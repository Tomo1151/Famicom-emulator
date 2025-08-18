package apu

type Envelope struct {
	rate    uint8
	enabled bool
	loop    bool
	counter uint8
	divider uint8
}

func (e *Envelope) Init(rate uint8, enabled bool, loop bool) {
	e.rate = rate
	e.enabled = enabled
	e.loop = loop
	e.counter = 0x0F
	e.divider = rate + 1
}

func (e *Envelope) volume() float32 {
	if e.enabled {
		return float32(e.counter) / 15.0
	} else {
		return float32(e.rate) / 15.0
	}
}

func (e *Envelope) tick() {
	e.divider--

	if e.divider != 0 {
		return
	}

	// 分周器が励起(divider == 0)したら
	if e.counter != 0 {
		e.counter--
	} else {
		if e.loop {
			e.counter = 0x0F
		}
	}
	e.divider = e.rate + 1
}

type LengthCounter struct {
	counter uint8
	enabled bool
}

func (lc *LengthCounter) setCount(count uint8) {
	lengthCounterTable := [32]uint8{
		0x05,
		0x7F,
		0x0A,
		0x01,
		0x14,
		0x02,
		0x28,
		0x03,
		0x50,
		0x04,
		0x1E,
		0x05,
		0x07,
		0x06,
		0x0D,
		0x07,
		0x06,
		0x08,
		0x0C,
		0x09,
		0x18,
		0x0A,
		0x30,
		0x0B,
		0x60,
		0x0C,
		0x24,
		0x0D,
		0x08,
		0x0E,
		0x10,
		0x0F,
	}
	lc.counter = lengthCounterTable[count]
}

func (lc *LengthCounter) tick() {
	if !lc.enabled {
		return
	}

	if lc.counter > 0 {
		lc.counter--
	}
}

func (lc *LengthCounter) isMuted() bool {
	return lc.enabled && lc.counter == 0
}
