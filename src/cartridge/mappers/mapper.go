package mappers

const (
	BANK_SIZE     uint   = 16 * 1024
	PRG_ROM_START uint16 = 0x8000
	PRG_ROM_END   uint16 = 0xFFFF
)

type Mapper interface {
	Init([]uint8)
	ReadProgramROM(uint16) uint8
	Write(uint16, uint8)

	GetMapperInfo() string
	GetProgramROM() []uint8
}
