package apu

import "time"

const (
	SQUARE_WAVE_ENABLED = iota
	SQUARE_WAVE_NOTE
	SQUARE_WAVE_ENVELOPE
	SQUARE_WAVE_ENVELOPE_TICK
	SQUARE_WAVE_LENGTH_COUNTER
	SQUARE_WAVE_LENGTH_COUNTER_TICK
	SQUARE_WAVE_SWEEP
	SQUARE_WAVE_SWEEP_FREQUENCY
	SQUARE_WAVE_SWEEP_TICK
	SQUARE_WAVE_CHANNEL
	SQUARE_WAVE_RESET
)

type SquareWaveEventType uint

// MARK: 矩形波データの構造体
type SquareWave struct {
	channelNumber uint
	freq          float32
	phase         float32
	channel       chan SquareWaveEvent
	sender        chan ChannelEvent
	note          SquareNote
	envelope      Envelope
	lengthCounter LengthCounter
	sweepUnit     SweepUnit
	buffer        *RingBuffer
	enabled       bool
}

// MARK: 可変部分の構造体
type SquareNote struct {
	duty float32
}

type SquareWaveEvent struct {
	eventType         SquareWaveEventType
	note              *SquareNote
	envelopeData      *EnvelopeData
	lengthCounterData *LengthCounterData
	sweepUnitData     *SweepUnitData
	frequency         *uint16
	enabled           bool
}

// MARK: 矩形波データ
var squareWave1 SquareWave
var squareWave2 SquareWave

// MARK: PCM波形生成のgoroutine
func (sw *SquareWave) generatePCM() {
	// バッファを事前確保
	pcmBuffer := make([]float32, CHUNK_SIZE)

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
					}
				case SQUARE_WAVE_ENVELOPE: // ENVELOPEイベント
					if event.envelopeData != nil {
						sw.envelope.data = *event.envelopeData
					}
				case SQUARE_WAVE_ENVELOPE_TICK: // ENVELOPE TICKイベント
					sw.envelope.tick()
				case SQUARE_WAVE_LENGTH_COUNTER: // LENGTH COUNTERイベント
					if event.lengthCounterData != nil {
						sw.lengthCounter.data = *event.lengthCounterData
					}
				case SQUARE_WAVE_LENGTH_COUNTER_TICK: // LENGTH COUNTER TICKイベント
					sw.lengthCounter.tick()
					sw.sender <- ChannelEvent{
						length: sw.lengthCounter.counter,
					}
				case SQUARE_WAVE_SWEEP: // SWEEPイベント
					if event.sweepUnitData != nil {
						sw.sweepUnit.data = *event.sweepUnitData
					}
				case SQUARE_WAVE_SWEEP_FREQUENCY: // SWEEP FREQUENCYイベント
					if event.frequency != nil {
						sw.sweepUnit.frequency = *event.frequency
					}
				case SQUARE_WAVE_SWEEP_TICK: // SWEEP TICKイベント
					sw.sweepUnit.tick(&sw.lengthCounter, *(&sw.channelNumber) == 1)
				case SQUARE_WAVE_RESET: // RESETイベント
					sw.envelope.reset()
					sw.lengthCounter.reset()
					sw.sweepUnit.reset()
					sw.phase = 0.0 // 音符が変わったらphaseをリセット
				}
			default:
				// 新しい音符がない場合は現在の音符を継続
				break eventLoop
			}
		}

		// バッファに十分なデータがある場合は少し待つ
		if sw.buffer.Available() > BUFFER_SIZE/2 {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		// pcmBufferを再利用
		for i := range pcmBuffer {
			pcmBuffer[i] = 0.0
		}

		// 現在の音符の周波数に基づいてphaseIncrementを計算
		frequency := sw.sweepUnit.getFrequency()
		phaseIncrement := frequency / float32(sampleHz)

		for i := range CHUNK_SIZE {
			var sample float32

			if sw.phase <= sw.note.duty {
				sample = MAX_VOLUME // 正の波形
			} else {
				sample = -MAX_VOLUME // 負の波形
			}

			if !sw.enabled || sw.lengthCounter.isMuted() || sw.sweepUnit.isMuted() {
				sample = 0.0
			}

			pcmBuffer[i] = sample * sw.envelope.volume()

			if sw.sweepUnit.frequency != 0.0 {
				sw.phase += phaseIncrement
				if sw.phase >= 1.0 {
					sw.phase -= 1.0
				}
			}
		}

		// リングバッファに書き込み
		sw.buffer.Write(pcmBuffer)
	}
}