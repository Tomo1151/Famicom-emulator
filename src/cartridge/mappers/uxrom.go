package mappers

import "fmt"

// MARK: UxROM (マッパー2) の定義
type UxROM struct {
	bank       uint8

	IsCharacterRAM bool
	Mirroring Mirroring
	ProgramROM []uint8
	CharacterROM []uint8
}

// MARK: マッパーの初期化
func (u *UxROM) Init(rom []uint8) {
	u.bank = 0x00

	programROM, characterROM := GetROMs(rom)
	u.IsCharacterRAM = GetCharacterROMSize(rom) == 0
	u.Mirroring = GetSimpleMirroring(rom)
	u.ProgramROM = programROM
	u.CharacterROM = characterROM
}

// MARK: ROMスペースへの書き込み
func (u *UxROM) Write(address uint16, data uint8) {
	u.bank = data
}

// MARK: プログラムROMの読み取り
func (u *UxROM) ReadProgramROM(address uint16) uint8 {
	// 最後のバンク番号
	bankMax := uint(len(u.ProgramROM)) / BANK_SIZE

	switch {
	case PRG_ROM_START <= address && address <= 0xBFFF:
		// 前半部分はバンク選択
		bank := uint(u.bank & 0x0F)
		return u.ProgramROM[uint(address)-0x8000+BANK_SIZE*bank]
	case 0xC000 <= address && address <= PRG_ROM_END:
		// 後半部分は固定
		return u.ProgramROM[uint(address)-0xC000+BANK_SIZE*(bankMax-1)]
	default:
		panic(fmt.Sprintf("Erorr: unexpected PRG ROM space: %04X", address))
	}
}

// MARK: キャラクタROMの読み取り
func (u *UxROM) ReadCharacterROM(address uint16) uint8 {
	return u.CharacterROM[address]
}

// MARK: ミラーリングの取得
func (u *UxROM) GetMirroring() Mirroring {
	return u.Mirroring
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (u *UxROM) GetIsCharacterRAM() bool {
	return u.IsCharacterRAM
}

// MARK: プログラムROMの取得
func (u *UxROM) GetProgramROM() []uint8 {
	return u.ProgramROM
}

// MARK: キャラクタROMの取得
func (u *UxROM) GetCharacterROM() []uint8 {
	return u.CharacterROM
}

// MARK: マッパー名の取得
func (u *UxROM) GetMapperInfo() string {
	return "UxROM (Mapper 2)"
}