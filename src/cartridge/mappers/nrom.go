package mappers

type NROM struct {
	ProgramROM []uint8
}

func (n *NROM) Init(ProgramROM []uint8) {
	n.ProgramROM = ProgramROM
}

func (n *NROM) Write(address uint16, data uint8) {}

func (n *NROM) ReadProgramROM(address uint16) uint8 {
	// カートリッジは$8000-$FFFFにマッピングされるためオフセット分引く
	romAddress := address - 0x8000

	// 16kBのROM(小さいROM)でアドレスが16kB以上の場合はミラーリング
	if len(n.ProgramROM) == 0x4000 && romAddress >= 0x4000 {
		romAddress %= 0x4000
	}
	return n.ProgramROM[romAddress]
}

func (n *NROM) GetProgramROM() []uint8 {
	return n.ProgramROM
}

func (n *NROM) GetMapperInfo() string {
	return "NROM (Mapper 0)"
}