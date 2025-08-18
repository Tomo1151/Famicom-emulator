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
	MAX_VOLUME = 0.8
	toneHz   = 440
	sampleHz = 44100
	BUFFER_SIZE = 8192 // リングバッファサイズ
)

// MARK: APUの定義
type APU struct {
	// CH1
	Ch1Register SquareWaveRegister
	Ch1Channel chan SquareWaveEvent
	Ch1Buffer *RingBuffer

	// CH2
	Ch2Register SquareWaveRegister
	Ch2Channel chan SquareWaveEvent
	Ch2Buffer *RingBuffer

	// CH3
	Ch3Register TriangleWaveRegister
	Ch3Channel chan TriangleWaveEvent
	Ch3Buffer *RingBuffer

	// CH4
	Ch4Register NoiseWaveRegister
	Ch4Channel chan NoiseWaveEvent
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

	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_NOTE,
		note: &SquareNote{
			duty: a.Ch1Register.getDuty(),
		},
	}

	envelopeData := EnvelopeData{}
	envelopeData.Init(
		a.Ch1Register.volume,
		a.Ch1Register.envelope,
		!a.Ch1Register.keyOffCounter,
	)
	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENVELOPE,
		envelopeData: &envelopeData,
	}

	lengthCounterData := LengthCounterData{}
	lengthCounterData.Init(
		a.Ch1Register.keyOffCount,
		a.Ch1Register.keyOffCounter,
	)

	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_LENGTH_COUNTER,
		lengthCounterData: &lengthCounterData,
	}

	sweepUnitData := SweepUnitData{}
	sweepUnitData.Init(
		a.Ch1Register.frequency,
		a.Ch1Register.sweepShift,
		a.Ch1Register.sweepDirection,
		a.Ch1Register.sweepPeriod,
		a.Ch1Register.sweepEnabled,
	)

	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_SWEEP,
		sweepUnitData: &sweepUnitData,
	}

	if address == 0x4003 {
		a.Ch1Channel <- SquareWaveEvent{
			eventType: SQUARE_WAVE_RESET,
		}
	}
}

// MARK: 2chへの書き込みメソッド（矩形波）
func (a *APU) Write2ch(address uint16, data uint8) {
	a.Ch2Register.write(address, data)

	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_NOTE,
		note: &SquareNote{
			duty: a.Ch2Register.getDuty(),
		},
	}

	envelopeData := EnvelopeData{}
	envelopeData.Init(
		a.Ch2Register.volume,
		a.Ch2Register.envelope,
		!a.Ch2Register.keyOffCounter,
	)
	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENVELOPE,
		envelopeData: &envelopeData,
	}

	lengthCounterData := LengthCounterData{}
	lengthCounterData.Init(
		a.Ch2Register.keyOffCount,
		a.Ch2Register.keyOffCounter,
	)

	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_LENGTH_COUNTER,
		lengthCounterData: &lengthCounterData,
	}


	sweepUnitData := SweepUnitData{}
	sweepUnitData.Init(
		a.Ch2Register.frequency,
		a.Ch2Register.sweepShift,
		a.Ch2Register.sweepDirection,
		a.Ch2Register.sweepPeriod,
		a.Ch2Register.sweepEnabled,
	)

	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_SWEEP,
		sweepUnitData: &sweepUnitData,
	}

	if address == 0x4007 {
		a.Ch2Channel <- SquareWaveEvent{
			eventType: SQUARE_WAVE_RESET,
		}
	}
}

// MARK: 3chへの書き込みメソッド（三角波）
func (a *APU) Write3ch(address uint16, data uint8) {
	a.Ch3Register.write(address, data)

	a.Ch3Channel <- TriangleWaveEvent{
		eventType: TRIANGLE_WAVE_NOTE,
		note: &TriangleNote{
			hz: a.Ch3Register.getFrequency(),
		},
	}

	lengthCounterData := LengthCounterData{}
	lengthCounterData.Init(
		a.Ch3Register.keyOffCount,
		a.Ch3Register.keyOffCounter,
	)
	a.Ch3Channel <- TriangleWaveEvent{
		eventType: TRIANGLE_WAVE_LENGTH_COUNTER,
		lengthCounterData: &lengthCounterData,
	}

	a.Ch3Channel <- TriangleWaveEvent{
		eventType: TRIANGLE_WAVE_LINEAR_COUNTER,
		linearCounterData: &LinearCounterData{
			count: a.Ch3Register.length,
		},
	}

	if address == 0x400B {
		a.Ch3Channel <- TriangleWaveEvent{
			eventType: TRIANGLE_WAVE_RESET,
		}
	}
}

// MARK: 4chへの書き込みメソッド（ノイズ）
func (a *APU) Write4ch(address uint16, data uint8) {
	a.Ch4Register.write(address, data)

	a.Ch4Channel <- NoiseWaveEvent{
		eventType: NOISE_WAVE_NOTE,
		note: &NoiseNote{
			hz: a.Ch4Register.getFrequency(),
			noiseMode: a.Ch4Register.getMode(),
		},
	}

	envelopeData := EnvelopeData{}
	envelopeData.Init(
		a.Ch4Register.volume,
		a.Ch4Register.envelope,
		!a.Ch4Register.keyOffCounter,
	)
	a.Ch4Channel <- NoiseWaveEvent{
		eventType: NOISE_WAVE_ENVELOPE,
		envelopeData: &envelopeData,
	}

	lengthCounterData := LengthCounterData{}
	lengthCounterData.Init(
		a.Ch4Register.keyOffCount,
		a.Ch4Register.keyOffCounter,
	)
	a.Ch4Channel <- NoiseWaveEvent{
		eventType: TRIANGLE_WAVE_LENGTH_COUNTER,
		lengthCounterData: &lengthCounterData,
	}

	if address == 0x400F {
		a.Ch4Channel <- NoiseWaveEvent{
			eventType: NOISE_WAVE_RESET,
		}
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

	// 各チャンネルの状態によってミュートにする
	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENABLED,
		enabled: a.Status.enable1ch,
	}
	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENABLED,
		enabled: a.Status.enable2ch,
	}
	if !a.Status.is3chEnabled() {
		triangleWave.note.hz = 0.0
	}
	a.Ch4Channel <- NoiseWaveEvent{
		eventType: NOISE_WAVE_ENABLED,
		enabled: a.Status.enable4ch,
	}
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
func initSquareChannel(wave *SquareWave, buffer *RingBuffer) chan SquareWaveEvent {
	ch1Channel := make(chan SquareWaveEvent, 10)

	envelope := Envelope{}
	envelope.Init()
	lengthCounter := LengthCounter{}
	lengthCounter.Init()
	sweepUnit := SweepUnit{}
	sweepUnit.Init()

	// SquareWave構造体を初期化
	*wave = SquareWave{
		freq:   44100.0,
		phase:  0.0,
		channel: ch1Channel,
		buffer: buffer,
		note: SquareNote{
			duty:   0.0,
		},
		envelope: envelope,
		lengthCounter: lengthCounter,
		sweepUnit: sweepUnit,
		enabled: true,
	}

	// PCM生成のgoroutineを開始
	go wave.generatePCM()

	return ch1Channel
}

// MARK: 3chの初期化メソッド
func initTriangleChannel(buffer *RingBuffer) chan TriangleWaveEvent {
	ch3Channel := make(chan TriangleWaveEvent, 10)

	lengthCounter := LengthCounter{}
	lengthCounter.Init()
	linearCounter := LinearCounter{}
	linearCounter.Init()

	triangleWave = TriangleWave{
		freq: 44100.0,
		phase: 0.0,
		channel: ch3Channel,
		buffer: buffer,
		note: TriangleNote{
			hz: 0.0,
		},
		lengthCounter: lengthCounter,
		linearCounter: linearCounter,
		enabled: true,
	}

	go triangleWave.generatePCM()

	return ch3Channel
}

// MARK: 4ch の初期化メソッド
func initNoiseChannel(buffer *RingBuffer) chan NoiseWaveEvent {
	ch4Channel := make(chan NoiseWaveEvent, 10)

	lengthCounter := LengthCounter{}
	lengthCounter.Init()

	// NoiseWave構造体を初期化
	noiseWave = NoiseWave{
		freq:   44100.0,
		phase:  0.0,
		channel: ch4Channel,
		buffer: buffer,
		noise: false,
		note: NoiseNote{
			hz: 0,
			noiseMode: NOISE_MODE_SHORT,
		},
		lengthCounter: lengthCounter,

		longNoise: NoiseShiftRegister{},
		shortNoise: NoiseShiftRegister{},
	}

	noiseWave.shortNoise.InitWithShortMode()
	noiseWave.longNoise.InitWithLongMode()

	// PCM生成のgoroutineを開始
	go noiseWave.generatePCM()

	return ch4Channel
}

func (a *APU) sendEnvelopeTick() {
	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENVELOPE_TICK,
	}
	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENVELOPE_TICK,
	}
	a.Ch4Channel <- NoiseWaveEvent{
		eventType: NOISE_WAVE_ENVELOPE_TICK,
	}
}

func (a *APU) sendSweepTick() {
	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_SWEEP_TICK,
	}
	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_SWEEP_TICK,
	}
}

func (a *APU) sendLengthCounterTick() {
	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_LENGTH_COUNTER_TICK,
	}
	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_LENGTH_COUNTER_TICK,
	}
	a.Ch3Channel <- TriangleWaveEvent{
		eventType: SQUARE_WAVE_LENGTH_COUNTER_TICK,
	}
	a.Ch4Channel <- NoiseWaveEvent{
		eventType: SQUARE_WAVE_LENGTH_COUNTER_TICK,
	}
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
			/*
				割り込み　　： - - - f    60Hz
				長さカウンタ： - l - l   120Hz
				エンベロープ： e e e e   240Hz
			*/
			if a.counter == 2 || a.counter == 4 {
				// 長さカウンタとスイープ用のクロック生成
				a.sendLengthCounterTick()
				a.sendSweepTick()
			}
			if a.counter == 1 || a.counter == 2 || a.counter == 3 || a.counter ==4 {
				a.sendEnvelopeTick()
			}
			if a.counter == 4 {
				// 割り込みフラグをセット
				a.counter = 0
				if !a.frameCounter.getDisableIRQ() {
					a.Status.SetFrameIRQ()
				}
			}
		case 5:
			/*
				割り込み　　： - - - - -   割り込みフラグセット無し
				長さカウンタ： l - l - -    96Hz
				エンベロープ： e e e e -   192Hz
			*/
			if a.counter == 1 || a.counter == 3 {
				// 長さカウンタとスイープ用のクロック生成
				a.sendLengthCounterTick()
				a.sendSweepTick()
			}
			if a.counter == 1 || a.counter == 2 || a.counter == 3 || a.counter ==4 {
				a.sendEnvelopeTick()
			}
			if a.counter == 5 {
				a.counter = 0
			}
		default:
			panic(fmt.Sprintf("APU Error: unexpected Frame sequencer mode: %04X", mode))
		}
	}
}

