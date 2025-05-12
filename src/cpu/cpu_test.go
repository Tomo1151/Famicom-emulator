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