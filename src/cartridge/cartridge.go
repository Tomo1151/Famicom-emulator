package cartridge

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"Famicom-emulator/cartridge/mappers"
)

type Cartridge struct {
	Mapper mappers.Mapper
}

type Mirroring uint8

const (
	PRG_ROM_PAGE_SIZE uint = 16 * 1024 // 16kB
	CHR_ROM_PAGE_SIZE uint = 8 * 1024 // 8kB
)

// カートリッジ先頭のiNESタグ
var	NES_TAG = []uint8{0x4E, 0x45, 0x53, 0x1A}

// MARK: カートリッジの読み込み
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

	// MARK: マッパーオブジェクトを生成・設定
	rom := c.GetMapper(mapperNo)
	rom.Init(raw)
	c.Mapper = rom
	c.DumpInfo()

	return nil
}

// MARK: マッパーオブジェクトの取得
func (c *Cartridge) GetMapper(mapperNo uint8) mappers.Mapper {
	switch mapperNo {
	case 0:
		return &mappers.NROM{}
	case 1:
		return &mappers.SxROM{}
	case 2:
		return &mappers.UxROM{}
	case 3:
		return &mappers.CNROM{}
	default:
		return &mappers.NROM{}
	}
}

// MARK: カートリッジの情報を出力
func (c *Cartridge) DumpInfo() {
	fmt.Printf("Cartridge loaded:\n")
	fmt.Printf("  Mapper: %s\n", c.Mapper.GetMapperInfo())
	fmt.Printf("  PRG ROM Size: %d bytes\n", len(c.Mapper.GetProgramROM()))
	fmt.Printf("  CHR ROM Size: %d bytes\n", len(c.Mapper.GetCharacterROM()))
	fmt.Printf("  CHR RAM: %v\n", c.Mapper.GetIsCharacterRAM())
	var mirroringStr string
	switch c.Mapper.GetMirroring() {
	case mappers.MIRRORING_VERTICAL:
		mirroringStr = "Vertical"
	case mappers.MIRRORING_HORIZONTAL:
		mirroringStr = "Horizontal"
	case mappers.MIRRORING_FOUR_SCREEN:
		mirroringStr = "Four Screen"
	default:
		mirroringStr = "Unknown"
	}
	fmt.Printf("  Mirroring: %s\n", mirroringStr)
}