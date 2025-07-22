package ppu

import (
	"Famicom-emulator/cartridge"
	"fmt"
)

const (
	VRAM_SIZE          uint16 = 2 * 1024 // 2kB
	PALETTE_TABLE_SIZE  uint8 = 32
	OAM_DATA_SIZE      uint16 = 64 * 4
)

const (
	SCANLINE_START      = 0
	SCANLINE_POSTRENDER = 240
	SCANLINE_VBLANK     = 241
	SCANLINE_PRERENDER  = 261
	SCANLINE_END        = 341
)

// MARK: PPUの定義
type PPU struct {
	CHR_ROM []uint8
	PaletteTable [PALETTE_TABLE_SIZE+1]uint8
	vram [VRAM_SIZE]uint8
	oam [OAM_DATA_SIZE]uint8
	Mirroring cartridge.Mirroring

	control   ControlRegister // $2000
	mask   MaskRegister    // $2001
	status StatusRegister  // $2002
	scroll ScrollRegister  // $2005
	address   AddrRegister    // $2006

	scanline uint16 // 現在描画中のスキャンライン
	cycles uint // PPUサイクル
	internalDataBuffer uint8
	oamAddress uint8 // OAM書き込みのポインタ

	NMI *uint8
}

// MARK: PPUの初期化メソッド
func (p *PPU) Init(chr_rom []uint8, mirroring cartridge.Mirroring){
	p.CHR_ROM = chr_rom
	p.Mirroring = mirroring
	for addr := range p.vram { p.vram[addr] = 0x00 }
	for addr := range p.oam { p.oam[addr] = 0x00 }
	for addr := range p.PaletteTable { p.PaletteTable[addr] = 0x00 }
	
	p.control.Init()
	p.mask.Init()
	p.status.Init()
	p.scroll.Init()
	p.address.Init()

	p.oamAddress = 0
	p.scanline = 0
	p.cycles = 0
	p.internalDataBuffer = 0x00

	p.NMI = nil
}

// MARK: PPUアドレスレジスタへの書き込み
func (p *PPU) WriteToPPUAddrRegister(value uint8) {
	p.address.update(value)
}

// MARK: PPUコントロールレジスタ($2000)への書き込み
func (p *PPU) WriteToPPUControlRegister(value uint8) {
	prev := p.control.GenerateVBlankNMI()
	p.control.update(value)

	// VBlank中にGenerateNMIが立つタイミングでNMIを発生させる
	if !prev && p.control.GenerateVBlankNMI() && p.status.IsInVBlank() {
		*p.NMI = 0x01
	}
}

// MARK: PPUマスクレジスタ($2001)への書き込み
func (p *PPU) WriteToPPUMaskRegister(value uint8) {
	p.mask.update(value)
}

// MARK: OAM ADDR($2003) への書き込み
func (p *PPU) WriteToOAMAddressRegister(addr uint8) {
	p.oamAddress = addr
}

// MARK: PPUスクロールレジスタ($2005)への書き込み
func (p *PPU) WriteToPPUScrollRegister(data uint8) {
	p.scroll.Write(data)
}

// MARK: OAM DATA($4014) への書き込み
func (p *PPU) WriteToOAMDataRegister(data uint8) {
	p.oam[p.oamAddress] = data
	p.oamAddress++
}

// MARK: DMA転送を行う ([256]u8 の配列のアドレスを受け取る)
func (p *PPU) DMATransfer(bytes *[256]uint8) {
	for _, byte := range *bytes {
		p.oam[p.oamAddress] = byte
		p.oamAddress++
	}
}


// MARK: VRAMアドレスをインクリメント
func (p *PPU) incrementVRAMAddress() {
	p.address.increment(p.control.GetVRAMAddrIncrement())
}

// MARK: VRAMへの書き込み
func (p *PPU) WriteVRAM(value uint8) {
	/*
		PPUメモリマップ

		$0000-$1FFF $2000 パレットテーブル (CHR ROM)
		$2000-$3EFF $1F00 ネームテーブル (VRAM)
		$3F00-$3FFF $0100 パレット
		$4000-$FFFF $4000 $0000-$3FFF のミラーリング
	*/

	addr := p.address.get()
	p.incrementVRAMAddress()

	switch {
	case addr <= 0x1FFF:
		// panic(fmt.Sprintf("addr space 0x0000..0x1FFF is not expected to write, requested: %04X", addr))
		return
	case 0x2000 <= addr && addr <= 0x2FFF:
		p.vram[p.mirrorVRAMAddress(addr)] = value
	case 0x3000 <= addr && addr <= 0x3EFF:
		// fmt.Printf("Error: unexpected vram write to $%04X\n", addr)
		return
	case 0x3F00 <= addr && addr <= 0x3F1F:
		// アドレスのミラーリング
		if addr == 0x3F10 ||
			 addr == 0x3F14 ||
			 addr == 0x3F18 ||
			 addr == 0x3FC {
			addr -= 0x10
		}
		p.PaletteTable[addr - 0x3F00] = value
	case 0x3F20 <= addr && addr <= 0x3FFF:
		p.PaletteTable[(addr - 0x3F00)%32] = value
	default:
		panic(fmt.Sprintf("Unexpected write to mirrored space: %04X", addr))
	}
}

// MARK: PPUステータスレジスタの読み取り
func (p *PPU) ReadPPUStatus() uint8 {
	status := p.status.ToByte()
	p.status.ClearVBlankStatus()
	p.scroll.ResetLatch()
	p.address.ResetLatch()
	return status
}

// MARK: OAM DATAの読み取り
func (p *PPU) ReadOAMData() uint8 {
	return p.oam[p.oamAddress]
}

// MARK: VRAMの読み取り
func (p *PPU) ReadVRAM() uint8 {
	/*
		PPUメモリマップ

		$0000-$1FFF $2000 パレットテーブル (CHR ROM)
		$2000-$3EFF $1F00 ネームテーブル (VRAM)
		$3F00-$3FFF $0100 パレット
		$4000-$FFFF $4000 $0000-$3FFF のミラーリング
	*/

	addr := p.address.get()
	p.incrementVRAMAddress()

	switch {
	case addr <= 0x1FFF:
		value := p.internalDataBuffer
		p.internalDataBuffer = p.CHR_ROM[addr]
		return value
	case 0x2000 <= addr && addr <= 0x2FFF:
		// 一回遅れで値は反映されるため，内部バッファを更新し，元のバッファ値を返す
		value := p.internalDataBuffer
		p.internalDataBuffer = p.vram[p.mirrorVRAMAddress(addr)]
		return value
		// fmt.Println("@TODO read from VRAM")
	case 0x3000 <= addr && addr <= 0x3EFF:
		panic(fmt.Sprintf("addr space 0x3000..0x3eff is not expected to read, requested: %04X", addr))
	case 0x3F00 <= addr && addr <= 0x3F1F:
		// アドレスのミラーリング
		if addr == 0x3F10 ||
			 addr == 0x3F14 ||
			 addr == 0x3F18 ||
			 addr == 0x3FC {
			addr -= 0x10
		}
		return p.PaletteTable[addr - 0x3F00]
	case 0x3F20 <= addr && addr <= 0x3FFF:
		return p.PaletteTable[(addr - 0x3F00)%32]
	default:
		panic(fmt.Sprintf("Unexpected read to mirrored space: %04X", addr))
	}
}

// MARK: VRAMアドレスのミラーリング
func (p *PPU) mirrorVRAMAddress(addr uint16) uint16 {
	// 0x3000-0x3eff から 0x2000 - 0x2eff へミラーリング
	mirroredVRAMAddr := addr & PPU_VRAM_MIRROR_MASK

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
		switch nameTable {
		case 2, 1:
			return vramIndex - 0x400
		case 3:
			return vramIndex - 0x800
		}
	}

	return vramIndex
}

// MARK: 待機しているNMIを取得
func (p *PPU) GetNMI() *uint8 {
	if p.NMI != nil {
		value := *p.NMI
		p.NMI = nil
		return &value
	} else {
		return nil
	}
}

func (p *PPU) isSprite0Hit(cycles uint) bool {
	x := uint(p.oam[0])
	y := uint(p.oam[3])
	return y == uint(p.scanline) && x <= cycles && p.mask.SpriteEnable
}

// MARK: サイクルを進める
func (p *PPU) Tick(cycles uint) bool {
	p.cycles += cycles

	if p.cycles >= SCANLINE_END {
		if p.isSprite0Hit(cycles) {
			p.status.SetSpriteZeroHit(true)
		}

		// サイクル数を0に戻す
		p.cycles -= SCANLINE_END

		// スキャンラインを進める
		p.scanline++

		// VBlankに突入
		if p.scanline == SCANLINE_VBLANK {
			p.status.SetVBlankStatus(true)
			p.status.SetSpriteZeroHit(false)
			if p.control.GenerateVBlankNMI() {
				// p.status.SetVBlankStatus(true)
				// NMIを設定
				nmiValue := uint8(1)
				p.NMI = &nmiValue
			}
		}

		// プリレンダーラインに到達した時
		if p.scanline > SCANLINE_PRERENDER {
			p.scanline = 0
			p.NMI = nil
			p.status.ClearVBlankStatus()
			p.status.SetSpriteZeroHit(false)
			return true
		}
	}
	return false
}