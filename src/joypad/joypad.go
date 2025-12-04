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
}

// MARK: JoyPadの初期化メソッド
func (j *JoyPad) Init() {
	j.strobe = false
	j.ButtonIndex = 0
	j.State = 0
}

// MARK: JoyPadへの書き込み
func (j *JoyPad) Write(data uint8) {
	j.strobe = data&1 == 1
	if j.strobe {
		j.ButtonIndex = 0
	}
}

// MARK: JoyPadの読み取り
func (j *JoyPad) Read() uint8 {
	if j.ButtonIndex > 7 {
		return 0x01
	}
	response := (j.State & (1 << j.ButtonIndex)) >> j.ButtonIndex

	if !j.strobe && j.ButtonIndex <= 7 {
		j.ButtonIndex++
	}

	return response
}

// MARK: ボタン押下状態のセット
func (j *JoyPad) SetButtonPressed(buttonIndex JoyPadButton, pressed bool) {
	if pressed {
		j.State |= (1 << buttonIndex)
	} else {
		j.State &^= (1 << buttonIndex)
	}
}
