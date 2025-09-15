package ppu

import (
	"Famicom-emulator/cartridge/mappers"
	"fmt"
)

const (
	SCREEN_WIDTH  uint = 256
	SCREEN_HEIGHT uint = 240
	FRAME_WIDTH   uint = SCREEN_WIDTH * 2
	FRAME_HEIGHT  uint = 240
	TILE_SIZE     uint = 8
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
		paletteIdx = (attrByte) & 0b11
	} else if tileColumn % 4 / 2 == 1 && tileRow % 4 / 2 == 0 {
		paletteIdx = (attrByte >> 2) & 0b11
	} else if tileColumn % 4 / 2 == 0 && tileRow % 4 / 2 == 1 {
		paletteIdx = (attrByte >> 4) & 0b11
	} else if tileColumn % 4 / 2 == 1 && tileRow % 4 / 2 == 1 {
		paletteIdx = (attrByte >> 6) & 0b11
	} else {
		panic("Error: unexpected palette value")
	}

	var paletteStart uint = 1 + uint(paletteIdx) * 4
	color := [4]uint8{
		ppu.PaletteTable[0],
		ppu.PaletteTable[paletteStart+0],
		ppu.PaletteTable[paletteStart+1],
		ppu.PaletteTable[paletteStart+2],
	}

	return color
}

func getSpritePalette(ppu *PPU, paletteIndex uint8) [4]uint8 {
	var start uint = 0x11 + uint(paletteIndex * 4)
	return [4]uint8{
		0,
		ppu.PaletteTable[start + 0],
		ppu.PaletteTable[start + 1],
		ppu.PaletteTable[start + 2],
	}
}


func RenderSprite(ppu *PPU, frame *Frame) {
	for i := len(ppu.oam) - 4; i >= 0; i -= 4 {
		tileIndex := uint16(ppu.oam[i + 1])
		tileX := uint(ppu.oam[i + 3])
		tileY := uint(ppu.oam[i])
		flipV := (ppu.oam[i + 2] >> 7) & 1 == 1
		flipH := (ppu.oam[i + 2] >> 6) & 1 == 1
		palleteIndex := ppu.oam[i + 2] & 0b11
		spritePalette := getSpritePalette(ppu, palleteIndex)

		bank := ppu.control.GetSpritePatternTableAddress()
		tileBasePtr :=(bank+tileIndex*16)

		var tile [TILE_SIZE*2]uint8
		for j := range TILE_SIZE*2 {
			tile[j] = ppu.Mapper.ReadCharacterROM(tileBasePtr+uint16(j))
		}

		for y := range TILE_SIZE {
			upper := tile[y]
			lower := tile[y+TILE_SIZE]

			for x := int(TILE_SIZE)-1; x >= 0; x-- {
				value := (1 & lower) << 1 | (1 & upper)
				upper >>= 1
				lower >>= 1

				if value == 0 { continue }

				rgb := PALETTE[spritePalette[value]]

				if !flipH && !flipV {
					frame.setPixelAt(tileX + uint(x), tileY + y, rgb)
				} else if flipH && !flipV {
					frame.setPixelAt(tileX + TILE_SIZE-1 - uint(x), tileY + y, rgb)
				} else if !flipH && flipV {
					frame.setPixelAt(tileX + uint(x), tileY + TILE_SIZE-1 - y, rgb)
				} else if flipH && flipV {
					frame.setPixelAt(tileX + TILE_SIZE-1 - uint(x), tileY + TILE_SIZE-1 - y, rgb)
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
		tileBasePtr := (bank+tileIndex*16)
		palette := getBGPalette(ppu, &attrributeTable, tileX, tileY)

		var tile [TILE_SIZE*2]uint8
		for j := range TILE_SIZE*2 {
			tile[j] = ppu.Mapper.ReadCharacterROM(tileBasePtr+uint16(j))
		}

		for y := range TILE_SIZE {
			upper := tile[y]
			lower := tile[y+TILE_SIZE]

			for x := int(TILE_SIZE-1); x >= 0; x-- {
				value := (1 & lower) << 1 | (1 & upper)
				upper >>= 1
				lower >>= 1

				pixelX := tileX * TILE_SIZE + uint(x)
				pixelY := tileY * TILE_SIZE + uint(y)

				if pixelX >= viewport.x1 && pixelX < viewport.x2 && pixelY >= viewport.y1 && pixelY < viewport.y2 {
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
	mirroring := ppu.Mapper.GetMirroring()

	if (mirroring == mappers.MIRRORING_VERTICAL &&
		(ppu.control.GetBaseNameTableAddress() == 0x2000 ||
		ppu.control.GetBaseNameTableAddress() == 0x2800)) {
		primaryNameTable = ppu.vram[0x000:0x400]
		secondaryNameTable = ppu.vram[0x400:0x800]
	} else if (mirroring == mappers.MIRRORING_HORIZONTAL &&
		(ppu.control.GetBackgroundPatternTableAddress() == 0x2000 ||
		ppu.control.GetBaseNameTableAddress() == 0x2400)) {
		primaryNameTable = ppu.vram[0x000:0x400]
		secondaryNameTable = ppu.vram[0x400:0x800]
	} else if (mirroring == mappers.MIRRORING_VERTICAL &&
		(ppu.control.GetBaseNameTableAddress() == 0x2400 ||
		ppu.control.GetBaseNameTableAddress() == 0x2C00)) {
		primaryNameTable = ppu.vram[0x400:0x800]
		secondaryNameTable = ppu.vram[0x000:0x400]
	} else if (mirroring == mappers.MIRRORING_HORIZONTAL &&
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
