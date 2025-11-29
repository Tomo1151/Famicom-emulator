package config

import "github.com/veandco/go-sdl2/sdl"

var DefaultConfig = &Config{
	Render: RenderConfig{
		SCALE_FACTOR:             3,
		DOUBLE_BUFFERING_ENABLED: true,
	},
	APU: ApuConfig{
		SOUND_VOLUME: 1.0,
		LOG_ENABLED:  false,
		MUTE_1CH:     false,
		MUTE_2CH:     false,
		MUTE_3CH:     false,
		MUTE_4CH:     false,
		MUTE_5CH:     false,
	},
	CPU: CpuConfig{
		LOG_ENABLED: false,
	},
	Control: ControlConfig{
		KEY_1P: KeyConfig{
			BUTTON_A:      sdl.K_k,
			BUTTON_B:      sdl.K_j,
			BUTTON_UP:     sdl.K_w,
			BUTTON_DOWN:   sdl.K_s,
			BUTTON_RIGHT:  sdl.K_d,
			BUTTON_LEFT:   sdl.K_a,
			BUTTON_START:  sdl.K_RETURN,
			BUTTON_SELECT: sdl.K_BACKSPACE,
		},
		KEY_2P: KeyConfig{
			BUTTON_A:      sdl.K_UNDERSCORE,
			BUTTON_B:      sdl.K_SLASH,
			BUTTON_UP:     sdl.K_g,
			BUTTON_DOWN:   sdl.K_b,
			BUTTON_RIGHT:  sdl.K_n,
			BUTTON_LEFT:   sdl.K_v,
			BUTTON_START:  sdl.K_KP_ENTER,
			BUTTON_SELECT: sdl.K_BACKSPACE,
		},
		GamepadAxisThreshold: 8000,
	},
}

// MARK: Configの定義
type Config struct {
	CPU     CpuConfig
	APU     ApuConfig
	Render  RenderConfig
	Control ControlConfig
}

// MARK: ApuConfigの定義
type ApuConfig struct {
	SOUND_VOLUME float32
	LOG_ENABLED  bool
	MUTE_1CH     bool
	MUTE_2CH     bool
	MUTE_3CH     bool
	MUTE_4CH     bool
	MUTE_5CH     bool
}

// MARK: CpuConfigの定義
type CpuConfig struct {
	LOG_ENABLED bool
}

// MARK: RenderConfigの定義
type RenderConfig struct {
	SCALE_FACTOR             int
	DOUBLE_BUFFERING_ENABLED bool
}

// MARK: ControllerConfigの定義
type ControlConfig struct {
	KEY_1P               KeyConfig
	KEY_2P               KeyConfig
	GamepadAxisThreshold int16
}

// MARK: KeyConfigの定義
type KeyConfig struct {
	BUTTON_A      sdl.Keycode
	BUTTON_B      sdl.Keycode
	BUTTON_UP     sdl.Keycode
	BUTTON_DOWN   sdl.Keycode
	BUTTON_RIGHT  sdl.Keycode
	BUTTON_LEFT   sdl.Keycode
	BUTTON_START  sdl.Keycode
	BUTTON_SELECT sdl.Keycode
}
