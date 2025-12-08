package ui

import (
	"Famicom-emulator/ppu"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: GameWindow の定義
type GameWindow struct {
	window       *sdl.Window
	renderer     *sdl.Renderer
	texture      *sdl.Texture
	canvas       *ppu.Canvas
	isFullscreen bool
	scale        int
	onClose      func()
}

// MARK: GameWindow の作成メソッド
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

	r, err := sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		w.Destroy()
		return nil, err
	}

	// フルスクリーン時にアスペクト比を保つため，論理サイズを指定
	if err := r.SetLogicalSize(int32(ppu.SCREEN_WIDTH), int32(ppu.SCREEN_HEIGHT)); err != nil {
		r.Destroy()
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

	return &GameWindow{window: w, renderer: r, texture: t, canvas: canvas, isFullscreen: false, scale: scale, onClose: onClose}, nil
}

// MARK: ウィンドウのID取得メソッド
func (g *GameWindow) ID() uint32 {
	id, _ := g.window.GetID()
	return id
}

// MARK: イベント処理メソッド
func (g *GameWindow) HandleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_CLOSE && g.onClose != nil {
			g.onClose()
		}
	case *sdl.KeyboardEvent:
		if e.State == sdl.PRESSED {
			switch e.Keysym.Sym {
			case sdl.K_F12:
				if g.isFullscreen {
					g.window.SetFullscreen(0)
					g.setScale(g.scale)
				} else {
					g.setScale(1)
					g.window.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
				}
				g.isFullscreen = !g.isFullscreen
			}
		}
	}
}

// MARK: スケール設定メソッド
func (g *GameWindow) setScale(s int) {
	// 1 ~ 8 の間に設定
	s = min(max(s, 1), 8)

	if g.window != nil {
		g.window.SetSize(int32(ppu.SCREEN_WIDTH)*int32(s),
			int32(ppu.SCREEN_HEIGHT)*int32(s))
	}
}

// MARK: ウィンドウの更新メソッド
func (g *GameWindow) Update() {
	// 現在描画中のバッファを元に画面を更新
	buf := g.canvas.FrontBuffer()
	g.texture.Update(nil, unsafe.Pointer(&(*buf)[0]), int(g.canvas.Width*3))
}

// MARK: 描画メソッド
func (g *GameWindow) Render() {
	g.renderer.Clear()
	g.renderer.Copy(g.texture, nil, nil)
	g.renderer.Present()
}

// MARK: SDLリソースの解放メソッド
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
