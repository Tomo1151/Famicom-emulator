package apu

/*
#include <stdint.h>
void AudioMixCallback(void* userdata, uint8_t* stream, int length);
*/
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数定義
const (
	CPUClock           = 1_789_772.5 // 1.78MHz
	SampleRate         = 44100       // 44.1kHz
	APU_CYCLE_INTERVAL = 7457
)

// MARK: APUの定義
type TAPU struct {
	cycles uint
	step   uint8

	channel1 SquareWaveChannel
	channel2 SquareWaveChannel

	frameSequencer FrameSequencer
	status         StatusRegister

	buffer BlipBuffer

	sampleClock uint64
	userHandle  cgo.Handle
}

// MARK: APUの初期化メソッド
func (a *TAPU) Init() {
	a.cycles = 0
	a.step = 0

	a.channel1 = SquareWaveChannel{}
	a.channel1.Init()
	a.channel2 = SquareWaveChannel{}
	a.channel2.Init()

	a.frameSequencer = FrameSequencer{}
	a.frameSequencer.Init()

	a.status = StatusRegister{}
	a.status.Init()

	a.buffer.Init(SampleRate)

	// オーディオデバイスの初期化
	a.initAudioDevice()
}

// MARK: オーディオデバイスの初期化メソッド
func (a *TAPU) initAudioDevice() {
	handle := cgo.NewHandle(a)
	a.userHandle = handle

	spec := &sdl.AudioSpec{
		Freq:     SampleRate,
		Format:   sdl.AUDIO_F32,
		Channels: 1,
		Samples:  2048,
		Callback: sdl.AudioCallback(C.AudioMixCallback),
		UserData: unsafe.Pointer(uintptr(handle)),
	}

	if err := sdl.OpenAudio(spec, nil); err != nil {
		panic(err)
	}

	// オーディオ再生開始
	sdl.PauseAudio(false)
}

// MARK: ステータスレジスタの読み込みメソッド
func (a *TAPU) ReadStatus() uint8 {
	status := a.status.ToByte()
	status &= 0xF0
	// @TODO: 各チャンネルの長さカウントを反映させる
	a.status.ClearFrameIRQ()
	return status
}

// MARK: ステータスレジスタの書き込みメソッド
func (a *TAPU) WriteStatus(data uint8) {
	prev := a.status.ToByte()
	a.status.update(data)

	// @TODO: ミュートと長さカウンタのリセットも行う
	if (prev&(1<<STATUS_REG_ENABLE_1CH_POS)) != 0 && !a.status.is1chEnabled() {
		a.channel1.lengthCounter.counter = 0
	}
	if (prev&(1<<STATUS_REG_ENABLE_2CH_POS)) != 0 && !a.status.is2chEnabled() {
		a.channel2.lengthCounter.counter = 0
	}
}

// MARK: フレームIRQを取得
func (a *TAPU) FrameIRQ() bool {
	return a.status.FrameIRQ()
}

// MARK: フレームシーケンサの書き込みメソッド
func (a *TAPU) WriteFrameSequencer(data uint8) {
	a.frameSequencer.update(data)
	a.step = 0
	a.cycles = 0
	a.status.ClearFrameIRQ()
}

// MARK: 1chへの書き込みメソッド (矩形波)
func (a *TAPU) Write1ch(address uint16, data uint8) {
	a.channel1.register.write(address, data)

	switch address {
	case 0x4000:
		// エンベロープ更新
		a.channel1.envelope.data.rate = a.channel1.register.volume
		a.channel1.envelope.data.enabled = a.channel1.register.envelope
		a.channel1.envelope.data.loop = !a.channel1.register.keyOffCounter
	case 0x4001:
		// スイープ更新
		a.channel1.sweepUnit.data.shift = a.channel1.register.sweepShift
		a.channel1.sweepUnit.data.direction = a.channel1.register.sweepDirection
		a.channel1.sweepUnit.data.timerCount = a.channel1.register.sweepPeriod
		a.channel1.sweepUnit.data.enabled = a.channel1.register.sweepEnabled
	case 0x4002:
		a.channel1.sweepUnit.frequency = a.channel1.register.frequency
	case 0x4003:
		// 長さカウンタのリセット (有効時のみ)
		if a.status.is1chEnabled() {
			a.channel1.lengthCounter.data.Init(
				a.channel1.register.keyOffCount,
				!a.channel1.register.keyOffCounter,
			)
			a.channel1.lengthCounter.reset()
			a.channel1.envelope.reset()
			a.channel1.sweepUnit.frequency = a.channel1.register.frequency
			a.channel1.sweepUnit.reset()
			a.channel1.phase = 0
		}
	}
}

// MARK: 2chへの書き込みメソッド (矩形波)
func (a *TAPU) Write2ch(address uint16, data uint8) {
	a.channel2.register.write(address, data)

	switch address {
	case 0x4004:
		// エンベロープ更新
		a.channel2.envelope.data.rate = a.channel2.register.volume
		a.channel2.envelope.data.enabled = a.channel2.register.envelope
		a.channel2.envelope.data.loop = !a.channel2.register.keyOffCounter
	case 0x4005:
		// スイープ更新
		a.channel2.sweepUnit.data.shift = a.channel2.register.sweepShift
		a.channel2.sweepUnit.data.direction = a.channel2.register.sweepDirection
		a.channel2.sweepUnit.data.timerCount = a.channel2.register.sweepPeriod
		a.channel2.sweepUnit.data.enabled = a.channel2.register.sweepEnabled
	case 0x4006:
		a.channel2.sweepUnit.frequency = a.channel2.register.frequency
	case 0x4007:
		// 長さカウンタのリセット (有効時のみ)
		if a.status.is2chEnabled() {
			a.channel2.lengthCounter.data.Init(
				a.channel2.register.keyOffCount,
				!a.channel2.register.keyOffCounter,
			)
			a.channel2.lengthCounter.reset()
			a.channel2.envelope.reset()
			a.channel2.sweepUnit.frequency = a.channel2.register.frequency
			a.channel2.sweepUnit.reset()
			a.channel2.phase = 0
		}
	}
}

//export AudioMixCallback
func AudioMixCallback(userdata unsafe.Pointer, stream *C.uint8_t, length C.int) {
	// apu := (*TAPU)(userdata)

	// // stream を float32 のスライスに変換
	// out := (*[1 << 24]float32)(unsafe.Pointer(&stream))[: length/4 : length/4]

	// // Blip_Buffer から stream にサンプルを書き込む
	// apu.buffer.Read(out, len(out))
	handle := cgo.Handle(uintptr(userdata))
	apu := handle.Value().(*TAPU)

	// stream を float32 のスライスに変換
	// AUDIO_F32フォーマットなので、バッファはfloat32の配列として扱います。
	n := int(length) / 4
	out := unsafe.Slice((*float32)(unsafe.Pointer(stream)), n)
	fmt.Println("Audio sample:", out[0], out[1], out[2])
	// Blip_Buffer から stream にサンプルを書き込む
	apu.buffer.Read(out, len(out))
}

// MARK: APUのサイクルを進める
func (a *TAPU) Tick(cycles uint) {
	a.cycles += cycles
	a.sampleClock += uint64(cycles)
	a.clockFrameSequencer()

	// サンプル生成
	sample1ch := a.channel1.output(cycles)
	sample2ch := a.channel2.output(cycles)

	mixed := (sample1ch + sample2ch) / 2

	a.buffer.addDelta(a.sampleClock, mixed)
	if a.cycles%100 == 0 {
		fmt.Println("TAPU mixed: ", mixed)
	}
}

// MARK: エンベロープのクロック
func (a *TAPU) clockEnvelopes() {
	a.channel1.envelope.tick()
	a.channel2.envelope.tick()
}

// MARK: スイープユニットのクロック
func (a *TAPU) clockSweepUnits() {
	a.channel1.sweepUnit.tick(
		&a.channel1.lengthCounter,
		true,
	)
	a.channel2.sweepUnit.tick(
		&a.channel2.lengthCounter,
		false,
	)
}

// MARK: 線形カウンタのクロック
func (a *TAPU) clockLinearCounter() {
}

// MARK: 長さカウンタのクロック
func (a *TAPU) clockLengthCounter() {
	a.channel1.lengthCounter.tick()
	a.channel2.lengthCounter.tick()
}

// MARK: フレームシーケンサのクロック
func (a *TAPU) clockFrameSequencer() {
	if a.cycles >= APU_CYCLE_INTERVAL {
		a.cycles %= APU_CYCLE_INTERVAL
		a.step++
		mode := a.frameSequencer.Mode()

		switch mode {
		case 4:
			/*
				エンベロープ： e e e e   240Hz
				長さカウンタ： - l - l   120Hz
				割り込み　　： - - - f    60Hz
			*/
			if a.step == 1 || a.step == 2 || a.step == 3 || a.step == 4 {
				// エンベロープと線形カウンタのクロック生成
				a.clockEnvelopes()
				a.clockLinearCounter()
			}
			if a.step == 2 || a.step == 4 {
				// 長さカウンタとスイープユニットのクロック生成
				a.clockLengthCounter()
				a.clockSweepUnits()
			}
			if a.step == 4 {
				// 割り込みフラグのセット
				a.step = 0
				if !a.frameSequencer.DisableIRQ() {
					a.status.SetFrameIRQ()
				}
			}
		case 5:
			/*
				エンベロープ： e e e e -   192Hz
				長さカウンタ： l - l - -    96Hz
				割り込み　　： - - - - -   割り込みフラグセット無し
			*/
			if a.step == 1 || a.step == 2 || a.step == 3 || a.step == 4 {
				// エンベロープと線形カウンタのクロック生成
				a.clockEnvelopes()
				a.clockLinearCounter()
			}
			if a.step == 1 || a.step == 3 {
				// 長さカウンタとスイープユニットのクロック生成
				a.clockLengthCounter()
				a.clockSweepUnits()
			}
			if a.step == 5 {
				a.step = 0
			}
		}
	}
}

type AudioChannel interface {
	output() float32
}

type AudioChannelRegister interface {
	write(uint16, uint8)
}

type SquareWaveChannel struct {
	register      SquareWaveRegister
	envelope      Envelope
	lengthCounter LengthCounter
	sweepUnit     SweepUnit
	phase         float32
}

func (swc *SquareWaveChannel) Init() {
	swc.register = SquareWaveRegister{}
	swc.register.Init()
	swc.envelope = Envelope{}
	swc.envelope.Init()
	swc.lengthCounter = LengthCounter{}
	swc.lengthCounter.Init()
	swc.sweepUnit = SweepUnit{}
	swc.sweepUnit.Init()
}

func (swc *SquareWaveChannel) output(cycles uint) float32 {
	frequency := swc.sweepUnit.frequency
	if frequency < 8 || frequency > 0x7FF || swc.lengthCounter.isMuted() || swc.sweepUnit.isMuted() {
		// ミュートの時は0.0を返す
		return 0.0
	}

	// 進める位相 (進んだクロック数 / 1周期に必要なクロック数)
	period := float32(16.0 * (frequency + 1))
	swc.phase += float32(cycles) / period

	if swc.phase >= 1.0 {
		// 0.0 ~ 1.0 の範囲に制限
		swc.phase -= 1.0
	}

	value := 0.0
	if swc.phase <= swc.register.Duty() {
		value = 1.0
	} else {
		value = -1.0
	}

	return float32(value) * swc.envelope.volume()
}
