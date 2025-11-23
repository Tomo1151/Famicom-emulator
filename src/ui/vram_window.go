package ui

import (
	"Famicom-emulator/cartridge/mappers"
	"Famicom-emulator/ppu"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: NameTableWindow の定義
type NameTableWindow struct {
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

// MARK: NameTableWindow の作成メソッド
func NewNameTableWindow(p *ppu.PPU, scale int, onClose func(id uint32)) (*NameTableWindow, error) {
	const tileSize = 8
	const cols = 32
	const rows = 30
	// Show 4 nametables arranged as 2x2 (classic layout)
	const w = cols * 2 * tileSize // 512 (two tables horizontally)
	const h = rows * 2 * tileSize // 480 (two tables vertically)

	win, err := sdl.CreateWindow(
		"Name Table Viewer",
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

	return &NameTableWindow{window: win, renderer: r, texture: t, ppu: p, buf: buf, onClose: onClose, baseW: w, baseH: h, scale: scale}, nil
}

// MARK: ウィンドウのIDを取得するメソッド
func (n *NameTableWindow) ID() uint32 {
	id, _ := n.window.GetID()
	return id
}

// MARK: イベント処理メソッド
func (n *NameTableWindow) HandleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_CLOSE {
			n.requestClose()
		}
	case *sdl.KeyboardEvent:
		if e.State == sdl.PRESSED {
			switch e.Keysym.Sym {
			case sdl.K_ESCAPE:
				n.requestClose()
			case sdl.K_UP:
				n.setScale(n.scale + 1)
			case sdl.K_DOWN:
				n.setScale(n.scale - 1)
			}
		}
	}
}

// MARK: スケール設定メソッド
func (n *NameTableWindow) setScale(s int) {
	if s < 1 {
		s = 1
	}
	if s > 8 {
		s = 8
	}
	if s == n.scale {
		return
	}
	n.scale = s
	// Resize the window to match the new scale; texture stays the same size
	if n.window != nil {
		n.window.SetSize(int32(n.baseW*n.scale), int32(n.baseH*n.scale))
	}
}

// MARK: 更新メソッド
func (n *NameTableWindow) Update() {
	vramPtr := n.ppu.VRAM()
	paletteTable := n.ppu.PaletteTable()
	bankBase := n.ppu.BackgroundPatternTableAddress()

	const cols = 32
	const rows = 30
	const ntSize = 1024
	const tileSize = 8
	width := cols * 2 * tileSize

	// NT0: 左上, NT1: 右上, NT2: 左下, NT3: 右下
	for nt := range 4 {
		// マッパーをもとにVRAMの基準を計算
		mirr := n.ppu.Mapper.Mirroring()
		var physicalPage int
		switch mirr {
		case mappers.MIRRORING_VERTICAL:
			// Vertical: NT0==NT2, NT1==NT3 -> pages: 0,1,0,1
			physicalPage = nt % 2
		case mappers.MIRRORING_HORIZONTAL:
			// Horizontal: NT0==NT1, NT2==NT3 -> pages: 0,0,1,1
			physicalPage = nt / 2
		default:
			physicalPage = nt % 2
		}
		base := physicalPage * ntSize
		attrStart := base + 0x3C0

		xOffset := (nt % 2) * cols * tileSize
		yOffset := (nt / 2) * rows * tileSize

		// 全てのタイルに対して処理
		for ty := range rows {
			for tx := range cols {
				idx := base + ty*cols + tx
				tileIndex := (*vramPtr)[idx]

				// 属性テーブルのインデックスを計算
				attrTableIdx := (ty/4)*8 + (tx / 4)
				attrByte := (*vramPtr)[attrStart+attrTableIdx]

				// パレットインデックスの計算 (各2x2ブロックごとに2bit)
				sx := (tx % 4) / 2
				sy := (ty % 4) / 2
				var paletteIdx uint8
				if sx == 0 && sy == 0 {
					paletteIdx = (attrByte) & 0b11
				} else if sx == 1 && sy == 0 {
					paletteIdx = (attrByte >> 2) & 0b11
				} else if sx == 0 && sy == 1 {
					paletteIdx = (attrByte >> 4) & 0b11
				} else {
					paletteIdx = (attrByte >> 6) & 0b11
				}

				// パレットの決定
				paletteStart := 1 + uint(paletteIdx)*4
				paletteIndices := [4]uint8{
					(*paletteTable)[0],
					(*paletteTable)[paletteStart+0],
					(*paletteTable)[paletteStart+1],
					(*paletteTable)[paletteStart+2],
				}

				// タイルパターンのフェッチ
				tileBase := bankBase + uint16(tileIndex)*uint16(tileSize*2)

				for row := range tileSize {
					upper := n.ppu.Mapper.ReadCharacterRom(tileBase + uint16(row))
					lower := n.ppu.Mapper.ReadCharacterRom(tileBase + uint16(row) + uint16(tileSize))

					for col := range tileSize {
						bit := (((lower >> (7 - uint(col))) & 1) << 1) | ((upper >> (7 - uint(col))) & 1)
						colorIdx := paletteIndices[bit]
						color := ppu.PALETTE[colorIdx]

						px := xOffset + tx*tileSize + col
						py := yOffset + ty*tileSize + row
						pos := (py*width + px) * 3
						n.buf[pos+0] = color[0]
						n.buf[pos+1] = color[1]
						n.buf[pos+2] = color[2]
					}
				}
			}
		}
	}

	n.texture.Update(nil, unsafe.Pointer(&n.buf[0]), int(width*3))
}

// MARK: 描画メソッド
func (n *NameTableWindow) Render() {
	n.renderer.Clear()
	n.renderer.Copy(n.texture, nil, nil)
	n.renderer.Present()
}

// MARK: SDLリソースの解放メソッド
func (n *NameTableWindow) Close() {
	if n.texture != nil {
		n.texture.Destroy()
	}
	if n.renderer != nil {
		n.renderer.Destroy()
	}
	if n.window != nil {
		n.window.Destroy()
	}
	n.texture = nil
	n.renderer = nil
	n.window = nil
}

// MARK: ウィンドウを閉じるメソッド
func (n *NameTableWindow) requestClose() {
	if n.onClose != nil {
		n.onClose(n.ID())
	}
}
