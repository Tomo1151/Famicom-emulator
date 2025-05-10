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


	// BRK命令
	instructionSet[0x00] = Instruction{
		Opecode: 0x00,
		Code: BRK,
		AddressingMode: Implied,
		Bytes: 1,
		Cycles: 7,
		PageCycles: 0,
		Handler: c.brk,
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