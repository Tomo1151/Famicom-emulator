package ppu

import (
	"math/bits"
)

// MARK: 定数定義
const (
	CONTROL_REG_NAMETABLE1_POS uint8 = iota
	CONTROL_REG_NAMETABLE2_POS
	CONTROL_REG_VRAM_ADDR_INC_POS
	CONTROL_REG_SP_PATTERN_ADDR_POS
	CONTROL_REG_BG_PATTERN_ADDR_POS
	CONTROL_REG_SP_SIZE_POS
	CONTROL_REG_MASTER_SLAVE_POS
	CONTROL_REG_GENERATE_NMI_POS
)

const (
	MASK_REG_GRAYSCALE uint8 = iota
	MASK_REG_LEFTMOST_BG_ENABLE_POS
	MASK_REG_LEFTMOST_SP_ENABLE_POS
	MASK_REG_BG_ENABLE_POS
	MASK_REG_SP_ENABLE_POS
	MASK_REG_EMPHASIZE_RED_POS
	MASK_REG_EMPHASIZE_GREEN_POS
	MASK_REG_EMPHASIZE_BLUE_POS
)

const (
	STATUS_REG_SPRITE_OVERFLOW uint8 = 5
	STATUS_REG_SPRITE_ZERO_HIT uint8 = 6
	STATUS_REG_VBLANK_FLAG     uint8 = 7
)

// MARK: コントロールレジスタ ($2000)
type ControlRegister struct {
	/*
		PPU コントロールレジスタ

		7 6 5 4 3 2 1 0 ビット
		------- -------

		V P H B S I N N
		| | | | | | L |
		| | | | | |   +- N: ネームテーブルの基準アドレス
		| | | | | |         (0 = $2000; 1 = $2400; 2 = $2800; 3 = $2C00)
		| | | | | +----- I: VRAMアドレスの増分 (CPU の PPUDATA 読み書き毎)
		| | | | |           (0: +1, VRAM上での横方向; 1: +32, VRAM上での縦方向)
		| | | | +------- S: スプライトのパターンテーブルアドレス (8x8のスプライトのみ)
		| | | |             (0: $0000; 1: $1000; 8x16モードでは不使用)
		| | | +--------- B: 背景のパターンテーブルアドレス (0: $0000; 1: $1000)
		| | +----------- H: スプライトサイズ (0: 8x8 px; 1: 8x16 px)
		| +------------- P: PPU マスター/スレーブの選択
		+--------------- V: Vblank開始時に NMI を発生させるか否か (0: off, 1: on)
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
func NewControlRegister() ControlRegister {
	return ControlRegister{
		nameTable1:               false,
		nameTable2:               false,
		vramAddressIncrement:     false,
		spritePatternAddress:     false,
		backgroundPatternAddress: false,
		spriteSize:               false,
		masterSlaveSelect:        false,
		generateNMI:              false,
	}
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
		value |= 1 << CONTROL_REG_VRAM_ADDR_INC_POS
	}
	if cr.spritePatternAddress {
		value |= 1 << CONTROL_REG_SP_PATTERN_ADDR_POS
	}
	if cr.backgroundPatternAddress {
		value |= 1 << CONTROL_REG_BG_PATTERN_ADDR_POS
	}
	if cr.spriteSize {
		value |= 1 << CONTROL_REG_SP_SIZE_POS
	}
	if cr.masterSlaveSelect {
		value |= 1 << CONTROL_REG_MASTER_SLAVE_POS
	}
	if cr.generateNMI {
		value |= 1 << CONTROL_REG_GENERATE_NMI_POS
	}

	return value
}

// uint8の値をコントロールレジスタオブジェクトへ反映するメソッド
func (cr *ControlRegister) SetFromByte(value uint8) {
	cr.nameTable1 = (value & (1 << CONTROL_REG_NAMETABLE1_POS)) != 0
	cr.nameTable2 = (value & (1 << CONTROL_REG_NAMETABLE2_POS)) != 0
	cr.vramAddressIncrement = (value & (1 << CONTROL_REG_VRAM_ADDR_INC_POS)) != 0
	cr.spritePatternAddress = (value & (1 << CONTROL_REG_SP_PATTERN_ADDR_POS)) != 0
	cr.backgroundPatternAddress = (value & (1 << CONTROL_REG_BG_PATTERN_ADDR_POS)) != 0
	cr.spriteSize = (value & (1 << CONTROL_REG_SP_SIZE_POS)) != 0
	cr.masterSlaveSelect = (value & (1 << CONTROL_REG_MASTER_SLAVE_POS)) != 0
	cr.generateNMI = (value & (1 << CONTROL_REG_GENERATE_NMI_POS)) != 0
}

// MARK: マスクレジスタ ($2001)
type MaskRegister struct {
	/*
		PPU マスクレジスタ
		7 6 5 4 3 2 1 0
		------- -------
		B G R s b M m G
		| | | | | | | |
		| | | | | | | +- G: カラー/モノクロフラグ (0: カラー, 1: モノクロ)
		| | | | | | +--- m: 画面左端8pxの背景描画 (0: 非表示, 1: 表示)
		| | | | | +----- M: 画面左端8pxのスプライト描画 (0: 非表示, 1: 表示)
		| | | | +------- b: 背景の描画 (0: 非表示, 1: 表示)
		| | | +--------- s: スプライトの描画 (0: 非表示, 1: 表示)
		| | +----------- R: 赤色を強調 (0: 強調しない, 1: 強調する)
		| +------------- G: 緑色を強調 (0: 強調しない, 1: 強調する)
		+--------------- B: 青色を強調 (0: 強調しない, 1: 強調する)
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

// MARK: マスクレジスタのコンストラクタ
func NewMaskRegister() MaskRegister {
	return MaskRegister{
		grayscale:                false,
		leftmostBackgroundEnable: false,
		leftmostSpriteEnable:     false,
		backgroundEnable:         false,
		spriteEnable:             false,
		emphasizeRed:             false,
		emphasizeGreen:           false,
		emphasizeBlue:            false,
	}
}

// マスクレジスタをuint8へ変換するメソッド
func (mr *MaskRegister) ToByte() uint8 {
	var value uint8 = 0x00

	if mr.grayscale {
		value |= 1 << MASK_REG_GRAYSCALE
	}
	if mr.leftmostBackgroundEnable {
		value |= 1 << MASK_REG_LEFTMOST_BG_ENABLE_POS
	}
	if mr.leftmostSpriteEnable {
		value |= 1 << MASK_REG_LEFTMOST_SP_ENABLE_POS
	}
	if mr.backgroundEnable {
		value |= 1 << MASK_REG_BG_ENABLE_POS
	}
	if mr.spriteEnable {
		value |= 1 << MASK_REG_SP_ENABLE_POS
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
func (mr *MaskRegister) SetFromByte(value uint8) {
	mr.grayscale = (value & (1 << MASK_REG_GRAYSCALE)) != 0
	mr.leftmostBackgroundEnable = (value & (1 << MASK_REG_LEFTMOST_BG_ENABLE_POS)) != 0
	mr.leftmostSpriteEnable = (value & (1 << MASK_REG_LEFTMOST_SP_ENABLE_POS)) != 0
	mr.backgroundEnable = (value & (1 << MASK_REG_BG_ENABLE_POS)) != 0
	mr.spriteEnable = (value & (1 << MASK_REG_SP_ENABLE_POS)) != 0
	mr.emphasizeRed = (value & (1 << MASK_REG_EMPHASIZE_RED_POS)) != 0
	mr.emphasizeGreen = (value & (1 << MASK_REG_EMPHASIZE_GREEN_POS)) != 0
	mr.emphasizeBlue = (value & (1 << MASK_REG_EMPHASIZE_BLUE_POS)) != 0
}

// MARK: ステータスレジスタ ($2002)
type StatusRegister struct {
	/*
		PPU ステータスレジスタ

		7 6 5 4 3 2 1 0
		------- -------
		V S O - - - - -
		| | |
		| | |
		| | +----------- O: スプライトのオーバーフローフラグ
		| +------------- S: スプライト 0 ヒットフラグ
		+--------------- V: VBlank フラグ
	*/

	spriteOverflow bool
	spriteZeroHit  bool
	vBlankFlag     bool
}

// MARK: ステータスレジスタのコンストラクタ
func NewStatusRegister() StatusRegister {
	return StatusRegister{
		spriteOverflow: false,
		spriteZeroHit:  false,
		vBlankFlag:     false,
	}
}

// スプライトオーバーフローの状態を取得するメソッド
func (sr *StatusRegister) SpriteOverflow() bool {
	return sr.spriteOverflow
}

// スプライト0ヒットフラグの状態を取得するメソッド
func (sr *StatusRegister) SpriteZeroHit() bool {
	return sr.spriteZeroHit
}

// VBlankフラグの状態を取得するメソッド
func (sr *StatusRegister) VBlank() bool {
	return sr.vBlankFlag
}

// スプライトオーバーフローフラグの設定メソッド
func (sr *StatusRegister) SetSpriteOverflow(status bool) {
	sr.spriteOverflow = status
}

// スプライト0ヒットの設定メソッド
func (sr *StatusRegister) SetSpriteZeroHit(status bool) {
	sr.spriteZeroHit = status
}

// VBlankフラグの設定メソッド
func (sr *StatusRegister) SetVBlank(status bool) {
	sr.vBlankFlag = status
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
func (sr *StatusRegister) SetFromByte(value uint8) {
	sr.spriteOverflow = (value & (1 << STATUS_REG_SPRITE_OVERFLOW)) != 0
	sr.spriteZeroHit = (value & (1 << STATUS_REG_SPRITE_ZERO_HIT)) != 0
	sr.vBlankFlag = (value & (1 << STATUS_REG_VBLANK_FLAG)) != 0
}

// MARK: アドレスレジスタ [V/Tレジスタ] (PPU内部)
type AddressRegiseter struct {
	/*
		PPU アドレスレジスタ

		14            8  7              0
		---------------  ----------------
		y y y  N N  Y Y  Y Y Y  X X X X X
		L + |  L |  L +  + + |  L + + + |
		    |    |           |          +- X: タイルの画面内列番号 X (0-31)
		    |    |           +------------ Y: タイルの画面内行番号 Y (0-29)
		    |    +------------------------ N: ネームテーブル選択
		    +----------------------------- y: タイル内のY座標 (0-7)
	*/

	coarseX   uint8
	coarseY   uint8
	nameTable uint8
	fineY     uint8
}

// MARK: アドレスレジスタのコンストラクタ
func NewAddressRegister() AddressRegiseter {
	return AddressRegiseter{
		coarseX:   0x00,
		coarseY:   0x00,
		nameTable: 0x00,
		fineY:     0x00,
	}
}

// ネームテーブル選択の更新メソッド
func (ar *AddressRegiseter) updateNameTable(value uint8) {
	/*
		T: ...GH.. ........ ← value: ......GH
	*/

	ar.nameTable = value & 0x03
}

// スクロール値の更新メソッド
func (ar *AddressRegiseter) updateScroll(value uint8, latch bool) {
	/*
		1回目の書き込み (w = 0): Xスクロールのセット
		T: ........ ...ABCDE ← value: ABCDEFGH

		2回目の書き込み (w = 1): Yスクロールのセット
		T: .FGH..AB CDE..... ← value: ABCDEFGH
	*/

	if !latch {
		// Xスクロールのセット
		ar.coarseX = value >> 3
	} else {
		// Yスクロールのセット
		ar.fineY = value & 0x07
		ar.coarseY = value >> 3
	}
}

// VRAMアドレスの更新メソッド ($2006)
func (ar *AddressRegiseter) updateAddress(value uint8, latch bool) {
	/*
		1回目の書き込み (w = 0): 上位バイトのセット
			.yyyNNYY YYYXXXXX
		    -------- --------
		T:  ..CDEFGH ........ ← value: ABCDEFGH

		2回目の書き込み (w = 1): 下位バイトのセット
			.yyyNNYY YYYXXXXX
			-------- --------
		T:  ........ ABCDEFGH ← value: ABCDEFGH
	*/

	if !latch {
		// 上位バイトの書き込み
		ar.fineY = (value >> 4) & 0x03                     // ABCD_EFGH → 00CD (14ビット目は常にクリアされる)
		ar.nameTable = (value >> 2) & 0x03                 // ABCD_EFGH → 00EF
		ar.coarseY = (value&0x03)<<3 | (ar.coarseY & 0x07) // Yの上位2バイトのみを更新
	} else {
		// 下位バイトの書き込み
		ar.coarseY = (ar.coarseY & 0x18) | ((value >> 5) & 0x07)
		ar.coarseX = value & 0x1F
	}
}

// 水平方向のVRAMアドレスをインクリメントするメソッド
func (ar *AddressRegiseter) incrementHorizontal() {
	if ar.coarseX < 31 {
		ar.coarseX++
	} else {
		// 31を超えたら0に戻し、水平ネームテーブルを切り替える (ビット0を反転)
		ar.coarseX = 0
		ar.nameTable ^= 0b01
	}
}

// 垂直方向のVRAMアドレスをインクリメントするメソッド
func (ar *AddressRegiseter) incrementVertical() {
	if ar.fineY < 7 {
		ar.fineY++
	} else {
		// 7を超えたら0に戻し、Coarse Yをインクリメント
		ar.fineY = 0
		ar.coarseY++

		switch ar.coarseY {
		case 32: // メモリの端
			ar.coarseY = 0
		case 30: // 画面の端
			ar.coarseY = 0
			ar.nameTable ^= 0b10
		}
	}
}

// アドレスレジスタ間で全ての値をコピーするメソッド
func (ar *AddressRegiseter) copyAllBitsTo(target *AddressRegiseter) {
	target.fineY = ar.fineY
	target.nameTable = ar.nameTable
	target.coarseX = ar.coarseX
	target.coarseY = ar.coarseY
}

// アドレスレジスタ間でX座標の値をコピーするメソッド
func (ar *AddressRegiseter) copyHorizontalBitsTo(target *AddressRegiseter) {
	target.nameTable = (target.nameTable & 0b10) | (ar.nameTable & 0b01)
	target.coarseX = ar.coarseX
}

// アドレスレジスタ間でY座標の値をコピーするメソッド
func (ar *AddressRegiseter) copyVerticalBitsTo(target *AddressRegiseter) {
	target.fineY = ar.fineY
	target.nameTable = (target.nameTable & 0b01) | (ar.nameTable & 0b10)
	target.coarseY = ar.coarseY
}

// アドレスレジスタをuint16へ変換するメソッド
func (ar *AddressRegiseter) ToWord() uint16 {
	var value uint16 = 0x00
	value |= uint16(ar.fineY) << 12
	value |= uint16(ar.nameTable) << 10
	value |= uint16(ar.coarseY) << 5
	value |= uint16(ar.coarseX)

	return value
}

// uint16からアドレスレジスタオブジェクトに変換するメソッド
func (ar *AddressRegiseter) SetFromWord(value uint16) {
	ar.fineY = uint8((value >> 12) & 0x07)
	ar.nameTable = uint8((value >> 10) & 0x03)
	ar.coarseY = uint8((value >> 5) & 0x1F)
	ar.coarseX = uint8(value & 0x1F)
}

// MARK: Xレジスタ (PPU内部)
type XRegister struct {
	/*
		PPU Xレジスタ

		7 6 5 4 3 2 1 0
		------- -------
		- - - - - X X X
		          L + |
		              +- X: スクロールX座標
	*/

	fineX uint8
}

// MARK: Xレジスタのコンストラクタ
func NewXRegister() XRegister {
	return XRegister{
		fineX: 0x00,
	}
}

// Xレジスタの更新メソッド
func (xr *XRegister) update(value uint8) {
	/*
		1回目の書き込み (w = 0): X座標の下位3ビットをセット
		X: ....... .....FGH ← value: ABCDEFGH

		2回目の書き込み (w = 1): 何もしない
	*/

	xr.fineX = value & 0x07 // 下位3bitに書き込み
}

// MARK: Wレジスタ (PPU内部)
type WRegister struct {
	/*
		PPU Wレジスタ

		7 6 5 4 3 2 1 0
		------- -------
		- - - - - - - W
		              |
		              +- W: 書き込みラッチ
	*/

	latch bool
}

// MARK: Wレジスタのコンストラクタ
func NewWRegister() WRegister {
	return WRegister{
		latch: false,
	}
}

// Wレジスタの反転メソッド
func (wr *WRegister) toggle() {
	wr.latch = !wr.latch
}

// Wレジスタの初期化メソッド
func (wr *WRegister) reset() {
	wr.latch = false
}

// MARK: 背景用ラッチ
type BackgroundLatch struct {
	nameTable    uint8 // 描画する背景タイルの番号
	attribute    uint8 // 背景の属性情報
	patternLower uint8 // 背景タイルのビットプレーン 0
	patternUpper uint8 // 背景タイルのビットプレーン 1
}

func NewBackgroundLatch() BackgroundLatch {
	return BackgroundLatch{
		nameTable:    0x00,
		attribute:    0x00,
		patternLower: 0x00,
		patternUpper: 0x00,
	}
}

// MARK: スプライト用ラッチ
type SpriteLatch struct {
	patternLower uint8 // スプライトタイルのビットプレーン 0
	patternUpper uint8 // スプライトタイルのビットプレーン 1
}

func NewSpriteLatch() SpriteLatch {
	return SpriteLatch{
		patternLower: 0x00,
		patternUpper: 0x00,
	}
}

// MARK: 背景シフトレジスタの定義
type BackgroundShiftRegister struct {
	attributeLower uint16 // 属性情報の 0 ビット目
	attributeUpper uint16 // 属性情報の 1 ビット目
	patternLower   uint16 // 背景タイルのビットプレーン 0
	patternUpper   uint16 // 背景タイルのビットプレーン 1
}

func NewBackgroundShiftRegister() BackgroundShiftRegister {
	return BackgroundShiftRegister{
		attributeLower: 0x0000,
		attributeUpper: 0x0000,
		patternLower:   0x0000,
		patternUpper:   0x0000,
	}
}

func (sr *BackgroundShiftRegister) shift() {
	sr.attributeLower <<= 1
	sr.attributeUpper <<= 1
	sr.patternLower <<= 1
	sr.patternUpper <<= 1
}

func (sr *BackgroundShiftRegister) load(latch *BackgroundLatch) {
	// 新しい値を書き込むために下位バイトをクリア
	sr.patternLower &= 0xFF00
	sr.patternUpper &= 0xFF00
	sr.attributeLower &= 0xFF00
	sr.attributeUpper &= 0xFF00

	// 新しい値を下位バイトに書き込み
	// ビットプレーン
	sr.patternLower |= uint16(latch.patternLower)
	sr.patternUpper |= uint16(latch.patternUpper)

	// 属性(パレット)情報
	// 次のフェッチタイミングまでの分(8ピクセル分)書き込む
	if latch.attribute&0b01 != 0 {
		sr.attributeLower |= 0xFF
	}
	if latch.attribute&0b10 != 0 {
		sr.attributeUpper |= 0xFF
	}
}

// MARK: スプライトシフトレジスタの定義
type SpriteShiftRegister struct {
	attributes   uint8 // スプライト属性
	patternLower uint8 // スプライトタイルのビットプレーン 0
	patternUpper uint8 // スプライトタイルのビットプレーン 1
	xDistance    uint8 // スクロール値
	isSpriteZero bool  // このシフタがスプライト0のものかどうか
}

func NewSpriteShiftRegister() SpriteShiftRegister {
	return SpriteShiftRegister{
		attributes:   0x00,
		patternLower: 0x00,
		patternUpper: 0x00,
		xDistance:    0x00,
		isSpriteZero: false,
	}
}

func (sr *SpriteShiftRegister) shift() {
	sr.patternLower <<= 1
	sr.patternUpper <<= 1
}

func (sr *SpriteShiftRegister) load(latch *SpriteLatch, flipH bool) {
	patternLower := latch.patternLower
	patternUpper := latch.patternUpper
	if flipH {
		// 水平反転
		patternLower = bits.Reverse8(patternLower)
		patternUpper = bits.Reverse8(patternUpper)
	}
	sr.patternLower = patternLower
	sr.patternUpper = patternUpper
}

func (sr *SpriteShiftRegister) reset() {
	sr.attributes = 0x00
	sr.patternLower = 0x00
	sr.patternUpper = 0x00
	sr.xDistance = 0x00
	sr.isSpriteZero = false
}
