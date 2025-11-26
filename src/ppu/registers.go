package ppu

// import "Famicom-emulator/bus"

// PPUレジスタ関連の定数
const (
	// PPUのI/Oレジスタのアドレス
	PPU_CTRL   uint16 = 0x2000
	PPU_MASK   uint16 = 0x2001
	PPU_STATUS uint16 = 0x2002
	OAM_ADDR   uint16 = 0x2003
	OAM_DATA   uint16 = 0x2004
	PPU_SCROLL uint16 = 0x2005
	PPU_ADDR   uint16 = 0x2006
	PPU_DATA   uint16 = 0x2007
	OAM_DMA    uint16 = 0x4014

	PPU_REG_START = 0x2000
	PPU_REG_END   = 0x3FFF

	PPU_VRAM_MIRROR_MASK = 0b101111_11111111
)

const (
	// コントロールレジスタ ($2000) のビット位置
	CONTROL_REG_NAMETABLE1_POS             uint8 = 0
	CONTROL_REG_NAMETABLE2_POS             uint8 = 1
	CONTROL_REG_VRAM_ADD_INCREMENT_POS     uint8 = 2
	CONTROL_REG_SPRITE_PATTERN_ADDR_POS    uint8 = 3
	CONTROL_REG_BACKROUND_PATTERN_ADDR_POS uint8 = 4
	CONTROL_REG_SPRITE_SIZE_POS            uint8 = 5
	CONTROL_REG_MASTER_SLAVE_SELECT_POS    uint8 = 6
	CONTROL_REG_GENERATE_NMI_POS           uint8 = 7

	// マスクレジスタ ($2001) のビット位置
	MASK_REG_GRAYSCALE                  uint8 = 0
	MASK_REG_LEFTMOST_BACKGROUND_ENABLE uint8 = 1
	MASK_REG_LEFTMOST_SPRITE_ENABLE     uint8 = 2
	MASK_REG_BACKGROUND_ENABLE          uint8 = 3
	MASK_REG_SPRITE_ENABLE              uint8 = 4
	MASK_REG_EMPHASIZE_RED_POS          uint8 = 5
	MASK_REG_EMPHASIZE_GREEN_POS        uint8 = 6
	MASK_REG_EMPHASIZE_BLUE_POS         uint8 = 7

	// ステータスレジスタ ($2002) のビット位置
	STATUS_REG_SPRITE_OVERFLOW uint8 = 5
	STATUS_REG_SPRITE_ZERO_HIT uint8 = 6
	STATUS_REG_VBLANK_FLAG     uint8 = 7
)

// MARK: コントロールレジスタ ($2000)
type ControlRegister struct {
	/*
		7654 3210
		---- ----
		VPHB SINN
		|||| ||||
		|||| ||++- ネームテーブルの基準アドレス
		|||| ||    (0 = $2000; 1 = $2400; 2 = $2800; 3 = $2C00)
		|||| |+--- VRAMアドレスの増分 (CPU の PPUDATA 読み書き毎)
		|||| |     (0: +1, VRAM上での横方向; 1: +32, VRAM上での縦方向)
		|||| +---- スプライトのパターンテーブルアドレス (8x8のスプライトのみ)
		||||       (0: $0000; 1: $1000; 8x16モードでは不使用)
		|||+------ 背景のパターンテーブルアドレス (0: $0000; 1: $1000)
		||+------- スプライトサイズ (0: 8x8 px; 1: 8x16 px)
		|+-------- PPU マスター/スレーブの選択
		|
		+--------- Vblank開始時に NMI を発生させるか否か (0: off, 1: on)
	*/

	nameTable1               bool
	nameTable2               bool
	vramAddressIncrement     bool
	spritePatternAddress     bool
	backgroundPatternAddress bool
	spriteSize               bool
	masterSlaveSelect        bool
	generateNMI              bool
}

// コントロールレジスタのコンストラクタ
func (cr *ControlRegister) Init() {
	cr.update(0b0000_0000)
}

// VRAMアドレスの増分を取得するメソッド
func (cr *ControlRegister) VRAMAddressIncrement() uint8 {
	if !cr.vramAddressIncrement {
		return 1
	} else {
		return 32
	}
}

// ネームテーブルの基準アドレスを取得するメソッド
func (cr *ControlRegister) BaseNameTableAddress() uint16 {
	if cr.nameTable1 && cr.nameTable2 {
		return 0x2C00
	} else if cr.nameTable2 {
		return 0x2800
	} else if cr.nameTable1 {
		return 0x2400
	} else {
		return 0x2000
	}
}

// スプライトのパターンテーブルの基準アドレスを取得するメソッド
func (cr *ControlRegister) SpritePatternTableAddress() uint16 {
	if !cr.spritePatternAddress {
		return 0x0000
	} else {
		return 0x1000
	}
}

// 背景のパターンテーブルの基準アドレスを取得するメソッド
func (cr *ControlRegister) BackgroundPatternTableAddress() uint16 {
	if !cr.backgroundPatternAddress {
		return 0x0000
	} else {
		return 0x1000
	}
}

// スプライトのサイズを取得するメソッド
func (cr *ControlRegister) SpriteSize() uint8 {
	if !cr.spriteSize {
		return 8
	} else {
		return 16
	}
}

// PPUのマスター/スレーブを取得するメソッド
func (cr *ControlRegister) MasterSlaveSelect() uint8 {
	if !cr.masterSlaveSelect {
		return 0
	} else {
		return 1
	}
}

// VBlankNMIの状態を取得するメソッド
func (cr *ControlRegister) GenerateNMI() bool {
	return cr.generateNMI
}

// コントロールレジスタをuint8へ変換するメソッド
func (cr *ControlRegister) ToByte() uint8 {
	var value uint8 = 0x00

	if cr.nameTable1 {
		value |= 1 << CONTROL_REG_NAMETABLE1_POS
	}
	if cr.nameTable2 {
		value |= 1 << CONTROL_REG_NAMETABLE2_POS
	}
	if cr.vramAddressIncrement {
		value |= 1 << CONTROL_REG_VRAM_ADD_INCREMENT_POS
	}
	if cr.spritePatternAddress {
		value |= 1 << CONTROL_REG_SPRITE_PATTERN_ADDR_POS
	}
	if cr.backgroundPatternAddress {
		value |= 1 << CONTROL_REG_BACKROUND_PATTERN_ADDR_POS
	}
	if cr.spriteSize {
		value |= 1 << CONTROL_REG_SPRITE_SIZE_POS
	}
	if cr.masterSlaveSelect {
		value |= 1 << CONTROL_REG_MASTER_SLAVE_SELECT_POS
	}
	if cr.generateNMI {
		value |= 1 << CONTROL_REG_GENERATE_NMI_POS
	}

	return value
}

// uint8の値をコントロールレジスタオブジェクトへ反映するメソッド
func (cr *ControlRegister) update(value uint8) {
	cr.nameTable1 = (value & (1 << CONTROL_REG_NAMETABLE1_POS)) != 0
	cr.nameTable2 = (value & (1 << CONTROL_REG_NAMETABLE2_POS)) != 0
	cr.vramAddressIncrement = (value & (1 << CONTROL_REG_VRAM_ADD_INCREMENT_POS)) != 0
	cr.spritePatternAddress = (value & (1 << CONTROL_REG_SPRITE_PATTERN_ADDR_POS)) != 0
	cr.backgroundPatternAddress = (value & (1 << CONTROL_REG_BACKROUND_PATTERN_ADDR_POS)) != 0
	cr.spriteSize = (value & (1 << CONTROL_REG_SPRITE_SIZE_POS)) != 0
	cr.masterSlaveSelect = (value & (1 << CONTROL_REG_MASTER_SLAVE_SELECT_POS)) != 0
	cr.generateNMI = (value & (1 << CONTROL_REG_GENERATE_NMI_POS)) != 0
}

// MARK: マスクレジスタ ($2001)
type MaskRegister struct {
	/*
	   7654 3210
	   ---- ----
	   BGRs bMmG
	   |||| ||||
	   |||| |||+- カラー/モノクロフラグ (0: カラー, 1: モノクロ)
	   |||| ||+-- 1: 画面左端8pxの背景を描画, 0: 非表示
	   |||| |+--- 1: 画面左端8pxのスプライトを描画, 0: 非表示
	   |||| +---- 1: 背景を描画
	   |||+------ 1: スプライトを描画
	   ||+------- 赤色を強調
	   |+-------- 緑色を強調
	   +--------- 青色を強調
	*/

	grayscale                bool
	leftmostBackgroundEnable bool
	leftmostSpriteEnable     bool
	backgroundEnable         bool
	spriteEnable             bool
	emphasizeRed             bool
	emphasizeGreen           bool
	emphasizeBlue            bool
}

func (mr *MaskRegister) Init() {
	mr.update(0b0000_0000)
}

// マスクレジスタをuint8へ変換するメソッド
func (mr *MaskRegister) ToByte() uint8 {
	var value uint8 = 0x00

	if mr.grayscale {
		value |= 1 << MASK_REG_GRAYSCALE
	}
	if mr.leftmostBackgroundEnable {
		value |= 1 << MASK_REG_LEFTMOST_BACKGROUND_ENABLE
	}
	if mr.leftmostSpriteEnable {
		value |= 1 << MASK_REG_LEFTMOST_SPRITE_ENABLE
	}
	if mr.backgroundEnable {
		value |= 1 << MASK_REG_BACKGROUND_ENABLE
	}
	if mr.spriteEnable {
		value |= 1 << MASK_REG_SPRITE_ENABLE
	}
	if mr.emphasizeRed {
		value |= 1 << MASK_REG_EMPHASIZE_RED_POS
	}
	if mr.emphasizeGreen {
		value |= 1 << MASK_REG_EMPHASIZE_GREEN_POS
	}
	if mr.emphasizeBlue {
		value |= 1 << MASK_REG_EMPHASIZE_BLUE_POS
	}

	return value
}

// uint8の値をマスクレジスタオブジェクトへ反映するメソッド
func (mr *MaskRegister) update(value uint8) {
	mr.grayscale = (value & (1 << MASK_REG_GRAYSCALE)) != 0
	mr.leftmostBackgroundEnable = (value & (1 << MASK_REG_LEFTMOST_BACKGROUND_ENABLE)) != 0
	mr.leftmostSpriteEnable = (value & (1 << MASK_REG_LEFTMOST_SPRITE_ENABLE)) != 0
	mr.backgroundEnable = (value & (1 << MASK_REG_BACKGROUND_ENABLE)) != 0
	mr.spriteEnable = (value & (1 << MASK_REG_SPRITE_ENABLE)) != 0
	mr.emphasizeRed = (value & (1 << MASK_REG_EMPHASIZE_RED_POS)) != 0
	mr.emphasizeGreen = (value & (1 << MASK_REG_EMPHASIZE_GREEN_POS)) != 0
	mr.emphasizeBlue = (value & (1 << MASK_REG_EMPHASIZE_BLUE_POS)) != 0
}

// MARK: ステータスレジスタ ($2002)
type StatusRegister struct {
	/*
		7  bit  0
		---- ----
		VSOx xxxx
		|||
		|||
		||+------- スプライトのオーバーフローフラグ (バグあり)
		|+-------- スプライト 0 ヒット
		+--------- Vblank フラグ, Statusレジスタを読まれるタイミングでクリアされる
	*/

	spriteOverflow bool
	spriteZeroHit  bool
	vBlankFlag     bool
}

// ステータスレジスタのコンストラクタ
func (sr *StatusRegister) Init() {
	sr.update(0b0001_0000)
}

// VBlankフラグの設定メソッド
func (sr *StatusRegister) SetVBlankStatus(status bool) {
	sr.vBlankFlag = status
}

// VBlankフラグの状態
func (sr *StatusRegister) ClearVBlankStatus() {
	sr.vBlankFlag = false
}

// VBlank期間中かどうかを返すメソッド
func (sr *StatusRegister) VBlank() bool {
	return sr.vBlankFlag
}

// スプライト0ヒットの設定メソッド
func (sr *StatusRegister) SetSpriteZeroHit(status bool) {
	sr.spriteZeroHit = status
}

// スプライトオーバーフローフラグの設定メソッド
func (sr *StatusRegister) SetSpriteOverflow(status bool) {
	sr.spriteOverflow = status
}

// ステータスレジスタをuint8へ変換するメソッド
func (sr *StatusRegister) ToByte() uint8 {
	var value uint8 = 0x00

	if sr.spriteOverflow {
		value |= 1 << STATUS_REG_SPRITE_OVERFLOW
	}
	if sr.spriteZeroHit {
		value |= 1 << STATUS_REG_SPRITE_ZERO_HIT
	}
	if sr.vBlankFlag {
		value |= 1 << STATUS_REG_VBLANK_FLAG
	}

	return uint8(value)
}

// uint8の値をステータスレジスタオブジェクトへ反映するメソッド
func (sr *StatusRegister) update(value uint8) {
	sr.spriteOverflow = (value & (1 << STATUS_REG_SPRITE_OVERFLOW)) != 0
	sr.spriteZeroHit = (value & (1 << STATUS_REG_SPRITE_ZERO_HIT)) != 0
	sr.vBlankFlag = (value & (1 << STATUS_REG_VBLANK_FLAG)) != 0
}

// MARK: T/Vレジスタ (PPU 内部)
type InternalAddressRegiseter struct {
	/*
		yyy NN YYYYY XXXXX
		||| || ||||| +++++-- タイルの画面内列番号 X (0-31)
		||| || +++++-------- タイルの画面内行番号 Y (0-29)
		||| ++-------------- nametable select
		+++----------------- タイル内の Y 座標 (0-7)
	*/
	fineY     uint8
	nameTable uint8
	coarseY   uint8
	coarseX   uint8
}

// T/Vレジスタの初期化メソッド
func (iar *InternalAddressRegiseter) Init() {
	iar.updateNameTable(0x00)
	iar.fineY = 0x00
	iar.coarseX = 0x00
	iar.coarseY = 0x00
}

// ネームテーブル位置の更新メソッド
func (iar *InternalAddressRegiseter) updateNameTable(value uint8) {
	/*
		t: ...GH.. ........ <- value: ......GH
	*/
	iar.nameTable = value & 0x03
}

// スクロール値の更新メソッド
func (iar *InternalAddressRegiseter) updateScroll(value uint8, w *InternalWRegister) {
	/*
		1回目の書き込み (w = 0) → X座標のセット
		t: ....... ...ABCDE <- value: ABCDEFGH
		w:                  <- 1

		2回目の書き込み (w = 1) → Y座標のセット
		t: FGH..AB CDE..... <- value: ABCDEFGH
		w:                  <- 0
	*/

	if !w.latch {
		// Xのスクロール値のセット
		iar.coarseX = (value & 0xF8) >> 3
	} else {
		// Yのスクロール値のセット
		iar.fineY = value & 0x07
		iar.coarseY = (value & 0xF8) >> 3 // coarseYの下位1ビットを維持
	}

	// Wレジスタの反転
	w.toggle()
}

func (iar *InternalAddressRegiseter) updateAddress(value uint8, w *InternalWRegister) {
	/*
		1回目の書き込み (w = 0) → 上位バイトのセット
		t: .CDEFGH ........ <- d: ..CDEFGH
					<unused>     <- d: AB......
		t: Z...... ........ <- 0 (bit Z is cleared)
		w:                  <- 1

		2回目の書き込み (w = 1) → 下位バイトのセット
		t: ....... ABCDEFGH <- d: ABCDEFGH
		v: <...all bits...> <- t: <...all bits...>
		w:                  <- 0
	*/

	if !w.latch {
		// 上位バイトの書き込み
		// t: .CDEFGH ........ <- value: ..CDEFGH
		// tのビット14はクリアされる
		iar.fineY = (value >> 4) & 0x07
		iar.nameTable = (value >> 2) & 0x03
		iar.coarseY = (iar.coarseY & 0x07) | ((value & 0x03) << 3)
	} else {
		// 下位バイトの書き込み
		// t: ....... ABCDEFGH <- value: ABCDEFGH
		iar.coarseY = (iar.coarseY & 0x18) | ((value >> 5) & 0x07)
		iar.coarseX = value & 0x1F
	}

	// Wレジスタの反転
	w.toggle()
}

// 水平方向のVRAMアドレスをインクリメントするメソッド
func (iar *InternalAddressRegiseter) incrementCoarseX() {
	// Coarse Xが31未満ならインクリメント
	if iar.coarseX < 31 {
		iar.coarseX++
	} else {
		// 31なら0に戻し、水平ネームテーブルを切り替える (ビット0を反転)
		iar.coarseX = 0
		iar.nameTable ^= 0b01
	}
}

// 垂直方向のVRAMアドレスをインクリメントするメソッド
func (iar *InternalAddressRegiseter) incrementY() {
	// Fine Yが7未満ならインクリメント
	if iar.fineY < 7 {
		iar.fineY++
	} else {
		// 7なら0に戻し、Coarse Yをインクリメント
		iar.fineY = 0
		y := iar.coarseY
		switch y {
		case 29:
			// 画面の最後のタイル行ならCoarse Yを0に戻し、垂直ネームテーブルを切り替える (ビット1を反転)
			y = 0
			iar.nameTable ^= 0b10
		case 31:
			// Coarse Yが31（属性テーブルなどの領域）に達した場合、0に戻す
			y = 0
		default:
			// それ以外はインクリメント
			y++
		}
		iar.coarseY = y
	}
}

// 全てのビットを別のレジスタにコピーするメソッド
func (iar *InternalAddressRegiseter) copyAllBitsTo(iar_to *InternalAddressRegiseter) {
	iar_to.fineY = iar.fineY
	iar_to.nameTable = iar.nameTable
	iar_to.coarseX = iar.coarseX
	iar_to.coarseY = iar.coarseY
}

// 水平方向のビットを別のレジスタにコピーするメソッド
func (iar *InternalAddressRegiseter) copyHorizontalBitsTo(iar_to *InternalAddressRegiseter) {
	// HBlank直前に使う
	iar_to.nameTable = (iar_to.nameTable & 0b10) | (iar.nameTable & 0b01)
	iar_to.coarseX = iar.coarseX
}

// 垂直方向のビットを別のレジスタにコピーするメソッド
func (iar *InternalAddressRegiseter) copyVerticalBitsTo(iar_to *InternalAddressRegiseter) {
	// VBlank直前に使う
	iar_to.fineY = iar.fineY
	iar_to.nameTable = (iar_to.nameTable & 0b01) | (iar.nameTable & 0b10)
	iar_to.coarseY = iar.coarseY
}

// V/Tレジスタをuint16へ変換するメソッド
func (iar *InternalAddressRegiseter) ToByte() uint16 {
	var value uint16 = 0x00
	value |= uint16(iar.fineY) << 12
	value |= uint16(iar.nameTable) << 10
	value |= uint16(iar.coarseY) << 5
	value |= uint16(iar.coarseX)

	return value
}

// uint16からV/Tレジスタオブジェクトに変換するメソッド
func (iar *InternalAddressRegiseter) SetFromWord(value uint16) {
	iar.fineY = uint8((value >> 12) & 0x07)
	iar.nameTable = uint8((value >> 10) & 0x03)
	iar.coarseY = uint8((value >> 5) & 0x1F)
	iar.coarseX = uint8(value & 0x1F)
}

// MARK: Xレジスタ (PPU 内部)
type InternalXRegister struct {
	fineX uint8
}

// Xレジスタの初期化メソッド
func (ixr *InternalXRegister) Init() {
	ixr.update(0x00)
}

// Xレジスタの更新メソッド
func (ixr *InternalXRegister) update(value uint8) {
	/*
		1回目の書き込み (w = 0) → X座標のセット
		x: ....... .....FGH <- value: ABCDEFGH
	*/
	ixr.fineX &= ^uint8(0x07) // 元の値をクリア
	ixr.fineX |= value & 0x07 // 下位3bitに書き込み
}

// MARK: Wレジスタ (PPU 内部)
type InternalWRegister struct {
	latch bool
}

// Wレジスタの初期化メソッド
func (iwr *InternalWRegister) Init() {
	iwr.reset()
}

// Wレジスタの反転メソッド
func (iwr *InternalWRegister) toggle() {
	iwr.latch = !iwr.latch
}

// Wレジスタの初期化メソッド
func (iwr *InternalWRegister) reset() {
	iwr.latch = false
}
