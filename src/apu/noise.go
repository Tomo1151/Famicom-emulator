package apu

const (
	NOISE_WAVE_ENABLED = iota
	NOISE_WAVE_NOTE
	NOISE_WAVE_ENVELOPE
	NOISE_WAVE_ENVELOPE_TICK
	NOISE_WAVE_LENGTH_COUNTER
	NOISE_WAVE_LENGTH_COUNTER_TICK
	NOISE_WAVE_CHANNEL
	NOISE_WAVE_RESET
)

type NoiseWaveEventType uint

// MARK: 矩形波データの構造体
type NoiseWave struct {
	freq          float32
	phase         float32
	channel       chan NoiseWaveEvent
	sender        chan ChannelEvent
	note          NoiseNote
	envelope      Envelope
	noise         bool
	lengthCounter LengthCounter
	buffer        *RingBuffer

	longNoise  NoiseShiftRegister
	shortNoise NoiseShiftRegister

	enabled bool
}

// MARK: 可変部分の構造体
type NoiseNote struct {
	hz        float32
	noiseMode NoiseRegisterMode
}

type NoiseWaveEvent struct {
	eventType         NoiseWaveEventType
	note              *NoiseNote
	envelopeData      *EnvelopeData
	lengthCounterData *LengthCounterData
	enabled           bool
}

// MARK: 矩形波データ
var noiseWave NoiseWave

// MARK: PCM波形生成のgoroutine
func (nw *NoiseWave) generatePCM() {
	for {
		// チャンネルから新しい音符を受信
	eventLoop:
		for {
			select {
			case event := <-nw.channel:
				switch event.eventType {
				case NOISE_WAVE_ENABLED:
					nw.enabled = event.enabled // ENABLEDイベント
				case NOISE_WAVE_NOTE: // NOTEイベント
					if event.note != nil {
						nw.note = *event.note
						nw.phase = 0.0 // 音符が変わったらphaseをリセット
					}
				case NOISE_WAVE_ENVELOPE: // ENVELOPEイベント
					if event.envelopeData != nil {
						nw.envelope.data = *event.envelopeData
					}
				case NOISE_WAVE_ENVELOPE_TICK: // ENVELOPE TICKイベント
					nw.envelope.tick()
				case NOISE_WAVE_LENGTH_COUNTER: // LENGTH COUNTER TICKイベント
					if event.lengthCounterData != nil {
						nw.lengthCounter.data = *event.lengthCounterData
					}
				case NOISE_WAVE_LENGTH_COUNTER_TICK: // LENGTH COUNTER TICKイベント
					nw.lengthCounter.tick()
					nw.sender <- ChannelEvent{
						length: nw.lengthCounter.counter,
					}
				case NOISE_WAVE_RESET: // RESETイベント
					nw.envelope.reset()
					nw.lengthCounter.reset()
				}
			default:
				// 新しい音符がない場合は現在の音符を継続
				break eventLoop
			}
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
			if nw.envelope.volume() > 0 { // ボリュームチェックを追加
				if nw.noise {
					sample = 0.0 // ノイズがtrueの場合は無音
				} else {
					sample = MAX_VOLUME * nw.envelope.volume()
				}
			} else {
				sample = 0.0
			}

			if !nw.enabled || nw.lengthCounter.isMuted() {
				sample = 0.0
			}

			pcmBuffer[i] = sample
		}

		// リングバッファに書き込み
		nw.buffer.Write(pcmBuffer)
	}
}