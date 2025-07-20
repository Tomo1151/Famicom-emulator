package main

import (
	"Famicom-emulator/bus"
	"Famicom-emulator/cartridge"
	"Famicom-emulator/cpu"
	"Famicom-emulator/joypad"
	"Famicom-emulator/ppu"
	"log"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const SCALE_FACTOR = 3

func main() {
	filedata, err := os.ReadFile("../rom/nestest.nes")

	if err != nil {
		log.Fatalf("Error occured in 'os.ReadFile()'")
	}

	cart := cartridge.Cartridge{}
	err = cart.Load(filedata)
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}

	frame := ppu.Frame{}
	frame.Init()

	pad := joypad.JoyPad{}
	pad.Init()

	// デバッグ：フレームバッファの最初の数バイトを確認
	log.Printf("Frame buffer first 12 bytes: %v", frame.Buffer[:12])
	log.Printf("Frame buffer size: %d", len(frame.Buffer))

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Famicom", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(frame.Width)*SCALE_FACTOR, int32(frame.Height)*SCALE_FACTOR, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	// SDL2テクスチャを作成
	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_RGB24,
		sdl.TEXTUREACCESS_STREAMING,
		int32(frame.Width), int32(frame.Height))
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	bus := bus.Bus{}
	bus.InitWithCartridge(&cart, &pad, func(p *ppu.PPU, j *joypad.JoyPad) {
		// PPUコールバック内でフレームをレンダリング
		ppu.Render(*p, &frame)
	})

	c := cpu.CPU{}
	c.InitWithCartridge(bus, true) // デバッグログを無効化

	// PPUのRender関数がVRAMの初期状態（全て0）でどう動作するかテスト
	log.Printf("PPU Render function test with initial VRAM state")

	// CPU実行用のゴルーチンと制御変数
	var running bool = true
	var mu sync.Mutex

	// CPU実行をゴルーチンで別スレッドで実行
	go func() {
		c.RunWithCallback(func(cpu *cpu.CPU) {
			mu.Lock()
			isRunning := running
			mu.Unlock()

			if !isRunning {
				// 実行を停止するためのパニックを使用
				// より良い方法があれば置き換え可能
				panic("CPU execution stopped")
			}

			// フレームレート制御（CPUを少し休ませる）
			time.Sleep(time.Microsecond * 16) // 約60kHz
		})
	}()

	// メインイベントループ
	for running {
		// イベント処理
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				mu.Lock()
				running = false
				mu.Unlock()
				return
			case *sdl.KeyboardEvent:
				// ESCで終了
				if e.Keysym.Sym == sdl.K_ESCAPE && e.State == sdl.PRESSED {
					mu.Lock()
					running = false
					mu.Unlock()
					return
				}

				// Aボタン
				if e.Keysym.Sym == sdl.K_k {
					if e.State == sdl.PRESSED {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_A_POSITION, true)
						// fmt.Println("SDL: BUTTON A Pressed.")
					} else {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_A_POSITION, false)
					}
				}

				// Bボタン
				if e.Keysym.Sym == sdl.K_j {
					if e.State == sdl.PRESSED {
						// fmt.Println("SDL: BUTTON B Pressed.")
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_B_POSITION, true)
						// fmt.Printf("A Button pressed: state=0x%02X\n", pad.State)
					} else {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_B_POSITION, false)
						// fmt.Printf("A Button released: state=0x%02X\n", pad.State)
					}
				}

				// 上ボタン
				if e.Keysym.Sym == sdl.K_w {
					if e.State == sdl.PRESSED {
						// fmt.Println("SDL: BUTTON UP Pressed.")
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_UP_POSITION, true)
					} else {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_UP_POSITION, false)
					}
				}

				// 下ボタン
				if e.Keysym.Sym == sdl.K_s {
					if e.State == sdl.PRESSED {
						// fmt.Println("SDL: BUTTON DOWN Pressed.")
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_DOWN_POSITION, true)
					} else {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_DOWN_POSITION, false)
					}
				}

				// 左ボタン
				if e.Keysym.Sym == sdl.K_a {
					if e.State == sdl.PRESSED {
						// fmt.Println("SDL: BUTTON LEFT Pressed.")
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_LEFT_POSITION, true)
					} else {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_LEFT_POSITION, false)
					}
				}

				// 右ボタン
				if e.Keysym.Sym == sdl.K_d {
					if e.State == sdl.PRESSED {
						// fmt.Println("SDL: BUTTON RIGHT Pressed.")
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_RIGHT_POSITION, true)
					} else {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_RIGHT_POSITION, false)
					}
				}

				// スタートボタン
				if e.Keysym.Sym == sdl.K_RETURN {
					if e.State == sdl.PRESSED {
						// fmt.Println("SDL: BUTTON START Pressed.")
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_START_POSITION, true)
					} else {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_START_POSITION, false)
					}
				}

				// セレクトボタン
				if e.Keysym.Sym == sdl.K_BACKSPACE {
					if e.State == sdl.PRESSED {
						// fmt.Println("SDL: BUTTON SELECT Pressed.")
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_SELECT_POSITION, true)
					} else {
						bus.SetJoypad1ButtonPressed(joypad.JOYPAD_BUTTON_SELECT_POSITION, false)
					}
				}
			}
		}

		// テクスチャを更新（unsafe.Pointerを使用）
		err = texture.Update(nil, unsafe.Pointer(&frame.Buffer[0]), int(frame.Width*3))
		if err != nil {
			panic(err)
		}

		// 画面を再描画
		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		// フレームレート制御
		sdl.Delay(16) // 約60FPS
	}
}
