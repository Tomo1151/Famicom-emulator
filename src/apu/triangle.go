package apu

// MARK: 矩形波データの構造体
type TriangleWave struct {
	freq       float32
	phase      float32
	channel    chan TriangleNote
	note       TriangleNote
	buffer *RingBuffer
}

// MARK: 可変部分の構造体
type TriangleNote struct {
	hz     float32
}

// MARK: 三角波データ
var triangleWave TriangleWave

// MARK: PCM波形生成のgoroutine
func (tw *TriangleWave) generatePCM() {
	for {
		// チャンネルから新しい音符を受信
		select {
		case note := <-tw.channel:
			tw.note = note
			tw.phase = 0.0 // 音符が変わったらphaseをリセット
		default:
			// 新しい音符がない場合は現在の音符を継続
		}

		// バッファに十分なデータがある場合は少し待つ
		if tw.buffer.Available() > BUFFER_SIZE/2 {
			continue
		}

		// 小さなバッファでPCMサンプルを生成
		const chunkSize = 512 // チャンクサイズを大きくしてアンダーランを防ぐ
		pcmBuffer := make([]float32, chunkSize)

		// 現在の音符の周波数に基づいてphaseIncrementを計算
		phaseIncrement := float32(tw.note.hz) / float32(sampleHz)

		for i := range chunkSize {
			tw.phase += phaseIncrement
			if tw.phase >= 1.0 {
				tw.phase -= 1.0
			}

			var sample float32
			if tw.phase < 0.5 {
				sample = tw.phase // 上がっていく部分
			} else {
				sample = 1.0-tw.phase // 下がっていく
			}

			pcmBuffer[i] = (sample - 0.25) * 4 // 真ん中へずらす, ボリュームは固定
		}

		// リングバッファに書き込み
		tw.buffer.Write(pcmBuffer)
	}
}