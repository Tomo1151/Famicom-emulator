package main

import (
	"Famicom-emulator/cartridge"
	"Famicom-emulator/cpu"
	"Famicom-emulator/ppu"
	"fmt"
	"log"
	"os"
)

func main() {
	filedata, err := os.ReadFile("../rom/SuperMarioBros/MAP0.NES")

	if err != nil {
		log.Fatalf("Error occured in 'os.ReadFile()'")
	}


	cart := cartridge.Cartridge{}
	err = cart.Load(filedata)
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}

	for i := range 0xFF {
		pixels := cart.CharacterROM[i*16:(i+1)*16]
		tile := ppu.GetTile(pixels)
		ppu.DumpTile(tile)
		fmt.Println()
	}

	// tileFrame := ppu.ShowTile(cart.CharacterROM, 1, 0)

	// if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
	// 	panic(err)
	// }
	// defer sdl.Quit()

	// window, err := sdl.CreateWindow("Frame Buffer", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
	// 	int32(tileFrame.Width), int32(tileFrame.Height), sdl.WINDOW_SHOWN)
	// if err != nil {
	// 	panic(err)
	// }
	// defer window.Destroy()

	// renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	// if err != nil {
	// 	panic(err)
	// }
	// defer renderer.Destroy()

	// texture, err := renderer.CreateTexture(
	// 	sdl.PIXELFORMAT_RGB24,
	// 	sdl.TEXTUREACCESS_STREAMING,
	// 	int32(tileFrame.Width), int32(tileFrame.Height))
	// if err != nil {
	// 	panic(err)
	// }
	// defer texture.Destroy()
	// // テクスチャにピクセルデータをアップロード
	// err = texture.Update(nil, unsafe.Pointer(&tileFrame.Buffer[0]), int(tileFrame.Width)*3)
	// if err != nil {
	// 	panic(err)
	// }

	// // 描画
	// renderer.Clear()
	// renderer.Copy(texture, nil, nil)
	// renderer.Present()

	// sdl.Delay(100000)
}


func createWindow(c *cpu.CPU) {
	// // SDL2の初期化
	// if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
	// 	log.Fatalf("Failed to initialize SDL: %s", err)
	// }
	// defer sdl.Quit()

	// rand.Seed(time.Now().UnixNano())
	// scaleFactor := 10.0

	// // ウィンドウの作成（中央配置）
	// window, err := sdl.CreateWindow(
	// 	"Famicom emulator",
	// 	sdl.WINDOWPOS_CENTERED,
	// 	sdl.WINDOWPOS_CENTERED,
	// 	int32(32 * scaleFactor),
	// 	int32(32 * scaleFactor),
	// 	sdl.WINDOW_SHOWN,
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to create window: %s", err)
	// }
	// defer window.Destroy()

	// // レンダラーの作成（VSyncあり）
	// renderer, err := sdl.CreateRenderer(
	// 	window,
	// 	-1,
	// 	sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC,
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to create renderer: %s", err)
	// }
	// defer renderer.Destroy()

	// // レンダラーのスケール設定
	// if err := renderer.SetScale(float32(scaleFactor), float32(scaleFactor)); err != nil {
	// 	log.Fatalf("Failed to set scale: %s", err)
	// }

	// テクスチャの作成
	// texture, err := renderer.CreateTexture(
	// 	sdl.PIXELFORMAT_RGB24,
	// 	sdl.TEXTUREACCESS_STREAMING,
	// 	32, 32,
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to create texture: %s", err)
	// }
	// defer texture.Destroy()

	// // 画面状態配列を初期化
	// var screenState [32 * 3 * 32]uint8
	// for i := range screenState {
	// 	screenState[i] = 0
	// }

	// running := true

	// // メインループ
	// c.Run(func(c *cpu.CPU) {
	// 	fmt.Println(c.Trace())

	// 	// イベント処理
	// 	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
	// 		switch e := event.(type) {
	// 		case *sdl.QuitEvent:
	// 			running = false
	// 		case *sdl.KeyboardEvent:
	// 			if e.Keysym.Sym == sdl.K_ESCAPE && e.State == sdl.PRESSED {
	// 				running = false
	// 			}
	// 		}
	// 	}


	// 	// CPUのフレームレート調整
	// 	time.Sleep(70 * time.Microsecond)

	// 	// 終了条件をチェック
	// 	if !running {
	// 		sdl.Quit()
	// 		os.Exit(0)
	// 	}
	// })
}
