package cpu

import (
	"fmt"
	"log"
)

const (
	CPU_WRAM_SIZE = 0xFFFF
)

// MARK: CPUの定義
type CPU struct {
	Registers registers // レジスタ
	InstructionSet instructionSet // 命令セット

	wram [CPU_WRAM_SIZE+1]uint8 // WRAM

	log bool // デバッグ出力フラグ
}

// MARK: CPUの作成関数
func CreateCPU(debug bool) *CPU {
	cpu := &CPU{}
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
func (c *CPU) Step() {
	c.execute(func() {})
}

func (c *CPU) execute(callback func()) {
	callback()

		opecode := c.ReadByteFromWRAM(c.Registers.PC)
	instruction, exists := c.InstructionSet[opecode]

	if !exists {
		log.Fatalf("Error: Unknown opecode $%02X at PC=%04X", opecode, c.Registers.PC)
	}

	if c.log {
		fmt.Printf("opecode: $%02X (%s) from: $%04X\n", opecode, instruction.Code.ToString(), c.Registers.PC)
	}

	// オペコード分プログラムカウンタを進める
	c.Registers.PC++
	instruction.Handler(instruction.AddressingMode)

	if instruction.Code != JMP && instruction.Code != JSR && instruction.Code != RTS && instruction.Code != RTI {
		// オペランド分プログラムカウンタを進める (オペコードの分 -1)
		c.Registers.PC += uint16(instruction.Bytes-1)
	}

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
	case Indirect:
		ptr := c.ReadWordFromWRAM(c.Registers.PC)
		// ページ境界をまたぐ際のバグを再現
		if (ptr & 0xFF) == 0xFF {
			lower := c.ReadByteFromWRAM(ptr)
			upper := c.ReadByteFromWRAM(ptr & 0xFF00)
			return uint16(upper) << 8 | uint16(lower)
		} else {
			return c.ReadWordFromWRAM(ptr)
		}
	case IndirectXIndexed:
		base := c.ReadByteFromWRAM(c.Registers.PC)
		ptr := base + c.Registers.X
		return c.ReadWordFromWRAM(uint16(ptr))
	case IndirectYIndexed:
		base := c.ReadByteFromWRAM(c.Registers.PC)
		return c.ReadWordFromWRAM(uint16(base)) + uint16(c.Registers.Y)
	case Relative:
		offset := int8(c.ReadByteFromWRAM(c.Registers.PC))
		return uint16(offset)
	case Accumulator:
		log.Fatalf("Error: Mode Accumulator doesn't take any operands")
		return 0x0000
	case Implied:
		log.Fatalf("Error: Mode Implied doesn't take any operands")
		return 0x0000
	default:
		log.Fatalf("Error: Unsupported addressing type '%v'", mode)
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



// MARK: スタック操作
func (c *CPU) pushByte(value uint8) {
	stack_addr := 0x0100 | uint16(c.Registers.SP)
	c.WriteByteToWRAM(stack_addr, value)
	c.Registers.SP--
}

func (c *CPU) pushWord(value uint16) {
	stack_addr := 0x0100 | uint16(c.Registers.SP)
	c.WriteByteToWRAM(stack_addr, (uint8(value >> 8)))
	c.Registers.SP--

	stack_addr = 0x0100 | uint16(c.Registers.SP)
	c.WriteByteToWRAM(stack_addr, (uint8(value & 0xFF)))
	c.Registers.SP--
}

func (c *CPU) popByte() uint8 {
	c.Registers.SP++
	stack_addr := 0x0100 | uint16(c.Registers.SP)
	value := c.ReadByteFromWRAM(stack_addr)
	return value
}

func (c *CPU) popWord() uint16 {
	c.Registers.SP++
	stack_addr := 0x0100 | uint16(c.Registers.SP)
	lower := c.ReadByteFromWRAM(stack_addr)

	c.Registers.SP++
	stack_addr = 0x0100 | uint16(c.Registers.SP)
	upper := c.ReadByteFromWRAM(stack_addr)

	return uint16(upper) << 8 | uint16(lower)
}


// MARK: ADC命令の実装
func (c *CPU) adc(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)
	sum := uint16(c.Registers.A) + uint16(value)

	if c.Registers.P.Carry {
		sum++
	}

	result := uint8(sum)

	// キャリーフラグの設定 (結果が8bitを超えるか)
	c.Registers.P.Carry = sum > 0xFF

	// 符号付きオーバーフローの検出
	// 両方の入力の符号が同じで結果の符号が異なる場合にオーバーフロー
	c.Registers.P.Overflow = ((c.Registers.A ^ value) & 0x80) == 0 && ((c.Registers.A ^ result) & 0x80) != 0

	c.updateNZFlags(result)
	c.Registers.A = result
}

// MARK: AND命令の実装
func (c *CPU) and(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)
	c.Registers.A &= value

	c.updateNZFlags(c.Registers.A)
}

// MARK: ASL命令の実装
func (c *CPU) asl(mode AddressingMode) {
	if mode == Accumulator {
		c.Registers.P.Carry = (c.Registers.A >> 7) != 0
		c.Registers.A = c.Registers.A << 1
		c.updateNZFlags(c.Registers.A)
	} else {
		addr := c.getOperandAddress(mode)
		value := c.ReadByteFromWRAM(addr)
		c.Registers.P.Carry = (value >> 7) != 0
		value <<= 1
		c.WriteByteToWRAM(addr, value)
		c.updateNZFlags(value)
	}
}

// MARK: BCC命令の実装
func (c *CPU) bcc(mode AddressingMode) {
	if !c.Registers.P.Carry {
		offset := c.getOperandAddress(mode)
		c.Registers.PC = uint16(int32(c.Registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
	}
}

// MARK: BCS命令の実装
func (c *CPU) bcs(mode AddressingMode) {
	if c.Registers.P.Carry {
		offset := c.getOperandAddress(mode)
		c.Registers.PC = uint16(int32(c.Registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
	}
}

// MARK: BEQ命令の実装
func (c *CPU) beq(mode AddressingMode) {
	if c.Registers.P.Zero {
		offset := c.getOperandAddress(mode)
		c.Registers.PC = uint16(int32(c.Registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
	}
}

// MARK: BIT命令の実装
func (c *CPU) bit(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)
	mask := c.Registers.A

	c.Registers.P.Zero = (value & mask) == 0x00
	c.Registers.P.Overflow = (value >> 6) != 0
	c.Registers.P.Negative = (value >> 7) != 0
}

// MARK: BMI命令の実装
func (c *CPU) bmi(mode AddressingMode) {
	if c.Registers.P.Negative {
		offset := c.getOperandAddress(mode)
		c.Registers.PC = uint16(int32(c.Registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
	}
}

// MARK: BNE命令の実装
func (c *CPU) bne(mode AddressingMode) {
	if !c.Registers.P.Zero {
		offset := c.getOperandAddress(mode)
		c.Registers.PC = uint16(int32(c.Registers.PC) + int32(offset))
	}
}

// MARK: BPL命令の実装
func (c *CPU) bpl(mode AddressingMode) {
	if !c.Registers.P.Negative {
		offset := c.getOperandAddress(mode)
		c.Registers.PC = uint16(int32(c.Registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
	}
}

// MARK: BRK命令の実装
func (c *CPU) brk(mode AddressingMode) {
	c.pushWord(c.Registers.PC + 1)
	c.Registers.P.Break = true
	c.pushByte(c.Registers.P.ToByte())
	c.Registers.PC = c.ReadWordFromWRAM(0xFFFE)
}

// MARK: BVC命令の実装
func (c *CPU) bvc(mode AddressingMode) {
	if !c.Registers.P.Overflow {
		offset := c.getOperandAddress(mode)
		c.Registers.PC = uint16(int32(c.Registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
	}
}

// MARK: BVS命令の実装
func (c *CPU) bvs(mode AddressingMode) {
	if c.Registers.P.Overflow {
		offset := c.getOperandAddress(mode)
		c.Registers.PC = uint16(int32(c.Registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
	}
}

// MARK: CLC命令の実装
func (c *CPU) clc(mode AddressingMode) {
	c.Registers.P.Carry = false
}

// MARK: CLD命令の実装
func (c *CPU) cld(mode AddressingMode) {
	c.Registers.P.Decimal = false
}

// MARK: CLI命令の実装
func (c *CPU) cli(mode AddressingMode) {
	c.Registers.P.Interrupt = false
}

// MARK: CLV命令の実装
func (c *CPU) clv(mode AddressingMode) {
	c.Registers.P.Overflow = false
}

// MARK: CMP命令の実装
func (c *CPU) cmp(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)

	c.Registers.P.Carry = c.Registers.A >= value
	c.updateNZFlags(c.Registers.A - value)
}

// MARK: CPX命令の実装
func (c *CPU) cpx(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)

	c.Registers.P.Carry = c.Registers.X >= value
	c.updateNZFlags(c.Registers.X - value)
}

// MARK: CPY命令の実装
func (c *CPU) cpy(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)

	c.Registers.P.Carry = c.Registers.Y >= value
	c.updateNZFlags(c.Registers.Y - value)
}

// MARK: DEC命令の実装
func (c *CPU) dec(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr) - 1
	c.WriteByteToWRAM(addr, value)
	c.updateNZFlags(value)
}

// MARK: DEX命令の実装
func (c *CPU) dex(mode AddressingMode) {
	c.Registers.X--
	c.updateNZFlags(c.Registers.X)
}

// MARK: DEY命令の実装
func (c *CPU) dey(mode AddressingMode) {
	c.Registers.Y--
	c.updateNZFlags(c.Registers.Y)
}

// MARK: EOR命令の実装
func (c *CPU) eor(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)
	c.Registers.A ^= value
	c.updateNZFlags(c.Registers.A)
}

// MARK: INC命令の実装
func (c *CPU) inc(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr) + 1
	c.WriteByteToWRAM(addr, value)
	c.updateNZFlags(value)
}

// MARK: INX命令の実装
func (c *CPU) inx(mode AddressingMode) {
	c.Registers.X++
	c.updateNZFlags(c.Registers.X)
}

// MARK: INY命令の実装
func (c *CPU) iny(mode AddressingMode) {
	c.Registers.Y++
	c.updateNZFlags(c.Registers.Y)
}

// MARK: JMP命令の実装
func (c *CPU) jmp(mode AddressingMode) {
	c.Registers.PC = c.getOperandAddress(mode)
}

// MARK: JSR命令の実装
func (c *CPU) jsr(mode AddressingMode) {
	c.pushWord(c.Registers.PC + 1)
	addr := c.getOperandAddress(mode)
	c.Registers.PC = addr
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
	addr := c.getOperandAddress(mode)
	operand := c.ReadByteFromWRAM(addr)

	c.Registers.X = uint8(operand)
	c.updateNZFlags(c.Registers.X)
}

// MARK: LDY命令の実装
func (c *CPU) ldy(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	operand := c.ReadByteFromWRAM(addr)

	c.Registers.Y = uint8(operand)
	c.updateNZFlags(c.Registers.Y)
}

// MARK: LSR命令の実装
func (c *CPU) lsr(mode AddressingMode) {
	if mode == Accumulator {
		c.Registers.P.Carry = (c.Registers.A & 0x01) != 0
		c.Registers.A = c.Registers.A >> 1
		c.updateNZFlags(c.Registers.A)
	} else {
		addr := c.getOperandAddress(mode)
		value := c.ReadByteFromWRAM(addr)
		c.Registers.P.Carry = (value & 0x01) != 0
		value >>= 1
		c.WriteByteToWRAM(addr, value)
		c.updateNZFlags(value)
	}
}

// MARK: NOP命令の実装
func (c *CPU) nop(mode AddressingMode) {
}

// MARK: ORA命令の実装
func (c *CPU) ora(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)

	c.Registers.A |= value
	c.updateNZFlags(c.Registers.A)
}

// MARK: PHA命令の実装
func (c *CPU) pha(mode AddressingMode) {
	c.pushByte(c.Registers.A)
}

// MARK: PHP命令の実装
func (c *CPU) php(mode AddressingMode) {
	c.pushByte(c.Registers.P.ToByte())
}

// MARK: PLA命令の実装
func (c *CPU) pla(mode AddressingMode) {
	c.Registers.A = c.popByte()
	c.updateNZFlags(c.Registers.A)
}

// MARK: PLP命令の実装
func (c *CPU) plp(mode AddressingMode) {
	c.Registers.P.SetFromByte(c.popByte())
}

// MARK: ROL命令の実装
func (c *CPU) rol(mode AddressingMode) {
	if mode == Accumulator {
		carry := c.Registers.A >> 7 != 0
		c.Registers.A = c.Registers.A << 1

		if (c.Registers.P.Carry) {
			c.Registers.A |= 0x01
		}

		c.Registers.P.Carry = carry
		c.updateNZFlags(c.Registers.A)
	} else {
		addr := c.getOperandAddress(mode)
		value := c.ReadByteFromWRAM(addr)

		carry := value >> 7 != 0
		value <<= 1

		if c.Registers.P.Carry {
			value |= 0x01
		}

		c.Registers.P.Carry = carry
		c.Registers.P.Negative = value >> 7 != 0
		c.updateNZFlags(value)

		c.WriteByteToWRAM(addr, value)
	}
}

// MARK: ROR命令の実装
func (c *CPU) ror(mode AddressingMode) {
	if mode == Accumulator {
		carry := c.Registers.A & 0x01 != 0
		c.Registers.A = c.Registers.A >> 1
		
		if (c.Registers.P.Carry) {
			c.Registers.A |= 1 << 7
		}

		c.Registers.P.Carry = carry
		c.updateNZFlags(c.Registers.A)
	} else {
		addr := c.getOperandAddress(mode)
		value := c.ReadByteFromWRAM(addr)

		carry := value & 0x01 != 0
		value >>= 1

		if c.Registers.P.Carry {
			value |= 1 << 7
		}

		c.Registers.P.Carry = carry
		c.Registers.P.Negative = value >> 7 != 0
		c.updateNZFlags(value)

		c.WriteByteToWRAM(addr, value)
	}
}

// MARK: RTI命令の実装
func (c *CPU) rti(mode AddressingMode) {
	status := c.popByte()
	addr := c.popWord()

	c.Registers.P.SetFromByte(status)
	c.Registers.PC = addr
	c.Registers.P.Break = false
}

// MARK: RTS命令の実装
func (c *CPU) rts(mode AddressingMode) {
	addr := c.popWord()
	c.Registers.PC = addr + 1
}

// MARK: SBC命令の実装
func (c *CPU) sbc(mode AddressingMode) {
	addr := c.getOperandAddress(mode)
	value := c.ReadByteFromWRAM(addr)

	sum := uint16(c.Registers.A) + uint16(^value)

	if c.Registers.P.Carry {
		sum++
	}

	result := uint8(sum)

	// フラグ設定
	c.Registers.P.Carry = sum > 0xFF
	c.Registers.P.Overflow = ((c.Registers.A ^ value) & 0x80) != 0 && ((c.Registers.A ^ result) & 0x80) != 0

	c.updateNZFlags(result)
	c.Registers.A = result
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


// MARK: デバッグ用表示メソッド
func (c *CPU) Dump() {
	fmt.Printf("REG_A: $%02X\n", c.Registers.A)
	fmt.Printf("REG_X: $%02X\n", c.Registers.X)
	fmt.Printf("REG_Y: $%02X\n", c.Registers.Y)
	fmt.Printf("REG_SP: $%02X\n", c.Registers.SP)
	fmt.Printf("REG_PC: $%04X\n", c.Registers.PC)
	fmt.Println("P.Negative: ", c.Registers.P.Negative)
	fmt.Println("P.Zero: ", c.Registers.P.Zero)
	fmt.Println("P.Carry: ", c.Registers.P.Carry)
	fmt.Printf("P.Overflow: %v\n\n", c.Registers.P.Overflow)
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
		c.Step()
	}
}