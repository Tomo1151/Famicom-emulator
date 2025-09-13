package mappers

import "fmt"

type UxROM struct {
	bank       uint8
	ProgramROM []uint8
}

func (u *UxROM) Init(ProgramROM []uint8) {
	u.bank = 0x00
	u.ProgramROM = ProgramROM
}

func (u *UxROM) Write(address uint16, data uint8) {
	u.bank = data
}

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

func (u *UxROM) GetProgramROM() []uint8 {
	return u.ProgramROM
}

func (u *UxROM) GetMapperInfo() string {
	return "UxROM (Mapper 2)"
}