package apu

/*
typedef unsigned char Uint8;
typedef float Float32;
void MixedAudioCallback(void *userdata, Uint8 *stream, int len);
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	CPU_CLOCK = 1_789_772.5 // 1.78MHz
	APU_CYCLE_INTERVAL = 7457
	MAX_VOLUME = 0.4
	toneHz   = 440
	sampleHz = 44100
	BUFFER_SIZE = 8192 // リングバッファサイズ
)

// MARK: APUの定義
type APU struct {
	// CH1
	Ch1Register SquareWaveRegister
	Ch1Channel chan SquareNote
	Ch1Buffer *RingBuffer

	// CH2
	Ch2Register SquareWaveRegister
	Ch2Channel chan SquareNote
	Ch2Buffer *RingBuffer

	// CH3
	Ch3Register TriangleWaveRegister
	Ch3Channel chan TriangleNote
	Ch3Buffer *RingBuffer

	// CH4
	Ch4Register NoiseWaveRegister
	Ch4Channel chan NoiseNote
	Ch4Buffer *RingBuffer

	frameCounter FrameCounter
	cycles uint
	counter uint
	Status StatusRegister
}

// MARK: APUの初期化メソッド
func (a *APU) Init() {
	// CH1
	a.Ch1Register = SquareWaveRegister{}
	a.Ch1Register.Init()
	a.Ch1Buffer = &RingBuffer{}
	a.Ch1Buffer.Init()
	a.Ch1Channel = initSquareChannel(&squareWave1, a.Ch1Buffer)

	// CH2
	a.Ch2Register = SquareWaveRegister{}
	a.Ch2Register.Init()
	a.Ch2Buffer = &RingBuffer{}
	a.Ch2Buffer.Init()
	a.Ch2Channel = initSquareChannel(&squareWave2, a.Ch2Buffer)

	// CH3
	a.Ch3Register = TriangleWaveRegister{}
	a.Ch3Register.Init()
	a.Ch3Buffer = &RingBuffer{}
	a.Ch3Buffer.Init()
	a.Ch3Channel = initTriangleChannel(a.Ch3Buffer)

	// CH4
	a.Ch4Register = NoiseWaveRegister{}
	a.Ch4Register.Init()
	a.Ch4Buffer = &RingBuffer{}
	a.Ch4Buffer.Init()
	a.Ch4Channel = initNoiseChannel(a.Ch4Buffer)

	a.frameCounter = FrameCounter{}
	a.frameCounter.Init()

	a.cycles = 0
	a.Status = StatusRegister{}
	a.Status.Init()

	// オーディオデバイスの初期化
	a.initAudioDevice()
}

// MARK: 1chへの書き込みメソッド（矩形波）
func (a *APU) Write1ch(address uint16, data uint8) {
	a.Ch1Register.write(address, data)

	// SDL側に送信
	var volume float32
	if a.Ch1Register.isEnabled() {
		volume = a.Ch1Register.getVolume()
	} else {
		volume = 0.0
	}
	a.Ch1Channel <- SquareNote{
		hz: a.Ch1Register.getFrequency(),
		duty: a.Ch1Register.getDuty(),
		volume: volume,
	}
}

// MARK: 2chへの書き込みメソッド（矩形波）
func (a *APU) Write2ch(address uint16, data uint8) {
	a.Ch2Register.write(address, data)

	var volume float32
	if a.Ch2Register.isEnabled() {
		volume = a.Ch2Register.getVolume()
	} else {
		volume = 0.0
	}
	a.Ch2Channel <- SquareNote{
		hz: a.Ch2Register.getFrequency(),
		duty: a.Ch2Register.getDuty(),
		volume: volume,
	}
}

// MARK: 3chへの書き込みメソッド（三角波）
func (a *APU) Write3ch(address uint16, data uint8) {
	a.Ch3Register.write(address, data)

	a.Ch3Channel <- TriangleNote{
		hz: a.Ch3Register.getFrequency(),
	}
}

// MARK: 4chへの書き込みメソッド（ノイズ）
func (a *APU) Write4ch(address uint16, data uint8) {
	a.Ch4Register.write(address, data)

	a.Ch4Channel <- NoiseNote{
		hz: a.Ch4Register.getFrequency(),
		volume: a.Ch4Register.getVolume(),
		noiseMode: a.Ch4Register.getMode(),
	}
}

// MARK: フレームカウンタの書き込みメソッド
func (a *APU) WriteFrameCounter(data uint8) {
	a.frameCounter.update(data)
	a.counter = 0
	a.cycles = 0
}

// MARK: ステータスレジスタの読み取りメソッド
func (a *APU) ReadStatus() uint8 {
	a.Status.ClearFrameIRQ()
	return a.Status.ToByte()
}

// MARK: ステータスレジスタの書き込みメソッド
func (a *APU) WriteStatus(data uint8) {
	a.Status.update(data)
}


// MARK: 全チャンネルをミックスした音声生成コールバック
//export MixedAudioCallback
func MixedAudioCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length) / 4
	buffer := unsafe.Slice((*float32)(unsafe.Pointer(stream)), n)

	// 1chのデータの読み込み
	ch1Buffer := make([]float32, n)
	squareWave1.buffer.Read(ch1Buffer)

	// 2chのデータの読み込み
	ch2Buffer := make([]float32, n)
	squareWave2.buffer.Read(ch2Buffer)

	// 3chのデータの読み込み
	ch3Buffer := make([]float32, n)
	triangleWave.buffer.Read(ch3Buffer)

	// 4chのデータの読み込み
	ch4Buffer := make([]float32, n)
	noiseWave.buffer.Read(ch4Buffer)

	for i := range n {
		// ミックス
		mixed := (ch1Buffer[i] + ch2Buffer[i] + ch3Buffer[i] + ch4Buffer[i]) / 75

		if mixed > MAX_VOLUME {
			mixed = MAX_VOLUME
		} else if mixed < -MAX_VOLUME {
			mixed = -MAX_VOLUME
		}

		buffer[i] = mixed
	}
}

// MARK: オーディオの初期化メソッド
func (a *APU) initAudioDevice() {
	spec := &sdl.AudioSpec{
		Freq: sampleHz,
		Format: sdl.AUDIO_F32,
		Channels: 1,
		Samples: 2048,
		Callback: sdl.AudioCallback(C.MixedAudioCallback),
	}
	if err := sdl.OpenAudio(spec, nil); err != nil {
		panic(err)
	}

	// オーディオ再生開始
	sdl.PauseAudio(false)
}

// MARK: 1ch/2chの初期化メソッド
func initSquareChannel(wave *SquareWave, buffer *RingBuffer) chan SquareNote {
	ch1Channel := make(chan SquareNote, 10)

	// SquareWave構造体を初期化
	*wave = SquareWave{
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
	go wave.generatePCM()

	return ch1Channel
}

// MARK: 3chの初期化メソッド
func initTriangleChannel(buffer *RingBuffer) chan TriangleNote {
	ch3Channel := make(chan TriangleNote, 10)

	triangleWave = TriangleWave{
		freq: 44100.0,
		phase: 0.0,
		channel: ch3Channel,
		buffer: buffer,
		note: TriangleNote{
			hz: 0.0,
		},
	}

	go triangleWave.generatePCM()

	return ch3Channel
}

// MARK: 4ch の初期化メソッド
func initNoiseChannel(buffer *RingBuffer) chan NoiseNote {
	ch4Channel := make(chan NoiseNote, 10)

	// NoiseWave構造体を初期化
	noiseWave = NoiseWave{
		freq:   44100.0,
		phase:  0.0,
		channel: ch4Channel,
		buffer: buffer,
		noise: false,
		note: NoiseNote{
			hz: 0,
			volume: 0.0,
			noiseMode: NOISE_MODE_SHORT,
		},

		longNoise: NoiseShiftRegister{},
		shortNoise: NoiseShiftRegister{},
	}

	noiseWave.shortNoise.InitWithShortMode()
	noiseWave.longNoise.InitWithLongMode()

	// PCM生成のgoroutineを開始
	go noiseWave.generatePCM()

	return ch4Channel
}

// MARK: CPUと同期してサイクルを進めるメソッド
func (a *APU) Tick(cycles uint) {
	a.cycles++

	if a.cycles >= APU_CYCLE_INTERVAL {
		a.cycles %= APU_CYCLE_INTERVAL
		a.counter++

		mode := a.frameCounter.getMode()

		switch mode {
		case 4:
			if a.counter == 2 || a.counter == 4 {
				// 長さカウンタとスイープ用のクロック生成
			}
			if a.counter == 4 {
				// 割り込みフラグをセット
				a.counter = 0
				a.Status.SetFrameIRQ()
			}
			if a.counter == 1 || a.counter == 2 || a.counter == 3 || a.counter ==4 {}
		case 5:
			if a.counter == 2 || a.counter == 4 {
				// 長さカウンタとスイープ用のクロック生成
			}
			if a.counter == 1 || a.counter == 2 || a.counter == 3 || a.counter ==4 {}
		default:
			panic(fmt.Sprintf("APU Error: unexpected Frame sequencer mode: %04X", mode))
		}
	}
}

