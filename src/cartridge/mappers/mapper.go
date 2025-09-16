package mappers

const (
	BANK_SIZE         uint = 16 * 1024 // 16kB
	PRG_ROM_PAGE_SIZE uint = 16 * 1024 // 16kB
	CHR_ROM_PAGE_SIZE uint = 8 * 1024  // 8kB
	PRG_RAM_SIZE      uint = 8 * 1024  // 8kB

	PRG_ROM_START uint16 = 0x8000
	PRG_ROM_END   uint16 = 0xFFFF
	PRG_RAM_START uint16 = 0x6000
)

type Mirroring uint8

const (
	MIRRORING_VERTICAL Mirroring = iota
	MIRRORING_HORIZONTAL
	MIRRORING_FOUR_SCREEN
)

// MARK: マッパーのインターフェース
type Mapper interface {
	Init([]uint8)
	ReadProgramROM(uint16) uint8
	ReadCharacterROM(uint16) uint8
	ReadProgramRAM(uint16) uint8
	WriteToCharacterROM(uint16, uint8)
	WriteToProgramRAM(uint16, uint8)
	Write(uint16, uint8)

	GenerateScanlineIRQ(uint16, bool)
	GetIRQ() bool

	GetMapperInfo() string
	GetIsCharacterRAM() bool
	GetMirroring() Mirroring
	GetProgramROM() []uint8
	GetCharacterROM() []uint8
}

// MARK: カートリッジのバイナリからプログラムROMとキャラクタROMを取得
func GetROMs(rom []uint8) ([]uint8, []uint8) {
	// それぞれのROMのアドレスとサイズを計算
	programROMStart := GetProgramROMStartAddress(rom)
	programmROMSize := GetProgramROMSize(rom)
	characterROMStart := GetCharacterROMStartAddress(programROMStart, programmROMSize)
	characterROMSize := GetCharacterROMSize(rom)

	var programROM []uint8
	var characterROM []uint8

	programROM = rom[programROMStart:(programROMStart + programmROMSize)]
	if characterROMSize == 0 {
		characterROM = make([]uint8, CHR_ROM_PAGE_SIZE)
	} else {
		characterROM = rom[characterROMStart:(characterROMStart + characterROMSize)]
	}

	return programROM, characterROM
}

// MARK: シンプルなミラーリングの取得
func GetSimpleMirroring(rom []uint8) Mirroring {
	isFourScreen := (rom[6] & 0b1000) != 0
	isVertical := (rom[6] & 0b0001) != 0

	var mirroring Mirroring

	if isFourScreen {
		mirroring = MIRRORING_FOUR_SCREEN
	} else if isVertical {
		mirroring = MIRRORING_VERTICAL
	} else {
		mirroring = MIRRORING_HORIZONTAL
	}

	return mirroring
}

// MARK: カートリッジのバイナリからプログラムROMのスタートアドレスを取得
func GetProgramROMStartAddress(rom []uint8) uint {
	skipTrainer := (rom[6] & 0b100) != 0
	var trainerOffset uint
	if skipTrainer {
		trainerOffset = 512
	} else {
		trainerOffset = 0
	}
	return 16 + trainerOffset
}

// MARK: カートリッジのバイナリからプログラムROMのサイズを取得
func GetProgramROMSize(rom []uint8) uint {
	return uint(rom[4]) * PRG_ROM_PAGE_SIZE
}

// MARK: カートリッジのバイナリからキャラクタROMのスタートアドレスを取得
func GetCharacterROMStartAddress(programROMStart uint, programROMSize uint) uint {
	return programROMStart + programROMSize
}

// MARK: カートリッジのバイナリからキャラクタROMのサイズを取得
func GetCharacterROMSize(rom []uint8) uint {
	return uint(rom[5]) * CHR_ROM_PAGE_SIZE
}