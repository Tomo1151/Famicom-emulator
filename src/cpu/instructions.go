package cpu

// MARK: アドレッシングモードの定義
type AddressingMode uint8

const (
	Implied          AddressingMode = iota // impl
	Accumulator                            // A
	Immediate                              // #
	ZeroPage                               // zpg
	ZeroPageXIndexed                       // zpg,X
	ZeroPageYIndexed                       // zpg,Y
	Absolute                               // abs
	AbsoluteXIndexed                       // abs,X
	AbsoluteYIndexed                       // abs,Y
	Relative                               // rel
	Indirect                               // ind
	IndexedIndirect                        // X,ind
	IndirectIndexed                        // ind,Y
)

// MARK: 命令の定義
type instruction struct {
	Mnemonic       string
	Opcode         uint8
	AddressingMode AddressingMode
	Bytes          uint8
	Cycles         uint8
	Jump           bool
	Handler        func(mode AddressingMode)
}

// MARK: 命令セットの定義
type instructionSet [256]instruction

// MARK: 命令セットの生成関数
func generateInstructionSet(c *CPU) instructionSet {
	instructionSet := instructionSet{}

	// MARK: 算術演算系 公式命令
	// ADC命令

	instructionSet[0x69] = instruction{
		Mnemonic:       "ADC",
		Opcode:         0x69,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.adc,
	}

	instructionSet[0x65] = instruction{
		Mnemonic:       "ADC",
		Opcode:         0x65,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.adc,
	}

	instructionSet[0x75] = instruction{
		Mnemonic:       "ADC",
		Opcode:         0x75,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.adc,
	}

	instructionSet[0x6D] = instruction{
		Mnemonic:       "ADC",
		Opcode:         0x6D,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.adc,
	}

	instructionSet[0x7D] = instruction{
		Mnemonic:       "ADC",
		Opcode:         0x7D,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.adc,
	}

	instructionSet[0x79] = instruction{
		Mnemonic:       "ADC",
		Opcode:         0x79,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.adc,
	}

	instructionSet[0x61] = instruction{
		Mnemonic:       "ADC",
		Opcode:         0x61,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.adc,
	}

	instructionSet[0x71] = instruction{
		Mnemonic:       "ADC",
		Opcode:         0x71,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.adc,
	}

	// DEC命令
	instructionSet[0xC6] = instruction{
		Mnemonic:       "DEC",
		Opcode:         0xC6,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.dec,
	}

	instructionSet[0xD6] = instruction{
		Mnemonic:       "DEC",
		Opcode:         0xD6,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.dec,
	}

	instructionSet[0xCE] = instruction{
		Mnemonic:       "DEC",
		Opcode:         0xCE,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.dec,
	}

	instructionSet[0xDE] = instruction{
		Mnemonic:       "DEC",
		Opcode:         0xDE,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.dec,
	}

	// DEX命令
	instructionSet[0xCA] = instruction{
		Mnemonic:       "DEX",
		Opcode:         0xCA,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.dex,
	}

	// DEY命令
	instructionSet[0x88] = instruction{
		Mnemonic:       "DEY",
		Opcode:         0x88,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.dey,
	}

	// INC命令
	instructionSet[0xE6] = instruction{
		Mnemonic:       "INC",
		Opcode:         0xE6,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.inc,
	}

	instructionSet[0xF6] = instruction{
		Mnemonic:       "INC",
		Opcode:         0xF6,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.inc,
	}

	instructionSet[0xEE] = instruction{
		Mnemonic:       "INC",
		Opcode:         0xEE,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.inc,
	}

	instructionSet[0xFE] = instruction{
		Mnemonic:       "INC",
		Opcode:         0xFE,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.inc,
	}

	// INX命令
	instructionSet[0xE8] = instruction{
		Mnemonic:       "INX",
		Opcode:         0xE8,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.inx,
	}

	// INY命令
	instructionSet[0xC8] = instruction{
		Mnemonic:       "INY",
		Opcode:         0xC8,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.iny,
	}

	// SBC命令
	instructionSet[0xE9] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xE9,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.sbc,
	}

	instructionSet[0xE5] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xE5,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.sbc,
	}

	instructionSet[0xF5] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xF5,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.sbc,
	}

	instructionSet[0xED] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xED,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.sbc,
	}

	instructionSet[0xFD] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xFD,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.sbc,
	}

	instructionSet[0xF9] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xF9,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.sbc,
	}

	instructionSet[0xE1] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xE1,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.sbc,
	}

	instructionSet[0xF1] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xF1,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.sbc,
	}

	// MARK: ビット演算系 公式命令
	// AND命令
	instructionSet[0x29] = instruction{
		Mnemonic:       "AND",
		Opcode:         0x29,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.and,
	}

	instructionSet[0x25] = instruction{
		Mnemonic:       "AND",
		Opcode:         0x25,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.and,
	}

	instructionSet[0x35] = instruction{
		Mnemonic:       "AND",
		Opcode:         0x35,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.and,
	}

	instructionSet[0x2D] = instruction{
		Mnemonic:       "AND",
		Opcode:         0x2D,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.and,
	}

	instructionSet[0x3D] = instruction{
		Mnemonic:       "AND",
		Opcode:         0x3D,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.and,
	}

	instructionSet[0x39] = instruction{
		Mnemonic:       "AND",
		Opcode:         0x39,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.and,
	}

	instructionSet[0x21] = instruction{
		Mnemonic:       "AND",
		Opcode:         0x21,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.and,
	}

	instructionSet[0x31] = instruction{
		Mnemonic:       "AND",
		Opcode:         0x31,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.and,
	}

	// BIT命令
	instructionSet[0x24] = instruction{
		Mnemonic:       "BIT",
		Opcode:         0x24,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.bit,
	}

	instructionSet[0x2C] = instruction{
		Mnemonic:       "BIT",
		Opcode:         0x2C,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.bit,
	}

	// EOR命令
	instructionSet[0x49] = instruction{
		Mnemonic:       "EOR",
		Opcode:         0x49,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.eor,
	}

	instructionSet[0x45] = instruction{
		Mnemonic:       "EOR",
		Opcode:         0x45,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.eor,
	}

	instructionSet[0x55] = instruction{
		Mnemonic:       "EOR",
		Opcode:         0x55,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.eor,
	}

	instructionSet[0x4D] = instruction{
		Mnemonic:       "EOR",
		Opcode:         0x4D,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.eor,
	}

	instructionSet[0x5D] = instruction{
		Mnemonic:       "EOR",
		Opcode:         0x5D,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.eor,
	}

	instructionSet[0x59] = instruction{
		Mnemonic:       "EOR",
		Opcode:         0x59,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.eor,
	}

	instructionSet[0x41] = instruction{
		Mnemonic:       "EOR",
		Opcode:         0x41,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.eor,
	}

	instructionSet[0x51] = instruction{
		Mnemonic:       "EOR",
		Opcode:         0x51,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.eor,
	}

	// ORA命令
	instructionSet[0x09] = instruction{
		Mnemonic:       "ORA",
		Opcode:         0x09,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.ora,
	}

	instructionSet[0x05] = instruction{
		Mnemonic:       "ORA",
		Opcode:         0x05,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.ora,
	}

	instructionSet[0x15] = instruction{
		Mnemonic:       "ORA",
		Opcode:         0x15,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.ora,
	}

	instructionSet[0x0D] = instruction{
		Mnemonic:       "ORA",
		Opcode:         0x0D,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.ora,
	}

	instructionSet[0x1D] = instruction{
		Mnemonic:       "ORA",
		Opcode:         0x1D,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.ora,
	}

	instructionSet[0x19] = instruction{
		Mnemonic:       "ORA",
		Opcode:         0x19,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.ora,
	}

	instructionSet[0x01] = instruction{
		Mnemonic:       "ORA",
		Opcode:         0x01,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.ora,
	}

	instructionSet[0x11] = instruction{
		Mnemonic:       "ORA",
		Opcode:         0x11,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.ora,
	}

	// MARK: ビットシフト系 公式命令
	// ASL命令
	instructionSet[0x0A] = instruction{
		Mnemonic:       "ASL",
		Opcode:         0x0A,
		AddressingMode: Accumulator,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.asl,
	}

	instructionSet[0x06] = instruction{
		Mnemonic:       "ASL",
		Opcode:         0x06,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.asl,
	}

	instructionSet[0x16] = instruction{
		Mnemonic:       "ASL",
		Opcode:         0x16,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.asl,
	}

	instructionSet[0x0E] = instruction{
		Mnemonic:       "ASL",
		Opcode:         0x0E,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.asl,
	}

	instructionSet[0x1E] = instruction{
		Mnemonic:       "ASL",
		Opcode:         0x1E,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.asl,
	}

	// LSR命令
	instructionSet[0x4A] = instruction{
		Mnemonic:       "LSR",
		Opcode:         0x4A,
		AddressingMode: Accumulator,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.lsr,
	}

	instructionSet[0x46] = instruction{
		Mnemonic:       "LSR",
		Opcode:         0x46,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.lsr,
	}

	instructionSet[0x56] = instruction{
		Mnemonic:       "LSR",
		Opcode:         0x56,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.lsr,
	}

	instructionSet[0x4E] = instruction{
		Mnemonic:       "LSR",
		Opcode:         0x4E,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.lsr,
	}

	instructionSet[0x5E] = instruction{
		Mnemonic:       "LSR",
		Opcode:         0x5E,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.lsr,
	}

	// ROL命令
	instructionSet[0x2A] = instruction{
		Mnemonic:       "ROL",
		Opcode:         0x2A,
		AddressingMode: Accumulator,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.rol,
	}

	instructionSet[0x26] = instruction{
		Mnemonic:       "ROL",
		Opcode:         0x26,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.rol,
	}

	instructionSet[0x36] = instruction{
		Mnemonic:       "ROL",
		Opcode:         0x36,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.rol,
	}

	instructionSet[0x2E] = instruction{
		Mnemonic:       "ROL",
		Opcode:         0x2E,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.rol,
	}

	instructionSet[0x3E] = instruction{
		Mnemonic:       "ROL",
		Opcode:         0x3E,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.rol,
	}

	// ROR命令
	instructionSet[0x6A] = instruction{
		Mnemonic:       "ROR",
		Opcode:         0x6A,
		AddressingMode: Accumulator,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.ror,
	}

	instructionSet[0x66] = instruction{
		Mnemonic:       "ROR",
		Opcode:         0x66,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.ror,
	}

	instructionSet[0x76] = instruction{
		Mnemonic:       "ROR",
		Opcode:         0x76,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.ror,
	}

	instructionSet[0x6E] = instruction{
		Mnemonic:       "ROR",
		Opcode:         0x6E,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.ror,
	}

	instructionSet[0x7E] = instruction{
		Mnemonic:       "ROR",
		Opcode:         0x7E,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.ror,
	}

	// MARK: 条件分岐系 公式命令
	// BCC命令
	instructionSet[0x90] = instruction{
		Mnemonic:       "BCC",
		Opcode:         0x90,
		AddressingMode: Relative,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.bcc,
	}

	// BCS命令
	instructionSet[0xB0] = instruction{
		Mnemonic:       "BCS",
		Opcode:         0xB0,
		AddressingMode: Relative,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.bcs,
	}

	// BEQ命令
	instructionSet[0xF0] = instruction{
		Mnemonic:       "BEQ",
		Opcode:         0xF0,
		AddressingMode: Relative,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.beq,
	}

	// BMI命令
	instructionSet[0x30] = instruction{
		Mnemonic:       "BMI",
		Opcode:         0x30,
		AddressingMode: Relative,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.bmi,
	}

	// BNE命令
	instructionSet[0xD0] = instruction{
		Mnemonic:       "BNE",
		Opcode:         0xD0,
		AddressingMode: Relative,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.bne,
	}

	// BPL命令
	instructionSet[0x10] = instruction{
		Mnemonic:       "BPL",
		Opcode:         0x10,
		AddressingMode: Relative,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.bpl,
	}

	// BVC命令
	instructionSet[0x50] = instruction{
		Mnemonic:       "BVC",
		Opcode:         0x50,
		AddressingMode: Relative,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.bvc,
	}

	// BVS命令
	instructionSet[0x70] = instruction{
		Mnemonic:       "BVS",
		Opcode:         0x70,
		AddressingMode: Relative,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.bvs,
	}

	// MARK: ジャンプ系 公式命令
	// BRK命令
	instructionSet[0x00] = instruction{
		Mnemonic:       "BRK",
		Opcode:         0x00,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         7,
		Jump:           true,
		Handler:        c.brk,
	}

	// JMP命令
	instructionSet[0x4C] = instruction{
		Mnemonic:       "JMP",
		Opcode:         0x4C,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         3,
		Jump:           true,
		Handler:        c.jmp,
	}

	instructionSet[0x6C] = instruction{
		Mnemonic:       "JMP",
		Opcode:         0x6C,
		AddressingMode: Indirect,
		Bytes:          3,
		Cycles:         5,
		Jump:           true,
		Handler:        c.jmp,
	}

	// JSR命令
	instructionSet[0x20] = instruction{
		Mnemonic:       "JSR",
		Opcode:         0x20,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Jump:           true,
		Handler:        c.jsr,
	}

	// RTI命令
	instructionSet[0x40] = instruction{
		Mnemonic:       "RTI",
		Opcode:         0x40,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         6,
		Jump:           true,
		Handler:        c.rti,
	}

	// RTS命令
	instructionSet[0x60] = instruction{
		Mnemonic:       "RTS",
		Opcode:         0x60,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         6,
		Jump:           true,
		Handler:        c.rts,
	}

	// MARK: フラグ操作系 公式命令
	// CLC命令
	instructionSet[0x18] = instruction{
		Mnemonic:       "CLC",
		Opcode:         0x18,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.clc,
	}

	// CLD命令
	instructionSet[0xD8] = instruction{
		Mnemonic:       "CLD",
		Opcode:         0xD8,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.cld,
	}

	// CLI命令
	instructionSet[0x58] = instruction{
		Mnemonic:       "CLI",
		Opcode:         0x58,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.cli,
	}

	// CLV命令
	instructionSet[0xB8] = instruction{
		Mnemonic:       "CLV",
		Opcode:         0xB8,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.clv,
	}

	// SEC命令
	instructionSet[0x38] = instruction{
		Mnemonic:       "SEC",
		Opcode:         0x38,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.sec,
	}

	// SED命令
	instructionSet[0xF8] = instruction{
		Mnemonic:       "SED",
		Opcode:         0xF8,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.sed,
	}

	// SEI命令
	instructionSet[0x78] = instruction{
		Mnemonic:       "SEI",
		Opcode:         0x78,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.sei,
	}

	// MARK: 比較系 公式命令
	// CMP命令
	instructionSet[0xC9] = instruction{
		Mnemonic:       "CMP",
		Opcode:         0xC9,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.cmp,
	}

	instructionSet[0xC5] = instruction{
		Mnemonic:       "CMP",
		Opcode:         0xC5,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.cmp,
	}

	instructionSet[0xD5] = instruction{
		Mnemonic:       "CMP",
		Opcode:         0xD5,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.cmp,
	}

	instructionSet[0xCD] = instruction{
		Mnemonic:       "CMP",
		Opcode:         0xCD,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.cmp,
	}

	instructionSet[0xDD] = instruction{
		Mnemonic:       "CMP",
		Opcode:         0xDD,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.cmp,
	}

	instructionSet[0xD9] = instruction{
		Mnemonic:       "CMP",
		Opcode:         0xD9,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.cmp,
	}

	instructionSet[0xC1] = instruction{
		Mnemonic:       "CMP",
		Opcode:         0xC1,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.cmp,
	}

	instructionSet[0xD1] = instruction{
		Mnemonic:       "CMP",
		Opcode:         0xD1,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.cmp,
	}

	// CPX命令
	instructionSet[0xE0] = instruction{
		Mnemonic:       "CPX",
		Opcode:         0xE0,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.cpx,
	}

	instructionSet[0xE4] = instruction{
		Mnemonic:       "CPX",
		Opcode:         0xE4,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.cpx,
	}

	instructionSet[0xEC] = instruction{
		Mnemonic:       "CPX",
		Opcode:         0xEC,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.cpx,
	}

	// CPY命令
	instructionSet[0xC0] = instruction{
		Mnemonic:       "CPY",
		Opcode:         0xC0,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.cpy,
	}

	instructionSet[0xC4] = instruction{
		Mnemonic:       "CPY",
		Opcode:         0xC4,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.cpy,
	}

	instructionSet[0xCC] = instruction{
		Mnemonic:       "CPY",
		Opcode:         0xCC,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.cpy,
	}

	// MARK: データアクセス系 公式命令
	// LDA命令
	instructionSet[0xA9] = instruction{
		Mnemonic:       "LDA",
		Opcode:         0xA9,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.lda,
	}

	instructionSet[0xA5] = instruction{
		Mnemonic:       "LDA",
		Opcode:         0xA5,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.lda,
	}

	instructionSet[0xB5] = instruction{
		Mnemonic:       "LDA",
		Opcode:         0xB5,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.lda,
	}

	instructionSet[0xAD] = instruction{
		Mnemonic:       "LDA",
		Opcode:         0xAD,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.lda,
	}

	instructionSet[0xBD] = instruction{
		Mnemonic:       "LDA",
		Opcode:         0xBD,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.lda,
	}

	instructionSet[0xB9] = instruction{
		Mnemonic:       "LDA",
		Opcode:         0xB9,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.lda,
	}

	instructionSet[0xA1] = instruction{
		Mnemonic:       "LDA",
		Opcode:         0xA1,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.lda,
	}

	instructionSet[0xB1] = instruction{
		Mnemonic:       "LDA",
		Opcode:         0xB1,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.lda,
	}

	// LDX命令
	instructionSet[0xA2] = instruction{
		Mnemonic:       "LDX",
		Opcode:         0xA2,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.ldx,
	}

	instructionSet[0xA6] = instruction{
		Mnemonic:       "LDX",
		Opcode:         0xA6,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.ldx,
	}

	instructionSet[0xB6] = instruction{
		Mnemonic:       "LDX",
		Opcode:         0xB6,
		AddressingMode: ZeroPageYIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.ldx,
	}

	instructionSet[0xAE] = instruction{
		Mnemonic:       "LDX",
		Opcode:         0xAE,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.ldx,
	}

	instructionSet[0xBE] = instruction{
		Mnemonic:       "LDX",
		Opcode:         0xBE,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.ldx,
	}

	// LDY命令
	instructionSet[0xA0] = instruction{
		Mnemonic:       "LDY",
		Opcode:         0xA0,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.ldy,
	}

	instructionSet[0xA4] = instruction{
		Mnemonic:       "LDY",
		Opcode:         0xA4,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.ldy,
	}

	instructionSet[0xB4] = instruction{
		Mnemonic:       "LDY",
		Opcode:         0xB4,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.ldy,
	}

	instructionSet[0xAC] = instruction{
		Mnemonic:       "LDY",
		Opcode:         0xAC,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.ldy,
	}

	instructionSet[0xBC] = instruction{
		Mnemonic:       "LDY",
		Opcode:         0xBC,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.ldy,
	}

	// STA命令
	instructionSet[0x85] = instruction{
		Mnemonic:       "STA",
		Opcode:         0x85,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.sta,
	}

	instructionSet[0x95] = instruction{
		Mnemonic:       "STA",
		Opcode:         0x95,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.sta,
	}

	instructionSet[0x8D] = instruction{
		Mnemonic:       "STA",
		Opcode:         0x8D,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.sta,
	}

	instructionSet[0x9D] = instruction{
		Mnemonic:       "STA",
		Opcode:         0x9D,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         5,
		Handler:        c.sta,
	}

	instructionSet[0x99] = instruction{
		Mnemonic:       "STA",
		Opcode:         0x99,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         5,
		Handler:        c.sta,
	}

	instructionSet[0x81] = instruction{
		Mnemonic:       "STA",
		Opcode:         0x81,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.sta,
	}

	instructionSet[0x91] = instruction{
		Mnemonic:       "STA",
		Opcode:         0x91,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.sta,
	}

	// STX命令
	instructionSet[0x86] = instruction{
		Mnemonic:       "STX",
		Opcode:         0x86,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.stx,
	}

	instructionSet[0x96] = instruction{
		Mnemonic:       "STX",
		Opcode:         0x96,
		AddressingMode: ZeroPageYIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.stx,
	}

	instructionSet[0x8E] = instruction{
		Mnemonic:       "STX",
		Opcode:         0x8E,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.stx,
	}

	// STY命令
	instructionSet[0x84] = instruction{
		Mnemonic:       "STY",
		Opcode:         0x84,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.sty,
	}

	instructionSet[0x94] = instruction{
		Mnemonic:       "STY",
		Opcode:         0x94,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.sty,
	}

	instructionSet[0x8C] = instruction{
		Mnemonic:       "STY",
		Opcode:         0x8C,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.sty,
	}

	// MARK: スタック操作系 公式命令
	// PHA命令
	instructionSet[0x48] = instruction{
		Mnemonic:       "PHA",
		Opcode:         0x48,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         3,
		Handler:        c.pha,
	}

	// PHP命令
	instructionSet[0x08] = instruction{
		Mnemonic:       "PHP",
		Opcode:         0x08,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         3,
		Handler:        c.php,
	}

	// PLA命令
	instructionSet[0x68] = instruction{
		Mnemonic:       "PLA",
		Opcode:         0x68,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         4,
		Handler:        c.pla,
	}

	// PLP命令
	instructionSet[0x28] = instruction{
		Mnemonic:       "PLP",
		Opcode:         0x28,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         4,
		Handler:        c.plp,
	}

	// MARK: データ転送系 公式命令
	// TAX命令
	instructionSet[0xAA] = instruction{
		Mnemonic:       "TAX",
		Opcode:         0xAA,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.tax,
	}

	// TAY命令
	instructionSet[0xA8] = instruction{
		Mnemonic:       "TAY",
		Opcode:         0xA8,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.tay,
	}

	// TSX命令
	instructionSet[0xBA] = instruction{
		Mnemonic:       "TSX",
		Opcode:         0xBA,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.tsx,
	}

	// TXA命令
	instructionSet[0x8A] = instruction{
		Mnemonic:       "TXA",
		Opcode:         0x8A,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.txa,
	}

	// TXS命令
	instructionSet[0x9A] = instruction{
		Mnemonic:       "TXS",
		Opcode:         0x9A,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.txs,
	}

	// TYA命令
	instructionSet[0x98] = instruction{
		Mnemonic:       "TYA",
		Opcode:         0x98,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.tya,
	}

	// NOP命令
	instructionSet[0xEA] = instruction{
		Mnemonic:       "NOP",
		Opcode:         0xEA,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.nop,
	}

	// MARK: 非公式命令
	// ALR命令 (ASR)
	instructionSet[0x4B] = instruction{
		Mnemonic:       "ALR",
		Opcode:         0x4B,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.alr,
	}

	// ANC命令 (AAC)
	instructionSet[0x0B] = instruction{
		Mnemonic:       "ANC",
		Opcode:         0x0B,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.anc,
	}

	instructionSet[0x2B] = instruction{
		Mnemonic:       "ANC",
		Opcode:         0x2B,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.anc,
	}

	// ARR命令 (ARR)
	instructionSet[0x6B] = instruction{
		Mnemonic:       "ARR",
		Opcode:         0x6B,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.arr,
	}

	// AXS命令 (SBX / SAX)
	instructionSet[0xCB] = instruction{
		Mnemonic:       "AXS",
		Opcode:         0xCB,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.axs,
	}

	// LXA命令 (ATX / LXA / OAL)
	instructionSet[0xAB] = instruction{
		Mnemonic:       "LXA",
		Opcode:         0xAB,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.lxa,
	}

	// SAX命令 (AAX / AXS)
	instructionSet[0x87] = instruction{
		Mnemonic:       "SAX",
		Opcode:         0x87,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.sax,
	}

	instructionSet[0x97] = instruction{
		Mnemonic:       "SAX",
		Opcode:         0x97,
		AddressingMode: ZeroPageYIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.sax,
	}

	instructionSet[0x83] = instruction{
		Mnemonic:       "SAX",
		Opcode:         0x83,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.sax,
	}

	instructionSet[0x8F] = instruction{
		Mnemonic:       "SAX",
		Opcode:         0x8F,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.sax,
	}

	// AHX命令 (AXA / SHA)
	instructionSet[0x93] = instruction{
		Mnemonic:       "AHX",
		Opcode:         0x93,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.ahx,
	}

	instructionSet[0x9F] = instruction{
		Mnemonic:       "AHX",
		Opcode:         0x9F,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         5,
		Handler:        c.ahx,
	}

	// DCP命令 (DCM)
	instructionSet[0xC7] = instruction{
		Mnemonic:       "DCP",
		Opcode:         0xC7,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.dcp,
	}

	instructionSet[0xD7] = instruction{
		Mnemonic:       "DCP",
		Opcode:         0xD7,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.dcp,
	}

	instructionSet[0xCF] = instruction{
		Mnemonic:       "DCP",
		Opcode:         0xCF,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.dcp,
	}

	instructionSet[0xDF] = instruction{
		Mnemonic:       "DCP",
		Opcode:         0xDF,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.dcp,
	}

	instructionSet[0xDB] = instruction{
		Mnemonic:       "DCP",
		Opcode:         0xDB,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.dcp,
	}

	instructionSet[0xC3] = instruction{
		Mnemonic:       "DCP",
		Opcode:         0xC3,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.dcp,
	}

	instructionSet[0xD3] = instruction{
		Mnemonic:       "DCP",
		Opcode:         0xD3,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.dcp,
	}

	// ISC命令 (ISB / INS)
	instructionSet[0xE7] = instruction{
		Mnemonic:       "ISC",
		Opcode:         0xE7,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.isc,
	}

	instructionSet[0xF7] = instruction{
		Mnemonic:       "ISC",
		Opcode:         0xF7,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.isc,
	}

	instructionSet[0xEF] = instruction{
		Mnemonic:       "ISC",
		Opcode:         0xEF,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.isc,
	}

	instructionSet[0xFF] = instruction{
		Mnemonic:       "ISC",
		Opcode:         0xFF,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.isc,
	}

	instructionSet[0xFB] = instruction{
		Mnemonic:       "ISC",
		Opcode:         0xFB,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.isc,
	}

	instructionSet[0xE3] = instruction{
		Mnemonic:       "ISC",
		Opcode:         0xE3,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.isc,
	}

	instructionSet[0xF3] = instruction{
		Mnemonic:       "ISC",
		Opcode:         0xF3,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.isc,
	}

	// LAS命令 (LAR / LAE)
	instructionSet[0xBB] = instruction{
		Mnemonic:       "LAS",
		Opcode:         0xBB,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.las,
	}

	// LAX命令 (LAX)
	instructionSet[0xA7] = instruction{
		Mnemonic:       "LAX",
		Opcode:         0xA7,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.lax,
	}

	instructionSet[0xB7] = instruction{
		Mnemonic:       "LAX",
		Opcode:         0xB7,
		AddressingMode: ZeroPageYIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.lax,
	}

	instructionSet[0xAF] = instruction{
		Mnemonic:       "LAX",
		Opcode:         0xAF,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.lax,
	}

	instructionSet[0xBF] = instruction{
		Mnemonic:       "LAX",
		Opcode:         0xBF,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.lax,
	}

	instructionSet[0xA3] = instruction{
		Mnemonic:       "LAX",
		Opcode:         0xA3,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.lax,
	}

	instructionSet[0xB3] = instruction{
		Mnemonic:       "LAX",
		Opcode:         0xB3,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.lax,
	}

	// RLA命令 (RLA)
	instructionSet[0x27] = instruction{
		Mnemonic:       "RLA",
		Opcode:         0x27,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.rla,
	}

	instructionSet[0x37] = instruction{
		Mnemonic:       "RLA",
		Opcode:         0x37,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.rla,
	}

	instructionSet[0x2F] = instruction{
		Mnemonic:       "RLA",
		Opcode:         0x2F,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.rla,
	}

	instructionSet[0x3F] = instruction{
		Mnemonic:       "RLA",
		Opcode:         0x3F,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.rla,
	}

	instructionSet[0x3B] = instruction{
		Mnemonic:       "RLA",
		Opcode:         0x3B,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.rla,
	}

	instructionSet[0x23] = instruction{
		Mnemonic:       "RLA",
		Opcode:         0x23,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.rla,
	}

	instructionSet[0x33] = instruction{
		Mnemonic:       "RLA",
		Opcode:         0x33,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.rla,
	}

	// RRA命令 (RRA)
	instructionSet[0x67] = instruction{
		Mnemonic:       "RRA",
		Opcode:         0x67,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.rra,
	}

	instructionSet[0x77] = instruction{
		Mnemonic:       "RRA",
		Opcode:         0x77,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.rra,
	}

	instructionSet[0x6F] = instruction{
		Mnemonic:       "RRA",
		Opcode:         0x6F,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.rra,
	}

	instructionSet[0x7F] = instruction{
		Mnemonic:       "RRA",
		Opcode:         0x7F,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.rra,
	}

	instructionSet[0x7B] = instruction{
		Mnemonic:       "RRA",
		Opcode:         0x7B,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.rra,
	}

	instructionSet[0x63] = instruction{
		Mnemonic:       "RRA",
		Opcode:         0x63,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.rra,
	}

	instructionSet[0x73] = instruction{
		Mnemonic:       "RRA",
		Opcode:         0x73,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.rra,
	}

	// SBC命令 (SBC)
	instructionSet[0xEB] = instruction{
		Mnemonic:       "SBC",
		Opcode:         0xEB,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.sbc,
	}

	// SLO命令 (ASO)
	instructionSet[0x07] = instruction{
		Mnemonic:       "SLO",
		Opcode:         0x07,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.slo,
	}

	instructionSet[0x17] = instruction{
		Mnemonic:       "SLO",
		Opcode:         0x17,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.slo,
	}

	instructionSet[0x0F] = instruction{
		Mnemonic:       "SLO",
		Opcode:         0x0F,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.slo,
	}

	instructionSet[0x1F] = instruction{
		Mnemonic:       "SLO",
		Opcode:         0x1F,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.slo,
	}

	instructionSet[0x1B] = instruction{
		Mnemonic:       "SLO",
		Opcode:         0x1B,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.slo,
	}

	instructionSet[0x03] = instruction{
		Mnemonic:       "SLO",
		Opcode:         0x03,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.slo,
	}

	instructionSet[0x13] = instruction{
		Mnemonic:       "SLO",
		Opcode:         0x13,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.slo,
	}

	// SRE命令 (LSE)
	instructionSet[0x47] = instruction{
		Mnemonic:       "SRE",
		Opcode:         0x47,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         5,
		Handler:        c.sre,
	}

	instructionSet[0x57] = instruction{
		Mnemonic:       "SRE",
		Opcode:         0x57,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         6,
		Handler:        c.sre,
	}

	instructionSet[0x4F] = instruction{
		Mnemonic:       "SRE",
		Opcode:         0x4F,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         6,
		Handler:        c.sre,
	}

	instructionSet[0x5F] = instruction{
		Mnemonic:       "SRE",
		Opcode:         0x5F,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.sre,
	}

	instructionSet[0x5B] = instruction{
		Mnemonic:       "SRE",
		Opcode:         0x5B,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         7,
		Handler:        c.sre,
	}

	instructionSet[0x43] = instruction{
		Mnemonic:       "SRE",
		Opcode:         0x43,
		AddressingMode: IndexedIndirect,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.sre,
	}

	instructionSet[0x53] = instruction{
		Mnemonic:       "SRE",
		Opcode:         0x53,
		AddressingMode: IndirectIndexed,
		Bytes:          2,
		Cycles:         8,
		Handler:        c.sre,
	}

	// TAS命令 (XAS / SHS)
	instructionSet[0x9B] = instruction{
		Mnemonic:       "TAS",
		Opcode:         0x9B,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         5,
		Handler:        c.tas,
	}

	// SHX命令 (SXA / XAS)
	instructionSet[0x9E] = instruction{
		Mnemonic:       "SHX",
		Opcode:         0x9E,
		AddressingMode: AbsoluteYIndexed,
		Bytes:          3,
		Cycles:         5,
		Handler:        c.shx,
	}

	// SHY命令 (SYA / SAY)
	instructionSet[0x9C] = instruction{
		Mnemonic:       "SHY",
		Opcode:         0x9C,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         5,
		Handler:        c.shy,
	}

	// KIL命令 (JAM / HLT)
	instructionSet[0x02] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x02,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0x12] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x12,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0x22] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x22,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0x32] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x32,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0x42] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x42,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0x52] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x52,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0x62] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x62,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0x72] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x72,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0x92] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0x92,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0xB2] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0xB2,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0xD2] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0xD2,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	instructionSet[0xF2] = instruction{
		Mnemonic:       "KIL",
		Opcode:         0xF2,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         0,
		Handler:        c.kil,
	}

	// NOP命令 (NOP)
	instructionSet[0x1A] = instruction{
		Mnemonic:       "NOP",
		Opcode:         0x1A,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.nop,
	}

	instructionSet[0x3A] = instruction{
		Mnemonic:       "NOP",
		Opcode:         0x3A,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.nop,
	}

	instructionSet[0x5A] = instruction{
		Mnemonic:       "NOP",
		Opcode:         0x5A,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.nop,
	}

	instructionSet[0x7A] = instruction{
		Mnemonic:       "NOP",
		Opcode:         0x7A,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.nop,
	}

	instructionSet[0xDA] = instruction{
		Mnemonic:       "NOP",
		Opcode:         0xDA,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.nop,
	}

	instructionSet[0xFA] = instruction{
		Mnemonic:       "NOP",
		Opcode:         0xFA,
		AddressingMode: Implied,
		Bytes:          1,
		Cycles:         2,
		Handler:        c.nop,
	}

	// DOP命令 (NOP / SKB / SKW)
	instructionSet[0x04] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x04,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.dop,
	}

	instructionSet[0x14] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x14,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.dop,
	}

	instructionSet[0x34] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x34,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.dop,
	}

	instructionSet[0x44] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x44,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.dop,
	}

	instructionSet[0x54] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x54,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.dop,
	}

	instructionSet[0x64] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x64,
		AddressingMode: ZeroPage,
		Bytes:          2,
		Cycles:         3,
		Handler:        c.dop,
	}

	instructionSet[0x74] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x74,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.dop,
	}

	instructionSet[0x80] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x80,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.dop,
	}

	instructionSet[0x82] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x82,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.dop,
	}

	instructionSet[0x89] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0x89,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.dop,
	}

	instructionSet[0xC2] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0xC2,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.dop,
	}

	instructionSet[0xD4] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0xD4,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.dop,
	}

	instructionSet[0xE2] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0xE2,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.dop,
	}

	instructionSet[0xF4] = instruction{
		Mnemonic:       "DOP",
		Opcode:         0xF4,
		AddressingMode: ZeroPageXIndexed,
		Bytes:          2,
		Cycles:         4,
		Handler:        c.dop,
	}

	// TOP命令 (IGN)
	instructionSet[0x0C] = instruction{
		Mnemonic:       "TOP",
		Opcode:         0x0C,
		AddressingMode: Absolute,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.top,
	}

	instructionSet[0x1C] = instruction{
		Mnemonic:       "TOP",
		Opcode:         0x1C,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.top,
	}

	instructionSet[0x3C] = instruction{
		Mnemonic:       "TOP",
		Opcode:         0x3C,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.top,
	}

	instructionSet[0x5C] = instruction{
		Mnemonic:       "TOP",
		Opcode:         0x5C,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.top,
	}

	instructionSet[0x7C] = instruction{
		Mnemonic:       "TOP",
		Opcode:         0x7C,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.top,
	}

	instructionSet[0xDC] = instruction{
		Mnemonic:       "TOP",
		Opcode:         0xDC,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.top,
	}

	instructionSet[0xFC] = instruction{
		Mnemonic:       "TOP",
		Opcode:         0xFC,
		AddressingMode: AbsoluteXIndexed,
		Bytes:          3,
		Cycles:         4,
		Handler:        c.top,
	}

	// XAA命令 (ANE)
	instructionSet[0x8B] = instruction{
		Mnemonic:       "XAA",
		Opcode:         0x8B,
		AddressingMode: Immediate,
		Bytes:          2,
		Cycles:         2,
		Handler:        c.xaa,
	}

	return instructionSet
}
