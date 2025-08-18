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