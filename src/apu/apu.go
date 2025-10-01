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
	CPU_CLOCK          = 1_789_772.5 // 1.78MHz
	APU_CYCLE_INTERVAL = 7457
	MAX_VOLUME         = 0.8
	toneHz             = 440
	sampleHz           = 44100
	BUFFER_SIZE        = 16384 // リングバッファサイズ
	CHUNK_SIZE         = 1024
)

var (
	ch1BufferPool = make([]float32, BUFFER_SIZE)
	ch2BufferPool = make([]float32, BUFFER_SIZE)
	ch3BufferPool = make([]float32, BUFFER_SIZE)
	ch4BufferPool = make([]float32, BUFFER_SIZE)
)

// MARK: APUの定義
type APU struct {
	// CH1
	Ch1Register    SquareWaveRegister
	Ch1Channel     chan SquareWaveEvent
	Ch1Receiver    chan ChannelEvent
	Ch1Buffer      *RingBuffer
	Ch1LengthCount uint8

	// CH2
	Ch2Register    SquareWaveRegister
	Ch2Channel     chan SquareWaveEvent
	Ch2Receiver    chan ChannelEvent
	Ch2Buffer      *RingBuffer
	Ch2LengthCount uint8

	// CH3
	Ch3Register    TriangleWaveRegister
	Ch3Channel     chan TriangleWaveEvent
	Ch3Receiver    chan ChannelEvent
	Ch3Buffer      *RingBuffer
	Ch3LengthCount uint8

	// CH4
	Ch4Register    NoiseWaveRegister
	Ch4Channel     chan NoiseWaveEvent
	Ch4Receiver    chan ChannelEvent
	Ch4Buffer      *RingBuffer
	Ch4LengthCount uint8

	// CH5
	Ch5Register    DMCRegister
	Ch5Channel     chan DMCWaveEvent
	Ch5Receiver    chan ChannelEvent
	Ch5Buffer      *RingBuffer
	Ch5LengthCount uint8

	frameCounter FrameCounter
	cycles       uint
	counter      uint
	Status       StatusRegister
}

type ChannelEvent struct {
	length uint8
}

// MARK: APUの初期化メソッド
func (a *APU) Init() {
	// CH1
	a.Ch1Register = SquareWaveRegister{}
	a.Ch1Register.Init()
	a.Ch1Buffer = &RingBuffer{}
	a.Ch1Buffer.Init()
	a.Ch1Channel, a.Ch1Receiver = initSquareChannel(1, &squareWave1, a.Ch1Buffer)
	a.Ch1LengthCount = 0

	// CH2
	a.Ch2Register = SquareWaveRegister{}
	a.Ch2Register.Init()
	a.Ch2Buffer = &RingBuffer{}
	a.Ch2Buffer.Init()
	a.Ch2Channel, a.Ch2Receiver = initSquareChannel(2, &squareWave2, a.Ch2Buffer)
	a.Ch2LengthCount = 0

	// CH3
	a.Ch3Register = TriangleWaveRegister{}
	a.Ch3Register.Init()
	a.Ch3Buffer = &RingBuffer{}
	a.Ch3Buffer.Init()
	a.Ch3Channel, a.Ch3Receiver = initTriangleChannel(a.Ch3Buffer)
	a.Ch3LengthCount = 0

	// CH4
	a.Ch4Register = NoiseWaveRegister{}
	a.Ch4Register.Init()
	a.Ch4Buffer = &RingBuffer{}
	a.Ch4Buffer.Init()
	a.Ch4Channel, a.Ch4Receiver = initNoiseChannel(a.Ch4Buffer)
	a.Ch4LengthCount = 0

	// CH5
	a.Ch5Register = DMCRegister{}
	a.Ch5Register.Init()
	a.Ch5Buffer = &RingBuffer{}
	a.Ch5Buffer.Init()
	a.Ch5Channel, a.Ch5Receiver = a.initDMCChannel(a.Ch5Buffer)
	a.Ch5LengthCount = 0

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

	if address == 0x4000 {
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
			eventType:    SQUARE_WAVE_ENVELOPE,
			envelopeData: &envelopeData,
		}
	}

	if address == 0x4000 || address == 0x4003 {
		lengthCounterData := LengthCounterData{}
		lengthCounterData.Init(
			a.Ch1Register.keyOffCount,
			a.Ch1Register.keyOffCounter,
		)

		a.Ch1Channel <- SquareWaveEvent{
			eventType:         SQUARE_WAVE_LENGTH_COUNTER,
			lengthCounterData: &lengthCounterData,
		}
	}

	if address == 0x4001 {
		sweepUnitData := SweepUnitData{}
		sweepUnitData.Init(
			a.Ch1Register.sweepShift,
			a.Ch1Register.sweepDirection,
			a.Ch1Register.sweepPeriod,
			a.Ch1Register.sweepEnabled,
		)

		a.Ch1Channel <- SquareWaveEvent{
			eventType:     SQUARE_WAVE_SWEEP,
			sweepUnitData: &sweepUnitData,
		}
	}

	if address == 0x4002 || address == 0x4003 {
		a.Ch1Channel <- SquareWaveEvent{
			eventType: SQUARE_WAVE_SWEEP_FREQUENCY,
			frequency: &a.Ch1Register.frequency,
		}
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

	if address == 0x4004 {
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
			eventType:    SQUARE_WAVE_ENVELOPE,
			envelopeData: &envelopeData,
		}
	}

	if address == 0x4004 || address == 0x4007 {
		lengthCounterData := LengthCounterData{}
		lengthCounterData.Init(
			a.Ch2Register.keyOffCount,
			a.Ch2Register.keyOffCounter,
		)

		a.Ch2Channel <- SquareWaveEvent{
			eventType:         SQUARE_WAVE_LENGTH_COUNTER,
			lengthCounterData: &lengthCounterData,
		}
	}

	if address == 0x4005 {
		sweepUnitData := SweepUnitData{}
		sweepUnitData.Init(
			a.Ch2Register.sweepShift,
			a.Ch2Register.sweepDirection,
			a.Ch2Register.sweepPeriod,
			a.Ch2Register.sweepEnabled,
		)

		a.Ch2Channel <- SquareWaveEvent{
			eventType:     SQUARE_WAVE_SWEEP,
			sweepUnitData: &sweepUnitData,
		}
	}

	if address == 0x4006 || address == 0x4007 {
		a.Ch2Channel <- SquareWaveEvent{
			eventType: SQUARE_WAVE_SWEEP_FREQUENCY,
			frequency: &a.Ch2Register.frequency,
		}
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

	if address == 0x400A || address == 0x400B {
		a.Ch3Channel <- TriangleWaveEvent{
			eventType: TRIANGLE_WAVE_NOTE,
			note: &TriangleNote{
				hz: a.Ch3Register.getFrequency(),
			},
		}
	}

	if address == 0x4008 || address == 0x400B {
		lengthCounterData := LengthCounterData{}
		lengthCounterData.Init(
			a.Ch3Register.keyOffCount,
			a.Ch3Register.keyOffCounter,
		)
		a.Ch3Channel <- TriangleWaveEvent{
			eventType:         TRIANGLE_WAVE_LENGTH_COUNTER,
			lengthCounterData: &lengthCounterData,
		}
	}

	if address == 0x4008 {
		linearCounterData := LinearCounterData{}
		linearCounterData.Init(
			a.Ch3Register.length,
			a.Ch3Register.keyOffCounter,
		)
		a.Ch3Channel <- TriangleWaveEvent{
			eventType:         TRIANGLE_WAVE_LINEAR_COUNTER,
			linearCounterData: &linearCounterData,
		}
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

	if address == 0x400E {
		a.Ch4Channel <- NoiseWaveEvent{
			eventType: NOISE_WAVE_NOTE,
			note: &NoiseNote{
				hz:        a.Ch4Register.getFrequency(),
				noiseMode: a.Ch4Register.getMode(),
			},
		}
	}

	if address == 0x400C {
		envelopeData := EnvelopeData{}
		envelopeData.Init(
			a.Ch4Register.volume,
			a.Ch4Register.envelope,
			!a.Ch4Register.keyOffCounter,
		)
		a.Ch4Channel <- NoiseWaveEvent{
			eventType:    NOISE_WAVE_ENVELOPE,
			envelopeData: &envelopeData,
		}
	}

	if address == 0x400C || address == 0x400F {
		lengthCounterData := LengthCounterData{}
		lengthCounterData.Init(
			a.Ch4Register.keyOffCount,
			a.Ch4Register.keyOffCounter,
		)
		a.Ch4Channel <- NoiseWaveEvent{
			eventType:         NOISE_WAVE_LENGTH_COUNTER,
			lengthCounterData: &lengthCounterData,
		}
	}

	if address == 0x400F {
		a.Ch4Channel <- NoiseWaveEvent{
			eventType: NOISE_WAVE_RESET,
		}
	}
}

// MARK: 5chへの書き込みメソッド（DMC）
func (a *APU) Write5ch(address uint16, data uint8) {
	a.Ch5Register.write(address, data)
}

// MARK: フレームカウンタの書き込みメソッド
func (a *APU) WriteFrameCounter(data uint8) {
	a.frameCounter.update(data)
	a.counter = 0
	a.cycles = 0
}

// MARK: ステータスレジスタの読み取りメソッド
func (a *APU) ReadStatus() uint8 {
	status := a.Status.ToByte()

	a.receiveEvents()
	status = status & 0xF0

	if a.Ch1LengthCount == 0 {
		status |= 0 << 0
	} else {
		status |= 1 << 0
	}
	if a.Ch2LengthCount == 0 {
		status |= 0 << 1
	} else {
		status |= 1 << 1
	}
	if a.Ch3LengthCount == 0 {
		status |= 0 << 2
	} else {
		status |= 1 << 2
	}
	if a.Ch4LengthCount == 0 {
		status |= 0 << 3
	} else {
		status |= 1 << 3
	}
	if a.Ch5LengthCount == 0 {
		status |= 0 << 4
	} else {
		status |= 1 << 4
	}

	a.Status.ClearFrameIRQ()

	return status
}

// MARK: ステータスレジスタの書き込みメソッド
func (a *APU) WriteStatus(data uint8) {
	// 更新前の状態を保存
	wasCh1Enabled := a.Status.is1chEnabled()
	wasCh2Enabled := a.Status.is2chEnabled()
	wasCh3Enabled := a.Status.is3chEnabled()
	wasCh4Enabled := a.Status.is4chEnabled()
	wasCh5Enabled := a.Status.is5chEnabled()

	a.Status.update(data)

	// 各チャンネルの状態によってミュートにする
	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENABLED,
		enabled:   a.Status.is1chEnabled(),
		changed:   wasCh1Enabled && !a.Status.is1chEnabled(),
	}
	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENABLED,
		enabled:   a.Status.is2chEnabled(),
		changed:   wasCh2Enabled && !a.Status.is2chEnabled(),
	}
	a.Ch3Channel <- TriangleWaveEvent{
		eventType: TRIANGLE_WAVE_ENABLED,
		enabled:   a.Status.is3chEnabled(),
		changed:   wasCh3Enabled && !a.Status.is3chEnabled(),
	}
	a.Ch4Channel <- NoiseWaveEvent{
		eventType: NOISE_WAVE_ENABLED,
		enabled:   a.Status.is4chEnabled(),
		changed:   wasCh4Enabled && !a.Status.is4chEnabled(),
	}
	a.Ch5Channel <- DMCWaveEvent{
		eventType: DMC_WAVE_ENABLED,
		enabled:   a.Status.is5chEnabled(),
		changed:   wasCh5Enabled && !a.Status.is5chEnabled(),
	}

	// disableされた時に長さカウンタも落とす (halt)
	if wasCh1Enabled && !a.Status.is1chEnabled() {
		a.Ch1LengthCount = 0
	}
	if wasCh2Enabled && !a.Status.is2chEnabled() {
		a.Ch2LengthCount = 0
	}
	if wasCh3Enabled && !a.Status.is3chEnabled() {
		a.Ch3LengthCount = 0
	}
	if wasCh4Enabled && !a.Status.is4chEnabled() {
		a.Ch4LengthCount = 0
	}
}

// MARK: 全チャンネルをミックスした音声生成コールバック
//
//export MixedAudioCallback
func MixedAudioCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length) / 4
	buffer := unsafe.Slice((*float32)(unsafe.Pointer(stream)), n)

	// 事前確保バッファを使用
	ch1Buffer := ch1BufferPool[:n]
	ch2Buffer := ch2BufferPool[:n]
	ch3Buffer := ch3BufferPool[:n]
	ch4Buffer := ch4BufferPool[:n]

	// 1chのデータの読み込み
	squareWave1.buffer.Read(ch1Buffer)

	// 2chのデータの読み込み
	squareWave2.buffer.Read(ch2Buffer)

	// 3chのデータの読み込み
	triangleWave.buffer.Read(ch3Buffer)

	// 4chのデータの読み込み
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
		Freq:     sampleHz,
		Format:   sdl.AUDIO_F32,
		Channels: 1,
		Samples:  2048,
		Callback: sdl.AudioCallback(C.MixedAudioCallback),
	}
	if err := sdl.OpenAudio(spec, nil); err != nil {
		panic(err)
	}

	// オーディオ再生開始
	sdl.PauseAudio(false)
}

// MARK: 1ch/2chの初期化メソッド
func initSquareChannel(channelNumber uint, wave *SquareWave, buffer *RingBuffer) (chan SquareWaveEvent, chan ChannelEvent) {
	ch1Channel := make(chan SquareWaveEvent, 100)
	sendChannel := make(chan ChannelEvent, 100)

	envelope := Envelope{}
	envelope.Init()
	lengthCounter := LengthCounter{}
	lengthCounter.Init()
	sweepUnit := SweepUnit{}
	sweepUnit.Init()

	// SquareWave構造体を初期化
	*wave = SquareWave{
		channelNumber: channelNumber,
		freq:          44100.0,
		phase:         0.0,
		channel:       ch1Channel,
		sender:        sendChannel,
		buffer:        buffer,
		note: SquareNote{
			duty: 0.0,
		},
		envelope:      envelope,
		lengthCounter: lengthCounter,
		sweepUnit:     sweepUnit,
		enabled:       true,
	}

	// PCM生成のgoroutineを開始
	go wave.generatePCM()

	return ch1Channel, sendChannel
}

// MARK: 3chの初期化メソッド
func initTriangleChannel(buffer *RingBuffer) (chan TriangleWaveEvent, chan ChannelEvent) {
	ch3Channel := make(chan TriangleWaveEvent, 100)
	sendChannel := make(chan ChannelEvent, 100)

	lengthCounter := LengthCounter{}
	lengthCounter.Init()
	linearCounter := LinearCounter{}
	linearCounter.Init()

	triangleWave = TriangleWave{
		freq:    44100.0,
		phase:   0.0,
		channel: ch3Channel,
		sender:  sendChannel,
		buffer:  buffer,
		note: TriangleNote{
			hz: 0.0,
		},
		lengthCounter: lengthCounter,
		linearCounter: linearCounter,
		enabled:       true,
	}

	go triangleWave.generatePCM()

	return ch3Channel, sendChannel
}

// MARK: 4ch の初期化メソッド
func initNoiseChannel(buffer *RingBuffer) (chan NoiseWaveEvent, chan ChannelEvent) {
	ch4Channel := make(chan NoiseWaveEvent, 100)
	sendChannel := make(chan ChannelEvent, 100)

	lengthCounter := LengthCounter{}
	lengthCounter.Init()

	// NoiseWave構造体を初期化
	noiseWave = NoiseWave{
		freq:    44100.0,
		phase:   0.0,
		channel: ch4Channel,
		sender:  sendChannel,
		buffer:  buffer,
		noise:   false,
		note: NoiseNote{
			hz:        0,
			noiseMode: NOISE_MODE_SHORT,
		},
		lengthCounter: lengthCounter,

		longNoise:  NoiseShiftRegister{},
		shortNoise: NoiseShiftRegister{},
		enabled:    true,
	}

	noiseWave.shortNoise.InitWithShortMode()
	noiseWave.longNoise.InitWithLongMode()

	// PCM生成のgoroutineを開始
	go noiseWave.generatePCM()

	return ch4Channel, sendChannel
}

// MARK: 5chの初期化メソッド
func (a *APU) initDMCChannel(buffer *RingBuffer) (chan DMCWaveEvent, chan ChannelEvent) {
	ch5Channel := make(chan DMCWaveEvent, 1000)
	sendChannel := make(chan ChannelEvent, 1000)

	dpcmWave = DMCWave{
		freq:    44100.0,
		phase:   0.0,
		channel: ch5Channel,
		sender:  sendChannel,
		note:    DMCNote{},
		buffer:  buffer,
		enabled: true,
	}

	go dpcmWave.generatePCM()

	return ch5Channel, sendChannel
}

func (a *APU) sendEnvelopeTick() {
	a.Ch1Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENVELOPE_TICK,
	}
	a.Ch2Channel <- SquareWaveEvent{
		eventType: SQUARE_WAVE_ENVELOPE_TICK,
	}
	a.Ch3Channel <- TriangleWaveEvent{
		eventType: TRIANGLE_WAVE_LINEAR_COUNTER_TICK,
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
		eventType: TRIANGLE_WAVE_LENGTH_COUNTER_TICK,
	}
	a.Ch4Channel <- NoiseWaveEvent{
		eventType: NOISE_WAVE_LENGTH_COUNTER_TICK,
	}
}

// MARK: イベント受け取りのメソッド
func (a *APU) receiveEvents() {
ch1EventLoop:
	for {
		select {
		case event := <-a.Ch1Receiver:
			a.Ch1LengthCount = event.length
		default:
			break ch1EventLoop
		}
	}

ch2EventLoop:
	for {
		select {
		case event := <-a.Ch2Receiver:
			a.Ch2LengthCount = event.length
		default:
			break ch2EventLoop
		}
	}

ch3EventLoop:
	for {
		select {
		case event := <-a.Ch3Receiver:
			a.Ch3LengthCount = event.length
		default:
			break ch3EventLoop
		}
	}

ch4EventLoop:
	for {
		select {
		case event := <-a.Ch4Receiver:
			a.Ch4LengthCount = event.length
		default:
			break ch4EventLoop
		}
	}

	// ch5EventLoop:
	// for {
	// 	select {
	// 	case event := <-a.Ch5Receiver:
	// 		a.Ch5LengthCount = event.length
	// 	default:
	// 		break ch5EventLoop
	// 	}
	// }
}

// MARK: CPUと同期してサイクルを進めるメソッド
func (a *APU) Tick(cycles uint) {
	a.cycles += cycles

	if a.cycles >= APU_CYCLE_INTERVAL {
		a.cycles %= APU_CYCLE_INTERVAL
		a.counter++

		a.receiveEvents()
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
			if a.counter == 4 {
				// 割り込みフラグをセット
				a.counter = 0
				if !a.frameCounter.getDisableIRQ() {
					a.Status.SetFrameIRQ()
				}
			}
			if a.counter == 1 || a.counter == 2 || a.counter == 3 || a.counter == 4 {
				a.sendEnvelopeTick()
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
			if a.counter == 1 || a.counter == 2 || a.counter == 3 || a.counter == 4 {
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
