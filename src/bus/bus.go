package bus

import (
	"Famicom-emulator/cartridge"
)

const (
	CPU_WRAM_SIZE  = 2 * 1024 // 2kB
	CPU_WRAM_START = 0x0000
	CPU_WRAM_END   = 0x1FFF

	PPU_REG_START  = 0x2000
	PPU_REG_END    = 0x3FFF
)

type Bus struct {
	wram [CPU_WRAM_SIZE+1]uint8
	cartridge cartridge.Cartridge
}


// MARK: Busの初期化メソッド
func (b *Bus) Init() {
	for addr := range b.wram {
		b.wram[addr] = 0x00
	}
}

func (b *Bus) InitWithCartridge(cartridge *cartridge.Cartridge) {
	for addr := range b.wram {
		b.wram[addr] = 0x00
	}
	b.cartridge = *cartridge
}


// MARK: WRAMの読み取り/書き込み
func (b *Bus) ReadByteFrom(address uint16) uint8 {
	switch {
	case CPU_WRAM_START <= address && address <= CPU_WRAM_END:
		ptr := address & 0b00000111_11111111 // 11bitにマスク
		return b.wram[ptr]
	case PPU_REG_START <= address && address <= PPU_REG_END:
		// ptr := address & 0b00100000_00000111
		// fmt.Printf("READ (PPU): $04%X\n", ptr)
		return 0x0000
	case 0x8000 <= address && address <= 0xFFFF:
		return b.ReadProgramROM(address)
	default:
		// fmt.Printf("READ (out of bounds): $%04X\n", address)
		return 0x0000
	}
}

func (b *Bus) ReadWordFrom(address uint16) uint16 {
	lower := b.ReadByteFrom(address)
	upper := b.ReadByteFrom(address + 1)

	return uint16(upper) << 8 | uint16(lower)
}

func (b *Bus) WriteByteAt(address uint16, data uint8) {
	switch {
	case CPU_WRAM_START <= address && address <= CPU_WRAM_END:
		ptr := address & 0b00000111_11111111 // 11bitにマスク
		b.wram[ptr] = data
	case PPU_REG_START <= address && address <= PPU_REG_END:
		// ptr := address & 0b00100000_00000111
		// fmt.Printf("READ (PPU): $%04X\n", ptr)
	case 0x8000 <= address && address <= 0xFFFF:
		// fmt.Printf("READ (CART): $%04X\n", address)
	default:
		// fmt.Printf("READ (out of bounds): $%04X\n", address)
	}
}

func (b *Bus) WriteWordAt(address uint16, data uint16) {
	upper := uint8(data >> 8)
	lower := uint8(data & 0xFF)
	b.WriteByteAt(address, lower)
	b.WriteByteAt(address + 1, upper)
}

func (b *Bus) ReadProgramROM(address uint16) uint8 {
	// fmt.Printf("READ PRG: $%04X -> $%04X\n", address, address - 0x8000)
	// fmt.Println(b.cartridge)
	address -= 0x8000
	if len(b.cartridge.ProgramROM) == 0x4000 && address >= 0x4000 {
		address %= 0x4000
	}
	// fmt.Printf("CARTRIDGE: $%04X\n", address)
	return b.cartridge.ProgramROM[address]
}