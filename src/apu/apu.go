package apu

/*
#include <stdint.h>
#include <stdlib.h>
void AudioMixCallback(void* userdata, uint8_t* stream, int length);
*/
import "C"
import (
	"Famicom-emulator/config"
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数定義
const (
	CPU_CLOCK          = 1_789_772.5 // 1.78MHz
	SAMPLE_RATE        = 44100       // 44.1kHz
	APU_CYCLE_INTERVAL = 7457        // 分周器の間隔
	BUFFER_SIZE        = 1024 * 4    // サンプルバッファサイズ
	MAX_VOLUME         = 1.0         // 最大音量
)

// MARK: 変数定義
var (
	// 各チャンネルのバッファの事前確保
	ch1Buffer [BUFFER_SIZE]float32
	ch2Buffer [BUFFER_SIZE]float32
	ch3Buffer [BUFFER_SIZE]float32
	ch4Buffer [BUFFER_SIZE]float32
	ch5Buffer [BUFFER_SIZE]float32
)

// CPUバスからデータを読み取るための関数型
type CpuBusReader func(uint16) uint8

// MARK: APUの定義
type APU struct {
	cycles uint
	step   uint8

	// チャンネル
	channel1 SquareWaveChannel
	channel2 SquareWaveChannel
	channel3 TriangleWaveChannel
	channel4 NoiseWaveChannel
	channel5 DMCWaveChannel

	frameCounter FrameCounter
	status       StatusRegister

	sampleClock uint64
	cpuRead     CpuBusReader

	// 各チャンネルの前回レベルを保持
	prevLevel1 float32
	prevLevel2 float32
	prevLevel3 float32
	prevLevel4 float32
	prevLevel5 float32

	isAudioInitialized bool
	config             config.Config
}

// MARK: APUの初期化メソッド
func (a *APU) Init(reader CpuBusReader, config config.Config) {
	a.cycles = 0
	a.step = 0
	a.cpuRead = reader
	a.config = config

	// 各チャンネルの初期化
	a.channel1.Init(a.config.APU.LOG_ENABLED)
	a.channel2.Init(a.config.APU.LOG_ENABLED)
	a.channel3.Init(a.config.APU.LOG_ENABLED)
	a.channel4.Init(a.config.APU.LOG_ENABLED)
	a.channel5.Init(a.cpuRead, a.config.APU.LOG_ENABLED)

	a.frameCounter.Init()
	a.status.Init()

	a.prevLevel1 = 0.0
	a.prevLevel2 = 0.0
	a.prevLevel3 = 0.0
	a.prevLevel4 = 0.0
	a.prevLevel5 = 0.0

	a.WriteFrameSequencer(0x00)

	// オーディオデバイスの初期化
	if !a.isAudioInitialized {
		a.initAudioDevice()
	}
}

// MARK: オーディオデバイスの初期化メソッド
func (a *APU) initAudioDevice() {
	handle := cgo.NewHandle(a)
	hptr := C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	*(*C.uintptr_t)(hptr) = C.uintptr_t(handle)
	spec := &sdl.AudioSpec{
		Freq:     SAMPLE_RATE,
		Format:   sdl.AUDIO_F32,
		Channels: 1,
		Samples:  BUFFER_SIZE / 2,
		Callback: sdl.AudioCallback(C.AudioMixCallback),
		UserData: hptr,
	}

	if err := sdl.OpenAudio(spec, nil); err != nil {
		panic(err)
	}

	// オーディオ再生開始
	sdl.PauseAudio(false)
	a.isAudioInitialized = true
}

// MARK: SDLのオーディオコールバック
//
//export AudioMixCallback
func AudioMixCallback(userdata unsafe.Pointer, stream *C.uint8_t, length C.int) {
	// APUの参照を取得
	if userdata == nil {
		return
	}

	// userdata に格納した uintptr を読み出す
	apuPointer := uintptr(*(*C.uintptr_t)(userdata))
	if apuPointer == 0 {
		return
	}

	h := cgo.Handle(apuPointer)
	apu, ok := h.Value().(*APU)
	if !ok || apu == nil {
		return
	}

	n := int(length) / 4
	buffer := unsafe.Slice((*float32)(unsafe.Pointer(stream)), n)

	ch1 := ch1Buffer[:n]
	ch2 := ch2Buffer[:n]
	ch3 := ch3Buffer[:n]
	ch4 := ch4Buffer[:n]
	ch5 := ch5Buffer[:n]

	if !apu.config.APU.MUTE_1CH {
		apu.channel1.buffer.Read(ch1, n)
	} else {
		apu.channel1.buffer.Fill(ch1, n, 0.0, apu.sampleClock)
	}
	if !apu.config.APU.MUTE_2CH {
		apu.channel2.buffer.Read(ch2, n)
	} else {
		apu.channel2.buffer.Fill(ch2, n, 0.0, apu.sampleClock)
	}
	if !apu.config.APU.MUTE_3CH {
		apu.channel3.buffer.Read(ch3, n)
	} else {
		apu.channel3.buffer.Fill(ch3, n, 0.0, apu.sampleClock)
	}
	if !apu.config.APU.MUTE_4CH {
		apu.channel4.buffer.Read(ch4, n)
	} else {
		apu.channel4.buffer.Fill(ch4, n, 0.0, apu.sampleClock)
	}
	if !apu.config.APU.MUTE_5CH {
		apu.channel5.buffer.Read(ch5, n)
	} else {
		apu.channel5.buffer.Fill(ch5, n, 0.0, apu.sampleClock)
	}

	for i := range n {
		// 全チャンネルをミックス
		mixed := mixSamples(ch1[i], ch2[i], ch3[i], ch4[i], ch5[i])

		if mixed > MAX_VOLUME {
			mixed = MAX_VOLUME
		} else if mixed < -MAX_VOLUME {
			mixed = -MAX_VOLUME
		}

		// SDLへサンプルとして渡す
		buffer[i] = mixed * apu.config.APU.SOUND_VOLUME
	}
}

// MARK: APUのサイクルを進める
func (a *APU) Tick(cycles uint) {
	a.cycles += cycles
	a.sampleClock += uint64(cycles)
	a.clockFrameSequencer()

	// DMCタイマーを進める
	a.channel5.tick(cycles)

	// 現在のレベルを計算
	var currentLevel1, currentLevel2, currentLevel3, currentLevel4, currentLevel5 float32
	// if a.status.is1chEnabled() {
	currentLevel1 = a.channel1.output(cycles)
	// }
	// if a.status.is2chEnabled() {
	currentLevel2 = a.channel2.output(cycles)
	// }
	if a.status.is3chEnabled() {
		currentLevel3 = a.channel3.output(cycles)
	}
	if a.status.is4chEnabled() {
		currentLevel4 = a.channel4.output(cycles)
	}

	// 5chは書き込み以外でレベルが変化しないため，ミュートに関係なく出力値を拾う
	currentLevel5 = a.channel5.output()

	// 前回レベルとの差分を計算
	delta1 := currentLevel1 - a.prevLevel1
	delta2 := currentLevel2 - a.prevLevel2
	delta3 := currentLevel3 - a.prevLevel3
	delta4 := currentLevel4 - a.prevLevel4
	delta5 := currentLevel5 - a.prevLevel5

	// レベルが変化した場合のみ、差分をバッファに追加
	if delta1 != 0 {
		a.channel1.buffer.addDelta(a.sampleClock, delta1)
		a.prevLevel1 = currentLevel1
	}
	if delta2 != 0 {
		a.channel2.buffer.addDelta(a.sampleClock, delta2)
		a.prevLevel2 = currentLevel2
	}
	if delta3 != 0 {
		a.channel3.buffer.addDelta(a.sampleClock, delta3)
		a.prevLevel3 = currentLevel3
	}
	if delta4 != 0 {
		a.channel4.buffer.addDelta(a.sampleClock, delta4)
		a.prevLevel4 = currentLevel4
	}
	if delta5 != 0 {
		a.channel5.buffer.addDelta(a.sampleClock, delta5)
		a.prevLevel5 = currentLevel5
	}
}

// MARK: ステータスレジスタの読み込みメソッド
func (a *APU) ReadStatus() uint8 {
	var status uint8 = 0

	// 1~4ch: length counter > 0 を見る
	if a.channel1.lengthCounter.counter > 0 {
		status |= 1 << STATUS_REG_ENABLE_1CH_POS
	}
	if a.channel2.lengthCounter.counter > 0 {
		status |= 1 << STATUS_REG_ENABLE_2CH_POS
	}
	if a.channel3.lengthCounter.counter > 0 {
		status |= 1 << STATUS_REG_ENABLE_3CH_POS
	}
	if a.channel4.lengthCounter.counter > 0 {
		status |= 1 << STATUS_REG_ENABLE_4CH_POS
	}

	// @FIXME DMCの再生状態を正しく反映する
	// 5ch: DMCが再生中かどうか
	status |= 1 << STATUS_REG_ENABLE_5CH_POS // TODO: DMC 実装に合わせて

	// FrameIRQ / DMCIRQ フラグの反映
	if a.status.FrameIRQ() {
		status |= 1 << STATUS_REG_ENABLE_FRAME_IRQ_POS
	}
	if a.status.EnableDMCIRQ() {
		status |= 1 << STATUS_REG_ENABLE_DMC_IRQ_POS
	}

	// $4015の読み込みはFrameIRQフラグをクリアする
	a.status.ClearFrameIRQ()
	return status
}

// MARK: ステータスレジスタの書き込みメソッド
func (a *APU) WriteStatus(data uint8) {
	prev := a.status.ToByte()
	a.status.update(data)

	// @TODO: ミュートと長さカウンタのリセットも行う
	/*
		有効ビットがクリアされると（$4015経由）、長さカウンタは強制的に0に設定され、有効ビットが再度セットされるまで変更できなくなる（長さカウンタの以前の値は破棄）。
		有効ビットをセットしても、すぐには効果はない。
	*/
	if (prev&(1<<STATUS_REG_ENABLE_1CH_POS)) != 0 && !a.status.is1chEnabled() {
		a.channel1.lengthCounter.counter = 0
	}
	if (prev&(1<<STATUS_REG_ENABLE_2CH_POS)) != 0 && !a.status.is2chEnabled() {
		a.channel2.lengthCounter.counter = 0
	}
	if (prev&(1<<STATUS_REG_ENABLE_3CH_POS)) != 0 && !a.status.is3chEnabled() {
		a.channel3.lengthCounter.counter = 0
	}
	if (prev&(1<<STATUS_REG_ENABLE_4CH_POS)) != 0 && !a.status.is4chEnabled() {
		a.channel4.lengthCounter.counter = 0
	}
	if (prev&(1<<STATUS_REG_ENABLE_5CH_POS)) != 0 && !a.status.is5chEnabled() {
		a.channel5.setEnabled(false)
	} else if (prev&(1<<STATUS_REG_ENABLE_5CH_POS)) == 0 && a.status.is5chEnabled() {
		a.channel5.setEnabled(true)
	}

	// $4015への書き込みはFrameIRQフラグをクリアする
	a.status.ClearFrameIRQ()
}

// MARK: フレームIRQを取得
func (a *APU) FrameIRQ() bool {
	return a.status.FrameIRQ()
}

// MARK: フレームシーケンサの書き込みメソッド
func (a *APU) WriteFrameSequencer(data uint8) {
	a.frameCounter.update(data)

	/*
		@NOTE:
			5ステップモード時のみ$4017の書き込みの副作用で halfフレーム/quarterフレーム信号を生成する
	*/
	if a.frameCounter.Mode() == 5 {
		a.clockEnvelopes()
		a.clockLengthCounter()
		a.clockSweepUnits()
		a.status.ClearFrameIRQ()
	}

	a.step = 0
	a.cycles = 0
}

// MARK: 1chへの書き込みメソッド (矩形波)
func (a *APU) Write1ch(address uint16, data uint8) {
	a.channel1.register.write(address, data)

	// @FIXME 既にレジスタに値が反映されているため、AudioChannel側でapply()などを用意し、一本化できるかも
	switch address {
	case 0x4000:
		/*
			$4000		ddld nnnn
				7-6 d   デューティ
				5   l   エンベロープループ
				4   d   エンベロープ無効
				3-0 n   ボリューム/エンベロープ周期
		*/
		a.channel1.duty = a.channel1.register.Duty()
		a.channel1.envelope.update(
			a.channel1.register.Volume(),
			a.channel1.register.EnvelopeLoop(),
			a.channel1.register.EnvelopeEnabled(),
		)
		a.channel1.lengthCounter.update(
			a.channel1.register.keyOffCount,
			a.channel1.register.LengthCounterHalt(),
		)
	case 0x4001:
		/*
			$4001		eppp nsss
				7   e   スイープ有効
				6-4 p   スイープ周期
				3   n   スイープ方向
				2-0 s   スイープ量
		*/
		a.channel1.sweepUnit.update(
			a.channel1.register.sweepShift,
			a.channel1.register.sweepDirection,
			a.channel1.register.sweepPeriod,
			a.channel1.register.sweepEnabled,
		)
	case 0x4002:
		/*
			$4002		llll llll
				7-0 l   チャンネル周期下位
		*/
		a.channel1.sweepUnit.frequency = a.channel1.register.frequency
	case 0x4003:
		/*
			$4003		cccc chhh
				7-3 c   長さカウンタインデクス
				2-0 h   チャンネル周期上位

				$4003への書き込みは長さカウンタのリロード，エンベロープの再起動，パルス生成器の位相のリセットが発生する
		*/
		a.channel1.sweepUnit.frequency = a.channel1.register.frequency
		a.channel1.lengthCounter.update(
			a.channel1.register.keyOffCount,
			a.channel1.register.LengthCounterHalt(),
		)
		a.channel1.envelope.reset()
		a.channel1.sweepUnit.reset()
		a.channel1.sequencer = 0
		a.channel1.timerReload = (uint16(a.channel1.register.frequency) + 1) * 2
		a.channel1.timer = a.channel1.timerReload

		if a.status.is1chEnabled() {
			a.channel1.lengthCounter.reload()
		}
	}
}

// MARK: 2chへの書き込みメソッド (矩形波)
func (a *APU) Write2ch(address uint16, data uint8) {
	a.channel2.register.write(address, data)

	switch address {
	case 0x4004:
		/*
			$4004		ddld nnnn
				7-6 d   デューティ
				5   l   エンベロープループ
				4   d   エンベロープ無効
				3-0 n   ボリューム/エンベロープ周期
		*/
		a.channel2.duty = a.channel2.register.Duty()
		a.channel2.envelope.update(
			a.channel2.register.Volume(),
			a.channel2.register.EnvelopeLoop(),
			a.channel2.register.EnvelopeEnabled(),
		)
		a.channel2.lengthCounter.update(
			a.channel2.register.keyOffCount,
			a.channel2.register.LengthCounterHalt(),
		)
	case 0x4005:
		/*
			$4005		eppp nsss
				7   e   スイープ有効
				6-4 p   スイープ周期
				3   n   スイープ方向
				2-0 s   スイープ量
		*/
		a.channel2.sweepUnit.update(
			a.channel2.register.sweepShift,
			a.channel2.register.sweepDirection,
			a.channel2.register.sweepPeriod,
			a.channel2.register.sweepEnabled,
		)
	case 0x4006:
		/*
			$4006		llll llll
				7-0 l   チャンネル周期下位
		*/
		a.channel2.sweepUnit.frequency = a.channel2.register.frequency
	case 0x4007:
		/*
			$4007		cccc chhh
				7-3 c   長さカウンタインデックス
				2-0 h   チャンネル周期上位

				$4007への書き込みは長さカウンタのリロード，エンベロープの再起動，パルス生成器の位相のリセットが発生する
		*/
		a.channel2.sweepUnit.frequency = a.channel2.register.frequency
		a.channel2.lengthCounter.update(
			a.channel2.register.keyOffCount,
			a.channel2.register.LengthCounterHalt(),
		)
		a.channel2.envelope.reset()
		a.channel2.sweepUnit.reset()
		a.channel2.sequencer = 0
		a.channel2.timerReload = (uint16(a.channel2.register.frequency) + 1) * 2
		a.channel2.timer = a.channel2.timerReload

		if a.status.is2chEnabled() {
			a.channel2.lengthCounter.reload()
		}
	}
}

// MARK: 3chの書き込みメソッド (三角波)
func (a *APU) Write3ch(address uint16, data uint8) {
	a.channel3.register.write(address, data)

	switch address {
	case 0x4008:
		/*
			$4008  clll llll
				7   c   長さカウンタ無効フラグ
				6-0 l   線形カウンタ
		*/
		a.channel3.lengthCounter.update(
			a.channel3.register.keyOffCount,
			a.channel3.register.LengthCounterHalt(),
		)
		a.channel3.linearCounter.update(
			a.channel3.register.length,
			a.channel3.register.keyOffCounter,
		)
	case 0x400A:
		/*
			$400A  llll llll
				7-0 l   チャンネル周期下位
		*/
		a.channel3.frequency = a.channel3.register.frequency
	case 0x400B:
		/*
			$400B  llll lhhh
				7-3 l   長さカウンタインデクス
				2-0 h   チャンネル周期上位
		*/
		a.channel3.frequency = a.channel3.register.frequency
		a.channel3.lengthCounter.update(
			a.channel3.register.keyOffCount,
			a.channel3.register.LengthCounterHalt(),
		)
		a.channel3.linearCounter.setReload()

		if a.status.is3chEnabled() {
			a.channel3.lengthCounter.reload()
		}
	}
}

// MARK: 4chの書き込みメソッド (ノイズ)
func (a *APU) Write4ch(address uint16, data uint8) {
	a.channel4.register.write(address, data)

	// @FIXME 既にレジスタに値が反映されているため、AudioChannel側でapply()などを用意し、一本化できるかも
	switch address {
	case 0x400C:
		/*
			$400C   --le nnnn
				5   l   エンベロープループ、長さカウンタ無効
				4   e   エンベロープ無効フラグ
				3-0 n   ボリューム/エンベロープ周期
		*/
		a.channel4.envelope.update(
			a.channel4.register.Volume(),
			a.channel4.register.EnvelopeLoop(),
			a.channel4.register.EnvelopeEnabled(),
		)
		a.channel4.lengthCounter.update(
			a.channel4.register.keyOffCount,
			a.channel4.register.LengthCounterHalt(),
		)
	case 0x400E:
		/*
			$400E   s--- pppp
				7   s   ランダム生成モード
				3-0 p   タイマ周期インデクス
		*/
		a.channel4.mode = a.channel4.register.Mode()
		a.channel4.shiftRegister.mode = a.channel4.register.Mode()
		a.channel4.index = a.channel4.register.frequency
	case 0x400F:
		/*
			$400F   llll l---
				7-3 l   長さインデクス

			$4003への書き込みは長さカウンタのリロード，エンベロープの再起動，パルス生成器の位相のリセットが発生する
		*/
		a.channel4.index = a.channel4.register.frequency
		a.channel4.lengthCounter.update(
			a.channel4.register.keyOffCount,
			a.channel4.register.LengthCounterHalt(),
		)
		a.channel4.envelope.reset()
		a.channel4.phase = 0

		if a.status.is4chEnabled() {
			a.channel4.lengthCounter.reload()
		}
	}
}

// MARK: 5chの書き込みメソッド (DMC)
func (a *APU) Write5ch(address uint16, data uint8) {
	a.channel5.register.write(address, data)

	switch address {
	case 0x4010:
		/*
			$4010    il-- ffff
				7   i    割り込み有効フラグ
				6   l    ループフラグ
				3-0 f    周期インデックス
		*/
		a.channel5.timerReload = dmcFrequencyTable[a.channel5.register.frequencyIndex]
		a.channel5.timer = a.channel5.timerReload
	case 0x4011:
		/*
			$4011    -ddd dddd
				6-0 d    デルタカウンタ初期値
		*/
		a.channel5.deltaCounter = a.channel5.register.deltaCounter
	case 0x4012:
		/*
			$4012    aaaa aaaa
				7-0 a    サンプル開始アドレス
		*/
		// a.channel5.baseAddress = uint16(a.channel5.register.sampleStartAddress)*0x40 + 0xC000
	case 0x4013:
		/*
			$4013    llll llll
				7-0 l    サンプルバイト数

				llll.llll0001 = (l * 16) + 1
		*/
		// a.channel5.byteCount = (uint16(data) << 4) + 1
	}
}

// MARK: バッファのフラッシュ
func (a *APU) EndFrame() {
	// フレームの終わりまでの時間を処理するため、現在のクロックを渡す
	a.channel1.buffer.endFrame(a.sampleClock)
	a.channel2.buffer.endFrame(a.sampleClock)
	a.channel3.buffer.endFrame(a.sampleClock)
	a.channel4.buffer.endFrame(a.sampleClock)
	a.channel5.buffer.endFrame(a.sampleClock)
}

// MARK: エンベロープのクロック (1ch/2ch/4ch)
func (a *APU) clockEnvelopes() {
	a.channel1.envelope.tick()
	a.channel2.envelope.tick()
	a.channel4.envelope.tick()
}

// MARK: スイープユニットのクロック (1ch/2ch)
func (a *APU) clockSweepUnits() {
	a.channel1.sweepUnit.tick(
		&a.channel1.lengthCounter,
		true,
	)
	a.channel2.sweepUnit.tick(
		&a.channel2.lengthCounter,
		false,
	)
}

// MARK: 線形カウンタのクロック (3ch)
func (a *APU) clockLinearCounter() {
	a.channel3.linearCounter.tick()
}

// MARK: 長さカウンタのクロック (1ch/2ch/3ch/4ch)
func (a *APU) clockLengthCounter() {
	a.channel1.lengthCounter.tick()
	a.channel2.lengthCounter.tick()
	a.channel3.lengthCounter.tick()
	a.channel4.lengthCounter.tick()
}

// MARK: フレームシーケンサのクロック
func (a *APU) clockFrameSequencer() {
	if a.cycles >= APU_CYCLE_INTERVAL {
		// フレームシーケンサは入力の1.789MHzを7457分周する
		a.cycles %= APU_CYCLE_INTERVAL
		a.step++
		mode := a.frameCounter.Mode()

		switch mode {
		case 4:
			/*
				エンベロープ/線型カウンタ： e e e e   240Hz
				長さカウンタ/スイープ　　： - l - l   120Hz
				割り込み　　 　　　　　　： - - - f    60Hz
			*/
			if a.step == 1 || a.step == 2 || a.step == 3 || a.step == 4 {
				// エンベロープと線形カウンタのクロック生成 (quarter frame)
				a.clockEnvelopes()
				a.clockLinearCounter()
			}
			if a.step == 2 || a.step == 4 {
				// 長さカウンタとスイープユニットのクロック生成 (half frame)
				a.clockLengthCounter()
				a.clockSweepUnits()
			}
			if a.step == 4 {
				// 割り込みフラグのセット
				a.step = 0
				if !a.frameCounter.DisableIRQ() {
					a.status.SetFrameIRQ()
				}
			}
		case 5:
			/*
				エンベロープ/線型カウンタ： e e e - e   192Hz
				長さカウンタ/スイープ　　： - l - - l    96Hz
				割り込み　　 　　　　　　： - - - - -   割り込みフラグセット無し
			*/
			if a.step == 1 || a.step == 2 || a.step == 3 || a.step == 4 {
				// エンベロープと線形カウンタのクロック生成 (quarter frame)
				a.clockEnvelopes()
				a.clockLinearCounter()
			}
			if a.step == 2 || a.step == 5 {
				// 長さカウンタとスイープユニットのクロック生成 (half frame)
				a.clockLengthCounter()
				a.clockSweepUnits()
			}
			if a.step == 5 {
				a.step = 0
			}
		}
	}
}

// MARK: リセット
func (a *APU) Reset() {
	a.WriteStatus(0x00)
	a.status.Init()

	a.channel1.buffer.Sync(0.0, a.sampleClock)
	a.channel2.buffer.Sync(0.0, a.sampleClock)
	a.channel3.buffer.Sync(0.0, a.sampleClock)
	a.channel4.buffer.Sync(0.0, a.sampleClock)
	a.channel5.buffer.Sync(0.0, a.sampleClock)

	a.prevLevel1 = 0.0
	a.prevLevel2 = 0.0
	a.prevLevel3 = 0.0
	a.prevLevel4 = 0.0
	a.prevLevel5 = 0.0
}

// MARK: デバッグ用ログ出力切り替え
func (a *APU) ToggleLog() {
	if a.config.APU.LOG_ENABLED {
		fmt.Println("[APU] Debug log: OFF")
	} else {
		fmt.Println("[APU] Debug log: ON")
	}
	a.config.APU.LOG_ENABLED = !a.config.APU.LOG_ENABLED
	a.channel1.ToggleLog()
	a.channel2.ToggleLog()
	a.channel3.ToggleLog()
	a.channel4.ToggleLog()
	a.channel5.ToggleLog()
}

// MARK: 1chミュート切り替え
func (a *APU) ToggleMute1ch() {
	if a.config.APU.MUTE_1CH {
		fmt.Println("[APU] 1ch (square): Unmuted")
	} else {
		fmt.Println("[APU] 1ch (square): Muted")
	}
	a.config.APU.MUTE_1CH = !a.config.APU.MUTE_1CH

	if a.config.APU.MUTE_1CH {
		// ミュートした瞬間はバッファを0に同期して過去サンプルを破棄
		a.channel1.buffer.Sync(0.0, a.sampleClock)
		// APU 側の prevLevel も 0 にして差分発生を防ぐ
		a.prevLevel1 = 0.0
	} else {
		// ミュート解除：現在の出力レベルに同期して過渡を抑える
		current := a.channel1.output(0)
		a.prevLevel1 = current
		a.channel1.buffer.Sync(current, a.sampleClock)
	}
}

// MARK: 2chミュート切り替え
func (a *APU) ToggleMute2ch() {
	if a.config.APU.MUTE_2CH {
		fmt.Println("[APU] 2ch (square): Unmuted")
	} else {
		fmt.Println("[APU] 2ch (square): Muted")
	}
	a.config.APU.MUTE_2CH = !a.config.APU.MUTE_2CH

	if a.config.APU.MUTE_2CH {
		a.channel2.buffer.Sync(0.0, a.sampleClock)
		a.prevLevel2 = 0.0
	} else {
		current := a.channel2.output(0)
		a.prevLevel2 = current
		a.channel2.buffer.Sync(current, a.sampleClock)
	}
}

// MARK: 3chミュート切り替え
func (a *APU) ToggleMute3ch() {
	if a.config.APU.MUTE_3CH {
		fmt.Println("[APU] 3ch (triangle): Unmuted")
	} else {
		fmt.Println("[APU] 3ch (triangle): Muted")
	}
	a.config.APU.MUTE_3CH = !a.config.APU.MUTE_3CH

	if a.config.APU.MUTE_3CH {
		a.channel3.buffer.Sync(0.0, a.sampleClock)
		a.prevLevel2 = 0.0
	} else {
		current := a.channel3.output(0)
		a.prevLevel3 = current
		a.channel3.buffer.Sync(current, a.sampleClock)
	}
}

// MARK: 4chミュート切り替え
func (a *APU) ToggleMute4ch() {
	if a.config.APU.MUTE_4CH {
		fmt.Println("[APU] 4ch (noise): Unmuted")
	} else {
		fmt.Println("[APU] 4ch (noise): Muted")
	}
	a.config.APU.MUTE_4CH = !a.config.APU.MUTE_4CH

	if a.config.APU.MUTE_4CH {
		a.channel4.buffer.Sync(0.0, a.sampleClock)
		a.prevLevel4 = 0.0
	} else {
		current := a.channel4.output(0)
		a.prevLevel4 = current
		a.channel4.buffer.Sync(current, a.sampleClock)
	}
}

// MARK: 5chミュート切り替え
func (a *APU) ToggleMute5ch() {
	if a.config.APU.MUTE_5CH {
		fmt.Println("[APU] 5ch (DPCM): Unmuted")
	} else {
		fmt.Println("[APU] 5ch (DPCM): Muted")
	}
	a.config.APU.MUTE_5CH = !a.config.APU.MUTE_5CH

	if a.config.APU.MUTE_5CH {
		a.channel5.buffer.Sync(0.0, a.sampleClock)
		a.prevLevel5 = 0.0
	} else {
		current := a.channel5.output()
		a.prevLevel5 = current
		a.channel5.buffer.Sync(current, a.sampleClock)
	}
}

// MARK: SDLコールバックのサンプル数を取得するメソッド
func AudioCallbackSampleCount() int {
	return BUFFER_SIZE / 2
}

// MARK: 最新の全チャンネルのサンプルを取得するメソッド
func GetRecentChannelSamples(n int) [][]float32 {
	if n <= 0 {
		return nil
	}
	if n > BUFFER_SIZE {
		n = BUFFER_SIZE
	}
	return [][]float32{
		ch1Buffer[:n],
		ch2Buffer[:n],
		ch3Buffer[:n],
		ch4Buffer[:n],
		ch5Buffer[:n],
	}
}

// MARK: 各チャンネルのサンプルを適切なバランスでミックスする関数
func mixSamples(pulse1 float32, pulse2 float32, triangle float32, noise float32, dmc float32) float32 {
	/*
		output = pulse_out + tnd_out

		                           95.88
		pulse_out = ------------------------------------
								(8128 / (pulse1 + pulse2)) + 100

																			159.79
		tnd_out = -------------------------------------------------------------
																				1
							----------------------------------------------------- + 100
								(triangle / 8227) + (noise / 12241) + (dmc / 22638)
	*/
	// 矩形波チャンネルのミックス
	var pulseOut float32
	pulseSum := pulse1 + pulse2
	if pulseSum > 0 {
		pulseOut = 95.88 / (8128/pulseSum + 100)
	}

	// 三角波、ノイズ、DMCチャンネルのミックス
	var tndOut float32
	tndSum := (triangle / 8227.0) + (noise / 12241.0) + (dmc / 22638.0)
	if tndSum > 0 {
		tndOut = 159.79 / (1/tndSum + 100)
	}

	return pulseOut + tndOut
}
