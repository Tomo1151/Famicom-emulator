package cpu

// レジスタ関連の定数
const (
	// ステータスレジスタのビット位置
	STATUS_REG_CARRY_POS uint8 = iota
	STATUS_REG_ZERO_POS
	STATUS_REG_INTERRUPT_POS
	STATUS_REG_DECIMAL_POS
	STATUS_REG_BREAK_POS
	STATUS_REG_RESERVED_POS
	STATUS_REG_OVERFLOW_POS
	STATUS_REG_NEGATIVE_POS
)

// レジスタの定義
type registers struct {
	A  uint8          // アキュムレータ: 8bit
	X  uint8          // インデックスレジスタ: 8bit
	Y  uint8          // インデックスレジスタ: 8bit
	SP uint8          // スタックポインタ: 8bit
	PC uint16         // プログラムカウンタ: 16bit
	P  statusRegister // ステータスレジスタ: 8bit
}

// ステータスレジスタ(P)の定義
type statusRegister struct {
	Negative  bool // N: 演算結果の最上位ビットが1の時にセット
	Overflow  bool // V: P演算結果がオーバーフローした時にセット
	Reserved  bool // R: 予約済み，常にセット
	Break     bool // B: BRK発生時にセットされIRQ発生時にクリア
	Decimal   bool // D: 0 -> デフォルト， 1 -> BCDモード (NESではBCDモードは未実装)
	Interrupt bool // I: 0 -> IRQ許可， 1 -> IRQ禁止
	Zero      bool // Z: 演算結果が0の時にセット
	Carry     bool // C: キャリー発生時にセット
}

// ステータスレジスタ(P)をuint8へ変換するメソッド
func (s *statusRegister) ToByte() uint8 {
	var value uint8 = 0x00

	if s.Negative {
		value |= 1 << STATUS_REG_NEGATIVE_POS
	}
	if s.Overflow {
		value |= 1 << STATUS_REG_OVERFLOW_POS
	}
	if s.Reserved {
		value |= 1 << STATUS_REG_RESERVED_POS
	}
	if s.Break {
		value |= 1 << STATUS_REG_BREAK_POS
	}
	if s.Decimal {
		value |= 1 << STATUS_REG_DECIMAL_POS
	}
	if s.Interrupt {
		value |= 1 << STATUS_REG_INTERRUPT_POS
	}
	if s.Zero {
		value |= 1 << STATUS_REG_ZERO_POS
	}
	if s.Carry {
		value |= 1 << STATUS_REG_CARRY_POS
	}

	return value
}

// uint8からステータスレジスタオブジェクトへ変換するメソッド
func (s *statusRegister) SetFromByte(value uint8) {
	s.Negative = (value & (1 << STATUS_REG_NEGATIVE_POS)) != 0
	s.Overflow = (value & (1 << STATUS_REG_OVERFLOW_POS)) != 0
	s.Reserved = (value & (1 << STATUS_REG_RESERVED_POS)) != 0
	s.Break = (value & (1 << STATUS_REG_BREAK_POS)) != 0
	s.Decimal = (value & (1 << STATUS_REG_DECIMAL_POS)) != 0
	s.Interrupt = (value & (1 << STATUS_REG_INTERRUPT_POS)) != 0
	s.Zero = (value & (1 << STATUS_REG_ZERO_POS)) != 0
	s.Carry = (value & (1 << STATUS_REG_CARRY_POS)) != 0
}
