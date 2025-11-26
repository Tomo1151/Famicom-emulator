package cpu

import (
	"fmt"
	"log"

	"Famicom-emulator/bus"
)

// MARK: CPUの定義
type CPU struct {
	registers      registers      // レジスタ
	InstructionSet instructionSet // 命令セット

	bus bus.Bus
	log bool // デバッグ出力フラグ
}

// MARK: CPUの初期化メソッド (カートリッジ無し，デバッグ・テスト用)
func (c *CPU) InitForTest(debug bool) {
	c.registers = registers{
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
		// PC: c.ReadWordFrom(0xFFFC),
	}
	c.bus = bus.Bus{}
	c.bus.InitForTest()
	c.InstructionSet = generateInstructionSet(c)
	c.log = debug
}

// MARK: CPUの初期化メソッド (Bus有り)
func (c *CPU) Init(bus bus.Bus, debug bool) {
	c.bus = bus
	c.registers = registers{
		A: 0x00,
		X: 0x00,
		Y: 0x00,
		P: statusRegister{
			Negative:  false,
			Overflow:  false,
			Reserved:  true,
			Break:     false,
			Decimal:   false,
			Interrupt: true,
			Zero:      false,
			Carry:     false,
		},
		SP: 0xFD,
		PC: c.ReadWordFrom(0xFFFC),
	}
	c.InstructionSet = generateInstructionSet(c)
	c.log = debug
}

// MARK:  命令の実行
func (c *CPU) Step() {
	// 命令のフェッチ
	opecode := c.ReadByteFrom(c.registers.PC)

	// 命令のデコード
	instruction, exists := c.InstructionSet[opecode]

	if !exists {
		log.Fatalf("Error: Unknown opecode $%02X at PC=%04X", opecode, c.registers.PC)
	}

	_, isPageCrossed := c.calcOperandAddress(instruction.AddressingMode)
	instruction.Handler(instruction.AddressingMode)
	if isPageCrossed {
		c.bus.Tick(uint(1))
	}

	if !instruction.Jump {
		// オペランド分プログラムカウンタを進める (オペコードの分 -1)
		c.registers.PC += uint16(instruction.Bytes)
	}

	c.bus.Tick(uint(instruction.Cycles))
}

// MARK: ループ実行
func (c *CPU) Run() {
	c.RunWithCallback(func(c *CPU) {
		if c.log {
			fmt.Println(c.Trace())
		}
	})
}

func (c *CPU) RunWithCallback(callback func(c *CPU)) {
	for {
		// NMIが発生したら処理をする
		nmi := c.bus.NMI()
		if nmi {
			c.interrupt(NMI)
		}

		apuIrq := c.bus.APUIRQ()
		mapperIrq := c.bus.MapperIRQ()

		// APUまたはマッパーでIRQが発生していて割込み禁止フラグが立っていないならIRQを処理
		if !c.registers.P.Interrupt && (apuIrq || mapperIrq) {
			c.interrupt(IRQ)
		}

		// コールバックを実行
		callback(c)

		// CPUの処理を進める
		c.Step()
	}
}

// MARK: NMIのハンドリング
func (c *CPU) interrupt(interrupt Interrupt) {
	// 現在のPCを退避
	c.pushWord(c.registers.PC)

	// ステータスレジスタをスタックにプッシュ
	status := c.registers.P
	status.Break = interrupt.BFlagMask&0b0001_0000 == 1
	status.Reserved = interrupt.BFlagMask&0b0010_0000 == 1
	c.pushByte(status.ToByte())
	c.registers.P.Interrupt = true

	c.bus.Tick(uint(interrupt.CPUCycles))
	c.registers.PC = c.ReadWordFrom(interrupt.VectorAddress) // 割り込みベクタ
}

// MARK: ワーキングメモリの参照 (1byte)
func (c *CPU) ReadByteFrom(address uint16) uint8 {
	return c.bus.ReadByteFrom(address)
}

// MARK: ワーキングメモリの参照 (2byte)
func (c *CPU) ReadWordFrom(address uint16) uint16 {
	return c.bus.ReadWordFrom(address)
}

// MARK: ワーキングメモリへの書き込み (1byte)
func (c *CPU) WriteByteAt(address uint16, data uint8) {
	c.bus.WriteByteAt(address, data)
}

// MARK: ワーキングメモリへの書き込み (2byte)
func (c *CPU) WriteWordAt(address uint16, data uint16) {
	c.bus.WriteWordAt(address, data)
}

func (c *CPU) isPageCrossed(address1 uint16, address2 uint16) bool {
	cond := (address1 & 0xFF00) != (address2 & 0xFF00)
	// fmt.Println("page crossed")
	return cond
}

// MARK: アドレッシングモードからオペランドアドレスを計算
func (c *CPU) calcOperandAddress(mode AddressingMode) (uint16, bool) {
	switch mode {
	case Immediate:
		return c.registers.PC + 1, false
	case ZeroPage:
		return uint16(c.ReadByteFrom(c.registers.PC + 1)), false
	case Absolute:
		return c.ReadWordFrom(c.registers.PC + 1), false
	case ZeroPageXIndexed:
		base := c.ReadByteFrom(c.registers.PC + 1)
		return uint16(base + c.registers.X), false
	case ZeroPageYIndexed:
		base := c.ReadByteFrom(c.registers.PC + 1)
		return uint16(base + c.registers.Y), false
	case AbsoluteXIndexed:
		base := c.ReadWordFrom(c.registers.PC + 1)
		ptr := base + uint16(c.registers.X)
		return ptr, c.isPageCrossed(base, ptr)
	case AbsoluteYIndexed:
		base := c.ReadWordFrom(c.registers.PC + 1)
		ptr := base + uint16(c.registers.Y)
		return ptr, c.isPageCrossed(base, ptr)
	case Indirect:
		ptr := c.ReadWordFrom(c.registers.PC + 1)
		// ページ境界をまたぐ際のバグを再現
		if (ptr & 0xFF) == 0xFF {
			lower := c.ReadByteFrom(ptr)
			upper := c.ReadByteFrom(ptr & 0xFF00)
			return uint16(upper)<<8 | uint16(lower), false
		} else {
			return c.ReadWordFrom(ptr), false
		}
	case IndirectXIndexed:
		base := c.ReadByteFrom(c.registers.PC + 1)
		ptr := uint8(base + c.registers.X)
		lower := c.ReadByteFrom(uint16(ptr))
		upper := c.ReadByteFrom(uint16(ptr+1) & 0xFF)
		return uint16(upper)<<8 | uint16(lower), false
	case IndirectYIndexed:
		base := c.ReadByteFrom(c.registers.PC + 1)
		ptr := uint8(base)
		lower := c.ReadByteFrom(uint16(ptr))
		upper := c.ReadByteFrom(uint16(ptr+1) & 0xFF)
		derefBase := uint16(upper)<<8 | uint16(lower)
		deref := derefBase + uint16(c.registers.Y)
		return deref, c.isPageCrossed(deref, derefBase)
	case Relative:
		offset := int8(c.ReadByteFrom(c.registers.PC + 1))
		return uint16(offset), false
	case Accumulator:
		// log.Fatalf("Error: Mode Accumulator doesn't take any operands")
		return 0x0000, false
	case Implied:
		// log.Fatalf("Error: Mode Implied doesn't take any operands")
		return 0x0000, false
	default:
		// log.Fatalf("Error: Unsupported addressing type '%v'", mode)
		return 0x0000, false
	}
}

// MARK: フラグ(N, Z)の更新
func (c *CPU) updateNZFlags(result uint8) {
	// Nフラグの更新
	if (result >> 7) != 0 {
		c.registers.P.Negative = true
	} else {
		c.registers.P.Negative = false
	}

	// Zフラグの更新
	if result == 0 {
		c.registers.P.Zero = true
	} else {
		c.registers.P.Zero = false
	}
}

// MARK: スタック操作
func (c *CPU) pushByte(value uint8) {
	stack_addr := 0x0100 | uint16(c.registers.SP)
	c.WriteByteAt(stack_addr, value)
	c.registers.SP--
}

func (c *CPU) pushWord(value uint16) {
	stack_addr := 0x0100 | uint16(c.registers.SP)
	c.WriteByteAt(stack_addr, (uint8(value >> 8)))
	c.registers.SP--

	stack_addr = 0x0100 | uint16(c.registers.SP)
	c.WriteByteAt(stack_addr, (uint8(value & 0xFF)))
	c.registers.SP--
}

func (c *CPU) popByte() uint8 {
	c.registers.SP++
	stack_addr := 0x0100 | uint16(c.registers.SP)
	value := c.ReadByteFrom(stack_addr)
	return value
}

func (c *CPU) popWord() uint16 {
	c.registers.SP++
	stack_addr := 0x0100 | uint16(c.registers.SP)
	lower := c.ReadByteFrom(stack_addr)

	c.registers.SP++
	stack_addr = 0x0100 | uint16(c.registers.SP)
	upper := c.ReadByteFrom(stack_addr)

	return uint16(upper)<<8 | uint16(lower)
}

// MARK: AAC命令の実装
func (c *CPU) aac(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	c.registers.A &= value

	c.updateNZFlags(c.registers.A)
	c.registers.P.Carry = c.registers.P.Negative
}

// MARK: AAX命令の実装
func (c *CPU) aax(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	result := c.registers.X & c.registers.A

	c.WriteByteAt(addr, result)
}

// MARK: ADC命令の実装
func (c *CPU) adc(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	sum := uint16(c.registers.A) + uint16(value)

	if c.registers.P.Carry {
		sum++
	}

	result := uint8(sum)

	// キャリーフラグの設定 (結果が8bitを超えるか)
	c.registers.P.Carry = sum > 0xFF

	// 符号付きオーバーフローの検出
	// 両方の入力の符号が同じで結果の符号が異なる場合にオーバーフロー
	c.registers.P.Overflow = ((c.registers.A^value)&0x80) == 0 && ((c.registers.A^result)&0x80) != 0

	c.updateNZFlags(result)
	c.registers.A = result
}

// MARK: AND命令の実装
func (c *CPU) and(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	c.registers.A &= value

	c.updateNZFlags(c.registers.A)
}

// MARK: ARR命令の実装
func (c *CPU) arr(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	c.registers.A &= value

	// 1ビット右回転
	c.registers.A = c.registers.A >> 1

	if c.registers.P.Carry {
		c.registers.A |= 1 << 7
	}

	c.registers.P.Carry = c.registers.A>>6 != 0
	c.registers.P.Overflow = (c.registers.A >> 6) != (c.registers.A >> 5) // XOR
	c.updateNZFlags(c.registers.A)

}

// MARK: ASL命令の実装
func (c *CPU) asl(mode AddressingMode) {
	if mode == Accumulator {
		c.registers.P.Carry = (c.registers.A >> 7) != 0
		c.registers.A = c.registers.A << 1
		c.updateNZFlags(c.registers.A)
	} else {
		addr, _ := c.calcOperandAddress(mode)
		value := c.ReadByteFrom(addr)
		c.registers.P.Carry = (value >> 7) != 0
		value <<= 1
		c.WriteByteAt(addr, value)
		c.updateNZFlags(value)
	}
}

// MARK: ASR命令の実装
func (c *CPU) asr(mode AddressingMode) {
	c.and(mode)
	c.lsr(Accumulator)
}

// MARK: ATX命令の実装
func (c *CPU) atx(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	c.registers.A &= value
	c.registers.X = c.registers.A
	c.updateNZFlags(c.registers.X)
}

// MARK: AXA命令の実装
func (c *CPU) axa(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	result := (c.registers.X & c.registers.A) & 7
	c.WriteByteAt(addr, result)
}

// MARK: AXS命令の実装
func (c *CPU) axs(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	c.registers.X &= c.registers.A

	c.registers.P.Carry = c.registers.X >= value
	c.registers.X -= value

	c.updateNZFlags(c.registers.X)
}

// MARK: BCC命令の実装
func (c *CPU) bcc(mode AddressingMode) {
	if !c.registers.P.Carry {
		c.bus.Tick(1)
		offset, _ := c.calcOperandAddress(mode)
		jumpAddr := uint16(int32(c.registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
		if c.isPageCrossed(c.registers.PC, jumpAddr) {
			c.bus.Tick(1)
		}
		c.registers.PC = jumpAddr
	}
}

// MARK: BCS命令の実装
func (c *CPU) bcs(mode AddressingMode) {
	if c.registers.P.Carry {
		c.bus.Tick(1)
		offset, _ := c.calcOperandAddress(mode)
		jumpAddr := uint16(int32(c.registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
		if c.isPageCrossed(c.registers.PC, jumpAddr) {
			c.bus.Tick(1)
		}
		c.registers.PC = jumpAddr
	}
}

// MARK: BEQ命令の実装
func (c *CPU) beq(mode AddressingMode) {
	if c.registers.P.Zero {
		c.bus.Tick(1)
		offset, _ := c.calcOperandAddress(mode)
		jumpAddr := uint16(int32(c.registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
		if c.isPageCrossed(c.registers.PC, jumpAddr) {
			c.bus.Tick(1)
		}
		c.registers.PC = jumpAddr
	}
}

// MARK: BIT命令の実装
func (c *CPU) bit(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	mask := c.registers.A

	c.registers.P.Zero = (value & mask) == 0x00
	c.registers.P.Overflow = (value & 0b0100_0000) != 0
	c.registers.P.Negative = (value & 0b1000_0000) != 0
}

// MARK: BMI命令の実装
func (c *CPU) bmi(mode AddressingMode) {
	if c.registers.P.Negative {
		c.bus.Tick(1)
		offset, _ := c.calcOperandAddress(mode)
		jumpAddr := uint16(int32(c.registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
		if c.isPageCrossed(c.registers.PC, jumpAddr) {
			c.bus.Tick(1)
		}
		c.registers.PC = jumpAddr
	}
}

// MARK: BNE命令の実装
func (c *CPU) bne(mode AddressingMode) {
	if !c.registers.P.Zero {
		c.bus.Tick(1)
		offset, _ := c.calcOperandAddress(mode)
		jumpAddr := uint16(int32(c.registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
		if c.isPageCrossed(c.registers.PC, jumpAddr) {
			c.bus.Tick(1)
		}
		c.registers.PC = jumpAddr
	}
}

// MARK: BPL命令の実装
func (c *CPU) bpl(mode AddressingMode) {
	if !c.registers.P.Negative {
		c.bus.Tick(1)
		offset, _ := c.calcOperandAddress(mode)
		jumpAddr := uint16(int32(c.registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
		if c.isPageCrossed(c.registers.PC, jumpAddr) {
			c.bus.Tick(1)
		}
		c.registers.PC = jumpAddr
	}
}

// MARK: BRK命令の実装
func (c *CPU) brk(mode AddressingMode) {
	if c.registers.P.Interrupt {
		return
	}

	c.pushWord(c.registers.PC + 1)
	c.registers.P.Break = true
	c.pushByte(c.registers.P.ToByte())
	c.registers.PC = c.ReadWordFrom(0xFFFE)
}

// MARK: BVC命令の実装
func (c *CPU) bvc(mode AddressingMode) {
	if !c.registers.P.Overflow {
		c.bus.Tick(1)
		offset, _ := c.calcOperandAddress(mode)
		jumpAddr := uint16(int32(c.registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
		if c.isPageCrossed(c.registers.PC, jumpAddr) {
			c.bus.Tick(1)
		}
		c.registers.PC = jumpAddr
	}
}

// MARK: BVS命令の実装
func (c *CPU) bvs(mode AddressingMode) {
	if c.registers.P.Overflow {
		c.bus.Tick(1)
		offset, _ := c.calcOperandAddress(mode)
		jumpAddr := uint16(int32(c.registers.PC) + int32(offset)) // 符号反転させなずに足すためint32を用いる
		if c.isPageCrossed(c.registers.PC, jumpAddr) {
			c.bus.Tick(1)
		}
		c.registers.PC = jumpAddr
	}
}

// MARK: CLC命令の実装
func (c *CPU) clc(mode AddressingMode) {
	c.registers.P.Carry = false
}

// MARK: CLD命令の実装
func (c *CPU) cld(mode AddressingMode) {
	c.registers.P.Decimal = false
}

// MARK: CLI命令の実装
func (c *CPU) cli(mode AddressingMode) {
	c.registers.P.Interrupt = false
}

// MARK: CLV命令の実装
func (c *CPU) clv(mode AddressingMode) {
	c.registers.P.Overflow = false
}

// MARK: CMP命令の実装
func (c *CPU) cmp(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)

	c.registers.P.Carry = c.registers.A >= value
	c.updateNZFlags(c.registers.A - value)
}

// MARK: CPX命令の実装
func (c *CPU) cpx(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)

	c.registers.P.Carry = c.registers.X >= value
	c.updateNZFlags(c.registers.X - value)
}

// MARK: CPY命令の実装
func (c *CPU) cpy(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)

	c.registers.P.Carry = c.registers.Y >= value
	c.updateNZFlags(c.registers.Y - value)
}

// MARK: DCP命令の実装
func (c *CPU) dcp(mode AddressingMode) {
	c.dec(mode)
	c.cmp(mode)
}

// MARK: DEC命令の実装
func (c *CPU) dec(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr) - 1
	c.WriteByteAt(addr, value)
	c.updateNZFlags(value)
}

// MARK: DEX命令の実装
func (c *CPU) dex(mode AddressingMode) {
	c.registers.X--
	c.updateNZFlags(c.registers.X)
}

// MARK: DEY命令の実装
func (c *CPU) dey(mode AddressingMode) {
	c.registers.Y--
	c.updateNZFlags(c.registers.Y)
}

// MARK: DOP命令の実装
func (c *CPU) dop(mode AddressingMode) {
}

// MARK: EOR命令の実装
func (c *CPU) eor(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	c.registers.A ^= value
	c.updateNZFlags(c.registers.A)
}

// MARK: INC命令の実装
func (c *CPU) inc(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr) + 1
	c.WriteByteAt(addr, value)
	c.updateNZFlags(value)
}

// MARK: INX命令の実装
func (c *CPU) inx(mode AddressingMode) {
	c.registers.X++
	c.updateNZFlags(c.registers.X)
}

// MARK: INY命令の実装
func (c *CPU) iny(mode AddressingMode) {
	c.registers.Y++
	c.updateNZFlags(c.registers.Y)
}

// MARK: ISC命令の実装
func (c *CPU) isc(mode AddressingMode) {
	c.inc(mode)
	c.sbc(mode)
}

// MARK: JMP命令の実装
func (c *CPU) jmp(mode AddressingMode) {
	c.registers.PC, _ = c.calcOperandAddress(mode)
}

// MARK: JSR命令の実装
func (c *CPU) jsr(mode AddressingMode) {
	c.pushWord(c.registers.PC + 2)
	addr, _ := c.calcOperandAddress(mode)
	c.registers.PC = addr
}

// MARK: KIL命令の実装
func (c *CPU) kil(mode AddressingMode) {
}

// MARK: LAR命令の実装
func (c *CPU) lar(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	result := c.registers.SP & value

	c.registers.A = result
	c.registers.X = result
	c.registers.SP = result
	c.updateNZFlags(result)
}

// MARK: LAX命令の実装
func (c *CPU) lax(mode AddressingMode) {
	c.lda(mode)
	c.tax(mode)
}

// MARK: LDA命令の実装
func (c *CPU) lda(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	operand := c.ReadByteFrom(addr)

	c.registers.A = uint8(operand)
	c.updateNZFlags(c.registers.A)
}

// MARK: LDX命令の実装
func (c *CPU) ldx(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	operand := c.ReadByteFrom(addr)

	c.registers.X = uint8(operand)
	c.updateNZFlags(c.registers.X)
}

// MARK: LDY命令の実装
func (c *CPU) ldy(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	operand := c.ReadByteFrom(addr)

	c.registers.Y = uint8(operand)
	c.updateNZFlags(c.registers.Y)
}

// MARK: LSR命令の実装
func (c *CPU) lsr(mode AddressingMode) {
	if mode == Accumulator {
		c.registers.P.Carry = (c.registers.A & 0x01) != 0
		c.registers.A = c.registers.A >> 1
		c.updateNZFlags(c.registers.A)
	} else {
		addr, _ := c.calcOperandAddress(mode)
		value := c.ReadByteFrom(addr)
		c.registers.P.Carry = (value & 0x01) != 0
		value >>= 1
		c.WriteByteAt(addr, value)
		c.updateNZFlags(value)
	}
}

// MARK: NOP命令の実装
func (c *CPU) nop(mode AddressingMode) {
}

// MARK: ORA命令の実装
func (c *CPU) ora(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)

	c.registers.A |= value
	c.updateNZFlags(c.registers.A)
}

// MARK: PHA命令の実装
func (c *CPU) pha(mode AddressingMode) {
	c.pushByte(c.registers.A)
}

// MARK: PHP命令の実装
func (c *CPU) php(mode AddressingMode) {
	// PHPでプッシュされるステータスレジスタはブレークフラグが立つ
	// 参考: https://pgate1.at-ninja.jp/NES_on_FPGA/nes_cpu.htm#trap
	c.pushByte(c.registers.P.ToByte() | 0x30)
}

// MARK: PLA命令の実装
func (c *CPU) pla(mode AddressingMode) {
	c.registers.A = c.popByte()
	c.updateNZFlags(c.registers.A)
}

// MARK: PLP命令の実装
func (c *CPU) plp(mode AddressingMode) {
	value := c.popByte()
	// PLPでフラグレジスタを復元するときには常にBreakはリセット, Reservedはセット?
	value = (value &^ 0x10) | 0x20
	c.registers.P.SetFromByte(value)
}

// MARK: ROL命令の実装
func (c *CPU) rol(mode AddressingMode) {
	if mode == Accumulator {
		carry := c.registers.A>>7 != 0
		c.registers.A = c.registers.A << 1

		if c.registers.P.Carry {
			c.registers.A |= 0x01
		}

		c.registers.P.Carry = carry
		c.updateNZFlags(c.registers.A)
	} else {
		addr, _ := c.calcOperandAddress(mode)
		value := c.ReadByteFrom(addr)

		carry := value>>7 != 0
		value <<= 1

		if c.registers.P.Carry {
			value |= 0x01
		}

		c.registers.P.Carry = carry
		c.registers.P.Negative = value>>7 != 0
		c.updateNZFlags(value)

		c.WriteByteAt(addr, value)
	}
}

// MARK: RLA命令の実装
func (c *CPU) rla(mode AddressingMode) {
	c.rol(mode)
	c.and(mode)
}

// MARK: ROR命令の実装
func (c *CPU) ror(mode AddressingMode) {
	if mode == Accumulator {
		carry := c.registers.A&0x01 != 0
		c.registers.A = c.registers.A >> 1

		if c.registers.P.Carry {
			c.registers.A |= 1 << 7
		}

		c.registers.P.Carry = carry
		c.updateNZFlags(c.registers.A)
	} else {
		addr, _ := c.calcOperandAddress(mode)
		value := c.ReadByteFrom(addr)

		carry := value&0x01 != 0
		value >>= 1

		if c.registers.P.Carry {
			value |= 1 << 7
		}

		c.registers.P.Carry = carry
		c.registers.P.Negative = value>>7 != 0
		c.updateNZFlags(value)

		c.WriteByteAt(addr, value)
	}
}

// MARK: RRA命令の実装
func (c *CPU) rra(mode AddressingMode) {
	c.ror(mode)
	c.adc(mode)
}

// MARK: RTI命令の実装
func (c *CPU) rti(mode AddressingMode) {
	status := c.popByte()
	addr := c.popWord()

	// RTIかにて復帰時には常にBreakはリセット, Reservedはセット？
	status = (status &^ 0x10) | 0x20
	c.registers.P.SetFromByte(status)
	c.registers.PC = addr
	c.registers.P.Break = false
}

// MARK: RTS命令の実装
func (c *CPU) rts(mode AddressingMode) {
	addr := c.popWord()
	c.registers.PC = addr + 1
}

// MARK: SBC命令の実装
func (c *CPU) sbc(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)

	sum := uint16(c.registers.A) + uint16(^value)

	if c.registers.P.Carry {
		sum++
	}

	result := uint8(sum)

	// フラグ設定
	c.registers.P.Carry = sum > 0xFF
	c.registers.P.Overflow = ((c.registers.A^value)&0x80) != 0 && ((c.registers.A^result)&0x80) != 0

	c.updateNZFlags(result)
	c.registers.A = result
}

// MARK: SEC命令の実装
func (c *CPU) sec(mode AddressingMode) {
	c.registers.P.Carry = true
}

// MARK: SED命令の実装
func (c *CPU) sed(mode AddressingMode) {
	c.registers.P.Decimal = true
}

// MARK: SEI命令の実装
func (c *CPU) sei(mode AddressingMode) {
	c.registers.P.Interrupt = true
}

// MARK: SLO命令の実装
func (c *CPU) slo(mode AddressingMode) {
	c.asl(mode)
	c.ora(mode)
}

// MARK: SRE命令の実装
func (c *CPU) sre(mode AddressingMode) {
	c.lsr(mode)
	c.eor(mode)
}

// MARK: STA命令の実装
func (c *CPU) sta(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	c.WriteByteAt(addr, c.registers.A)
}

// MARK: STX命令の実装
func (c *CPU) stx(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	c.WriteByteAt(addr, c.registers.X)
}

// MARK: STY命令の実装
func (c *CPU) sty(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	c.WriteByteAt(addr, c.registers.Y)
}

// MARK: SXA命令の実装
func (c *CPU) sxa(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	result := c.registers.X & (uint8(addr>>8) + 1)
	c.WriteByteAt(addr, result)
}

// MARK: SYA命令の実装
func (c *CPU) sya(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	result := c.registers.Y & (uint8(addr>>8) + 1)
	c.WriteByteAt(addr, result)
}

// MARK: TAX命令の実装
func (c *CPU) tax(mode AddressingMode) {
	c.registers.X = c.registers.A
	c.updateNZFlags(c.registers.X)
}

// MARK: TAY命令の実装
func (c *CPU) tay(mode AddressingMode) {
	c.registers.Y = c.registers.A
	c.updateNZFlags(c.registers.Y)
}

// MARK: TOP命令の実装
func (c *CPU) top(mode AddressingMode) {
}

// MARK: TSX命令の実装
func (c *CPU) tsx(mode AddressingMode) {
	c.registers.X = c.registers.SP
	c.updateNZFlags(c.registers.X)
}

// MARK: TXA命令
func (c *CPU) txa(mode AddressingMode) {
	c.registers.A = c.registers.X
	c.updateNZFlags(c.registers.A)
}

// MARK: TXS命令
func (c *CPU) txs(mode AddressingMode) {
	c.registers.SP = c.registers.X
}

// MARK: TYA命令
func (c *CPU) tya(mode AddressingMode) {
	c.registers.A = c.registers.Y
	c.updateNZFlags(c.registers.A)
}

// MARK: XAA命令の実装
func (c *CPU) xaa(mode AddressingMode) {
	// @NOTE 未定義動作
	addr, _ := c.calcOperandAddress(mode)
	value := c.ReadByteFrom(addr)
	c.registers.A = (c.registers.A | 0x80) & c.registers.X & value
}

// MARK: XAS命令の実装
func (c *CPU) xas(mode AddressingMode) {
	addr, _ := c.calcOperandAddress(mode)
	c.registers.SP = c.registers.X & c.registers.A
	result := c.registers.SP & (uint8(addr>>8) + 1)
	c.WriteByteAt(addr, result)
}

// MARK: canPeek: トレース時に安全に読み取れるアドレスか (副作用やpanicを避ける)
func (c *CPU) canPeek(addr uint16) bool {
	// WRAM (含ミラー)
	if addr < 0x2000 {
		return true
	}

	// PPUレジスタ ($2000-$3FFF) のうち副作用の小さいものだけ許可
	// ミラーを正規化 ($2000 + (addr & 7))
	if addr >= 0x2000 && addr <= 0x3FFF {
		m := 0x2000 + (addr & 0x0007)
		switch m {
		case 0x2000, // PPUCTRL (読み出しはラッチ値で副作用なし)
			0x2001: // PPUMASK
			return true
		default:
			return false
		}
	}

	// APU / IO
	if addr == 0x4015 { // APU STATUS (読み出し副作用なし想定)
		return true
	}
	// 0x4016/0x4017 (JoyPad) は読み出しでシフト進行するため除外

	// 拡張領域 (多くのマッパでは ROM/RAM/未使用) - bus.ReadByteFrom は panic しない
	if addr >= 0x4020 && addr < 0x6000 {
		return true
	}

	// カートリッジRAM/ROM/マッパ
	if addr >= 0x6000 {
		return true
	}

	return false
}

// MARK: デバッグ用表示メソッド
func (c *CPU) Trace() string {
	pc := c.registers.PC
	opcode := c.ReadByteFrom(pc)
	inst, ok := c.InstructionSet[opcode]
	if !ok {
		return fmt.Sprintf("%04X  %02X        ???                         A:%02X X:%02X Y:%02X P:%02X SP:%02X",
			pc, opcode, c.registers.A, c.registers.X, c.registers.Y, c.registers.P.ToByte(), c.registers.SP)
	}

	var b1, b2 uint8
	if inst.Bytes > 1 {
		b1 = c.ReadByteFrom(pc + 1)
	}
	if inst.Bytes > 2 {
		b2 = c.ReadByteFrom(pc + 2)
	}

	hexDump := fmt.Sprintf("%02X", opcode)
	switch inst.Bytes {
	case 2:
		hexDump = fmt.Sprintf("%02X %02X", opcode, b1)
	case 3:
		hexDump = fmt.Sprintf("%02X %02X %02X", opcode, b1, b2)
	}
	hexDump = fmt.Sprintf("%-8s", hexDump)

	operandStr := ""
	effAddr := uint16(0)

	mn := inst.Code.ToString()
	isStore := mn == "STA" || mn == "STX" || mn == "STY" || mn == "SAX" || mn == "AAX"

	peek := func(addr uint16) (uint8, bool) {
		if c.canPeek(addr) {
			return c.ReadByteFrom(addr), true
		}
		return 0, false
	}

	switch inst.AddressingMode {
	case Implied:
	case Accumulator:
		operandStr = "A"
	case Immediate:
		operandStr = fmt.Sprintf("#$%02X", b1)
	case Relative:
		offset := int8(b1)
		target := pc + 2 + uint16(offset)
		operandStr = fmt.Sprintf("$%04X", target)
	case ZeroPage:
		effAddr = uint16(b1)
		if !isStore {
			if v, ok := peek(effAddr); ok {
				operandStr = fmt.Sprintf("$%02X = %02X", b1, v)
				break
			}
		}
		operandStr = fmt.Sprintf("$%02X", b1)
	case ZeroPageXIndexed:
		base := b1
		effAddr = uint16(uint8(base + c.registers.X))
		if !isStore {
			if v, ok := peek(effAddr); ok {
				operandStr = fmt.Sprintf("$%02X,X @ %02X = %02X", base, effAddr, v)
				break
			}
		}
		operandStr = fmt.Sprintf("$%02X,X @ %02X", base, effAddr)
	case ZeroPageYIndexed:
		base := b1
		effAddr = uint16(uint8(base + c.registers.Y))
		if !isStore {
			if v, ok := peek(effAddr); ok {
				operandStr = fmt.Sprintf("$%02X,Y @ %02X = %02X", base, effAddr, v)
				break
			}
		}
		operandStr = fmt.Sprintf("$%02X,Y @ %02X", base, effAddr)
	case Absolute:
		effAddr = uint16(b1) | (uint16(b2) << 8)
		if opcode == 0x20 || opcode == 0x4C { // JSR/JMP
			operandStr = fmt.Sprintf("$%04X", effAddr)
		} else if !isStore {
			if v, ok := peek(effAddr); ok {
				operandStr = fmt.Sprintf("$%04X = %02X", effAddr, v)
			} else {
				operandStr = fmt.Sprintf("$%04X", effAddr)
			}
		} else {
			operandStr = fmt.Sprintf("$%04X", effAddr)
		}
	case AbsoluteXIndexed:
		base := uint16(b1) | (uint16(b2) << 8)
		effAddr = base + uint16(c.registers.X)
		if !isStore {
			if v, ok := peek(effAddr); ok {
				operandStr = fmt.Sprintf("$%04X,X @ %04X = %02X", base, effAddr, v)
				break
			}
		}
		operandStr = fmt.Sprintf("$%04X,X @ %04X", base, effAddr)
	case AbsoluteYIndexed:
		base := uint16(b1) | (uint16(b2) << 8)
		effAddr = base + uint16(c.registers.Y)
		if !isStore {
			if v, ok := peek(effAddr); ok {
				operandStr = fmt.Sprintf("$%04X,Y @ %04X = %02X", base, effAddr, v)
				break
			}
		}
		operandStr = fmt.Sprintf("$%04X,Y @ %04X", base, effAddr)
	case Indirect:
		ptr := uint16(b1) | (uint16(b2) << 8)
		var target uint16
		if ptr&0x00FF == 0x00FF {
			low := c.ReadByteFrom(ptr)
			high := c.ReadByteFrom(ptr & 0xFF00)
			target = uint16(high)<<8 | uint16(low)
		} else {
			target = c.ReadWordFrom(ptr)
		}
		operandStr = fmt.Sprintf("($%04X) = %04X", ptr, target)
	case IndirectXIndexed:
		base := b1
		ptr := uint8(base + c.registers.X)
		low := c.ReadByteFrom(uint16(ptr))
		high := c.ReadByteFrom(uint16(ptr+1) & 0x00FF)
		effAddr = uint16(high)<<8 | uint16(low)
		if !isStore {
			if v, ok := peek(effAddr); ok {
				operandStr = fmt.Sprintf("($%02X,X) @ %02X = %04X = %02X", base, ptr, effAddr, v)
				break
			}
		}
		operandStr = fmt.Sprintf("($%02X,X) @ %02X = %04X", base, ptr, effAddr)
	case IndirectYIndexed:
		base := b1
		low := c.ReadByteFrom(uint16(base))
		high := c.ReadByteFrom(uint16(base+1) & 0x00FF)
		baseAddr := uint16(high)<<8 | uint16(low)
		effAddr = baseAddr + uint16(c.registers.Y)
		if !isStore {
			if v, ok := peek(effAddr); ok {
				operandStr = fmt.Sprintf("($%02X),Y = %04X @ %04X = %02X", base, baseAddr, effAddr, v)
				break
			}
		}
		operandStr = fmt.Sprintf("($%02X),Y = %04X @ %04X", base, baseAddr, effAddr)
	default:
	}

	asm := fmt.Sprintf("%04X  %s %4s %s",
		pc,
		hexDump,
		inst.Code.ToString(),
		operandStr)

	return fmt.Sprintf("%-47s A:%02X X:%02X Y:%02X P:%02X SP:%02X",
		asm,
		c.registers.A, c.registers.X, c.registers.Y,
		c.registers.P.ToByte(), c.registers.SP)
}

// MARK: デバッグ用ログ出力切り替え
func (c *CPU) ToggleLog() {
	if c.log {
		fmt.Println("[CPU] Debug log: OFF")
	} else {
		fmt.Println("[CPU] Debug log: ON")
	}
	c.log = !c.log
}

// MARK: デバッグ用実行メソッド
func (c *CPU) REPL(commands []uint8) {
	c.InitForTest(true)

	for addr, opecode := range commands {
		c.WriteByteAt(uint16(addr), opecode)
	}

	for i := range commands {
		// BRK命令まで実行
		if commands[i] == 0x00 {
			return
		}
		c.Step()
	}
}

// RunCycles executes CPU instructions until at least targetCycles CPU cycles
// have been advanced on the bus. This is intended for frame-sliced execution
// where the caller controls per-frame timing.
func (c *CPU) RunCycles(targetCycles uint) {
	var executed uint = 0

	for executed < targetCycles {
		// Handle interrupts
		nmi := c.bus.NMI()
		if nmi {
			c.interrupt(NMI)
		}

		apuIrq := c.bus.APUIRQ()
		mapperIrq := c.bus.MapperIRQ()
		if !c.registers.P.Interrupt && (apuIrq || mapperIrq) {
			c.interrupt(IRQ)
		}

		prev := c.bus.Cycles()
		c.Step()
		post := c.bus.Cycles()

		var delta uint
		if post >= prev {
			delta = post - prev
		} else {
			delta = post
		}
		executed += delta
	}
}
