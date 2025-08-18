package apu

const (
	SQUARE_WAVE_ENABLED = iota
	SQUARE_WAVE_NOTE
	SQUARE_WAVE_ENVELOPE
	SQUARE_WAVE_ENVELOPE_TICK
)

type SquareWaveEventType uint

// MARK: 矩形波データの構造体
type SquareWave struct {
	freq     float32
	phase    float32
	channel  chan SquareWaveEvent
	note     SquareNote
	envelope Envelope
	buffer   *RingBuffer
	enabled  bool
}

// MARK: 可変部分の構造体
type SquareNote struct {
	hz   float32
	duty float32
}

type SquareWaveEvent struct {
	eventType SquareWaveEventType
	note      *SquareNote
	envelope  *Envelope
	enabled   bool
}

// MARK: 矩形波データ
var squareWave1 SquareWave
var squareWave2 SquareWave

// MARK: PCM波形生成のgoroutine
func (sw *SquareWave) generatePCM() {
	for {
		// チャンネルから新しい音符を受信
	eventLoop:
		for {

			select {
			case event := <-sw.channel:
				switch event.eventType {
				case SQUARE_WAVE_ENABLED: // ENABLEDイベント
					sw.enabled = event.enabled
				case SQUARE_WAVE_NOTE: // NOTEイベント
					if event.note != nil {
						sw.note = *event.note
						sw.phase = 0.0 // 音符が変わったらphaseをリセット
					}
				case SQUARE_WAVE_ENVELOPE:
					if event.envelope != nil {
						sw.envelope = *event.envelope
					}
				case SQUARE_WAVE_ENVELOPE_TICK:
					if event.envelope != nil {
						sw.envelope.tick()
					}
				}
			default:
				// 新しい音符がない場合は現在の音符を継続
				break eventLoop
			}
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
			if sw.envelope.volume() > 0 { // ボリュームが0より大きい場合のみ音を出す
				if sw.phase < sw.note.duty {
					sample = MAX_VOLUME // 正の波形
				} else {
					sample = -MAX_VOLUME // 負の波形
				}
			} else {
				sample = 0.0 // 無音（中央値）
			}

			if !sw.enabled {
				sample = 0.0
			}
			pcmBuffer[i] = sample * sw.envelope.volume()
		}

		// リングバッファに書き込み
		sw.buffer.Write(pcmBuffer)
	}
}