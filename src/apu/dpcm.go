package apu

const (
	DPCM_WAVE_ENABLED = iota
	DPCM_WAVE_NOTE
	DPCM_WAVE_CHANNEL
	DPCM_WAVE_RESET
)

type DPCMWaveEventType uint

type DPCMWave struct {
	freq    float32
	phase   float32
	channel chan DPCMWaveEvent
	sender  chan ChannelEvent
	note    DPCMNote
	buffer  *RingBuffer
	enabled bool
}

type DPCMNote struct{}

type DPCMWaveEvent struct {
	eventType DPCMWaveEventType
	note      *DPCMNote
	enabled   bool
	changed   bool
}

var dpcmWave DPCMWave

func (dw *DPCMWave) generatePCM() {
	for {
		// チャンネルから新しい音符を受信
	eventLoop:
		for {
			select {
			case event := <-dw.channel:
				switch event.eventType {
				case DPCM_WAVE_ENABLED: // ENABLEDイベント
					dw.enabled = event.enabled
				case DPCM_WAVE_NOTE: // NOTEイベント
					if event.note != nil {
						dw.note = *event.note
						dw.phase = 0.0
					}
				case DPCM_WAVE_RESET:
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
