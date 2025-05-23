package ppu

import "Famicom-emulator/bus"

const (
	PPU_CTRL uint16 = 0x2000
	PPU_MASK uint16 = 0x2001
	PPU_STATUS uint16 = 0x2002
	OAM_ADDR uint16 = 0x2003
	OAM_DATA uint16 = 0x2004
	PPU_SCROLL uint16 = 0x2005
	PPU_ADDR uint16 = 0x2006
	PPU_DATA uint16 = 0x2007
	OAM_DMA uint16 = 0x4014

	PPU_ADDR_MIRROR_MASK = 0b111111_11111111 // 14ビット
)

type AddrRegister struct {
	upper uint8
	lower uint8
	isUpper bool
}

func (ar *AddrRegister) Init() {
	ar.upper = 0x00
	ar.lower = 0x00
	ar.isUpper = true
}

func (ar *AddrRegister) set(data uint16) {
	ar.upper = uint8(data >> 8)
	ar.lower = uint8(data & 0xFF)
}

func (ar *AddrRegister) get() uint16 {
	return uint16(ar.upper) << 8 | uint16(ar.lower)
}

func (ar *AddrRegister) update(data uint8) {
	// 1回目の書き込みは上位ビット, 2回目は下位ビット
	if ar.isUpper {
		ar.upper = data
	} else {
		ar.lower = data
	}

	// アドレスのミラーリング
	if ar.get() > bus.PPU_REG_END {
		ar.set(ar.get() & PPU_ADDR_MIRROR_MASK)
	}

	ar.isUpper = !ar.isUpper
}

func (ar *AddrRegister) increment(step uint8) {
	// 方向によって 1 or 32 増やす
	current := ar.get()
	result := current + uint16(step)

	// アドレスのミラーリング
	if result > bus.PPU_REG_END {
		ar.set(result & PPU_ADDR_MIRROR_MASK)
	}
}

func (ar *AddrRegister) resetLatch() {
	ar.isUpper = true
}