package apu

import "time"

const (
	TRIANGLE_WAVE_ENABLED = iota
	TRIANGLE_WAVE_NOTE
	TRIANGLE_WAVE_LENGTH_COUNTER
	TRIANGLE_WAVE_LENGTH_COUNTER_TICK
	TRIANGLE_WAVE_LINEAR_COUNTER
	TRIANGLE_WAVE_LINEAR_COUNTER_TICK
	TRIANGLE_WAVE_CHANNEL
	TRIANGLE_WAVE_RESET
)

type TriangleWaveEventType uint

// MARK: 矩形波データの構造体
type TriangleWave struct {
	freq          float32
	phase         float32
	channel       chan TriangleWaveEvent
	sender        chan ChannelEvent
	note          TriangleNote
	lengthCounter LengthCounter
	linearCounter LinearCounter
	buffer        *RingBuffer
	enabled       bool
}

// MARK: 可変部分の構造体
type TriangleNote struct {
	hz float32
}

type TriangleWaveEvent struct {
	eventType         TriangleWaveEventType
	note              *TriangleNote
	lengthCounterData *LengthCounterData
	linearCounterData *LinearCounterData
	enabled           bool
	changed           bool
}

// MARK: 三角波データ
var triangleWave TriangleWave

// MARK: PCM波形生成のgoroutine
func (tw *TriangleWave) generatePCM() {
	pcmBuffer := make([]float32, CHUNK_SIZE)

	for {
		// チャンネルから新しい音符を受信
	eventLoop:
		for {
			select {
			case event := <-tw.channel:
				switch event.eventType {
				case TRIANGLE_WAVE_ENABLED: // ENABLEDイベント
					tw.enabled = event.enabled

					// disableされた時に長さカウンタも落とす (halt)
					if event.changed {
						tw.lengthCounter.counter = 0
					}
				case TRIANGLE_WAVE_NOTE: // NOTEイベント
					if event.note != nil {
						tw.note = *event.note
					}
				case TRIANGLE_WAVE_LINEAR_COUNTER: // LINEAR COUNTERイベント
					if event.linearCounterData != nil {
						tw.linearCounter.data = *event.linearCounterData
						tw.linearCounter.reset()
					}
				case TRIANGLE_WAVE_LENGTH_COUNTER: // LENGTH COUNTERイベント
					if event.lengthCounterData != nil {
						tw.lengthCounter.data = *event.lengthCounterData
					}
				case TRIANGLE_WAVE_LENGTH_COUNTER_TICK: // LENGTH COUNTER TICKイベント
					tw.lengthCounter.tick()
					tw.sender <- ChannelEvent{
						length: tw.lengthCounter.counter,
					}
				case TRIANGLE_WAVE_LINEAR_COUNTER_TICK: // LINEAR COUNTER TICKイベント
					tw.linearCounter.tick()
				case TRIANGLE_WAVE_RESET: // RESETイベント
					tw.lengthCounter.reload()
					tw.linearCounter.reset()
					tw.phase = 0.0 // 音符が変わったらphaseをリセット
				}
			default:
				// 新しい音符がない場合は現在の音符を継続
				break eventLoop
			}
		}

		// バッファに十分なデータがある場合は少し待つ
		if tw.buffer.Available() > BUFFER_SIZE/2 {
			time.Sleep(10 * time.Nanosecond)
		}

		for i := range pcmBuffer {
			pcmBuffer[i] = 0.0
		}

		// 現在の音符の周波数に基づいてphaseIncrementを計算
		phaseIncrement := float32(tw.note.hz) / float32(sampleHz)

		for i := range CHUNK_SIZE {
			var sample float32
			if tw.phase < 0.5 {
				sample = tw.phase // 上がっていく部分
			} else {
				sample = 1.0 - tw.phase // 下がっていく部分
			}

			if tw.enabled && !tw.linearCounter.isMuted() && !tw.lengthCounter.isMuted() {
				pcmBuffer[i] = (sample - 0.25) * 4 * MAX_VOLUME // 真ん中へずらす, ボリュームは固定
			} else {
				pcmBuffer[i] = 0.0
			}

			tw.phase += phaseIncrement
			if tw.phase >= 1.0 {
				tw.phase -= 1.0
			}
		}

		// リングバッファに書き込み
		tw.buffer.Write(pcmBuffer)
	}
}
