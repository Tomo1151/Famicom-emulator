package apu

// MARK: 矩形波データの構造体
type NoiseWave struct {
	freq       float32
	phase      float32
	channel    chan NoiseNote
	note       NoiseNote
	noise 			bool
	buffer *RingBuffer

	longNoise NoiseShiftRegister
	shortNoise NoiseShiftRegister
}

// MARK: 可変部分の構造体
type NoiseNote struct {
	hz     float32
	volume float32
	noiseMode NoiseRegisterMode
}

// MARK: 矩形波データ
var noiseWave NoiseWave

// MARK: PCM波形生成のgoroutine
func (nw *NoiseWave) generatePCM() {
	for {
		// チャンネルから新しい音符を受信
		select {
		case note := <-nw.channel:
			nw.note = note
			nw.phase = 0.0 // 音符が変わったらphaseをリセット
		default:
			// 新しい音符がない場合は現在の音符を継続
		}

		// バッファに十分なデータがある場合は少し待つ
		if nw.buffer.Available() > BUFFER_SIZE/2 {
			continue
		}

		// 小さなバッファでPCMサンプルを生成
		const chunkSize = 512
		pcmBuffer := make([]float32, chunkSize)
		
		// 現在の音符の周波数に基づいてphaseIncrementを計算
		phaseIncrement := float32(nw.note.hz) / float32(sampleHz)

		for i := range chunkSize {
			nw.phase += phaseIncrement
			if nw.phase >= 1.0 {
				nw.phase -= 1.0

				// ノイズシフトレジスタを更新
				switch nw.note.noiseMode {
				case NOISE_MODE_LONG:
					nw.noise = nw.longNoise.next()
				case NOISE_MODE_SHORT:
					nw.noise = nw.shortNoise.next()
				}
			}

			var sample float32
			if nw.note.volume > 0 { // ボリュームチェックを追加
				if nw.noise {
					sample = 0.0 // ノイズがtrueの場合は無音
				} else {
					sample = MAX_VOLUME * nw.note.volume
				}
			} else {
				sample = 0.0
			}

			pcmBuffer[i] = sample
		}

		// リングバッファに書き込み
		nw.buffer.Write(pcmBuffer)
	}
}