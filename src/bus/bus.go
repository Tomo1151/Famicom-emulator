package bus

import (
	"Famicom-emulator/apu"
	"Famicom-emulator/cartridge"
	"Famicom-emulator/config"
	"Famicom-emulator/joypad"
	"Famicom-emulator/ppu"
)

const (
	CPU_WRAM_SIZE  = 2 * 1024 // 2kB
	CPU_WRAM_START = 0x0000
	CPU_WRAM_END   = 0x1FFF

	PPU_REG_START = 0x2000
	PPU_REG_END   = 0x3FFF

	PRG_ROM_START = 0x8000
	PRG_ROM_END   = 0xFFFF
)

// MARK: Busの定義
type Bus struct {
	wram      [CPU_WRAM_SIZE + 1]uint8 // CPUのWRAM (2kB)
	cartridge *cartridge.Cartridge     // カートリッジ
	ppu       *ppu.PPU                 // PPU
	apu       *apu.APU                 // APU
	joypad1   *joypad.JoyPad           // ポインタに変更
	joypad2   *joypad.JoyPad           // コントローラ (2P)
	cycles    uint                     // CPUサイクル
	callback  func(*ppu.PPU, *ppu.Canvas, *joypad.JoyPad, *joypad.JoyPad)
	canvas    *ppu.Canvas
	config    *config.Config
}

// MARK: Busの初期化メソッド (カートリッジ無し，デバッグ・テスト用)
func (b *Bus) InitForTest() {
	for addr := range b.wram {
		b.wram[addr] = 0x00
	}
}

// MARK: Busの初期化メソッド (ConnectComponents後に呼ばれる)
func (b *Bus) Init(callback func(*ppu.PPU, *ppu.Canvas, *joypad.JoyPad, *joypad.JoyPad)) {
	for addr := range b.wram {
		b.wram[addr] = 0x00
	}
	b.callback = callback
	b.canvas = &ppu.Canvas{}
	b.canvas.Init(*b.config)
}

// MARK: Canvasを取得
func (b *Bus) Canvas() *ppu.Canvas {
	return b.canvas
}

// MARK: Busに各コンポーネントを接続
func (b *Bus) ConnectComponents(
	ppu *ppu.PPU,
	apu *apu.APU,
	cartridge *cartridge.Cartridge,
	joypad1 *joypad.JoyPad,
	joypad2 *joypad.JoyPad,
	config *config.Config,
) {
	// 設定を反映
	b.config = config

	// コンポーネントをBusと接続
	b.ppu = ppu
	b.apu = apu
	b.cartridge = cartridge
	b.joypad1 = joypad1
	b.joypad2 = joypad2

	// 各コンポーネントを初期化
	b.ppu.Init(b.cartridge.Mapper(), *b.config)
	b.apu.Init(b.ReadByteFrom, *b.config)
	b.joypad1.Init()
	b.joypad2.Init()
}

// MARK: NMIを取得
func (b *Bus) NMI() bool {
	return b.ppu.PollNmiStatus()
}

// MARK: APUのIRQを取得
func (b *Bus) APUIRQ() bool {
	return b.apu.FrameIRQ()
}

// MARK: マッパーのIRQを取得
func (b *Bus) MapperIRQ() bool {
	return b.cartridge.Mapper().IRQ()
}

// MARK: リセット
func (b *Bus) Reset() {
	b.Shutdown()
	b.apu.Reset()
}

// MARK: 終了処理
func (b *Bus) Shutdown() {
	b.cartridge.Mapper().Save()
}

// MARK: サイクルを進める
func (b *Bus) Tick(cycles uint) {
	b.cycles += cycles

	nmiBefore := b.ppu.Nmi()

	frameEnd := false

	// PPUはCPUの3倍のクロック周波数
	for range cycles * 3 {
		if b.ppu.Tick(b.canvas, 1) {
			frameEnd = true

			// Canvasをバッファを交換し，すぐにPPUが次のレンダリングを行っても混ざらないように
			b.canvas.Swap()
		}
	}

	// APUと同期
	b.apu.Tick(cycles)

	nmiAfter := b.ppu.Nmi()
	if frameEnd || (!nmiBefore && nmiAfter) {
		b.apu.EndFrame()
		b.callback(b.ppu, b.canvas, b.joypad1, b.joypad2)
	}
}

// MARK: メモリの読み取り (1byte)
func (b *Bus) ReadByteFrom(address uint16) uint8 {
	/*
		CPUメモリマップ

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
	case address == 0x2000: // PPU_CTRL
		return b.ppu.ReadOpenBus()
	case address == 0x2001: // PPU_MASK
		return b.ppu.ReadOpenBus()
	case address == 0x2002: // PPU_STATUS
		return b.ppu.ReadPPUStatus()
	case address == 0x2003: // OAM_ADDR
		return b.ppu.ReadOpenBus()
	case address == 0x2004: // OAM_DATA
		return b.ppu.ReadOAMData() // OAMはDMA転送を使用するため，ほとんど使わないはず?
	case address == 0x2005: // PPU_SCROLL
		return b.ppu.ReadOpenBus()
	case address == 0x2006: // PPU_ADDR
		return b.ppu.ReadOpenBus()
	case address == 0x2007: // PPU_DATA
		return b.ppu.ReadVRAM()
	case 0x2008 <= address && address <= PPU_REG_END: // PPUレジスタのミラーリング
		// $2000 ~ $2007 (8bytesを繰り返すようにマスク)
		ptr := 0x2000 | (address & 0x07)
		return b.ReadByteFrom(ptr)
	case address == 0x4014: // OAM_DATA (DMA)
		// @NOTE 本来はCPU側のOpenBusを返すべき
		return 0x00
	case address == 0x4015: // APU
		return b.apu.ReadStatus()
	case address == 0x4016: // JOYPAD (1P)
		return b.joypad1.Read()
	case address == 0x4017: // JOYPAD (2P)
		return b.joypad2.Read()
	case 0x6000 <= address && address <= 0x7FFF: // プログラムRAM
		return b.cartridge.Mapper().ReadProgramRam(address)
	case PRG_ROM_START <= address && address <= PRG_ROM_END: // プログラムROM
		return b.cartridge.Mapper().ReadProgramRom(address)
	default:
		return 0x00
	}
}

// MARK: メモリの読み取り (2byte)
func (b *Bus) ReadWordFrom(address uint16) uint16 {
	lower := b.ReadByteFrom(address)
	upper := b.ReadByteFrom(address + 1)

	return uint16(upper)<<8 | uint16(lower)
}

// MARK: メモリの書き込み (1byte)
func (b *Bus) WriteByteAt(address uint16, data uint8) {
	/*
		CPU メモリマップ

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
	case address == 0x2000: // PPU_CTRL
		b.ppu.WriteToPPUControlRegister(data)
	case address == 0x2001: // PPU_MASK
		b.ppu.WriteToPPUMaskRegister(data)
	case address == 0x2002: // PPU_STATUS
		b.ppu.WriteToPPUStatusRegister(data)
	case address == 0x2003: // OAM_ADDR
		b.ppu.WriteToOAMAddressRegister(data)
	case address == 0x2004: // OAM_DATA
		b.ppu.WriteToOAMDataRegister(data)
	case address == 0x2005: // PPU_SCROLL
		b.ppu.WriteToPPUInternalRegister(address, data)
	case address == 0x2006: // PPU_ADDR
		b.ppu.WriteToPPUInternalRegister(address, data)
	case address == 0x2007: // PPU_DATA
		b.ppu.WriteVRAM(data)
	case 0x2008 <= address && address <= PPU_REG_END: // PPUレジスタのミラーリング
		// $2008 ~ $3FFF は $2000 ~ $2007 (8bytesを繰り返すようにマスク) へミラーリング
		ptr := 0x2000 | (address & 0x07)
		b.WriteByteAt(ptr, data)
	case 0x4000 <= address && address <= 0x4003: // APU 1ch
		b.apu.Write1ch(address, data)
	case 0x4004 <= address && address <= 0x4007: // APU 2ch
		b.apu.Write2ch(address, data)
	case address == 0x4008: // APU 3ch
		b.apu.Write3ch(address, data)
	case address == 0x400A: // APU 3ch
		b.apu.Write3ch(address, data)
	case address == 0x400B: // APU 3ch
		b.apu.Write3ch(address, data)
	case address == 0x400C: // APU 4ch
		b.apu.Write4ch(address, data)
	case address == 0x400E: // APU 4ch
		b.apu.Write4ch(address, data)
	case address == 0x400F: // APU 4ch
		b.apu.Write4ch(address, data)
	case 0x4010 <= address && address <= 0x4013: // APU 5ch
		b.apu.Write5ch(address, data)
		// fmt.Printf("W 5ch: %04X -> %02X\n", address, data)
	case address == 0x4014: // DMA転送
		var buffer [256]uint8
		upper := uint16(data) << 8
		for i := range 256 {
			buffer[i] = b.ReadByteFrom(upper + uint16(i))
		}

		// OAM DMA は 513 / 514 CPU サイクル を消費する
		var dmaCpuCycles uint
		if b.cycles%2 == 0 {
			dmaCpuCycles = 513
		} else {
			dmaCpuCycles = 514
		}
		b.cycles += dmaCpuCycles

		// PPU のサイクルを進める
		for range dmaCpuCycles * 3 {
			b.ppu.Tick(b.canvas, 1)
		}

		// APU のサイクルを進める
		for range dmaCpuCycles {
			b.apu.Tick(1)
		}

		// 用意されたデータを転送
		b.ppu.DMATransfer(&buffer)
	case address == 0x4015: // APU
		b.apu.WriteStatus(data)
	case address == 0x4016: // コントローラ (1P/2P)
		// @FIXME 2PのB・A同時押しの読み取りミスが多い
		b.joypad1.Write(data)
		b.joypad2.Write(data)
	case address == 0x4017: // APU フレームカウンタ
		b.apu.WriteFrameSequencer(data)
	case 0x6000 <= address && address <= 0x7FFF: // プログラムRAM
		b.cartridge.Mapper().WriteToProgramRam(address, data)
	case PRG_ROM_START <= address && address <= PRG_ROM_END: // プログラムROM
		b.cartridge.Mapper().Write(address, data)
	default:
	}
}

// MARK: メモリの書き込み (2byte)
func (b *Bus) WriteWordAt(address uint16, data uint16) {
	upper := uint8(data >> 8)
	lower := uint8(data & 0xFF)
	b.WriteByteAt(address, lower)
	b.WriteByteAt(address+1, upper)
}

// MARK: 現在のサイクル数の取得
func (b *Bus) Cycles() uint {
	return b.cycles
}
