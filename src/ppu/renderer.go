package ppu

import (
	"Famicom-emulator/cartridge/mappers"
)

const (
	SCREEN_WIDTH    uint = 256
	SCREEN_HEIGHT   uint = 240
	CANVAS_WIDTH    uint = SCREEN_WIDTH * 2
	CANVAS_HEIGHT   uint = 240
	OAM_SPRITE_SIZE uint = 4
	OAM_SPRITE_X    uint = 3
	OAM_SPRITE_Y    uint = 0
	OAM_SPRITE_TILE uint = 1
	OAM_SPRITE_ATTR uint = 2
	SPRITE_MAX      uint = 8
	TILE_SIZE       uint = 8
)

// MARK: Canvasの定義
type Canvas struct {
	Width  uint
	Height uint
	Buffer [uint(CANVAS_WIDTH) * uint(CANVAS_HEIGHT)*3]byte
}

// MARK: キャンバスの初期化メソッド
func (c *Canvas) Init() {
	c.Width = CANVAS_WIDTH
	c.Height = CANVAS_HEIGHT
}

// MARK: キャンバスの指定した座標に色をセット
func (c *Canvas) setPixelAt(x uint, y uint, palette [3]uint8) {
	if x >= c.Width || y >= c.Height { return }

	basePtr := (y * CANVAS_WIDTH + x) * 3
	c.Buffer[basePtr+0] = palette[0]  // R
	c.Buffer[basePtr+1] = palette[1]  // G
	c.Buffer[basePtr+2] = palette[2]  // B
}

// MARK: キャンバスの指定した座標が透明ピクセルかを判別
func (c *Canvas) isBgTransparentAt(ppu *PPU, x uint, y uint) bool {
	basePtr := (y * CANVAS_WIDTH + x) * 3
	isBgTransparent := c.Buffer[basePtr+0] == PALETTE[ppu.PaletteTable[0]][0] && c.Buffer[basePtr+1] == PALETTE[ppu.PaletteTable[0]][1] && c.Buffer[basePtr+2] == PALETTE[ppu.PaletteTable[0]][2]
	return isBgTransparent
}

// MARK: Rectの定義
type Rect struct {
	x1 uint
	y1 uint
	x2 uint
	y2 uint
}

// MARK: BG面のカラーパレットを取得
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

// MARK: スプライトのカラーパレットを取得
func getSpritePalette(ppu *PPU, paletteIndex uint8) [4]uint8 {
	var start uint = 0x11 + uint(paletteIndex * 4)
	return [4]uint8{
		0,
		ppu.PaletteTable[start + 0],
		ppu.PaletteTable[start + 1],
		ppu.PaletteTable[start + 2],
	}
}


// MARK: 指定したスキャンラインのBG面を描画
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

// MARK: 指定したスキャンラインのスプライトを描画
func RenderScanlineSprite(ppu *PPU, canvas *Canvas, scanline uint16) {
	// スプライトが無効であれば描画しない
	if !ppu.mask.SpriteEnable { return }

	// スプライトサイズの取得 (8 / 16)
	spriteHeight := ppu.control.GetSpriteSize()
	spriteCount, sprites := FindScanlineSprite(ppu, spriteHeight, scanline)

	// スプライトの描画
	for i := range spriteCount {
		// 逆順に評価する (重なり順のため)
		index := (spriteCount-1) - i

		/*
			タイル属性
			bit 76543210
					VHP...CC

			V: 垂直反転
			H: 水平反転
			P: 優先度 (0:前面, 1:背面)
			C: パレット
		*/

		// 描画するスプライトを取得
		sprite := sprites[index]
		spriteY := uint16(sprite[OAM_SPRITE_Y])
		spriteX := uint16(sprite[OAM_SPRITE_X])
		tileIndex := uint16(sprite[OAM_SPRITE_TILE])
		attributes := sprite[OAM_SPRITE_ATTR]
		priority := (attributes >> 5) & 1

		flipV := (attributes >> 7) & 1 == 1
		flipH := (attributes >> 6) & 1 == 1
		paletteIndex := attributes & 0b11
		palette := getSpritePalette(ppu, paletteIndex)

		// スプライトの何行目を描画するかを判定
		var tileY uint16
		if flipV {
			tileY = (spriteY + uint16(spriteHeight-1)) - scanline
		} else {
			tileY = scanline - spriteY
		}

		var bank uint16
		if spriteHeight == 8 {
			/*
				8x8モード
			*/
			bank = ppu.control.GetSpritePatternTableAddress()
		} else {
			/*
				8x16モード

				タイル選択は8x16モードの時のみ特殊 (8x8のときはタイルの番号)
				bit 76543210
						TTTTTTTP

				P: パターンテーブル選択。0:$0000, 1:$1000
				T: スプライト上半分のタイル ID を 2*T とし、下半分を 2*T+1 とする
			*/
			bank = (tileIndex & 0x01) * 0x1000
			tileIndex &= 0xFE

			// 下半分のとき
			if tileY >= uint16(TILE_SIZE) {
				tileIndex++
				tileY -= uint16(TILE_SIZE)
			}
		}

		// キャラクタROMからタイルデータを取得
		tileBasePointer := bank + tileIndex * uint16(TILE_SIZE*2)
		upper := ppu.Mapper.ReadCharacterROM(tileBasePointer + tileY)
		lower := ppu.Mapper.ReadCharacterROM(tileBasePointer + tileY + uint16(TILE_SIZE))

		// タイルデータを描画
		for x := range TILE_SIZE {
			var value uint8
			if flipH {
				// 水平反転の場合
				value = (lower & 1) << 1 | (upper & 1)
				upper >>= 1
				lower >>= 1
			} else {
				// 反転がない場合
				value = ((lower>>7) & 1) << 1 | ((upper >> 7) & 1)
				upper <<= 1
				lower <<= 1
			}

			// 透明ピクセルは描画しない
			if value == 0 { continue }

			actualX := uint(spriteX) + uint(x)

			// 画面外のピクセルは描画しない
			if actualX >= SCREEN_WIDTH { continue }

			if priority == 0 || canvas.isBgTransparentAt(ppu, actualX, uint(scanline)) {
				canvas.setPixelAt(actualX, uint(scanline), PALETTE[palette[value]])
			}
		}
	}
}

// MARK: 指定したスキャンラインに重なるスプライトを探索
func FindScanlineSprite(ppu *PPU, spriteHeight uint8, scanline uint16) (uint,  *[SPRITE_MAX][OAM_SPRITE_SIZE]uint8) {
	var sprites [SPRITE_MAX][OAM_SPRITE_SIZE]uint8 // 1スキャンラインに配置するスプライト (8個まで)

	var spriteCount uint = 0
	for i := range len(ppu.oam) / 4 {
		index := uint(i*4)
		/*
			struct Sprite{
					U8 y;
					U8 tile;
					U8 attr;
					U8 x;
			};
		*/
		spriteY := uint16(ppu.oam[index]) // OAM各スプライトの0バイト目がY座標

		// スプライトが現在のスキャンラインに収まっているかをチェックする
		if scanline >= spriteY && scanline < spriteY + uint16(spriteHeight) {
			if spriteCount < SPRITE_MAX {
				sprites[spriteCount][OAM_SPRITE_Y] = ppu.oam[index+OAM_SPRITE_Y] // Y座標
				sprites[spriteCount][OAM_SPRITE_TILE] = ppu.oam[index+OAM_SPRITE_TILE] // タイル選択
				sprites[spriteCount][OAM_SPRITE_ATTR] = ppu.oam[index+OAM_SPRITE_ATTR] // 属性
				sprites[spriteCount][OAM_SPRITE_X] = ppu.oam[index+OAM_SPRITE_X] // X座標
				spriteCount++
			} else {
				// 最大表示数を超えたらフラグを立てて抜ける
				ppu.status.SetSpriteOverflow(true)
				break
			}
		}
	}
	return spriteCount, &sprites
}

// MARK: 指定したスキャンラインを描画
func RenderScanline(ppu *PPU, canvas *Canvas, scanline uint16) {
	RenderScanlineBackground(ppu, canvas, scanline)
	RenderScanlineSprite(ppu, canvas, scanline)
}

// MARK: 指定した座標のネームテーブルを取得
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
