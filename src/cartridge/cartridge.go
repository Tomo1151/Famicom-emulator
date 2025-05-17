package cartridge

import (
	"errors"
	"log"
	"reflect"
)

type Cartridge struct {
	ProgramROM []uint8
	CharacterROM []uint8
	Mapper uint8
	ScreenMirroring Mirroring
}

type Mirroring uint8

const (
	VERTICAL Mirroring = iota
	HORIZONTAL
	FOUR_SCREEN
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

	mapper := (raw[7] & 0xF0) | (raw[6] >> 4)
	iNESVer := (raw[7] >> 2) & 0b11
	if iNESVer != 0 {
		log.Fatalf("NES2.0 format is not supported")
		return errors.New("Unsupported iNES version")
	}

	isFourScreen := (raw[6] & 0b1000) != 0
	isVertical := (raw[6] & 0b0001) != 0
	
	var mirroring Mirroring
	
	if isFourScreen {
		mirroring = FOUR_SCREEN
	} else if isVertical {
		mirroring = VERTICAL
	} else {
		mirroring = HORIZONTAL
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

	// fmt.Printf("PRG_ROM_START: %04X, SIZE: %d\n", prgROMStart, prgROMSize)
	// fmt.Printf("CHR_ROM_START: %04X, SIZE: %d\n", chrROMStart, chrROMSize)

	c.ProgramROM = raw[prgROMStart:(prgROMStart+prgROMSize)]
	c.CharacterROM = raw[chrROMStart:(chrROMStart+chrROMSize)]
	c.Mapper = mapper
	c.ScreenMirroring = mirroring

	// fmt.Printf("PRG: %v\n", c.ProgramROM[0:8])
	// fmt.Printf("CHR: %v\n", c.CharacterROM[0:8])
	// fmt.Printf("Map: %v\n", c.Mapper)
	// fmt.Printf("Mir: %v\n", c.ScreenMirroring)

	return nil
}