package main

import (
	"Famicom-emulator/cartridge"
	"Famicom-emulator/cpu"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

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
		fmt.Println(c.Trace())

		// イベント処理
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.Keysym.Sym == sdl.K_ESCAPE && e.State == sdl.PRESSED {
					running = false
				}
			}
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

// デバッグ用スネークゲーム
func runCPUTestGame(cartridge *cartridge.Cartridge) {
	c := cpu.CreateCPU(true)

	c.InitWithCartridge(cartridge, true)
	c.Registers.PC = 0xC000

	// ウィンドウとレンダラーを準備
	createWindow(c)
}