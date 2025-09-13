package cartridge

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"Famicom-emulator/cartridge/mappers"
)

type Cartridge struct {
	IsCHRRAM bool
	CharacterROM []uint8
	Mapper mappers.Mapper
	ScreenMirroring Mirroring
}

type Mirroring uint8

const (
	MIRRORING_VERTICAL Mirroring = iota
	MIRRORING_HORIZONTAL
	MIRRORING_FOUR_SCREEN
)

const (
	PRG_ROM_PAGE_SIZE uint = 16 * 1024 // 16kB
	CHR_ROM_PAGE_SIZE uint = 8 * 1024 // 8kB
)

var	NES_TAG = []uint8{0x4E, 0x45, 0x53, 0x1A}


func (c *Cartridge) Load(raw []uint8) error {
	if !reflect.DeepEqual(raw[0:4], NES_TAG) {
		log.Fatalf("Error: invalid cartridge header '%v'", raw[0:4])
		return errors.New("Invalid cartridge header")
	}

	mapperNo := (raw[7] & 0xF0) | (raw[6] >> 4)
	iNESVer := (raw[7] >> 2) & 0b11
	if iNESVer != 0 {
		log.Fatalf("NES2.0 format is not supported")
		return errors.New("Unsupported iNES version")
	}

	isFourScreen := (raw[6] & 0b1000) != 0
	isMIRRORING_VERTICAL := (raw[6] & 0b0001) != 0
	
	var mirroring Mirroring
	
	if isFourScreen {
		mirroring = MIRRORING_FOUR_SCREEN
	} else if isMIRRORING_VERTICAL {
		mirroring = MIRRORING_VERTICAL
	} else {
		mirroring = MIRRORING_HORIZONTAL
	}

	prgROMSize := uint16(raw[4]) * uint16(PRG_ROM_PAGE_SIZE)
	chrROMSize := uint16(raw[5]) * uint16(CHR_ROM_PAGE_SIZE)

	skipTrainer := (raw[6] & 0b100) != 0

	var trainerOffset uint16
	if skipTrainer {
		trainerOffset = 512
	} else {
		trainerOffset = 0
	}

	prgROMStart := 16 + trainerOffset
	chrROMStart := prgROMStart + prgROMSize
	var chr_rom []uint8

	if chrROMSize == 0 {
		chr_rom = make([]uint8, CHR_ROM_PAGE_SIZE)
	} else {
		chr_rom = raw[chrROMStart:(chrROMStart+chrROMSize)]
	}

	// fmt.Printf("PRG_ROM_START: %04X, SIZE: %d\n", prgROMStart, prgROMSize)
	// fmt.Printf("CHR_ROM_START: %04X, SIZE: %d\n", chrROMStart, chrROMSize)

	rom := c.GetMapper(mapperNo)
	rom.Init(raw[prgROMStart:(prgROMStart+prgROMSize)])

	c.IsCHRRAM = chrROMSize == 0
	c.CharacterROM = chr_rom
	c.Mapper = rom
	c.ScreenMirroring = mirroring

	c.DumpInfo()

	return nil
}

func (c *Cartridge) GetMapper(mapperNo uint8) mappers.Mapper {
	switch mapperNo {
	case 0:
		return &mappers.NROM{}
	case 1:
		return &mappers.UxROM{}
	default:
		return &mappers.NROM{}
	}
}

func (c *Cartridge) DumpInfo() {
	fmt.Printf("Cartridge loaded:\n")
	fmt.Printf("  Mapper: %s\n", c.Mapper.GetMapperInfo())
	fmt.Printf("  PRG ROM Size: %d bytes\n", len(c.Mapper.GetProgramROM()))
	fmt.Printf("  CHR ROM Size: %d bytes\n", len(c.CharacterROM))
	fmt.Printf("  CHR RAM: %v\n", c.IsCHRRAM)
	var mirroringStr string
	switch c.ScreenMirroring {
	case MIRRORING_VERTICAL:
		mirroringStr = "Vertical"
	case MIRRORING_HORIZONTAL:
		mirroringStr = "Horizontal"
	case MIRRORING_FOUR_SCREEN:
		mirroringStr = "Four Screen"
	default:
		mirroringStr = "Unknown"
	}
	fmt.Printf("  Mirroring: %s\n", mirroringStr)
}