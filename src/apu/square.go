package apu

// MARK: 変数定義
var (
	SQUARE_DUTY_TABLE = [4][8]uint8{
		{0, 1, 0, 0, 0, 0, 0, 0}, // 12.5%
		{0, 1, 1, 0, 0, 0, 0, 0}, // 25.0%
		{0, 1, 1, 1, 1, 0, 0, 0}, // 50.0%
		{1, 0, 0, 1, 1, 1, 1, 1}, // 75.0%
	}
)

// MARK: 矩形波チャンネルの定義
type SquareWaveChannel struct {
	register      SquareWaveRegister // @FIXME レジスタはAPUに持たせ、ここは参照にする
	envelope      Envelope
	lengthCounter LengthCounter
	sweepUnit     SweepUnit
	duty          uint8
	timerReload   uint16
	timer         uint16
	sequencer     uint8
	buffer        BlipBuffer
}

// MARK: 矩形波チャンネルの初期化メソッド
func (swc *SquareWaveChannel) Init(log bool) {
	swc.register = SquareWaveRegister{}
	swc.register.Init()
	swc.envelope = Envelope{}
	swc.envelope.Init()
	swc.lengthCounter = LengthCounter{}
	swc.lengthCounter.Init()
	swc.sweepUnit = SweepUnit{}
	swc.sweepUnit.Init()
	swc.buffer.Init(log)
}

// MARK: 矩形波チャンネルの出力メソッド
func (swc *SquareWaveChannel) output(cycles uint) float32 {
	frequency := swc.sweepUnit.frequency
	if frequency < 8 || frequency > 0x7FF || swc.lengthCounter.isMuted() || swc.sweepUnit.isMuted() {
		return 0.0
	}

	swc.timerReload = (uint16(frequency) + 1) * 2
	if swc.timer == 0 {
		swc.timer = swc.timerReload
	}

	cyclesLeft := uint(cycles)
	for cyclesLeft > 0 {
		if swc.timer > uint16(cyclesLeft) {
			swc.timer -= uint16(cyclesLeft)
			break
		}
		cyclesLeft -= uint(swc.timer)
		swc.timer = swc.timerReload
		swc.sequencer = (swc.sequencer + 1) & 7
	}

	if frequency < 8 || frequency > 0x7FF || swc.lengthCounter.isMuted() || swc.sweepUnit.isMuted() {
		// ミュートの時は0.0を返す
		return 0.0
	}

	// デューティテーブルを参照
	if SQUARE_DUTY_TABLE[swc.duty][swc.sequencer] == 1 {
		return swc.envelope.Volume()
	} else {
		return 0.0
	}
}

// MARK: デバッグ出力切り替え
func (s *SquareWaveChannel) ToggleLog() {
	s.buffer.ToggleLog()
}
