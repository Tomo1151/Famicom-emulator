package mappers

const (
	PRG_RAM_SIZE uint = 8 * 1024 // 8kB
)

// MARK: MMC1 (マッパー1) の定義
type MMC1 struct {
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
func (m *MMC1) Init(rom []uint8) {
	m.shiftRegister = 0x10
	m.shiftCount = 0

	m.control = 0x0C
	m.chrBank0 = 0
	m.chrBank1 = 0
	m.prgBank = 0

	programROM, characterROM := GetROMs(rom)
	m.IsCharacterRAM = GetCharacterROMSize(rom) == 0
	m.ProgramROM = programROM
	m.CharacterROM = characterROM

	// プログラムRAMの初期化
	for i := range m.ProgramRAM {
		m.ProgramRAM[i] = 0xFF
	}
}

// MARK: ROMスペースへの書き込み
func (m *MMC1) Write(address uint16, data uint8) {
	if data&0x80 != 0 {
		m.resetShiftRegister()
		return
	}

	// 5bitのシフトレジスタの最上位ビットにdataを入れる
	m.shiftRegister >>= 1
	m.shiftRegister |= ((data & 0x01) << 4)
	m.shiftCount++

	// 5回目の書き込み時のみアドレスによって返すバンクの変更を行う (前4回はその準備)
	if m.shiftCount == 5 {
		switch {
		case PRG_ROM_START <= address && address <= 0x9FFF:
			m.control = m.shiftRegister
		case 0xA000 <= address && address <= 0xBFFF:
			m.chrBank0 = m.shiftRegister
		case 0xC000 <= address && address <= 0xDFFF:
			m.chrBank1 = m.shiftRegister
		case 0xE000 <= address && address <= PRG_ROM_END:
			m.prgBank = m.shiftRegister
		}
		m.resetShiftRegister()
	}
}

// MARK: プログラムROMの読み取り
func (m *MMC1) ReadProgramROM(address uint16) uint8 {
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
	bankMax := uint(len(m.ProgramROM)) / BANK_SIZE

	romBaseAddress := uint(address - PRG_ROM_START)

	switch (m.control & 0x0C) >> 2 {
	case 0, 1:
		// バンク番号の下位ビットを無視，32KBを$8000~に割り当て
		bank := m.prgBank & 0x1E
		return m.ProgramROM[romBaseAddress+(BANK_SIZE*uint(bank))]
	case 2:
		// 最初のバンクを$8000~に固定，16KBバンクを$C000~に割り当て
		bank := m.prgBank & 0x1F

		switch {
		case PRG_ROM_START <= address && address <= 0xBFFF:
			return m.ProgramROM[romBaseAddress]
		case 0xC000 <= address && address <= PRG_ROM_END:
			return m.ProgramROM[uint(address-0xC000)+(BANK_SIZE*uint(bank))]
		default:
			panic("Error: unexpected program rom bank mode")
		}
	case 3:
		// 最後のバンクを$C000~に固定，16KBバンクを$8000~に割り当て
		bank := m.prgBank & 0x1F

		switch {
		case PRG_ROM_START <= address && address <= 0xBFFF:
			return m.ProgramROM[romBaseAddress+(BANK_SIZE*uint(bank))]
		case 0xC000 <= address && address <= PRG_ROM_END:
			return m.ProgramROM[uint(address-0xC000)+(BANK_SIZE*(bankMax-1))]
		default:
			panic("Error: unexpected program rom bank mode")
		}
	default:
		panic("Error: unexpected program rom bank mode")
	}
}

// MARK: キャラクタROMのアドレス計算
func (m *MMC1) getCharacterROMAddress(address uint16) uint16 {
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

	switch (m.control & 0x10) >> 4 {
	case 0:
		// 一度に8Bを切り替え
		bank := uint16(m.chrBank0 & 0x1F)
		return address + (uint16(BANK_SIZE) * bank)
	case 1:
		// 二つの別々の4KBバンクを割り当て
		bank := uint16(m.chrBank0 & 0x1F)

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
func (m *MMC1) ReadCharacterROM(address uint16) uint8 {
	return m.CharacterROM[m.getCharacterROMAddress(address)]
}

// MARK: キャラクタROMへの書き込み
func (m *MMC1) WriteToCharacterROM(address uint16, data uint8) {
	m.CharacterROM[m.getCharacterROMAddress(address)] = data
}

// MARK: プログラムRAMの読み取り
func (m *MMC1) ReadProgramRAM(address uint16) uint8 {
	return m.ProgramROM[address]
}

// MARK: プログラムRAMへの書き込み
func (m *MMC1) WriteToProgramRAM(address uint16, data uint8) {
	m.ProgramRAM[address-0x6000] = data
}

// MARK: シフトレジスタのリセット
func (m *MMC1) resetShiftRegister() {
	m.shiftRegister = 0x10
	m.shiftCount = 0
}

// MARK: ミラーリングの取得
func (m *MMC1) GetMirroring() Mirroring {
	switch m.control & 0x03 {
	case 2:
		return MIRRORING_VERTICAL
	case 3:
		return MIRRORING_HORIZONTAL
	default:
		panic("Error: unsupported mirroring mode")
	}
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (m *MMC1) GetIsCharacterRAM() bool {
	return m.IsCharacterRAM
}

// MARK: プログラムROMの取得
func (m *MMC1) GetProgramROM() []uint8 {
	return m.ProgramROM
}

// MARK: キャラクタROMの取得
func (m *MMC1) GetCharacterROM() []uint8 {
	return m.CharacterROM
}

// MARK: マッパー名の取得
func (m *MMC1) GetMapperInfo() string {
	return "MMC1 (Mapper 1)"
}