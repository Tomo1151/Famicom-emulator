package ppu

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"

	"Famicom-emulator/cartridge/mappers"
	"Famicom-emulator/config"
)

// MARK: 定数定義
const (
	PPU_VRAM_SIZE          = 2 * 1024             // 1024kB
	PPU_PALETTE_TABLE_SIZE = 32                   // 32色 (背景/スプライト各16色ずつ)
	PPU_PRIMARY_OAM_SIZE   = 64 * OAM_SPRITE_SIZE // 64スプライト
	PPU_SECONDARY_OAM_SIZE = 8 * OAM_SPRITE_SIZE  // 8スプライト

	PPU_ADDRESS_START      = 0x0000
	PPU_ADDRESS_END        = 0xFFFF
	PPU_VRAM_ADDRESS_SPACE = 0x4000

	PPU_MEMORY_ADDRESS_MASK    = 0x3FFF // 14ビット
	PPU_NAMETABLE_ADDRESS_MASK = 0x2FFF // ミラーリング直前のアドレス

	TILE_SIZE             = 8 // 1タイルのサイズ (幅，高さ)
	MAX_SPRITE_COUNT      = 8 // スプライトの最大同時表示数
	OPEN_BUS_DECAY_CYCLES = 3_000_000
)

const (
	SCANLINE_START      = 0
	SCANLINE_POSTRENDER = 240
	SCANLINE_VBLANK     = 241
	SCANLINE_PRERENDER  = 261
	SCANLINE_END        = 340
)

const (
	OAM_SPRITE_Y_POS uint = iota
	OAM_SPRITE_TILE_POS
	OAM_SPRITE_ATTR_POS
	OAM_SPRITE_X_POS

	OAM_SPRITE_SIZE = 4 // 1スプライトあたりのバイト数
)

const (
	COLOR_EMPHASIZE_FACTOR = 0.75
	SPRITE_ZERO_NOT_FOUND  = 0xFF
)

// MARK: PPUの定義
type PPU struct {
	vram         [PPU_VRAM_SIZE]uint8          // Video RAM
	oam          [PPU_PRIMARY_OAM_SIZE]uint8   // Object Attribute Memory
	secondaryOAM [PPU_SECONDARY_OAM_SIZE]uint8 // Secondary OAM
	paletteTable [PPU_PALETTE_TABLE_SIZE]uint8 // Palette RAM

	// IOレジスタ
	control ControlRegister // $2000
	mask    MaskRegister    // $2001
	status  StatusRegister  // $2002

	// 内部レジスタ
	t AddressRegiseter
	v AddressRegiseter
	x XRegister
	w WRegister

	// 描画用シフトレジスタ
	backgroundLatch BackgroundLatch
	backgroundShift BackgroundShiftRegister
	spriteLatch     SpriteLatch
	spriteShifts    [8]SpriteShiftRegister

	mapper mappers.Mapper // カートリッジ (CHR ROM) への参照
	canvas *Canvas        // 描画キャンバスの参照

	dot               uint
	scanline          uint
	oamAddress        uint8
	dataBuffer        uint8
	spriteCount       uint
	spriteZeroIndex   uint
	nmi               bool
	nmiCounter        uint
	nmiSuppress       bool
	isOddFrame        bool
	openBus           uint8
	openBusDecayTimer int

	config config.Config
	frame  uint64
}

// MARK: PPUのコンストラクタ
func (p *PPU) Init(mapper mappers.Mapper, canvas *Canvas, config config.Config) {
	p.control = NewControlRegister()
	p.mask = NewMaskRegister()
	p.status = NewStatusRegister()
	p.t = NewAddressRegister()
	p.v = NewAddressRegister()
	p.x = NewXRegister()
	p.w = NewWRegister()
	p.mapper = mapper
	p.config = config
	p.canvas = canvas
	p.dot = 0
	p.scanline = 0
	p.oamAddress = 0x00
	p.dataBuffer = 0x00
	p.spriteCount = 0
	p.spriteZeroIndex = SPRITE_ZERO_NOT_FOUND
	p.nmi = false
	p.nmiCounter = 0
	p.nmiSuppress = false
	p.isOddFrame = false
	p.openBus = 0x00
	p.frame = 0
}

// MARK: PPUクロックの更新
func (p *PPU) Tick(cycles uint) {
	for range cycles {
		// 描画設定を取得
		isRenderingEnabled := p.mask.backgroundEnable || p.mask.spriteEnable

		// スキャンライン位置によって処理
		switch {
		case SCANLINE_START <= p.scanline && p.scanline < SCANLINE_POSTRENDER:
			p.tickVisibleScanline(isRenderingEnabled)
		case p.scanline == SCANLINE_VBLANK:
			p.tickVBlankScanline()
		case p.scanline == SCANLINE_PRERENDER:
			p.tickPreRenderScanline(isRenderingEnabled)
		}

		// 奇数フレームで背景描画が有効の場合は1サイクルスキップされる
		if p.scanline == SCANLINE_PRERENDER && p.dot == 339 &&
			p.isOddFrame && p.mask.backgroundEnable {
			p.dot++
		}

		// サイクル, スキャンラインを進める
		p.incrementCycles()
	}
}

// MARK: PPUコントロールレジスタの読み取り (CPU: $2000)
func (p *PPU) ReadPPUControl() uint8 {
	return p.control.ToByte()
}

// MARK: PPUマスクレジスタの読み取り (CPU: $2001)
func (p *PPU) ReadPPUMask() uint8 {
	return p.mask.ToByte()
}

// MARK: PPUステータスレジスタの読み取り (CPU: $2002)
func (p *PPU) ReadPPUStatus() uint8 {
	status := p.status.ToByte()
	if p.scanline == SCANLINE_VBLANK && p.dot == 0 {
		p.nmiSuppress = true
	}
	p.status.SetVBlank(false) // 読み取りでVBlankフラグとラッチがクリアされる
	p.w.reset()

	value := status | (p.openBus)&0x1F
	p.openBus = value
	return value
}

// MARK: OAMデータの読み取り (CPU: $2004)
func (p *PPU) ReadOAMData() uint8 {
	value := p.oam[p.oamAddress]

	// 属性バイトの bit 2~4 は未実装のため 0 として読みだされる
	if p.oamAddress%4 == 2 {
		value &= 0xE3
	}

	// オープンバスに影響を与える
	p.refreshOpenBus(value)

	return value
}

// MARK: PPUデータの読み取り (CPU: $2007)
func (p *PPU) ReadPPUData() uint8 {
	address := p.v.ToWord() & PPU_MEMORY_ADDRESS_MASK // $4000-$FFFF のミラーリング
	p.incrementVRAMAddress()

	// CPUからの読み取りは内部バッファにより一回分遅延する
	value := p.dataBuffer
	p.dataBuffer = p.ReadPPUMemory(address)

	// パレットテーブルのみ遅延無しで読み取り
	if 0x3F00 <= address && address <= 0x3FFF {
		value = p.ReadPPUMemory(address)
		p.dataBuffer = p.ReadPPUMemory(address & PPU_NAMETABLE_ADDRESS_MASK)
	}

	// オープンバスに影響を与える
	p.refreshOpenBus(value)

	return value
}

// MARK: PPUメモリマップの読み取り
func (p *PPU) ReadPPUMemory(address uint16) uint8 {
	/*
		PPU メモリマップ
		(範囲 / サイズ / 対象)

		$0000-$1FFF 0x2000 パターンテーブル
		$2000-$2FFF 0x1000 ネームテーブル
		$3000-$3EFF 0x0F00 ネームテーブルのミラーリング
		$3F00-$3F1F 0x0020 パレットテーブル
		$3F20-$3FFF 0x00E0 パレットテーブルのミラーリング
		$4000-$FFFF 0x4000 $0000-$3FFF のミラーリング
	*/

	switch {
	case PPU_ADDRESS_START <= address && address <= 0x1FFF: // パターンテーブル (CHR ROM)
		return p.mapper.ReadCharacterRom(address)
	case 0x2000 <= address && address <= 0x3EFF: // ネームテーブル (VRAM)
		vramAddress := p.mirrorVRAMAddress(address & PPU_NAMETABLE_ADDRESS_MASK)
		return p.vram[vramAddress]
	case 0x3F00 <= address && address <= 0x3FFF: // パレットテーブル (Palette RAM)
		paletteTableIndex := (address - 0x3F00) % PPU_PALETTE_TABLE_SIZE
		if paletteTableIndex >= 0x10 && paletteTableIndex%4 == 0 {
			paletteTableIndex -= 0x10 // $3F10, $3F14, $3F18, $3F1C は $3F00 番台にミラーされる
		}
		return p.paletteTable[paletteTableIndex]
	default:
		return 0x00
	}
}

// MARK: PPUコントロールレジスタの書き込み (CPU: $2000)
func (p *PPU) WritePPUControl(value uint8) {
	prev := p.control.GenerateNMI()

	p.control.SetFromByte(value)
	p.t.updateNameTable(value)

	// VBlank中にGenerateNMIがセットされたタイミングでNMIが発生
	if !prev && p.control.GenerateNMI() && p.status.VBlank() {
		p.nmi = true
		p.nmiCounter = 1
	} else if prev && !p.control.GenerateNMI() && p.nmiCounter > 0 {
		p.nmi = false
		p.nmiCounter = 0
	}

	// オープンバスに影響を与える
	p.refreshOpenBus(value)
}

// MARK: PPUマスクレジスタの書き込み (CPU: $2001)
func (p *PPU) WritePPUMask(value uint8) {
	p.mask.SetFromByte(value)

	// オープンバスに影響を与える
	p.refreshOpenBus(value)
}

// MARK: PPUステータスレジスタの書き込み (CPU: $2002)
func (p *PPU) WritePPUStatus(value uint8) {
	// オープンバスに影響を与える
	p.refreshOpenBus(value)
}

// MARK: OAMアドレスの書き込み (CPU: $2003)
func (p *PPU) WriteOAMAddress(value uint8) {
	p.oamAddress = value

	// オープンバスに影響を与える
	p.refreshOpenBus(value)
}

// MARK: OAMデータの書き込み (CPU: $2004)
func (p *PPU) WriteOAMData(value uint8) {
	p.oam[p.oamAddress] = value
	p.oamAddress++ // OAMアドレスは自動でインクリメントされる

	// オープンバスに影響を与える
	p.refreshOpenBus(value)
}

// MARK: PPUスクロールの書き込み (CPU: $2005)
func (p *PPU) WritePPUScroll(value uint8) {
	if !p.w.latch {
		p.x.update(value) // 1回目はXレジスタも更新 (fineX)
	}

	p.t.updateScroll(value, p.w.latch) // Tレジスタは毎回更新 (fineY / coarseX / coarseY)
	p.w.toggle()

	// オープンバスに影響を与える
	p.refreshOpenBus(value)
}

// MARK: PPUアドレスの書き込み (CPU: $2006)
func (p *PPU) WritePPUAddress(value uint8) {
	p.t.updateAddress(value, p.w.latch)

	if p.w.latch {
		// 2回目の書き込み時はTレジスタをVレジスタにコピー
		p.t.copyAllBitsTo(&p.v)
	}

	p.w.toggle()

	// オープンバスに影響を与える
	p.refreshOpenBus(value)
}

// MARK: PPUデータの書き込み (CPU: $2007)
func (p *PPU) WritePPUData(value uint8) {
	address := p.v.ToWord() & PPU_MEMORY_ADDRESS_MASK // $4000-$FFFF のミラーリング
	p.incrementVRAMAddress()
	p.WritePPUMemory(address, value)

	// オープンバスに影響を与える
	p.refreshOpenBus(value)
}

// MARK: PPUメモリマップへの書き込み
func (p *PPU) WritePPUMemory(address uint16, value uint8) {
	/*
		PPU メモリマップ
		(範囲 / サイズ / 対象)

		$0000-$1FFF 0x2000 パターンテーブル
		$2000-$2FFF 0x1000 ネームテーブル
		$3000-$3EFF 0x0F00 ネームテーブルのミラーリング
		$3F00-$3F1F 0x0020 パレットテーブル
		$3F20-$3FFF 0x00E0 パレットテーブルのミラーリング
		$4000-$FFFF 0x4000 $0000-$3FFF のミラーリング
	*/

	switch {
	case PPU_ADDRESS_START <= address && address <= 0x1FFF: // パターンテーブル (CHR RAM)
		if p.mapper.IsCharacterRam() {
			p.mapper.WriteToCharacterRom(address, value)
		}
	case 0x2000 <= address && address <= 0x3EFF: // ネームテーブル (VRAM)
		vramAddress := p.mirrorVRAMAddress(address & PPU_NAMETABLE_ADDRESS_MASK)
		p.vram[vramAddress] = value
	case 0x3F00 <= address && address <= 0x3FFF: // パレットテーブル (Palette RAM)
		paletteTableIndex := (address - 0x3F00) % PPU_PALETTE_TABLE_SIZE
		if paletteTableIndex >= 0x10 && paletteTableIndex%4 == 0 {
			paletteTableIndex -= 0x10 // $3F10, $3F14, $3F18, $3F1C は $3F00 番台にミラーされる
		}
		p.paletteTable[paletteTableIndex] = value
	default:
	}
}

// MARK: DMA転送の実行 (CPU: $4014)
func (p *PPU) DMATransfer(bytes *[256]uint8) {
	for _, value := range *bytes {
		p.oam[p.oamAddress] = value
		p.oamAddress++
	}
}

// MARK: 可視スキャンラインの処理
func (p *PPU) tickVisibleScanline(isRenderingEnabled bool) {
	if p.dot == 1 {
		// ラインの先頭でセカンダリOAMを初期化
		p.clearSecondaryOAM()
	}
	if p.dot == 256 {
		p.evaluateNextLineSprite()
	}

	// 描画
	if 1 <= p.dot && p.dot <= 256 {
		p.renderPixel()

		if isRenderingEnabled {
			p.shiftSpriteRegisters()
		}
	}

	if !isRenderingEnabled {
		return
	}

	// 背景フェッチ
	p.fetchBackgroundPipeline()

	// スクロール更新
	if p.dot == 256 {
		p.v.incrementVertical()
	}
	if p.dot == 257 {
		p.t.copyHorizontalBitsTo(&p.v)
		p.oamAddress = 0x00
	}

	// マッパー割り込みの生成
	if p.dot == 260 {
		p.mapper.GenerateScanlineIRQ(uint16(p.scanline), isRenderingEnabled)
	}

	// スプライトフェッチ
	p.fetchSpritePipeline()
}

// MARK: VBlankラインの処理
func (p *PPU) tickVBlankScanline() {
	if p.dot == 1 {
		if !p.nmiSuppress {
			// ラインの先頭でVBlankフラグを立てる
			p.status.SetVBlank(true)
			// fmt.Printf("VBlank %d,%d\n", p.scanline, p.dot)
			if p.control.GenerateNMI() {
				p.nmi = true
				p.nmiCounter = 3
			}
		}
		p.nmiSuppress = false
	}
}

// MARK: プリレンダーラインの処理
func (p *PPU) tickPreRenderScanline(isRenderingEnabled bool) {
	if p.dot == 1 {
		p.status.SetVBlank(false)
		p.status.SetSpriteZeroHit(false)
		p.status.SetSpriteOverflow(false)
		p.clearSecondaryOAM()
	}
	if p.dot == 256 {
		p.evaluateNextLineSprite()
	}

	if !isRenderingEnabled {
		return
	}

	p.fetchBackgroundPipeline()

	if p.dot == 256 {
		p.v.incrementVertical()
	}
	if p.dot == 257 {
		p.t.copyHorizontalBitsTo(&p.v)
		p.oamAddress = 0x00
	}
	if 280 <= p.dot && p.dot <= 304 {
		p.t.copyVerticalBitsTo(&p.v)
	}

	// スプライトフェッチ
	p.fetchSpritePipeline()
}

// MARK: 次のスキャンラインで描画するスプライトの評価
func (p *PPU) evaluateNextLineSprite() {
	p.spriteCount = 0
	spriteHeight := p.control.SpriteSize()

	// 次のスキャンラインを計算，プリレンダーライン以降は次フレームの0ライン目になる
	nextScanline := p.scanline + 1
	if nextScanline > SCANLINE_PRERENDER {
		nextScanline = 0
	}

	// プライマリOAMを順番に評価
	for i := range len(p.oam) / OAM_SPRITE_SIZE {
		primaryBase := uint(i) * OAM_SPRITE_SIZE
		spriteY := uint(p.oam[primaryBase+OAM_SPRITE_Y_POS]) + 1 // OAMのY座標は表示座標-1のため補正

		// 次のラインに重なっているか判定
		if nextScanline < spriteY || nextScanline >= spriteY+uint(spriteHeight) {
			continue
		}

		// 重なっているスプライトをプライマリOAMからセカンダリOAMへロード
		if p.spriteCount < MAX_SPRITE_COUNT {
			// プライマリOAMの先頭であればスプライト0として位置を記憶
			if i == 0 {
				p.spriteZeroIndex = p.spriteCount
			}

			// セカンダリOAMにプライマリOAMの中身をコピー
			secondaryBase := p.spriteCount * OAM_SPRITE_SIZE
			p.secondaryOAM[secondaryBase+OAM_SPRITE_Y_POS] = p.oam[primaryBase+OAM_SPRITE_Y_POS]
			p.secondaryOAM[secondaryBase+OAM_SPRITE_TILE_POS] = p.oam[primaryBase+OAM_SPRITE_TILE_POS]
			p.secondaryOAM[secondaryBase+OAM_SPRITE_ATTR_POS] = p.oam[primaryBase+OAM_SPRITE_ATTR_POS]
			p.secondaryOAM[secondaryBase+OAM_SPRITE_X_POS] = p.oam[primaryBase+OAM_SPRITE_X_POS]

			// ライン上のスプライト数をインクリメント
			p.spriteCount++
		} else {
			// 9個目が存在する場合スプライトオーバーフローフラグをセットし評価を止める
			p.status.SetSpriteOverflow(true)
			break
		}
	}
}

// MARK: 背景フェッチパイプライン
func (p *PPU) fetchBackgroundPipeline() {
	if (1 <= p.dot && p.dot <= 256) || (321 <= p.dot && p.dot <= 336) {
		// 背景パターンのフェッチ
		p.backgroundShift.shift()
		p.fetchBackground()
	}
}

// MARK: スプライトフェッチパイプライン
func (p *PPU) fetchSpritePipeline() {
	if 257 <= p.dot && p.dot <= 320 {
		// 次のライン用にタイルをフェッチ
		spriteIndex := (p.dot - 257) / TILE_SIZE
		p.fetchSprite(spriteIndex)
	}
}

// MARK: 背景ネームテーブル/属性テーブルのフェッチ
func (p *PPU) fetchBackground() {
	/*
		背景フェッチタイミング (8ドット毎)

		| 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | ...
		-----------------------------------------
		| - |  NT   |  AT   | BG lo | BG hi | ...

		1 ~ 7ドット目のフェッチはラッチに保存され，8ドット目でラッチからシフタにまとめて反映される
	*/

	switch p.dot % TILE_SIZE {
	case 1: // ネームテーブルのフェッチ
		nameTableAddress := p.getNameTableAddress()
		tile := p.ReadPPUMemory(nameTableAddress)
		p.backgroundLatch.nameTable = tile
	case 3: // 属性テーブルのフェッチ
		attributeTableAddress := p.getAttributeTableAddress()
		attribute := p.ReadPPUMemory(attributeTableAddress)
		// Vレジスタの位置に応じて該当する2ビットを抽出する
		shift := ((p.v.coarseY>>1)&1)<<2 | ((p.v.coarseX>>1)&1)<<1
		p.backgroundLatch.attribute = (attribute >> shift) & 0x03
	case 5: // パターンテーブル(下位)のフェッチ
		patternTableAddress := p.getBackgroundPatternAddress(false)
		pattern := p.ReadPPUMemory(patternTableAddress)
		p.backgroundLatch.patternLower = pattern
	case 7: // パターンテーブル(上位)のフェッチ
		patternTableAddress := p.getBackgroundPatternAddress(true)
		pattern := p.ReadPPUMemory(patternTableAddress)
		p.backgroundLatch.patternUpper = pattern
	case 0: // ラッチからシフトレジスタへロード
		p.backgroundShift.load(&p.backgroundLatch)
		p.v.incrementHorizontal()
	}
}

// MARK: スプライトネームテーブルのフェッチ
func (p *PPU) fetchSprite(index uint) {
	/*
		スプライトフェッチタイミング (8ドット毎)

		| 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
		-----------------------------------
		| - |   -   |   -   | Splow | SpHi  |
	*/

	// 現在のラインで使用されないシフトレジスタをリセット
	if p.spriteCount <= index {
		p.spriteShifts[index].reset()
		return
	}

	switch p.dot % TILE_SIZE {
	case 5: // パターンテーブル(下位)のフェッチ
		address := p.getSpritePatternAddress(index, false)
		p.spriteLatch.patternLower = p.ReadPPUMemory(address)
	case 7: // パターンテーブル(上位)のフェッチ
		address := p.getSpritePatternAddress(index, true)
		p.spriteLatch.patternUpper = p.ReadPPUMemory(address)
	case 0: // ラッチ・セカンダリOAMからシフトレジスタへロード
		basePtr := index * OAM_SPRITE_SIZE

		spriteX := p.secondaryOAM[basePtr+OAM_SPRITE_X_POS]
		attributes := p.secondaryOAM[basePtr+OAM_SPRITE_ATTR_POS]
		flipH := (attributes>>6)&1 == 1

		p.spriteShifts[index].load(&p.spriteLatch, flipH)
		p.spriteShifts[index].attributes = attributes
		p.spriteShifts[index].xDistance = spriteX
		p.spriteShifts[index].isSpriteZero = (index == p.spriteZeroIndex)
	}
}

// MARK: スプライトシフトレジスタのシフト
func (p *PPU) shiftSpriteRegisters() {
	for i := range p.spriteShifts {
		if p.spriteShifts[i].xDistance > 0 {
			// 描画位置までは描画位置カウンタを減算
			p.spriteShifts[i].xDistance--
		} else {
			// 描画位置になったら描画するドットを送り始める
			p.spriteShifts[i].shift()
		}
	}
}

// MARK: ピクセルの書き込み
func (p *PPU) renderPixel() {
	screenX := p.dot - 1
	screenY := p.scanline

	// 現在のドットの背景・スプライトパターン，透過状態を取得
	bgPattern, bgAttribute := p.getBackgroundPixel()
	spPattern, spAttributes, isSpriteZero, spPriority := p.getSpritePixel()

	bgOpaque := (bgPattern != 0)
	spOpaque := (spPattern != 0)

	// スプライト0ヒット判定
	if isSpriteZero && bgOpaque && spOpaque && screenX < SCREEN_WIDTH-1 {
		// 左端タイルのマスクがされている時は左端ではスプライト0ヒットが起こらない
		leftMasked := !p.mask.leftmostBackgroundEnable || !p.mask.leftmostSpriteEnable
		if !(leftMasked && screenX < TILE_SIZE) {
			p.status.SetSpriteZeroHit(true)
		}
	}

	// 描画設定の適用
	bgOpaque = bgOpaque && p.config.Ppu.BACKGROUND_ENABLED
	spOpaque = spOpaque && p.config.Ppu.SPRITE_ENABLED

	// 優先順位に基づいて色を決定
	var color sdl.Color
	switch {
	case !bgOpaque && !spOpaque: // 両方透明
		address := p.v.ToWord() & PPU_MEMORY_ADDRESS_MASK
		if !p.mask.backgroundEnable && !p.mask.spriteEnable && 0x3F00 <= address && address <= 0x3FFF {
			paletteTableIndex := (address - 0x3F00) % PPU_PALETTE_TABLE_SIZE
			if paletteTableIndex >= 0x10 && paletteTableIndex%4 == 0 {
				paletteTableIndex -= 0x10 // $3F10, $3F14, $3F18, $3F1C は $3F00 番台にミラーされる
			}
			color = PALETTE[p.paletteTable[paletteTableIndex]]
		} else {
			color = PALETTE[p.paletteTable[0x00]]
		}
	case !bgOpaque && spOpaque: // 背景のみ透明
		color = p.getSpriteColor(spAttributes, spPattern)
	case bgOpaque && !spOpaque: // スプライトのみ透明
		color = p.getBackgroundColor(bgAttribute, bgPattern)
	case bgOpaque && spOpaque: // 両方不透明
		// スプライト優先の場合
		if spPriority == 0 {
			color = p.getSpriteColor(spAttributes, spPattern)
		} else {
			color = p.getBackgroundColor(bgAttribute, bgPattern)
		}
	}

	// 決定した色を描画
	p.canvas.SetPixel(screenX, screenY, p.getEmphasizedColor(color))
}

// MARK: VRAMアドレスをミラーリング
func (p *PPU) mirrorVRAMAddress(address uint16) uint16 {
	/*
		VRAM ネームテーブル
		(範囲 / サイズ / 対象)

		$2000-$2400 0x0400 画面1
		$2400-$2800 0x0400 画面2
	*/

	vramAddress := address - 0x2000 // 先頭オフセットを引きVRAMのアドレスに変換
	mirroring := p.mapper.Mirroring()

	/*
		ネームテーブルの位置を求める
		[ 0 ][ 1 ]
		[ 2 ][ 3 ]
	*/
	nameTableIndex := vramAddress / 0x0400

	switch mirroring {
	case mappers.MIRRORING_VERTICAL:
		/*
			[ A ][ B ] $2000 $2400
			[ a ][ b ] $2800 $2C00

			A: $2000
			a: $2800 → $2000
			B: $2400
			b: $2C00 → $2400
		*/

		switch nameTableIndex {
		case 2, 3:
			vramAddress -= 0x0800
		}
	case mappers.MIRRORING_HORIZONTAL:
		/*
			[ A ][ a ] $2000 $2400
			[ B ][ b ] $2800 $2C00

			A: $2000
			a: $2400 → $2000
			B: $2800 → $2400
			b: $2C00 → $2400
		*/

		switch nameTableIndex {
		case 1, 2:
			vramAddress -= 0x0400
		case 3:
			vramAddress -= 0x0800
		}
	}

	return vramAddress
}

// MARK: VRAMアドレスのインクリメント
func (p *PPU) incrementVRAMAddress() {
	step := uint16(p.control.VRAMAddressIncrement())
	address := (p.v.ToWord() + step)
	p.v.SetFromWord(address & PPU_MEMORY_ADDRESS_MASK)
}

// MARK: サイクルを進める
func (p *PPU) incrementCycles() {
	// ドットを進める
	p.dot++
	if p.nmiCounter > 0 {
		p.nmiCounter--
	}

	// 端に達したらドットを0に戻しスキャンラインを進める
	if p.dot > SCANLINE_END {
		p.dot = 0
		p.scanline++

		// プリレンダーラインに達したらスキャンラインを0に戻しフレーム数を進める
		if p.scanline > SCANLINE_PRERENDER {
			p.scanline = 0
			p.frame++
			p.isOddFrame = !p.isOddFrame
		}
	}
}

// MARK: セカンダリOAMのクリア
func (p *PPU) clearSecondaryOAM() {
	for i := range p.secondaryOAM {
		p.secondaryOAM[i] = 0xFF
	}
	p.spriteZeroIndex = SPRITE_ZERO_NOT_FOUND
}

// MARK: Vレジスタから描画中の行・列のネームテーブルのアドレスを取得
func (p *PPU) getNameTableAddress() uint16 {
	/*
		基準となるネームテーブルの始点 ($2000 | $2400 | $2800 | $2C00) と
		Vレジスタの下位12ビット(NN YYYYY XXXXX)の論理和で現在の行・列のアドレスが求まる
	*/
	return 0x2000 | (p.v.ToWord() & 0x0FFF)
}

// MARK: Vレジスタから描画中の属性テーブルのアドレスを取得
func (p *PPU) getAttributeTableAddress() uint16 {
	/*
		ベースアドレス: $23C0 (ネームテーブルの末尾)
		p.v & 0x0C00: どの画面（ネームテーブル）かを選択
		(v >> 4) & 0x38: Y座標 (行) をメタタイル単位(4行ごと)に変換
		(v >> 2) & 0x07: X座標 (列) をメタタイル単位(4列ごと)に変換
	*/

	v := p.v.ToWord()
	return 0x23C0 | (v & 0x0C00) | ((v >> 4) & 0x38) | ((v >> 2) & 0x07)
}

// MARK: Vレジスタから描画中の背景タイルのパターンテーブルのアドレスを取得
func (p *PPU) getBackgroundPatternAddress(isUpper bool) uint16 {
	base := p.control.BackgroundPatternTableAddress()
	tile := p.backgroundLatch.nameTable
	fineY := (p.v.ToWord() >> 12) & 0x07
	offset := 0
	if isUpper {
		offset = TILE_SIZE
	}
	return base + (uint16(tile) * TILE_SIZE * 2) + fineY + uint16(offset)
}

// MARK: 描画中のスプライトタイルのパターンテーブルのアドレスを取得
func (p *PPU) getSpritePatternAddress(index uint, isUpper bool) uint16 {
	// スプライトの情報を取得
	secondaryBase := index * OAM_SPRITE_SIZE
	spriteY := p.secondaryOAM[secondaryBase+OAM_SPRITE_Y_POS] + 1 // OAMのY座標は表示座標-1のため補正
	spriteTile := p.secondaryOAM[secondaryBase+OAM_SPRITE_TILE_POS]
	attributes := p.secondaryOAM[secondaryBase+OAM_SPRITE_ATTR_POS]
	spriteHeight := p.control.SpriteSize()

	targetScanline := p.scanline
	// 画面外の場合は次スキャンラインを対象にする
	if 257 <= p.dot && p.dot <= 320 {
		targetScanline++
		// プリレンダーラインを超えた場合次フレームの0ライン目にする
		if targetScanline > SCANLINE_PRERENDER {
			targetScanline = 0
		}
	}

	// タイル内の行数
	tileY := uint16(targetScanline - uint(spriteY))

	// 垂直反転の場合はタイル内行数も反転
	flipV := ((attributes >> 7) & 0x01) == 1
	if flipV {
		tileY = (uint16(spriteHeight) - 1) - tileY
	}

	offset := 0
	if isUpper {
		offset = TILE_SIZE
	}

	var base uint16
	if spriteHeight == TILE_SIZE {
		// 8x8の場合
		base = p.control.SpritePatternTableAddress()
	} else {
		// 8x16の場合
		base = uint16(spriteTile&0x01) * 0x1000 // タイル番号のビット0がネームテーブル選択
		spriteTile = spriteTile & 0xFE
		if tileY >= TILE_SIZE {
			spriteTile++
			tileY -= TILE_SIZE
		}
	}

	return base + (uint16(spriteTile) * TILE_SIZE * 2) + tileY + uint16(offset)
}

// MARK: 背景ピクセルの取得
func (p *PPU) getBackgroundPixel() (pixel uint8, attribute uint8) {
	// 背景描画が無効の場合は 0 を返す
	if !p.mask.backgroundEnable {
		return 0x00, 0x00
	}

	// 左端8ピクセルの背景描画が無効でX座標が8より小さい場合は 0 を返す
	screenX := p.dot - 1
	if screenX < TILE_SIZE && !p.mask.leftmostBackgroundEnable {
		return 0x00, 0x00
	}

	shift := (TILE_SIZE*2 - 1) - p.x.fineX

	// ピクセルの upper / lower の見ているビットを取り出し
	bit0 := (p.backgroundShift.patternLower >> uint16(shift)) & 0x01
	bit1 := (p.backgroundShift.patternUpper >> uint16(shift)) & 0x01
	pixel = uint8((bit1 << 1) | bit0)

	// 属性情報の upper / lower の見ているビットを取り出し
	attrBit0 := (p.backgroundShift.attributeLower >> uint16(shift)) & 0x01
	attrBit1 := (p.backgroundShift.attributeUpper >> uint16(shift)) & 0x01
	attribute = uint8((attrBit1 << 1) | attrBit0)

	return pixel, attribute
}

// MARK: スプライトピクセルの取得
func (p *PPU) getSpritePixel() (pixel uint8, attributes uint8, isSpriteZero bool, priority uint8) {
	// スプライト描画が無効の場合は 0 を返す
	if !p.mask.spriteEnable {
		return 0x00, 0x00, false, 0x00
	}

	// 左端8ピクセルのスプライト描画が無効でX座標が8より小さい場合は 0 を返す
	screenX := p.dot - 1
	if screenX < TILE_SIZE && !p.mask.leftmostSpriteEnable {
		return 0x00, 0x00, false, 0x00
	}

	shift := TILE_SIZE - 1
	for i := range len(p.spriteShifts) {
		if p.spriteShifts[i].xDistance == 0 {
			// シフトレジスタから1ビットを取り出し
			bit0 := (p.spriteShifts[i].patternLower >> shift) & 0x01
			bit1 := (p.spriteShifts[i].patternUpper >> shift) & 0x01
			pixel = uint8((bit1 << 1) | bit0)

			// 透明なピクセルであれば無視
			if pixel == 0 {
				continue
			}

			// 透明でなければこのピクセルの情報を返す
			attributes = p.spriteShifts[i].attributes
			isSpriteZero = p.spriteShifts[i].isSpriteZero
			priority = (attributes >> 5) & 0x01
			return pixel, attributes, isSpriteZero, priority
		}
	}
	return 0x00, 0x00, false, 0x00
}

// MARK: 属性情報とピクセルの値から色を取得するメソッド
func (p *PPU) getBackgroundColor(attribute uint8, pattern uint8) sdl.Color {
	// 透明の場合は背景色を返す
	if pattern == 0x00 {
		paletteIndex := p.paletteTable[0x00]
		return PALETTE[paletteIndex]
	}

	paletteTableIndex := ((attribute & 0x03) << 2) + pattern
	paletteIndex := p.paletteTable[paletteTableIndex]
	return PALETTE[paletteIndex]
}

// MARK: 属性情報とピクセルの値からスプライトの色を取得するメソッド
func (p *PPU) getSpriteColor(attributes uint8, pattern uint8) sdl.Color {
	// 透明の場合は背景色を返す
	if pattern == 0x00 {
		paletteIndex := p.paletteTable[0x00]
		return PALETTE[paletteIndex]
	}

	// スプライトパレットの開始位置 0x10 をオフセットとして加算
	paletteTableIndex := 0x10 + ((attributes & 0x03) << 2) + pattern
	paletteIndex := p.paletteTable[paletteTableIndex]
	return PALETTE[paletteIndex]
}

// MARK: マスクレジスタの値から色強調を反映した色を取得するメソッド
func (p *PPU) getEmphasizedColor(baseColor sdl.Color) sdl.Color {
	if !p.mask.emphasizeRed && !p.mask.emphasizeGreen && !p.mask.emphasizeBlue {
		return baseColor
	}

	r := baseColor.R
	g := baseColor.G
	b := baseColor.B

	if !p.mask.emphasizeRed {
		r = uint8(float32(r) * COLOR_EMPHASIZE_FACTOR)
	}
	if !p.mask.emphasizeGreen {
		g = uint8(float32(g) * COLOR_EMPHASIZE_FACTOR)
	}
	if !p.mask.emphasizeBlue {
		b = uint8(float32(b) * COLOR_EMPHASIZE_FACTOR)
	}

	return sdl.Color{R: r, G: g, B: b}
}

// MARK: 待機中のNMI状態をチェックするメソッド
func (p *PPU) PollNMI() bool {
	if p.nmi && p.nmiCounter == 0 {
		p.nmi = false
		p.nmiCounter = 0
		return true
	} else {
		return false
	}
}

// MARK: オープンバスの読み取り
func (p *PPU) ReadOpenBus() uint8 {
	return p.openBus
}

// MARK: オープンバスのリフレッシュ
func (p *PPU) refreshOpenBus(value uint8) {
	p.openBus = value
	p.openBusDecayTimer = OPEN_BUS_DECAY_CYCLES
}

// MARK: NMIを取得するメソッド
func (p *PPU) NMI() bool {
	return p.nmi && p.nmiCounter == 0
}

// MARK: 経過フレーム数を取得するメソッド
func (p *PPU) Frame() uint64 {
	return p.frame
}

// MARK: スプライトパターンテーブルアドレスの取得メソッド
func (p *PPU) SpritePatternTableAddress() uint16 {
	return p.control.SpritePatternTableAddress()
}

// MARK: 背景パターンテーブルアドレスの取得メソッド
func (p *PPU) BackgroundPatternTableAddress() uint16 {
	return p.control.BackgroundPatternTableAddress()
}

// MARK: VRAMの取得メソッド
func (p *PPU) Vram() *[PPU_VRAM_SIZE]uint8 {
	return &p.vram
}

// MARK: OAMの取得メソッド
func (p *PPU) Oam() *[PPU_PRIMARY_OAM_SIZE]uint8 {
	return &p.oam
}

// MARK: パレットテーブルの取得メソッド
func (p *PPU) PaletteTable() *[PPU_PALETTE_TABLE_SIZE]uint8 {
	return &p.paletteTable
}

// MARK: 背景パレットの取得メソッド
func (p *PPU) BackgroundColorPalette(attributeTable *[]uint8, tileColumn, tileRow uint) [4]uint8 {
	attrTableIdx := tileRow/4*TILE_SIZE + tileColumn/4
	attrByte := (*attributeTable)[attrTableIdx]

	var paletteIdx uint8
	if tileColumn%4/2 == 0 && tileRow%4/2 == 0 {
		paletteIdx = (attrByte) & 0b11
	} else if tileColumn%4/2 == 1 && tileRow%4/2 == 0 {
		paletteIdx = (attrByte >> 2) & 0b11
	} else if tileColumn%4/2 == 0 && tileRow%4/2 == 1 {
		paletteIdx = (attrByte >> 4) & 0b11
	} else if tileColumn%4/2 == 1 && tileRow%4/2 == 1 {
		paletteIdx = (attrByte >> 6) & 0b11
	} else {
		panic("Error: unexpected palette value")
	}

	var paletteStart uint = 1 + uint(paletteIdx)*4
	color := [4]uint8{
		p.paletteTable[0],
		p.paletteTable[paletteStart+0],
		p.paletteTable[paletteStart+1],
		p.paletteTable[paletteStart+2],
	}

	return color
}

// MARK: スプライトサイズの取得メソッド
func (p *PPU) SpriteSize() uint8 {
	return p.control.SpriteSize()
}

// MARK: マッパー取得メソッド
func (p *PPU) Mapper() mappers.Mapper {
	return p.mapper
}

// MARK: コンフィグのBG描画有効/無効の取得メソッド
func (p *PPU) ConfigBackgroundEnabled() bool {
	return p.config.Ppu.BACKGROUND_ENABLED
}

// MARK: コンフィグのBG描画有効/無効の切り替えメソッド
func (p *PPU) ToggleBackgroundEnabled() {
	if p.config.Ppu.BACKGROUND_ENABLED {
		fmt.Println("[PPU] Background: OFF")
	} else {
		fmt.Println("[PPU] Background: ON")
	}
	p.config.Ppu.BACKGROUND_ENABLED = !p.config.Ppu.BACKGROUND_ENABLED
}

// MARK: コンフィグのSprite描画有効/無効の取得メソッド
func (p *PPU) ConfigSpriteEnabled() bool {
	return p.config.Ppu.SPRITE_ENABLED
}

// MARK: コンフィグのスプライト描画有効/無効の切り替えメソッド
func (p *PPU) ToggleSpriteEnabled() {
	if p.config.Ppu.SPRITE_ENABLED {
		fmt.Println("[PPU] Sprite: OFF")
	} else {
		fmt.Println("[PPU] Sprite: ON")
	}
	p.config.Ppu.SPRITE_ENABLED = !p.config.Ppu.SPRITE_ENABLED
}
