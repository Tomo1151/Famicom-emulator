package apu

// MARK: 矩形波チャンネルの定義
type SquareWaveChannel struct {
	register      SquareWaveRegister // @FIXME レジスタはAPUに持たせ、ここは参照にする
	envelope      Envelope
	lengthCounter LengthCounter
	sweepUnit     SweepUnit
	duty          float32
	phase         float32
	buffer        BlipBuffer
}

// MARK: 矩形波チャンネルの初期化メソッド
func (swc *SquareWaveChannel) Init() {
	swc.register = SquareWaveRegister{}
	swc.register.Init()
	swc.envelope = Envelope{}
	swc.envelope.Init()
	swc.lengthCounter = LengthCounter{}
	swc.lengthCounter.Init()
	swc.sweepUnit = SweepUnit{}
	swc.sweepUnit.Init()
	swc.buffer.Init()
}

// MARK: 矩形波チャンネルの出力メソッド
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
	if swc.phase <= swc.duty {
		value = 1.0
	} else {
		value = -1.0
	}

	return float32(value) * swc.envelope.volume()
}
