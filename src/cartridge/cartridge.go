package cartridge

type Cartridge struct {
	ProgramROM []uint8
	CharacterROM []uint8
	Mapper uint8
	ScreenMirroring Mirroring
}

type Mirroring uint8

const (
	VERTICAL uint8 = iota
	HORIZONTAL
	FOUR_SCREEN
)

func (c *Cartridge) Load(raw []uint8) {
	
}