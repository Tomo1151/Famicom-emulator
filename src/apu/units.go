package apu

// MARK: 変数定義
var (
	LENGTH_COUNTER_TABLE = [32]uint8{
		0x0A, 0xFE, 0x14, 0x02, 0x28, 0x04, 0x50, 0x06,
		0xA0, 0x08, 0x3C, 0x0A, 0x0E, 0x0C, 0x1A, 0x0E,
		0x0C, 0x10, 0x18, 0x12, 0x30, 0x14, 0x60, 0x16,
		0xC0, 0x18, 0x48, 0x1A, 0x10, 0x1C, 0x20, 0x1E,
	}
)

// MARK: エンベロープの定義
type Envelope struct {
	counter uint8
	divider uint8
	rate    uint8
	enabled bool
	loop    bool
}

// MARK: エンベロープの初期化メソッド
func (e *Envelope) Init() {
	e.counter = 0x0F
	e.rate = 0
	e.divider = e.rate + 1
	e.loop = false
	e.enabled = false
}

// MARK: エンベロープからボリュームを取得するメソッド
func (e *Envelope) Volume() float32 {
	if e.enabled {
		return float32(e.counter)
	} else {
		return float32(e.rate)
	}
}

// MARK: エンベロープをリセットするメソッド
func (e *Envelope) reset() {
	e.counter = 0x0F
	e.divider = e.rate + 1
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
		if e.loop {
			e.reset()
		}
	}
	e.divider = e.rate + 1
}

// MARK: エンベロープの更新メソッド
func (e *Envelope) update(rate uint8, loop bool, enabled bool) {
	e.rate = rate
	e.loop = loop
	e.enabled = enabled
}

// MARK: スイープの定義
type SweepUnit struct {
	frequency  uint16
	counter    uint8
	mute       bool
	reload     bool
	shift      uint8
	direction  uint8
	timerCount uint8
	enabled    bool
}

// MARK: スイープの初期化メソッド
func (su *SweepUnit) Init() {
	su.frequency = 0
	su.counter = 0
	su.mute = true
	su.reload = false
	su.shift = 0
	su.direction = 0
	su.timerCount = 0
	su.enabled = false
}

// MARK: スイープのリセット
func (su *SweepUnit) reset() {
	su.counter = 0
	su.mute = false
	su.reload = true
}

// MARK: スイープユニットによるチャンネルの無効化
func (su *SweepUnit) isMuted() bool {
	return su.mute
}

// MARK: スイープユニットの更新メソッド
func (su *SweepUnit) update(shift uint8, direction uint8, period uint8, enabled bool) {
	su.shift = shift
	su.direction = direction
	su.timerCount = period
	su.enabled = enabled
	su.reload = true
}

// MARK: スイープのサイクルを進めるメソッド
func (su *SweepUnit) tick(lengthCounter *LengthCounter, isNot bool) {
	if su.counter > 0 {
		su.counter--
	}

	if su.reload || su.counter == 0 {
		su.counter = su.timerCount + 1
		if su.reload {
			su.reload = false
		}
	} else {
		return
	}

	if !su.enabled || su.shift == 0 || lengthCounter.isMuted() || su.frequency < 8 {
		return
	}

	var target uint16
	diff := su.frequency >> uint16(su.shift)

	if su.direction == 0 {
		target = su.frequency + diff
	} else {
		if isNot {
			target = su.frequency - (diff + 1)
		} else {
			target = su.frequency - diff
		}
	}

	if target > 0x7FF || target < 8 {
		su.mute = true
		return
	}

	su.frequency = target
	su.mute = false
}

// MARK: 線形カウンタの定義
type LinearCounter struct {
	counter uint8
	reload  bool
	count   uint8
	enabled bool
}

// MARK: 線形カウンタの初期化メソッド
func (lc *LinearCounter) Init() {
	lc.counter = 0
	lc.reload = false
	lc.count = 0
	lc.enabled = false
}

// MARK: 線形カウンタのサイクルを進めるメソッド
func (lc *LinearCounter) tick() {
	if lc.reload {
		lc.counter = lc.count
	} else if lc.counter > 0 {
		lc.counter--
	}

	if !lc.enabled {
		lc.reload = false
	}
}

// MARK: 線形カウンタが終了したかを取得するメソッド
func (lc *LinearCounter) isMuted() bool {
	return lc.counter == 0
}

// MARK: 線型カウンタをリロードするメソッド
func (lc *LinearCounter) setReload() {
	lc.reload = true
}

// MARK: 線型カウンタの更新メソッド
func (lc *LinearCounter) update(count uint8, enabled bool) {
	lc.count = count
	lc.enabled = enabled
}

// MARK: 長さカウンタの定義
type LengthCounter struct {
	counter uint8
	count   uint8
	enabled bool
}

// MARK: 長さカウンタの初期化メソッド
func (lc *LengthCounter) Init() {
	lc.counter = 0
	lc.count = 0
	lc.enabled = false
}

// MARK: 長さカウンタが終了したかを取得するメソッド
func (lc *LengthCounter) isMuted() bool {
	return lc.counter == 0
}

// MARK: 長さカウンタをリセットするメソッド
func (lc *LengthCounter) reload() {
	lc.counter = lc.count
}

// MARK: 長さカウンタの更新メソッド
func (lc *LengthCounter) update(count uint8, enabled bool) {
	lc.count = LENGTH_COUNTER_TABLE[count]
	lc.enabled = enabled
}

// MARK: 長さカウンタのサイクルを進めるメソッド
func (lc *LengthCounter) tick() {
	if lc.enabled {
		return
	}

	if lc.counter > 0 {
		lc.counter--
	}
}
