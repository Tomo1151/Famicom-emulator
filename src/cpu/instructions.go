package cpu

import "fmt"

type Instruction struct {
	Opecode uint8 // オペコード
	Code InstructionCode // 命令の種類
	AddressingMode AddressingMode // アドレッシングモード
	Bytes uint8 // 命令のバイト数
	Cycles uint8 // 基本サイクル数
	PageCycles uint8 // ページ協会を越えた場合の追加サイクル
	Handler InstructionHandler // 命令の実装
}

// 命令の種類の型定義
type InstructionCode uint8

// アドレッシングモード型の定義
type AddressingMode uint8

type InstructionHandler func(mode AddressingMode)

// 命令セット型の定義
type instructionSet map[uint8]Instruction



// 定数の定義
const (
	Implied AddressingMode = iota // impl
	Accumulator // A
	Immediate // #
	ZeroPage // zpg
	ZeroPageXIndexed // zpg,X
	ZeroPageYIndexed // zpg,Y
	Absolute // abs
	AbsoluteXIndexed // abs,X
	AbsoluteYIndexed // abs,Y
	Relative // rel
	Indirect // Ind
	IndirectXIndexed // X,Ind
	IndirectYIndexed // Ind,Y
)


const (
	ADC InstructionCode = iota
	AND
	ASL
	BCC
	BCS
	BEQ
	BIT
	BMI
	BNE
	BPL
	BRK
	BVC
	BVS
	CLC
	CLD
	CLI
	CLV
	CMP
	CPX
	CPY
	DEC
	DEX
	DEY
	EOR
	INC
	INX
	INY
	JMP
	JSR
	LDA
	LDX
	LDY
	LSR
	NOP
	ORA
	PHA
	PHP
	PLA
	PLP
	ROL
	ROR
	RTI
	RTS
	SBC
	SEC
	SED
	SEI
	STA
	STX
	STY
	TAX
	TAY
	TSX
	TXA
	TXS
	TYA
)


// 命令セットの生成
func generateInstructionSet(c *CPU) instructionSet {
	instructionSet := make(instructionSet)

	// MARK: ADC命令
	instructionSet[0x69] = Instruction{
		Opecode: 0x69,
		Code: ADC,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.adc,
	}

	instructionSet[0x65] = Instruction{
		Opecode: 0x65,
		Code: ADC,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.adc,
	}

	instructionSet[0x75] = Instruction{
		Opecode: 0x75,
		Code: ADC,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.adc,
	}

	instructionSet[0x6D] = Instruction{
		Opecode: 0x6D,
		Code: ADC,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.adc,
	}

	instructionSet[0x7D] = Instruction{
		Opecode: 0x7D,
		Code: ADC,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.adc,
	}

	instructionSet[0x79] = Instruction{
		Opecode: 0x79,
		Code: ADC,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.adc,
	}

	instructionSet[0x61] = Instruction{
		Opecode: 0x61,
		Code: ADC,
		AddressingMode: IndirectXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.adc,
	}

	instructionSet[0x71] = Instruction{
		Opecode: 0x71,
		Code: ADC,
		AddressingMode: IndirectYIndexed,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 1,
		Handler: c.adc,
	}


	// MARK: AND命令
	instructionSet[0x29] = Instruction{
		Opecode: 0x29,
		Code: AND,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.and,
	}

	instructionSet[0x25] = Instruction{
		Opecode: 0x25,
		Code: AND,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.and,
	}

	instructionSet[0x35] = Instruction{
		Opecode: 0x35,
		Code: AND,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.and,
	}

	instructionSet[0x2D] = Instruction{
		Opecode: 0x2D,
		Code: AND,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.and,
	}

	instructionSet[0x3D] = Instruction{
		Opecode: 0x3D,
		Code: AND,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.and,
	}

	instructionSet[0x39] = Instruction{
		Opecode: 0x39,
		Code: AND,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.and,
	}

	instructionSet[0x21] = Instruction{
		Opecode: 0x21,
		Code: AND,
		AddressingMode: IndirectXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.and,
	}

	instructionSet[0x31] = Instruction{
		Opecode: 0x31,
		Code: AND,
		AddressingMode: IndirectYIndexed,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 1,
		Handler: c.and,
	}


	// MARK: ASL命令
	instructionSet[0x0A] = Instruction{
		Opecode: 0x0A,
		Code: ASL,
		AddressingMode: Accumulator,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.asl,
	}

	instructionSet[0x06] = Instruction{
		Opecode: 0x06,
		Code: ASL,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.asl,
	}

	instructionSet[0x16] = Instruction{
		Opecode: 0x16,
		Code: ASL,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.asl,
	}

	instructionSet[0x0E] = Instruction{
		Opecode: 0x0E,
		Code: ASL,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.asl,
	}

	instructionSet[0x1E] = Instruction{
		Opecode: 0x1E,
		Code: ASL,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 7,
		PageCycles: 0,
		Handler: c.asl,
	}


	// MARK: BCC命令
	instructionSet[0x90] = Instruction{
		Opecode: 0x90,
		Code: BCC,
		AddressingMode: Relative,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.bcc,
	}


	// MARK: BCS命令
	instructionSet[0xB0] = Instruction{
		Opecode: 0xB0,
		Code: BCS,
		AddressingMode: Relative,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.bcs,
	}


	// MARK: BEQ命令
	instructionSet[0xF0] = Instruction{
		Opecode: 0xF0,
		Code: BEQ,
		AddressingMode: Relative,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.beq,
	}


	// MARK: BIT命令
	instructionSet[0x24] = Instruction{
		Opecode: 0x24,
		Code: BIT,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.bit,
	}

	instructionSet[0x2C] = Instruction{
		Opecode: 0x2C,
		Code: BIT,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.bit,
	}


	// MARK: BMI命令
	instructionSet[0x30] = Instruction{
		Opecode: 0x30,
		Code: BMI,
		AddressingMode: Relative,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.bmi,
	}


	// MARK: BNE命令
	instructionSet[0xD0] = Instruction{
		Opecode: 0xD0,
		Code: BNE,
		AddressingMode: Relative,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.bne,
	}


	// MARK: BPL命令
	instructionSet[0x10] = Instruction{
		Opecode: 0x10,
		Code: BPL,
		AddressingMode: Relative,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.bpl,
	}


	// MARK: BRK命令
	instructionSet[0x00] = Instruction{
		Opecode: 0x00,
		Code: BRK,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 7,
		PageCycles: 0,
		Handler: c.brk,
	}


	// MARK: BVC命令
	instructionSet[0x50] = Instruction{
		Opecode: 0x50,
		Code: BVC,
		AddressingMode: Relative,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.bvc,
	}


	// MARK: BVS命令
	instructionSet[0x70] = Instruction{
		Opecode: 0x70,
		Code: BVS,
		AddressingMode: Relative,
		Bytes: 2,
		Cycles: 2, // @NOTE +1 if branch succeeds +2 if to a new page
		PageCycles: 0,
		Handler: c.bvs,
	}


	// MARK: CLC命令
	instructionSet[0x18] = Instruction{
		Opecode: 0x18,
		Code: CLC,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.clc,
	}


	// MARK: CLD命令
	instructionSet[0xD8] = Instruction{
		Opecode: 0xD8,
		Code: CLD,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.cld,
	}


	// MARK: CLI命令
	instructionSet[0x58] = Instruction{
		Opecode: 0x58,
		Code: CLI,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.cli,
	}


	// MARK: CLV命令
	instructionSet[0xB8] = Instruction{
		Opecode: 0xB8,
		Code: CLV,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.clv,
	}


	// MARK: CMP命令
	instructionSet[0xC9] = Instruction{
		Opecode: 0xC9,
		Code: CMP,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.cmp,
	}

	instructionSet[0xC5] = Instruction{
		Opecode: 0xC5,
		Code: CMP,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.cmp,
	}

	instructionSet[0xD5] = Instruction{
		Opecode: 0xD5,
		Code: CMP,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.cmp,
	}

	instructionSet[0xCD] = Instruction{
		Opecode: 0xCD,
		Code: CMP,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.cmp,
	}

	instructionSet[0xDD] = Instruction{
		Opecode: 0xDD,
		Code: CMP,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.cmp,
	}

	instructionSet[0xD9] = Instruction{
		Opecode: 0xD9,
		Code: CMP,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.cmp,
	}

	instructionSet[0xC1] = Instruction{
		Opecode: 0xC1,
		Code: CMP,
		AddressingMode: IndirectXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.cmp,
	}

	instructionSet[0xD1] = Instruction{
		Opecode: 0xD1,
		Code: CMP,
		AddressingMode: IndirectYIndexed,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 1,
		Handler: c.cmp,
	}


	// MARK: CPX命令
	instructionSet[0xE0] = Instruction{
		Opecode: 0xE0,
		Code: CPX,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.cpx,
	}

	instructionSet[0xE4] = Instruction{
		Opecode: 0xE4,
		Code: CPX,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.cpx,
	}

	instructionSet[0xEC] = Instruction{
		Opecode: 0xEC,
		Code: CPX,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.cpx,
	}


	// MARK: CPY命令
	instructionSet[0xC0] = Instruction{
		Opecode: 0xC0,
		Code: CPY,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.cpy,
	}

	instructionSet[0xC4] = Instruction{
		Opecode: 0xC4,
		Code: CPY,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.cpy,
	}

	instructionSet[0xCC] = Instruction{
		Opecode: 0xCC,
		Code: CPY,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.cpy,
	}


	// MARK: DEC命令
	instructionSet[0xC6] = Instruction{
		Opecode: 0xC6,
		Code: DEC,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.dec,
	}

	instructionSet[0xD6] = Instruction{
		Opecode: 0xD6,
		Code: DEC,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.dec,
	}

	instructionSet[0xCE] = Instruction{
		Opecode: 0xCE,
		Code: DEC,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.dec,
	}

	instructionSet[0xDE] = Instruction{
		Opecode: 0xDE,
		Code: DEC,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 7,
		PageCycles: 0,
		Handler: c.dec,
	}


	// MARK: DEX命令
	instructionSet[0xCA] = Instruction{
		Opecode: 0xCA,
		Code: DEX,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.dex,
	}


	// MARK: DEY命令
	instructionSet[0x88] = Instruction{
		Opecode: 0x88,
		Code: DEY,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.dey,
	}


	// MARK: EOR命令
	instructionSet[0x49] = Instruction{
		Opecode: 0x49,
		Code: EOR,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.eor,
	}

	instructionSet[0x45] = Instruction{
		Opecode: 0x45,
		Code: EOR,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.eor,
	}

	instructionSet[0x55] = Instruction{
		Opecode: 0x55,
		Code: EOR,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.eor,
	}

	instructionSet[0x4D] = Instruction{
		Opecode: 0x4D,
		Code: EOR,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.eor,
	}

	instructionSet[0x5D] = Instruction{
		Opecode: 0x5D,
		Code: EOR,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.eor,
	}

	instructionSet[0x59] = Instruction{
		Opecode: 0x59,
		Code: EOR,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.eor,
	}

	instructionSet[0x41] = Instruction{
		Opecode: 0x41,
		Code: EOR,
		AddressingMode: IndirectXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.eor,
	}

	instructionSet[0x51] = Instruction{
		Opecode: 0x51,
		Code: EOR,
		AddressingMode: IndirectYIndexed,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 1,
		Handler: c.eor,
	}


	// MARK: INC命令
	instructionSet[0xE6] = Instruction{
		Opecode: 0xE6,
		Code: INC,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.inc,
	}

	instructionSet[0xF6] = Instruction{
		Opecode: 0xF6,
		Code: INC,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.inc,
	}

	instructionSet[0xEE] = Instruction{
		Opecode: 0xEE,
		Code: INC,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.inc,
	}

	instructionSet[0xFE] = Instruction{
		Opecode: 0xFE,
		Code: INC,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 7,
		PageCycles: 0,
		Handler: c.inc,
	}


	// MARK: INX命令
	instructionSet[0xE8] = Instruction{
		Opecode: 0xE8,
		Code: INX,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.inx,
	}


	// MARK: INY命令
	instructionSet[0xC8] = Instruction{
		Opecode: 0xC8,
		Code: INY,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.iny,
	}


	// MARK: JMP命令
	instructionSet[0x4C] = Instruction{
		Opecode: 0x4C,
		Code: JMP,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.jmp,
	}

	instructionSet[0x6C] = Instruction{
		Opecode: 0x6C,
		Code: JMP,
		AddressingMode: Indirect,
		Bytes: 3,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.jmp,
	}


	// MARK: JSR命令
	instructionSet[0x20] = Instruction{
		Opecode: 0x20,
		Code: JSR,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.jsr,
	}


	// MARK: LDA命令
	instructionSet[0xA9] = Instruction{
		Opecode: 0xA9,
		Code: LDA,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.lda,
	}

	instructionSet[0xA5] = Instruction{
		Opecode: 0xA5,
		Code: LDA,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.lda,
	}

	instructionSet[0xB5] = Instruction{
		Opecode: 0xB5,
		Code: LDA,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.lda,
	}

	instructionSet[0xAD] = Instruction{
		Opecode: 0xAD,
		Code: LDA,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.lda,
	}

	instructionSet[0xBD] = Instruction{
		Opecode: 0xBD,
		Code: LDA,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.lda,
	}

	instructionSet[0xB9] = Instruction{
		Opecode: 0xB9,
		Code: LDA,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.lda,
	}

	instructionSet[0xA1] = Instruction{
		Opecode: 0xA1,
		Code: LDA,
		AddressingMode: IndirectXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.lda,
	}

	instructionSet[0xB1] = Instruction{
		Opecode: 0xB1,
		Code: LDA,
		AddressingMode: IndirectYIndexed,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 1,
		Handler: c.lda,
	}


	// MARK: LDX命令
	instructionSet[0xA2] = Instruction{
		Opecode: 0xA2,
		Code: LDX,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.ldx,
	}

	instructionSet[0xA6] = Instruction{
		Opecode: 0xA6,
		Code: LDX,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.ldx,
	}

	instructionSet[0xB6] = Instruction{
		Opecode: 0xB6,
		Code: LDX,
		AddressingMode: ZeroPageYIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.ldx,
	}

	instructionSet[0xAE] = Instruction{
		Opecode: 0xAE,
		Code: LDX,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.ldx,
	}

	instructionSet[0xBE] = Instruction{
		Opecode: 0xBE,
		Code: LDX,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.ldx,
	}


	// MARK: LDY命令
	instructionSet[0xA0] = Instruction{
		Opecode: 0xA0,
		Code: LDY,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.ldy,
	}

	instructionSet[0xA4] = Instruction{
		Opecode: 0xA4,
		Code: LDY,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.ldy,
	}

	instructionSet[0xB4] = Instruction{
		Opecode: 0xB4,
		Code: LDY,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.ldy,
	}

	instructionSet[0xAC] = Instruction{
		Opecode: 0xAC,
		Code: LDY,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.ldy,
	}

	instructionSet[0xBC] = Instruction{
		Opecode: 0xBC,
		Code: LDY,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.ldy,
	}


	// MARK: LSR命令
	instructionSet[0x4A] = Instruction{
		Opecode: 0x4A,
		Code: LSR,
		AddressingMode: Accumulator,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.lsr,
	}

	instructionSet[0x46] = Instruction{
		Opecode: 0x46,
		Code: LSR,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.lsr,
	}

	instructionSet[0x56] = Instruction{
		Opecode: 0x56,
		Code: LSR,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.lsr,
	}

	instructionSet[0x4E] = Instruction{
		Opecode: 0x4E,
		Code: LSR,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.lsr,
	}

	instructionSet[0x5E] = Instruction{
		Opecode: 0x5E,
		Code: LSR,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 7,
		PageCycles: 0,
		Handler: c.lsr,
	}


	// MARK: NOP命令
	instructionSet[0xEA] = Instruction{
		Opecode: 0xEA,
		Code: NOP,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.nop,
	}


	// MARK: ORA命令
	instructionSet[0x09] = Instruction{
		Opecode: 0x09,
		Code: ORA,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.ora,
	}

	instructionSet[0x05] = Instruction{
		Opecode: 0x05,
		Code: ORA,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.ora,
	}

	instructionSet[0x15] = Instruction{
		Opecode: 0x15,
		Code: ORA,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.ora,
	}

	instructionSet[0x0D] = Instruction{
		Opecode: 0x0D,
		Code: ORA,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.ora,
	}

	instructionSet[0x1D] = Instruction{
		Opecode: 0x1D,
		Code: ORA,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.ora,
	}

	instructionSet[0x19] = Instruction{
		Opecode: 0x19,
		Code: ORA,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.ora,
	}

	instructionSet[0x01] = Instruction{
		Opecode: 0x01,
		Code: ORA,
		AddressingMode: IndirectXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.ora,
	}

	instructionSet[0x11] = Instruction{
		Opecode: 0x11,
		Code: ORA,
		AddressingMode: IndirectYIndexed,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 1,
		Handler: c.ora,
	}


	// MARK: PHA命令
	instructionSet[0x48] = Instruction{
		Opecode: 0x48,
		Code: PHA,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.pha,
	}


	// MARK: PHP命令
	instructionSet[0x08] = Instruction{
		Opecode: 0x08,
		Code: PHP,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.php,
	}


	// MARK: PLA命令
	instructionSet[0x68] = Instruction{
		Opecode: 0x68,
		Code: PLA,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.pla,
	}


	// MARK: PLP命令
	instructionSet[0x28] = Instruction{
		Opecode: 0x28,
		Code: PLP,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.plp,
	}


	// MARK: ROL命令
	instructionSet[0x2A] = Instruction{
		Opecode: 0x2A,
		Code: ROL,
		AddressingMode: Accumulator,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.rol,
	}

	instructionSet[0x26] = Instruction{
		Opecode: 0x26,
		Code: ROL,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.rol,
	}

	instructionSet[0x36] = Instruction{
		Opecode: 0x36,
		Code: ROL,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.rol,
	}

	instructionSet[0x2E] = Instruction{
		Opecode: 0x2E,
		Code: ROL,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.rol,
	}

	instructionSet[0x3E] = Instruction{
		Opecode: 0x3E,
		Code: ROL,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 7,
		PageCycles: 0,
		Handler: c.rol,
	}


	// MARK: ROR命令
	instructionSet[0x6A] = Instruction{
		Opecode: 0x6A,
		Code: ROR,
		AddressingMode: Accumulator,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.ror,
	}

	instructionSet[0x66] = Instruction{
		Opecode: 0x66,
		Code: ROR,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.ror,
	}

	instructionSet[0x76] = Instruction{
		Opecode: 0x76,
		Code: ROR,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.ror,
	}

	instructionSet[0x6E] = Instruction{
		Opecode: 0x6E,
		Code: ROR,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.ror,
	}

	instructionSet[0x7E] = Instruction{
		Opecode: 0x7E,
		Code: ROR,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 7,
		PageCycles: 0,
		Handler: c.ror,
	}


	// MARK: RTI命令
	instructionSet[0x40] = Instruction{
		Opecode: 0x40,
		Code: RTI,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.rti,
	}


	// MARK: RTS命令
	instructionSet[0x60] = Instruction{
		Opecode: 0x60,
		Code: RTS,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.rts,
	}


	// MARK: SBC命令
	instructionSet[0xE9] = Instruction{
		Opecode: 0xE9,
		Code: SBC,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.sbc,
	}

	instructionSet[0xE5] = Instruction{
		Opecode: 0xE5,
		Code: SBC,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.sbc,
	}

	instructionSet[0xF5] = Instruction{
		Opecode: 0xF5,
		Code: SBC,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.sbc,
	}

	instructionSet[0xED] = Instruction{
		Opecode: 0xED,
		Code: SBC,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.sbc,
	}

	instructionSet[0xFD] = Instruction{
		Opecode: 0xE9,
		Code: SBC,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.sbc,
	}

	instructionSet[0xF9] = Instruction{
		Opecode: 0xF9,
		Code: SBC,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 1,
		Handler: c.sbc,
	}

	instructionSet[0xE1] = Instruction{
		Opecode: 0xE1,
		Code: SBC,
		AddressingMode: IndirectXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.sbc,
	}

	instructionSet[0xF1] = Instruction{
		Opecode: 0xF1,
		Code: SBC,
		AddressingMode: IndirectYIndexed,
		Bytes: 2,
		Cycles: 5,
		PageCycles: 1,
		Handler: c.sbc,
	}


	// MARK: SEC命令
	instructionSet[0x38] = Instruction{
		Opecode: 0x38,
		Code: SEC,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.sec,
	}


	// MARK: SED命令
	instructionSet[0xF8] = Instruction{
		Opecode: 0xF8,
		Code: SED,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.sed,
	}


	// MARK: SEI命令
	instructionSet[0x78] = Instruction{
		Opecode: 0x78,
		Code: SEI,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.sei,
	}


	// MARK: STA命令
	instructionSet[0x85] = Instruction{
		Opecode: 0x85,
		Code: STA,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.sta,
	}

	instructionSet[0x95] = Instruction{
		Opecode: 0x95,
		Code: STA,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.sta,
	}

	instructionSet[0x8D] = Instruction{
		Opecode: 0x8D,
		Code: STA,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.sta,
	}

	instructionSet[0x9D] = Instruction{
		Opecode: 0x9D,
		Code: STA,
		AddressingMode: AbsoluteXIndexed,
		Bytes: 3,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.sta,
	}

	instructionSet[0x99] = Instruction{
		Opecode: 0x99,
		Code: STA,
		AddressingMode: AbsoluteYIndexed,
		Bytes: 3,
		Cycles: 5,
		PageCycles: 0,
		Handler: c.sta,
	}

	instructionSet[0x81] = Instruction{
		Opecode: 0x81,
		Code: STA,
		AddressingMode: IndirectXIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.sta,
	}

	instructionSet[0x91] = Instruction{
		Opecode: 0x91,
		Code: STA,
		AddressingMode: IndirectYIndexed,
		Bytes: 2,
		Cycles: 6,
		PageCycles: 0,
		Handler: c.sta,
	}


	// MARK: STX命令
	instructionSet[0x86] = Instruction{
		Opecode: 0x86,
		Code: STX,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.stx,
	}

	instructionSet[0x96] = Instruction{
		Opecode: 0x96,
		Code: STX,
		AddressingMode: ZeroPageYIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.stx,
	}

	instructionSet[0x8E] = Instruction{
		Opecode: 0x8E,
		Code: STX,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.stx,
	}


	// MARK: STY命令
	instructionSet[0x84] = Instruction{
		Opecode: 0x84,
		Code: STY,
		AddressingMode: ZeroPage,
		Bytes: 2,
		Cycles: 3,
		PageCycles: 0,
		Handler: c.sty,
	}

	instructionSet[0x94] = Instruction{
		Opecode: 0x94,
		Code: STY,
		AddressingMode: ZeroPageXIndexed,
		Bytes: 2,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.sty,
	}

	instructionSet[0x8C] = Instruction{
		Opecode: 0x8C,
		Code: STY,
		AddressingMode: Absolute,
		Bytes: 3,
		Cycles: 4,
		PageCycles: 0,
		Handler: c.sty,
	}


	// MARK: TAX命令
	instructionSet[0xAA] = Instruction{
		Opecode: 0xAA,
		Code: TAX,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.tax,
	}


	// MARK: TAY命令
	instructionSet[0xA8] = Instruction{
		Opecode: 0xA8,
		Code: TAY,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.tay,
	}


	// MARK: TSX命令
	instructionSet[0xBA] = Instruction{
		Opecode: 0xBA,
		Code: TSX,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.tsx,
	}


	// MARK: TXA命令
	instructionSet[0x8A] = Instruction{
		Opecode: 0x8A,
		Code: TXA,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.txa,
	}


	// MARK: TXS命令
	instructionSet[0x9A] = Instruction{
		Opecode: 0x9A,
		Code: TXS,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.txs,
	}


	// MARK: TYA命令
	instructionSet[0x98] = Instruction{
		Opecode: 0x98,
		Code: TYA,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 2,
		PageCycles: 0,
		Handler: c.tya,
	}

	return instructionSet
}


// 命令名取得メソッド
func (ic InstructionCode) ToString() string {
	names := [...]string {
		"ADC", "AND", "ASL", "BCC", "BCS", "BEQ",	"BIT", "BMI","BNE", "BPL",	"BRK","BVC","BVS","CLC","CLD","CLI","CLV","CMP","CPX","CPY","DEC","DEX","DEY","EOR","INC","INX","INY","JMP","JSR","LDA","LDX","LDY","LSR","NOP","ORA","PHA","PHP","PLA","PLP","ROL","ROR","RTI","RTS","SBC","SEC","SED","SEI","STA","STX","STY","TAX","TAY","TSX","TXA","TXS","TYA",
	}

	if int(ic) < len(names) {
		return names[ic]
	} else {
		return fmt.Sprintf("Unknown(%d)", ic)
	}
}