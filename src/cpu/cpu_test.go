package cpu

import "testing"

func TestREPL_LDA_TAX(t *testing.T) {
	c := CreateCPU(false)
	// LDA #$42, TAX, BRK
	commands := []uint8{0xA9, 0x42, 0xAA, 0x00}
	c.REPL(commands)
	if c.Registers.A != 0x42 {
		t.Errorf("A = 0x%02X; want 0x42", c.Registers.A)
	}
	if c.Registers.X != 0x42 {
		t.Errorf("X = 0x%02X; want 0x42", c.Registers.X)
	}
}

func TestREPL_ZeroFlag(t *testing.T) {
	c := CreateCPU(false)
	// LDA #$00, BRK
	commands := []uint8{0xA9, 0x00, 0x00}
	c.REPL(commands)
	if !c.Registers.P.Zero {
		t.Error("Zero flag = false; want true")
	}
}

func TestREPL_NegativeFlag(t *testing.T) {
	c := CreateCPU(false)
	// LDA #$FF, BRK
	commands := []uint8{0xA9, 0xFF, 0x00}
	c.REPL(commands)
	if !c.Registers.P.Negative {
		t.Error("Negative flag = false; want true")
	}
}