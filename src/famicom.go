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
	"runtime"
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

	gamepad1 sdl.JoystickID       // SDLのコントローラID (1P)
	gamepad2 sdl.JoystickID       // SDLのコントローラID (2P)
	adapter1 joypad.JoyPadAdapter // 1Pコントローラのアダプタ
	adapter2 joypad.JoyPadAdapter // 2Pコントローラのアダプタ

	config  *config.Config
	windows *ui.WindowManager
}

// MARK: Famicomの初期化メソッド
func (f *Famicom) Init(cartridge cartridge.Cartridge, config *config.Config) {
	// ROMファイルのロード
	f.cartridge = cartridge
	err := f.cartridge.Load()
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}

	// 各コンポーネントの接続
	f.config = config
	f.bus.ConnectComponents(
		&f.ppu,
		&f.apu,
		&f.cartridge,
		&f.joypad1,
		&f.joypad2,
		f.config,
	)
}

// MARK: Famicomの起動
func (f *Famicom) Start() {
	// SDLの初期化
	runtime.LockOSThread() // SDLはMainスレッド上で動かす必要がある
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
			f.adapter1.Init(gamepad1.Name())
			defer gamepad1.Close()
		}
	}
	if sdl.NumJoysticks() > 1 {
		gamepad2 = sdl.GameControllerOpen(1)
		if gamepad2 != nil {
			f.gamepad2 = gamepad2.Joystick().InstanceID()
			f.adapter2.Init(gamepad2.Name())
			fmt.Println("Controller opened for 2P:", gamepad2.Name())
			defer gamepad2.Close()
		}
	}

	// 状態変数の定義
	eventPump := sdl.PollEvent

	// フレーム同期用チャネル: PPU がフレーム完了を通知するが、
	// メインスレッドが CPU を駆動しているので非ブロッキングで通知する。
	frameCh := make(chan struct{}, 1) // PPU -> main (non-blocking, buffered)

	// Busの初期化とフレーム毎に実行されるコールバックの定義
	// PPU のフレーム到達を非ブロッキングで通知する（CPU はメインループが駆動する）。
	f.bus.Init(func(p *ppu.PPU, c *ppu.Canvas, j1 *joypad.JoyPad, j2 *joypad.JoyPad) {
		select {
		case frameCh <- struct{}{}:
		default:
		}
	})

	// ゲームウィンドウの作成
	gameWindow, err := ui.NewGameWindow(f.config.Render.SCALE_FACTOR, f.bus.Canvas(), func() {
		f.requestShutdown()
	})
	if err != nil {
		panic(err)
	}
	f.windows.Add(gameWindow)

	// CPU の作成（フレーム単位でメインが駆動する）
	f.cpu.Init(f.bus, f.config.CPU.LOG_ENABLED)

	// メインループ: メインが CPU をフレーム単位で駆動し、描画とイベント処理を行う。
	// フレームあたりの CPU サイクル数を計算して RunCycles に渡す。
	const ntscCpuClock = 1789773
	baseCycles := int(ntscCpuClock / FRAME_PER_SECOND)        // 29829
	remainderPerFrame := int(ntscCpuClock % FRAME_PER_SECOND) // 33
	remAcc := 0

	lastFrameTime := time.Now()

	for {
		// フレームのサイクル数を計算
		remAcc += remainderPerFrame
		extra := 0
		if remAcc >= FRAME_PER_SECOND {
			extra = 1
			remAcc -= FRAME_PER_SECOND
		}
		cyclesThisFrame := uint(baseCycles + extra)

		// CPU をフレーム分だけ実行する（これがエミュレータの実時間を決める）
		f.cpu.RunCycles(cyclesThisFrame)

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
					case sdl.K_F1:
						if f.windows != nil {
							if _, err := f.windows.ToggleOptionWindow(f.config); err != nil {
								log.Printf("failed to toggle option window: %v", err)
							}
						}
					case sdl.K_F2:
						if f.windows != nil {
							if _, err := f.windows.ToggleNameTableWindow(&f.ppu, f.config.Render.SCALE_FACTOR); err != nil {
								log.Printf("failed to toggle name table window: %v", err)
							}
						}
					case sdl.K_F3:
						if f.windows != nil {
							if _, err := f.windows.ToggleCharacterWindow(&f.ppu, f.config.Render.SCALE_FACTOR); err != nil {
								log.Printf("failed to toggle character window: %v", err)
							}
						}
					case sdl.K_F4:
						if f.windows != nil {
							if _, err := f.windows.ToggleAudioWindow(&f.apu, f.config.Render.SCALE_FACTOR); err != nil {
								log.Printf("failed to toggle audio window: %v", err)
							}
						}
					case sdl.K_F10:
						f.apu.ToggleLog()
					case sdl.K_F11:
						f.cpu.ToggleLog()
					case sdl.K_1:
						f.apu.ToggleMute1ch()
					case sdl.K_2:
						f.apu.ToggleMute2ch()
					case sdl.K_3:
						f.apu.ToggleMute3ch()
					case sdl.K_4:
						f.apu.ToggleMute4ch()
					case sdl.K_5:
						f.apu.ToggleMute5ch()
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

			// JoyPad状態の更新
			f.updateJoyPad(&f.joypad1, &f.keyboard1, &f.controller1)
			f.updateJoyPad(&f.joypad2, &f.keyboard2, &f.controller2)

			f.windows.HandleEvent(event)
		}

		// 描画（PPU フレームに合わせて）
		f.windows.RenderAll()

		// フレームレート制御: PPU が速すぎて音より先行するのを防ぐ
		frameDuration := time.Second / FRAME_PER_SECOND
		now := time.Now()
		if elapsed := now.Sub(lastFrameTime); elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
		lastFrameTime = time.Now()
	}
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
	case f.config.Control.KEY_1P.BUTTON_A:
		c1.A = pressed
	case f.config.Control.KEY_1P.BUTTON_B:
		c1.B = pressed
	case f.config.Control.KEY_1P.BUTTON_UP:
		c1.Up = pressed
	case f.config.Control.KEY_1P.BUTTON_DOWN:
		c1.Down = pressed
	case f.config.Control.KEY_1P.BUTTON_RIGHT:
		c1.Right = pressed
	case f.config.Control.KEY_1P.BUTTON_LEFT:
		c1.Left = pressed
	case f.config.Control.KEY_1P.BUTTON_START:
		c1.Start = pressed
	case f.config.Control.KEY_1P.BUTTON_SELECT:
		c1.Select = pressed

	// 2P
	case f.config.Control.KEY_2P.BUTTON_A:
		c2.A = pressed
	case f.config.Control.KEY_2P.BUTTON_B:
		c2.B = pressed
	case f.config.Control.KEY_2P.BUTTON_UP:
		c2.Up = pressed
	case f.config.Control.KEY_2P.BUTTON_DOWN:
		c2.Down = pressed
	case f.config.Control.KEY_2P.BUTTON_RIGHT:
		c2.Right = pressed
	case f.config.Control.KEY_2P.BUTTON_LEFT:
		c2.Left = pressed
	case f.config.Control.KEY_2P.BUTTON_START:
		c2.Start = pressed
	case f.config.Control.KEY_2P.BUTTON_SELECT:
		c2.Select = pressed
	}
}

// MARK: コントローラーのボタン状態を検知するメソッド
func (f *Famicom) handleButtonPress(e *sdl.ControllerButtonEvent, c *InputState) {
	var adapter joypad.JoyPadAdapter
	switch e.Which {
	case f.gamepad1:
		adapter = f.adapter1
	case f.gamepad2:
		adapter = f.adapter2
	default:
		return
	}

	pressed := e.State == sdl.PRESSED
	switch e.Button {
	case adapter.ButtonA():
		c.A = pressed
	case adapter.ButtonB():
		c.B = pressed
	case adapter.ButtonUp():
		c.Up = pressed
	case adapter.ButtonDown():
		c.Down = pressed
	case adapter.ButtonRight():
		c.Right = pressed
	case adapter.ButtonLeft():
		c.Left = pressed
	case adapter.ButtonStart():
		c.Start = pressed
	case adapter.ButtonSelect():
		c.Select = pressed
	}
}

// MARK: コントローラーのスティック状態を検知するメソッド
func (f *Famicom) handleAxisMotion(e *sdl.ControllerAxisEvent, c *InputState) {
	switch e.Axis {
	case 0: // X軸 (左スティック左右)
		if e.Value < -f.config.Control.GamepadAxisThreshold {
			c.Left = true
			c.Right = false
		} else if e.Value > f.config.Control.GamepadAxisThreshold {
			c.Left = false
			c.Right = true
		} else {
			c.Left = false
			c.Right = false
		}
	case 1: // Y軸 (左スティック上下)
		if e.Value < -f.config.Control.GamepadAxisThreshold {
			c.Up = true
			c.Down = false
		} else if e.Value > f.config.Control.GamepadAxisThreshold {
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
