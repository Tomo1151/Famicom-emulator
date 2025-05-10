package cpu

import (
	"fmt"
	"log"
)

// MARK: CPUの定義
type CPU struct {
	Registers registers // レジスタ
	InstructionSet instructionSet // 命令セット

	wram [0xFFFF]uint8 // WRAM

	log bool // デバッグ出力フラグ
}

// MARK: CPUの作成関数
func CreateCPU(debug bool) CPU {
	cpu := CPU{}
	cpu.Init(debug)
	return cpu
}

// MARK: CPUの初期化メソッド
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
		PC: c.ReadWordFromWRAM(0xFFFC),
	}
	c.InstructionSet = generateInstructionSet(c)
	c.log = debug
}

// MARK:  命令の実行
func (c *CPU) Execute() {
	opecode := c.ReadByteFromWRAM(c.Registers.PC)
	instruction := c.InstructionSet[opecode]

	if c.log {
		fmt.Printf("opecode: $%02X (%s) from: $%04X\n", opecode, instruction.Code.ToString(), c.Registers.PC)
	}

	// オペコード分プログラムカウンタを進める
	c.Registers.PC++
	instruction.Handler(instruction.AddressingMode)

	// オペランド分プログラムカウンタを進める (オペコードの分 -1)
	c.Registers.PC = uint16(c.Registers.PC + uint16(instruction.Bytes-1))

	if c.log {
		fmt.Printf("PC: $%04X\n\n", c.Registers.PC)
	}
}

// MARK: ワーキングメモリの参照 (1byte)
func (c *CPU) ReadByteFromWRAM(address uint16) uint8 {
	return c.wram[address]
}

// MARK: ワーキングメモリの参照 (2byte)
func (c *CPU) ReadWordFromWRAM(address uint16) uint16 {
	lower := c.ReadByteFromWRAM(address)
	upper := c.ReadByteFromWRAM(address + 1)

	return uint16(upper) << 8 | uint16((lower))
}

// MARK: ワーキングメモリへの書き込み (1byte)
func (c *CPU) WriteByteToWRAM(address uint16, data uint8) {
	c.wram[address] = data
}

// MARK: ワーキングメモリへの書き込み (2byte)
func (c *CPU) WriteWordToWRAM(address uint16, data uint16) {
	upper := uint8(data >> 8)
	lower := uint8(data & 0xFF)
	c.WriteByteToWRAM(address, lower)
	c.WriteByteToWRAM(address + 1, upper)
}


// MARK: アドレッシングモードからオペランドアドレスを計算
func (c *CPU) getOperandAddress(mode AddressingMode) uint16 {
	switch mode {
	case Immediate:
		return c.Registers.PC
	case ZeroPage:
		return uint16(c.ReadByteFromWRAM(c.Registers.PC))
	case Absolute:
		return c.ReadWordFromWRAM(c.Registers.PC)
	case ZeroPageXIndexed:
		origin := c.ReadByteFromWRAM(c.Registers.PC)
		return uint16(origin + c.Registers.X)
	case ZeroPageYIndexed:
		origin := c.ReadByteFromWRAM(c.Registers.PC)
		return uint16(origin + c.Registers.Y)
	case AbsoluteXIndexed:
		origin := c.ReadWordFromWRAM(c.Registers.PC)
		return origin + uint16(c.Registers.X)
	case AbsoluteYIndexed:
		origin := c.ReadWordFromWRAM(c.Registers.PC)
		return origin + uint16(c.Registers.Y)
	case IndirectXIndexed:
		base := c.ReadByteFromWRAM(c.Registers.PC)
		ptr := base + c.Registers.X
		lower := c.ReadByteFromWRAM(uint16(ptr))
		upper := c.ReadByteFromWRAM(uint16(ptr + 1))
		return uint16(upper) << 8 | uint16(lower)
	case IndirectYIndexed:
		base := c.ReadByteFromWRAM(c.Registers.PC)
		ptr := base + c.Registers.Y
		lower := c.ReadByteFromWRAM(uint16(ptr))
		upper := c.ReadByteFromWRAM(uint16(ptr + 1))
		return uint16(upper) << 8 | uint16(lower)
	default:
		log.Fatalf("Unsupported addressing type: %v", mode)
		return 0x0000
	}
}


// MARK: 計算フラグ(N, V, Z, C)の更新
func (c *CPU) updateCalcFlags(result uint8) {
	// Nフラグの更新
	if (result >> 7) != 0  {
		c.Registers.P.Negative = true
	} else {
		c.Registers.P.Negative = false
	}

	// Vフラグの更新 @TODO 実装

	// Zフラグの更新
	if result == 0 {
		c.Registers.P.Zero = true
	} else {
		c.Registers.P.Zero = false
	}

	// Cフラグの更新 @TODO 実装
}


// MARK: ADC命令の実装
func (c *CPU) adc(mode AddressingMode) {
	if c.log {
		fmt.Printf("*ADC* mode: $%02X", mode)
	}

	// @TODO 実装
}

// MARK: LDA命令の実装
func (c *CPU) lda(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	operand := c.ReadByteFromWRAM(addr)

	c.Registers.A = uint8(operand)
	c.updateCalcFlags(c.Registers.A)
}

// MARK: STA命令の実装
func (c *CPU) sta(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	c.WriteByteToWRAM(addr, c.Registers.A)
}

// MARK: TAX命令の実装
func (c *CPU) tax(mode AddressingMode) {
	c.Registers.X = c.Registers.A
	c.updateCalcFlags(c.Registers.X)
}

// MARK: BRK命令
func (c *CPU) brk(mode AddressingMode) {
}


// MARK: デバッグ用実行メソッド
func (c *CPU) REPL(commands []uint8) {
	c.Init(true)

	for addr, opecode := range commands {
		c.WriteByteToWRAM(uint16(addr), opecode)
	}

	for i := range commands {
		// BRK命令まで実行
		if commands[i] == 0x00 {
			return
		}
		c.Execute()
	}
}