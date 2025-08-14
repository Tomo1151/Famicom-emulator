package apu

/*
typedef unsigned char Uint8;
void SineWave(void *userdata, Uint8 *stream, int len);
void SquareWaveCallback(void *userdata, Uint8 *stream, int len);
*/
import "C"
import (
	"log"
	"math"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	MAX_VOLUME = 50
	toneHz   = 440
	sampleHz = 48000
	dPhase   = 2 * math.Pi * toneHz / sampleHz
)

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
	}

	// 現在の音符の周波数に基づいてphaseIncrementを計算
	phaseIncrement := float32(squareWave.note.hz) / float32(sampleHz)

	for i := 0; i < n; i += 2 {
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
		buf[i+1] = sample
	}
}



func PlaySquareWave() {
    // SquareWave構造体を初期化
    squareWave = SquareWave{
        freq:   44100.0,
        phase:  0.0,
        channel: make(chan SquareNote),
        note: SquareNote{
            hz: 261.626,
            volume: 0.1,
            duty:   0.5,
        },
    }

    spec := &sdl.AudioSpec{
        Freq:     sampleHz,
        Format:   sdl.AUDIO_U8,
        Channels: 2,
        Samples:  sampleHz,
        Callback: sdl.AudioCallback(C.SquareWaveCallback),
    }
    if err := sdl.OpenAudio(spec, nil); err != nil {
        log.Println(err)
        return
    }

    // オーディオ再生開始
    sdl.PauseAudio(false)

    // 2秒待機（先頭の音）
    sdl.Delay(2000)

    // 次の音符に切り替え
    squareWave.channel <- SquareNote{
        hz: 293.665,
        volume: 0.1,
        duty:   0.75,
    }

    // さらに2秒待機
    sdl.Delay(2000)

    // sdl.CloseAudio()
    close(squareWave.channel)
}