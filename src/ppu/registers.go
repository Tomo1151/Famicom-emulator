package ppu

// import "Famicom-emulator/bus"

// PPUレジスタ関連の定数
const (
	// PPUのI/Oレジスタのアドレス
	PPU_CTRL uint16 = 0x2000
	PPU_MASK uint16 = 0x2001
	PPU_STATUS uint16 = 0x2002
	OAM_ADDR uint16 = 0x2003
	OAM_DATA uint16 = 0x2004
	PPU_SCROLL uint16 = 0x2005
	PPU_ADDR uint16 = 0x2006
	PPU_DATA uint16 = 0x2007
	OAM_DMA uint16 = 0x4014

	PPU_REG_START  = 0x2000
	PPU_REG_END    = 0x3FFF

	// PPU_ADDRのアドレスマスク
	PPU_ADDR_MIRROR_MASK = 0b111111_11111111
	PPU_VRAM_MIRROR_MASK = 0b101111_11111111
)

const (
	// コントロールレジスタ ($2000) のビット位置
	CONTROL_REG_NAMETABLE1_POS uint8 = 0
	CONTROL_REG_NAMETABLE2_POS uint8 = 1
	CONTROL_REG_VRAM_ADD_INCREMENT_POS uint8 = 2
	CONTROL_REG_SPRITE_PATTERN_ADDR_POS uint8 = 3
	CONTROL_REG_BACKROUND_PATTERN_ADDR_POS uint8 = 4
	CONTROL_REG_SPRITE_SIZE_POS uint8 = 5
	CONTROL_REG_MASTER_SLAVE_SELECT_POS uint8 = 6
	CONTROL_REG_GENERATE_NMI_POS uint8 = 7

	// マスクレジスタ ($2001) のビット位置
	MASK_REG_GRAYSCALE uint8 = 0
	MASK_REG_LEFTMOST_BACKGROUND_ENABLE uint8 = 1
	MASK_REG_LEFTMOST_SPRITE_ENABLE uint8 = 2
	MASK_REG_BACKGROUND_ENABLE uint8 = 3
	MASK_REG_SPRITE_ENABLE uint8 = 4
	MASK_REG_EMPHASIZE_RED_POS uint8 = 5
	MASK_REG_EMPHASIZE_GREEN_POS uint8 = 6
	MASK_REG_EMPHASIZE_BLUE_POS uint8 = 7

	// ステータスレジスタ ($2002) のビット位置
	STATUS_REG_SPRITE_OVERFLOW uint8 = 5
	STATUS_REG_SPRITE_ZERO_HIT uint8 = 6
	STATUS_REG_VBLANK_FLAG uint8 = 7
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

	NameTable1 bool
	NameTable2 bool
	VRAMAddIncrement bool
	SpritePatternAddress bool
	BackgroundPatternAddress bool
	SpriteSize bool
	MasterSlaveSelect bool
	GenerateNMI bool
}

// コントロールレジスタのコンストラクタ
func (cr *ControlRegister) Init() {
	cr.update(0b0000_0000)
}

// VRAMアドレスの増分を取得するメソッド
func (cr *ControlRegister) GetVRAMAddrIncrement() uint8 {
	if !cr.VRAMAddIncrement {
		return 1
	} else {
		return 32
	}
}

// ネームテーブルの基準アドレスを取得するメソッド
func (cr *ControlRegister) GetBaseNameTableAddress() uint16 {
	if cr.NameTable1 && cr.NameTable2 {
		return 0x2C00
	} else if cr.NameTable2 {
		return 0x2800
	} else if cr.NameTable1 {
		return 0x2400
	} else {
		return 0x2000
	}
}

// スプライトのパターンテーブルの基準アドレスを取得するメソッド
func (cr *ControlRegister) GetSpritePatternTableAddress() uint16 {
	if !cr.SpritePatternAddress {
		return 0x0000
	} else {
		return 0x1000
	}
}

// 背景のパターンテーブルの基準アドレスを取得するメソッド
func (cr *ControlRegister) GetBackgroundPatternTableAddress() uint16 {
	if !cr.BackgroundPatternAddress {
		return 0x0000
	} else {
		return 0x1000
	}
}

// スプライトのサイズを取得するメソッド
func (cr *ControlRegister) GetSpriteSize() uint8 {
	if !cr.SpriteSize {
		return 8
	} else {
		return 16
	}
}

// PPUのマスター/スレーブを取得するメソッド
func (cr *ControlRegister) GetPPUMasterSlaveSelect() uint8 {
	if !cr.MasterSlaveSelect {
		return 0
	} else {
		return 1
	}
}

// VBlankNMIの状態を取得するメソッド
func (cr *ControlRegister) GenerateVBlankNMI() bool {
	return cr.GenerateNMI
}

// コントロールレジスタをuint8へ変換するメソッド
func (cr *ControlRegister) ToByte() uint8 {
	var value uint8 = 0x00

	if cr.NameTable1 {
		value |= 1 << CONTROL_REG_NAMETABLE1_POS
	}
	if cr.NameTable2 {
		value |= 1 << CONTROL_REG_NAMETABLE2_POS
	}
	if cr.VRAMAddIncrement {
		value |= 1 << CONTROL_REG_VRAM_ADD_INCREMENT_POS
	}
	if cr.SpritePatternAddress {
		value |= 1 << CONTROL_REG_SPRITE_PATTERN_ADDR_POS
	}
	if cr.BackgroundPatternAddress {
		value |= 1 << CONTROL_REG_BACKROUND_PATTERN_ADDR_POS
	}
	if cr.SpriteSize {
		value |= 1 << CONTROL_REG_SPRITE_SIZE_POS
	}
	if cr.MasterSlaveSelect {
		value |= 1 << CONTROL_REG_MASTER_SLAVE_SELECT_POS
	}
	if cr.GenerateNMI {
		value |= 1 << CONTROL_REG_GENERATE_NMI_POS
	}

	return value
}

// uint8の値をコントロールレジスタオブジェクトへ反映するメソッド
func (cr *ControlRegister) update(value uint8) {
	cr.NameTable1 = (value & (1 << CONTROL_REG_NAMETABLE1_POS)) != 0
	cr.NameTable2 = (value & (1 << CONTROL_REG_NAMETABLE2_POS)) != 0
	cr.VRAMAddIncrement = (value & (1 << CONTROL_REG_VRAM_ADD_INCREMENT_POS)) != 0
	cr.SpritePatternAddress = (value & (1 << CONTROL_REG_SPRITE_PATTERN_ADDR_POS)) != 0
	cr.BackgroundPatternAddress = (value & (1 << CONTROL_REG_BACKROUND_PATTERN_ADDR_POS)) != 0
	cr.SpriteSize = (value & (1 << CONTROL_REG_SPRITE_SIZE_POS)) != 0
	cr.MasterSlaveSelect = (value & (1 << CONTROL_REG_MASTER_SLAVE_SELECT_POS)) != 0
	cr.GenerateNMI = (value & (1 << CONTROL_REG_GENERATE_NMI_POS)) != 0
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

	Grayscale bool
	LeftmostBackgroundEnable bool
	LeftmostSpriteEnable bool
	BackgroundEnable bool
	SpriteEnable bool
	EmphasizeRed bool
	EmphasizeGreen bool
	EmphasizeBlue bool
}

func (mr *MaskRegister) Init() {
	mr.update(0b0000_0000)
}

// マスクレジスタをuint8へ変換するメソッド
func (mr *MaskRegister) ToByte() uint8 {
	var value uint8 = 0x00

	if mr.Grayscale {
		value |= 1 << MASK_REG_GRAYSCALE
	}
	if mr.LeftmostBackgroundEnable {
		value |= 1 << MASK_REG_LEFTMOST_BACKGROUND_ENABLE
	}
	if mr.LeftmostSpriteEnable {
		value |= 1 << MASK_REG_LEFTMOST_SPRITE_ENABLE
	}
	if mr.BackgroundEnable {
		value |= 1 << MASK_REG_BACKGROUND_ENABLE
	}
	if mr.SpriteEnable {
		value |= 1 << MASK_REG_SPRITE_ENABLE
	}
	if mr.EmphasizeRed {
		value |= 1 << MASK_REG_EMPHASIZE_RED_POS
	}
	if mr.EmphasizeGreen {
		value |= 1 << MASK_REG_EMPHASIZE_GREEN_POS
	}
	if mr.EmphasizeBlue {
		value |= 1 << MASK_REG_EMPHASIZE_BLUE_POS
	}

	return value
}

// uint8の値をマスクレジスタオブジェクトへ反映するメソッド
func (mr *MaskRegister) update(value uint8) {
	mr.Grayscale = (value & (1 << MASK_REG_GRAYSCALE)) != 0
	mr.LeftmostBackgroundEnable = (value & (1 << MASK_REG_LEFTMOST_BACKGROUND_ENABLE)) != 0
	mr.LeftmostSpriteEnable = (value & (1 << MASK_REG_LEFTMOST_SPRITE_ENABLE)) != 0
	mr.BackgroundEnable = (value & (1 << MASK_REG_BACKGROUND_ENABLE)) != 0
	mr.SpriteEnable = (value & (1 << MASK_REG_SPRITE_ENABLE)) != 0
	mr.EmphasizeRed = (value & (1 << MASK_REG_EMPHASIZE_RED_POS)) != 0
	mr.EmphasizeGreen = (value & (1 << MASK_REG_EMPHASIZE_GREEN_POS)) != 0
	mr.EmphasizeBlue = (value & (1 << MASK_REG_EMPHASIZE_BLUE_POS)) != 0
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

	SpriteOverflow bool
	SpriteZeroHit bool
	VBlankFlag bool
}

// ステータスレジスタのコンストラクタ
func (sr *StatusRegister) Init() {
	sr.update(0b0001_0000)
}

// VBlankフラグの設定メソッド
func (sr *StatusRegister) SetVBlankStatus(status bool) {
	sr.VBlankFlag = status
}

// VBlankフラグの状態
func (sr *StatusRegister) ClearVBlankStatus() {
	sr.VBlankFlag = false
}

// VBlank期間中かどうかを返すメソッド
func (sr *StatusRegister) IsInVBlank() bool {
	return sr.VBlankFlag
}

// スプライト0ヒットの設定メソッド
func (sr *StatusRegister) SetSpriteZeroHit(status bool) {
	sr.SpriteZeroHit = status
}

// スプライトオーバーフローフラグの設定メソッド
func (sr *StatusRegister) SetSpriteOverflow(status bool) {
	sr.SpriteOverflow = status
}

// ステータスレジスタをuint8へ変換するメソッド
func (sr *StatusRegister) ToByte() uint8 {
	var value uint8 = 0x00

	if sr.SpriteOverflow {
		value |= 1 << STATUS_REG_SPRITE_OVERFLOW
	}
	if sr.SpriteZeroHit {
		value |= 1 << STATUS_REG_SPRITE_ZERO_HIT
	}
	if sr.VBlankFlag {
		value |= 1 << STATUS_REG_VBLANK_FLAG
	}

	return uint8(value)
}

// uint8の値をステータスレジスタオブジェクトへ反映するメソッド
func (sr *StatusRegister) update(value uint8) {
	sr.SpriteOverflow = (value & (1 << STATUS_REG_SPRITE_OVERFLOW)) != 0
	sr.SpriteZeroHit  = (value & (1 << STATUS_REG_SPRITE_ZERO_HIT)) != 0
	sr.VBlankFlag     = (value & (1 << STATUS_REG_VBLANK_FLAG)) != 0
}


// MARK: スクロールレジスタ ($2005)
// Deprecated: PPUに t/v/x/w 内部レジスタを持たせるように修正したため
type ScrollRegister struct {
	/*
		書き込み1回目
		7  bit  0
		---- ----
		XXXX XXXX
		|||| ||||
		++++-++++- X scroll bits 7-0 (bit 8 in PPUCTRL bit 0)

		書き込み2回目
		7  bit  0
		---- ----
		YYYY YYYY
		|||| ||||
		++++-++++- Y scroll bits 7-0 (bit 8 in PPUCTRL bit 1)
	*/

	ScrollX uint8 // スクロール値(X)
	ScrollY uint8 // スクロール値(Y)
	writeLatch bool // 現在のビットが上位ビットかどうか
}

// スクロールレジスタのコンストラクタ
func (sr *ScrollRegister) Init() {
	sr.ScrollX = 0x00
	sr.ScrollY = 0x00
	sr.writeLatch = false
}

// スクロールレジスタの書き込みメソッド (1度目はX, 2度目はYの値として書き込む)
func (sr *ScrollRegister) Write(data uint8) {
	if !sr.writeLatch {
		sr.ScrollX = data
	} else {
		sr.ScrollY = data
	}
	sr.writeLatch = !sr.writeLatch
}

// 書き込みラッチのリセットメソッド
func (sr *ScrollRegister) ResetLatch() {
	sr.writeLatch = false
}


// MARK: アドレスレジスタ ($2006)
// Deprecated: PPUに t/v/x/w 内部レジスタを持たせるように修正したため
type AddrRegister struct {
	upper uint8 // 上位ビット
	lower uint8 // 下位ビット
	writeLatch bool // 現在のビットが上位ビットかどうか
}

// アドレスレジスタのコンストラクタ
func (ar *AddrRegister) Init() {
	ar.upper = 0x00
	ar.lower = 0x00
	ar.writeLatch = true // Wレジスタ (書き込みが最初か2回目かを記憶する)
}

// アドレスレジスタにデータ(16bit)をセットするメソッド
func (ar *AddrRegister) set(data uint16) {
	ar.upper = uint8(data >> 8)
	ar.lower = uint8(data & 0xFF)
}

// アドレスレジスタのデータ(16bit)を取得するメソッド
func (ar *AddrRegister) get() uint16 {
	return uint16(ar.upper) << 8 | uint16(ar.lower)
}

// １バイトずつ書き込むメソッド
func (ar *AddrRegister) update(data uint8) {
	// 1回目の書き込みは上位ビット, 2回目は下位ビット
	if ar.writeLatch {
		ar.upper = data
	} else {
		ar.lower = data
	}

	// アドレスのミラーリング
	if ar.get() > PPU_REG_END {
		ar.set(ar.get() & PPU_ADDR_MIRROR_MASK)
	}

	// ビット位置を変更
	ar.writeLatch = !ar.writeLatch
}

// アドレスをインクリメントするメソッド
func (ar *AddrRegister) increment(step uint8) {
	// 方向によって 1 or 32 増やす
	current := ar.get()
	result := current + uint16(step)

	// アドレスのミラーリング
	if result > PPU_REG_END {
		ar.set(result & PPU_ADDR_MIRROR_MASK)
	} else {
		ar.set(result)
	}
}

// 書き込みラッチのリセットメソッド
func (ar *AddrRegister) ResetLatch() {
	ar.writeLatch = true
}
