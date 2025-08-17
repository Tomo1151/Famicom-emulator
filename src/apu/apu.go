package apu

/*
typedef unsigned char Uint8;
void SquareWaveCallback(void *userdata, Uint8 *stream, int len);
*/
import "C"
import (
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

// MARK: APUの定義
type APU struct {
	Ch1Register SquareWaveRegister
	Ch1Channel chan SquareNote
	Ch1Buffer *RingBuffer
}

// MARK: APUの初期化メソッド
func (a *APU) Init() {
	a.Ch1Register = SquareWaveRegister{
		toneVolume:    0,
		sweep:         0,
		freqLow:       0,
		freqHighKeyOn: 0,
	}
	a.Ch1Buffer = &RingBuffer{}
	a.Ch1Channel = init1ch(a.Ch1Buffer)
}

// MARK: 1chへの書き込みメソッド（矩形波）
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

// MARK: 矩形波生成メソッド
//export SquareWaveCallback
func SquareWaveCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	buf := unsafe.Slice((*C.Uint8)(stream), n)

	// リングバッファから直接読み込み
	readBuffer := make([]uint8, n)
	squareWave.buffer.Read(readBuffer)

	// バッファをコピー
	for i := range n {
		buf[i] = C.Uint8(readBuffer[i])
	}
}

// MARK: 1chの初期化メソッド
func init1ch(buffer *RingBuffer) chan SquareNote {
	ch1Channel := make(chan SquareNote, 10) // バッファ付きチャンネル
	// SquareWave構造体を初期化
	squareWave = SquareWave{
		freq:   44100.0,
		phase:  0.0,
		channel: ch1Channel,
		buffer: buffer,
		note: SquareNote{
			hz: 0,
			volume: 0.0,
			duty:   0.0,
		},
	}

	// PCM生成のgoroutineを開始
	go squareWave.generatePCM()

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

	return ch1Channel
}
