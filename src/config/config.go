package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数宣言
const (
	CONFIG_FILE_PATH = "./config.json"
)

// MARK: デフォルトのキーコンフィグ
var DefaultControl = ControlConfig{
	KEY_1P: SDLKeyConfig{
		BUTTON_A:      sdl.K_k,
		BUTTON_B:      sdl.K_j,
		BUTTON_UP:     sdl.K_w,
		BUTTON_DOWN:   sdl.K_s,
		BUTTON_RIGHT:  sdl.K_d,
		BUTTON_LEFT:   sdl.K_a,
		BUTTON_START:  sdl.K_RETURN,
		BUTTON_SELECT: sdl.K_BACKSPACE,
	},
	KEY_2P: SDLKeyConfig{
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
	Cpu     CpuConfig     `json:"cpu"`
	Ppu     PpuConfig     `json:"ppu"`
	Apu     ApuConfig     `json:"apu"`
	Render  RenderConfig  `json:"render"`
	Rom     RomConfig     `json:"rom"`
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

// MARK: PpuConfigの定義
type PpuConfig struct {
	BACKGROUND_ENABLED bool `json:"background"`
	SPRITE_ENABLED     bool `json:"sprite"`
}

// MARK: RenderConfigの定義
type RenderConfig struct {
	SCALE_FACTOR             int  `json:"scale"`
	DOUBLE_BUFFERING_ENABLED bool `json:"doubleBuffering"`
	FULLSCREEN               bool `json:"fullscreen"`
}

// MARK: RomConfigの定義
type RomConfig struct {
	AUTO_LOADING bool `json:"autoLoading"`
}

// MARK: ControllerConfigの定義
type ControlConfig struct {
	RAW_KEY_1P           KeyConfig `json:"key1p"`
	RAW_KEY_2P           KeyConfig `json:"key2p"`
	KEY_1P               SDLKeyConfig
	KEY_2P               SDLKeyConfig
	GamepadAxisThreshold int16 `json:"gamepadAxisThreshold"`
}

// MARK: KeyConfigの定義
type SDLKeyConfig struct {
	BUTTON_A      sdl.Keycode
	BUTTON_B      sdl.Keycode
	BUTTON_UP     sdl.Keycode
	BUTTON_DOWN   sdl.Keycode
	BUTTON_RIGHT  sdl.Keycode
	BUTTON_LEFT   sdl.Keycode
	BUTTON_START  sdl.Keycode
	BUTTON_SELECT sdl.Keycode
}

// MARK: KeyConfigの定義
type KeyConfig struct {
	BUTTON_A      string `json:"buttonA"`
	BUTTON_B      string `json:"buttonB"`
	BUTTON_UP     string `json:"buttonUp"`
	BUTTON_DOWN   string `json:"buttonDown"`
	BUTTON_RIGHT  string `json:"buttonRight"`
	BUTTON_LEFT   string `json:"buttonLeft"`
	BUTTON_START  string `json:"buttonStart"`
	BUTTON_SELECT string `json:"buttonSelect"`
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
		config = Config{
			Render: RenderConfig{
				SCALE_FACTOR: 3,
			},
			Control: DefaultControl,
		}
	}

	return &config
}

// MARK: JSONからConfig構造体へ変換
func ParseFromJson(file []byte) (Config, error) {
	var config Config
	err := json.Unmarshal(file, &config)

	// キーコンフィグをstringからsdl.KeyCodeに変換して登録
	config.Control.KEY_1P = MapKeyConfig(config.Control.RAW_KEY_1P)
	config.Control.KEY_2P = MapKeyConfig(config.Control.RAW_KEY_2P)

	return config, err
}

// MARK: KeyConfigをSDLKeyConfigに変換
func MapKeyConfig(raw KeyConfig) SDLKeyConfig {
	var out SDLKeyConfig
	rvOut := reflect.ValueOf(&out).Elem()
	rvIn := reflect.ValueOf(raw)

	for i := 0; i < rvIn.NumField(); i++ {
		fieldName := rvIn.Type().Field(i).Name
		rawValue := rvIn.Field(i).String()
		outField := rvOut.FieldByName(fieldName)

		if outField.IsValid() && outField.CanSet() {
			outField.SetInt(int64(sdl.GetKeyFromName(rawValue)))
		}
	}

	return out
}
