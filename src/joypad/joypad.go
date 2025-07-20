package joypad

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

type JoyPadButton uint8

type JoyPad struct {
	strobe bool
	ButtonIndex uint8
	State uint8
}

func (j *JoyPad) Init() {
	j.strobe = false
	j.ButtonIndex = 0
	j.State = 0
}

func (j *JoyPad) Write(data uint8) {
	j.strobe = data & 1 == 1
	
	// Reset index when strobe is enabled (NES behavior)
	if j.strobe {
		j.ButtonIndex = 0
	}
}

func (j *JoyPad) Read() uint8 {
	var response uint8
	
	if j.ButtonIndex >= 8 {
		return 0x01 // 8番目以降は常に1を返す
	}
	
	// 現在のボタンの状態を取得
	response = (j.State >> j.ButtonIndex) & 1
	
	// ストローブが無効の場合のみインデックスを進める
	if !j.strobe {
		j.ButtonIndex++
	}
	
	return response
}

func (j *JoyPad) SetButtonPressed(button JoyPadButton, pressed bool) {
	if pressed {
		j.State |= (1 << button)
	} else {
		j.State &^= (1 << button)
	}
}