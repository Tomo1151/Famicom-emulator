package mappers

import "fmt"

// MARK: UxROM (マッパー2) の定義
type UxROM struct {
	name string
	bank uint8

	isCharacterRam bool
	mirroring      Mirroring
	programRom     []uint8
	characterRom   []uint8
}

// MARK: マッパーの初期化
func (u *UxROM) Init(name string, rom []uint8, save []uint8) {
	u.name = name
	u.bank = 0x00

	programRom, characterRom := roms(rom)
	u.isCharacterRam = characterRomSize(rom) == 0
	u.mirroring = simpleMirroring(rom)
	u.programRom = programRom
	u.characterRom = characterRom
}

// MARK: ROMスペースへの書き込み
func (u *UxROM) Write(address uint16, data uint8) {
	u.bank = data
}

// MARK: プログラムROMの読み取り
func (u *UxROM) ReadProgramRom(address uint16) uint8 {
	// 最後のバンク番号
	bankMax := uint(len(u.programRom)) / BANK_SIZE

	switch {
	case PRG_ROM_START <= address && address <= 0xBFFF:
		// 前半部分はバンク選択
		bank := uint(u.bank & 0x0F)
		return u.programRom[uint(address)-0x8000+BANK_SIZE*bank]
	case 0xC000 <= address && address <= PRG_ROM_END:
		// 後半部分は固定
		return u.programRom[uint(address)-0xC000+BANK_SIZE*(bankMax-1)]
	default:
		panic(fmt.Sprintf("Erorr: unexpected PRG ROM space: %04X", address))
	}
}

// MARK: キャラクタROMの読み取り
func (u *UxROM) ReadCharacterRom(address uint16) uint8 {
	return u.characterRom[address]
}

// MARK: キャラクタROMへの書き込み
func (u *UxROM) WriteToCharacterRom(address uint16, data uint8) {
	u.characterRom[address] = data
}

// MARK: プログラムRAMの読み取り
func (u *UxROM) ReadProgramRam(address uint16) uint8 {
	panic("Error: unsupported read program RAM on UxROM")
}

// MARK: プログラムRAMへの書き込み
func (u *UxROM) WriteToProgramRam(address uint16, data uint8) {}

// MARK: セーブデータの書き出し
func (u *UxROM) Save() {}

// MARK: スキャンラインによってIRQを発生させる
func (u *UxROM) GenerateScanlineIRQ(scanline uint16, backgroundEnable bool) {}

// MARK: IRQ状態の取得
func (u *UxROM) IRQ() bool { return false }

// MARK: ミラーリングの取得
func (u *UxROM) Mirroring() Mirroring {
	return u.mirroring
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (u *UxROM) IsCharacterRam() bool {
	return u.isCharacterRam
}

// MARK: プログラムROMの取得
func (u *UxROM) ProgramRom() []uint8 {
	return u.programRom
}

// MARK: キャラクタROMの取得
func (u *UxROM) CharacterRom() []uint8 {
	return u.characterRom
}

// MARK: マッパー名の取得
func (u *UxROM) MapperInfo() string {
	return "UxROM (Mapper 2)"
}

// MARK: マッパーのシャローコピーの取得
func (u *UxROM) Clone() Mapper {
	copy := *u
	return &copy
}
