package bus

import (
	"Famicom-emulator/cartridge"
	"Famicom-emulator/ppu"
	"fmt"
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
	ppu ppu.PPU
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
	b.ppu = ppu.PPU{}
	b.ppu.Init(b.cartridge.CharacterROM, b.cartridge.ScreenMirroring)
}


// MARK: WRAMの読み取り/書き込み
func (b *Bus) ReadByteFrom(address uint16) uint8 {
	switch {
	case CPU_WRAM_START <= address && address <= CPU_WRAM_END:
		ptr := address & 0b00000111_11111111 // 11bitにマスク
		return b.wram[ptr]
	case
			address == 0x2000 ||
			address == 0x2001 ||
			address == 0x2003 ||
			address == 0x2005 ||
			address == 0x2006 ||
			address == 0x4014:
		panic(fmt.Sprintf("Error: attempt to read from write only ppu address $%04X", address))
	case address == 0x2007:
		b.ppu.ReadVRAM()
	case 0x2008 <= address && address <= PPU_REG_END:
		ptr := address & 0b00100000_00000111
		return b.ReadByteFrom(ptr)
		// fmt.Printf("READ (PPU): $04%X\n", ptr)
	case 0x8000 <= address:
		return b.ReadProgramROM(address)
	default:
		fmt.Printf("Ignoring memory access at $%04X", address)
		return 0x00
		// panic(fmt.Sprintf("Error: illegal memory read $%04X\n", address))
	}
	return 0x00
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
	case address == 0x2000:
		b.ppu.WriteToPPUControlRegister(data)
	case address == 0x2006:
		b.ppu.WriteToPPUAddrRegister(data)
	// case address == 0x2007:
		// b.ppu.WriteToData(data)
	case 0x2008 <= address && address <= PPU_REG_END:
		ptr := address & 0b00100000_00000111
		b.WriteByteAt(ptr, data)
		// fmt.Printf("READ (PPU): $%04X\n", ptr)
	case 0x8000 <= address:
		panic(fmt.Sprintf("Error: attempt to write to cartridge ROM space $%04X, 0x%02X\n", address, data))
	default:
		fmt.Printf("Ignoring memory write to $%04X", address)
		// panic(fmt.Sprintf("Error: illegal memory write $%04X, 0x%02X\n", address, data))
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