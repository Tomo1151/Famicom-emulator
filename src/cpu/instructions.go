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

type InstructionHandler func(addressing AddressingMode, operand uint16)

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

	// ADC命令
	instructionSet[0x69] = Instruction{
		Opecode: 0x69,
		Code: ADC,
		AddressingMode: Immediate,
		Bytes: 2,
		Cycles: 2,
		PageCycles: 1,
		Handler: c.adc,
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