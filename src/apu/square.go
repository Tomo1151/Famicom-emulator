package apu

// MARK: 矩形波データの構造体
type SquareWave struct {
	freq       float32
	phase      float32
	channel    chan SquareNote
	note       SquareNote
	buffer *RingBuffer
}

// MARK: 可変部分の構造体
type SquareNote struct {
	hz     float32
	volume float32
	duty   float32
}

// MARK: 矩形波データ
var squareWave1 SquareWave
var squareWave2 SquareWave

// MARK: PCM波形生成のgoroutine
func (sw *SquareWave) generatePCM() {
	for {
		// チャンネルから新しい音符を受信
		select {
		case note := <-sw.channel:
			sw.note = note
			sw.phase = 0.0 // 音符が変わったらphaseをリセット
		default:
			// 新しい音符がない場合は現在の音符を継続
		}

		// バッファに十分なデータがある場合は少し待つ
		if sw.buffer.Available() > BUFFER_SIZE/2 {
			continue
		}

		// 小さなバッファでPCMサンプルを生成
		const chunkSize = 512 // チャンクサイズを大きくしてアンダーランを防ぐ
		pcmBuffer := make([]float32, chunkSize)
		
		// 現在の音符の周波数に基づいてphaseIncrementを計算
		phaseIncrement := float32(sw.note.hz) / float32(sampleHz)

		for i := range chunkSize {
			sw.phase += phaseIncrement
			if sw.phase >= 1.0 {
				sw.phase -= 1.0
			}

			var sample float32
			if sw.note.volume > 0 { // ボリュームが0より大きい場合のみ音を出す
				if sw.phase < sw.note.duty {
					sample = MAX_VOLUME // 正の波形
				} else {
					sample = -MAX_VOLUME // 負の波形
				}
			} else {
				sample = 0.0 // 無音（中央値）
			}
			pcmBuffer[i] = sample * sw.note.volume
		}

		// リングバッファに書き込み
		sw.buffer.Write(pcmBuffer)
	}
}