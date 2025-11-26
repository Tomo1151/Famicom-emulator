package mappers

import (
	"fmt"
	"os"
)

// MARK: MMC1 SxROM (マッパー1) の定義
type SxROM struct {
	name string

	shiftRegister uint8
	shiftCount    uint8

	control  uint8
	chrBank0 uint8
	chrBank1 uint8
	prgBank  uint8

	isCharacterRam bool
	programRom     []uint8
	characterRom   []uint8
	programRam     [PRG_RAM_SIZE]uint8
}

// MARK: マッパーの初期化
func (s *SxROM) Init(name string, rom []uint8, save []uint8) {
	s.name = name

	s.shiftRegister = 0x10
	s.shiftCount = 0

	s.control = 0x0C
	s.chrBank0 = 0
	s.chrBank1 = 0
	s.prgBank = 0

	programRom, characterRom := roms(rom)
	s.isCharacterRam = characterRomSize(rom) == 0
	s.programRom = programRom
	s.characterRom = characterRom

	// プログラムRAMの初期化
	for i := range s.programRam {
		s.programRam[i] = 0xFF
	}

	// セーブデータの読み込み
	if len(save) != 0 {
		copy(s.programRam[:], save)
	}
}

// MARK: ROMスペースへの書き込み
func (s *SxROM) Write(address uint16, data uint8) {
	if data&0x80 != 0 {
		s.resetShiftRegister()
		return
	}

	// 5bitのシフトレジスタの最上位ビットにdataを入れる
	s.shiftRegister >>= 1
	s.shiftRegister |= ((data & 0x01) << 4)
	s.shiftCount++

	// 5回目の書き込み時のみアドレスによって返すバンクの変更を行う (前4回はその準備)
	if s.shiftCount == 5 {
		switch {
		case PRG_ROM_START <= address && address <= 0x9FFF:
			s.control = s.shiftRegister
		case 0xA000 <= address && address <= 0xBFFF:
			s.chrBank0 = s.shiftRegister
		case 0xC000 <= address && address <= 0xDFFF:
			s.chrBank1 = s.shiftRegister
		case 0xE000 <= address && address <= PRG_ROM_END:
			s.prgBank = s.shiftRegister
		}
		s.resetShiftRegister()
	}
}

// MARK: プログラムROMの読み取り
func (s *SxROM) ReadProgramRom(address uint16) uint8 {
	/*
		4bit0
		-----
		CPPMM
		|||||
		|||++- ミラーリング: (0: 1画面, 下位バンク / 1: 1画面, 上位バンク
		|||               2: 垂直
		|||               3: 水平
		|++--- プログラムROM バンクモード (0 / 1: バンク番号の下位ビットを無視，32KBを$8000~に割り当て
		|                         2: 最初のバンクを$8000~に固定，16KBバンクを$C000~に割り当て
		|                         3: 最後のバンクを$C000~に固定，16KBバンクを$8000~に割り当て
		+----- キャラクタROM バンクモード (0: 一度に8KBを切り替え / 1: 2角別々の4KBバンクを割り当て
	*/

	// 最後のバンク番号
	bankMax := uint(len(s.programRom)) / BANK_SIZE

	romBaseAddress := uint(address - PRG_ROM_START)

	switch (s.control & 0x0C) >> 2 {
	case 0, 1:
		// バンク番号の下位ビットを無視，32KBを$8000~に割り当て
		bank := s.prgBank & 0x1E
		return s.programRom[romBaseAddress+(BANK_SIZE*uint(bank))]
	case 2:
		// 最初のバンクを$8000~に固定，16KBバンクを$C000~に割り当て
		bank := s.prgBank & 0x1F

		switch {
		case PRG_ROM_START <= address && address <= 0xBFFF:
			return s.programRom[romBaseAddress]
		case 0xC000 <= address && address <= PRG_ROM_END:
			return s.programRom[uint(address-0xC000)+(BANK_SIZE*uint(bank))]
		default:
			panic("Error: unexpected program rom bank mode")
		}
	case 3:
		// 最後のバンクを$C000~に固定，16KBバンクを$8000~に割り当て
		bank := s.prgBank & 0x1F

		switch {
		case PRG_ROM_START <= address && address <= 0xBFFF:
			return s.programRom[romBaseAddress+(BANK_SIZE*uint(bank))]
		case 0xC000 <= address && address <= PRG_ROM_END:
			return s.programRom[uint(address-0xC000)+(BANK_SIZE*(bankMax-1))]
		default:
			panic("Error: unexpected program rom bank mode")
		}
	default:
		panic("Error: unexpected program rom bank mode")
	}
}

// MARK: キャラクタROMのアドレス計算
func (s *SxROM) calcCharacterRomAddress(address uint16) uint16 {
	/*
		4bit0
		-----
		CPPMM
		|||||
		|||++- ミラーリング: (0: 1画面, 下位バンク / 1: 1画面, 上位バンク
		|||               2: 垂直
		|||               3: 水平
		|++--- プログラムROM バンクモード (0 / 1: バンク番号の下位ビットを無視，32KBを$8000~に割り当て
		|                         2: 最初のバンクを$8000~に固定，16KBバンクを$C000~に割り当て
		|                         3: 最後のバンクを$C000~に固定，16KBバンクを$8000~に割り当て
		+----- キャラクタROM バンクモード (0: 一度に8KBを切り替え / 1: 2角別々の4KBバンクを割り当て
	*/

	switch (s.control & 0x10) >> 4 {
	case 0:
		// 一度に8Bを切り替え
		bank := uint16(s.chrBank0 & 0x1F)
		return address + (uint16(BANK_SIZE) * bank)
	case 1:
		// 二つの別々の4KBバンクを割り当て
		bank := uint16(s.chrBank0 & 0x1F)

		switch {
		case address <= 0x0000 && address <= 0x0FFF:
			return address + (uint16(BANK_SIZE) * bank)
		case address <= 0x1000 && address <= 0x1FFF:
			return address - 0x1000 + (uint16(BANK_SIZE) * bank)
		default:
			panic("Error: unexpected character rom bank mode")
		}
	default:
		panic("Error: unexpected character rom bank mode")
	}
}

// MARK: キャラクタROMの読み取り
func (s *SxROM) ReadCharacterRom(address uint16) uint8 {
	return s.characterRom[s.calcCharacterRomAddress(address)]
}

// MARK: キャラクタROMへの書き込み
func (s *SxROM) WriteToCharacterRom(address uint16, data uint8) {
	s.characterRom[s.calcCharacterRomAddress(address)] = data
}

// MARK: プログラムRAMの読み取り
func (s *SxROM) ReadProgramRam(address uint16) uint8 {
	return s.programRam[address-PRG_RAM_START]
}

// MARK: プログラムRAMへの書き込み
func (s *SxROM) WriteToProgramRam(address uint16, data uint8) {
	s.programRam[address-PRG_RAM_START] = data

	// セーブデータの書き出し
	os.WriteFile(SAVE_DATA_DIR+s.name+".save", s.programRam[:], 0644)
}

// MARK: セーブデータの書き出し
func (s *SxROM) Save() {
	err := os.WriteFile(SAVE_DATA_DIR+s.name+".save", s.programRam[:], 0644)
	if err != nil {
		fmt.Printf("Error saving game data: %v\n", err)
	} else {
		fmt.Printf("Game saved to: %s\n", SAVE_DATA_DIR+s.name+".save")
	}
}

// MARK: シフトレジスタのリセット
func (s *SxROM) resetShiftRegister() {
	s.shiftRegister = 0x10
	s.shiftCount = 0
}

// MARK: スキャンラインによってIRQを発生させる
func (s *SxROM) GenerateScanlineIRQ(scanline uint16, backgroundEnable bool) {}

// MARK: IRQ状態の取得
func (s *SxROM) IRQ() bool { return false }

// MARK: ミラーリングの取得
func (s *SxROM) Mirroring() Mirroring {
	switch s.control & 0x03 {
	case 2:
		return MIRRORING_VERTICAL
	case 3:
		return MIRRORING_HORIZONTAL
	default:
		panic("Error: unsupported mirroring mode")
	}
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (s *SxROM) IsCharacterRam() bool {
	return s.isCharacterRam
}

// MARK: プログラムROMの取得
func (s *SxROM) ProgramRom() []uint8 {
	return s.programRom
}

// MARK: キャラクタROMの取得
func (s *SxROM) CharacterRom() []uint8 {
	return s.characterRom
}

// MARK: マッパー名の取得
func (s *SxROM) MapperInfo() string {
	return "MMC1 SxROM (Mapper 1)"
}

// MARK: マッパーのシャローコピーの取得
func (s *SxROM) Clone() Mapper {
	copy := *s
	return &copy
}
