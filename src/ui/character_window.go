package ui

import (
	"Famicom-emulator/ppu"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数定義
const (
	TABLE_COLUMNS = 16
	TABLE_ROWS    = 16
	TABLES        = 2

	PALETTE_COLUMNS = 16
	PALETTE_ROWS    = 2
	SWATCH          = 16
)

// MARK: CharacterWindowの定義
type CharacterWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	ppu      *ppu.PPU
	buffer   []byte
	onClose  func(id uint32)
	baseW    int
	baseH    int
	tileH    int
	scale    int
}

// MARK: CharacterWindow の作成メソッド
func NewCharacterWindow(p *ppu.PPU, scale int, onClose func(id uint32)) (*CharacterWindow, error) {
	tileH := TABLE_ROWS * int(ppu.TILE_SIZE) // 128
	paletteH := SWATCH * PALETTE_ROWS        // 32

	width := TABLE_COLUMNS * int(ppu.TILE_SIZE) * TABLES // 256
	height := tileH + paletteH                           // 160

	win, err := sdl.CreateWindow(
		"CHR ROM Viewer",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(width*scale),
		int32(height*scale),
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

	t, err := r.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_STREAMING, int32(width), int32(height))
	if err != nil {
		r.Destroy()
		win.Destroy()
		return nil, err
	}

	buffer := make([]byte, width*height*3)

	return &CharacterWindow{window: win, renderer: r, texture: t, ppu: p, buffer: buffer, onClose: onClose, baseW: width, baseH: height, tileH: tileH, scale: scale}, nil
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
			case sdl.K_PLUS, sdl.K_SEMICOLON:
				c.setScale(c.scale + 1)
			case sdl.K_MINUS:
				c.setScale(c.scale - 1)
			}
		}
	}
}

// MARK: スケール設定メソッド
func (c *CharacterWindow) setScale(s int) {
	// 1 ~ 8 の間に設定
	s = min(max(s, 1), 8)
	if s == c.scale {
		return
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

	width := TABLE_COLUMNS * ppu.TILE_SIZE * TABLES

	// パレットはグレースケールで用意
	paletteTable := c.ppu.PaletteTable()
	bgPalette := [4]uint8{
		(*paletteTable)[0],
		(*paletteTable)[1],
		(*paletteTable)[2],
		(*paletteTable)[3],
	}

	// タイル一覧描画
	for t := range TABLES {
		for ty := range TABLE_ROWS {
			for tx := range TABLE_COLUMNS {
				tileIdx := t*256 + ty*TABLE_COLUMNS + tx

				basePx := tx*int(ppu.TILE_SIZE) + t*TABLE_COLUMNS*int(ppu.TILE_SIZE)
				basePy := ty * int(ppu.TILE_SIZE)

				ppuBase := uint16(tileIdx * 16)

				// フレーム最初のマッパーを使用
				mapper := c.ppu.MapperSnapshot()

				for row := range int(ppu.TILE_SIZE) {
					b0 := mapper.ReadCharacterRom(ppuBase + uint16(row))
					b1 := mapper.ReadCharacterRom(ppuBase + uint16(row) + 8)
					for col := range int(ppu.TILE_SIZE) {
						bit := (((b1 >> (7 - uint(col))) & 1) << 1) | ((b0 >> (7 - uint(col))) & 1)
						index := bgPalette[bit]
						color := ppu.PALETTE[index]

						px := basePx + col
						py := basePy + row
						pos := (py*int(width) + px) * 3
						c.buffer[pos+0] = color[0]
						c.buffer[pos+1] = color[1]
						c.buffer[pos+2] = color[2]
					}
				}
			}
		}
	}

	// パレット一覧描画
	paletteTop := c.tileH
	for i := range PALETTE_ROWS * PALETTE_COLUMNS {
		px0 := (i % PALETTE_COLUMNS) * SWATCH
		py0 := (i/PALETTE_COLUMNS)*SWATCH + paletteTop

		index := (*paletteTable)[i]
		color := ppu.PALETTE[index]

		for dy := range SWATCH {
			py := py0 + dy
			for dx := range SWATCH {
				px := px0 + dx
				pos := (py*int(width) + px) * 3
				c.buffer[pos+0] = color[0]
				c.buffer[pos+1] = color[1]
				c.buffer[pos+2] = color[2]
			}
		}
	}

	c.texture.Update(nil, unsafe.Pointer(&c.buffer[0]), int(width*3))
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
