package mappers

// MARK: CNROM (マッパー3) の定義
type CNROM struct {
	bank       uint8

	IsCharacterRAM bool
	Mirroring Mirroring
	ProgramROM []uint8
	CharacterROM []uint8
}

// MARK: マッパーの初期化
func (c *CNROM) Init(rom []uint8) {
	c.bank = 0x00

	programROM, characterROM := GetROMs(rom)
	c.IsCharacterRAM = GetCharacterROMSize(rom) == 0
	c.Mirroring = GetSimpleMirroring(rom)
	c.ProgramROM = programROM
	c.CharacterROM = characterROM
}

// MARK: ROMスペースへの書き込み
func (c *CNROM) Write(address uint16, data uint8) {
	c.bank = data
}

// MARK: プログラムROMの読み取り
func (c *CNROM) ReadProgramROM(address uint16) uint8 {
	return c.ProgramROM[uint(address)-uint(PRG_ROM_START)]
}

// MARK: キャラクタROMの読み取り
func (c *CNROM) ReadCharacterROM(address uint16) uint8 {
	// CNROMはバンクセレクトの下位2ビットのみを使う
	return c.CharacterROM[uint(address)+BANK_SIZE*uint(c.bank & 0x03)]
}

// MARK: キャラクタROMへの書き込み
func (c *CNROM) WriteToCharacterROM(address uint16, data uint8) {
	c.CharacterROM[address] = data
}

// MARK: プログラムRAMの読み取り
func (c *CNROM) ReadProgramRAM(address uint16) uint8 {
	panic("Error: unsupported read program RAM on CNROM")
}

// MARK: プログラムRAMへの書き込み
func (c *CNROM) WriteToProgramRAM(address uint16, data uint8) {}

// MARK: ミラーリングの取得
func (c *CNROM) GetMirroring() Mirroring {
	return c.Mirroring
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (c *CNROM) GetIsCharacterRAM() bool {
	return c.IsCharacterRAM
}

// MARK: プログラムROMの取得
func (c *CNROM) GetProgramROM() []uint8 {
	return c.ProgramROM
}

// MARK: キャラクタROMの取得
func (c *CNROM) GetCharacterROM() []uint8 {
	return c.CharacterROM
}

// MARK: マッパー名の取得
func (c *CNROM) GetMapperInfo() string {
	return "CNROM (Mapper 3)"
}