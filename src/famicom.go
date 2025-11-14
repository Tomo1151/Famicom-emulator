package main

import (
	"Famicom-emulator/apu"
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

// MARK: Famicomの定義
type Famicom struct {
	cpu       cpu.CPU
	ppu       ppu.PPU
	apu       apu.TAPU
	joypad1   joypad.JoyPad
	joypad2   joypad.JoyPad
	bus       bus.Bus
	cartridge cartridge.Cartridge

	keyboard1   InputState // 1Pの入力状態 (キーボード)
	keyboard2   InputState // 2Pの入力状態 (キーボード)
	controller1 InputState // 1Pの入力状態 (コントローラ)
	controller2 InputState // 2Pの入力状態 (コントローラ)

	gamepad1 sdl.JoystickID // SDLのコントローラID (1P)
	gamepad2 sdl.JoystickID // SDLのコントローラID (2P)
}

// MARK: Famicomの初期化メソッド
func (f *Famicom) Init(cartridge cartridge.Cartridge) {
	// ROMファイルのロード
	f.cartridge = cartridge
	err := f.cartridge.Load()
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}

	// 各コンポーネントの定義 / 接続
	f.cpu = cpu.CPU{}
	f.ppu = ppu.PPU{}
	f.apu = apu.TAPU{}
	f.joypad1 = joypad.JoyPad{}
	f.joypad2 = joypad.JoyPad{}
	f.bus = bus.Bus{}
	f.bus.ConnectComponents(
		&f.ppu,
		&f.apu,
		&f.cartridge,
		&f.joypad1,
		&f.joypad2,
	)

	// 入力データの定義
	f.keyboard1 = InputState{}
	f.keyboard2 = InputState{}
	f.controller1 = InputState{}
	f.controller2 = InputState{}
}

// MARK: Famicomの起動
func (f *Famicom) Start() {
	// SDLの初期化
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMECONTROLLER); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// ウィンドウの作成
	window, err := sdl.CreateWindow("Famicom emu", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(ppu.SCREEN_WIDTH)*SCALE_FACTOR, int32(ppu.SCREEN_HEIGHT)*SCALE_FACTOR, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	// 接続済みコントローラを検知
	var gamepad1, gamepad2 *sdl.GameController
	if sdl.NumJoysticks() == 0 {
		fmt.Println("No controller detected")
	}
	if sdl.NumJoysticks() > 0 {
		gamepad1 = sdl.GameControllerOpen(0)
		if gamepad1 != nil {
			f.gamepad1 = gamepad1.Joystick().InstanceID()
			fmt.Println("Controller opened for 1P:", gamepad1.Name())
			defer gamepad1.Close()
		}
	}
	if sdl.NumJoysticks() > 1 {
		gamepad2 = sdl.GameControllerOpen(1)
		if gamepad2 != nil {
			f.gamepad2 = gamepad2.Joystick().InstanceID()
			fmt.Println("Controller opened for 2P:", gamepad2.Name())
			defer gamepad2.Close()
		}
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

	// SDL2イベントポンプを取得
	eventPump := sdl.PollEvent

	// 1フレーム目の時間を取得
	var lastFrameTime = time.Now()

	// BusのNMIコールバックで描画とイベント処理
	f.bus.Init(func(p *ppu.PPU, c *ppu.Canvas, j1 *joypad.JoyPad, j2 *joypad.JoyPad) {

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
				f.bus.Shutdown()
				os.Exit(0)
			case *sdl.KeyboardEvent:
				if e.Keysym.Sym == sdl.K_ESCAPE && e.State == sdl.PRESSED {
					f.bus.Shutdown()
					os.Exit(0)
				}
				if e.Keysym.Sym == sdl.K_F12 && e.State == sdl.PRESSED {
					f.cpu.Log = !f.cpu.Log
				}
				f.handleKeyPress(e, &f.keyboard1, &f.keyboard2)
			case *sdl.ControllerButtonEvent:
				switch e.Which {
				case f.gamepad1:
					f.handleButtonPress(e, &f.controller1)
				case f.gamepad2:
					f.handleButtonPress(e, &f.controller2)
				}
			case *sdl.ControllerAxisEvent:
				switch e.Which {
				case f.gamepad1:
					f.handleAxisMotion(e, &f.controller1)
				case f.gamepad2:
					f.handleAxisMotion(e, &f.controller2)
				}
			}

			// 操作結果を反映
			f.updateJoyPad(j1, &f.keyboard1, &f.controller1)
			f.updateJoyPad(j2, &f.keyboard2, &f.controller2)
		}
	})

	// CPUの初期化
	f.cpu.Init(f.bus, false)

	// CPUの起動
	f.cpu.Run()
}

// MARK: キーボードの状態を検知
func (f *Famicom) handleKeyPress(e *sdl.KeyboardEvent, c1 *InputState, c2 *InputState) {
	pressed := e.State == sdl.PRESSED
	switch e.Keysym.Sym {
	// 1P
	case sdl.K_k:
		c1.A = pressed
	case sdl.K_j:
		c1.B = pressed
	case sdl.K_w:
		c1.Up = pressed
	case sdl.K_s:
		c1.Down = pressed
	case sdl.K_a:
		c1.Left = pressed
	case sdl.K_d:
		c1.Right = pressed
	case sdl.K_RETURN, sdl.K_KP_ENTER:
		c1.Start = pressed
	case sdl.K_BACKSPACE:
		c1.Select = pressed

	// 2P
	case sdl.K_COLON:
		c2.A = pressed
	case sdl.K_SEMICOLON:
		c2.B = pressed
	case sdl.K_t:
		c2.Up = pressed
	case sdl.K_g:
		c2.Down = pressed
	case sdl.K_f:
		c2.Left = pressed
	case sdl.K_h:
		c2.Right = pressed
	case sdl.K_GREATER:
		c2.Select = pressed
	case sdl.K_TAB:
		c2.Start = pressed
	}
}

// MARK: コントローラーのボタン状態を検知
func (f *Famicom) handleButtonPress(e *sdl.ControllerButtonEvent, c *InputState) {
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
func (f *Famicom) handleAxisMotion(e *sdl.ControllerAxisEvent, c *InputState) {
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
func (f *Famicom) updateJoyPad(j *joypad.JoyPad, k *InputState, c *InputState) {
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
