package main

import (
	"Famicom-emulator/cartridge"
	"Famicom-emulator/cpu"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	fmt.Println("Hello, world!")

	filedata, err := os.ReadFile("../rom/nestest.nes")

	if err != nil {
		log.Fatalf("Error occured in 'os.ReadFile()'")
	}

	// fmt.Println(filedata)

	cart := cartridge.Cartridge{}
	err = cart.Load(filedata)
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}



	runCPUTestGame(&cart)
}

func createWindow(c *cpu.CPU) {
	// SDL2の初期化
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatalf("Failed to initialize SDL: %s", err)
	}
	defer sdl.Quit()

	rand.Seed(time.Now().UnixNano())
	scaleFactor := 10.0

	// ウィンドウの作成（中央配置）
	window, err := sdl.CreateWindow(
		"Famicom emulator",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(32 * scaleFactor),
		int32(32 * scaleFactor),
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		log.Fatalf("Failed to create window: %s", err)
	}
	defer window.Destroy()

	// レンダラーの作成（VSyncあり）
	renderer, err := sdl.CreateRenderer(
		window,
		-1,
		sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC,
	)
	if err != nil {
		log.Fatalf("Failed to create renderer: %s", err)
	}
	defer renderer.Destroy()

	// レンダラーのスケール設定
	if err := renderer.SetScale(float32(scaleFactor), float32(scaleFactor)); err != nil {
		log.Fatalf("Failed to set scale: %s", err)
	}

	// テクスチャの作成
	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_RGB24,
		sdl.TEXTUREACCESS_STREAMING,
		32, 32,
	)
	if err != nil {
		log.Fatalf("Failed to create texture: %s", err)
	}
	defer texture.Destroy()

	// 画面状態配列を初期化
	var screenState [32 * 3 * 32]uint8
	for i := range screenState {
		screenState[i] = 0
	}

	running := true

	// メインループ
	c.Run(func(c *cpu.CPU) {
		// イベント処理
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.Keysym.Sym == sdl.K_ESCAPE && e.State == sdl.PRESSED {
					running = false
				}

				// キー入力の処理（押された時のみ）
				if e.State == sdl.PRESSED {
					switch e.Keysym.Sym {
					case sdl.K_w:
						c.WriteByteAt(0xFF, 0x77)
					case sdl.K_s:
						c.WriteByteAt(0xFF, 0x73)
					case sdl.K_a:
						c.WriteByteAt(0xFF, 0x61)
					case sdl.K_d:
						c.WriteByteAt(0xFF, 0x64)
					default:
					}
				}
			}
		}

		// 0xFEに乱数を書き込み（1-15の範囲）
		c.WriteByteAt(0xFE, uint8(rand.Intn(15)+1))

		// 画面状態の更新をチェックし、変更があれば描画
		if readScreenState(c, &screenState) {
			// テクスチャを更新
			texture.Update(nil, unsafe.Pointer(&screenState), 32*3)

			// レンダラーをクリア
			renderer.Clear()

			// テクスチャをレンダラーにコピー
			renderer.Copy(texture, nil, nil)

			// 画面に表示
			renderer.Present()
		}

		// CPUのフレームレート調整
		time.Sleep(70 * time.Microsecond)

		// 終了条件をチェック
		if !running {
			sdl.Quit()
			os.Exit(0)
		}
	})
}

func readScreenState(c *cpu.CPU, frame *[32 * 3 * 32]uint8) bool {
	frameIdx := 0
	update := false
	for i := 0x0200; i < 0x0600; i++ {
		colorIdx := c.ReadByteFrom(uint16(i))
		col := color(colorIdx)
		if frame[frameIdx] != col.R || frame[frameIdx+1] != col.G || frame[frameIdx+2] != col.B {
			frame[frameIdx] = col.R
			frame[frameIdx+1] = col.G
			frame[frameIdx+2] = col.B
			update = true
		}

		frameIdx += 3
	}

	return update
}

func color(bytes uint8) sdl.Color {
	switch bytes {
	case 0:
		return sdl.Color{R: 0, G: 0, B: 0, A: 255} // BLACK
	case 1:
		return sdl.Color{R: 255, G: 255, B: 255, A: 255} // WHITE
	case 2, 9:
		return sdl.Color{R: 128, G: 128, B: 128, A: 255} // GREY
	case 3, 10:
		return sdl.Color{R: 255, G: 0, B: 0, A: 255} // RED
	case 4, 11:
		return sdl.Color{R: 0, G: 255, B: 0, A: 255} // GREEN
	case 5, 12:
		return sdl.Color{R: 0, G: 0, B: 255, A: 255} // BLUE
	case 6, 13:
		return sdl.Color{R: 255, G: 0, B: 255, A: 255} // MAGENTA
	case 7, 14:
		return sdl.Color{R: 255, G: 255, B: 0, A: 255} // YELLOW
	default:
		return sdl.Color{R: 0, G: 255, B: 255, A: 255} // CYAN
	}
}

// デバッグ用スネークゲーム
func runCPUTestGame(cartridge *cartridge.Cartridge) {
	c := cpu.CreateCPU(true)

	c.InitWithCartridge(cartridge, true)

	fmt.Printf("Entry point: $%04X\n", c.ReadByteFrom(0xFFFC))

	// ウィンドウとレンダラーを準備
	createWindow(c)
}