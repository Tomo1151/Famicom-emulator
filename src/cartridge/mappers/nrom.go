package mappers

// MARK: NROM (マッパー0) の定義
type NROM struct {
	name           string
	isCharacterRam bool
	mirroring      Mirroring
	programRom     []uint8
	characterRom   []uint8
}

// MARK: マッパーの初期化
func (n *NROM) Init(name string, rom []uint8, save []uint8) {
	programRom, characterRom := roms(rom)

	n.name = name
	n.isCharacterRam = characterRomSize(rom) == 0
	n.mirroring = simpleMirroring(rom)
	n.programRom = programRom
	n.characterRom = characterRom
}

// MARK: ROMスペースへの書き込み
func (n *NROM) Write(address uint16, data uint8) {}

// MARK: プログラムROMの読み取り
func (n *NROM) ReadProgramRom(address uint16) uint8 {
	// カートリッジは$8000-$FFFFにマッピングされるためオフセット分引く
	romAddress := address - 0x8000

	// 16kBのROM(小さいROM)でアドレスが16kB以上の場合はミラーリング
	if len(n.programRom) == 0x4000 && romAddress >= 0x4000 {
		romAddress %= 0x4000
	}
	return n.programRom[romAddress]
}

// MARK: キャラクタROMの読み取り
func (n *NROM) ReadCharacterRom(address uint16) uint8 {
	return n.characterRom[address]
}

// MARK: キャラクタROMへの書き込み
func (n *NROM) WriteToCharacterRom(address uint16, data uint8) {}

// MARK: プログラムRAMの読み取り
func (n *NROM) ReadProgramRam(address uint16) uint8 {
	panic("Error: unsupported read program ROM on NROM")
}

// MARK: プログラムRAMへの書き込み
func (n *NROM) WriteToProgramRam(address uint16, data uint8) {}

// MARK: セーブデータの書き出し
func (n *NROM) Save() {}

// MARK: スキャンラインによってIRQを発生させる
func (n *NROM) GenerateScanlineIRQ(scanline uint16, backgroundEnable bool) {}

// MARK: IRQ状態の取得
func (n *NROM) IRQ() bool { return false }

// MARK: ミラーリングの取得
func (n *NROM) Mirroring() Mirroring {
	return n.mirroring
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (n *NROM) IsCharacterRam() bool {
	return n.isCharacterRam
}

// MARK: プログラムROMの取得
func (n *NROM) ProgramRom() []uint8 {
	return n.programRom
}

// MARK: キャラクタROMの取得
func (n *NROM) CharacterRom() []uint8 {
	return n.characterRom
}

// MARK: マッパー名の取得
func (n *NROM) MapperInfo() string {
	return "NROM (Mapper 0)"
}

// MARK: マッパーのシャローコピーの取得
func (n *NROM) Clone() Mapper {
	copy := *n
	return &copy
}
