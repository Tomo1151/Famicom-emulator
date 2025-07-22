package ppu

import (
	"Famicom-emulator/cartridge"
	"fmt"
)

const (
	SCREEN_WIDTH  uint = 256
	SCREEN_HEIGHT uint = 240
	FRAME_WIDTH   uint = SCREEN_WIDTH * 2
	FRAME_HEIGHT  uint = 240
	TILE_SIZE     uint = 8
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

type Rect struct {
	x1 uint
	y1 uint
	x2 uint
	y2 uint
}

func (f *Frame) Init() {
	f.Width = FRAME_WIDTH
	f.Height = FRAME_HEIGHT
}

func (f *Frame) setPixelAt(x uint, y uint, palette [3]uint8) {
	if x >= f.Width || y >= f.Height { return }

	basePtr := (y * FRAME_WIDTH + x) * 3
	f.Buffer[basePtr+0] = palette[0]  // R
	f.Buffer[basePtr+1] = palette[1]  // G
	f.Buffer[basePtr+2] = palette[2]  // B
}

func getBGPalette(ppu *PPU, attrributeTable *[]uint8, tileColumn uint, tileRow uint) [4]uint8 {
	attrTableIdx := tileRow / 4 * TILE_SIZE + tileColumn / 4
	attrByte := (*attrributeTable)[attrTableIdx]

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

func getSpritePalette(ppu *PPU, paletteIndex uint8) [4]uint8 {
	start := 0x11 + (paletteIndex * 4)
	return [4]uint8{
		0,
		ppu.PaletteTable[start + 0],
		ppu.PaletteTable[start + 1],
		ppu.PaletteTable[start + 2],
	}
}

func RenderBackground(ppu *PPU, frame *Frame) {
	bank := ppu.control.GetBackgroundPatternTableAddress()

	for i := range 0x03C0 {
		tileIndex := uint16(ppu.vram[i])
		tileX := uint(i % 32)
		tileY := uint(i / 32)
		tileBasePtr :=(bank+tileIndex*16)
		tile := ppu.CHR_ROM[tileBasePtr:tileBasePtr+16]
		attrTable := ppu.vram[0x3C0:0x400]
		palette := getBGPalette(ppu, &attrTable, tileX, tileY)

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

func RenderSprite(ppu *PPU, frame *Frame) {
	// fmt.Println(ppu.oam)
	for i := len(ppu.oam) - 4; i >= 0; i -= 4 {
		// fmt.Println("Sprite: ", i, "rendered.")
		tileIndex := uint16(ppu.oam[i + 1])
		tileX := uint(ppu.oam[i + 3])
		tileY := uint(ppu.oam[i])
		flipV := (ppu.oam[i + 2] >> 7) & 1 == 1
		flipH := (ppu.oam[i + 2] >> 6) & 1 == 1
		palleteIndex := ppu.oam[i + 2] & 0b11
		spritePalette := getSpritePalette(ppu, palleteIndex)

		bank := ppu.control.GetSpritePatternTableAddress()
		tileBasePtr :=(bank+tileIndex*16)
		tile := ppu.CHR_ROM[tileBasePtr:tileBasePtr+16]

		for y := range TILE_SIZE {
			upper := tile[y]
			lower := tile[y+TILE_SIZE]

			for x := range TILE_SIZE {
				bit0 := (lower >> (7 - x)) & 1
				bit1 := (upper >> (7 - x)) & 1
				value := (bit1 << 1) | bit0
				if value == 0 { continue }

				rgb := PALETTE[spritePalette[value]]
				// if y == 0 && x == 0 {
				// 	fmt.Printf("rgb: %02X", spritePalette[value])
				// }

				if !flipH && !flipV {
					frame.setPixelAt(tileX + x, tileY + y, rgb)
				} else if flipH && !flipV {
					frame.setPixelAt(tileX + TILE_SIZE-1 - x, tileY + y, rgb)
				} else if !flipH && flipV {
					frame.setPixelAt(tileX + x, tileY + TILE_SIZE-1 - y, rgb)
				} else if flipH && flipV {
					frame.setPixelAt(tileX + TILE_SIZE-1 - x, tileY + TILE_SIZE-1 - y, rgb)
				}
			}
		}

	}
}

func RenderNameTable(ppu *PPU, frame *Frame, nameTable *[]uint8, viewport Rect, shiftX int, shiftY int) {
	bank := ppu.control.GetBackgroundPatternTableAddress()
	attrributeTable := (*nameTable)[0x3C0:0x400]

	for i := range 0x3C0 {
		tileIndex := uint16((*nameTable)[i])
		tileX := uint(i % 32)
		tileY := uint(i / 32)
		tileBasePtr :=(bank+tileIndex*16)
		tile := ppu.CHR_ROM[tileBasePtr:tileBasePtr+16]
		palette := getBGPalette(ppu, &attrributeTable, tileX, tileY)

		for y := range TILE_SIZE {
			upper := tile[y]
			lower := tile[y+TILE_SIZE]

			for x := range TILE_SIZE {
				bit0 := (lower >> (7 - x)) & 1
				bit1 := (upper >> (7 - x)) & 1
				value := (bit1 << 1) | bit0

				pixelX := tileX * TILE_SIZE + x
				pixelY := tileY * TILE_SIZE + y

				if pixelX >= uint(viewport.x1) && pixelX < uint(viewport.x2) && pixelY >= uint(viewport.y1) && pixelY < uint(viewport.y2) {
					frame.setPixelAt(uint(shiftX + int(pixelX)), uint(shiftY + int(pixelY)), PALETTE[palette[value]])
				}
			}
		}
	}
}

func Render(ppu *PPU, frame *Frame) {
	scrollX := uint(ppu.scroll.ScrollX)
	scrollY := uint(ppu.scroll.ScrollY)

	var primary, secondary = getNameTables(ppu)

	RenderNameTable(
		ppu,
		frame,
		primary,
		Rect{scrollX, scrollY, SCREEN_WIDTH, SCREEN_HEIGHT},
		-int(scrollX),
		-int(scrollY),
	)

	if scrollX > 0 {
		RenderNameTable(
			ppu,
			frame,
			secondary,
			Rect{0, 0, scrollX, SCREEN_HEIGHT},
			int(SCREEN_WIDTH - scrollX),
			0,
		)
	} else if scrollY > 0 {
		RenderNameTable(
			ppu,
			frame,
			secondary,
			Rect{0, 0, SCREEN_WIDTH, scrollY},
			0,
			int(SCREEN_HEIGHT - scrollY),
		)
	}

	// RenderBackground(ppu, frame)
	RenderSprite(ppu, frame)
}

func getNameTables(ppu *PPU) (*[]uint8, *[]uint8) {
	var primaryNameTable []uint8
	var secondaryNameTable []uint8
	if (ppu.Mirroring == cartridge.MIRRORING_VERTICAL &&
		(ppu.control.GetBaseNameTableAddress() == 0x2000 ||
		ppu.control.GetBaseNameTableAddress() == 0x2800)) {
		primaryNameTable = ppu.vram[0x000:0x400]
		secondaryNameTable = ppu.vram[0x400:0x800]
	} else if (ppu.Mirroring == cartridge.MIRRORING_HORIZONTAL &&
		(ppu.control.GetBackgroundPatternTableAddress() == 0x2000 ||
		ppu.control.GetBaseNameTableAddress() == 0x2400)) {
		primaryNameTable = ppu.vram[0x000:0x400]
		secondaryNameTable = ppu.vram[0x400:0x800]
	} else if (ppu.Mirroring == cartridge.MIRRORING_VERTICAL &&
		(ppu.control.GetBaseNameTableAddress() == 0x2400 ||
		ppu.control.GetBaseNameTableAddress() == 0x2C00)) {
		primaryNameTable = ppu.vram[0x400:0x800]
		secondaryNameTable = ppu.vram[0x000:0x400]
	} else if (ppu.Mirroring == cartridge.MIRRORING_HORIZONTAL &&
		(ppu.control.GetBackgroundPatternTableAddress() == 0x2800 ||
		ppu.control.GetBaseNameTableAddress() == 0x2C00)) {
		primaryNameTable = ppu.vram[0x400:0x800]
		secondaryNameTable = ppu.vram[0x000:0x400]
	} else {
		primaryNameTable = ppu.vram[0x000:0x400]
		secondaryNameTable = ppu.vram[0x400:0x800]
	}
	return &primaryNameTable, &secondaryNameTable
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
