package mappers

// MARK: NROM (マッパー0) の定義
type NROM struct {
	IsCharacterRAM bool
	Mirroring Mirroring
	ProgramROM   []uint8
	CharacterROM []uint8
}

// MARK: マッパーの初期化
func (n *NROM) Init(rom []uint8) {
	programRom, characterROM := GetROMs(rom)

	n.IsCharacterRAM = GetCharacterROMSize(rom) == 0
	n.Mirroring = GetSimpleMirroring(rom)
	n.ProgramROM = programRom
	n.CharacterROM = characterROM
}

// MARK: ROMスペースへの書き込み
func (n *NROM) Write(address uint16, data uint8) {}

// MARK: プログラムROMの読み取り
func (n *NROM) ReadProgramROM(address uint16) uint8 {
	// カートリッジは$8000-$FFFFにマッピングされるためオフセット分引く
	romAddress := address - 0x8000

	// 16kBのROM(小さいROM)でアドレスが16kB以上の場合はミラーリング
	if len(n.ProgramROM) == 0x4000 && romAddress >= 0x4000 {
		romAddress %= 0x4000
	}
	return n.ProgramROM[romAddress]
}

// MARK: キャラクタROMの読み取り
func (n *NROM) ReadCharacterROM(address uint16) uint8 {
	return n.CharacterROM[address]
}

// MARK: キャラクタROMへの書き込み
func (n *NROM) WriteToCharacterROM(address uint16, data uint8) {}

// MARK: プログラムRAMの読み取り
func (n *NROM) ReadProgramRAM(address uint16) uint8 {
	panic("Error: unsupported read program RAM on NROM")
}

// MARK: プログラムRAMへの書き込み
func (n *NROM) WriteToProgramRAM(address uint16, data uint8) {}

// MARK: ミラーリングの取得
func (n *NROM) GetMirroring() Mirroring {
	return n.Mirroring
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (n *NROM) GetIsCharacterRAM() bool {
	return n.IsCharacterRAM
}

// MARK: プログラムROMの取得
func (n *NROM) GetProgramROM() []uint8 {
	return n.ProgramROM
}

// MARK: キャラクタROMの取得
func (n *NROM) GetCharacterROM() []uint8 {
	return n.CharacterROM
}

// MARK: マッパー名の取得
func (n *NROM) GetMapperInfo() string {
	return "NROM (Mapper 0)"
}