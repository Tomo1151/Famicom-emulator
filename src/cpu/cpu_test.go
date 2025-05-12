package cpu

import (
	"testing"
)

// テストヘルパー関数：CPUを初期化する
func setupCPU() *CPU {
    cpu := CreateCPU(false)
    // PCを固定アドレスに設定（テスト用）
    cpu.Registers.PC = 0x0200
    return cpu
}

// テストヘルパー関数：レジスタの値をチェックする
func checkRegister(t *testing.T, name string, got, want uint8) {
    if got != want {
        t.Errorf("%s register = %#02x, want %#02x", name, got, want)
    }
}

// テストヘルパー関数：フラグの値をチェックする
func checkFlag(t *testing.T, name string, got, want bool) {
    if got != want {
        t.Errorf("%s flag = %v, want %v", name, got, want)
    }
}

// TestAllCPUInstructions は全ての命令のテストをまとめて実行する統合テストです
func TestAllCPUInstructions(t *testing.T) {
    // フラグ操作命令テスト
    t.Run("Flag operations", func(t *testing.T) {
        TestSEC(t)
        TestCLC(t)
        TestCLV(t)
        TestSEI(t)
        TestCLI(t)
        TestSED(t)
        TestCLD(t)
    })

    // レジスタ操作命令テスト
    t.Run("Register operations", func(t *testing.T) {
        TestLDA(t)
        TestLDX(t)
        TestLDY(t)
        TestSTA(t)
        TestSTX(t)
        TestSTY(t)
        TestTAX(t)
        TestTAY(t)
        TestTXA(t)
        TestTYA(t)
        TestTXS(t)
        TestTSX(t)
    })

    // 加算・減算命令テスト
    t.Run("Addition and subtraction", func(t *testing.T) {
        TestADC(t)
        TestSBC(t)
    })

    // ビット演算命令テスト
    t.Run("Bit operations", func(t *testing.T) {
        TestAND(t)
        TestORA(t)
        TestEOR(t)
        TestBIT(t)
        TestLSR(t)
        TestASL(t)
        TestROL(t)
        TestROR(t)
    })

    // レジスタ比較命令テスト
    t.Run("Register comparisons", func(t *testing.T) {
        TestCMP(t)
        TestCPX(t)
        TestCPY(t)
    })

    // スタック操作命令テスト
    t.Run("Stack operations", func(t *testing.T) {
        TestPHA(t)
        TestPHP(t)
        TestPLA(t)
        TestPLP(t)
    })

    // インクリメント・デクリメント命令テスト
    t.Run("Increment and decrement", func(t *testing.T) {
        TestINX(t)
        TestINY(t)
        TestDEX(t)
        TestDEY(t)
        TestINC(t)
        TestDEC(t)
    })

    // ジャンプ命令テスト
    t.Run("Jump and control flow", func(t *testing.T) {
        TestJMP(t)
        TestJSR(t)
        TestRTS(t)
        TestRTI(t)
        TestBRK(t)
        TestNOP(t)
    })
}


// MARK: フラグ操作
// TestSEC はSEC命令（キャリーフラグをセット）をテストします
func TestSEC(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedCarry bool
    }{
        {
            name:       "SEC - carry false to true",
            opcode:     0x38,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x38) // SEC命令
            },
            expectedCarry: true,
        },
        {
            name:       "SEC - carry already true",
            opcode:     0x38,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Carry = true
                c.WriteByteToWRAM(c.Registers.PC, 0x38) // SEC命令
            },
            expectedCarry: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
        })
    }
}

// TestCLC はCLC命令（キャリーフラグをクリア）をテストします
func TestCLC(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedCarry bool
    }{
        {
            name:       "CLC - carry true to false",
            opcode:     0x18,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Carry = true
                c.WriteByteToWRAM(c.Registers.PC, 0x18) // CLC命令
            },
            expectedCarry: false,
        },
        {
            name:       "CLC - carry already false",
            opcode:     0x18,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x18) // CLC命令
            },
            expectedCarry: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
        })
    }
}

// TestCLV はCLV命令（オーバーフローフラグをクリア）をテストします
func TestCLV(t *testing.T) {
    tests := []struct {
        name            string
        opcode          uint8
        addrMode        AddressingMode
        setupCPU        func(*CPU)
        expectedOverflow bool
    }{
        {
            name:       "CLV - overflow true to false",
            opcode:     0xB8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Overflow = true
                c.WriteByteToWRAM(c.Registers.PC, 0xB8) // CLV命令
            },
            expectedOverflow: false,
        },
        {
            name:       "CLV - overflow already false",
            opcode:     0xB8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Overflow = false
                c.WriteByteToWRAM(c.Registers.PC, 0xB8) // CLV命令
            },
            expectedOverflow: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Overflow", c.Registers.P.Overflow, tt.expectedOverflow)
        })
    }
}

// TestSEI はSEI命令（割り込み禁止フラグをセット）をテストします
func TestSEI(t *testing.T) {
    tests := []struct {
        name              string
        opcode            uint8
        addrMode          AddressingMode
        setupCPU          func(*CPU)
        expectedInterrupt bool
    }{
        {
            name:       "SEI - interrupt false to true",
            opcode:     0x78,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Interrupt = false
                c.WriteByteToWRAM(c.Registers.PC, 0x78) // SEI命令
            },
            expectedInterrupt: true,
        },
        {
            name:       "SEI - interrupt already true",
            opcode:     0x78,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Interrupt = true
                c.WriteByteToWRAM(c.Registers.PC, 0x78) // SEI命令
            },
            expectedInterrupt: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Interrupt", c.Registers.P.Interrupt, tt.expectedInterrupt)
        })
    }
}

// TestCLI はCLI命令（割り込み禁止フラグをクリア）をテストします
func TestCLI(t *testing.T) {
    tests := []struct {
        name              string
        opcode            uint8
        addrMode          AddressingMode
        setupCPU          func(*CPU)
        expectedInterrupt bool
    }{
        {
            name:       "CLI - interrupt true to false",
            opcode:     0x58,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Interrupt = true
                c.WriteByteToWRAM(c.Registers.PC, 0x58) // CLI命令
            },
            expectedInterrupt: false,
        },
        {
            name:       "CLI - interrupt already false",
            opcode:     0x58,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Interrupt = false
                c.WriteByteToWRAM(c.Registers.PC, 0x58) // CLI命令
            },
            expectedInterrupt: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Interrupt", c.Registers.P.Interrupt, tt.expectedInterrupt)
        })
    }
}

// TestSED はSED命令（デシマルモードフラグをセット）をテストします
func TestSED(t *testing.T) {
    tests := []struct {
        name            string
        opcode          uint8
        addrMode        AddressingMode
        setupCPU        func(*CPU)
        expectedDecimal bool
    }{
        {
            name:       "SED - decimal false to true",
            opcode:     0xF8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Decimal = false
                c.WriteByteToWRAM(c.Registers.PC, 0xF8) // SED命令
            },
            expectedDecimal: true,
        },
        {
            name:       "SED - decimal already true",
            opcode:     0xF8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Decimal = true
                c.WriteByteToWRAM(c.Registers.PC, 0xF8) // SED命令
            },
            expectedDecimal: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Decimal", c.Registers.P.Decimal, tt.expectedDecimal)
        })
    }
}

// TestCLD はCLD命令（デシマルモードフラグをクリア）をテストします
func TestCLD(t *testing.T) {
    tests := []struct {
        name            string
        opcode          uint8
        addrMode        AddressingMode
        setupCPU        func(*CPU)
        expectedDecimal bool
    }{
        {
            name:       "CLD - decimal true to false",
            opcode:     0xD8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Decimal = true
                c.WriteByteToWRAM(c.Registers.PC, 0xD8) // CLD命令
            },
            expectedDecimal: false,
        },
        {
            name:       "CLD - decimal already false",
            opcode:     0xD8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.P.Decimal = false
                c.WriteByteToWRAM(c.Registers.PC, 0xD8) // CLD命令
            },
            expectedDecimal: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Decimal", c.Registers.P.Decimal, tt.expectedDecimal)
        })
    }
}


// MARK: レジスタ操作
// LDA命令のテスト (Load Accumulator)
func TestLDA(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupMemory   func(*CPU)
        expectedA     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "LDA Immediate - positive value",
            opcode:     0xA9,
            addrMode:   Immediate,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xA9) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedA:    0x42,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDA Immediate - zero value",
            opcode:     0xA9,
            addrMode:   Immediate,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xA9) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x00) // オペランド: 0x00
            },
            expectedA:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "LDA Immediate - negative value",
            opcode:     0xA9,
            addrMode:   Immediate,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xA9) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 0x80 (負の値)
            },
            expectedA:    0x80,
            expectedZero: false,
            expectedNeg:  true,
        },
        {
            name:       "LDA Zero Page",
            opcode:     0xA5,
            addrMode:   ZeroPage,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xA5) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: ZPアドレス0x42
                c.WriteByteToWRAM(0x42, 0x37) // 0x42に値を設定
            },
            expectedA:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDA Zero Page,X",
            opcode:     0xB5,
            addrMode:   ZeroPageXIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xB5) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: ZPアドレス0x42
                c.WriteByteToWRAM(0x52, 0x37) // 0x52 (0x42+0x10) に値を設定
            },
            expectedA:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDA Absolute",
            opcode:     0xAD,
            addrMode:   Absolute,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xAD) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x37) // 0x4480に値を設定
            },
            expectedA:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDA Absolute,X",
            opcode:     0xBD,
            addrMode:   AbsoluteXIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xBD) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x37) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedA:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDA Absolute,Y",
            opcode:     0xB9,
            addrMode:   AbsoluteYIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xB9) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x37) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedA:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDA Indirect,X",
            opcode:     0xA1,
            addrMode:   IndirectXIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.X = 0x04
                c.WriteByteToWRAM(c.Registers.PC, 0xA1) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x24, 0x74) // 0x24 (0x20+0x04) に低バイト
                c.WriteByteToWRAM(0x25, 0x20) // 0x25 に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2074, 0x37) // 0x2074に値を設定
            },
            expectedA:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDA Indirect,Y",
            opcode:     0xB1,
            addrMode:   IndirectYIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xB1) // LDA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x74) // 0x20に低バイト
                c.WriteByteToWRAM(0x21, 0x20) // 0x21に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2084, 0x37) // 0x2084 (0x2074+0x10) に値を設定
            },
            expectedA:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupMemory(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// LDX命令のテスト (Load X Register)
func TestLDX(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupMemory   func(*CPU)
        expectedX     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "LDX Immediate",
            opcode:     0xA2,
            addrMode:   Immediate,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xA2) // LDX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedX:    0x42,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDX Zero Page",
            opcode:     0xA6,
            addrMode:   ZeroPage,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xA6) // LDX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: ZPアドレス0x42
                c.WriteByteToWRAM(0x42, 0x37) // 0x42に値を設定
            },
            expectedX:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDX Zero Page,Y",
            opcode:     0xB6,
            addrMode:   ZeroPageYIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xB6) // LDX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: ZPアドレス0x42
                c.WriteByteToWRAM(0x52, 0x37) // 0x52 (0x42+0x10) に値を設定
            },
            expectedX:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDX Absolute",
            opcode:     0xAE,
            addrMode:   Absolute,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xAE) // LDX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x37) // 0x4480に値を設定
            },
            expectedX:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDX Absolute,Y",
            opcode:     0xBE,
            addrMode:   AbsoluteYIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xBE) // LDX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x37) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedX:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupMemory(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "X", c.Registers.X, tt.expectedX)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// LDY命令のテスト (Load Y Register)
func TestLDY(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupMemory   func(*CPU)
        expectedY     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "LDY Immediate",
            opcode:     0xA0,
            addrMode:   Immediate,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xA0) // LDY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedY:    0x42,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDY Zero Page",
            opcode:     0xA4,
            addrMode:   ZeroPage,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xA4) // LDY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: ZPアドレス0x42
                c.WriteByteToWRAM(0x42, 0x37) // 0x42に値を設定
            },
            expectedY:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDY Zero Page,X",
            opcode:     0xB4,
            addrMode:   ZeroPageXIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xB4) // LDY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: ZPアドレス0x42
                c.WriteByteToWRAM(0x52, 0x37) // 0x52 (0x42+0x10) に値を設定
            },
            expectedY:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDY Absolute",
            opcode:     0xAC,
            addrMode:   Absolute,
            setupMemory: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xAC) // LDY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x37) // 0x4480に値を設定
            },
            expectedY:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "LDY Absolute,X",
            opcode:     0xBC,
            addrMode:   AbsoluteXIndexed,
            setupMemory: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xBC) // LDY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x37) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedY:    0x37,
            expectedZero: false,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupMemory(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "Y", c.Registers.Y, tt.expectedY)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// STA命令のテスト (Store Accumulator)
func TestSTA(t *testing.T) {
    tests := []struct {
        name           string
        opcode         uint8
        addrMode       AddressingMode
        setupCPU       func(*CPU)
        checkMemory    func(*testing.T, *CPU)
    }{
        {
            name:       "STA Zero Page",
            opcode:     0x85,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x85) // STA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x20) != 0x42 {
                    t.Errorf("Memory at $20 = %#02x, want %#02x", c.ReadByteFromWRAM(0x20), 0x42)
                }
            },
        },
        {
            name:       "STA Zero Page,X",
            opcode:     0x95,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x08
                c.WriteByteToWRAM(c.Registers.PC, 0x95) // STA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x28) != 0x42 { // 0x20 + 0x08
                    t.Errorf("Memory at $28 = %#02x, want %#02x", c.ReadByteFromWRAM(0x28), 0x42)
                }
            },
        },
        {
            name:       "STA Absolute",
            opcode:     0x8D,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x8D) // STA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x00) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x30) // オペランド: 高バイト (0x3000)
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x3000) != 0x42 {
                    t.Errorf("Memory at $3000 = %#02x, want %#02x", c.ReadByteFromWRAM(0x3000), 0x42)
                }
            },
        },
        {
            name:       "STA Absolute,X",
            opcode:     0x9D,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x9D) // STA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x00) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x30) // オペランド: 高バイト (0x3000)
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x3010) != 0x42 { // 0x3000 + 0x10
                    t.Errorf("Memory at $3010 = %#02x, want %#02x", c.ReadByteFromWRAM(0x3010), 0x42)
                }
            },
        },
        {
            name:       "STA Absolute,Y",
            opcode:     0x99,
            addrMode:   AbsoluteYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x99) // STA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x00) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x30) // オペランド: 高バイト (0x3000)
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x3010) != 0x42 { // 0x3000 + 0x10
                    t.Errorf("Memory at $3010 = %#02x, want %#02x", c.ReadByteFromWRAM(0x3010), 0x42)
                }
            },
        },
        {
            name:       "STA Indirect,X",
            opcode:     0x81,
            addrMode:   IndirectXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x04
                c.WriteByteToWRAM(c.Registers.PC, 0x81) // STA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x24, 0x00) // 0x24 (0x20+0x04) に低バイト
                c.WriteByteToWRAM(0x25, 0x30) // 0x25に高バイト (→ 0x3000)
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x3000) != 0x42 {
                    t.Errorf("Memory at $3000 = %#02x, want %#02x", c.ReadByteFromWRAM(0x3000), 0x42)
                }
            },
        },
        {
            name:       "STA Indirect,Y",
            opcode:     0x91,
            addrMode:   IndirectYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x91) // STA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x00) // 0x20に低バイト
                c.WriteByteToWRAM(0x21, 0x30) // 0x21に高バイト (→ 0x3000)
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x3010) != 0x42 { // 0x3000 + 0x10
                    t.Errorf("Memory at $3010 = %#02x, want %#02x", c.ReadByteFromWRAM(0x3010), 0x42)
                }
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)

            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // メモリ内容を検証
            tt.checkMemory(t, c)
        })
    }
}

// STX命令のテスト (Store X Register)
func TestSTX(t *testing.T) {
    tests := []struct {
        name           string
        opcode         uint8
        addrMode       AddressingMode
        setupCPU       func(*CPU)
        checkMemory    func(*testing.T, *CPU)
    }{
        {
            name:       "STX Zero Page",
            opcode:     0x86,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x86) // STX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x20) != 0x42 {
                    t.Errorf("Memory at $20 = %#02x, want %#02x", c.ReadByteFromWRAM(0x20), 0x42)
                }
            },
        },
        {
            name:       "STX Zero Page,Y",
            opcode:     0x96,
            addrMode:   ZeroPageYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.Registers.Y = 0x08
                c.WriteByteToWRAM(c.Registers.PC, 0x96) // STX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x28) != 0x42 { // 0x20 + 0x08
                    t.Errorf("Memory at $28 = %#02x, want %#02x", c.ReadByteFromWRAM(0x28), 0x42)
                }
            },
        },
        {
            name:       "STX Absolute",
            opcode:     0x8E,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x8E) // STX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x00) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x30) // オペランド: 高バイト (0x3000)
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x3000) != 0x42 {
                    t.Errorf("Memory at $3000 = %#02x, want %#02x", c.ReadByteFromWRAM(0x3000), 0x42)
                }
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // メモリ内容を検証
            tt.checkMemory(t, c)
        })
    }
}

// STY命令のテスト (Store Y Register)
func TestSTY(t *testing.T) {
    tests := []struct {
        name           string
        opcode         uint8
        addrMode       AddressingMode
        setupCPU       func(*CPU)
        checkMemory    func(*testing.T, *CPU)
    }{
        {
            name:       "STY Zero Page",
            opcode:     0x84,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x84) // STY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x20) != 0x42 {
                    t.Errorf("Memory at $20 = %#02x, want %#02x", c.ReadByteFromWRAM(0x20), 0x42)
                }
            },
        },
        {
            name:       "STY Zero Page,X",
            opcode:     0x94,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.Registers.X = 0x08
                c.WriteByteToWRAM(c.Registers.PC, 0x94) // STY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x28) != 0x42 { // 0x20 + 0x08
                    t.Errorf("Memory at $28 = %#02x, want %#02x", c.ReadByteFromWRAM(0x28), 0x42)
                }
            },
        },
        {
            name:       "STY Absolute",
            opcode:     0x8C,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x8C) // STY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x00) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x30) // オペランド: 高バイト (0x3000)
            },
            checkMemory: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x3000) != 0x42 {
                    t.Errorf("Memory at $3000 = %#02x, want %#02x", c.ReadByteFromWRAM(0x3000), 0x42)
                }
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // メモリ内容を検証
            tt.checkMemory(t, c)
        })
    }
}

// TestTAX はTAX命令（アキュムレータからXレジスタへの転送）をテストします
func TestTAX(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedX     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "TAX - positive value",
            opcode:     0xAA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xAA) // TAX命令
            },
            expectedX:    0x42,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "TAX - zero value",
            opcode:     0xAA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x00
                c.WriteByteToWRAM(c.Registers.PC, 0xAA) // TAX命令
            },
            expectedX:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "TAX - negative value",
            opcode:     0xAA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x80
                c.WriteByteToWRAM(c.Registers.PC, 0xAA) // TAX命令
            },
            expectedX:    0x80,
            expectedZero: false,
            expectedNeg:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "X", c.Registers.X, tt.expectedX)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestTAY はTAY命令（アキュムレータからYレジスタへの転送）をテストします
func TestTAY(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedY     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "TAY - positive value",
            opcode:     0xA8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xA8) // TAY命令
            },
            expectedY:    0x42,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "TAY - zero value",
            opcode:     0xA8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x00
                c.WriteByteToWRAM(c.Registers.PC, 0xA8) // TAY命令
            },
            expectedY:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "TAY - negative value",
            opcode:     0xA8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x80
                c.WriteByteToWRAM(c.Registers.PC, 0xA8) // TAY命令
            },
            expectedY:    0x80,
            expectedZero: false,
            expectedNeg:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "Y", c.Registers.Y, tt.expectedY)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestTXA はTXA命令（Xレジスタからアキュムレータへの転送）をテストします
func TestTXA(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedA     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "TXA - positive value",
            opcode:     0x8A,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x8A) // TXA命令
            },
            expectedA:    0x42,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "TXA - zero value",
            opcode:     0x8A,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x00
                c.WriteByteToWRAM(c.Registers.PC, 0x8A) // TXA命令
            },
            expectedA:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "TXA - negative value",
            opcode:     0x8A,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x80
                c.WriteByteToWRAM(c.Registers.PC, 0x8A) // TXA命令
            },
            expectedA:    0x80,
            expectedZero: false,
            expectedNeg:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestTYA はTYA命令（Yレジスタからアキュムレータへの転送）をテストします
func TestTYA(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedA     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "TYA - positive value",
            opcode:     0x98,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x98) // TYA命令
            },
            expectedA:    0x42,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "TYA - zero value",
            opcode:     0x98,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x00
                c.WriteByteToWRAM(c.Registers.PC, 0x98) // TYA命令
            },
            expectedA:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "TYA - negative value",
            opcode:     0x98,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x80
                c.WriteByteToWRAM(c.Registers.PC, 0x98) // TYA命令
            },
            expectedA:    0x80,
            expectedZero: false,
            expectedNeg:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestTXS はTXS命令（XレジスタからSPへの転送）をテストします
// 注意: TXS命令はフラグに影響を与えません
func TestTXS(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedSP    uint8
    }{
        {
            name:       "TXS - normal value",
            opcode:     0x9A,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x9A) // TXS命令
            },
            expectedSP: 0x42,
        },
        {
            name:       "TXS - high value",
            opcode:     0x9A,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0xFF
                c.WriteByteToWRAM(c.Registers.PC, 0x9A) // TXS命令
            },
            expectedSP: 0xFF,
        },
        {
            name:       "TXS - low value",
            opcode:     0x9A,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x00
                c.WriteByteToWRAM(c.Registers.PC, 0x9A) // TXS命令
            },
            expectedSP: 0x00,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証（SPレジスタのみ）
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
        })
    }
}

// TestTSX はTSX命令（SPからXレジスタへの転送）をテストします
func TestTSX(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedX     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "TSX - positive value",
            opcode:     0xBA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xBA) // TSX命令
            },
            expectedX:    0x42,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "TSX - zero value",
            opcode:     0xBA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0x00
                c.WriteByteToWRAM(c.Registers.PC, 0xBA) // TSX命令
            },
            expectedX:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "TSX - negative value",
            opcode:     0xBA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0x80
                c.WriteByteToWRAM(c.Registers.PC, 0xBA) // TSX命令
            },
            expectedX:    0x80,
            expectedZero: false,
            expectedNeg:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "X", c.Registers.X, tt.expectedX)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}


// MARK: 加算/減算
// TestADC はADC命令（キャリー付き加算）をテストします
func TestADC(t *testing.T) {
    tests := []struct {
        name             string
        opcode           uint8
        addrMode         AddressingMode
        setupCPU         func(*CPU)
        expectedA        uint8
        expectedZero     bool
        expectedNeg      bool
        expectedCarry    bool
        expectedOverflow bool
    }{
        {
            name:       "ADC Immediate - basic addition",
            opcode:     0x69,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x10
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x69) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x15) // オペランド: 0x15
            },
            expectedA:        0x25, // 0x10 + 0x15 = 0x25
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
        {
            name:       "ADC Immediate - addition with initial carry set",
            opcode:     0x69,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x10
                c.Registers.P.Carry = true
                c.WriteByteToWRAM(c.Registers.PC, 0x69) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x15) // オペランド: 0x15
            },
            expectedA:        0x26, // 0x10 + 0x15 + 1 (Carry) = 0x26
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
        {
            name:       "ADC Immediate - carry out",
            opcode:     0x69,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xFF
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x69) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x01) // オペランド: 0x01
            },
            expectedA:        0x00, // 0xFF + 0x01 = 0x100 (下位8bit = 0x00)
            expectedZero:     true,
            expectedNeg:      false,
            expectedCarry:    true, // キャリー発生
            expectedOverflow: false,
        },
        {
            name:       "ADC Immediate - overflow case (positive+positive=negative)",
            opcode:     0x69,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x7F // 01111111 (127)
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x69) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x01) // オペランド: 0x01
            },
            expectedA:        0x80, // 0x7F + 0x01 = 0x80 (-128 as signed)
            expectedZero:     false,
            expectedNeg:      true, // 負数（最上位ビットが1）
            expectedCarry:    false,
            expectedOverflow: true, // 符号付きオーバーフロー発生
        },
        {
            name:       "ADC Immediate - overflow case (negative+negative=positive)",
            opcode:     0x69,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x80 // 10000000 (-128)
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x69) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 0x80 (-128)
            },
            expectedA:        0x00, // 0x80 + 0x80 = 0x100 (下位8bit = 0x00)
            expectedZero:     true,
            expectedNeg:      false,
            expectedCarry:    true, // キャリー発生
            expectedOverflow: true, // 符号付きオーバーフロー発生
        },
        {
            name:       "ADC Zero Page",
            opcode:     0x65,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x65) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x13) // 0x20に値を設定
            },
            expectedA:        0x55, // 0x42 + 0x13 = 0x55
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
        {
            name:       "ADC Zero Page,X",
            opcode:     0x75,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x10
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x75) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x13) // 0x30 (0x20+0x10) に値を設定
            },
            expectedA:        0x55, // 0x42 + 0x13 = 0x55
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
        {
            name:       "ADC Absolute",
            opcode:     0x6D,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x6D) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x13) // 0x4480に値を設定
            },
            expectedA:        0x55, // 0x42 + 0x13 = 0x55
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
        {
            name:       "ADC Absolute,X",
            opcode:     0x7D,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x10
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x7D) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x13) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedA:        0x55, // 0x42 + 0x13 = 0x55
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
        {
            name:       "ADC Absolute,Y",
            opcode:     0x79,
            addrMode:   AbsoluteYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.Y = 0x10
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x79) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x13) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedA:        0x55, // 0x42 + 0x13 = 0x55
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
        {
            name:       "ADC Indirect,X",
            opcode:     0x61,
            addrMode:   IndirectXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x04
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x61) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x24, 0x74) // 0x24 (0x20+0x04) に低バイト
                c.WriteByteToWRAM(0x25, 0x20) // 0x25 に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2074, 0x13) // 0x2074に値を設定
            },
            expectedA:        0x55, // 0x42 + 0x13 = 0x55
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
        {
            name:       "ADC Indirect,Y",
            opcode:     0x71,
            addrMode:   IndirectYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.Y = 0x10
                c.Registers.P.Carry = false
                c.WriteByteToWRAM(c.Registers.PC, 0x71) // ADC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x74) // 0x20に低バイト
                c.WriteByteToWRAM(0x21, 0x20) // 0x21に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2084, 0x13) // 0x2084 (0x2074+0x10) に値を設定
            },
            expectedA:        0x55, // 0x42 + 0x13 = 0x55
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    false,
            expectedOverflow: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Overflow", c.Registers.P.Overflow, tt.expectedOverflow)
        })
    }
}

// TestSBC はSBC命令（キャリー付き減算）をテストします
func TestSBC(t *testing.T) {
    tests := []struct {
        name             string
        opcode           uint8
        addrMode         AddressingMode
        setupCPU         func(*CPU)
        expectedA        uint8
        expectedZero     bool
        expectedNeg      bool
        expectedCarry    bool
        expectedOverflow bool
    }{
        {
            name:       "SBC Immediate - basic subtraction with carry set",
            opcode:     0xE9,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.P.Carry = true // ボローなし（1 = 借りなし、0 = 借りあり）
                c.WriteByteToWRAM(c.Registers.PC, 0xE9) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x30) // オペランド: 0x30
            },
            expectedA:        0x20, // 0x50 - 0x30 = 0x20
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
        {
            name:       "SBC Immediate - subtraction with borrow (carry clear)",
            opcode:     0xE9,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.P.Carry = false // ボローあり（0 = 借りあり）
                c.WriteByteToWRAM(c.Registers.PC, 0xE9) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x30) // オペランド: 0x30
            },
            expectedA:        0x1F, // 0x50 - 0x30 - 1 (ボロー) = 0x1F
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
        {
            name:       "SBC Immediate - borrow out (carry clear result)",
            opcode:     0xE9,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x30
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xE9) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x40) // オペランド: 0x40
            },
            expectedA:        0xF0, // 0x30 - 0x40 = 0xF0 (下位8bit)
            expectedZero:     false,
            expectedNeg:      true, // 負数（最上位ビットが1）
            expectedCarry:    false, // ボローあり
            expectedOverflow: false,
        },
        {
            name:       "SBC Immediate - overflow case (positive-negative=negative)",
            opcode:     0xE9,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50 // 01010000 (正数)
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xE9) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0xB0) // オペランド: 0xB0 (負数)
            },
            expectedA:        0xA0, // 0x50 - 0xB0 = 0xA0 (下位8bit)
            expectedZero:     false,
            expectedNeg:      true, // 負数
            expectedCarry:    false, // ボローあり
            expectedOverflow: true, // 符号付きオーバーフロー発生 (正 - 負 = 負)
        },
        {
            name:       "SBC Immediate - overflow case (negative-positive=positive)",
            opcode:     0xE9,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x90 // 10010000 (負数)
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xE9) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x10) // オペランド: 0x10 (正数)
            },
            expectedA:        0x80, // 0x90 - 0x10 = 0x80 (下位8bit)
            expectedZero:     false,
            expectedNeg:      true, // 負数
            expectedCarry:    true, // ボローなし
            expectedOverflow: false, // 符号付きオーバーフローは発生しない (負 - 正)
        },
        {
            name:       "SBC Zero Page",
            opcode:     0xE5,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xE5) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x30) // 0x20に値を設定
            },
            expectedA:        0x20, // 0x50 - 0x30 = 0x20
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
        {
            name:       "SBC Zero Page,X",
            opcode:     0xF5,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.X = 0x10
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xF5) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x30) // 0x30 (0x20+0x10) に値を設定
            },
            expectedA:        0x20, // 0x50 - 0x30 = 0x20
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
        {
            name:       "SBC Absolute",
            opcode:     0xED,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xED) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x30) // 0x4480に値を設定
            },
            expectedA:        0x20, // 0x50 - 0x30 = 0x20
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
        {
            name:       "SBC Absolute,X",
            opcode:     0xFD,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.X = 0x10
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xFD) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x30) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedA:        0x20, // 0x50 - 0x30 = 0x20
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
        {
            name:       "SBC Absolute,Y",
            opcode:     0xF9,
            addrMode:   AbsoluteYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.Y = 0x10
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xF9) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x30) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedA:        0x20, // 0x50 - 0x30 = 0x20
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
        {
            name:       "SBC Indirect,X",
            opcode:     0xE1,
            addrMode:   IndirectXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.X = 0x04
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xE1) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x24, 0x74) // 0x24 (0x20+0x04) に低バイト
                c.WriteByteToWRAM(0x25, 0x20) // 0x25 に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2074, 0x30) // 0x2074に値を設定
            },
            expectedA:        0x20, // 0x50 - 0x30 = 0x20
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
        {
            name:       "SBC Indirect,Y",
            opcode:     0xF1,
            addrMode:   IndirectYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x50
                c.Registers.Y = 0x10
                c.Registers.P.Carry = true // ボローなし
                c.WriteByteToWRAM(c.Registers.PC, 0xF1) // SBC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x74) // 0x20に低バイト
                c.WriteByteToWRAM(0x21, 0x20) // 0x21に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2084, 0x30) // 0x2084 (0x2074+0x10) に値を設定
            },
            expectedA:        0x20, // 0x50 - 0x30 = 0x20
            expectedZero:     false,
            expectedNeg:      false,
            expectedCarry:    true, // ボローなし
            expectedOverflow: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Overflow", c.Registers.P.Overflow, tt.expectedOverflow)
        })
    }
}


// MARK: ビット演算
// TestAND はAND命令（論理積）をテストします
func TestAND(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedA     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "AND Immediate - basic AND operation",
            opcode:     0x29,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x29) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x0F) // オペランド: 0x0F (00001111)
            },
            expectedA:    0x00, // 0xF0 & 0x0F = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "AND Immediate - result with bits set",
            opcode:     0x29,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xFF // 11111111
                c.WriteByteToWRAM(c.Registers.PC, 0x29) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0xAA) // オペランド: 0xAA (10101010)
            },
            expectedA:    0xAA, // 0xFF & 0xAA = 0xAA
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "AND Zero Page",
            opcode:     0x25,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x25) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x0F) // 0x20に値を設定 (00001111)
            },
            expectedA:    0x00, // 0xF0 & 0x0F = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "AND Zero Page,X",
            opcode:     0x35,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x35) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x0F) // 0x30 (0x20+0x10) に値を設定 (00001111)
            },
            expectedA:    0x00, // 0xF0 & 0x0F = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "AND Absolute",
            opcode:     0x2D,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x2D) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x0F) // 0x4480に値を設定 (00001111)
            },
            expectedA:    0x00, // 0xF0 & 0x0F = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "AND Absolute,X",
            opcode:     0x3D,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x3D) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x0F) // 0x4490 (0x4480+0x10) に値を設定 (00001111)
            },
            expectedA:    0x00, // 0xF0 & 0x0F = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "AND Absolute,Y",
            opcode:     0x39,
            addrMode:   AbsoluteYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x39) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x0F) // 0x4490 (0x4480+0x10) に値を設定 (00001111)
            },
            expectedA:    0x00, // 0xF0 & 0x0F = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "AND Indirect,X",
            opcode:     0x21,
            addrMode:   IndirectXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x04
                c.WriteByteToWRAM(c.Registers.PC, 0x21) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x24, 0x74) // 0x24 (0x20+0x04) に低バイト
                c.WriteByteToWRAM(0x25, 0x20) // 0x25 に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2074, 0x0F) // 0x2074に値を設定 (00001111)
            },
            expectedA:    0x00, // 0xF0 & 0x0F = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "AND Indirect,Y",
            opcode:     0x31,
            addrMode:   IndirectYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x31) // AND命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x74) // 0x20に低バイト
                c.WriteByteToWRAM(0x21, 0x20) // 0x21に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2084, 0x0F) // 0x2084 (0x2074+0x10) に値を設定 (00001111)
            },
            expectedA:    0x00, // 0xF0 & 0x0F = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestORA はORA命令（論理和）をテストします
func TestORA(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedA     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "ORA Immediate - basic OR operation",
            opcode:     0x09,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x09) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x0F) // オペランド: 0x0F (00001111)
            },
            expectedA:    0xFF, // 0xF0 | 0x0F = 0xFF
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "ORA Immediate - zero result",
            opcode:     0x09,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x00 // 00000000
                c.WriteByteToWRAM(c.Registers.PC, 0x09) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x00) // オペランド: 0x00 (00000000)
            },
            expectedA:    0x00, // 0x00 | 0x00 = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "ORA Zero Page",
            opcode:     0x05,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x05) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x0F) // 0x20に値を設定 (00001111)
            },
            expectedA:    0xFF, // 0xF0 | 0x0F = 0xFF
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "ORA Zero Page,X",
            opcode:     0x15,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x15) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x0F) // 0x30 (0x20+0x10) に値を設定 (00001111)
            },
            expectedA:    0xFF, // 0xF0 | 0x0F = 0xFF
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "ORA Absolute",
            opcode:     0x0D,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x0D) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x0F) // 0x4480に値を設定 (00001111)
            },
            expectedA:    0xFF, // 0xF0 | 0x0F = 0xFF
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "ORA Absolute,X",
            opcode:     0x1D,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x1D) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x0F) // 0x4490 (0x4480+0x10) に値を設定 (00001111)
            },
            expectedA:    0xFF, // 0xF0 | 0x0F = 0xFF
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "ORA Absolute,Y",
            opcode:     0x19,
            addrMode:   AbsoluteYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x19) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x0F) // 0x4490 (0x4480+0x10) に値を設定 (00001111)
            },
            expectedA:    0xFF, // 0xF0 | 0x0F = 0xFF
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "ORA Indirect,X",
            opcode:     0x01,
            addrMode:   IndirectXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x04
                c.WriteByteToWRAM(c.Registers.PC, 0x01) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x24, 0x74) // 0x24 (0x20+0x04) に低バイト
                c.WriteByteToWRAM(0x25, 0x20) // 0x25 に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2074, 0x0F) // 0x2074に値を設定 (00001111)
            },
            expectedA:    0xFF, // 0xF0 | 0x0F = 0xFF
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "ORA Indirect,Y",
            opcode:     0x11,
            addrMode:   IndirectYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x11) // ORA命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x74) // 0x20に低バイト
                c.WriteByteToWRAM(0x21, 0x20) // 0x21に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2084, 0x0F) // 0x2084 (0x2074+0x10) に値を設定 (00001111)
            },
            expectedA:    0xFF, // 0xF0 | 0x0F = 0xFF
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestEOR はEOR命令（排他的論理和）をテストします
func TestEOR(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedA     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "EOR Immediate - basic XOR operation",
            opcode:     0x49,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x49) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0xFF) // オペランド: 0xFF (11111111)
            },
            expectedA:    0x0F, // 0xF0 ^ 0xFF = 0x0F
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "EOR Immediate - zero result",
            opcode:     0x49,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xFF // 11111111
                c.WriteByteToWRAM(c.Registers.PC, 0x49) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0xFF) // オペランド: 0xFF (11111111)
            },
            expectedA:    0x00, // 0xFF ^ 0xFF = 0x00
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "EOR Immediate - negative result",
            opcode:     0x49,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x0F // 00001111
                c.WriteByteToWRAM(c.Registers.PC, 0x49) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0xFF) // オペランド: 0xFF (11111111)
            },
            expectedA:    0xF0, // 0x0F ^ 0xFF = 0xF0
            expectedZero: false,
            expectedNeg:  true, // 負数（最上位ビットが1）
        },
        {
            name:       "EOR Zero Page",
            opcode:     0x45,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x45) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0xFF) // 0x20に値を設定 (11111111)
            },
            expectedA:    0x0F, // 0xF0 ^ 0xFF = 0x0F
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "EOR Zero Page,X",
            opcode:     0x55,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x55) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0xFF) // 0x30 (0x20+0x10) に値を設定 (11111111)
            },
            expectedA:    0x0F, // 0xF0 ^ 0xFF = 0x0F
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "EOR Absolute",
            opcode:     0x4D,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.WriteByteToWRAM(c.Registers.PC, 0x4D) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0xFF) // 0x4480に値を設定 (11111111)
            },
            expectedA:    0x0F, // 0xF0 ^ 0xFF = 0x0F
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "EOR Absolute,X",
            opcode:     0x5D,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x5D) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0xFF) // 0x4490 (0x4480+0x10) に値を設定 (11111111)
            },
            expectedA:    0x0F, // 0xF0 ^ 0xFF = 0x0F
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "EOR Absolute,Y",
            opcode:     0x59,
            addrMode:   AbsoluteYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x59) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0xFF) // 0x4490 (0x4480+0x10) に値を設定 (11111111)
            },
            expectedA:    0x0F, // 0xF0 ^ 0xFF = 0x0F
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "EOR Indirect,X",
            opcode:     0x41,
            addrMode:   IndirectXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.X = 0x04
                c.WriteByteToWRAM(c.Registers.PC, 0x41) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x24, 0x74) // 0x24 (0x20+0x04) に低バイト
                c.WriteByteToWRAM(0x25, 0x20) // 0x25 に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2074, 0xFF) // 0x2074に値を設定 (11111111)
            },
            expectedA:    0x0F, // 0xF0 ^ 0xFF = 0x0F
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "EOR Indirect,Y",
            opcode:     0x51,
            addrMode:   IndirectYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xF0 // 11110000
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x51) // EOR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x74) // 0x20に低バイト
                c.WriteByteToWRAM(0x21, 0x20) // 0x21に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2084, 0xFF) // 0x2084 (0x2074+0x10) に値を設定 (11111111)
            },
            expectedA:    0x0F, // 0xF0 ^ 0xFF = 0x0F
            expectedZero: false,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}


// MARK: ビット操作
// TestBIT はBIT命令（ビットテスト）をテストします
func TestBIT(t *testing.T) {
    tests := []struct {
        name             string
        opcode           uint8
        addrMode         AddressingMode
        setupCPU         func(*CPU)
        expectedA        uint8      // Aレジスタは変更されない
        expectedZero     bool
        expectedNeg      bool
        expectedOverflow bool
    }{
        {
            name:       "BIT Zero Page - Zero flag set (A & M == 0)",
            opcode:     0x24,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x0F // 00001111
                c.WriteByteToWRAM(c.Registers.PC, 0x24) // BIT命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0xF0) // 0x20に値を設定 (11110000)
            },
            expectedA:        0x0F, // 変更なし
            expectedZero:     true,  // A & M = 0
            expectedNeg:      true,  // M のビット7が1
            expectedOverflow: true,  // M のビット6が1
        },
        {
            name:       "BIT Zero Page - Zero flag clear (A & M != 0)",
            opcode:     0x24,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0xFF // 11111111
                c.WriteByteToWRAM(c.Registers.PC, 0x24) // BIT命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x0C) // 0x20に値を設定 (00001100)
            },
            expectedA:        0xFF, // 変更なし
            expectedZero:     false, // A & M != 0
            expectedNeg:      false, // M のビット7が0
            expectedOverflow: false, // M のビット6が0
        },
        {
            name:       "BIT Absolute",
            opcode:     0x2C,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x0F // 00001111
                c.WriteByteToWRAM(c.Registers.PC, 0x2C) // BIT命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x40) // 0x4480に値を設定 (01000000)
            },
            expectedA:        0x0F, // 変更なし
            expectedZero:     true,  // A & M = 0
            expectedNeg:      false, // M のビット7が0
            expectedOverflow: true,  // M のビット6が1
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA) // Aレジスタは変更されないことを確認
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
            checkFlag(t, "Overflow", c.Registers.P.Overflow, tt.expectedOverflow)
        })
    }
}

// TestLSR はLSR命令（論理右シフト）をテストします
func TestLSR(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        checkResult   func(*testing.T, *CPU)
        expectedCarry bool
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "LSR Accumulator - typical case",
            opcode:     0x4A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x82 // 10000010
                c.WriteByteToWRAM(c.Registers.PC, 0x4A) // LSR命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x41) // 01000001
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   false,
        },
        {
            name:       "LSR Accumulator - carry out",
            opcode:     0x4A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x01 // 00000001
                c.WriteByteToWRAM(c.Registers.PC, 0x4A) // LSR命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x00) // 00000000
            },
            expectedCarry: true,
            expectedZero:  true,
            expectedNeg:   false,
        },
        {
            name:       "LSR Zero Page",
            opcode:     0x46,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0x46) // LSR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x82) // 0x20に値を設定 (10000010)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x20) != 0x41 { // 01000001
                    t.Errorf("Memory at $20 = %#02x, want %#02x", c.ReadByteFromWRAM(0x20), 0x41)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   false,
        },
        {
            name:       "LSR Zero Page,X",
            opcode:     0x56,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x56) // LSR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x82) // 0x30 (0x20+0x10) に値を設定 (10000010)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x30) != 0x41 { // 01000001
                    t.Errorf("Memory at $30 = %#02x, want %#02x", c.ReadByteFromWRAM(0x30), 0x41)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   false,
        },
        {
            name:       "LSR Absolute",
            opcode:     0x4E,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0x4E) // LSR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x82) // 0x4480に値を設定 (10000010)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x4480) != 0x41 { // 01000001
                    t.Errorf("Memory at $4480 = %#02x, want %#02x", c.ReadByteFromWRAM(0x4480), 0x41)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   false,
        },
        {
            name:       "LSR Absolute,X",
            opcode:     0x5E,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x5E) // LSR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x82) // 0x4490 (0x4480+0x10) に値を設定 (10000010)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x4490) != 0x41 { // 01000001
                    t.Errorf("Memory at $4490 = %#02x, want %#02x", c.ReadByteFromWRAM(0x4490), 0x41)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            tt.checkResult(t, c)
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestASL はASL命令（算術左シフト）をテストします
func TestASL(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        checkResult   func(*testing.T, *CPU)
        expectedCarry bool
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "ASL Accumulator - typical case",
            opcode:     0x0A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x41 // 01000001
                c.WriteByteToWRAM(c.Registers.PC, 0x0A) // ASL命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x82) // 10000010
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ASL Accumulator - carry out",
            opcode:     0x0A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x80 // 10000000
                c.WriteByteToWRAM(c.Registers.PC, 0x0A) // ASL命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x00) // 00000000
            },
            expectedCarry: true,
            expectedZero:  true,
            expectedNeg:   false,
        },
        {
            name:       "ASL Zero Page",
            opcode:     0x06,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0x06) // ASL命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x41) // 0x20に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x20) != 0x82 { // 10000010
                    t.Errorf("Memory at $20 = %#02x, want %#02x", c.ReadByteFromWRAM(0x20), 0x82)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ASL Zero Page,X",
            opcode:     0x16,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x16) // ASL命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x41) // 0x30 (0x20+0x10) に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x30) != 0x82 { // 10000010
                    t.Errorf("Memory at $30 = %#02x, want %#02x", c.ReadByteFromWRAM(0x30), 0x82)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ASL Absolute",
            opcode:     0x0E,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0x0E) // ASL命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x41) // 0x4480に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x4480) != 0x82 { // 10000010
                    t.Errorf("Memory at $4480 = %#02x, want %#02x", c.ReadByteFromWRAM(0x4480), 0x82)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ASL Absolute,X",
            opcode:     0x1E,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0x1E) // ASL命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x41) // 0x4490 (0x4480+0x10) に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x4490) != 0x82 { // 10000010
                    t.Errorf("Memory at $4490 = %#02x, want %#02x", c.ReadByteFromWRAM(0x4490), 0x82)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            tt.checkResult(t, c)
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestROL はROL命令（左回転）をテストします
func TestROL(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        checkResult   func(*testing.T, *CPU)
        expectedCarry bool
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "ROL Accumulator - with carry in=0",
            opcode:     0x2A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x41 // 01000001
                c.Registers.P.Carry = false // キャリーなし
                c.WriteByteToWRAM(c.Registers.PC, 0x2A) // ROL命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x82) // 10000010
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROL Accumulator - with carry in=1",
            opcode:     0x2A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x41 // 01000001
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x2A) // ROL命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x83) // 10000011
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROL Accumulator - carry out",
            opcode:     0x2A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x80 // 10000000
                c.Registers.P.Carry = false // キャリーなし
                c.WriteByteToWRAM(c.Registers.PC, 0x2A) // ROL命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x00) // 00000000
            },
            expectedCarry: true,
            expectedZero:  true,
            expectedNeg:   false,
        },
        {
            name:       "ROL Zero Page",
            opcode:     0x26,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x26) // ROL命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x41) // 0x20に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x20) != 0x83 { // 10000011
                    t.Errorf("Memory at $20 = %#02x, want %#02x", c.ReadByteFromWRAM(0x20), 0x83)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROL Zero Page,X",
            opcode:     0x36,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x36) // ROL命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x41) // 0x30 (0x20+0x10) に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x30) != 0x83 { // 10000011
                    t.Errorf("Memory at $30 = %#02x, want %#02x", c.ReadByteFromWRAM(0x30), 0x83)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROL Absolute",
            opcode:     0x2E,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x2E) // ROL命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x41) // 0x4480に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x4480) != 0x83 { // 10000011
                    t.Errorf("Memory at $4480 = %#02x, want %#02x", c.ReadByteFromWRAM(0x4480), 0x83)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROL Absolute,X",
            opcode:     0x3E,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x3E) // ROL命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x41) // 0x4490 (0x4480+0x10) に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x4490) != 0x83 { // 10000011
                    t.Errorf("Memory at $4490 = %#02x, want %#02x", c.ReadByteFromWRAM(0x4490), 0x83)
                }
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            tt.checkResult(t, c)
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestROR はROR命令（右回転）をテストします
func TestROR(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        checkResult   func(*testing.T, *CPU)
        expectedCarry bool
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "ROR Accumulator - with carry in=0",
            opcode:     0x6A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x41 // 01000001
                c.Registers.P.Carry = false // キャリーなし
                c.WriteByteToWRAM(c.Registers.PC, 0x6A) // ROR命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x20) // 00100000
            },
            expectedCarry: true,  // ビット0が1だったため
            expectedZero:  false,
            expectedNeg:   false,
        },
        {
            name:       "ROR Accumulator - with carry in=1",
            opcode:     0x6A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x41 // 01000001
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x6A) // ROR命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0xA0) // 10100000
            },
            expectedCarry: true,  // ビット0が1だったため
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROR Accumulator - carry out=0",
            opcode:     0x6A,
            addrMode:   Accumulator,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x40 // 01000000
                c.Registers.P.Carry = false // キャリーなし
                c.WriteByteToWRAM(c.Registers.PC, 0x6A) // ROR命令
            },
            checkResult: func(t *testing.T, c *CPU) {
                checkRegister(t, "A", c.Registers.A, 0x20) // 00100000
            },
            expectedCarry: false,
            expectedZero:  false,
            expectedNeg:   false,
        },
        {
            name:       "ROR Zero Page",
            opcode:     0x66,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x66) // ROR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x41) // 0x20に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x20) != 0xA0 { // 10100000
                    t.Errorf("Memory at $20 = %#02x, want %#02x", c.ReadByteFromWRAM(0x20), 0xA0)
                }
            },
            expectedCarry: true,  // ビット0が1だったため
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROR Zero Page,X",
            opcode:     0x76,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x76) // ROR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x41) // 0x30 (0x20+0x10) に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x30) != 0xA0 { // 10100000
                    t.Errorf("Memory at $30 = %#02x, want %#02x", c.ReadByteFromWRAM(0x30), 0xA0)
                }
            },
            expectedCarry: true,  // ビット0が1だったため
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROR Absolute",
            opcode:     0x6E,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x6E) // ROR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x41) // 0x4480に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x4480) != 0xA0 { // 10100000
                    t.Errorf("Memory at $4480 = %#02x, want %#02x", c.ReadByteFromWRAM(0x4480), 0xA0)
                }
            },
            expectedCarry: true,  // ビット0が1だったため
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
        {
            name:       "ROR Absolute,X",
            opcode:     0x7E,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.Registers.P.Carry = true // キャリーあり
                c.WriteByteToWRAM(c.Registers.PC, 0x7E) // ROR命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x41) // 0x4490 (0x4480+0x10) に値を設定 (01000001)
            },
            checkResult: func(t *testing.T, c *CPU) {
                if c.ReadByteFromWRAM(0x4490) != 0xA0 { // 10100000
                    t.Errorf("Memory at $4490 = %#02x, want %#02x", c.ReadByteFromWRAM(0x4490), 0xA0)
                }
            },
            expectedCarry: true,  // ビット0が1だったため
            expectedZero:  false,
            expectedNeg:   true,  // 最上位ビットが1
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            tt.checkResult(t, c)
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}


// MARK: レジスタ比較
// TestCMP はCMP命令（比較 A レジスタ）をテストします
func TestCMP(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedCarry bool  // 結果: A >= M
        expectedZero  bool  // 結果: A == M
        expectedNeg   bool  // 結果: 比較の最上位ビット
    }{
        {
            name:       "CMP Immediate - A > M",
            opcode:     0xC9,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xC9) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x30) // オペランド: 0x30
            },
            expectedCarry: true,  // A > M なのでキャリーがセット
            expectedZero:  false, // A != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
        {
            name:       "CMP Immediate - A == M",
            opcode:     0xC9,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xC9) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedCarry: true,  // A == M なのでキャリーがセット
            expectedZero:  true,  // A == M なのでゼロがセット
            expectedNeg:   false, // 比較結果は0
        },
        {
            name:       "CMP Immediate - A < M",
            opcode:     0xC9,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x30
                c.WriteByteToWRAM(c.Registers.PC, 0xC9) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedCarry: false, // A < M なのでキャリーはクリア
            expectedZero:  false, // A != M なのでゼロはクリア
            expectedNeg:   true,  // 比較結果は負数
        },
        {
            name:       "CMP Zero Page",
            opcode:     0xC5,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xC5) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x30) // 0x20に値を設定
            },
            expectedCarry: true,  // A > M なのでキャリーがセット
            expectedZero:  false, // A != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
        {
            name:       "CMP Zero Page,X",
            opcode:     0xD5,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xD5) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x42) // 0x30 (0x20+0x10) に値を設定
            },
            expectedCarry: true,  // A == M なのでキャリーがセット
            expectedZero:  true,  // A == M なのでゼロがセット
            expectedNeg:   false, // 比較結果は0
        },
        {
            name:       "CMP Absolute",
            opcode:     0xCD,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x30
                c.WriteByteToWRAM(c.Registers.PC, 0xCD) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x42) // 0x4480に値を設定
            },
            expectedCarry: false, // A < M なのでキャリーはクリア
            expectedZero:  false, // A != M なのでゼロはクリア
            expectedNeg:   true,  // 比較結果は負数
        },
        {
            name:       "CMP Absolute,X",
            opcode:     0xDD,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xDD) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x30) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedCarry: true,  // A > M なのでキャリーがセット
            expectedZero:  false, // A != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
        {
            name:       "CMP Absolute,Y",
            opcode:     0xD9,
            addrMode:   AbsoluteYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xD9) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x30) // 0x4490 (0x4480+0x10) に値を設定
            },
            expectedCarry: true,  // A > M なのでキャリーがセット
            expectedZero:  false, // A != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
        {
            name:       "CMP Indirect,X",
            opcode:     0xC1,
            addrMode:   IndirectXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.X = 0x04
                c.WriteByteToWRAM(c.Registers.PC, 0xC1) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x24, 0x74) // 0x24 (0x20+0x04) に低バイト
                c.WriteByteToWRAM(0x25, 0x20) // 0x25 に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2074, 0x42) // 0x2074に値を設定
            },
            expectedCarry: true,  // A == M なのでキャリーがセット
            expectedZero:  true,  // A == M なのでゼロがセット
            expectedNeg:   false, // 比較結果は0
        },
        {
            name:       "CMP Indirect,Y",
            opcode:     0xD1,
            addrMode:   IndirectYIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42
                c.Registers.Y = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xD1) // CMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x74) // 0x20に低バイト
                c.WriteByteToWRAM(0x21, 0x20) // 0x21に高バイト (→ 0x2074)
                c.WriteByteToWRAM(0x2084, 0x30) // 0x2084 (0x2074+0x10) に値を設定
            },
            expectedCarry: true,  // A > M なのでキャリーがセット
            expectedZero:  false, // A != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestCPX はCPX命令（比較 X レジスタ）をテストします
func TestCPX(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedCarry bool  // 結果: X >= M
        expectedZero  bool  // 結果: X == M
        expectedNeg   bool  // 結果: 比較の最上位ビット
    }{
        {
            name:       "CPX Immediate - X > M",
            opcode:     0xE0,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xE0) // CPX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x30) // オペランド: 0x30
            },
            expectedCarry: true,  // X > M なのでキャリーがセット
            expectedZero:  false, // X != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
        {
            name:       "CPX Immediate - X == M",
            opcode:     0xE0,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xE0) // CPX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedCarry: true,  // X == M なのでキャリーがセット
            expectedZero:  true,  // X == M なのでゼロがセット
            expectedNeg:   false, // 比較結果は0
        },
        {
            name:       "CPX Immediate - X < M",
            opcode:     0xE0,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x30
                c.WriteByteToWRAM(c.Registers.PC, 0xE0) // CPX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedCarry: false, // X < M なのでキャリーはクリア
            expectedZero:  false, // X != M なのでゼロはクリア
            expectedNeg:   true,  // 比較結果は負数
        },
        {
            name:       "CPX Zero Page",
            opcode:     0xE4,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xE4) // CPX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x30) // 0x20に値を設定
            },
            expectedCarry: true,  // X > M なのでキャリーがセット
            expectedZero:  false, // X != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
        {
            name:       "CPX Absolute",
            opcode:     0xEC,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x30
                c.WriteByteToWRAM(c.Registers.PC, 0xEC) // CPX命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x30) // 0x4480に値を設定
            },
            expectedCarry: true,  // X == M なのでキャリーがセット
            expectedZero:  true,  // X == M なのでゼロがセット
            expectedNeg:   false, // 比較結果は0
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestCPY はCPY命令（比較 Y レジスタ）をテストします
func TestCPY(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedCarry bool  // 結果: Y >= M
        expectedZero  bool  // 結果: Y == M
        expectedNeg   bool  // 結果: 比較の最上位ビット
    }{
        {
            name:       "CPY Immediate - Y > M",
            opcode:     0xC0,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xC0) // CPY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x30) // オペランド: 0x30
            },
            expectedCarry: true,  // Y > M なのでキャリーがセット
            expectedZero:  false, // Y != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
        {
            name:       "CPY Immediate - Y == M",
            opcode:     0xC0,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xC0) // CPY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedCarry: true,  // Y == M なのでキャリーがセット
            expectedZero:  true,  // Y == M なのでゼロがセット
            expectedNeg:   false, // 比較結果は0
        },
        {
            name:       "CPY Immediate - Y < M",
            opcode:     0xC0,
            addrMode:   Immediate,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x30
                c.WriteByteToWRAM(c.Registers.PC, 0xC0) // CPY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x42) // オペランド: 0x42
            },
            expectedCarry: false, // Y < M なのでキャリーはクリア
            expectedZero:  false, // Y != M なのでゼロはクリア
            expectedNeg:   true,  // 比較結果は負数
        },
        {
            name:       "CPY Zero Page",
            opcode:     0xC4,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xC4) // CPY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // オペランド: ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x30) // 0x20に値を設定
            },
            expectedCarry: true,  // Y > M なのでキャリーがセット
            expectedZero:  false, // Y != M なのでゼロはクリア
            expectedNeg:   false, // 比較結果は正数
        },
        {
            name:       "CPY Absolute",
            opcode:     0xCC,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x30
                c.WriteByteToWRAM(c.Registers.PC, 0xCC) // CPY命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // オペランド: 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // オペランド: 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x30) // 0x4480に値を設定
            },
            expectedCarry: true,  // Y == M なのでキャリーがセット
            expectedZero:  true,  // Y == M なのでゼロがセット
            expectedNeg:   false, // 比較結果は0
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}


// MARK: スタック操作
// TestPHA はPHA命令（Push Accumulator）をテストします
func TestPHA(t *testing.T) {
    tests := []struct {
        name        string
        opcode      uint8
        addrMode    AddressingMode
        setupCPU    func(*CPU)
        expectedSP  uint8
        stackValue  uint8
    }{
        {
            name:     "PHA - Push Accumulator to Stack",
            opcode:   0x48,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x42  // テスト用の値
                c.Registers.SP = 0xFF // スタックポインタを初期化
                c.WriteByteToWRAM(c.Registers.PC, 0x48) // PHA命令
            },
            expectedSP: 0xFE,  // スタックポインタが1つ減少
            stackValue: 0x42,  // スタックに格納される値
        },
        {
            name:     "PHA - Push negative value",
            opcode:   0x48,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.A = 0x80  // 負の値
                c.Registers.SP = 0xFF // スタックポインタを初期化
                c.WriteByteToWRAM(c.Registers.PC, 0x48) // PHA命令
            },
            expectedSP: 0xFE,  // スタックポインタが1つ減少
            stackValue: 0x80,  // スタックに格納される値
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
            
            // スタックの内容を検証（最後にプッシュした値）
            stackAddr := 0x0100 | uint16(c.Registers.SP+1)
            stackValue := c.ReadByteFromWRAM(stackAddr)
            if stackValue != tt.stackValue {
                t.Errorf("Stack value at $%04X = %#02x, want %#02x", stackAddr, stackValue, tt.stackValue)
            }
        })
    }
}

// TestPHP はPHP命令（Push Processor Status）をテストします
func TestPHP(t *testing.T) {
    tests := []struct {
        name        string
        opcode      uint8
        addrMode    AddressingMode
        setupCPU    func(*CPU)
        expectedSP  uint8
    }{
        {
            name:     "PHP - Push Processor Status to Stack",
            opcode:   0x08,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                // 特定のフラグセットを設定
                c.Registers.P.Carry = true
                c.Registers.P.Zero = true
                c.Registers.P.Negative = false
                c.Registers.P.Overflow = false
                c.Registers.P.Decimal = false
                c.Registers.P.Interrupt = false
                c.Registers.SP = 0xFF // スタックポインタを初期化
                c.WriteByteToWRAM(c.Registers.PC, 0x08) // PHP命令
            },
            expectedSP: 0xFE,  // スタックポインタが1つ減少
        },
        {
            name:     "PHP - All flags set",
            opcode:   0x08,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                // すべてのフラグをセット
                c.Registers.P.Carry = true
                c.Registers.P.Zero = true
                c.Registers.P.Interrupt = true
                c.Registers.P.Decimal = true
                c.Registers.P.Break = true
                c.Registers.P.Overflow = true
                c.Registers.P.Negative = true
                c.Registers.SP = 0xFF // スタックポインタを初期化
                c.WriteByteToWRAM(c.Registers.PC, 0x08) // PHP命令
            },
            expectedSP: 0xFE,  // スタックポインタが1つ減少
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // スタックにプッシュされる予定のステータスバイトを記録
            expectedStatus := c.Registers.P.ToByte()
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
            
            // スタックの内容を検証（最後にプッシュしたステータス）
            stackAddr := 0x0100 | uint16(c.Registers.SP+1)
            stackValue := c.ReadByteFromWRAM(stackAddr)
            if stackValue != expectedStatus {
                t.Errorf("Stack status at $%04X = %#02x, want %#02x", stackAddr, stackValue, expectedStatus)
            }
        })
    }
}

// TestPLA はPLA命令（Pull Accumulator）をテストします
func TestPLA(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedA     uint8
        expectedSP    uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:     "PLA - Pull Accumulator from Stack (positive value)",
            opcode:   0x68,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0xFE // スタックポインタを設定
                c.WriteByteToWRAM(0x01FF, 0x42) // スタックの次の位置に値を配置
                c.WriteByteToWRAM(c.Registers.PC, 0x68) // PLA命令
            },
            expectedA: 0x42,   // 取得した値
            expectedSP: 0xFF,  // スタックポインタが1つ増加
            expectedZero: false,
            expectedNeg: false,
        },
        {
            name:     "PLA - Pull Accumulator from Stack (zero value)",
            opcode:   0x68,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0xFE // スタックポインタを設定
                c.WriteByteToWRAM(0x01FF, 0x00) // スタックの次の位置にゼロを配置
                c.WriteByteToWRAM(c.Registers.PC, 0x68) // PLA命令
            },
            expectedA: 0x00,   // 取得した値
            expectedSP: 0xFF,  // スタックポインタが1つ増加
            expectedZero: true,
            expectedNeg: false,
        },
        {
            name:     "PLA - Pull Accumulator from Stack (negative value)",
            opcode:   0x68,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0xFE // スタックポインタを設定
                c.WriteByteToWRAM(0x01FF, 0x80) // スタックの次の位置に負の値を配置
                c.WriteByteToWRAM(c.Registers.PC, 0x68) // PLA命令
            },
            expectedA: 0x80,   // 取得した値
            expectedSP: 0xFF,  // スタックポインタが1つ増加
            expectedZero: false,
            expectedNeg: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "A", c.Registers.A, tt.expectedA)
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestPLP はPLP命令（Pull Processor Status）をテストします
func TestPLP(t *testing.T) {
    tests := []struct {
        name            string
        opcode          uint8
        addrMode        AddressingMode
        setupCPU        func(*CPU)
        expectedSP      uint8
        expectedCarry   bool
        expectedZero    bool
        expectedInt     bool
        expectedDecimal bool
        expectedBreak   bool
        expectedOverflow bool
        expectedNeg     bool
    }{
        {
            name:     "PLP - Pull Processor Status from Stack (mixed flags)",
            opcode:   0x28,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0xFE // スタックポインタを設定
                // ステータスレジスタのバイト表現: Carry=1, Zero=1, 他=0 (0x03)
                c.WriteByteToWRAM(0x01FF, 0x03)
                c.WriteByteToWRAM(c.Registers.PC, 0x28) // PLP命令
            },
            expectedSP: 0xFF,       // スタックポインタが1つ増加
            expectedCarry: true,
            expectedZero: true,
            expectedInt: false,
            expectedDecimal: false,
            expectedBreak: false,
            expectedOverflow: false,
            expectedNeg: false,
        },
        {
            name:     "PLP - Pull Processor Status from Stack (all flags)",
            opcode:   0x28,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0xFE // スタックポインタを設定
                // すべてのフラグがセット (0xFF)
                c.WriteByteToWRAM(0x01FF, 0xFF)
                c.WriteByteToWRAM(c.Registers.PC, 0x28) // PLP命令
            },
            expectedSP: 0xFF,       // スタックポインタが1つ増加
            expectedCarry: true,
            expectedZero: true,
            expectedInt: true,
            expectedDecimal: true,
            expectedBreak: true,
            expectedOverflow: true,
            expectedNeg: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Interrupt", c.Registers.P.Interrupt, tt.expectedInt)
            checkFlag(t, "Decimal", c.Registers.P.Decimal, tt.expectedDecimal)
            checkFlag(t, "Break", c.Registers.P.Break, tt.expectedBreak)
            checkFlag(t, "Overflow", c.Registers.P.Overflow, tt.expectedOverflow)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}


// MARK: インクリメント / デクリメント
// TestINX はINX命令（インクリメントX）をテストします
func TestINX(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedX     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "INX - Normal increment",
            opcode:     0xE8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xE8) // INX命令
            },
            expectedX:    0x43,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "INX - Overflow from 0xFF to 0x00",
            opcode:     0xE8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0xFF
                c.WriteByteToWRAM(c.Registers.PC, 0xE8) // INX命令
            },
            expectedX:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "INX - Negative result",
            opcode:     0xE8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x7F
                c.WriteByteToWRAM(c.Registers.PC, 0xE8) // INX命令
            },
            expectedX:    0x80,
            expectedZero: false,
            expectedNeg:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "X", c.Registers.X, tt.expectedX)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestINY はINY命令（インクリメントY）をテストします
func TestINY(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedY     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "INY - Normal increment",
            opcode:     0xC8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xC8) // INY命令
            },
            expectedY:    0x43,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "INY - Overflow from 0xFF to 0x00",
            opcode:     0xC8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0xFF
                c.WriteByteToWRAM(c.Registers.PC, 0xC8) // INY命令
            },
            expectedY:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "INY - Negative result",
            opcode:     0xC8,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x7F
                c.WriteByteToWRAM(c.Registers.PC, 0xC8) // INY命令
            },
            expectedY:    0x80,
            expectedZero: false,
            expectedNeg:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "Y", c.Registers.Y, tt.expectedY)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestDEX はDEX命令（デクリメントX）をテストします
func TestDEX(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedX     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "DEX - Normal decrement",
            opcode:     0xCA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0xCA) // DEX命令
            },
            expectedX:    0x41,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "DEX - Underflow from 0x00 to 0xFF",
            opcode:     0xCA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x00
                c.WriteByteToWRAM(c.Registers.PC, 0xCA) // DEX命令
            },
            expectedX:    0xFF,
            expectedZero: false,
            expectedNeg:  true,
        },
        {
            name:       "DEX - Decrement to zero",
            opcode:     0xCA,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x01
                c.WriteByteToWRAM(c.Registers.PC, 0xCA) // DEX命令
            },
            expectedX:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "X", c.Registers.X, tt.expectedX)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestDEY はDEY命令（デクリメントY）をテストします
func TestDEY(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        expectedY     uint8
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "DEY - Normal decrement",
            opcode:     0x88,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x42
                c.WriteByteToWRAM(c.Registers.PC, 0x88) // DEY命令
            },
            expectedY:    0x41,
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "DEY - Underflow from 0x00 to 0xFF",
            opcode:     0x88,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x00
                c.WriteByteToWRAM(c.Registers.PC, 0x88) // DEY命令
            },
            expectedY:    0xFF,
            expectedZero: false,
            expectedNeg:  true,
        },
        {
            name:       "DEY - Decrement to zero",
            opcode:     0x88,
            addrMode:   Implied,
            setupCPU: func(c *CPU) {
                c.Registers.Y = 0x01
                c.WriteByteToWRAM(c.Registers.PC, 0x88) // DEY命令
            },
            expectedY:    0x00,
            expectedZero: true,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            checkRegister(t, "Y", c.Registers.Y, tt.expectedY)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestINC はINC命令（インクリメントメモリ）をテストします
func TestINC(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        checkResult   func(*testing.T, *CPU)
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "INC Zero Page - Normal increment",
            opcode:     0xE6,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xE6) // INC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x42) // 0x20に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x20)
                if value != 0x43 {
                    t.Errorf("Memory at $20 = %#02x, want %#02x", value, 0x43)
                }
            },
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "INC Zero Page - Overflow from 0xFF to 0x00",
            opcode:     0xE6,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xE6) // INC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0xFF) // 0x20に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x20)
                if value != 0x00 {
                    t.Errorf("Memory at $20 = %#02x, want %#02x", value, 0x00)
                }
            },
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "INC Zero Page,X - Negative result",
            opcode:     0xF6,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xF6) // INC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x7F) // 0x30 (0x20+0x10) に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x30)
                if value != 0x80 {
                    t.Errorf("Memory at $30 = %#02x, want %#02x", value, 0x80)
                }
            },
            expectedZero: false,
            expectedNeg:  true,
        },
        {
            name:       "INC Absolute",
            opcode:     0xEE,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xEE) // INC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x42) // 0x4480に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x4480)
                if value != 0x43 {
                    t.Errorf("Memory at $4480 = %#02x, want %#02x", value, 0x43)
                }
            },
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "INC Absolute,X",
            opcode:     0xFE,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xFE) // INC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x42) // 0x4490 (0x4480+0x10) に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x4490)
                if value != 0x43 {
                    t.Errorf("Memory at $4490 = %#02x, want %#02x", value, 0x43)
                }
            },
            expectedZero: false,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            tt.checkResult(t, c)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestDEC はDEC命令（デクリメントメモリ）をテストします
func TestDEC(t *testing.T) {
    tests := []struct {
        name          string
        opcode        uint8
        addrMode      AddressingMode
        setupCPU      func(*CPU)
        checkResult   func(*testing.T, *CPU)
        expectedZero  bool
        expectedNeg   bool
    }{
        {
            name:       "DEC Zero Page - Normal decrement",
            opcode:     0xC6,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xC6) // DEC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x42) // 0x20に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x20)
                if value != 0x41 {
                    t.Errorf("Memory at $20 = %#02x, want %#02x", value, 0x41)
                }
            },
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "DEC Zero Page - Underflow from 0x00 to 0xFF",
            opcode:     0xC6,
            addrMode:   ZeroPage,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xC6) // DEC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // ZPアドレス0x20
                c.WriteByteToWRAM(0x20, 0x00) // 0x20に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x20)
                if value != 0xFF {
                    t.Errorf("Memory at $20 = %#02x, want %#02x", value, 0xFF)
                }
            },
            expectedZero: false,
            expectedNeg:  true,
        },
        {
            name:       "DEC Zero Page,X - Decrement to zero",
            opcode:     0xD6,
            addrMode:   ZeroPageXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xD6) // DEC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x20) // ZPアドレス0x20
                c.WriteByteToWRAM(0x30, 0x01) // 0x30 (0x20+0x10) に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x30)
                if value != 0x00 {
                    t.Errorf("Memory at $30 = %#02x, want %#02x", value, 0x00)
                }
            },
            expectedZero: true,
            expectedNeg:  false,
        },
        {
            name:       "DEC Absolute",
            opcode:     0xCE,
            addrMode:   Absolute,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0xCE) // DEC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // 高バイト (0x4480)
                c.WriteByteToWRAM(0x4480, 0x42) // 0x4480に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x4480)
                if value != 0x41 {
                    t.Errorf("Memory at $4480 = %#02x, want %#02x", value, 0x41)
                }
            },
            expectedZero: false,
            expectedNeg:  false,
        },
        {
            name:       "DEC Absolute,X",
            opcode:     0xDE,
            addrMode:   AbsoluteXIndexed,
            setupCPU: func(c *CPU) {
                c.Registers.X = 0x10
                c.WriteByteToWRAM(c.Registers.PC, 0xDE) // DEC命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // 高バイト (0x4480)
                c.WriteByteToWRAM(0x4490, 0x42) // 0x4490 (0x4480+0x10) に初期値設定
            },
            checkResult: func(t *testing.T, c *CPU) {
                value := c.ReadByteFromWRAM(0x4490)
                if value != 0x41 {
                    t.Errorf("Memory at $4490 = %#02x, want %#02x", value, 0x41)
                }
            },
            expectedZero: false,
            expectedNeg:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            tt.checkResult(t, c)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}


// MARK: ジャンプ
// TestJMP はJMP命令（ジャンプ）をテストします
func TestJMP(t *testing.T) {
    tests := []struct {
        name      string
        opcode    uint8
        addrMode  AddressingMode
        setupCPU  func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "JMP Absolute",
            opcode:   0x4C,
            addrMode: Absolute,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0x4C) // JMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x34) // 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x12) // 高バイト (0x1234)
            },
            expectedPC: 0x1234, // 指定されたアドレスに変更
        },
        {
            name:     "JMP Indirect",
            opcode:   0x6C,
            addrMode: Indirect,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0x6C) // JMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0x80) // 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // 高バイト (0x4480)
                
                // 間接アドレス先のジャンプ先
                c.WriteByteToWRAM(0x4480, 0x34) // 低バイト
                c.WriteByteToWRAM(0x4481, 0x12) // 高バイト (0x1234)
            },
            expectedPC: 0x1234, // 間接アドレスの指す位置に変更
        },
        {
            name:     "JMP Indirect - Page Boundary Bug",
            opcode:   0x6C,
            addrMode: Indirect,
            setupCPU: func(c *CPU) {
                c.WriteByteToWRAM(c.Registers.PC, 0x6C) // JMP命令
                c.WriteByteToWRAM(c.Registers.PC+1, 0xFF) // 低バイト
                c.WriteByteToWRAM(c.Registers.PC+2, 0x44) // 高バイト (0x44FF)
                
                // 間接アドレス先（ページ境界）
                c.WriteByteToWRAM(0x44FF, 0x34) // 低バイト
                // 高バイトは0x4500でなく、同ページ内の0x4400から取得される（バグ）
                c.WriteByteToWRAM(0x4400, 0x12) // 高バイト (0x1234)
            },
            expectedPC: 0x1234, // バグによる特殊なアドレス取得
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestJSR はJSR命令（サブルーチンにジャンプ）をテストします
func TestJSR(t *testing.T) {
    tests := []struct {
        name      string
        opcode    uint8
        addrMode  AddressingMode
        setupCPU  func(*CPU)
        expectedPC uint16
        expectedSP uint8
        checkStack func(*testing.T, *CPU)
    }{
        {
            name:     "JSR Absolute",
            opcode:   0x20,
            addrMode: Absolute,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.SP = 0xFF // スタックポインタ初期化
                c.WriteByteToWRAM(0x0200, 0x20) // JSR命令
                c.WriteByteToWRAM(0x0201, 0x34) // 低バイト
                c.WriteByteToWRAM(0x0202, 0x12) // 高バイト (0x1234)
            },
            expectedPC: 0x1234, // ジャンプ先アドレス
            expectedSP: 0xFD, // スタックポインタが2バイト分減少
            checkStack: func(t *testing.T, c *CPU) {
                // スタックにはPC-1の値（0x0202-1 = 0x0201）が格納されている
                highByte := c.ReadByteFromWRAM(0x01FF) // SP+2の位置
                lowByte := c.ReadByteFromWRAM(0x01FE)  // SP+1の位置
                returnAddr := uint16(highByte)<<8 | uint16(lowByte)
                if returnAddr != 0x0202 {
                    t.Errorf("Return address on stack = %#04x, want %#04x", returnAddr, 0x0202)
                }
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
            if tt.checkStack != nil {
                tt.checkStack(t, c)
            }
        })
    }
}

// TestRTS はRTS命令（サブルーチンから復帰）をテストします
func TestRTS(t *testing.T) {
    tests := []struct {
        name      string
        opcode    uint8
        addrMode  AddressingMode
        setupCPU  func(*CPU)
        expectedPC uint16
        expectedSP uint8
    }{
        {
            name:     "RTS - Return from Subroutine",
            opcode:   0x60,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0xFD // スタックポインタ設定
                
                // スタックに復帰先アドレス-1 (0x1234-1 = 0x1233) を設定
                c.WriteByteToWRAM(0x01FE, 0x33) // 低バイト
                c.WriteByteToWRAM(0x01FF, 0x12) // 高バイト
                
                c.WriteByteToWRAM(c.Registers.PC, 0x60) // RTS命令
            },
            expectedPC: 0x1234, // 復帰先アドレス（スタック値+1）
            expectedSP: 0xFF, // スタックポインタが2バイト分復元
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
        })
    }
}

// TestRTI はRTI命令（割り込みから復帰）をテストします
func TestRTI(t *testing.T) {
    tests := []struct {
        name           string
        opcode         uint8
        addrMode       AddressingMode
        setupCPU       func(*CPU)
        expectedPC     uint16
        expectedSP     uint8
        expectedCarry  bool
        expectedZero   bool
        expectedInt    bool
        expectedBreak  bool
        expectedDec    bool
        expectedOverflow bool
        expectedNeg    bool
    }{
        {
            name:     "RTI - Return from Interrupt",
            opcode:   0x40,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0xFC // スタックポインタ設定
                
                // スタックに処理状態とアドレスを設定
                // ステータスレジスタ（Carry=1, Zero=1, 他=0）
                c.WriteByteToWRAM(0x01FD, 0x03)
                // 復帰先アドレス (0x1234)
                c.WriteByteToWRAM(0x01FE, 0x34) // 低バイト
                c.WriteByteToWRAM(0x01FF, 0x12) // 高バイト
                
                c.WriteByteToWRAM(c.Registers.PC, 0x40) // RTI命令
            },
            expectedPC: 0x1234, // 復帰先アドレス
            expectedSP: 0xFF, // スタックポインタが3バイト分復元
            expectedCarry: true,
            expectedZero: true,
            expectedInt: false,
            expectedBreak: false,
            expectedDec: false,
            expectedOverflow: false,
            expectedNeg: false,
        },
        {
            name:     "RTI - All Flags Set",
            opcode:   0x40,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.SP = 0xFC // スタックポインタ設定
                
                // スタックに処理状態とアドレスを設定
                // すべてのフラグセット（但しBreakとReservedは無視される）
                c.WriteByteToWRAM(0x01FD, 0xFF)
                // 復帰先アドレス (0x4321)
                c.WriteByteToWRAM(0x01FE, 0x21) // 低バイト
                c.WriteByteToWRAM(0x01FF, 0x43) // 高バイト
                
                c.WriteByteToWRAM(c.Registers.PC, 0x40) // RTI命令
            },
            expectedPC: 0x4321, // 復帰先アドレス
            expectedSP: 0xFF, // スタックポインタが3バイト分復元
            expectedCarry: true,
            expectedZero: true,
            expectedInt: true,
            expectedBreak: false, // RTIではBフラグは無視される
            expectedDec: true,
            expectedOverflow: true,
            expectedNeg: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
            checkFlag(t, "Carry", c.Registers.P.Carry, tt.expectedCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, tt.expectedZero)
            checkFlag(t, "Interrupt", c.Registers.P.Interrupt, tt.expectedInt)
            checkFlag(t, "Decimal", c.Registers.P.Decimal, tt.expectedDec)
            checkFlag(t, "Break", c.Registers.P.Break, tt.expectedBreak)
            checkFlag(t, "Overflow", c.Registers.P.Overflow, tt.expectedOverflow)
            checkFlag(t, "Negative", c.Registers.P.Negative, tt.expectedNeg)
        })
    }
}

// TestBRK はBRK命令（強制割り込み）をテストします
func TestBRK(t *testing.T) {
    tests := []struct {
        name        string
        opcode      uint8
        addrMode    AddressingMode
        setupCPU    func(*CPU)
        expectedPC  uint16
        expectedSP  uint8
        checkStack  func(*testing.T, *CPU)
    }{
        {
            name:     "BRK - Force Interrupt",
            opcode:   0x00,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.SP = 0xFF // スタックポインタ初期化
                
                // 割り込みベクタを設定
                c.WriteByteToWRAM(0xFFFE, 0x34) // 低バイト (IRQとBRKは同じベクタを使用)
                c.WriteByteToWRAM(0xFFFF, 0x12) // 高バイト (0x1234)
                
                c.WriteByteToWRAM(0x0200, 0x00) // BRK命令
            },
            expectedPC: 0x1234, // 割り込みベクタのアドレスにジャンプ
            expectedSP: 0xFC, // スタックポインタが3バイト分減少 (PC[2bytes]+P[1byte])
            checkStack: func(t *testing.T, c *CPU) {
                // スタックにはPC+2の値（0x0200+2 = 0x0202）とステータスが格納されている
                statusByte := c.ReadByteFromWRAM(0x01FD)
                highByte := c.ReadByteFromWRAM(0x01FF)
                lowByte := c.ReadByteFromWRAM(0x01FE)
                returnAddr := uint16(highByte)<<8 | uint16(lowByte)
                
                if returnAddr != 0x0202 {
                    t.Errorf("Return address on stack = %#04x, want %#04x", returnAddr, 0x0202)
                }
                
                // BRKでプッシュされるステータスバイトはBフラグが立っているはず
                if statusByte&0x10 == 0 {
                    t.Errorf("Break flag not set in pushed status byte: %#02x", statusByte)
                }
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
            checkRegister(t, "SP", c.Registers.SP, tt.expectedSP)
            checkFlag(t, "Interrupt", c.Registers.P.Interrupt, true) // BRK後は割り込み禁止になる
            
            if tt.checkStack != nil {
                tt.checkStack(t, c)
            }
        })
    }
}

// TestNOP はNOP命令（何もしない）をテストします
func TestNOP(t *testing.T) {
    tests := []struct {
        name      string
        opcode    uint8
        addrMode  AddressingMode
        setupCPU  func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "NOP - No Operation",
            opcode:   0xEA,
            addrMode: Implied,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.WriteByteToWRAM(0x0200, 0xEA) // NOP命令
            },
            expectedPC: 0x0201, // PCが1バイト進むだけ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // 各レジスタの初期値を保存
            oldA := c.Registers.A
            oldX := c.Registers.X
            oldY := c.Registers.Y
            oldSP := c.Registers.SP
            oldCarry := c.Registers.P.Carry
            oldZero := c.Registers.P.Zero
            oldInterrupt := c.Registers.P.Interrupt
            oldDecimal := c.Registers.P.Decimal
            oldBreak := c.Registers.P.Break
            oldOverflow := c.Registers.P.Overflow
            oldNegative := c.Registers.P.Negative
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // PC以外のレジスタが変わっていないことを検証
            checkRegister(t, "A", c.Registers.A, oldA)
            checkRegister(t, "X", c.Registers.X, oldX)
            checkRegister(t, "Y", c.Registers.Y, oldY)
            checkRegister(t, "SP", c.Registers.SP, oldSP)
            checkFlag(t, "Carry", c.Registers.P.Carry, oldCarry)
            checkFlag(t, "Zero", c.Registers.P.Zero, oldZero)
            checkFlag(t, "Interrupt", c.Registers.P.Interrupt, oldInterrupt)
            checkFlag(t, "Decimal", c.Registers.P.Decimal, oldDecimal)
            checkFlag(t, "Break", c.Registers.P.Break, oldBreak)
            checkFlag(t, "Overflow", c.Registers.P.Overflow, oldOverflow)
            checkFlag(t, "Negative", c.Registers.P.Negative, oldNegative)
            
            // PCだけが進んでいることを検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestBMI はBMI命令（Branch if Minus）をテストします
func TestBMI(t *testing.T) {
    tests := []struct {
        name       string
        opcode     uint8
        addrMode   AddressingMode
        setupCPU   func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "BMI - Branch taken (negative flag set, positive offset)",
            opcode:   0x30,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Negative = true // 負数フラグをセット
                c.WriteByteToWRAM(0x0200, 0x30) // BMI命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0212, // 0x0202 + 0x10 = 0x0212
        },
        {
            name:     "BMI - Branch taken (negative flag set, negative offset)",
            opcode:   0x30,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Negative = true // 負数フラグをセット
                c.WriteByteToWRAM(0x0200, 0x30) // BMI命令
                c.WriteByteToWRAM(0x0201, 0xF0) // オフセット: -16 (2の補数表現)
            },
            expectedPC: 0x01F2, // 0x0202 + 0xF0 (符号拡張で-16) = 0x01F2
        },
        {
            name:     "BMI - Branch not taken (negative flag clear)",
            opcode:   0x30,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Negative = false // 負数フラグをクリア
                c.WriteByteToWRAM(0x0200, 0x30) // BMI命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0202, // 分岐が行われないので、PC+2のみ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestBPL はBPL命令（Branch if Plus）をテストします
func TestBPL(t *testing.T) {
    tests := []struct {
        name       string
        opcode     uint8
        addrMode   AddressingMode
        setupCPU   func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "BPL - Branch taken (negative flag clear, positive offset)",
            opcode:   0x10,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Negative = false // 負数フラグをクリア
                c.WriteByteToWRAM(0x0200, 0x10) // BPL命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0212, // 0x0202 + 0x10 = 0x0212
        },
        {
            name:     "BPL - Branch taken (negative flag clear, negative offset)",
            opcode:   0x10,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Negative = false // 負数フラグをクリア
                c.WriteByteToWRAM(0x0200, 0x10) // BPL命令
                c.WriteByteToWRAM(0x0201, 0xF0) // オフセット: -16 (2の補数表現)
            },
            expectedPC: 0x01F2, // 0x0202 + 0xF0 (符号拡張で-16) = 0x01F2
        },
        {
            name:     "BPL - Branch not taken (negative flag set)",
            opcode:   0x10,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Negative = true // 負数フラグをセット
                c.WriteByteToWRAM(0x0200, 0x10) // BPL命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0202, // 分岐が行われないので、PC+2のみ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestBVS はBVS命令（Branch if Overflow Set）をテストします
func TestBVS(t *testing.T) {
    tests := []struct {
        name       string
        opcode     uint8
        addrMode   AddressingMode
        setupCPU   func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "BVS - Branch taken (overflow flag set, positive offset)",
            opcode:   0x70,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Overflow = true // オーバーフローフラグをセット
                c.WriteByteToWRAM(0x0200, 0x70) // BVS命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0212, // 0x0202 + 0x10 = 0x0212
        },
        {
            name:     "BVS - Branch taken (overflow flag set, negative offset)",
            opcode:   0x70,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Overflow = true // オーバーフローフラグをセット
                c.WriteByteToWRAM(0x0200, 0x70) // BVS命令
                c.WriteByteToWRAM(0x0201, 0xF0) // オフセット: -16 (2の補数表現)
            },
            expectedPC: 0x01F2, // 0x0202 + 0xF0 (符号拡張で-16) = 0x01F2
        },
        {
            name:     "BVS - Branch not taken (overflow flag clear)",
            opcode:   0x70,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Overflow = false // オーバーフローフラグをクリア
                c.WriteByteToWRAM(0x0200, 0x70) // BVS命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0202, // 分岐が行われないので、PC+2のみ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestBVC はBVC命令（Branch if Overflow Clear）をテストします
func TestBVC(t *testing.T) {
    tests := []struct {
        name       string
        opcode     uint8
        addrMode   AddressingMode
        setupCPU   func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "BVC - Branch taken (overflow flag clear, positive offset)",
            opcode:   0x50,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Overflow = false // オーバーフローフラグをクリア
                c.WriteByteToWRAM(0x0200, 0x50) // BVC命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0212, // 0x0202 + 0x10 = 0x0212
        },
        {
            name:     "BVC - Branch taken (overflow flag clear, negative offset)",
            opcode:   0x50,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Overflow = false // オーバーフローフラグをクリア
                c.WriteByteToWRAM(0x0200, 0x50) // BVC命令
                c.WriteByteToWRAM(0x0201, 0xF0) // オフセット: -16 (2の補数表現)
            },
            expectedPC: 0x01F2, // 0x0202 + 0xF0 (符号拡張で-16) = 0x01F2
        },
        {
            name:     "BVC - Branch not taken (overflow flag set)",
            opcode:   0x50,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Overflow = true // オーバーフローフラグをセット
                c.WriteByteToWRAM(0x0200, 0x50) // BVC命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0202, // 分岐が行われないので、PC+2のみ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestBCS はBCS命令（Branch if Carry Set）をテストします
func TestBCS(t *testing.T) {
    tests := []struct {
        name       string
        opcode     uint8
        addrMode   AddressingMode
        setupCPU   func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "BCS - Branch taken (carry flag set, positive offset)",
            opcode:   0xB0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Carry = true // キャリーフラグをセット
                c.WriteByteToWRAM(0x0200, 0xB0) // BCS命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0212, // 0x0202 + 0x10 = 0x0212
        },
        {
            name:     "BCS - Branch taken (carry flag set, negative offset)",
            opcode:   0xB0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Carry = true // キャリーフラグをセット
                c.WriteByteToWRAM(0x0200, 0xB0) // BCS命令
                c.WriteByteToWRAM(0x0201, 0xF0) // オフセット: -16 (2の補数表現)
            },
            expectedPC: 0x01F2, // 0x0202 + 0xF0 (符号拡張で-16) = 0x01F2
        },
        {
            name:     "BCS - Branch not taken (carry flag clear)",
            opcode:   0xB0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Carry = false // キャリーフラグをクリア
                c.WriteByteToWRAM(0x0200, 0xB0) // BCS命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0202, // 分岐が行われないので、PC+2のみ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestBCC はBCC命令（Branch if Carry Clear）をテストします
func TestBCC(t *testing.T) {
    tests := []struct {
        name       string
        opcode     uint8
        addrMode   AddressingMode
        setupCPU   func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "BCC - Branch taken (carry flag clear, positive offset)",
            opcode:   0x90,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Carry = false // キャリーフラグをクリア
                c.WriteByteToWRAM(0x0200, 0x90) // BCC命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0212, // 0x0202 + 0x10 = 0x0212
        },
        {
            name:     "BCC - Branch taken (carry flag clear, negative offset)",
            opcode:   0x90,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Carry = false // キャリーフラグをクリア
                c.WriteByteToWRAM(0x0200, 0x90) // BCC命令
                c.WriteByteToWRAM(0x0201, 0xF0) // オフセット: -16 (2の補数表現)
            },
            expectedPC: 0x01F2, // 0x0202 + 0xF0 (符号拡張で-16) = 0x01F2
        },
        {
            name:     "BCC - Branch not taken (carry flag set)",
            opcode:   0x90,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Carry = true // キャリーフラグをセット
                c.WriteByteToWRAM(0x0200, 0x90) // BCC命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0202, // 分岐が行われないので、PC+2のみ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestBEQ はBEQ命令（Branch if Equal）をテストします
func TestBEQ(t *testing.T) {
    tests := []struct {
        name       string
        opcode     uint8
        addrMode   AddressingMode
        setupCPU   func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "BEQ - Branch taken (zero flag set, positive offset)",
            opcode:   0xF0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Zero = true // ゼロフラグをセット
                c.WriteByteToWRAM(0x0200, 0xF0) // BEQ命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0212, // 0x0202 + 0x10 = 0x0212
        },
        {
            name:     "BEQ - Branch taken (zero flag set, negative offset)",
            opcode:   0xF0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Zero = true // ゼロフラグをセット
                c.WriteByteToWRAM(0x0200, 0xF0) // BEQ命令
                c.WriteByteToWRAM(0x0201, 0xF0) // オフセット: -16 (2の補数表現)
            },
            expectedPC: 0x01F2, // 0x0202 + 0xF0 (符号拡張で-16) = 0x01F2
        },
        {
            name:     "BEQ - Branch not taken (zero flag clear)",
            opcode:   0xF0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Zero = false // ゼロフラグをクリア
                c.WriteByteToWRAM(0x0200, 0xF0) // BEQ命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0202, // 分岐が行われないので、PC+2のみ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}

// TestBNE はBNE命令（Branch if Not Equal）をテストします
func TestBNE(t *testing.T) {
    tests := []struct {
        name       string
        opcode     uint8
        addrMode   AddressingMode
        setupCPU   func(*CPU)
        expectedPC uint16
    }{
        {
            name:     "BNE - Branch taken (zero flag clear, positive offset)",
            opcode:   0xD0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Zero = false // ゼロフラグをクリア
                c.WriteByteToWRAM(0x0200, 0xD0) // BNE命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0212, // 0x0202 + 0x10 = 0x0212
        },
        {
            name:     "BNE - Branch taken (zero flag clear, negative offset)",
            opcode:   0xD0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Zero = false // ゼロフラグをクリア
                c.WriteByteToWRAM(0x0200, 0xD0) // BNE命令
                c.WriteByteToWRAM(0x0201, 0xF0) // オフセット: -16 (2の補数表現)
            },
            expectedPC: 0x01F2, // 0x0202 + 0xF0 (符号拡張で-16) = 0x01F2
        },
        {
            name:     "BNE - Branch not taken (zero flag set)",
            opcode:   0xD0,
            addrMode: Relative,
            setupCPU: func(c *CPU) {
                c.Registers.PC = 0x0200
                c.Registers.P.Zero = true // ゼロフラグをセット
                c.WriteByteToWRAM(0x0200, 0xD0) // BNE命令
                c.WriteByteToWRAM(0x0201, 0x10) // オフセット: +16
            },
            expectedPC: 0x0202, // 分岐が行われないので、PC+2のみ
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := setupCPU()
            tt.setupCPU(c)
            
            // CPU実行サイクルを使用して命令を実行
            c.Execute()

            // 結果を検証
            if c.Registers.PC != tt.expectedPC {
                t.Errorf("PC = %#04x, want %#04x", c.Registers.PC, tt.expectedPC)
            }
        })
    }
}