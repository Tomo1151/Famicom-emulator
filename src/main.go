package main

import (
	"Famicom-emulator/bus"
	"Famicom-emulator/cartridge"
	"Famicom-emulator/cpu"
	"Famicom-emulator/joypad"
	"Famicom-emulator/ppu"
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const SCALE_FACTOR = 3

func main() {
	filedata, err := os.ReadFile("../rom/SuperMarioBros.nes")
	if err != nil {
		log.Fatalf("Error occured in 'os.ReadFile()'")
	}

	cart := cartridge.Cartridge{}
	err = cart.Load(filedata)
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Famicom", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(ppu.SCREEN_WIDTH)*SCALE_FACTOR, int32(ppu.SCREEN_HEIGHT)*SCALE_FACTOR, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

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

	var lastFrameTime = time.Now()

	// BusのNMIコールバックで描画とイベント処理
	bus := bus.Bus{}
	bus.InitWithCartridge(&cart, func(p *ppu.PPU, c *ppu.Canvas, j0 *joypad.JoyPad, j1 *joypad.JoyPad) {

		now := time.Now()
		elapsed := now.Sub(lastFrameTime)
		const frameDuration = time.Second / 60
		if elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
		lastFrameTime = time.Now()

		ppu.Render(p, c)
		texture.Update(nil, unsafe.Pointer(&c.Buffer[0]), int(c.Width*3))
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
				// キー入力をJoypadに反映
				switch e.Keysym.Sym {
				case sdl.K_k:
					j0.SetButtonPressed(joypad.JOYPAD_BUTTON_A_POSITION, e.State == sdl.PRESSED)
				case sdl.K_j:
					j0.SetButtonPressed(joypad.JOYPAD_BUTTON_B_POSITION, e.State == sdl.PRESSED)
				case sdl.K_w:
					j0.SetButtonPressed(joypad.JOYPAD_BUTTON_UP_POSITION, e.State == sdl.PRESSED)
				case sdl.K_s:
					j0.SetButtonPressed(joypad.JOYPAD_BUTTON_DOWN_POSITION, e.State == sdl.PRESSED)
				case sdl.K_a:
					j0.SetButtonPressed(joypad.JOYPAD_BUTTON_LEFT_POSITION, e.State == sdl.PRESSED)
				case sdl.K_d:
					j0.SetButtonPressed(joypad.JOYPAD_BUTTON_RIGHT_POSITION, e.State == sdl.PRESSED)
				case sdl.K_RETURN, sdl.K_KP_ENTER:
					j0.SetButtonPressed(joypad.JOYPAD_BUTTON_START_POSITION, e.State == sdl.PRESSED)
				case sdl.K_BACKSPACE:
					j0.SetButtonPressed(joypad.JOYPAD_BUTTON_SELECT_POSITION, e.State == sdl.PRESSED)
				}
			}
		}
	})

	c := cpu.CPU{}
	c.InitWithCartridge(bus, true)
	c.Run()
}
