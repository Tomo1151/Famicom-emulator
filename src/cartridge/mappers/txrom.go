package mappers

import (
	"fmt"
	"os"
)

const (
	TXROM_PRG_BANK_SIZE = 8 * 1024 // 8kB
	SCANLINE_POSTRENDER = 240
)

// MARK: MMC3 TxROM (マッパー4) の定義
type TxROM struct {
	name string

	bank       uint8
	bankData   [8]uint8
	ramProtect uint8
	irqLatch   uint8
	irqReload  bool
	irqEnable  bool
	irqCounter uint8
	irq        bool

	isCharacterRam bool
	mirroring      Mirroring
	programRom     []uint8
	characterRom   []uint8
	programRam     [PRG_RAM_SIZE]uint8
}

// MARK: マッパーの初期化
func (t *TxROM) Init(name string, rom []uint8, save []uint8) {
	programRom, characterROM := roms(rom)
	t.name = name
	t.bank = 0x00
	t.ramProtect = 0x00
	t.irqLatch = 0x00
	t.irqReload = false
	t.irqEnable = false

	for i := range t.bankData {
		t.bankData[i] = 0x00
	}

	t.isCharacterRam = characterRomSize(rom) == 0
	t.mirroring = simpleMirroring(rom)
	t.programRom = programRom
	t.characterRom = characterROM

	// プログラムRAMの初期化
	for i := range t.programRam {
		t.programRam[i] = 0xFF
	}

	// セーブデータの読み込み
	if len(save) != 0 {
		copy(t.programRam[:], save)
		t.ramProtect = 0x80
	}
}

// MARK: ROMスペースへの書き込み
func (t *TxROM) Write(address uint16, data uint8) {
	switch {
	case PRG_ROM_START <= address && address <= 0x9FFF:
		if address&0x01 == 0 {
			// バンクセレクト ($8000~$9FFE, 偶数)
			t.bank = data
		} else {
			// バンクデータ ($8001~$9FFF, 奇数)
			t.bankData[uint(t.bank&0x07)] = data
		}
	case 0xA000 <= address && address <= 0xBFFF:
		if address&0x01 == 0 {
			// ミラーリング ($A000~$BFFE, 偶数)
			if data&0x01 == 0 {
				t.mirroring = MIRRORING_VERTICAL
			} else {
				t.mirroring = MIRRORING_HORIZONTAL
			}
		} else {
			// プログラムRAM 保護 ($A001~$BFFF, 奇数)
			t.ramProtect = data

			/*
				@NOTE
				MMC3では一応機能するがあまり必要ない，MMC6との非互換性を避けるために実装しない
			*/
		}
	case 0xC000 <= address && address <= 0xDFFF:
		if address&0x01 == 0 {
			// IRQ ラッチ ($C000~$DFFE, 偶数)
			t.irqLatch = data
			t.irqCounter = data
		} else {
			// IRQ リロード ($C001~$DFFF, 奇数)
			t.irqReload = true
			t.irqCounter = 0
		}
	case 0xE000 <= address && address <= PRG_ROM_END:
		if address&0x01 == 0 {
			// IRQ 無効化 ($E000~$FFFE, 偶数)
			t.irqEnable = false
			t.irq = false
		} else {
			// IRQ 有効化 ($E001~$FFFD, 奇数)
			t.irqEnable = true
		}
	}
}

// MARK: プログラムROMの読み取り
func (t *TxROM) ReadProgramRom(address uint16) uint8 {
	/*
		mode         0      1
		$8000~$9FFF: R6    (-2)
		$A000~$BFFF: R7     R7
		$C000~$DFFF: (-2)   R6
		$E000~$FFFF: (-1)  (-1)
	*/
	bankMax := uint(len(t.programRom)) / TXROM_PRG_BANK_SIZE

	mode := t.bank & 0x40

	lastBank1 := uint(bankMax - 1)
	lastBank2 := uint(bankMax - 2)

	/*
		@NOTE
		> R6 and R7 will ignore the top two bits, as the MMC3 has only 6 PRG ROM address lines.
		https://www.nesdev.org/wiki/MMC3

		TxROMではR6/R7の上位2bitが無視されるためマスク
	*/
	r6Bank := uint(t.bankData[6]) & 0x3F
	r7Bank := uint(t.bankData[7]) & 0x3F

	switch mode {
	case 0:
		switch {
		case PRG_ROM_START <= address && address <= 0x9FFF:
			return t.programRom[uint(address-PRG_ROM_START)+r6Bank*TXROM_PRG_BANK_SIZE]
		case 0xA000 <= address && address <= 0xBFFF:
			return t.programRom[uint(address-0xA000)+r7Bank*TXROM_PRG_BANK_SIZE]
		case 0xC000 <= address && address <= 0xDFFF:
			return t.programRom[uint(address-0xC000)+lastBank2*TXROM_PRG_BANK_SIZE]
		case 0xE000 <= address && address <= PRG_ROM_END:
			return t.programRom[uint(address-0xE000)+lastBank1*TXROM_PRG_BANK_SIZE]
		default:
			panic(fmt.Sprintf("Error: unexpected program rom read: $%04X", address))
		}
	default:
		switch {
		case PRG_ROM_START <= address && address <= 0x9FFF:
			return t.programRom[uint(address-PRG_ROM_START)+lastBank2*TXROM_PRG_BANK_SIZE]
		case 0xA000 <= address && address <= 0xBFFF:
			return t.programRom[uint(address-0xA000)+r7Bank*TXROM_PRG_BANK_SIZE]
		case 0xC000 <= address && address <= 0xDFFF:
			return t.programRom[uint(address-0xC000)+r6Bank*TXROM_PRG_BANK_SIZE]
		case 0xE000 <= address && address <= PRG_ROM_END:
			return t.programRom[uint(address-0xE000)+lastBank1*TXROM_PRG_BANK_SIZE]
		default:
			panic(fmt.Sprintf("Error: unexpected program rom read: $%04X", address))
		}
	}
}

// MARK: キャラクタROMのアドレス計算
func (t *TxROM) calcCharacterRomAddress(address uint16) uint {
	/*
		mode          0    1
		$0000~$03FF: R0   R2
		$0400~$07FF:      R3
		$0800~$0BFF: R1   R4
		$0C00~$0FFF:      R5
		$1000~$13FF: R2   R0
		$1400~$17FF: R3
		$1800~$1BFF: R4   R1
		$1C00~$1FFF: R5
	*/
	var CHR_BANK_START uint16 = 0x0000
	var CHR_BANK_END uint16 = 0x1FFF
	var CHR_BANK_SIZE uint = 1 * 1024

	mode := t.bank & 0x80

	r0Bank := uint(t.bankData[0] & 0xFE)
	r1Bank := uint(t.bankData[1] & 0xFE)
	r2Bank := uint(t.bankData[2])
	r3Bank := uint(t.bankData[3])
	r4Bank := uint(t.bankData[4])
	r5Bank := uint(t.bankData[5])

	switch mode {
	case 0:
		switch {
		case CHR_BANK_START <= address && address <= 0x07FF:
			return uint(address-CHR_BANK_START) + r0Bank*CHR_BANK_SIZE
		case 0x0800 <= address && address <= 0x0FFF:
			return uint(address-0x0800) + r1Bank*CHR_BANK_SIZE
		case 0x1000 <= address && address <= 0x13FF:
			return uint(address-0x1000) + r2Bank*CHR_BANK_SIZE
		case 0x1400 <= address && address <= 0x17FF:
			return uint(address-0x1400) + r3Bank*CHR_BANK_SIZE
		case 0x1800 <= address && address <= 0x1BFF:
			return uint(address-0x1800) + r4Bank*CHR_BANK_SIZE
		case 0x1C00 <= address && address <= CHR_BANK_END:
			return uint(address-0x1C00) + r5Bank*CHR_BANK_SIZE
		default:
			panic(fmt.Sprintf("Error: unexpected character rom address: $%04X", address))
		}
	default:
		switch {
		case CHR_BANK_START <= address && address <= 0x03FF:
			return uint(address-CHR_BANK_START) + r2Bank*CHR_BANK_SIZE
		case 0x0400 <= address && address <= 0x07FF:
			return uint(address-0x0400) + r3Bank*CHR_BANK_SIZE
		case 0x0800 <= address && address <= 0x0BFF:
			return uint(address-0x0800) + r4Bank*CHR_BANK_SIZE
		case 0x0C00 <= address && address <= 0x0FFF:
			return uint(address-0x0C00) + r5Bank*CHR_BANK_SIZE
		case 0x1000 <= address && address <= 0x17FF:
			return uint(address-0x1000) + r0Bank*CHR_BANK_SIZE
		case 0x1800 <= address && address <= CHR_BANK_END:
			return uint(address-0x1800) + r1Bank*CHR_BANK_SIZE
		default:
			panic(fmt.Sprintf("Error: unexpected character rom address: $%04X", address))
		}
	}
}

// MARK: キャラクタROMの読み取り
func (t *TxROM) ReadCharacterRom(address uint16) uint8 {
	return t.characterRom[t.calcCharacterRomAddress(address)]
}

// MARK: キャラクタROMへの書き込み
func (t *TxROM) WriteToCharacterRom(address uint16, data uint8) {
	t.characterRom[t.calcCharacterRomAddress(address)] = data
}

// MARK: プログラムRAMの読み取り
func (t *TxROM) ReadProgramRam(address uint16) uint8 {
	// RAM有効ビットが立っている場合のみRAMから読み取り、それ以外は0xFF
	if t.ramProtect&0x80 != 0 {
		return t.programRam[address-PRG_RAM_START]
	}
	return 0xFF // 無効なRAMアクセスの場合は0xFFを返す
}

// MARK: プログラムRAMへの書き込み
func (t *TxROM) WriteToProgramRam(address uint16, data uint8) {
	// RAM保護が無効な場合のみ書き込む
	if t.ramProtect&0x80 != 0 && t.ramProtect&0x40 == 0 {
		t.programRam[address-PRG_RAM_START] = data
	}
}

// MARK: セーブデータの書き出し
func (t *TxROM) Save() {
	// RAM書き込みが有効な場合のみセーブ
	if t.ramProtect&0x80 != 0 {
		err := os.WriteFile(SAVE_DATA_DIR+t.name+".save", t.programRam[:], 0644)
		if err != nil {
			fmt.Printf("Error saving game data: %v\n", err)
		} else {
			fmt.Printf("Game saved to: %s\n", SAVE_DATA_DIR+t.name+".save")
		}
	}
}

// MARK: スキャンラインによってIRQを発生させる
func (t *TxROM) GenerateScanlineIRQ(scanline uint16, renderEnable bool) {
	if scanline <= SCANLINE_POSTRENDER && renderEnable {
		// リロードフラグが立っているか、カウンタが0なら、カウンタをラッチ値でリロード
		if t.irqReload || t.irqCounter == 0 {
			t.irqCounter = t.irqLatch
			t.irqReload = false
		} else {
			// そうでなければカウンタをデクリメント
			t.irqCounter--
		}

		// カウンタが0になり、かつIRQが有効ならIRQを発生
		if t.irqCounter == 0 && t.irqEnable {
			t.irq = true
		}
	}
}

// MARK: IRQ状態の取得
func (t *TxROM) IRQ() bool {
	value := t.irq
	t.irq = false
	return value
}

// MARK: ミラーリングの取得
func (t *TxROM) Mirroring() Mirroring {
	return t.mirroring
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (t *TxROM) IsCharacterRam() bool {
	return t.isCharacterRam
}

// MARK: プログラムROMの取得
func (t *TxROM) ProgramRom() []uint8 {
	return t.programRom
}

// MARK: キャラクタROMの取得
func (t *TxROM) CharacterRom() []uint8 {
	return t.characterRom
}

// MARK: マッパー名の取得
func (t *TxROM) MapperInfo() string {
	return "TxROM (Mapper 4)"
}

// MARK: マッパーのシャローコピーの取得
func (t *TxROM) Clone() Mapper {
	copy := *t
	return &copy
}
