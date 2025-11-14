package apu

type TriangleWaveChannel struct {
	register      TriangleWaveRegister
	lengthCounter LengthCounter
	linearCounter LinearCounter
	frequency     uint16
	phase         float32
	buffer        BlipBuffer
}

func (twc *TriangleWaveChannel) Init() {
	twc.register = TriangleWaveRegister{}
	twc.register.Init()
	twc.lengthCounter = LengthCounter{}
	twc.lengthCounter.Init()
	twc.linearCounter = LinearCounter{}
	twc.linearCounter.Init()
	twc.buffer.Init()
}

func (twc *TriangleWaveChannel) output(cycles uint) float32 {
	if twc.lengthCounter.isMuted() || twc.linearCounter.isMuted() || twc.frequency < 2 {
		return 0.0
	}

	period := float32((twc.frequency + 1) * 32)
	twc.phase += float32(cycles) / period

	if twc.phase >= 1.0 {
		// 0.0 ~ 1.0 の範囲に制限
		twc.phase -= 1.0
	}

	// -1.0 から 1.0 の範囲で線形に変化する三角波を生成
	var value float32
	if twc.phase < 0.5 {
		// 0.0 -> 0.5 の区間で 1.0 -> -1.0 に変化
		value = 1.0 - 4.0*twc.phase
	} else {
		// 0.5 -> 1.0 の区間で -1.0 -> 1.0 に変化
		value = -1.0 + 4.0*(twc.phase-0.5)
	}

	return value
}
