package ppu

import (
	"Famicom-emulator/cartridge"
	"fmt"
)

const (
	VRAM_SIZE uint16 = 2 * 1024 // 2kB
	PALETTE_TABLE_SIZE uint8 = 32
	OAM_DATA_SIZE uint16 = 64 * 4
)

// MARK: PPUの定義
type PPU struct {
	CHR_ROM []uint8
	PaletteTable [PALETTE_TABLE_SIZE+1]uint8
	vram [VRAM_SIZE+1]uint8
	oam [OAM_DATA_SIZE+1]uint8
	Mirroring cartridge.Mirroring

	addrRegister AddrRegister
	ctrlRegister ControlRegister

	internalDataBuffer uint8
}

// MARK: PPUの初期化メソッド
func (p *PPU) Init(chr_rom []uint8, mirroring cartridge.Mirroring){
	p.CHR_ROM = chr_rom
	p.Mirroring = mirroring
	for addr := range p.vram { p.vram[addr] = 0x00 }
	for addr := range p.oam { p.oam[addr] = 0x00 }
	for addr := range p.PaletteTable { p.PaletteTable[addr] = 0x00 }
	p.addrRegister.Init()
	p.ctrlRegister.Init()

	p.internalDataBuffer = 0x00
}

// MARK: PPUアドレスレジスタへの書き込み
func (p *PPU) WriteToPPUAddrRegister(value uint8) {
	p.addrRegister.update(value)
}

// PPUコントロールレジスタへの書き込み
func (p *PPU) WriteToPPUControlRegister(value uint8) {
	p.ctrlRegister.update(value)
}

// VRAMアドレス
func (p *PPU) incrementVRAMAddr() {
	p.addrRegister.increment(p.ctrlRegister.GetVRAMAddrIncrement())
}

func (p *PPU) ReadVRAM() uint8 {
	addr := p.addrRegister.get()
	p.incrementVRAMAddr()

	switch {
	case 0x0000 <= addr && addr <= 0x1FFF:
		value := p.internalDataBuffer
		p.internalDataBuffer = p.CHR_ROM[addr]
		return value
	case 0x2000 <= addr && addr <= 0x2FFF:
		value := p.internalDataBuffer
		p.internalDataBuffer = p.vram[p.mirrorVRAMAddr(addr)]
		return value
		// fmt.Println("@TODO read from VRAM")
	case 0x3000 <= addr && addr <= 0x3EFF:
		panic(fmt.Sprintf("addr space 0x3000..0x3eff is not expected to be used, requested: %04X", addr))
	case 0x3F00 <= addr && addr <= 0x3FFF:
		return p.PaletteTable[addr - 0x3F00]
	default:
		panic(fmt.Sprintf("Unexpected access to mirrored space: %04X", addr))
	}
}

func (p *PPU) mirrorVRAMAddr(addr uint16) uint16 {
	// 0x3000-0x3eff から 0x2000 - 0x2eff へミラーリング
	mirroredVRAMAddr := addr & PPU_ADDR_MIRROR_MASK

	// メモリアドレスをVRAMの配列用に補正 (VRAMの先頭アドレスを引く)
	vramIndex := mirroredVRAMAddr - 0x2000

	// ネームテーブルのインデックスを求める
	nameTable := vramIndex / 0x400

	// ネームテーブルのミラーリングがVerticalの場合
	// [ A ] [ B ] (一つのテーブルが 0x400 × 0x400，そのテーブルが 2 × 2)
	// [ a ] [ b ]
	if p.Mirroring == cartridge.MIRRORING_VERTICAL {
		if nameTable == 2 || nameTable == 3 {
			return vramIndex - 0x800
		}
	}

	// ネームテーブルのミラーリングがHorizontalの場合
	// [ A ] [ a ]
	// [ B ] [ b ]
	if p.Mirroring == cartridge.MIRRORING_HORIZONTAL {
		if nameTable == 2 || nameTable == 1 {
			return vramIndex - 0x400
		} else if nameTable == 3 {
			return vramIndex - 0x800
		}
	}

	return vramIndex
}