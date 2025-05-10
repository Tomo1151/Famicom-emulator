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


// MARK: フラグ(N, Z)の更新
func (c *CPU) updateNZFlags(result uint8) {
	// Nフラグの更新
	if (result >> 7) != 0  {
		c.Registers.P.Negative = true
	} else {
		c.Registers.P.Negative = false
	}

	// Zフラグの更新
	if result == 0 {
		c.Registers.P.Zero = true
		} else {
			c.Registers.P.Zero = false
		}
	}


	// MARK: フラグ(V, C)の更新
	func (c *CPU) updateVCFlags(prev uint8, result uint8) {
	// Vフラグの更新 @TODO 実装
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
	c.updateNZFlags(c.Registers.A)
}

// MARK: LDX命令の実装
func (c *CPU) ldx(mode AddressingMode) {
	// @TODO 実装
}

// MARK: LDY命令の実装
func (c *CPU) ldy(mode AddressingMode) {
	// @TODO 実装
}

// MARK: LSR命令の実装
func (c *CPU) lsr(mode AddressingMode) {
	// @TODO 実装
}

// MARK: NOP命令の実装
func (c *CPU) nop(mode AddressingMode) {
	// @TODO 実装
}

// MARK: ORA命令の実装
func (c *CPU) ora(mode AddressingMode) {
	// @TODO 実装
}

// MARK: PHA命令の実装
func (c *CPU) pha(mode AddressingMode) {
	// @TODO 実装
}

// MARK: PHP命令の実装
func (c *CPU) php(mode AddressingMode) {
	// @TODO 実装
}

// MARK: PLA命令の実装
func (c *CPU) pla(mode AddressingMode) {
	// @TODO 実装
}

// MARK: PLP命令の実装
func (c *CPU) plp(mode AddressingMode) {
	// @TODO 実装
}

// MARK: ROL命令の実装
func (c *CPU) rol(mode AddressingMode) {
	// @TODO 実装
}

// MARK: ROR命令の実装
func (c *CPU) ror(mode AddressingMode) {
	// @TODO 実装
}

// MARK: RTI命令の実装
func (c *CPU) rti(mode AddressingMode) {
	// @TODO 実装
}

// MARK: RTS命令の実装
func (c *CPU) rts(mode AddressingMode) {
	// @TODO 実装
}

// MARK: SBC命令の実装
func (c *CPU) sbc(mode AddressingMode) {
	// @TODO 実装
}

// MARK: SEC命令の実装
func (c *CPU) sec(mode AddressingMode) {
	c.Registers.P.Carry = true
}

// MARK: SED命令の実装
func (c *CPU) sed(mode AddressingMode) {
	c.Registers.P.Decimal = true
}

// MARK: SEI命令の実装
func (c *CPU) sei(mode AddressingMode) {
	c.Registers.P.Interrupt = true
}

// MARK: STA命令の実装
func (c *CPU) sta(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	c.WriteByteToWRAM(addr, c.Registers.A)
}

// MARK: STX命令の実装
func (c *CPU) stx(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	c.WriteByteToWRAM(addr, c.Registers.X)
}

// MARK: STY命令の実装
func (c *CPU) sty(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	c.WriteByteToWRAM(addr, c.Registers.Y)
}

// MARK: TAX命令の実装
func (c *CPU) tax(mode AddressingMode) {
	c.Registers.X = c.Registers.A
	c.updateNZFlags(c.Registers.X)
}

// MARK: TAY命令の実装
func (c *CPU) tay(mode AddressingMode) {
	c.Registers.Y = c.Registers.A
	c.updateNZFlags(c.Registers.Y)
}

// MARK: TSX命令の実装
func (c *CPU) tsx(mode AddressingMode) {
	c.Registers.X = c.Registers.SP
	c.updateNZFlags(c.Registers.X)
}

// MARK: TXA命令
func (c *CPU) txa(mode AddressingMode) {
	c.Registers.A = c.Registers.X
	c.updateNZFlags(c.Registers.A)
}

// MARK: TXS命令
func (c *CPU) txs(mode AddressingMode) {
	c.Registers.SP = c.Registers.X
}

// MARK: TYA命令
func (c *CPU) tya(mode AddressingMode) {
	c.Registers.A = c.Registers.Y
	c.updateNZFlags(c.Registers.A)
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