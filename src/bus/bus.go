package bus

import (
	"Famicom-emulator/cartridge"
	"Famicom-emulator/ppu"
	"fmt"
)

const (
	CPU_WRAM_SIZE  = 2 * 1024 // 2kB
	CPU_WRAM_START = 0x0000
	CPU_WRAM_END   = 0x1FFF

	PPU_REG_START  = 0x2000
	PPU_REG_END    = 0x3FFF
)

// MARK: Busの定義
type Bus struct {
	wram [CPU_WRAM_SIZE+1]uint8 // CPUのWRAM (2kB)
	cartridge cartridge.Cartridge // カートリッジ
	ppu ppu.PPU // PPU
}


// MARK: Busの初期化メソッド (カートリッジ無し，デバッグ・テスト用)
func (b *Bus) Init() {
	for addr := range b.wram {
		b.wram[addr] = 0x00
	}
}

// MARK: Busの初期化メソッド (カートリッジ有り)
func (b *Bus) InitWithCartridge(cartridge *cartridge.Cartridge) {
	for addr := range b.wram {
		b.wram[addr] = 0x00
	}
	b.cartridge = *cartridge
	b.ppu = ppu.PPU{}
	b.ppu.Init(b.cartridge.CharacterROM, b.cartridge.ScreenMirroring)
}


// MARK: メモリの読み取り (1byte)
func (b *Bus) ReadByteFrom(address uint16) uint8 {
	/*
		$0000–$07FF	$0800	2kBのWRAM

		$0800–$0FFF	$0800	$0000–$07FF (WRAM) のミラーリング×3
		$1000–$17FF	$0800
		$1800–$1FFF	$0800

		$2000–$2007	$0008	PPUレジスタ
		$2008–$3FFF	$1FF8	$2000–$2007 (PPUレジスタ) のミラーリング

		$4000–$4017	$0018	APU, I/O レジスタ
		$4018–$401F	$0008	APU, I/O レジスタのテスト用 (通常は無効)

		$4020–$FFFF  	$BFE0	未割り当て，カートリッジで使用可能
		• $6000–$7FFF $2000 カートリッジRAM
		• $8000–$FFFF $8000 カートリッジROMまたはマッパーレジスタ
	*/

	switch {
	case CPU_WRAM_START <= address && address <= CPU_WRAM_END: // WRAM
		ptr := address & 0b00000111_11111111 // 11bitにマスク
		return b.wram[ptr]
	case // PPUレジスタ
			address == 0x2000 ||
			address == 0x2001 ||
			address == 0x2003 ||
			address == 0x2005 ||
			address == 0x2006 ||
			address == 0x4014:
		panic(fmt.Sprintf("Error: attempt to read from write only ppu address $%04X", address))
	case address == 0x2007:
		return b.ppu.ReadVRAM()
	case 0x2008 <= address && address <= PPU_REG_END: // PPUレジスタのミラーリング
		// $2000 ~ $2007 (8bytesを繰り返すようにマスク)
		ptr := address & 0b00100000_00000111
		return b.ReadByteFrom(ptr)
	case 0x8000 <= address: // プログラムROM
		return b.ReadProgramROM(address)
	default:
		fmt.Printf("Ignoring memory access at $%04X\n", address)
		return 0x00
	}
}

// MARK: メモリの読み取り (2byte)
func (b *Bus) ReadWordFrom(address uint16) uint16 {
	lower := b.ReadByteFrom(address)
	upper := b.ReadByteFrom(address + 1)

	return uint16(upper) << 8 | uint16(lower)
}

// MARK: メモリの書き込み (1byte)
func (b *Bus) WriteByteAt(address uint16, data uint8) {
	/*
		$0000–$07FF	$0800	2kBのWRAM

		$0800–$0FFF	$0800	$0000–$07FF (WRAM) のミラーリング×3
		$1000–$17FF	$0800
		$1800–$1FFF	$0800

		$2000       $0001 PPUコントロールレジスタ
		$2001       $0001 PPUマスクレジスタ
		$2002       $0001 PPUステータスレジスタ
		$2003       $0001 OAMアドレスレジスタ (スプライトRAMアドレス)
		$2004       $0001 OAMデータレジスタ (スプライトRAMデータ)
		$2005       $0001 PPUスクロールレジスタ
		$2006       $0001 PPUアドレスレジスタ (VRAMアドレス)
		$2007       $0001 PPUデータレジスタ (VRAMデータ)
		$2008–$3FFF	$1FF8	$2000–$2007 (PPUレジスタ) のミラーリング

		$4014       $0001 OAMDMA (スプライトDMA)
		$4000–$4017	$0018	APU, I/O レジスタ
		$4018–$401F	$0008	APU, I/O レジスタのテスト用 (通常は無効)

		$4020–$FFFF  	$BFE0	未割り当て，カートリッジで使用可能
		• $6000–$7FFF $2000 カートリッジRAM
		• $8000–$FFFF $8000 カートリッジROMまたはマッパーレジスタ
	*/

	switch {
	case CPU_WRAM_START <= address && address <= CPU_WRAM_END: // WRAM
		ptr := address & 0b00000111_11111111 // 11bitにマスク
		b.wram[ptr] = data
	case address == 0x2000: // PPUCTRL
		b.ppu.WriteToPPUControlRegister(data)
	case address == 0x2006: // PPUADDR
		b.ppu.WriteToPPUAddrRegister(data)
	case address == 0x2007: // PPUDATA
		// b.ppu.WriteToData(data)
	case 0x2008 <= address && address <= PPU_REG_END: // PPUレジスタのミラーリング
		// $2000 ~ $2007 (8bytesを繰り返すようにマスク)
		ptr := address & 0b00100000_00000111
		b.WriteByteAt(ptr, data)
	case 0x8000 <= address: // プログラムROM
		panic(fmt.Sprintf("Error: attempt to write to cartridge ROM space $%04X, 0x%02X\n", address, data))
	default:
		fmt.Printf("Ignoring memory write to $%04X\n", address)
	}
}

// MARK: メモリの書き込み (2byte)
func (b *Bus) WriteWordAt(address uint16, data uint16) {
	upper := uint8(data >> 8)
	lower := uint8(data & 0xFF)
	b.WriteByteAt(address, lower)
	b.WriteByteAt(address + 1, upper)
}

// MARK: プログラムROMの読み取り
func (b *Bus) ReadProgramROM(address uint16) uint8 {
	// カートリッジは$8000-$FFFFにマッピングされるためオフセット分引く
	addr := address - 0x8000

	// 16kBのROMでアドレスが16kB以上の場合はミラーリング
	if len(b.cartridge.ProgramROM) == 0x4000 && addr >= 0x4000 {
		addr %= 0x4000
	}
	return b.cartridge.ProgramROM[addr]
}