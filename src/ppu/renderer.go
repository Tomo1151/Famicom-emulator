package ppu

import (
	"Famicom-emulator/cartridge/mappers"
	"fmt"
)

const (
	SCREEN_WIDTH  uint = 256
	SCREEN_HEIGHT uint = 240
	CANVAS_WIDTH   uint = SCREEN_WIDTH * 2
	CANVAS_HEIGHT  uint = 240
	TILE_SIZE     uint = 8
)

type Canvas struct {
	Width  uint
	Height uint
	Buffer [uint(CANVAS_WIDTH) * uint(CANVAS_HEIGHT)*3]byte
}

type Rect struct {
	x1 uint
	y1 uint
	x2 uint
	y2 uint
}

func (c *Canvas) Init() {
	c.Width = CANVAS_WIDTH
	c.Height = CANVAS_HEIGHT
}

func (c *Canvas) setPixelAt(x uint, y uint, palette [3]uint8) {
	if x >= c.Width || y >= c.Height { return }

	basePtr := (y * CANVAS_WIDTH + x) * 3
	c.Buffer[basePtr+0] = palette[0]  // R
	c.Buffer[basePtr+1] = palette[1]  // G
	c.Buffer[basePtr+2] = palette[2]  // B
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


func RenderSprite(ppu *PPU, canvas *Canvas) {
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
					canvas.setPixelAt(tileX + uint(x), tileY + y, rgb)
				} else if flipH && !flipV {
					canvas.setPixelAt(tileX + TILE_SIZE-1 - uint(x), tileY + y, rgb)
				} else if !flipH && flipV {
					canvas.setPixelAt(tileX + uint(x), tileY + TILE_SIZE-1 - y, rgb)
				} else if flipH && flipV {
					canvas.setPixelAt(tileX + TILE_SIZE-1 - uint(x), tileY + TILE_SIZE-1 - y, rgb)
				}
			}
		}

	}
}

func RenderNameTable(ppu *PPU, canvas *Canvas, nameTable *[]uint8, viewport Rect, shiftX int, shiftY int) {
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
					canvas.setPixelAt(uint(shiftX + int(pixelX)), uint(shiftY + int(pixelY)), PALETTE[palette[value]])
				}
			}
		}
	}
}

func RenderScanlineBackground(ppu *PPU, canvas *Canvas, scanline uint16) {
	// BGが無効であれば描画をしない
	if !ppu.mask.BackgroundEnable { return }

	// スクロール値を取得
	scrollX := uint(ppu.scroll.ScrollX)
	scrollY := uint(ppu.scroll.ScrollY)

	// 描画するY座標を計算
	globalY := scrollY + uint(scanline)
	actualY := globalY % TILE_SIZE // タイル内の何行目か

	// 画面の左端から右端まで
	for x := range SCREEN_WIDTH {
		// 描画するX座標を計算
		globalX := scrollX + x

		// 描画対象のネームテーブルを決定
		nameTable := getNameTableForPixel(ppu, globalX, globalY)

		// ネームテーブル内のタイル座標を計算
		tileX := (globalX / TILE_SIZE) % 32
		tileY := (globalY / TILE_SIZE) % 30

		// タイルのインデックスを取得
		tileIndex := uint16((nameTable)[tileY*32+tileX])

		// 属性テーブルからパレット情報を取得
		attributeTable := (nameTable)[0x3C0:0x400]
		palette := getBGPalette(ppu, &attributeTable, tileX, tileY)

		// パターンテーブルからタイルのピクセルデータを取得
		bank := ppu.control.GetBackgroundPatternTableAddress()
		tileBasePointer := bank + tileIndex * uint16(TILE_SIZE * 2)

		// 実際のY座標に対応するタイルデータを2バイト取得
		upper := ppu.Mapper.ReadCharacterROM(tileBasePointer + uint16(actualY))
		lower := ppu.Mapper.ReadCharacterROM(tileBasePointer + uint16(actualY) + uint16(TILE_SIZE))

		// 実際のY座標に対応するピクセルを計算
		actualX := (TILE_SIZE-1) - (globalX % TILE_SIZE)

		// そのピクセルの色を確定
		value := (lower >> uint8(actualX) & 1) << 1 | (upper >> uint8(actualX) & 1)

		// Canvasに描画
		canvas.setPixelAt(x, uint(scanline), PALETTE[palette[value]])
	}
}

// @FIXME 実装し直す
func RenderScanlineSprite(ppu *PPU, canvas *Canvas, scanline uint16) {
    if !ppu.mask.SpriteEnable {
        return
    }

    spriteHeight := ppu.control.GetSpriteSize()
    var secondaryOAM [8][4]uint8 // 1スキャンラインに描画するスプライト(最大8つ)
    spriteCount := 0

    // 1. スプライト評価 (OAMから現在のスキャンラインにあるスプライトを探す)
    for i := 0; i < len(ppu.oam)/4; i++ {
        spriteY := uint16(ppu.oam[i*4])

        // スプライトが現在のスキャンラインに収まっているかチェック
        if scanline >= spriteY && scanline < spriteY+uint16(spriteHeight) {
            if spriteCount < 8 {
                secondaryOAM[spriteCount][0] = ppu.oam[i*4+0] // Y
                secondaryOAM[spriteCount][1] = ppu.oam[i*4+1] // Tile Index
                secondaryOAM[spriteCount][2] = ppu.oam[i*4+2] // Attributes
                secondaryOAM[spriteCount][3] = ppu.oam[i*4+3] // X
                spriteCount++
            } else {
                // 9つ目以降のスプライトが見つかったらオーバーフローフラグを立てる
                ppu.status.SetSpriteOverflow(true)
                break
            }
        }
    }

    // 2. スプライト描画 (セカンダリOAMのスプライトを描画)
    // OAMのインデックスが小さいものが優先される(手前に描画)ため、逆順にループする
    for i := spriteCount - 1; i >= 0; i-- {
        sprite := secondaryOAM[i]
        spriteY := uint16(sprite[0])
        tileIndex := uint16(sprite[1])
        attributes := sprite[2]
        spriteX := uint(sprite[3])

        flipV := (attributes >> 7) & 1 == 1
        flipH := (attributes >> 6) & 1 == 1
        paletteIndex := attributes & 0b11
        spritePalette := getSpritePalette(ppu, paletteIndex)

        // スプライトタイルの何行目を描画するか
        var fineY uint16
        if flipV {
            fineY = (spriteY + uint16(spriteHeight-1)) - scanline
        } else {
            fineY = scanline - spriteY
        }

        var bank uint16
        if spriteHeight == 8 { // 8x8モード
            bank = ppu.control.GetSpritePatternTableAddress()
        } else { // 8x16モード
            bank = (tileIndex & 1) * 0x1000
            tileIndex &= 0xFE
        }

        // 8x16モードでタイルの下半分の場合
        if fineY >= 8 {
            tileIndex++
            fineY -= 8
        }

        tileBasePtr := bank + tileIndex*16
        upper := ppu.Mapper.ReadCharacterROM(tileBasePtr + fineY)
        lower := ppu.Mapper.ReadCharacterROM(tileBasePtr + fineY + 8)

        for x := 0; x < 8; x++ {
            var value uint8
            if flipH {
                value = (lower&1)<<1 | (upper & 1)
                upper >>= 1
                lower >>= 1
            } else {
                value = ((lower>>7)&1)<<1 | ((upper>>7) & 1)
                upper <<= 1
                lower <<= 1
            }

            if value == 0 {
                continue
            }

            pixelX := spriteX + uint(x)
            if pixelX >= SCREEN_WIDTH {
                continue
            }

            // 背景が透明な場合のみ描画
            // @TODO: 優先度(Attributeのbit5)の処理を追加
            rgb := PALETTE[spritePalette[value]]
            canvas.setPixelAt(pixelX, uint(scanline), rgb)
        }
    }
}

func RenderScanline(ppu *PPU, canvas *Canvas, scanline uint16) {
	RenderScanlineBackground(ppu, canvas, scanline)
	RenderScanlineSprite(ppu, canvas, scanline)
}

func Render(ppu *PPU, canvas *Canvas) {
	scrollX := uint(ppu.scroll.ScrollX)
	scrollY := uint(ppu.scroll.ScrollY)

	var nameTable0, nameTable1, nameTable2, nameTable3 = getNameTables(ppu)

	// 左上
	RenderNameTable(
		ppu,
		canvas,
		nameTable0,
		Rect{scrollX, scrollY, SCREEN_WIDTH, SCREEN_HEIGHT},
		-int(scrollX),
		-int(scrollY),
	)

	// 右上
	RenderNameTable(
		ppu,
		canvas,
		nameTable1,
		Rect{0, scrollY, scrollX, SCREEN_HEIGHT},
		int(SCREEN_WIDTH - scrollX),
		-int(scrollY),
	)

	// 左下
	RenderNameTable(
		ppu,
		canvas,
		nameTable2,
		Rect{scrollX, 0, SCREEN_WIDTH, scrollY},
		-int(scrollX),
		int(SCREEN_HEIGHT - scrollY),
	)

	// 右下
	RenderNameTable(
		ppu,
		canvas,
		nameTable3,
		Rect{0, 0, scrollX, scrollY},
		int(SCREEN_WIDTH - scrollX),
		int(SCREEN_HEIGHT - scrollY),
	)

	// RenderBackground(ppu, canvas)
	RenderSprite(ppu, canvas)
}

func getNameTables(ppu *PPU) (*[]uint8, *[]uint8, *[]uint8, *[]uint8) {
	var nameTables [4](*[]uint8)
	primaryNameTable := ppu.vram[0x000:0x400]
	secondaryNameTable := ppu.vram[0x400:0x800]
	mirroring := ppu.Mapper.GetMirroring()

	if (mirroring == mappers.MIRRORING_VERTICAL &&
		(ppu.control.GetBaseNameTableAddress() == 0x2000 ||
		ppu.control.GetBaseNameTableAddress() == 0x2800)) {
		nameTables[0] = &primaryNameTable
		nameTables[1] = &secondaryNameTable
		nameTables[2] = &primaryNameTable
		nameTables[3] = &secondaryNameTable
	} else if (mirroring == mappers.MIRRORING_HORIZONTAL &&
		(ppu.control.GetBaseNameTableAddress() == 0x2000 ||
		ppu.control.GetBaseNameTableAddress() == 0x2400)) {
		nameTables[0] = &primaryNameTable
		nameTables[1] = &primaryNameTable
		nameTables[2] = &secondaryNameTable
		nameTables[3] = &secondaryNameTable
	} else if (mirroring == mappers.MIRRORING_VERTICAL &&
		(ppu.control.GetBaseNameTableAddress() == 0x2400 ||
		ppu.control.GetBaseNameTableAddress() == 0x2C00)) {
		nameTables[0] = &secondaryNameTable
		nameTables[1] = &primaryNameTable
		nameTables[2] = &secondaryNameTable
		nameTables[3] = &primaryNameTable
	} else if (mirroring == mappers.MIRRORING_HORIZONTAL &&
		(ppu.control.GetBaseNameTableAddress() == 0x2800 ||
		ppu.control.GetBaseNameTableAddress() == 0x2C00)) {
		nameTables[0] = &secondaryNameTable
		nameTables[1] = &secondaryNameTable
		nameTables[2] = &primaryNameTable
		nameTables[3] = &primaryNameTable
	} else {
		panic("Error: unexpected name table pattern")
	}
	return nameTables[0], nameTables[1], nameTables[2], nameTables[3]
}

func getNameTableForPixel(ppu *PPU, x uint, y uint) []uint8 {
	mirroring := ppu.Mapper.GetMirroring()
	baseNameTableAddress := ppu.control.GetBaseNameTableAddress()
	primaryNameTable := ppu.vram[0x000:0x400]
	secondaryNameTable := ppu.vram[0x400:0x800]

	// 4画面を繋げたうちどの画面にピクセルがあるかを判定
	isRight := (x % (SCREEN_WIDTH*2)) >= SCREEN_WIDTH
	isBottom := (y % (SCREEN_HEIGHT*2)) >= SCREEN_HEIGHT

	var vNameTableIndex uint
	if !isBottom && !isRight {
		vNameTableIndex = 0 // 左上
	} else if !isBottom && isRight {
		vNameTableIndex = 1 // 右上
	} else if isBottom && !isRight {
		vNameTableIndex = 2 // 左下
	} else {
		vNameTableIndex = 3 // 右下
	}

	// 基準ネームテーブルアドレスとミラーリングから実際に使用するテーブルを判定
	var nameTableIndex uint
	switch baseNameTableAddress {
	case 0x2000:
		nameTableIndex = 0
	case 0x2400:
		nameTableIndex = 1
	case 0x2800:
		nameTableIndex = 2
	case 0x2C00:
		nameTableIndex = 3
	}

	// 仮想のテーブルインデックスと基準アドレスから最終的なテーブルのインデックスを計算
	index := (vNameTableIndex + nameTableIndex) % 4

	switch mirroring {
	case mappers.MIRRORING_VERTICAL:
		if index == 0 || index == 2 {
			return primaryNameTable
		} else {
			return secondaryNameTable
		}
	case mappers.MIRRORING_HORIZONTAL:
		if index == 0 || index == 1 {
			return primaryNameTable
		} else {
			return secondaryNameTable
		}
	default:
		// @FIXME FourScreenの対応
		return primaryNameTable
	}
}

func DumpCanvasWithASCII(canvas Canvas) {
	for y := range CANVAS_HEIGHT-1 {
		for x := range CANVAS_WIDTH-1 {
			color := canvas.Buffer[y*CANVAS_WIDTH+x]

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
