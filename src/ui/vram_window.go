package ui

import (
	"Famicom-emulator/cartridge/mappers"
	"Famicom-emulator/ppu"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: 定数定義
const (
	TILE_COLUMNS = 32
	TILE_ROWS    = 30
)

// MARK: NameTableWindow の定義
type NameTableWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	ppu      *ppu.PPU
	buffer   []byte
	onClose  func(id uint32)
	baseW    int
	baseH    int
	scale    int
}

// MARK: NameTableWindow の作成メソッド
func NewNameTableWindow(p *ppu.PPU, scale int, onClose func(id uint32)) (*NameTableWindow, error) {
	const width = int(TILE_COLUMNS * 2 * ppu.TILE_SIZE) // 512px: 2画面分
	const height = int(TILE_ROWS * 2 * ppu.TILE_SIZE)   // 480px: 2画面分
	win, err := sdl.CreateWindow(
		"Name Table Viewer",
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

	return &NameTableWindow{window: win, renderer: r, texture: t, ppu: p, buffer: buffer, onClose: onClose, baseW: width, baseH: height, scale: scale}, nil
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
			case sdl.K_PLUS, sdl.K_SEMICOLON:
				n.setScale(n.scale + 1)
			case sdl.K_MINUS:
				n.setScale(n.scale - 1)
			}
		}
	}
}

// MARK: スケール設定メソッド
func (n *NameTableWindow) setScale(s int) {
	// 1 ~ 8 の間に設定
	s = min(max(s, 1), 8)
	if s == n.scale {
		return
	}
	n.scale = s
	if n.window != nil {
		n.window.SetSize(int32(n.baseW*n.scale), int32(n.baseH*n.scale))
	}
}

// MARK: 更新メソッド
func (n *NameTableWindow) Update() {
	vramPtr := n.ppu.VRAM()

	const cols uint = 32
	const rows uint = 30
	const ntSize uint = 1024
	width := cols * 2 * ppu.TILE_SIZE

	// フレーム最初のマッパーでミラーリングを判別
	mapper := n.ppu.GetMapperForScanline(0)
	mirroring := mapper.Mirroring()
	bankBase := n.ppu.BackgroundPatternTableAddress()

	// 4つのネームテーブルを描画
	for nt := range 4 {
		xOffset := uint(nt%2) * cols * ppu.TILE_SIZE
		yOffset := uint(nt/2) * rows * ppu.TILE_SIZE

		var physicalPage uint
		switch mirroring {
		case mappers.MIRRORING_VERTICAL:
			physicalPage = uint(nt % 2) // 0,1,0,1
		case mappers.MIRRORING_HORIZONTAL:
			physicalPage = uint(nt / 2) // 0,0,1,1
		default:
			physicalPage = uint(nt % 2)
		}

		base := physicalPage * ntSize
		attrStart := base + 0x3C0
		attributeTable := (*vramPtr)[int(attrStart):int(attrStart+0x40)]

		for ty := range rows {
			for tx := range cols {
				idx := base + ty*cols + tx
				tileIndex := (*vramPtr)[idx]
				palette := n.ppu.BackgroundColorPalette(&attributeTable, tx, ty)
				tileBase := bankBase + uint16(tileIndex)*uint16(ppu.TILE_SIZE*2)

				// 各タイルの描画
				for row := range ppu.TILE_SIZE {
					upper := mapper.ReadCharacterRom(tileBase + uint16(row))
					lower := mapper.ReadCharacterRom(tileBase + uint16(row) + uint16(ppu.TILE_SIZE))

					for col := range ppu.TILE_SIZE {
						bit := (((lower >> (7 - uint(col))) & 1) << 1) | ((upper >> (7 - uint(col))) & 1)
						colorIdx := palette[bit]
						color := ppu.PALETTE[colorIdx]

						px := xOffset + tx*ppu.TILE_SIZE + col
						py := yOffset + ty*ppu.TILE_SIZE + row
						pos := (py*width + px) * 3
						n.buffer[pos+0] = color[0]
						n.buffer[pos+1] = color[1]
						n.buffer[pos+2] = color[2]
					}
				}
			}
		}
	}

	n.texture.Update(nil, unsafe.Pointer(&n.buffer[0]), int(width*3))
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
