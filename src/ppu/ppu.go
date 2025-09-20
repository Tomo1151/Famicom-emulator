package ppu

import (
	"Famicom-emulator/cartridge/mappers"
	"fmt"
)

// MARK: 定数定義
const (
	VRAM_SIZE          uint16 = 2 * 1024 // 2kB
	PALETTE_TABLE_SIZE  uint8 = 32
	OAM_DATA_SIZE      uint16 = 64 * 4
)

const (
	SCANLINE_START      = 0
	SCANLINE_POSTRENDER = 240
	SCANLINE_VBLANK     = 241
	SCANLINE_PRERENDER  = 261
	SCANLINE_END        = 341

	OAM_SPRITE_SIZE uint = 4
	OAM_SPRITE_X    uint = 3
	OAM_SPRITE_Y    uint = 0
	OAM_SPRITE_TILE uint = 1
	OAM_SPRITE_ATTR uint = 2
	SPRITE_MAX      uint = 8

	TILE_SIZE     uint = 8
)

// MARK: PPUの定義
type PPU struct {
	Mapper mappers.Mapper
	PaletteTable [PALETTE_TABLE_SIZE+1]uint8
	vram [VRAM_SIZE]uint8
	oam [OAM_DATA_SIZE]uint8

	// IOレジスタ
	control   ControlRegister // $2000
	mask   MaskRegister    // $2001
	status StatusRegister  // $2002
	// scroll ScrollRegister  // $2005
	// address   AddrRegister    // $2006

	// 内部レジスタ
	v uint16 // current VRAM address
	t uint16 // VRAM temporary register
	x uint16 // x scroll
	w bool // write latch

	scanline uint16 // 現在描画中のスキャンライン
	cycles uint // PPUサイクル
	internalDataBuffer uint8
	oamAddress uint8 // OAM書き込みのポインタ

	NMI *uint8

	lineBuffer [SCREEN_WIDTH]Pixel // 次のスキャンラインのバッファ
}

// MARK: PPUの初期化メソッド
func (p *PPU) Init(mapper mappers.Mapper){
	p.Mapper = mapper
	for addr := range p.vram { p.vram[addr] = 0x00 }
	for addr := range p.oam { p.oam[addr] = 0x00 }
	for addr := range p.PaletteTable { p.PaletteTable[addr] = 0x00 }

	// IOレジスタの初期化
	p.control.Init()
	p.mask.Init()
	p.status.Init()
	p.scroll.Init()
	p.address.Init()

	// 内部レジスタの初期化
	p.v = 0x00
	p.t = 0x00
	p.x = 0x00
	p.w = false

	p.oamAddress = 0
	p.scanline = 0
	p.cycles = 0
	p.internalDataBuffer = 0x00

	p.NMI = nil

	// ラインバッファの初期化
	for i := range p.lineBuffer {
		p.lineBuffer[i] = Pixel{
			PIXEL_TYPE_BACKGROUND, // type
			0x00, // priority
			PALETTE[p.PaletteTable[0]], // value (rgb palette)
		}
	}
}

// MARK: PPUコントロールレジスタ($2000)への書き込み
func (p *PPU) WriteToPPUControlRegister(value uint8) {
	prev := p.control.GenerateVBlankNMI()
	p.control.update(value)

	// VBlank中にGenerateNMIが立つタイミングでNMIを発生させる
	if !prev && p.control.GenerateVBlankNMI() && p.status.IsInVBlank() {
		*p.NMI = 0x01
	}
}

// MARK: PPUマスクレジスタ($2001)への書き込み
func (p *PPU) WriteToPPUMaskRegister(value uint8) {
	p.mask.update(value)
}

// MARK: OAM ADDR($2003) への書き込み
func (p *PPU) WriteToOAMAddressRegister(addr uint8) {
	p.oamAddress = addr
}

// // MARK: PPUスクロールレジスタ($2005)への書き込み
// func (p *PPU) WriteToPPUScrollRegister(data uint8) {
// 	p.scroll.Write(data)
// }

// // MARK: PPUアドレスレジスタ($2006)への書き込み
// func (p *PPU) WriteToPPUAddrRegister(value uint8) {
// 	p.address.update(value)
// }

// MARK: OAM DATA($4014) への書き込み
func (p *PPU) WriteToOAMDataRegister(data uint8) {
	p.oam[p.oamAddress] = data
	p.oamAddress++
}

// MARK: DMA転送を行う ([256]u8 の配列のアドレスを受け取る)
func (p *PPU) DMATransfer(bytes *[256]uint8) {
	for _, byte := range *bytes {
		p.oam[p.oamAddress] = byte
		p.oamAddress++
	}
}


// MARK: VRAMアドレスをインクリメント
func (p *PPU) incrementVRAMAddress() {
	p.address.increment(p.control.GetVRAMAddrIncrement())
}

// MARK: VRAMへの書き込み
func (p *PPU) WriteVRAM(value uint8) {
	/*
		PPUメモリマップ

		$0000-$1FFF $2000 パレットテーブル (CHR ROM)
		$2000-$3EFF $1F00 ネームテーブル (VRAM)
		$3F00-$3FFF $0100 パレット
		$4000-$FFFF $4000 $0000-$3FFF のミラーリング
	*/

	address := p.address.get()
	p.incrementVRAMAddress()

	switch {
	case address <= 0x1FFF:
		if p.Mapper.GetIsCharacterRAM() {
			p.Mapper.WriteToCharacterROM(address, value)
			// p.CHR_ROM[address] = value
		}
	case 0x2000 <= address && address <= 0x2FFF:
		p.vram[p.mirrorVRAMAddress(address)] = value
	case 0x3000 <= address && address <= 0x3EFF:
		// fmt.Printf("Error: unexpected vram write to $%04X\n", address)
		return
	case 0x3F00 <= address && address <= 0x3F1F:
		// アドレスのミラーリング
		if address == 0x3F10 ||
			 address == 0x3F14 ||
			 address == 0x3F18 ||
			 address == 0x3FC {
			address -= 0x10
		}
		p.PaletteTable[address - 0x3F00] = value
	case 0x3F20 <= address && address <= 0x3FFF:
		p.PaletteTable[(address - 0x3F00)%32] = value
	default:
		panic(fmt.Sprintf("Unexpected write to mirrored space: %04X", address))
	}
}

// MARK: PPUコントロールレジスタの読み取り
func (p *PPU) ReadPPUControl() uint8 {
	return p.control.ToByte()
}

// MARK: PPUマスクレジスタの読み取り
func (p *PPU) ReadPPUMask() uint8 {
	return p.mask.ToByte()
}

// MARK: PPUステータスレジスタの読み取り
func (p *PPU) ReadPPUStatus() uint8 {
	status := p.status.ToByte()
	p.status.ClearVBlankStatus()
	p.scroll.ResetLatch()
	p.address.ResetLatch()
	// @FIXME READ PPU STATUSした次のフレームはNMIを発生させない
	return status
}

// MARK: OAM DATAの読み取り
func (p *PPU) ReadOAMData() uint8 {
	return p.oam[p.oamAddress]
}

// MARK: VRAMの読み取り
func (p *PPU) ReadVRAM() uint8 {
	/*
		PPUメモリマップ

		$0000-$1FFF $2000 パレットテーブル (CHR ROM)
		$2000-$3EFF $1F00 ネームテーブル (VRAM)
		$3F00-$3FFF $0100 パレット
		$4000-$FFFF $4000 $0000-$3FFF のミラーリング
	*/

	address := p.address.get()
	p.incrementVRAMAddress()

	switch {
	case address <= 0x1FFF:
		value := p.internalDataBuffer
		p.internalDataBuffer = p.Mapper.ReadCharacterROM(address)
		return value
	case 0x2000 <= address && address <= 0x2FFF:
		// 一回遅れで値は反映されるため，内部バッファを更新し，元のバッファ値を返す
		value := p.internalDataBuffer
		p.internalDataBuffer = p.vram[p.mirrorVRAMAddress(address)]
		return value
	case 0x3000 <= address && address <= 0x3EFF:
		panic(fmt.Sprintf("Error: address space 0x3000..0x3eff is not expected to read, requested: %04X", address))
	case 0x3F00 <= address && address <= 0x3F1F:
		// アドレスのミラーリング
		if address == 0x3F10 ||
			 address == 0x3F14 ||
			 address == 0x3F18 ||
			 address == 0x3FC {
			address -= 0x10
		}
		return p.PaletteTable[address - 0x3F00]
	case 0x3F20 <= address && address <= 0x3FFF:
		return p.PaletteTable[(address - 0x3F00)%32]
	default:
		panic(fmt.Sprintf("Error: unexpected read to mirrored space: %04X", address))
	}
}

// MARK: VRAMアドレスのミラーリング
func (p *PPU) mirrorVRAMAddress(addr uint16) uint16 {
	// 0x3000-0x3eff から 0x2000 - 0x2eff へミラーリング
	mirroredVRAMAddr := addr & PPU_VRAM_MIRROR_MASK

	// メモリアドレスをVRAMの配列用に補正 (VRAMの先頭アドレスを引く)
	vramIndex := mirroredVRAMAddr - 0x2000

	// ネームテーブルのインデックスを求める
	nameTable := vramIndex / 0x400

	mirroring := p.Mapper.GetMirroring()

	// ネームテーブルのミラーリングがVerticalの場合
	// [ A ] [ B ] (一つのテーブルが 0x400 × 0x400，そのテーブルが 2 × 2)
	// [ a ] [ b ]
	if mirroring == mappers.MIRRORING_VERTICAL {
		if nameTable == 2 || nameTable == 3 {
			return vramIndex - 0x800
		}
	}

	// ネームテーブルのミラーリングがHorizontalの場合
	// [ A ] [ a ]
	// [ B ] [ b ]
	if mirroring == mappers.MIRRORING_HORIZONTAL {
		switch nameTable {
		case 2, 1:
			return vramIndex - 0x400
		case 3:
			return vramIndex - 0x800
		}
	}

	return vramIndex
}

// MARK: 待機しているNMIを取得
func (p *PPU) GetNMI() *uint8 {
	if p.NMI != nil {
		value := *p.NMI
		p.NMI = nil
		return &value
	} else {
		return nil
	}
}

// MARK: スプライト0ヒットの判定
func (p *PPU) isSpriteZeroHit(cycles uint) bool {
	x := uint(p.oam[3])
	y := uint(p.oam[0]) + 6 // スプライト0ヒットが反映されるまでのラグ

	/*
		@NOTE
		参考：https://www.nesdev.org/wiki/PPU_rendering
		> Sprite 0 hit acts as if the image starts at cycle 2 (which is the same cycle that the shifters shift for the first time), so the sprite 0 flag will be raised at this point at the earliest. Actual pixel output is delayed further due to internal render pipelining, and the first pixel is output during cycle 4.

		@FIXME
		スプライトの可視ピクセルを判定に加える，現在はそれをしておらず，SMBのコイン下半分のスプライトに合わせているため +6 になっているが，これが +4 になるはず
	*/
	return p.mask.SpriteEnable && y == uint(p.scanline) && x <= cycles
}

// MARK: 指定したピクセルで使用するネームテーブルを取得
func (p *PPU) getNameTableForPixel(x uint, y uint) []uint8 {
	mirroring := p.Mapper.GetMirroring()
	baseNameTableAddress := p.control.GetBaseNameTableAddress()
	primaryNameTable := p.vram[0x000:0x400]
	secondaryNameTable := p.vram[0x400:0x800]

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

// MARK: BG面のカラーパレットを取得
func (p *PPU) getBGPalette(attrributeTable *[]uint8, tileColumn uint, tileRow uint) [4]uint8 {
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
		p.PaletteTable[0],
		p.PaletteTable[paletteStart+0],
		p.PaletteTable[paletteStart+1],
		p.PaletteTable[paletteStart+2],
	}

	return color
}

// MARK: スプライトのカラーパレットを取得
func (p *PPU) getSpritePalette(paletteIndex uint8) [4]uint8 {
	var start uint = 0x11 + uint(paletteIndex * 4)
	return [4]uint8{
		0,
		p.PaletteTable[start + 0],
		p.PaletteTable[start + 1],
		p.PaletteTable[start + 2],
	}
}

// MARK: ラインバッファをクリア
func (p *PPU) ClearLineBuffer() {
	for x := range p.lineBuffer {
		p.lineBuffer[x].Type = PIXEL_TYPE_BACKGROUND
		p.lineBuffer[x].value = PALETTE[p.PaletteTable[0]]
		p.lineBuffer[x].priority = 0x00
	}
}


// MARK: 指定したスキャンラインに重なるスプライトを探索
func (p *PPU) FindScanlineSprite(spriteHeight uint8, scanline uint16) (uint,  *[SPRITE_MAX][OAM_SPRITE_SIZE]uint8) {
	var sprites [SPRITE_MAX][OAM_SPRITE_SIZE]uint8 // 1スキャンラインに配置するスプライト (8個まで)

	var spriteCount uint = 0
	for i := range len(p.oam) / 4 {
		index := uint(i*4)
		/*
			struct Sprite{
					U8 y;
					U8 tile;
					U8 attr;
					U8 x;
			};
		*/
		spriteY := uint16(p.oam[index]) // OAM各スプライトの0バイト目がY座標

		// スプライトが現在のスキャンラインに収まっているかをチェックする
		if scanline >= spriteY && scanline < spriteY + uint16(spriteHeight) {
			if spriteCount < SPRITE_MAX {
				sprites[spriteCount][OAM_SPRITE_Y] = p.oam[index+OAM_SPRITE_Y] // Y座標
				sprites[spriteCount][OAM_SPRITE_TILE] = p.oam[index+OAM_SPRITE_TILE] // タイル選択
				sprites[spriteCount][OAM_SPRITE_ATTR] = p.oam[index+OAM_SPRITE_ATTR] // 属性
				sprites[spriteCount][OAM_SPRITE_X] = p.oam[index+OAM_SPRITE_X] // X座標
				spriteCount++
			} else {
				// 最大表示数を超えたらフラグを立てて抜ける
				p.status.SetSpriteOverflow(true)
				break
			}
		}
	}
	return spriteCount, &sprites
}

// MARK: 指定したスキャンラインのBG面を計算
func (p *PPU) CalculateScanlineBackground(canvas *Canvas, scanline uint16) {
	// BGが無効であれば描画をしない
	if !p.mask.BackgroundEnable { return }

	// スクロール値を取得
	scrollX := uint(p.scroll.ScrollX)
	scrollY := uint(p.scroll.ScrollY)

	// 描画するY座標を計算
	globalY := scrollY + uint(scanline)
	actualY := globalY % TILE_SIZE // タイル内の何行目か

	// 画面の左端から右端まで
	for x := range SCREEN_WIDTH {
		// 左端のBG面描画フラグが無効であれば描画しない
		if !p.mask.LeftmostBackgroundEnable && x < TILE_SIZE { continue }

		// 描画するX座標を計算
		globalX := scrollX + x

		// 描画対象のネームテーブルを決定
		nameTable := p.getNameTableForPixel(globalX, globalY)

		// ネームテーブル内のタイル座標を計算
		tileX := (globalX / TILE_SIZE) % 32
		tileY := (globalY / TILE_SIZE) % 30

		// タイルのインデックスを取得
		tileIndex := uint16((nameTable)[tileY*32+tileX])

		// 属性テーブルからパレット情報を取得
		attributeTable := (nameTable)[0x3C0:0x400]
		palette := p.getBGPalette(&attributeTable, tileX, tileY)

		// パターンテーブルからタイルのピクセルデータを取得
		bank := p.control.GetBackgroundPatternTableAddress()
		tileBasePointer := bank + tileIndex * uint16(TILE_SIZE * 2)

		// 実際のY座標に対応するタイルデータを2バイト取得
		upper := p.Mapper.ReadCharacterROM(tileBasePointer + uint16(actualY))
		lower := p.Mapper.ReadCharacterROM(tileBasePointer + uint16(actualY) + uint16(TILE_SIZE))

		// 実際のY座標に対応するピクセルを計算
		actualX := (TILE_SIZE-1) - (globalX % TILE_SIZE)

		// そのピクセルの色を確定
		value := (lower >> uint8(actualX) & 1) << 1 | (upper >> uint8(actualX) & 1)

		// ラインバッファに登録
		p.lineBuffer[x].Type  = PIXEL_TYPE_BACKGROUND
		p.lineBuffer[x].value = PALETTE[palette[value]]
		p.lineBuffer[x].priority = 0x00
	}
}

// MARK: 指定したスキャンラインのスプライトを計算
func (p *PPU) CalculateScanlineSprite(canvas *Canvas, scanline uint16) {
	// スプライトが無効であれば描画しない
	if !p.mask.SpriteEnable { return }

	// スプライトサイズの取得 (8 / 16)
	spriteHeight := p.control.GetSpriteSize()
	spriteCount, sprites := p.FindScanlineSprite(spriteHeight, scanline)

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
		palette := p.getSpritePalette(paletteIndex)

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
			bank = p.control.GetSpritePatternTableAddress()
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
		upper := p.Mapper.ReadCharacterROM(tileBasePointer + tileY)
		lower := p.Mapper.ReadCharacterROM(tileBasePointer + tileY + uint16(TILE_SIZE))

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

			// 左端のスプライト描画フラグが無効であれば描画しない
			if !p.mask.LeftmostSpriteEnable && actualX < TILE_SIZE { continue }

			// 描画ピクセルの背景が透明かどうか
			isBgTransparent := p.lineBuffer[actualX].value == PALETTE[p.PaletteTable[0]]

			// スプライトの優先度が0または背景が透明であれば描画
			// @FIXME BG面より優先されないスプライトに，OAM上の順番が後ろが重なった時にBG面が最優先になるようにする (SMB3のパックンフラワー等)
			if priority == 0 || isBgTransparent {
				p.lineBuffer[actualX].Type = PIXEL_TYPE_SPRITE
				p.lineBuffer[actualX].value = PALETTE[palette[value]]
				p.lineBuffer[actualX].priority = priority
			}
		}
	}
}

// MARK: PPUを動かす
func (p *PPU) Tick(canvas *Canvas, cycles uint) bool {
	// サイクルを進める
	p.cycles += cycles

	if p.cycles >= SCANLINE_END {
		if p.isSpriteZeroHit(p.cycles) {
			p.status.SetSpriteZeroHit(true)
		}

		// サイクル数を0に戻す
		p.cycles -= SCANLINE_END

		// マッパーによるIRQの判定
		p.Mapper.GenerateScanlineIRQ(p.scanline, p.mask.BackgroundEnable || p.mask.SpriteEnable)

		// 可視領域のスキャンラインを描画
		if SCANLINE_START <= p.scanline && p.scanline < SCANLINE_POSTRENDER {
			RenderScanlineToCanvas(p, canvas, p.scanline)
		}

		// スキャンラインを進める
		p.scanline++

		// VBlankに突入
		if p.scanline == SCANLINE_VBLANK {
			p.status.SetVBlankStatus(true)
			p.status.SetSpriteZeroHit(false)
			if p.control.GenerateVBlankNMI() {
				// NMIを設定
				nmiValue := uint8(1)
				p.NMI = &nmiValue
			}
		}

		// プリレンダーラインに到達した時
		if p.scanline >= SCANLINE_PRERENDER+1 {
			p.scanline = 0
			p.NMI = nil
			p.status.SetSpriteZeroHit(false)
			p.status.ClearVBlankStatus()
			return true
		}
	}
	return false
}