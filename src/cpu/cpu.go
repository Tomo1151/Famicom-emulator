package cpu

import "fmt"

// CPUの定義
type CPU struct {
	Registers registers
}

// CPUの作成関数
func CreateCPU() CPU {
	cpu := CPU{}
	cpu.Init()
	return cpu
}

// CPUの初期化メソッド
func (c *CPU) Init() {
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
}

// 命令の実行
func (c * CPU) execute(opecode uint8) {
	switch opecode {
	default:
		fmt.Println(opecode)
	}
}