package apu

const (
	DMC_WAVE_ENABLED = iota
	DMC_WAVE_NOTE
	DMC_WAVE_CHANNEL
	DMC_WAVE_RESET
)

type DMCWaveEventType uint

type DMCWave struct {
	freq    float32
	phase   float32
	channel chan DMCWaveEvent
	sender  chan ChannelEvent
	note    DMCNote
	buffer  *RingBuffer
	enabled bool
}

type DMCNote struct{}

type DMCWaveEvent struct {
	eventType DMCWaveEventType
	note      *DMCNote
	enabled   bool
	changed   bool
}

var dpcmWave DMCWave

func (dw *DMCWave) generatePCM() {
	for {
		// チャンネルから新しい音符を受信
	eventLoop:
		for {
			select {
			case event := <-dw.channel:
				switch event.eventType {
				case DMC_WAVE_ENABLED: // ENABLEDイベント
					dw.enabled = event.enabled
				case DMC_WAVE_NOTE: // NOTEイベント
					if event.note != nil {
						dw.note = *event.note
						dw.phase = 0.0
					}
				case DMC_WAVE_RESET:
				}
			default:
				break eventLoop
			}
		}

		if dw.buffer.Available() > BUFFER_SIZE/2 {
			continue
		}
	}
}
