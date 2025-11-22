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
	phase         float32
	buffer        BlipBuffer
}

// MARK: 三角波チャンネルの初期化メソッド
func (twc *TriangleWaveChannel) Init() {
	twc.register = TriangleWaveRegister{}
	twc.register.Init()
	twc.lengthCounter = LengthCounter{}
	twc.lengthCounter.Init()
	twc.linearCounter = LinearCounter{}
	twc.linearCounter.Init()
	twc.buffer.Init()
}

// MARK: 三角波チャンネルの出力メソッド
func (twc *TriangleWaveChannel) output(cycles uint) float32 {
	if twc.lengthCounter.isMuted() || twc.linearCounter.isMuted() || twc.frequency < 2 {
		return 0.0
	}

	// タイマー値から現在のシーケンス位置を計算
	period := (twc.frequency + 1)
	twc.phase += float32(cycles)

	// 32ステップのシーケンサ
	step := uint(twc.phase/float32(period)) % 32

	return float32(triangleSequence[step])
}
