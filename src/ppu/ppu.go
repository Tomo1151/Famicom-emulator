package ppu

import (
	"Famicom-emulator/cartridge/mappers"
	"fmt"
)

// MARK: 定数定義
const (
	VRAM_SIZE          uint16 = 2 * 1024 // 2kB
	PALETTE_TABLE_SIZE uint8  = 32
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

	TILE_SIZE uint = 8
)

// MARK: PPUの定義
type PPU struct {
	Mapper       mappers.Mapper
	paletteTable [PALETTE_TABLE_SIZE + 1]uint8
	vram         [VRAM_SIZE]uint8
	oam          [OAM_DATA_SIZE]uint8

	// IOレジスタ
	control ControlRegister // $2000
	mask    MaskRegister    // $2001
	status  StatusRegister  // $2002

	// 内部レジスタ
	t InternalAddressRegiseter // 一時的な VRAM アドレスレジスタ
	v InternalAddressRegiseter // 現在の VRAM アドレスレジスタ
	x InternalXRegister        // x スクロール
	w InternalWRegister        // 書き込みラッチ

	// 1ライン描画開始時点のv（レンダラ用スナップショット）
	vLineStart InternalAddressRegiseter

	scanline           uint16 // 現在描画中のスキャンライン
	cycles             uint   // PPUサイクル
	internalDataBuffer uint8
	oamAddress         uint8 // OAM書き込みのポインタ

	nmi bool

	lineBuffer [SCREEN_WIDTH]Pixel // 次のスキャンラインのバッファ

	// デバッグウィンドウ用のスナップショット
	mapperSnapshots []mappers.Mapper
	vLineSnapshots  []InternalAddressRegiseter
}

// MARK: PPUの初期化メソッド
func (p *PPU) Init(mapper mappers.Mapper) {
	p.Mapper = mapper

	// VRAM/OAM/パレットの初期化
	for addr := range p.vram {
		p.vram[addr] = 0x00
	}
	for addr := range p.oam {
		p.oam[addr] = 0x00
	}
	for addr := range p.paletteTable {
		p.paletteTable[addr] = 0x00
	}

	// IOレジスタの初期化
	p.control.Init()
	p.mask.Init()
	p.status.Init()

	// 内部レジスタの初期化
	p.t.Init()
	p.v.Init()
	p.x.Init()
	p.w.Init()

	// vLineStartも初期化
	p.vLineStart.Init()

	p.oamAddress = 0
	p.scanline = 0
	p.cycles = 0
	p.internalDataBuffer = 0x00

	p.nmi = false

	// ラインバッファの初期化
	for i := range p.lineBuffer {
		p.lineBuffer[i] = Pixel{
			0x00,                       // priority
			PALETTE[p.paletteTable[0]], // background value (rgb palette)
			PALETTE[p.paletteTable[0]], // sprite value (rgb palette)
			true,                       // background transparent
			true,                       // sprite transparent
		}
	}

	p.mapperSnapshots = make([]mappers.Mapper, SCREEN_HEIGHT)
	p.vLineSnapshots = make([]InternalAddressRegiseter, SCREEN_HEIGHT)
}

// MARK: PPUコントロールレジスタ($2000)への書き込み
func (p *PPU) WriteToPPUControlRegister(value uint8) {
	prev := p.control.GenerateNMI()
	p.control.update(value)

	// tレジスタのネームテーブルビットを更新
	p.t.updateNameTable(value)

	// VBlank中にGenerateNMIが立つタイミングでNMIを発生させる
	if !prev && p.control.GenerateNMI() && p.status.VBlank() {
		p.nmi = true
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

// MARK: PPU内部レジスタへ(T/V/X/W)の書き込み
func (p *PPU) WriteToPPUInternalRegister(address uint16, data uint8) {
	switch address {
	case 0x2005: // PPU_SCROLL
		if !p.w.latch {
			p.x.update(data)
		}
		p.t.updateScroll(data, &p.w)
	case 0x2006: // PPU_ADDR
		beforeLatch := p.w.latch
		p.t.updateAddress(data, &p.w)

		if beforeLatch && !p.w.latch {
			p.t.copyAllBitsTo(&p.v)
		}
	}
}

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
	step := uint16(p.control.VRAMAddressIncrement())
	newAddr := (p.v.ToByte() + step) & 0x3FFF // 14ビットでマスク
	p.v.SetFromWord(newAddr)
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

	address := p.v.ToByte()
	p.incrementVRAMAddress()

	// $0000-$3FFF のミラーリング
	if address > 0x3FFF {
		address -= 0x4000
	}

	switch {
	case address <= 0x1FFF: // キャラクタROM
		if p.Mapper.IsCharacterRam() {
			p.Mapper.WriteToCharacterRom(address, value)
		}
	case 0x2000 <= address && address <= 0x2FFF: // VRAM
		p.vram[p.mirrorVRAMAddress(address)] = value
	case 0x3000 <= address && address <= 0x3EFF: // ネームテーブル
		return
	case 0x3F00 <= address && address <= 0x3F1F: // パレット
		// アドレスのミラーリング
		if address == 0x3F10 ||
			address == 0x3F14 ||
			address == 0x3F18 ||
			address == 0x3F1C {
			address -= 0x10
		}
		p.paletteTable[address-0x3F00] = value
	case 0x3F20 <= address && address <= 0x3FFF: // パレット (ミラーリング)
		p.paletteTable[(address-0x3F00)%32] = value
	default:
		panic(fmt.Sprintf("Unexpected write to vram space: %04X", address))
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
	p.w.reset()
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

	address := p.v.ToByte()
	p.incrementVRAMAddress()

	// $0000-$3FFF のミラーリング
	if address > 0x3FFF {
		address -= 0x4000
	}

	switch {
	case address <= 0x1FFF: // キャラクタROM
		value := p.internalDataBuffer
		p.internalDataBuffer = p.Mapper.ReadCharacterRom(address)
		return value
	case 0x2000 <= address && address <= 0x2FFF: // VRAM
		// 一回遅れで値は反映されるため，内部バッファを更新し，元のバッファ値を返す
		value := p.internalDataBuffer
		p.internalDataBuffer = p.vram[p.mirrorVRAMAddress(address)]
		return value
	case 0x3000 <= address && address <= 0x3EFF: // ネームテーブル
		panic(fmt.Sprintf("Error: address space 0x3000..0x3eff is not expected to read, requested: %04X", address))
	case 0x3F00 <= address && address <= 0x3F1F: // パレット
		// アドレスのミラーリング
		if address == 0x3F10 ||
			address == 0x3F14 ||
			address == 0x3F18 ||
			address == 0x3F1C {
			address -= 0x10
		}
		return p.paletteTable[address-0x3F00]
	case 0x3F20 <= address && address <= 0x3FFF: // パレット (ミラーリング)
		return p.paletteTable[(address-0x3F00)%32]
	default:
		panic(fmt.Sprintf("Error: unexpected read to vram space: %04X", address))
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

	mirroring := p.Mapper.Mirroring()

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
func (p *PPU) NMI() bool {
	if p.nmi {
		p.nmi = false
		return true
	} else {
		return false
	}
}

// MARK: 待機しているNMIを確認
func (p *PPU) CheckNMI() bool {
	return p.nmi
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
	return p.mask.spriteEnable && y == uint(p.scanline) && x <= cycles
}

// MARK: BG面のカラーパレットを取得
func (p *PPU) bgPalette(attrributeTable *[]uint8, tileColumn uint, tileRow uint) [4]uint8 {
	attrTableIdx := tileRow/4*TILE_SIZE + tileColumn/4
	attrByte := (*attrributeTable)[attrTableIdx]

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

// MARK: スプライトのカラーパレットを取得
func (p *PPU) spritePalette(paletteIndex uint8) [4]uint8 {
	var start uint = 0x11 + uint(paletteIndex*4)
	return [4]uint8{
		0,
		p.paletteTable[start+0],
		p.paletteTable[start+1],
		p.paletteTable[start+2],
	}
}

// MARK: ラインバッファをクリア
func (p *PPU) ClearLineBuffer() {
	for x := range p.lineBuffer {
		p.lineBuffer[x].backgroundValue = PALETTE[p.paletteTable[0]]
		p.lineBuffer[x].spriteValue = PALETTE[p.paletteTable[0]]
		p.lineBuffer[x].priority = 0x00
		p.lineBuffer[x].isSpriteTransparent = true
	}
}

// MARK: 指定したスキャンラインに重なるスプライトを探索
func (p *PPU) FindScanlineSprite(spriteHeight uint8, scanline uint16) (uint, *[SPRITE_MAX][OAM_SPRITE_SIZE]uint8) {
	var sprites [SPRITE_MAX][OAM_SPRITE_SIZE]uint8 // 1スキャンラインに配置するスプライト (8個まで)

	var spriteCount uint = 0
	for i := range len(p.oam) / 4 {
		index := uint(i * 4)
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
		if scanline >= spriteY && scanline < spriteY+uint16(spriteHeight) {
			if spriteCount < SPRITE_MAX {
				sprites[spriteCount][OAM_SPRITE_Y] = p.oam[index+OAM_SPRITE_Y]       // Y座標
				sprites[spriteCount][OAM_SPRITE_TILE] = p.oam[index+OAM_SPRITE_TILE] // タイル選択
				sprites[spriteCount][OAM_SPRITE_ATTR] = p.oam[index+OAM_SPRITE_ATTR] // 属性
				sprites[spriteCount][OAM_SPRITE_X] = p.oam[index+OAM_SPRITE_X]       // X座標
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

// MARK: スキャンライン開始時点のVレジスタからネームテーブルを取得
func (p *PPU) nameTable(v InternalAddressRegiseter) *[]uint8 {
	nameTableIndex := v.nameTable
	var nameTable []uint8

	primaryNameTable := p.vram[0x000:0x400]
	secondaryNameTable := p.vram[0x400:0x800]

	mirroring := p.Mapper.Mirroring()
	switch mirroring {
	case mappers.MIRRORING_VERTICAL:
		if nameTableIndex == 0 || nameTableIndex == 2 {
			nameTable = primaryNameTable
		} else {
			nameTable = secondaryNameTable
		}
	case mappers.MIRRORING_HORIZONTAL:
		if nameTableIndex == 0 || nameTableIndex == 1 {
			nameTable = primaryNameTable
		} else {
			nameTable = secondaryNameTable
		}
	default:
		nameTable = primaryNameTable
	}
	return &nameTable
}

// MARK: 指定したスキャンラインのBG面を計算
func (p *PPU) CalculateScanlineBackground(canvas *Canvas, scanline uint16) {
	// BGが無効であれば描画をしない
	if !p.mask.backgroundEnable {
		return
	}

	// 現在のVレジスタの状態をバックアップ（ライン開始時点の値を使う）
	v := p.vLineStart

	// 画面の左端から右端まで
	fineX := uint(p.x.fineX) // ここからはローカルで進める。p.xは書き換えない
	for x := range SCREEN_WIDTH {
		// 左端8pxの描画有無を判定（描画はしないが、アドレスの前進は必要）
		if p.mask.leftmostBackgroundEnable || x >= TILE_SIZE {
			// 現在のピクセル位置でのタイル座標を計算
			tileX := uint(v.coarseX)
			tileY := uint(v.coarseY)
			fineY := uint(v.fineY)

			// ネームテーブルの選択
			nameTable := *p.nameTable(v)

			// タイルのインデックスを取得
			tileIndex := uint16(nameTable[tileY*32+tileX])

			// 属性テーブルからパレット情報を取得
			attributeTable := nameTable[0x3C0:0x400]
			palette := p.bgPalette(&attributeTable, tileX, tileY)

			// パターンテーブルからタイルのピクセルデータを取得
			bank := p.control.BackgroundPatternTableAddress()
			tileBasePointer := bank + tileIndex*uint16(TILE_SIZE*2)

			upper := p.Mapper.ReadCharacterRom(tileBasePointer + uint16(fineY))
			lower := p.Mapper.ReadCharacterRom(tileBasePointer + uint16(fineY) + uint16(TILE_SIZE))

			// ピクセル位置を計算（fineXを使用）
			pixelIndex := (7 - (fineX % 8)) // 0..7
			value := ((lower>>pixelIndex)&1)<<1 | ((upper >> pixelIndex) & 1)
			color := PALETTE[palette[value]]

			// ラインバッファに登録
			p.lineBuffer[x].backgroundValue = color
			p.lineBuffer[x].priority = 0x00
			p.lineBuffer[x].isBgTransparent = color == PALETTE[p.paletteTable[0]]
		}

		// 次のピクセルへ進む（描画しない場合でも必ず進める）
		fineX++
		if fineX%8 == 0 {
			// タイル境界を越えたらタイルを進める
			v.incrementCoarseX()
		}
	}
}

// MARK: 指定したスキャンラインのスプライトを計算
func (p *PPU) CalculateScanlineSprite(canvas *Canvas, scanline uint16) {
	// スプライトが無効であれば描画しない
	if !p.mask.spriteEnable {
		return
	}

	// スプライトサイズの取得 (8 / 16)
	spriteHeight := p.control.SpriteSize()
	spriteCount, sprites := p.FindScanlineSprite(spriteHeight, scanline)

	// スプライトの描画
	for i := range spriteCount {
		// 逆順に評価する (重なり順のため)
		index := (spriteCount - 1) - i

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

		flipV := (attributes>>7)&1 == 1
		flipH := (attributes>>6)&1 == 1
		paletteIndex := attributes & 0b11
		palette := p.spritePalette(paletteIndex)

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
			bank = p.control.SpritePatternTableAddress()
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
		tileBasePointer := bank + tileIndex*uint16(TILE_SIZE*2)
		upper := p.Mapper.ReadCharacterRom(tileBasePointer + tileY)
		lower := p.Mapper.ReadCharacterRom(tileBasePointer + tileY + uint16(TILE_SIZE))

		// タイルデータを描画
		for x := range TILE_SIZE {
			var value uint8
			if flipH {
				// 水平反転の場合
				value = (lower&1)<<1 | (upper & 1)
				upper >>= 1
				lower >>= 1
			} else {
				// 反転がない場合
				value = ((lower>>7)&1)<<1 | ((upper >> 7) & 1)
				upper <<= 1
				lower <<= 1
			}

			// 透明ピクセルは描画しない
			if value == 0 {
				// @FIXME 飛ばすとOAMでの順番が若い透明なピクセルで上書きできない
				continue
			}

			actualX := uint(spriteX) + uint(x)

			// 画面外のピクセルは描画しない
			if actualX >= SCREEN_WIDTH {
				continue
			}

			// 左端のスプライト描画フラグが無効であれば描画しない
			if !p.mask.leftmostSpriteEnable && actualX < TILE_SIZE {
				continue
			}

			// スプライトの優先度が0または背景が透明であれば描画
			// @FIXME BG面より優先されないスプライトに，OAM上の順番が後ろが重なった時にBG面が最優先になるようにする (SMB3のパックンフラワー等)
			p.lineBuffer[actualX].spriteValue = PALETTE[palette[value]]
			p.lineBuffer[actualX].priority = priority
			p.lineBuffer[actualX].isSpriteTransparent = false
		}
	}
}

// MARK: PPUを動かす
func (p *PPU) Tick(canvas *Canvas, cycles uint) bool {
	// サイクルを進める
	p.cycles += cycles

	// ライン開始(サイクル1)でvのスナップショットを取る
	if p.cycles == 1 {
		// vは直前の321–336サイクルで2タイル進んでいるので、
		// 描画用スナップショットはtの水平ビットで補正して使用する
		p.vLineStart = p.v
		p.t.copyHorizontalBitsTo(&p.vLineStart)

		// vLineStart のデバッグウィンドウ用スナップショットを保存
		idx := int(p.scanline)
		if idx >= 0 && idx < len(p.vLineSnapshots) {
			p.vLineSnapshots[idx] = p.vLineStart
		}
	}

	isRenderingEnabled := p.mask.backgroundEnable || p.mask.spriteEnable
	isRenderLine := (SCANLINE_START <= p.scanline && p.scanline < SCANLINE_POSTRENDER)
	isPreRenderLine := p.scanline == SCANLINE_PRERENDER

	// スプライト0ヒットの判定（適切なタイミングで）
	if isRenderLine && p.isSpriteZeroHit(p.cycles) {
		p.status.SetSpriteZeroHit(true)
	}

	if isRenderingEnabled {
		// レンダリング中のサイクル処理
		if isRenderLine || isPreRenderLine {
			// 1-256サイクル: 各タイルをフェッチする間に水平アドレスをインクリメント
			if p.cycles >= 1 && p.cycles <= 256 {
				// 8サイクル毎（タイルフェッチ完了時）に水平アドレスをインクリメント
				if p.cycles%TILE_SIZE == 0 {
					p.v.incrementCoarseX()
				}
			}

			// 256サイクル: 垂直アドレスをインクリメント
			if p.cycles == 256 {
				p.v.incrementY()
			}

			// 257サイクル: 水平ビットのコピー (t -> v)
			if p.cycles == 257 {
				p.t.copyHorizontalBitsTo(&p.v)
			}

			// 321-336サイクル: 次のスキャンライン準備のため水平アドレスをインクリメント
			if p.cycles >= 321 && p.cycles <= 336 {
				if p.cycles%TILE_SIZE == 0 {
					p.v.incrementCoarseX()
				}
			}
		}

		// プリレンダーラインでのみ垂直ビットをコピー (t -> v)
		if isPreRenderLine && p.cycles >= 280 && p.cycles <= 304 {
			p.t.copyVerticalBitsTo(&p.v)
		}
	}

	if p.cycles >= SCANLINE_END {
		// サイクル数をリセット
		p.cycles = 0

		// マッパーによるIRQの判定
		p.Mapper.GenerateScanlineIRQ(p.scanline, isRenderingEnabled)

		// 可視領域のスキャンラインを描画
		if SCANLINE_START <= p.scanline && p.scanline < SCANLINE_POSTRENDER {
			RenderScanlineToCanvas(p, canvas, p.scanline)

			// デバッグウィンドウ用のマッパースナップショットを保存
			p.takeMapperSnapshot(p.scanline)
		}

		// スキャンラインを進める
		p.scanline++

		// VBlankに突入
		if p.scanline == SCANLINE_VBLANK {
			p.status.SetVBlankStatus(true)
			if p.control.GenerateNMI() {
				// NMIを設定
				p.nmi = true
			}
		}

		// プリレンダーラインに到達した時（フレーム終了）
		if p.scanline > SCANLINE_PRERENDER {
			p.scanline = 0
			p.nmi = false
			p.status.SetSpriteZeroHit(false)
			p.status.ClearVBlankStatus()
			return true
		}
	}
	return false
}

// MARK: PaletteTable の取得メソッド
func (p *PPU) PaletteTable() *[PALETTE_TABLE_SIZE + 1]uint8 {
	return &p.paletteTable
}

// MARK: VRAM の取得メソッド
func (p *PPU) VRAM() *[VRAM_SIZE]uint8 {
	return &p.vram
}

// MARK: 指定したスキャンラインのマッパーのスナップショットを保存
func (p *PPU) takeMapperSnapshot(scanline uint16) {
	idx := int(scanline)
	if idx < 0 || idx >= len(p.mapperSnapshots) {
		return
	}
	if p.Mapper == nil {
		p.mapperSnapshots[idx] = nil
		return
	}
	p.mapperSnapshots[idx] = p.Mapper.Clone()
}

// MARK: 指定したスキャンラインのマッパースナップショットを取得
func (p *PPU) GetMapperForScanline(scanline uint16) mappers.Mapper {
	idx := int(scanline)
	if idx >= 0 && idx < len(p.mapperSnapshots) && p.mapperSnapshots[idx] != nil {
		return p.mapperSnapshots[idx]
	}
	return p.Mapper
}

// MARK: 指定したスキャンラインのvLineStart のスナップショットを取得
func (p *PPU) GetVLineSnapshot(scanline uint16) InternalAddressRegiseter {
	idx := int(scanline)
	if idx >= 0 && idx < len(p.vLineSnapshots) {
		return p.vLineSnapshots[idx]
	}
	return p.vLineStart
}

// MARK: OAM の取得メソッド
func (p *PPU) OAM() *[OAM_DATA_SIZE]uint8 {
	return &p.oam
}

// MARK: BG pattern table address の取得メソッド
func (p *PPU) BackgroundPatternTableAddress() uint16 {
	return p.control.BackgroundPatternTableAddress()
}
