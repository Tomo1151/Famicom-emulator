package ppu

const (
	VRAM_SIZE uint16 = 2048
	PALETTE_TABLE_SIZE uint8 = 32
	OAM_DATA_SIZE uint16 = 256
)

type PPU struct {
	CHR_ROM []uint8
	PaletteTable [PALETTE_TABLE_SIZE+1]uint8
	vram [VRAM_SIZE+1]uint8
	
}
