package ui

import (
	"Famicom-emulator/ppu"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数定義
const (
	OAM_COLUMNS = 8
	OAM_ROWS    = 8
)

// MARK: OAMWindow の定義
type OAMWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	ppu      *ppu.PPU
	buffer   []byte
	onClose  func(id uint32)
	baseW    int
	baseH    int
	columns  int
	rows     int
	scale    int
}

// MARK: OAMWindow の作成メソッド
func NewOAMWindow(p *ppu.PPU, scale int, onClose func(id uint32)) (*OAMWindow, error) {
	spriteHeight := int(p.SpriteSize())
	columns := OAM_COLUMNS
	rows := OAM_ROWS
	if spriteHeight == 16 {
		columns = 16
		rows = 4
	}

	width := int(columns * int(ppu.TILE_SIZE))
	height := rows * spriteHeight

	win, err := sdl.CreateWindow(
		"OAM Viewer",
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
	return &OAMWindow{window: win, renderer: r, texture: t, ppu: p, buffer: buffer, onClose: onClose, baseW: width, baseH: height, columns: columns, rows: rows, scale: scale}, nil
}

// MARK: ウィンドウのID取得メソッド
func (o *OAMWindow) ID() uint32 {
	id, _ := o.window.GetID()
	return id
}

// MARK: イベント処理メソッド
func (o *OAMWindow) HandleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_CLOSE {
			o.requestClose()
		}
	case *sdl.KeyboardEvent:
		if e.State == sdl.PRESSED {
			switch e.Keysym.Sym {
			case sdl.K_ESCAPE:
				o.requestClose()
			case sdl.K_PLUS, sdl.K_SEMICOLON:
				o.setScale(o.scale + 1)
			case sdl.K_MINUS:
				o.setScale(o.scale - 1)
			}
		}
	}
}

// MARK: スケール設定メソッド
func (o *OAMWindow) setScale(s int) {
	// 1 ~ 8 の間に設定
	s = min(max(s, 1), 8)
	if s == o.scale {
		return
	}
	o.scale = s
	if o.window != nil {
		o.window.SetSize(int32(o.baseW*o.scale), int32(o.baseH*o.scale))
	}
}

// MARK: ウィンドウの更新メソッド
func (o *OAMWindow) Update() {
	width := int(o.columns * int(ppu.TILE_SIZE))
	spriteHeight := int(o.ppu.SpriteSize())

	// パレットはスプライトのパレット0固定 ($3F10..$3F13)
	paletteTable := o.ppu.PaletteTable()
	bgColor := ppu.PALETTE[(*paletteTable)[0]]
	palette0 := [4]uint8{(*paletteTable)[16], (*paletteTable)[17], (*paletteTable)[18], (*paletteTable)[19]}

	// フレーム最初のマッパーを使用
	mapper := o.ppu.MapperSnapshot()

	oam := o.ppu.OAM()

	// クリア（前フレームの残りが見えないように）
	for i := 0; i < len(o.buffer); i += 3 {
		o.buffer[i+0] = bgColor[0]
		o.buffer[i+1] = bgColor[1]
		o.buffer[i+2] = bgColor[2]
	}

	// 64スプライトを一覧表示
	for i := range 64 {
		tileIndex := (*oam)[i*4+1]
		attr := (*oam)[i*4+2]
		hFlip := (attr & (1 << 6)) != 0
		vFlip := (attr & (1 << 7)) != 0

		gridX := i % o.columns
		gridY := i / o.columns
		basePx := gridX * int(ppu.TILE_SIZE)
		basePy := gridY * spriteHeight

		// 描画対象のスプライトがウィンドウのバッファ外ならスキップ
		if basePy+spriteHeight > o.baseH {
			continue
		}

		for row := range spriteHeight {
			srcRow := row
			if vFlip {
				srcRow = spriteHeight - 1 - row
			}

			var tileBase uint16
			var tileRow int
			if spriteHeight == 8 {
				bankBase := o.ppu.SpritePatternTableAddress()
				tileBase = bankBase + uint16(tileIndex)*uint16(ppu.TILE_SIZE*2)
				tileRow = srcRow
			} else {
				// 8x16スプライト:
				// タイル番号 bit0 がパターンテーブル選択、bit1-7 がタイル番号(上側)を表す
				bankBase := uint16(tileIndex&1) * 0x1000
				topTile := tileIndex & 0xFE
				tileNum := topTile
				if srcRow >= 8 {
					tileNum = topTile + 1
					tileRow = srcRow - 8
				} else {
					tileRow = srcRow
				}
				tileBase = bankBase + uint16(tileNum)*uint16(ppu.TILE_SIZE*2)
			}

			b0 := mapper.ReadCharacterRom(tileBase + uint16(tileRow))
			b1 := mapper.ReadCharacterRom(tileBase + uint16(tileRow) + uint16(ppu.TILE_SIZE))

			for col := range int(ppu.TILE_SIZE) {
				srcCol := col
				if hFlip {
					srcCol = int(ppu.TILE_SIZE) - 1 - col
				}
				bit := (((b1 >> (7 - uint(srcCol))) & 1) << 1) | ((b0 >> (7 - uint(srcCol))) & 1)
				if bit == 0 {
					continue
				}
				colorIdx := palette0[bit]
				color := ppu.PALETTE[colorIdx]

				px := basePx + col
				py := basePy + row
				pos := (py*width + px) * 3
				o.buffer[pos+0] = color[0]
				o.buffer[pos+1] = color[1]
				o.buffer[pos+2] = color[2]
			}
		}
	}

	o.texture.Update(nil, unsafe.Pointer(&o.buffer[0]), int(width*3))
}

// MARK: 描画メソッド
func (o *OAMWindow) Render() {
	o.renderer.Clear()
	o.renderer.Copy(o.texture, nil, nil)
	o.renderer.Present()
}

// MARK: SDLリソースの解放メソッド
func (o *OAMWindow) Close() {
	if o.texture != nil {
		o.texture.Destroy()
	}
	if o.renderer != nil {
		o.renderer.Destroy()
	}
	if o.window != nil {
		o.window.Destroy()
	}
	o.texture = nil
	o.renderer = nil
	o.window = nil
}

// MARK: ウィンドウを閉じるメソッド
func (o *OAMWindow) requestClose() {
	if o.onClose != nil {
		o.onClose(o.ID())
	}
}
