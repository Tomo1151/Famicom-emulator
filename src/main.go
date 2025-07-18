package main

import (
	"Famicom-emulator/bus"
	"Famicom-emulator/cartridge"
	"Famicom-emulator/cpu"
	"Famicom-emulator/ppu"
	"log"
	"os"
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


	// デバッグ：フレームバッファの最初の数バイトを確認
	log.Printf("Frame buffer first 12 bytes: %v", frame.Buffer[:12])
	log.Printf("Frame buffer size: %d", len(frame.Buffer))

	// ppu.DumpFrame(frame)

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Famicom", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(frame.Width) * SCALE_FACTOR, int32(frame.Height) * SCALE_FACTOR, sdl.WINDOW_SHOWN)
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
	bus.InitWithCartridge(&cart, func(p *ppu.PPU) {
		// PPUコールバック内でフレームをレンダリング
		ppu.Render(*p, &frame)
	})

	c := cpu.CPU{}
	c.InitWithCartridge(bus, false) // デバッグログを無効化

	// PPUのRender関数がVRAMの初期状態（全て0）でどう動作するかテスト
	log.Printf("PPU Render function test with initial VRAM state")

	// メインイベントループ
	running := true
	for running {
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

		// CPUを少しずつ実行（ノンブロッキング）
		for range 1000 { // 1フレームあたり1000命令実行
			c.Step()
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
