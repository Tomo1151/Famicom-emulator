package apu

// MARK: 長さカウンタのテーブル
var LENGTH_COUNTER_TABLE = [32]uint8{
	0x0A, 0xFE, 0x14, 0x02, 0x28, 0x04, 0x50, 0x06,
	0xA0, 0x08, 0x3C, 0x0A, 0x0E, 0x0C, 0x1A, 0x0E,
	0x0C, 0x10, 0x18, 0x12, 0x30, 0x14, 0x60, 0x16,
	0xC0, 0x18, 0x48, 0x1A, 0x10, 0x1C, 0x20, 0x1E,
}

// MARK: エンベロープの定義
type Envelope struct {
	data EnvelopeData

	counter uint8
	divider uint8
}

// MARK: エンベロープの初期化メソッド
func (e *Envelope) Init() {
	e.data = EnvelopeData{}
	e.data.Init(0, false, false)
	e.counter = 0x0F
	e.divider = e.data.rate + 1
}

// MARK: エンベロープからボリュームを取得するメソッド
func (e *Envelope) volume() float32 {
	if e.data.enabled {
		return float32(e.counter) / 15.0
	} else {
		return float32(e.data.rate) / 15.0
	}
}

// MARK: エンベロープをリセットするメソッド
func (e *Envelope) reset() {
	e.counter = 0x0F
	e.divider = e.data.rate + 1
}

// MARK: エンベロープのサイクルを進めるメソッド
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

// MARK: エンベロープの可変部分
type EnvelopeData struct {
	rate    uint8
	enabled bool
	loop    bool
}

// MARK: エンベロープの可変部分の初期化メソッド
func (ed *EnvelopeData) Init(rate uint8, enabled bool, loop bool) {
	ed.rate = rate
	ed.enabled = enabled
	ed.loop = loop
}

// MARK: スイープの定義
type SweepUnit struct {
	data      SweepUnitData
	frequency uint16
	counter   uint8
	mute      bool
}

// MARK: スイープの初期化メソッド
func (su *SweepUnit) Init() {
	su.data = SweepUnitData{}
	su.data.Init(0, 0, 0, false)
	su.frequency = 0
	su.counter = 0
	su.mute = true
}

// MARK: スイープの周波数を取得するメソッド
func (su *SweepUnit) Frequency() float32 {
	if su.frequency == 0 {
		return 0.0
	}
	return CPU_CLOCK / (16.0 * (float32(su.frequency) + 1.0))
}

// MARK: スイープのリセット
func (su *SweepUnit) reset() {
	su.counter = 0
	su.mute = false
}

// MARK: スイープユニットによるチャンネルの無効化
func (su *SweepUnit) isMuted() bool {
	return su.mute
}

// MARK: スイープのサイクルを進めるメソッド
func (su *SweepUnit) tick(lengthCounter *LengthCounter, isNot bool) {
	su.counter++

	if su.counter < su.data.timerCount+1 {
		return
	}

	su.counter = 0

	if !su.data.enabled || su.data.shift == 0 || lengthCounter.isMuted() {
		return
	}

	if su.data.direction == 0 { // 上
		su.frequency += (su.frequency >> su.data.shift)
	} else { // 下
		diff := su.frequency >> su.data.shift
		if isNot {
			// 1の補数を使用する(Ch1)場合
			su.frequency -= diff
		} else {
			// 2の補数を使用する(Ch2)場合
			su.frequency -= (diff + 1)
		}
	}

	su.mute = su.frequency < 0x08 || su.frequency > 0x7FF

	if su.frequency < 0x08 || su.frequency > 0x7FF {
		lengthCounter.counter = 0
	}
}

// MARK: スイープの可変部分
type SweepUnitData struct {
	shift      uint8
	direction  uint8
	timerCount uint8
	enabled    bool
}

// MARK: スイープの可変部分の初期化メソッド
func (sud *SweepUnitData) Init(shift uint8, direction uint8, timerCount uint8, enabled bool) {
	sud.shift = shift
	sud.direction = direction
	sud.timerCount = timerCount
	sud.enabled = enabled
}

// MARK: 線形カウンタの定義
type LinearCounter struct {
	data    LinearCounterData
	counter uint8
}

// MARK: 線形カウンタの初期化メソッド
func (lc *LinearCounter) Init() {
	lc.data = LinearCounterData{}
	lc.data.Init(0, false)
	lc.counter = 0
}

// MARK: 線形カウンタのサイクルを進めるメソッド
func (lc *LinearCounter) tick() {
	if !lc.data.enabled {
		return
	}

	if lc.counter > 0 {
		lc.counter--
	}
}

// MARK: 線形カウンタが終了したかを取得するメソッド
func (lc *LinearCounter) isMuted() bool {
	return lc.counter == 0
}

// MARK: 線形カウンタをリセットするメソッド
func (lc *LinearCounter) reset() {
	lc.counter = lc.data.count
}

// MARK: 線形カウンタの可変部分
type LinearCounterData struct {
	count   uint8
	enabled bool
}

// MARK: 線形カウンタの可変部分の初期化メソッド
func (lcd *LinearCounterData) Init(count uint8, enabled bool) {
	lcd.count = count
	lcd.enabled = enabled
}

// MARK: 長さカウンタの定義
type LengthCounter struct {
	data    LengthCounterData
	counter uint8
}

// MARK: 長さカウンタの初期化メソッド
func (lc *LengthCounter) Init() {
	lc.data = LengthCounterData{}
	lc.data.Init(0, false)
	lc.counter = 0
}

// MARK: 長さカウンタが終了したかを取得するメソッド
func (lc *LengthCounter) isMuted() bool {
	return lc.counter == 0
}

// MARK: 長さカウンタをリセットするメソッド
func (lc *LengthCounter) reset() {
	lc.counter = lc.data.count
}

// MARK: 長さカウンタのサイクルを進めるメソッド
func (lc *LengthCounter) tick() {
	if !lc.data.enabled {
		return
	}

	if lc.counter > 0 {
		lc.counter--
	}
}

// MARK: 長さカウンタの可変部分
type LengthCounterData struct {
	count   uint8
	enabled bool
}

// MARK: 長さカウンタの可変部分の初期化メソッド
func (lcd *LengthCounterData) Init(count uint8, enabled bool) {
	lcd.count = LENGTH_COUNTER_TABLE[count]
	lcd.enabled = enabled
}
