package mappers

// MARK: CNROM (マッパー3) の定義
type CNROM struct {
	name string
	bank uint8

	isCharacterRam bool
	mirroring      Mirroring
	programRom     []uint8
	characterRom   []uint8
}

// MARK: マッパーの初期化
func (c *CNROM) Init(name string, rom []uint8, save []uint8) {
	c.name = name
	c.bank = 0x00

	programRom, characterRom := roms(rom)
	c.isCharacterRam = characterRomSize(rom) == 0
	c.mirroring = simpleMirroring(rom)
	c.programRom = programRom
	c.characterRom = characterRom
}

// MARK: ROMスペースへの書き込み
func (c *CNROM) Write(address uint16, data uint8) {
	c.bank = data
}

// MARK: プログラムROMの読み取り
func (c *CNROM) ReadProgramRom(address uint16) uint8 {
	return c.programRom[uint(address)-uint(PRG_ROM_START)]
}

// MARK: キャラクタROMの読み取り
func (c *CNROM) ReadCharacterRom(address uint16) uint8 {
	// CNROMはバンクセレクトの下位2ビットのみを使う
	return c.characterRom[uint(address)+BANK_SIZE*uint(c.bank&0x03)]
}

// MARK: キャラクタROMへの書き込み
func (c *CNROM) WriteToCharacterRom(address uint16, data uint8) {
	c.characterRom[address] = data
}

// MARK: プログラムRAMの読み取り
func (c *CNROM) ReadProgramRam(address uint16) uint8 {
	panic("Error: unsupported read program RAM on CNROM")
}

// MARK: プログラムRAMへの書き込み
func (c *CNROM) WriteToProgramRam(address uint16, data uint8) {}

// MARK: セーブデータの書き出し
func (c *CNROM) Save() {}

// MARK: スキャンラインによってIRQを発生させる
func (c *CNROM) GenerateScanlineIRQ(scanline uint16, backgroundEnable bool) {}

// MARK: IRQ状態の取得
func (c *CNROM) IRQ() bool { return false }

// MARK: ミラーリングの取得
func (c *CNROM) Mirroring() Mirroring {
	return c.mirroring
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (c *CNROM) IsCharacterRam() bool {
	return c.isCharacterRam
}

// MARK: プログラムROMの取得
func (c *CNROM) ProgramRom() []uint8 {
	return c.programRom
}

// MARK: キャラクタROMの取得
func (c *CNROM) CharacterRom() []uint8 {
	return c.characterRom
}

// MARK: マッパー名の取得
func (c *CNROM) MapperInfo() string {
	return "CNROM (Mapper 3)"
}
