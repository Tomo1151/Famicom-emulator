package apu

var LENGTH_COUNTER_TABLE = [32]uint8{
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

type EnvelopeData struct {
	rate    uint8
	enabled bool
	loop    bool
}

func (ed *EnvelopeData) Init(rate uint8, enabled bool, loop bool) {
	ed.rate = rate
	ed.enabled = enabled
	ed.loop = loop
}

type Envelope struct {
	data EnvelopeData

	counter uint8
	divider uint8
}

func (e *Envelope) Init() {
	e.data = EnvelopeData{}
	e.data.Init(0, false, false)
	e.counter = 0x0F
	e.divider = e.data.rate + 1
}

func (e *Envelope) volume() float32 {
	if e.data.enabled {
		return float32(e.counter) / 15.0
	} else {
		return float32(e.data.rate) / 15.0
	}
}

func (e *Envelope) reset() {
	e.counter = 0x0F
	e.divider = e.data.rate + 1
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
		if e.data.loop {
			e.reset()
		}
	}
	e.divider = e.data.rate + 1
}

type SweepUnitData struct {
	frequency  uint16
	shift      uint8
	direction  uint8
	timerCount uint8
	enabled    bool
}

func (sud *SweepUnitData) Init(frequency uint16, shift uint8, direction uint8, timerCount uint8, enabled bool) {
	sud.frequency = frequency
	sud.direction = direction
	sud.timerCount = timerCount
	sud.enabled = enabled
}

type SweepUnit struct {
	data      SweepUnitData
	frequency uint16
	counter   uint8
}

func (su *SweepUnit) Init() {
	su.data = SweepUnitData{}
	su.data.Init(0, 0, 0, 0, false)
	su.frequency = 0
	su.counter = 0
}

func (su *SweepUnit) getFrequency() float32 {
	if su.frequency == 0 {
		return 0.0
	}
	return CPU_CLOCK / (16.0 * (float32(su.frequency) + 1.0))
}

func (su *SweepUnit) reset() {
	su.frequency = su.data.frequency
	su.counter = 0
}

func (su *SweepUnit) tick(lengthCounter *LengthCounter) {
	if !su.data.enabled || su.data.shift == 0 || lengthCounter.isMuted() {
		return
	}

	su.counter++

	if su.counter < su.data.timerCount+1 {
		return
	}

	su.counter = 0

	if su.data.direction == 0 { // 上
		su.frequency = su.frequency + (su.frequency >> uint16(su.data.shift))
	} else { // 下
		su.frequency = su.frequency - (su.frequency >> uint16(su.data.shift))
	}

	if su.frequency < 8 || su.frequency >= 0x7FF {
		su.frequency = 0
	}
}

type LinearCounterData struct {
	count uint8
}

type LinearCounter struct {
	data    LinearCounterData
	counter uint8
}

func (lc *LinearCounter) Init() {
	lc.data = LinearCounterData{count: 0}
	lc.counter = 0
}

func (lc *LinearCounter) tick() {
	if lc.counter > 0 {
		lc.counter--
	}
}

func (lc *LinearCounter) isMuted() bool {
	return lc.counter == 0
}

func (lc *LinearCounter) reset() {
	lc.counter = lc.data.count
}

type LengthCounterData struct {
	count   uint8
	enabled bool
}

func (lcd *LengthCounterData) Init(count uint8, enabled bool) {
	lcd.count = LENGTH_COUNTER_TABLE[count]
	lcd.enabled = enabled
}

type LengthCounter struct {
	data    LengthCounterData
	counter uint8
}

func (lc *LengthCounter) Init() {
	lc.data = LengthCounterData{}
	lc.data.Init(0, false)
	lc.counter = 0
}

func (lc *LengthCounter) isMuted() bool {
	return lc.data.enabled && lc.counter == 0
}

func (lc *LengthCounter) reset() {
	lc.counter = lc.data.count
}

func (lc *LengthCounter) tick() {
	if !lc.data.enabled {
		return
	}

	if lc.counter > 0 {
		lc.counter--
	}
}
