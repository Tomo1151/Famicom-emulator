package ui

import (
	"Famicom-emulator/ppu"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// NameTableWindow visualizes the PPU name tables in a compact form.
// It displays the two 1KB name tables side-by-side as a 64x30 image
// (32x30 tiles per table) and reuses a preallocated RGB buffer.
type NameTableWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	ppu      *ppu.PPU
	buf      []byte
	onClose  func(id uint32)
}

// NewNameTableWindow creates a lightweight name-table viewer.
func NewNameTableWindow(p *ppu.PPU, scale int, onClose func(id uint32)) (*NameTableWindow, error) {
	const tileSize = 8
	const cols = 32
	const rows = 30
	const w = cols * 2 * tileSize // 512
	const h = rows * tileSize     // 240

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

	return &NameTableWindow{window: win, renderer: r, texture: t, ppu: p, buf: buf, onClose: onClose}, nil
}

func (n *NameTableWindow) ID() uint32 {
	id, _ := n.window.GetID()
	return id
}

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
			}
		}
	}
}

func (n *NameTableWindow) Update() {
	// Render each tile as an 8x8 block using CHR ROM + name table attribute palettes.
	vramPtr := n.ppu.VRAM()
	paletteTable := n.ppu.PaletteTable()
	bankBase := n.ppu.BackgroundPatternTableAddress()

	// convenience aliases
	const cols = 32
	const rows = 30
	const ntSize = 1024
	const tileSize = 8
	width := cols * 2 * tileSize

	for nt := 0; nt < 2; nt++ {
		base := nt * ntSize
		// attribute table slice
		attrStart := base + 0x3C0
		// For each tile in the name table
		for ty := 0; ty < rows; ty++ {
			for tx := 0; tx < cols; tx++ {
				idx := base + ty*cols + tx
				tileIndex := (*vramPtr)[idx]

				// compute attribute table index
				attrTableIdx := (ty/4)*8 + (tx / 4)
				attrByte := (*vramPtr)[attrStart+attrTableIdx]

				// select palette quadrant
				var paletteIdx uint8
				if tx%4/2 == 0 && ty%4/2 == 0 {
					paletteIdx = (attrByte) & 0b11
				} else if tx%4/2 == 1 && ty%4/2 == 0 {
					paletteIdx = (attrByte >> 2) & 0b11
				} else if tx%4/2 == 0 && ty%4/2 == 1 {
					paletteIdx = (attrByte >> 4) & 0b11
				} else {
					paletteIdx = (attrByte >> 6) & 0b11
				}

				paletteStart := 1 + uint(paletteIdx)*4
				// paletteIndices are indices into ppu.PALETTE
				paletteIndices := [4]uint8{
					(*paletteTable)[0],
					(*paletteTable)[paletteStart+0],
					(*paletteTable)[paletteStart+1],
					(*paletteTable)[paletteStart+2],
				}

				// fetch tile pattern (16 bytes per tile)
				tileBase := bankBase + uint16(tileIndex)*uint16(tileSize*2)

				for row := 0; row < tileSize; row++ {
					upper := n.ppu.Mapper.ReadCharacterRom(tileBase + uint16(row))
					lower := n.ppu.Mapper.ReadCharacterRom(tileBase + uint16(row) + uint16(tileSize))
					for col := 0; col < tileSize; col++ {
						// Match PPU: lower plane contributes the high bit, upper plane the low bit
						bit := (((lower >> (7 - uint(col))) & 1) << 1) | ((upper >> (7 - uint(col))) & 1)
						colorIdx := paletteIndices[bit]
						color := ppu.PALETTE[colorIdx]

						px := nt*cols*tileSize + tx*tileSize + col
						py := ty*tileSize + row
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

func (n *NameTableWindow) Render() {
	n.renderer.Clear()
	n.renderer.Copy(n.texture, nil, nil)
	n.renderer.Present()
}

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

func (n *NameTableWindow) requestClose() {
	if n.onClose != nil {
		n.onClose(n.ID())
	}
}
