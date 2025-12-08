package apu

// MARK: 変数定義
var (
	triangleSequence = [32]uint8{
		15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}
)

// MARK: 三角波チャンネルの定義
type TriangleWaveChannel struct {
	register      TriangleWaveRegister
	lengthCounter LengthCounter
	linearCounter LinearCounter
	frequency     uint16
	timer         float32
	sequenceIndex int
	buffer        BlipBuffer
}

// MARK: 三角波チャンネルの初期化メソッド
func (twc *TriangleWaveChannel) Init(log bool) {
	twc.register = TriangleWaveRegister{}
	twc.register.Init()
	twc.lengthCounter = LengthCounter{}
	twc.lengthCounter.Init()
	twc.linearCounter = LinearCounter{}
	twc.linearCounter.Init()
	twc.buffer.Init(log)
	twc.timer = 0
	twc.sequenceIndex = 0
}

// MARK: 三角波チャンネルの出力メソッド
func (twc *TriangleWaveChannel) output(cycles uint) float32 {
	// ミュート状態でも現在の値を出力し続ける（位相は進めない）
	if twc.lengthCounter.isMuted() || twc.linearCounter.isMuted() || twc.frequency < 2 {
		return float32(triangleSequence[twc.sequenceIndex])
	}

	// タイマーを進める
	period := float32(twc.frequency + 1)
	twc.timer -= float32(cycles)
	for twc.timer <= 0 {
		twc.timer += period
		twc.sequenceIndex = (twc.sequenceIndex + 1) % 32
	}

	return float32(triangleSequence[twc.sequenceIndex])
}

// MARK: デバッグ出力切り替え
func (t *TriangleWaveChannel) ToggleLog() {
	t.buffer.ToggleLog()
}
