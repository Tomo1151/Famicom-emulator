package joypad

import "github.com/veandco/go-sdl2/sdl"

// MARK: 定数定義
// 対応コントローラ名
const (
	JOY_CON_R        = "Joy-Con (R)"
	JOY_CON_L        = "Joy-Con (L)"
	HVC_CONTROLLER_1 = "Nintendo HVC Controller (1)"
	HVC_CONTROLLER_2 = "Nintendo HVC Controller (2)"
)

// Joy-Con (R) 向けボタンマッピング
const (
	JOYCON_R_BUTTON_X    uint8 = 0
	JOYCON_R_BUTTON_A    uint8 = 1
	JOYCON_R_BUTTON_Y    uint8 = 2
	JOYCON_R_BUTTON_B    uint8 = 3
	JOYCON_R_BUTTON_HOME uint8 = 5
	JOYCON_R_BUTTON_PLUS uint8 = 6
	JOYCON_R_BUTTON_R    uint8 = 16
	JOYCON_R_BUTTON_ZR   uint8 = 16
)

// Nintendo HVC Controller 向けマッピング
const (
	HVC_BUTTON_A      = 0
	HVC_BUTTON_B      = 1
	HVC_BUTTON_UP     = 11
	HVC_BUTTON_DOWN   = 12
	HVC_BUTTON_LEFT   = 13
	HVC_BUTTON_RIGHT  = 14
	HVC_BUTTON_START  = 6
	HVC_BUTTON_SELECT = 4
	HVC_BUTTON_L      = 9
	HVC_BUTTON_R      = 10
)

// MARK: コントローラアダプタ
type JoyPadAdapter struct {
	name string
}

// MARK: コントローラアダプタの初期化メソッド
func (a *JoyPadAdapter) Init(name string) {
	a.name = name
}

// MARK: Aボタンのアクセサ
func (a *JoyPadAdapter) ButtonA() uint8 {
	switch a.name {
	case JOY_CON_L, JOY_CON_R:
		return JOYCON_R_BUTTON_A
	case HVC_CONTROLLER_1, HVC_CONTROLLER_2:
		return HVC_BUTTON_A
	default:
		return sdl.CONTROLLER_BUTTON_A
	}
}

// MARK: Bボタンのアクセサ
func (a *JoyPadAdapter) ButtonB() uint8 {
	switch a.name {
	case JOY_CON_L, JOY_CON_R:
		return JOYCON_R_BUTTON_B
	case HVC_CONTROLLER_1, HVC_CONTROLLER_2:
		return HVC_BUTTON_B
	default:
		return sdl.CONTROLLER_BUTTON_B
	}
}

// MARK: 十字キー上のアクセサ
func (a *JoyPadAdapter) ButtonUp() uint8 {
	switch a.name {
	case HVC_CONTROLLER_1, HVC_CONTROLLER_2:
		return HVC_BUTTON_UP
	default:
		return sdl.CONTROLLER_BUTTON_DPAD_UP
	}
}

// MARK: 十字キー下のアクセサ
func (a *JoyPadAdapter) ButtonDown() uint8 {
	switch a.name {
	case HVC_CONTROLLER_1, HVC_CONTROLLER_2:
		return HVC_BUTTON_DOWN
	default:
		return sdl.CONTROLLER_BUTTON_DPAD_DOWN
	}
}

// MARK: 十字キー右のアクセサ
func (a *JoyPadAdapter) ButtonRight() uint8 {
	switch a.name {
	case HVC_CONTROLLER_1, HVC_CONTROLLER_2:
		return HVC_BUTTON_RIGHT
	default:
		return sdl.CONTROLLER_BUTTON_DPAD_RIGHT
	}
}

// MARK: 十字キー左のアクセサ
func (a *JoyPadAdapter) ButtonLeft() uint8 {
	switch a.name {
	case HVC_CONTROLLER_1, HVC_CONTROLLER_2:
		return HVC_BUTTON_LEFT
	default:
		return sdl.CONTROLLER_BUTTON_DPAD_LEFT
	}
}

// MARK: スタートボタンのアクセサ
func (a *JoyPadAdapter) ButtonStart() uint8 {
	switch a.name {
	case JOY_CON_L, JOY_CON_R:
		return JOYCON_R_BUTTON_PLUS
	case HVC_CONTROLLER_1:
		return HVC_BUTTON_START
	case HVC_CONTROLLER_2:
		// HVC Controller 2P にはスタートボタンがないためRボタンを割り当て
		return HVC_BUTTON_R
	default:
		return sdl.CONTROLLER_BUTTON_START
	}
}

// MARK: セレクトボタンのアクセサ
func (a *JoyPadAdapter) ButtonSelect() uint8 {
	switch a.name {
	case JOY_CON_L, JOY_CON_R:
		return JOYCON_R_BUTTON_HOME
	case HVC_CONTROLLER_1:
		return HVC_BUTTON_SELECT
	case HVC_CONTROLLER_2:
		// HVC Controller 2P にはセレクトボタンがないためLボタンを割り当て
		return HVC_BUTTON_L
	default:
		return sdl.CONTROLLER_BUTTON_GUIDE
	}
}
