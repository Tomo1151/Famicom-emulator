package main

import (
	"Famicom-emulator/apu"
	"Famicom-emulator/bus"
	"Famicom-emulator/cartridge"
	"Famicom-emulator/config"
	"Famicom-emulator/cpu"
	"Famicom-emulator/joypad"
	"Famicom-emulator/ppu"
	"Famicom-emulator/ui"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数定義
const (
	FRAME_PER_SECOND = 60
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
	apu       apu.APU
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

	config  *config.Config
	windows *ui.WindowManager
}

// MARK: Famicomの初期化メソッド
func (f *Famicom) Init(cartridge cartridge.Cartridge, cfg *config.Config) {
	// ROMファイルのロード
	f.cartridge = cartridge
	err := f.cartridge.Load()
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}

	// 各コンポーネントの定義 / 接続
	f.cpu = cpu.CPU{}
	f.ppu = ppu.PPU{}
	f.apu = apu.APU{}
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

	if cfg != nil {
		f.config = cfg
	} else {
		f.config = &config.Config{
			SCALE_FACTOR: 3,
			SOUND_VOLUME: 1.0,
		}
	}

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

	// ウィンドウマネージャの作成
	f.windows = ui.NewWindowManager()
	defer f.windows.CloseAll()

	// ゲームコントローラの接続
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

	// 状態変数の定義
	eventPump := sdl.PollEvent
	lastFrameTime := time.Now()

	var optionWindowID uint32
	var optionWindowOpen bool

	var vramWindowID uint32
	var vramWindowOpen bool

	var chrWindowID uint32
	var chrWindowOpen bool

	var apuWindowID uint32
	var apuWindowOpen bool

	closeOptionWindow := func() {
		if !optionWindowOpen {
			return
		}
		if optionWindowID != 0 {
			f.windows.Remove(optionWindowID)
		}
		optionWindowOpen = false
		optionWindowID = 0
	}

	openOptionWindow := func() {
		if optionWindowOpen {
			return
		}
		optWin, err := ui.NewOptionWindow(f.config, func(uint32) {
			closeOptionWindow()
		})
		if err != nil {
			log.Printf("failed to open option window: %v", err)
			return
		}
		optionWindowOpen = true
		optionWindowID = optWin.ID()
		f.windows.Add(optWin)
	}

	closeVramWindow := func() {
		if !vramWindowOpen {
			return
		}
		if vramWindowID != 0 {
			f.windows.Remove(vramWindowID)
		}
		vramWindowOpen = false
		vramWindowID = 0
	}

	openVramWindow := func(p *ppu.PPU) {
		if vramWindowOpen {
			return
		}
		vwin, err := ui.NewNameTableWindow(p, f.config.SCALE_FACTOR, func(id uint32) {
			closeVramWindow()
		})
		if err != nil {
			log.Printf("failed to open Name Table window: %v", err)
			return
		}
		vramWindowOpen = true
		vramWindowID = vwin.ID()
		f.windows.Add(vwin)
	}

	openChrWindow := func(p *ppu.PPU) {
		cwin, err := ui.NewCharacterWindow(p, f.config.SCALE_FACTOR, func(id uint32) {
			// onClose
			if id != 0 {
				f.windows.Remove(id)
			}
			chrWindowOpen = false
			chrWindowID = 0
		})
		if err != nil {
			log.Printf("failed to open CHR window: %v", err)
		} else {
			chrWindowOpen = true
			chrWindowID = cwin.ID()
			f.windows.Add(cwin)
		}

	}

	closeChrWindow := func() {
		if !chrWindowOpen {
			return
		}
		if chrWindowID != 0 {
			f.windows.Remove(chrWindowID)
		}
		chrWindowOpen = false
		chrWindowID = 0
	}

	openAPUWindow := func() {
		if apuWindowOpen || f.windows == nil {
			return
		}

		aw, err := ui.NewAPUWindow(&f.apu, f.config.SCALE_FACTOR, func(id uint32) {
			if id != 0 {
				f.windows.Remove(id)
			}
			apuWindowOpen = false
			apuWindowID = 0
		})
		if err != nil {
			log.Printf("failed to open APU window: %v", err)
			return
		}
		apuWindowOpen = true
		apuWindowID = aw.ID()
		f.windows.Add(aw)
	}

	closeAPUWindow := func() {
		if !apuWindowOpen {
			return
		}
		if apuWindowID != 0 {
			f.windows.Remove(apuWindowID)
		}
		apuWindowOpen = false
		apuWindowID = 0
	}

	// Busの初期化とフレーム毎に実行されるコールバックの定義
	f.bus.Init(func(p *ppu.PPU, c *ppu.Canvas, j1 *joypad.JoyPad, j2 *joypad.JoyPad) {
		// フレームレート制御
		frameDuration := time.Second / FRAME_PER_SECOND
		now := time.Now()
		if elapsed := now.Sub(lastFrameTime); elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
		lastFrameTime = time.Now()

		// 全ウィンドウの描画
		f.windows.RenderAll()

		// イベント処理
		for event := eventPump(); event != nil; event = eventPump() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				f.requestShutdown()
			case *sdl.KeyboardEvent:
				if e.State == sdl.PRESSED {
					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						f.requestShutdown()
					case sdl.K_F12:
						f.cpu.ToggleLog()
					case sdl.K_F1:
						if optionWindowOpen {
							closeOptionWindow()
						} else {
							openOptionWindow()
						}
					case sdl.K_F2:
						if vramWindowOpen {
							closeVramWindow()
						} else {
							openVramWindow(p)
						}
					case sdl.K_F3:
						if chrWindowOpen {
							closeChrWindow()
						} else {
							openChrWindow(p)
						}
					case sdl.K_F4:
						if apuWindowOpen {
							closeAPUWindow()
						} else {
							openAPUWindow()
						}
					}
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

			f.windows.HandleEvent(event)
		}

		// コントローラの状態更新
		f.updateJoyPad(j1, &f.keyboard1, &f.controller1)
		f.updateJoyPad(j2, &f.keyboard2, &f.controller2)
	})

	// ゲームウィンドウの作成
	gameWindow, err := ui.NewGameWindow(f.config.SCALE_FACTOR, f.bus.Canvas(), func() {
		f.requestShutdown()
	})
	if err != nil {
		panic(err)
	}
	f.windows.Add(gameWindow)

	// CPU の作成と起動
	f.cpu.Init(f.bus, false)
	f.cpu.Run()
}

// MARK: ゲームの終了メソッド
func (f *Famicom) requestShutdown() {
	if f.windows != nil {
		f.windows.CloseAll()
	}
	f.bus.Shutdown()
	os.Exit(0)
}

// MARK: キーボードの状態を検知するメソッド
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

// MARK: コントローラーのボタン状態を検知するメソッド
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

// MARK: コントローラーのスティック状態を検知するメソッド
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

// MARK: JoyPadの状態を更新するメソッド
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
