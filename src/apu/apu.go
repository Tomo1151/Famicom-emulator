package apu

/*
#include <stdint.h>
void AudioMixCallback(void* userdata, uint8_t* stream, int length);
*/
import "C"
import (
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数定義
const (
	CPU_CLOCK          = 1_789_772.5 // 1.78MHz
	SAMPLE_RATE        = 44100       // 44.1kHz
	APU_CYCLE_INTERVAL = 7457        // 分周器の間隔
	BUFFER_SIZE        = 16384       // サンプルバッファサイズ
	MAX_VOLUME         = 0.8
)

// MARK: APUの定義
type APU struct {
	cycles uint
	step   uint8

	// チャンネル
	channel1 *SquareWaveChannel
	channel2 *SquareWaveChannel
	channel3 *TriangleWaveChannel
	channel4 *NoiseWaveChannel

	frameCounter FrameCounter
	status       StatusRegister

	sampleClock uint64

	// 各チャンネルの前回レベルを保持
	prevLevel1 float32
	prevLevel2 float32
	prevLevel3 float32
	prevLevel4 float32
}

// MARK: APUの初期化メソッド
func (a *APU) Init() {
	a.cycles = 0
	a.step = 0

	a.channel1 = &square1
	a.channel1.Init()
	a.channel2 = &square2
	a.channel2.Init()
	a.channel3 = &triangle
	a.channel3.Init()
	a.channel4 = &noise
	a.channel4.Init()

	a.frameCounter = FrameCounter{}
	a.frameCounter.Init()

	a.status = StatusRegister{}
	a.status.Init()

	a.prevLevel1 = 0.0
	a.prevLevel2 = 0.0
	a.prevLevel3 = 0.0
	a.prevLevel4 = 0.0

	// オーディオデバイスの初期化
	a.initAudioDevice()
}

// MARK: オーディオデバイスの初期化メソッド
func (a *APU) initAudioDevice() {
	spec := &sdl.AudioSpec{
		Freq:     SAMPLE_RATE,
		Format:   sdl.AUDIO_F32,
		Channels: 1,
		Samples:  2048,
		Callback: sdl.AudioCallback(C.AudioMixCallback),
	}

	if err := sdl.OpenAudio(spec, nil); err != nil {
		panic(err)
	}

	// オーディオ再生開始
	sdl.PauseAudio(false)
}

// MARK: SDLのオーディオコールバック
//
//export AudioMixCallback
func AudioMixCallback(userdata unsafe.Pointer, stream *C.uint8_t, length C.int) {
	n := int(length) / 4
	buffer := unsafe.Slice((*float32)(unsafe.Pointer(stream)), n)

	ch1 := make([]float32, BUFFER_SIZE)[:n]
	ch2 := make([]float32, BUFFER_SIZE)[:n]
	ch3 := make([]float32, BUFFER_SIZE)[:n]
	ch4 := make([]float32, BUFFER_SIZE)[:n]

	square1.buffer.Read(ch1, n)
	square2.buffer.Read(ch2, n)
	triangle.buffer.Read(ch3, n)
	noise.buffer.Read(ch4, n)

	for i := range n {
		// @FIXME mixのバランス
		mixed := (ch1[i] + ch2[i] + ch3[i] + ch4[i]) / 25
		// mixed := (ch1[i] + ch2[i]) / 25
		// mixed := mixSamples(ch1[i], ch2[i], ch3[i], ch4[i], float32(0))

		if mixed > MAX_VOLUME {
			mixed = MAX_VOLUME
		} else if mixed < -MAX_VOLUME {
			mixed = -MAX_VOLUME
		}
		buffer[i] = mixed
	}
}

// MARK: APUのサイクルを進める
func (a *APU) Tick(cycles uint) {
	a.cycles += cycles
	a.sampleClock += uint64(cycles)
	a.clockFrameSequencer()

	// 現在のレベルを計算
	currentLevel1 := a.channel1.output(cycles)
	currentLevel2 := a.channel2.output(cycles)
	currentLevel3 := a.channel3.output(cycles)
	currentLevel4 := a.channel4.output(cycles)

	// 前回レベルとの差分を計算
	delta1 := currentLevel1 - a.prevLevel1
	delta2 := currentLevel2 - a.prevLevel2
	delta3 := currentLevel3 - a.prevLevel3
	delta4 := currentLevel4 - a.prevLevel4

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
}

// MARK: ステータスレジスタの読み込みメソッド
func (a *APU) ReadStatus() uint8 {
	status := a.status.ToByte()
	status &= 0xF0
	// @TODO: 各チャンネルの長さカウントを反映させる
	a.status.ClearFrameIRQ()
	return status
}

// MARK: ステータスレジスタの書き込みメソッド
func (a *APU) WriteStatus(data uint8) {
	prev := a.status.ToByte()
	a.status.update(data)

	// @TODO: ミュートと長さカウンタのリセットも行う
	/*
		有効ビットがクリアされると（ $4015経由）、長さカウンタは強制的に0に設定され、有効ビットが再度セットされるまで変更できなくなります（長さカウンタの以前の値は失われます）。有効ビットをセットしても、すぐには効果はありません。
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
	a.status.ClearFrameIRQ()
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
		if a.status.is1chEnabled() {
			a.channel1.lengthCounter.update(
				a.channel1.register.keyOffCount,
				a.channel1.register.LengthCounterHalt(),
			)
			a.channel1.lengthCounter.reload()
			a.channel1.envelope.reset()
			a.channel1.sweepUnit.reset()
			a.channel1.phase = 0
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
		if a.status.is2chEnabled() {
			a.channel2.lengthCounter.update(
				a.channel2.register.keyOffCount,
				a.channel2.register.LengthCounterHalt(),
			)
			a.channel2.lengthCounter.reload()
			a.channel2.envelope.reset()
			a.channel2.sweepUnit.reset()
			a.channel2.phase = 0
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
		a.channel3.lengthCounter.reload()
		a.channel3.linearCounter.setReload()
		a.channel3.phase = 0
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
		if a.status.is4chEnabled() {
			a.channel4.lengthCounter.update(
				a.channel4.register.keyOffCount,
				a.channel4.register.LengthCounterHalt(),
			)
			a.channel4.lengthCounter.reload()
			a.channel4.envelope.reset()
			a.channel4.phase = 0
		}
	}
}

// MARK: バッファのフラッシュ
func (a *APU) EndFrame() {
	// フレームの終わりまでの時間を処理するため、現在のクロックを渡す
	a.channel1.buffer.endFrame(a.sampleClock)
	a.channel2.buffer.endFrame(a.sampleClock)
	a.channel3.buffer.endFrame(a.sampleClock)
	a.channel4.buffer.endFrame(a.sampleClock)
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

// MARK: 各チャンネルのサンプルをMixする関数
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
	pulseOut := 95.88 / ((8128 / (pulse1 + pulse2)) + 100)
	tndOut := 159.79 / ((1/(triangle/8227) + (noise / 12241) + (dmc / 22638)) + 100)
	return pulseOut + tndOut
}
