package ppu

import (
	"fmt"
)

const (
	FRAME_WIDTH  uint = 256
	FRAME_HEIGHT uint = 240
	TILE_SIZE    uint = 8
)

var (
	PALETTE = [64][3]uint8{
		{0x80, 0x80, 0x80}, {0x00, 0x3D, 0xA6}, {0x00, 0x12, 0xB0}, {0x44, 0x00, 0x96}, {0xA1, 0x00, 0x5E},
		{0xC7, 0x00, 0x28}, {0xBA, 0x06, 0x00}, {0x8C, 0x17, 0x00}, {0x5C, 0x2F, 0x00}, {0x10, 0x45, 0x00},
		{0x05, 0x4A, 0x00}, {0x00, 0x47, 0x2E}, {0x00, 0x41, 0x66}, {0x00, 0x00, 0x00}, {0x05, 0x05, 0x05},
		{0x05, 0x05, 0x05}, {0xC7, 0xC7, 0xC7}, {0x00, 0x77, 0xFF}, {0x21, 0x55, 0xFF}, {0x82, 0x37, 0xFA},
		{0xEB, 0x2F, 0xB5}, {0xFF, 0x29, 0x50}, {0xFF, 0x22, 0x00}, {0xD6, 0x32, 0x00}, {0xC4, 0x62, 0x00},
		{0x35, 0x80, 0x00}, {0x05, 0x8F, 0x00}, {0x00, 0x8A, 0x55}, {0x00, 0x99, 0xCC}, {0x21, 0x21, 0x21},
		{0x09, 0x09, 0x09}, {0x09, 0x09, 0x09}, {0xFF, 0xFF, 0xFF}, {0x0F, 0xD7, 0xFF}, {0x69, 0xA2, 0xFF},
		{0xD4, 0x80, 0xFF}, {0xFF, 0x45, 0xF3}, {0xFF, 0x61, 0x8B}, {0xFF, 0x88, 0x33}, {0xFF, 0x9C, 0x12},
		{0xFA, 0xBC, 0x20}, {0x9F, 0xE3, 0x0E}, {0x2B, 0xF0, 0x35}, {0x0C, 0xF0, 0xA4}, {0x05, 0xFB, 0xFF},
		{0x5E, 0x5E, 0x5E}, {0x0D, 0x0D, 0x0D}, {0x0D, 0x0D, 0x0D}, {0xFF, 0xFF, 0xFF}, {0xA6, 0xFC, 0xFF},
		{0xB3, 0xEC, 0xFF}, {0xDA, 0xAB, 0xEB}, {0xFF, 0xA8, 0xF9}, {0xFF, 0xAB, 0xB3}, {0xFF, 0xD2, 0xB0},
		{0xFF, 0xEF, 0xA6}, {0xFF, 0xF7, 0x9C}, {0xD7, 0xE8, 0x95}, {0xA6, 0xED, 0xAF}, {0xA2, 0xF2, 0xDA},
		{0x99, 0xFF, 0xFC}, {0xDD, 0xDD, 0xDD}, {0x11, 0x11, 0x11}, {0x11, 0x11, 0x11},
	}
)

type Frame struct {
	Width  uint
	Height uint
	Buffer [uint(FRAME_WIDTH) * uint(FRAME_HEIGHT)*3]byte
}

type Tile struct {
	Pixels [TILE_SIZE][TILE_SIZE]uint8
}

func (f *Frame) Init() {
	f.Width = FRAME_WIDTH
	f.Height = FRAME_HEIGHT
}

func (f *Frame) setPixelAt(x uint, y uint, palette [3]uint8) {
	basePtr := (y * FRAME_WIDTH + x) * 3
	f.Buffer[basePtr+0] = palette[0]  // R
	f.Buffer[basePtr+1] = palette[1]  // G
	f.Buffer[basePtr+2] = palette[2]  // B
}

func (f *Frame) SetTileAt(tileIndex uint, tile Tile) {
	var ox uint = (tileIndex % 32) * uint(TILE_SIZE)
	var oy uint = (tileIndex / 32) * uint(TILE_SIZE)

	fmt.Printf("set starts at (%d, %d)\n", ox, oy)

	DumpTile(tile)

	for y := range TILE_SIZE {
		for x := range TILE_SIZE {
			pixelY := oy + y
			pixelX := ox + x

			if pixelY >= FRAME_HEIGHT || pixelX >= FRAME_WIDTH { continue }
			index := pixelY * FRAME_WIDTH + pixelX
			if uint(index) >= uint(len(f.Buffer)) {
					fmt.Printf("Warning: index %d exceeds buffer size %d\n", index, len(f.Buffer))
					continue
			}

			bufferIndex := (pixelY * FRAME_WIDTH + pixelX) * 3
			var colorPallette [4][3]uint8 = [4][3]uint8{
				PALETTE[0x01],
				PALETTE[0x16],
				PALETTE[0x27],
				PALETTE[0x18],
			}
			color := tile.Pixels[y][x]
			f.Buffer[bufferIndex+0] = colorPallette[color][0] // R
			f.Buffer[bufferIndex+1] = colorPallette[color][1] // G
			f.Buffer[bufferIndex+2] = colorPallette[color][2] // B
		}
	}
}


func GetTile(tileData []byte) Tile {
	if len(tileData) != 16 {
		panic("Error: invalid tile data size")
	}

	tile := Tile{}

	for y := range TILE_SIZE {
		lower := tileData[y]
		upper := tileData[y+8]

		for x := range TILE_SIZE {
			bit0 := (lower >> (7 - x)) & 1
			bit1 := (upper >> (7 - x)) & 1
			color := (bit1 << 1) | bit0
			tile.Pixels[y][x] = color
		}
	}

	return tile
}

func DumpFrame(frame Frame) {
	for y := range FRAME_HEIGHT-1 {
		for x := range FRAME_WIDTH-1 {
			color := frame.Buffer[y*FRAME_WIDTH+x]

			switch color {
			case 0:
				fmt.Print(". ")
			case 1:
				fmt.Print(": ")
			case 2:
				fmt.Print("* ")
			case 3:
				fmt.Print("# ")
			}
		}
		fmt.Println()
	}
}

func DumpTile(tile Tile) {
	for y := range TILE_SIZE {
		for x := range TILE_SIZE {
			color := tile.Pixels[y][x]

			switch color {
			case 0:
				fmt.Print(". ")
			case 1:
				fmt.Print(": ")
			case 2:
				fmt.Print("* ")
			case 3:
				fmt.Print("# ")
			}
		}
		fmt.Println()
	}
}

func getColorFromPalette(ppu PPU, tileColumn uint, tileRow uint) [4]uint8 {
	attrTableIdx := tileRow / 4 * TILE_SIZE + tileColumn / 4
	attrByte := ppu.vram[0x3C0 + attrTableIdx]

	var paletteIdx uint8
	if tileColumn % 4 / 2 == 0 && tileRow % 4 / 2 == 0 {
		paletteIdx = (attrByte >> 0) & 0b11
	} else if tileColumn % 4 / 2 == 1 && tileRow % 4 / 2 == 0 {
		paletteIdx = (attrByte >> 2) & 0b11
	} else if tileColumn % 4 / 2 == 0 && tileRow % 4 / 2 == 1 {
		paletteIdx = (attrByte >> 4) & 0b11
	} else if tileColumn % 4 / 2 == 1 && tileRow % 4 / 2 == 1 {
		paletteIdx = (attrByte >> 6) & 0b11
	} else {
		panic("Error: unexpected palette value")
	}

	paletteStart := 1 + paletteIdx * 4
	color := [4]uint8{
		ppu.PaletteTable[0],
		ppu.PaletteTable[paletteStart+0],
		ppu.PaletteTable[paletteStart+1],
		ppu.PaletteTable[paletteStart+2],
	}

	return color
}

func Render(ppu PPU, frame *Frame) {
	var bank uint16
	if ppu.control.BackgroundPatternAddress {
		bank = 0x1000
	} else {
		bank = 0x0000
	}

	for i := range 0x03C0 {
		tileIndex := uint16(ppu.vram[i])
		tileX := uint(i % 32)
		tileY := uint(i / 32)
		tileBasePtr :=(bank+tileIndex*16)
		tile := ppu.CHR_ROM[tileBasePtr:tileBasePtr+16]
		palette := getColorFromPalette(ppu, tileX, tileY)

		for y := range TILE_SIZE {
			upper := tile[y]
			lower := tile[y+TILE_SIZE]

			for x := range TILE_SIZE {
				bit0 := (lower >> (7 - x)) & 1
				bit1 := (upper >> (7 - x)) & 1
				value := (bit1 << 1) | bit0
				frame.setPixelAt(tileX*TILE_SIZE + x, tileY*TILE_SIZE+y, PALETTE[palette[value]])
			}
		}
	}

}