package apu

/*
typedef unsigned char Uint8;
void SineWave(void *userdata, Uint8 *stream, int len);
void SquareWaveCallback(void *userdata, Uint8 *stream, int len);
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	CPU_CLOCK = 1_789_773 // 1.78MHz
	MAX_VOLUME = 2
	toneHz   = 440
	sampleHz = 44100
	BUFFER_SIZE = 8192 // リングバッファサイズ
)

// リングバッファ構造体
type RingBuffer struct {
	buffer    [BUFFER_SIZE]uint8
	writePos  int
	readPos   int
	mutex     sync.RWMutex
}

func (rb *RingBuffer) Write(data []uint8) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	for _, sample := range data {
		rb.buffer[rb.writePos] = sample
		rb.writePos = (rb.writePos + 1) % BUFFER_SIZE
	}
}

func (rb *RingBuffer) Read(data []uint8) int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	
	readCount := 0
	for i := range data {
		if rb.readPos == rb.writePos {
			// バッファが空の場合は無音（128はU8フォーマットの中央値）
			data[i] = 128
		} else {
			data[i] = rb.buffer[rb.readPos]
			rb.readPos = (rb.readPos + 1) % BUFFER_SIZE
			readCount++
		}
	}
	return readCount
}

func (rb *RingBuffer) Available() int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	
	if rb.writePos >= rb.readPos {
		return rb.writePos - rb.readPos
	}
	return BUFFER_SIZE - rb.readPos + rb.writePos
}


type APU struct {
	Ch1Register SquareWaveRegister
	Ch1Channel chan SquareNote
	ringBuffer *RingBuffer
}

type SquareWaveRegister struct {
	toneVolume    uint8
	sweep         uint8
	freqLow       uint8
	freqHighKeyOn uint8
}

func (a *APU) Init() {
	a.Ch1Register = SquareWaveRegister{
		toneVolume:    0,
		sweep:         0,
		freqLow:       0,
		freqHighKeyOn: 0,
	}
	a.ringBuffer = &RingBuffer{}
	a.Ch1Channel = init1ch(a.ringBuffer)
}

func (a *APU) Write1ch(address uint16, data uint8) {
	a.Ch1Register.write(address, data)

	// fmt.Printf("Ch1 write: %f hz / %f % / %f", a.Ch1Register.freq(), a.Ch1Register.duty(), a.Ch1Register.volume())

	// SDL側に送信
	var volume float32
	if a.Ch1Register.isEnabled() {
		volume = a.Ch1Register.volume()
	} else {
		volume = 0.0
	}
	a.Ch1Channel <- SquareNote{
		hz: a.Ch1Register.freq(),
		duty: a.Ch1Register.duty(),
		volume: volume,
	}
}

type SquareWave struct {
	freq   float32
	phase  float32
	channel chan SquareNote
	note SquareNote
	ringBuffer *RingBuffer
}

type SquareNote struct {
	hz     float32
	volume float32
	duty   float32
}


var squareWave SquareWave

// PCM波形生成のgoroutine
func (sw *SquareWave) generatePCM() {
	for {
		// チャンネルから新しい音符を受信
		select {
		case note := <-sw.channel:
			sw.note = note
			sw.phase = 0.0 // 音符が変わったらphaseをリセット
		default:
			// 新しい音符がない場合は現在の音符を継続
		}

		// バッファに十分なデータがある場合は少し待つ
		if sw.ringBuffer.Available() > BUFFER_SIZE/2 {
			continue
		}

		// 小さなバッファでPCMサンプルを生成
		const chunkSize = 512 // チャンクサイズを大きくしてアンダーランを防ぐ
		pcmBuffer := make([]uint8, chunkSize)
		
		// 現在の音符の周波数に基づいてphaseIncrementを計算
		phaseIncrement := float32(sw.note.hz) / float32(sampleHz)

		for i := range chunkSize {
			sw.phase += phaseIncrement
			if sw.phase >= 1.0 {
				sw.phase -= 1.0
			}

			var sample uint8
			if sw.note.volume > 0 { // ボリュームが0より大きい場合のみ音を出す
				if sw.phase < sw.note.duty {
					sample = uint8(128 + (MAX_VOLUME * sw.note.volume * 127 / 256)) // 正の波形
				} else {
					sample = uint8(128 - (MAX_VOLUME * sw.note.volume * 127 / 256)) // 負の波形
				}
			} else {
				sample = 128 // 無音（中央値）
			}
			pcmBuffer[i] = sample
		}

		// リングバッファに書き込み
		sw.ringBuffer.Write(pcmBuffer)
	}
}

//export SquareWaveCallback
func SquareWaveCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	buf := unsafe.Slice((*C.Uint8)(stream), n)

	// リングバッファから直接読み込み
	readBuffer := make([]uint8, n)
	squareWave.ringBuffer.Read(readBuffer)
	
	// バッファをコピー
	for i := range n {
		buf[i] = C.Uint8(readBuffer[i])
	}
}

func initSdlAudio() {
	spec := &sdl.AudioSpec{
		Freq:     sampleHz,
		Format:   sdl.AUDIO_U8,
		Channels: 1,
		Samples:  2048, // バッファサイズを大きくしてアンダーランを防ぐ
		Callback: sdl.AudioCallback(C.SquareWaveCallback),
	}
	if err := sdl.OpenAudio(spec, nil); err != nil {
		panic(err)
	}

	// オーディオ再生開始
	sdl.PauseAudio(false)
}


func init1ch(ringBuffer *RingBuffer) chan SquareNote {
	ch1Channel := make(chan SquareNote, 10) // バッファ付きチャンネル
	// SquareWave構造体を初期化
	squareWave = SquareWave{
		freq:   44100.0,
		phase:  0.0,
		channel: ch1Channel,
		ringBuffer: ringBuffer,
		note: SquareNote{
			hz: 0,
			volume: 0.0,
			duty:   0.0,
		},
	}

	// PCM生成のgoroutineを開始
	go squareWave.generatePCM()

	initSdlAudio()
	return ch1Channel
}

func (swr *SquareWaveRegister) write(address uint16, data uint8) {
	switch address {
	case 0x4000:
		swr.toneVolume = data
	case 0x4001:
		swr.sweep = data
	case 0x4002:
		swr.freqLow = data
	case 0x4003:
		swr.freqHighKeyOn = data
	default:
		panic(fmt.Sprintf("APU Error: Invalid write at: %04X", address))
	}
}

func (swr *SquareWaveRegister) duty() float32 {
	// 00: 12.5%, 01: 25.0%, 10: 50.0%, 11: 75.0%
	value := (swr.toneVolume & 0xC0) >> 6
	switch value {
	case 0b00:
		return 0.125
	case 0b01:
		return 0.25
	case 0b10:
		return 0.50
	case 0b11:
		return 0.75
	default:
		return 0.0
	}
}

func (swr *SquareWaveRegister) isEnabled() bool {
	// より長く音を鳴らすために、長さカウンタのチェックを緩くする
	// 実際のNESでは長さカウンタが複雑だが、今回は簡単にする
	return (swr.toneVolume & 0x0F) > 0 // ボリュームが0より大きければ有効
}

func (swr *SquareWaveRegister) volume() float32 {
	// 0が消音，15が最大 ※ スウィープ無効時のみ
	return float32(swr.toneVolume & 0x0F) / 15.0
}

func (swr *SquareWaveRegister) freq() float32 {
    value := ((uint16(swr.freqHighKeyOn) & 0x07) << 8) | uint16(swr.freqLow)
    return CPU_CLOCK / (16.0 * float32(value) + 1.0)
}

