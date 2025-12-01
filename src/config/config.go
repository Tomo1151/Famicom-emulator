package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数宣言
const (
	CONFIG_FILE_PATH = "./config.json"
)

// MARK: デフォルトのキーコンフィグ
var DefaultControl = ControlConfig{
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
		BUTTON_A:      sdl.K_SLASH,
		BUTTON_B:      sdl.K_PERIOD,
		BUTTON_UP:     sdl.K_g,
		BUTTON_DOWN:   sdl.K_b,
		BUTTON_RIGHT:  sdl.K_n,
		BUTTON_LEFT:   sdl.K_v,
		BUTTON_START:  sdl.K_KP_ENTER,
		BUTTON_SELECT: sdl.K_BACKSPACE,
	},
	GamepadAxisThreshold: 8000,
}

// MARK: Configの定義
type Config struct {
	CPU     CpuConfig     `json:"cpu"`
	APU     ApuConfig     `json:"apu"`
	Render  RenderConfig  `json:"render"`
	Control ControlConfig `json:"control"`
}

// MARK: ApuConfigの定義
type ApuConfig struct {
	SOUND_VOLUME float32 `json:"volume"`
	LOG_ENABLED  bool    `json:"log"`
	MUTE_1CH     bool    `json:"mute1ch"`
	MUTE_2CH     bool    `json:"mute2ch"`
	MUTE_3CH     bool    `json:"mute3ch"`
	MUTE_4CH     bool    `json:"mute4ch"`
	MUTE_5CH     bool    `json:"mute5ch"`
}

// MARK: CpuConfigの定義
type CpuConfig struct {
	LOG_ENABLED bool `json:"log"`
}

// MARK: RenderConfigの定義
type RenderConfig struct {
	SCALE_FACTOR             int  `json:"scale"`
	DOUBLE_BUFFERING_ENABLED bool `json:"doubleBuffering"`
}

// MARK: ControllerConfigの定義
type ControlConfig struct {
	KEY_1P               KeyConfig `json:"key1p"`
	KEY_2P               KeyConfig `json:"key2p"`
	GamepadAxisThreshold int16     `json:"gamepadAxisThreshold"`
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

// MARK: コンフィグファイルの読み込み
func LoadFromFile() *Config {
	// コンフィグファイルの読み込み
	configfile, err := os.ReadFile(CONFIG_FILE_PATH)
	if len(configfile) != 0 && err == nil {
		fmt.Println("Config file loaded.")
	} else {
		fmt.Println("Config file was not found.")
	}

	// ファイルからConfig構造体へパース
	config, err := ParseFromJson(configfile)
	if err != nil {
		fmt.Println("Config file contains invalid field.")

		// パースに失敗した場合は最低限の設定で起動
		config = &Config{
			Render: RenderConfig{
				SCALE_FACTOR: 3,
			},
			Control: DefaultControl,
		}
	}

	return config
}

// MARK: JSONからConfig構造体へ変換
func ParseFromJson(file []byte) (*Config, error) {
	var config Config
	err := json.Unmarshal(file, &config)

	// キーコンフィグだけデフォルトから設定
	config.Control.KEY_1P = DefaultControl.KEY_1P
	config.Control.KEY_2P = DefaultControl.KEY_2P

	return &config, err
}
