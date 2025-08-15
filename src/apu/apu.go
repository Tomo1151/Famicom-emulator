package apu

/*
typedef unsigned char Uint8;
void SineWave(void *userdata, Uint8 *stream, int len);
void SquareWaveCallback(void *userdata, Uint8 *stream, int len);
*/
import "C"
import (
	"fmt"
	"math"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	CPU_CLOCK = 1_789_773 // 1.78MHz
	MAX_VOLUME = 2
	toneHz   = 440
	sampleHz = 48000
	dPhase   = 2 * math.Pi * toneHz / sampleHz
)


type APU struct {
	Ch1Register SquareWaveRegister
	Ch1Channel chan SquareNote
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
	a.Ch1Channel = init1ch()
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
}

type SquareNote struct {
	hz     float32
	volume float32
	duty   float32
}


var squareWave SquareWave

//export SquareWaveCallback
func SquareWaveCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	buf := unsafe.Slice((*C.Uint8)(stream), n)

	// チャンネルから新しい音符を受信
	select {
	case note := <-squareWave.channel:
		squareWave.note = note
		squareWave.phase = 0.0 // 音符が変わったらphaseをリセット
	default:
		// squareWave.note = SquareNote{
		// 	hz: 0,
		// 	duty: 0,
		// 	volume: 0,
		// }
	}

	// 現在の音符の周波数に基づいてphaseIncrementを計算
	phaseIncrement := float32(squareWave.note.hz) / float32(sampleHz)

	for i := range n {
		squareWave.phase += phaseIncrement
		if squareWave.phase >= 1.0 {
			squareWave.phase -= 1.0
		}

		var sample C.Uint8
		if squareWave.phase < squareWave.note.duty {
			sample = C.Uint8(MAX_VOLUME * squareWave.note.volume)
		} else {
			sample = C.Uint8(0)
		}
		buf[i] = sample
	}
}

func initSdlAudio() {
	spec := &sdl.AudioSpec{
		Freq:     sampleHz,
		Format:   sdl.AUDIO_U8,
		Channels: 1,
		Samples:  1024,
		Callback: sdl.AudioCallback(C.SquareWaveCallback),
	}
	if err := sdl.OpenAudio(spec, nil); err != nil {
		panic(err)
	}

	// オーディオ再生開始
	sdl.PauseAudio(false)
}


func init1ch() chan SquareNote {
	ch1Channel := make(chan SquareNote)
	// SquareWave構造体を初期化
	squareWave = SquareWave{
		freq:   44100.0,
		phase:  0.0,
		channel: ch1Channel,
		note: SquareNote{
			hz: 0,
			volume: 0.0,
			duty:   0.0,
		},
	}

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
	// $4003 (freqHighKeyOn)のbit 7-3は長さカウンタ、これが0なら音は停止
	return (swr.freqHighKeyOn & 0xF8) != 0
}

func (swr *SquareWaveRegister) volume() float32 {
	// 0が消音，15が最大 ※ スウィープ無効時のみ
	return float32(swr.toneVolume & 0x0F) / 15.0
}

func (swr *SquareWaveRegister) freq() float32 {
    value := ((uint16(swr.freqHighKeyOn) & 0x07) << 8) | uint16(swr.freqLow)
    return CPU_CLOCK / (16.0 * float32(value) + 1.0)
}

