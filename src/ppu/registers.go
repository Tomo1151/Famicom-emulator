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

	// コントロールレジスタのビット位置
	CONTROL_REG_NAMETABLE1_POS uint8 = 0;
	CONTROL_REG_NAMETABLE2_POS uint8 = 1;
	CONTROL_REG_VRAM_ADD_INCREMENT_POS uint8 = 2;
	CONTROL_REG_SPRITE_PATTERN_ADDR_POS uint8 = 3;
	CONTROL_REG_BACKROUND_PATTERN_ADDR_POS uint8 = 4;
	CONTROL_REG_SPRITE_SIZE_POS uint8 = 5;
	CONTROL_REG_MASTER_SLAVE_SELECT_POS uint8 = 6;
	CONTROL_REG_GENERATE_NMI_POS uint8 = 7;


	// VRAMのアドレスマスク
	PPU_ADDR_MIRROR_MASK = 0b111111_11111111 // 14ビット
)


// MARK: アドレスレジスタ(0x2006)
type AddrRegister struct {
	upper uint8 // 上位ビット
	lower uint8 // 下位ビット
	isUpper bool // 現在のビットが上位ビットかどうか
}

// アドレスレジスタのコンストラクタ
func (ar *AddrRegister) Init() {
	ar.upper = 0x00
	ar.lower = 0x00
	ar.isUpper = true // 最初は上位ビットから見る
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
	if ar.isUpper {
		ar.upper = data
	} else {
		ar.lower = data
	}

	// アドレスのミラーリング
	if ar.get() > PPU_REG_END {
		ar.set(ar.get() & PPU_ADDR_MIRROR_MASK)
	}

	// ビット位置を変更
	ar.isUpper = !ar.isUpper
}

// アドレスをインクリメントするメソッド
func (ar *AddrRegister) increment(step uint8) {
	// 方向によって 1 or 32 増やす
	current := ar.get()
	result := current + uint16(step)

	// アドレスのミラーリング
	if result > PPU_REG_END {
		ar.set(result & PPU_ADDR_MIRROR_MASK)
	}
}

// ビット位置をリセット
func (ar *AddrRegister) ResetLatch() {
	ar.isUpper = true
}


// MARK: コントロールレジスタ
type ControlRegister struct {
	NameTable1 bool
	NameTable2 bool
	VRAMAddIncrement bool
	SpritePatternAddr bool
	BackgroundPatternAddr bool
	SpriteSize bool
	MasterSlaveSelect bool
	GenerateNMI bool
}

// コントロールレジスタのコンストラクタ
func (cr *ControlRegister) Init() {
	cr.update(0x00)
}

// VRAMアドレスの増分を取得するメソッド
func (cr *ControlRegister) GetVRAMAddrIncrement() uint8 {
	if !cr.VRAMAddIncrement {
		return 1
	} else {
		return 32
	}
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
	if cr.SpritePatternAddr {
		value |= 1 << CONTROL_REG_SPRITE_PATTERN_ADDR_POS
	}
	if cr.BackgroundPatternAddr {
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
	cr.SpritePatternAddr = (value & (1 << CONTROL_REG_SPRITE_PATTERN_ADDR_POS)) != 0
	cr.BackgroundPatternAddr = (value & (1 << CONTROL_REG_BACKROUND_PATTERN_ADDR_POS)) != 0
	cr.SpriteSize = (value & (1 << CONTROL_REG_SPRITE_SIZE_POS)) != 0
	cr.MasterSlaveSelect = (value & (1 << CONTROL_REG_MASTER_SLAVE_SELECT_POS)) != 0
	cr.GenerateNMI = (value & (1 << CONTROL_REG_GENERATE_NMI_POS)) != 0
}