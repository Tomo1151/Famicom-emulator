package cartridge

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"Famicom-emulator/cartridge/mappers"
)

type Cartridge struct {
	Mapper mappers.Mapper
}

type Mirroring uint8

const (
	PRG_ROM_PAGE_SIZE uint = 16 * 1024 // 16kB
	CHR_ROM_PAGE_SIZE uint = 8 * 1024  // 8kB

	SAVE_DATA_DIR = "../rom/saves/"
)

// カートリッジ先頭のiNESタグ
var NES_TAG = []uint8{0x4E, 0x45, 0x53, 0x1A}

// MARK: カートリッジの読み込み
func (c *Cartridge) Load(filename string) error {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filepath.Base(filename), ext)

	// ゲームROMの読み込み
	gamefile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error occured in 'os.ReadFile()'")
		return fmt.Errorf("Couldn't read file: %s", filename)
	}

	// savesディレクトリがなければ作成
	if _, err := os.Stat(SAVE_DATA_DIR); os.IsNotExist(err) {
		os.Mkdir(SAVE_DATA_DIR, 0755)
	}

	// セーブデータの読み込み
	savefile, err := os.ReadFile(SAVE_DATA_DIR + name + ".save")
	if err == nil {
		fmt.Println("Save data loaded")
	} else {
		fmt.Println("No save data found")
		savefile = []byte{}
	}

	// NESタグの検証
	if !reflect.DeepEqual(gamefile[0:4], NES_TAG) {
		log.Fatalf("Error: invalid cartridge header '%v'", gamefile[0:4])
		return errors.New("Invalid cartridge header")
	}

	// iNESヘッダとマッパーの検証
	mapperNo := (gamefile[7] & 0xF0) | (gamefile[6] >> 4)
	iNESVer := (gamefile[7] >> 2) & 0b11
	if iNESVer != 0 {
		log.Fatalf("NES2.0 format is not supported")
		return errors.New("Unsupported iNES version")
	}

	// マッパーオブジェクトを生成・設定
	rom := c.GetMapper(mapperNo)
	rom.Init(name, gamefile, savefile)
	c.Mapper = rom
	c.DumpInfo()

	return nil
}

// MARK: マッパーオブジェクトの取得
func (c *Cartridge) GetMapper(mapperNo uint8) mappers.Mapper {
	switch mapperNo {
	case 0x00:
		return &mappers.NROM{}
	case 0x01:
		return &mappers.SxROM{}
	case 0x02:
		return &mappers.UxROM{}
	case 0x03:
		return &mappers.CNROM{}
	case 0x04:
		return &mappers.TxROM{}
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
