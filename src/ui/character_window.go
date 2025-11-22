package ui

import (
	"Famicom-emulator/ppu"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: CharacterWindowの定義
type CharacterWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	ppu      *ppu.PPU
	buf      []byte
	onClose  func(id uint32)
	baseW    int
	baseH    int
	scale    int
}

// MARK: CharacterWindow の作成メソッド
func NewCharacterWindow(p *ppu.PPU, scale int, onClose func(id uint32)) (*CharacterWindow, error) {
	const tileSize = 8
	const colsPerTable = 16
	const rowsPerTable = 16
	const tables = 2

	w := colsPerTable * tileSize * tables // 256
	h := rowsPerTable * tileSize          // 128

	win, err := sdl.CreateWindow(
		"CHR ROM Viewer",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(w*scale),
		int32(h*scale),
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, err
	}

	r, err := sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		win.Destroy()
		return nil, err
	}

	t, err := r.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		r.Destroy()
		win.Destroy()
		return nil, err
	}

	buf := make([]byte, w*h*3)

	return &CharacterWindow{window: win, renderer: r, texture: t, ppu: p, buf: buf, onClose: onClose, baseW: w, baseH: h, scale: scale}, nil
}

// MARK: ウィンドウのID取得メソッド
func (c *CharacterWindow) ID() uint32 {
	id, _ := c.window.GetID()
	return id
}

// MARK: イベント処理メソッド
func (c *CharacterWindow) HandleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_CLOSE {
			c.requestClose()
		}
	case *sdl.KeyboardEvent:
		if e.State == sdl.PRESSED {
			switch e.Keysym.Sym {
			case sdl.K_ESCAPE:
				c.requestClose()
			case sdl.K_UP:
				c.setScale(c.scale + 1)
			case sdl.K_DOWN:
				c.setScale(c.scale - 1)
			}
		}
	}
}

// MARK: スケール設定メソッド
func (c *CharacterWindow) setScale(s int) {
	if s < 1 {
		s = 1
	}
	if s > 8 {
		s = 8
	}
	if s == c.scale {
		return
	}
	c.scale = s
	if c.window != nil {
		c.window.SetSize(int32(c.baseW*c.scale), int32(c.baseH*c.scale))
	}
}

// MARK: ウィンドウの更新メソッド
func (c *CharacterWindow) Update() {
	const tileSize = 8
	const colsPerTable = 16
	const rowsPerTable = 16
	const tables = 2

	width := colsPerTable * tileSize * tables

	// パレットはグレースケールで用意
	palette := [4][3]uint8{{0, 0, 0}, {85, 85, 85}, {170, 170, 170}, {255, 255, 255}}

	for t := range tables {
		for ty := range rowsPerTable {
			for tx := range colsPerTable {
				tileIdx := t*256 + ty*colsPerTable + tx

				basePx := tx*tileSize + t*colsPerTable*tileSize
				basePy := ty * tileSize

				ppuBase := uint16(tileIdx * 16)
				for row := range tileSize {
					b0 := c.ppu.Mapper.ReadCharacterRom(ppuBase + uint16(row))
					b1 := c.ppu.Mapper.ReadCharacterRom(ppuBase + uint16(row) + 8)
					for col := 0; col < tileSize; col++ {
						bit := (((b1 >> (7 - uint(col))) & 1) << 1) | ((b0 >> (7 - uint(col))) & 1)
						color := palette[bit]

						px := basePx + col
						py := basePy + row
						pos := (py*width + px) * 3
						c.buf[pos+0] = color[0]
						c.buf[pos+1] = color[1]
						c.buf[pos+2] = color[2]
					}
				}
			}
		}
	}

	c.texture.Update(nil, unsafe.Pointer(&c.buf[0]), int(width*3))
}

// MARK: 描画メソッド
func (c *CharacterWindow) Render() {
	c.renderer.Clear()
	c.renderer.Copy(c.texture, nil, nil)
	c.renderer.Present()
}

// MARK: SDLリソースの解放メソッド
func (c *CharacterWindow) Close() {
	if c.texture != nil {
		c.texture.Destroy()
	}
	if c.renderer != nil {
		c.renderer.Destroy()
	}
	if c.window != nil {
		c.window.Destroy()
	}
	c.texture = nil
	c.renderer = nil
	c.window = nil
}

// MARK: ウィンドウを閉じるメソッド
func (c *CharacterWindow) requestClose() {
	if c.onClose != nil {
		c.onClose(c.ID())
	}
}
