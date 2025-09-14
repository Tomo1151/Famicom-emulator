package mappers

const (
	PRG_RAM_SIZE uint = 8 * 1024 // 8kB
)

// MARK: MMC1 SxROM (マッパー1) の定義
type SxROM struct {
	shiftRegister uint8
	shiftCount    uint8

	control  uint8
	chrBank0 uint8
	chrBank1 uint8
	prgBank  uint8

	IsCharacterRAM bool
	ProgramROM     []uint8
	CharacterROM   []uint8
	ProgramRAM     [PRG_RAM_SIZE]uint8
}

// MARK: マッパーの初期化
func (s *SxROM) Init(rom []uint8) {
	s.shiftRegister = 0x10
	s.shiftCount = 0

	s.control = 0x0C
	s.chrBank0 = 0
	s.chrBank1 = 0
	s.prgBank = 0

	programROM, characterROM := GetROMs(rom)
	s.IsCharacterRAM = GetCharacterROMSize(rom) == 0
	s.ProgramROM = programROM
	s.CharacterROM = characterROM

	// プログラムRAMの初期化
	for i := range s.ProgramRAM {
		s.ProgramRAM[i] = 0xFF
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
func (s *SxROM) ReadProgramROM(address uint16) uint8 {
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
	bankMax := uint(len(s.ProgramROM)) / BANK_SIZE

	romBaseAddress := uint(address - PRG_ROM_START)

	switch (s.control & 0x0C) >> 2 {
	case 0, 1:
		// バンク番号の下位ビットを無視，32KBを$8000~に割り当て
		bank := s.prgBank & 0x1E
		return s.ProgramROM[romBaseAddress+(BANK_SIZE*uint(bank))]
	case 2:
		// 最初のバンクを$8000~に固定，16KBバンクを$C000~に割り当て
		bank := s.prgBank & 0x1F

		switch {
		case PRG_ROM_START <= address && address <= 0xBFFF:
			return s.ProgramROM[romBaseAddress]
		case 0xC000 <= address && address <= PRG_ROM_END:
			return s.ProgramROM[uint(address-0xC000)+(BANK_SIZE*uint(bank))]
		default:
			panic("Error: unexpected program rom bank mode")
		}
	case 3:
		// 最後のバンクを$C000~に固定，16KBバンクを$8000~に割り当て
		bank := s.prgBank & 0x1F

		switch {
		case PRG_ROM_START <= address && address <= 0xBFFF:
			return s.ProgramROM[romBaseAddress+(BANK_SIZE*uint(bank))]
		case 0xC000 <= address && address <= PRG_ROM_END:
			return s.ProgramROM[uint(address-0xC000)+(BANK_SIZE*(bankMax-1))]
		default:
			panic("Error: unexpected program rom bank mode")
		}
	default:
		panic("Error: unexpected program rom bank mode")
	}
}

// MARK: キャラクタROMのアドレス計算
func (s *SxROM) getCharacterROMAddress(address uint16) uint16 {
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
func (s *SxROM) ReadCharacterROM(address uint16) uint8 {
	return s.CharacterROM[s.getCharacterROMAddress(address)]
}

// MARK: キャラクタROMへの書き込み
func (s *SxROM) WriteToCharacterROM(address uint16, data uint8) {
	s.CharacterROM[s.getCharacterROMAddress(address)] = data
}

// MARK: プログラムRAMの読み取り
func (s *SxROM) ReadProgramRAM(address uint16) uint8 {
	return s.ProgramROM[address]
}

// MARK: プログラムRAMへの書き込み
func (s *SxROM) WriteToProgramRAM(address uint16, data uint8) {
	s.ProgramRAM[address-0x6000] = data
}

// MARK: シフトレジスタのリセット
func (s *SxROM) resetShiftRegister() {
	s.shiftRegister = 0x10
	s.shiftCount = 0
}

// MARK: ミラーリングの取得
func (s *SxROM) GetMirroring() Mirroring {
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
func (s *SxROM) GetIsCharacterRAM() bool {
	return s.IsCharacterRAM
}

// MARK: プログラムROMの取得
func (s *SxROM) GetProgramROM() []uint8 {
	return s.ProgramROM
}

// MARK: キャラクタROMの取得
func (s *SxROM) GetCharacterROM() []uint8 {
	return s.CharacterROM
}

// MARK: マッパー名の取得
func (s *SxROM) GetMapperInfo() string {
	return "MMC1 SxROM (Mapper 1)"
}