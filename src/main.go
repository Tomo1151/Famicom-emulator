package main

import (
	"Famicom-emulator/bus"
	"Famicom-emulator/cartridge"
	"Famicom-emulator/cpu"
	"Famicom-emulator/joypad"
	"Famicom-emulator/ppu"
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数定義
const (
	FRAME_PER_SECOND = 60
	SCALE_FACTOR     = 3

	INPUT_MODE_KEYBOARD = 0
	INPUT_MODE_JOYPAD   = 1
)

// MARK: InputStateの定義
type InputState struct {
	Left, Right, Up, Down bool
	A, B, Start, Select   bool
}

// MARK: main関数
func main() {
	// ROMファイルのロード

	// filedata, err := os.ReadFile("../rom/Kirby'sAdventure.nes")
	// filedata, err := os.ReadFile("../rom/SuperMarioBros.nes")
	filedata, err := os.ReadFile("../rom/SuperMarioBros3.nes")
	if err != nil {
		log.Fatalf("Error occured in 'os.ReadFile()'")
	}

	// カートリッジの作成と初期化
	cart := cartridge.Cartridge{}
	err = cart.Load(filedata)
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}

	// SDLの初期化
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMECONTROLLER); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// ウィンドウの作成
	window, err := sdl.CreateWindow("Famicom", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(ppu.SCREEN_WIDTH)*SCALE_FACTOR, int32(ppu.SCREEN_HEIGHT)*SCALE_FACTOR, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	// 接続済みコントローラを検知
	controller := sdl.GameControllerOpen(0)
	if controller == nil {
		fmt.Println("No controller detected")
	} else {
		fmt.Println("Controller opened:", controller.Name())
		defer controller.Close()
	}

	// レンダラーの作成
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	// テクスチャの作成
	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_RGB24,
		sdl.TEXTUREACCESS_STREAMING,
		int32(ppu.SCREEN_WIDTH), int32(ppu.SCREEN_HEIGHT))
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	// 入力データ
	keyboardState := InputState{}
	controllerState := InputState{}

	// SDL2イベントポンプを取得
	eventPump := sdl.PollEvent

	var lastFrameTime = time.Now()

	// BusのNMIコールバックで描画とイベント処理
	bus := bus.Bus{}
	bus.InitWithCartridge(&cart, func(p *ppu.PPU, c *ppu.Canvas, j0 *joypad.JoyPad, j1 *joypad.JoyPad) {

		// フレームレート調整 (60FPS)
		now := time.Now()
		elapsed := now.Sub(lastFrameTime)
		const frameDuration = time.Second / FRAME_PER_SECOND
		if elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
		lastFrameTime = time.Now()

		// キャンバスのバッファを元にテクスチャの更新
		texture.Update(nil, unsafe.Pointer(&c.Buffer[0]), int(c.Width*3))

		// 再レンダリング
		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		// SDLイベント処理
		for event := eventPump(); event != nil; event = eventPump() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			case *sdl.KeyboardEvent:
				if e.Keysym.Sym == sdl.K_ESCAPE && e.State == sdl.PRESSED {
					os.Exit(0)
				}
				handleKeyPress(e, &keyboardState)
			case *sdl.ControllerButtonEvent:
				handleButtonPress(e, &controllerState)
			case *sdl.ControllerAxisEvent:
				handleAxisMotion(e, &controllerState)
			}

			// 操作結果を反映
			updateJoyPad(j0, &keyboardState, &controllerState)
		}
	})

	c := cpu.CPU{}
	c.InitWithCartridge(bus, true)
	c.Run()
}

// MARK: キーボードの状態を検知
func handleKeyPress(e *sdl.KeyboardEvent, c *InputState) {
	pressed := e.State == sdl.PRESSED
	switch e.Keysym.Sym {
	case sdl.K_k:
		c.A = pressed
	case sdl.K_j:
		c.B = pressed
	case sdl.K_w:
		c.Up = pressed
	case sdl.K_s:
		c.Down = pressed
	case sdl.K_a:
		c.Left = pressed
	case sdl.K_d:
		c.Right = pressed
	case sdl.K_RETURN, sdl.K_KP_ENTER:
		c.Start = pressed
	case sdl.K_BACKSPACE:
		c.Select = pressed
	}
}

// MARK: コントローラーのボタン状態を検知
func handleButtonPress(e *sdl.ControllerButtonEvent, c *InputState) {
	pressed := e.State == sdl.PRESSED
	switch e.Button {
	case joypad.JOYCON_R_BUTTON_A, joypad.JOYCON_R_BUTTON_X:
		c.A = pressed
	case joypad.JOYCON_R_BUTTON_B, joypad.JOYCON_R_BUTTON_Y:
		c.B = pressed
	case joypad.JOYCON_R_BUTTON_PLUS:
		c.Start = pressed
	case joypad.JOYCON_R_BUTTON_HOME:
		c.Select = pressed
	}
}

// MARK: コントローラーのスティック状態を検知
func handleAxisMotion(e *sdl.ControllerAxisEvent, c *InputState) {
	const threshold = 8000 // デッドゾーン
	switch e.Axis {
	case 0: // X軸 (左スティック左右)
		if e.Value < -threshold {
			c.Left = true
			c.Right = false
		} else if e.Value > threshold {
			c.Left = false
			c.Right = true
		} else {
			c.Left = false
			c.Right = false
		}
	case 1: // Y軸 (左スティック上下)
		if e.Value < -threshold {
			c.Up = true
			c.Down = false
		} else if e.Value > threshold {
			c.Up = false
			c.Down = true
		} else {
			c.Up = false
			c.Down = false
		}
	}
}

// MARK: JoyPadの状態を更新
func updateJoyPad(j *joypad.JoyPad, k *InputState, c *InputState) {
	// キーボードとコントローラの入力を統合
	buttonA := k.A || c.A
	buttonB := k.B || c.B
	buttonStart := k.Start || c.Start
	buttonSelect := k.Select || c.Select
	buttonUp := k.Up || c.Up
	buttonDown := k.Down || c.Down
	buttonLeft := k.Left || c.Left
	buttonRight := k.Right || c.Right

	// JoyPadの状態を更新
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_A_POSITION, buttonA)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_B_POSITION, buttonB)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_START_POSITION, buttonStart)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_SELECT_POSITION, buttonSelect)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_UP_POSITION, buttonUp)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_DOWN_POSITION, buttonDown)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_LEFT_POSITION, buttonLeft)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_RIGHT_POSITION, buttonRight)
}
