package ui

import (
	"Famicom-emulator/ppu"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// GameWindow renders the main NES screen.
type GameWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	canvas   *ppu.Canvas
	onClose  func()
}

// NewGameWindow creates the primary game viewport.
func NewGameWindow(scale int, canvas *ppu.Canvas, onClose func()) (*GameWindow, error) {
	w, err := sdl.CreateWindow(
		"Famicom emu",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		int32(ppu.SCREEN_WIDTH)*int32(scale),
		int32(ppu.SCREEN_HEIGHT)*int32(scale),
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, err
	}

	r, err := sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		w.Destroy()
		return nil, err
	}

	t, err := r.CreateTexture(
		sdl.PIXELFORMAT_RGB24,
		sdl.TEXTUREACCESS_STREAMING,
		int32(ppu.SCREEN_WIDTH),
		int32(ppu.SCREEN_HEIGHT),
	)
	if err != nil {
		r.Destroy()
		w.Destroy()
		return nil, err
	}

	return &GameWindow{window: w, renderer: r, texture: t, canvas: canvas, onClose: onClose}, nil
}

// ID returns the SDL window identifier.
func (g *GameWindow) ID() uint32 {
	id, _ := g.window.GetID()
	return id
}

// HandleEvent reacts to SDL events routed to this window.
func (g *GameWindow) HandleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_CLOSE && g.onClose != nil {
			g.onClose()
		}
	}
}

// Update refreshes the SDL texture with the latest PPU frame buffer.
func (g *GameWindow) Update() {
	g.texture.Update(nil, unsafe.Pointer(&g.canvas.Buffer[0]), int(g.canvas.Width*3))
}

// Render draws the latest frame.
func (g *GameWindow) Render() {
	g.renderer.Clear()
	g.renderer.Copy(g.texture, nil, nil)
	g.renderer.Present()
}

// Close releases SDL resources.
func (g *GameWindow) Close() {
	if g.texture != nil {
		g.texture.Destroy()
	}
	if g.renderer != nil {
		g.renderer.Destroy()
	}
	if g.window != nil {
		g.window.Destroy()
	}
}
