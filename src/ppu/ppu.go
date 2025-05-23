package ppu

import "Famicom-emulator/cartridge"

const (
	VRAM_SIZE uint16 = 2 * 1024 // 2kB
	PALETTE_TABLE_SIZE uint8 = 32
	OAM_DATA_SIZE uint16 = 64 * 4
)

type PPU struct {
	CHR_ROM []uint8
	PaletteTable [PALETTE_TABLE_SIZE+1]uint8
	vram [VRAM_SIZE+1]uint8
	oam [OAM_DATA_SIZE+1]uint8
	Mirroring cartridge.Mirroring

	addrRegister AddrRegister
}

func (p *PPU) Init(chr_rom []uint8, mirroring cartridge.Mirroring){
	p.CHR_ROM = chr_rom
	p.Mirroring = mirroring
	for addr := range p.vram { p.vram[addr] = 0x00 }
	for addr := range p.oam { p.oam[addr] = 0x00 }
	for addr := range p.PaletteTable { p.PaletteTable[addr] = 0x00 }
	p.addrRegister.Init()
}

func (p *PPU) WriteToPPUAddrRegister(value uint8) {
	p.addrRegister.update(value)
}