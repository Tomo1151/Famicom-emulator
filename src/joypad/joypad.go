package joypad

// MARK: 定数定義
const (
	JOYPAD_BUTTON_A_POSITION JoyPadButton = iota
	JOYPAD_BUTTON_B_POSITION
	JOYPAD_BUTTON_SELECT_POSITION
	JOYPAD_BUTTON_START_POSITION
	JOYPAD_BUTTON_UP_POSITION
	JOYPAD_BUTTON_DOWN_POSITION
	JOYPAD_BUTTON_LEFT_POSITION
	JOYPAD_BUTTON_RIGHT_POSITION
)

// MARK: JoyPadButtonの定義
type JoyPadButton uint8

// MARK: JoyPadの定義
type JoyPad struct {
	strobe      bool
	ButtonIndex uint8
	State       uint8

	latchedState uint8 // $4016ストローブでラッチされたスナップショット
}

// MARK: JoyPadの初期化メソッド
func (j *JoyPad) Init() {
	j.strobe = false
	j.ButtonIndex = 0
	j.State = 0
	j.latchedState = 0
}

// MARK: JoyPadへの書き込み
func (j *JoyPad) Write(data uint8) {
	prevStrobe := j.strobe
	j.strobe = data&1 == 1

	// ストローブがhighになった時点で状態をラッチ
	if j.strobe {
		j.ButtonIndex = 0
		j.latchedState = j.State
		return
	}

	// ダウンエッジでシフト読み取り開始
	if prevStrobe && !j.strobe {
		j.ButtonIndex = 0
		j.latchedState = j.State
	}
}

// MARK: JoyPadの読み取り
func (j *JoyPad) Read() uint8 {
	// bit 6 を立てておく
	const openBus uint8 = 0x40

	// ストローブHigh: Aボタン(bit0)を返し続ける（インデックスは進めない）
	if j.strobe {
		return openBus | (j.State & 0x01)
	}

	// ストローブLow: ラッチした8bitを順番に返す
	var bit uint8
	if j.ButtonIndex < 8 {
		bit = (j.latchedState >> j.ButtonIndex) & 0x01
	} else {
		// 8回読み切った後は1を返し続ける
		bit = 0x01
	}

	j.ButtonIndex++
	return openBus | bit
}

// MARK: ボタン押下状態のセット
func (j *JoyPad) SetButtonPressed(buttonIndex JoyPadButton, pressed bool) {
	if pressed {
		j.State |= uint8(1) << uint8(buttonIndex)
	} else {
		j.State &^= uint8(1) << uint8(buttonIndex)
	}
}
