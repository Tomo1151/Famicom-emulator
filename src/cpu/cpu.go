package cpu

import "fmt"

// CPUの定義
type CPU struct {
	Registers registers // レジスタ
	InstructionSet instructionSet // 命令セット

	wram [0xFFFF]uint8 // WRAM

	log bool // デバッグ出力フラグ
}

// CPUの作成関数
func CreateCPU(debug bool) CPU {
	cpu := CPU{}
	cpu.Init(debug)
	return cpu
}

// CPUの初期化メソッド
func (c *CPU) Init(debug bool) {
	c.Registers = registers{
		A: 0x00,
		X: 0x00,
		Y: 0x00,
		P: statusRegister{
			Negative:  false,
			Overflow:  false,
			Reserved:  true,
			Break:     true,
			Decimal:   false,
			Interrupt: true,
			Zero:      false,
			Carry:     false,
		},
		SP: 0xFD,
		PC: 0x0000,
	}
	c.InstructionSet = generateInstructionSet(c)
	c.log = debug
}

// 命令の実行
func (c *CPU) Execute() {
	opecode := c.ReadWRAM(c.Registers.PC)
	instruction := c.InstructionSet[opecode]

	if c.log {
		fmt.Printf("opecode: $%02X (%s) from: $%04X\n", opecode, instruction.Code.ToString(), c.Registers.PC)
		fmt.Println("instr: ", instruction)
	}

	instruction.Handler(instruction.AddressingMode, 0)

	// プログラムカウンタを進める
	c.Registers.PC = uint16(c.Registers.PC + uint16(instruction.Bytes))

	if c.log {
		fmt.Printf("PC: $%04X\n\n", c.Registers.PC)
	}
}

// ワーキングメモリの参照
func (c *CPU) ReadWRAM(address uint16) uint8 {
	return c.wram[address]
}

// ワーキングメモリへの書き込み
func (c *CPU) WriteWRAM(address uint16, data uint8) {
	c.wram[address] = data
}

func (c *CPU) adc(addressing AddressingMode, operand uint16) {
	if c.log {
		fmt.Printf("*ADC* mode: $%02X operand: $%04X", addressing, operand)
	}
}