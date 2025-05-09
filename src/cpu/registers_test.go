package cpu

import "testing"

func TestStatusRegister_ToByte(t *testing.T) {
	tests := []struct {
		name string
		sr   statusRegister
		want uint8
	}{
		{
			name: "no flags",
			sr:   statusRegister{},
			want: 0x00,
		},
		{
			name: "carry",
			sr:   statusRegister{Carry: true},
			want: 1 << STATUS_REG_CARRY_POS,
		},
		{
			name: "zero",
			sr:   statusRegister{Zero: true},
			want: 1 << STATUS_REG_ZERO_POS,
		},
		{
			name: "interrupt",
			sr:   statusRegister{Interrupt: true},
			want: 1 << STATUS_REG_INTERRUPT_POS,
		},
		{
			name: "decimal",
			sr:   statusRegister{Decimal: true},
			want: 1 << STATUS_REG_DECIMAL_POS,
		},
		{
			name: "break",
			sr:   statusRegister{Break: true},
			want: 1 << STATUS_REG_BREAK_POS,
		},
		{
			name: "reserved",
			sr:   statusRegister{Reserved: true},
			want: 1 << STATUS_REG_RESERVED_POS,
		},
		{
			name: "overflow",
			sr:   statusRegister{Overflow: true},
			want: 1 << STATUS_REG_OVERFLOW_POS,
		},
		{
			name: "negative",
			sr:   statusRegister{Negative: true},
			want: 1 << STATUS_REG_NEGATIVE_POS,
		},
		{
			name: "all flags",
			sr: statusRegister{
				Carry:     true,
				Zero:      true,
				Interrupt: true,
				Decimal:   true,
				Break:     true,
				Reserved:  true,
				Overflow:  true,
				Negative:  true,
			},
			want: 0xFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sr.ToByte()
			if got != tt.want {
				t.Errorf("ToByte() = 0x%02X, want 0x%02X", got, tt.want)
			}
		})
	}
}

// TestStatusRegister_FromByte はFromByteメソッドをテストします
func TestStatusRegister_FromByte(t *testing.T) {
	tests := []struct {
		name  string
		input uint8
		want  statusRegister
	}{
		{
			name:  "no flags",
			input: 0x00,
			want:  statusRegister{},
		},
		{
			name:  "carry only",
			input: 1 << STATUS_REG_CARRY_POS,
			want:  statusRegister{Carry: true},
		},
		{
			name:  "zero only",
			input: 1 << STATUS_REG_ZERO_POS,
			want:  statusRegister{Zero: true},
		},
		{
			name:  "interrupt only",
			input: 1 << STATUS_REG_INTERRUPT_POS,
			want:  statusRegister{Interrupt: true},
		},
		{
			name:  "decimal only",
			input: 1 << STATUS_REG_DECIMAL_POS,
			want:  statusRegister{Decimal: true},
		},
		{
			name:  "break only",
			input: 1 << STATUS_REG_BREAK_POS,
			want:  statusRegister{Break: true},
		},
		{
			name:  "reserved only",
			input: 1 << STATUS_REG_RESERVED_POS,
			want:  statusRegister{Reserved: true},
		},
		{
			name:  "overflow only",
			input: 1 << STATUS_REG_OVERFLOW_POS,
			want:  statusRegister{Overflow: true},
		},
		{
			name:  "negative only",
			input: 1 << STATUS_REG_NEGATIVE_POS,
			want:  statusRegister{Negative: true},
		},
		// 組み合わせパターン
		{
			name:  "negative and zero",
			input: (1 << STATUS_REG_NEGATIVE_POS) | (1 << STATUS_REG_ZERO_POS),
			want:  statusRegister{Negative: true, Zero: true},
		},
		{
			name:  "carry, zero and interrupt",
			input: (1 << STATUS_REG_CARRY_POS) | (1 << STATUS_REG_ZERO_POS) | (1 << STATUS_REG_INTERRUPT_POS),
			want:  statusRegister{Carry: true, Zero: true, Interrupt: true},
		},
		{
			name:  "common arithmetic result (negative with carry)",
			input: (1 << STATUS_REG_NEGATIVE_POS) | (1 << STATUS_REG_CARRY_POS),
			want:  statusRegister{Negative: true, Carry: true},
		},
		{
			name:  "all flags",
			input: 0xFF,
			want: statusRegister{
				Carry:     true,
				Zero:      true,
				Interrupt: true,
				Decimal:   true,
				Break:     true,
				Reserved:  true,
				Overflow:  true,
				Negative:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got statusRegister
			got.FromByte(tt.input)

			// 各フラグを個別に検証
			if got.Negative != tt.want.Negative {
				t.Errorf("FromByte(%#02x).Negative = %v, want %v", tt.input, got.Negative, tt.want.Negative)
			}
			if got.Overflow != tt.want.Overflow {
				t.Errorf("FromByte(%#02x).Overflow = %v, want %v", tt.input, got.Overflow, tt.want.Overflow)
			}
			if got.Reserved != tt.want.Reserved {
				t.Errorf("FromByte(%#02x).Reserved = %v, want %v", tt.input, got.Reserved, tt.want.Reserved)
			}
			if got.Break != tt.want.Break {
				t.Errorf("FromByte(%#02x).Break = %v, want %v", tt.input, got.Break, tt.want.Break)
			}
			if got.Decimal != tt.want.Decimal {
				t.Errorf("FromByte(%#02x).Decimal = %v, want %v", tt.input, got.Decimal, tt.want.Decimal)
			}
			if got.Interrupt != tt.want.Interrupt {
				t.Errorf("FromByte(%#02x).Interrupt = %v, want %v", tt.input, got.Interrupt, tt.want.Interrupt)
			}
			if got.Zero != tt.want.Zero {
				t.Errorf("FromByte(%#02x).Zero = %v, want %v", tt.input, got.Zero, tt.want.Zero)
			}
			if got.Carry != tt.want.Carry {
				t.Errorf("FromByte(%#02x).Carry = %v, want %v", tt.input, got.Carry, tt.want.Carry)
			}
		})
	}
}

// TestStatusRegister_RoundTrip はToByte→FromByteの一貫性をテストします
func TestStatusRegister_RoundTrip(t *testing.T) {
	testCases := []statusRegister{
		{}, // 全てfalse
		{Carry: true},
		{Zero: true, Carry: true},
		{Interrupt: true, Break: true},
		{Decimal: true, Overflow: true, Negative: true},
		{ // 全てtrue
			Carry:     true,
			Zero:      true,
			Interrupt: true,
			Decimal:   true,
			Break:     true,
			Reserved:  true,
			Overflow:  true,
			Negative:  true,
		},
	}

	for i, original := range testCases {
		byteValue := original.ToByte()
		var reconverted statusRegister
		reconverted.FromByte(byteValue)

		// 元のステータスと再変換後のステータスが一致するか検証
		if original.Negative != reconverted.Negative ||
			original.Overflow != reconverted.Overflow ||
			original.Reserved != reconverted.Reserved ||
			original.Break != reconverted.Break ||
			original.Decimal != reconverted.Decimal ||
			original.Interrupt != reconverted.Interrupt ||
			original.Zero != reconverted.Zero ||
			original.Carry != reconverted.Carry {
			t.Errorf("Case %d: Round trip conversion failed. Original: %+v, After: %+v, Byte: 0x%02X",
				i, original, reconverted, byteValue)
		}
	}
}