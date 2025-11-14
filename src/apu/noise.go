package apu

// MARK: ノイズチャンネルの定義
type NoiseWaveChannel struct {
	register      NoiseWaveRegister
	envelope      Envelope
	lengthCounter LengthCounter
	mode          NoiseShiftMode
	shiftRegister NoiseShiftRegister
	prev          bool
	index         uint8
	phase         float32
	buffer        BlipBuffer
}

// MARK: ノイズチャンネルの初期化メソッド
func (nwc *NoiseWaveChannel) Init() {
	nwc.register = NoiseWaveRegister{}
	nwc.register.Init()
	nwc.envelope = Envelope{}
	nwc.envelope.Init()
	nwc.lengthCounter = LengthCounter{}
	nwc.lengthCounter.Init()
	nwc.mode = NOISE_MODE_SHORT
	nwc.shiftRegister = NoiseShiftRegister{}
	nwc.shiftRegister.InitWithShortMode()
	nwc.buffer = BlipBuffer{}
	nwc.buffer.Init()
}

// MARK: ノイズチャンネルの出力メソッド
func (nwc *NoiseWaveChannel) output(cycles uint) float32 {
	if nwc.lengthCounter.isMuted() {
		return 0.0
	}

	period := nwc.register.Frequency()
	nwc.phase += float32(cycles)

	if nwc.phase >= period {
		for nwc.phase >= period {
			nwc.prev = nwc.shiftRegister.next()
			nwc.phase -= period
		}
	}

	var value float32
	if !nwc.prev {
		value = MAX_VOLUME * float32(nwc.envelope.volume())
	} else {
		value = 0.0
	}

	return value
}
