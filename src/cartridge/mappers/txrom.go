package mappers

import "fmt"

// MARK: MMC3 TxROM (マッパー4) の定義
type TxROM struct {
	bank uint8
	bankData [8]uint8
	ramProtect uint8
	irqLatch uint8
	irqReload bool
	irqEnable bool

	IsCharacterRAM bool
	Mirroring Mirroring
	ProgramROM   []uint8
	CharacterROM []uint8
	ProgramRAM	 [PRG_RAM_SIZE]uint8
}

// MARK: マッパーの初期化
func (t *TxROM) Init(rom []uint8) {
	programRom, characterROM := GetROMs(rom)
	t.bank = 0x00
	t.ramProtect = 0x00
	t.irqLatch = 0x00
	t.irqReload = false
	t.irqEnable = false

	for i := range t.bankData { t.bankData[i] = 0x00 }
	
	t.IsCharacterRAM = GetCharacterROMSize(rom) == 0
	t.Mirroring = GetSimpleMirroring(rom)
	t.ProgramROM = programRom
	t.CharacterROM = characterROM
	// プログラムRAMの初期化

	for i := range t.ProgramRAM {
		t.ProgramRAM[i] = 0xFF
	}
}

// MARK: ROMスペースへの書き込み
func (t *TxROM) Write(address uint16, data uint8) {
	switch {
	case PRG_ROM_START <= address && address <= 0x9FFF:
		if address & 0x01 == 0 {
			// バンクセレクト ($8000~$9FFE, 偶数)
			t.bank = data
		} else {
			// バンクデータ ($8001~$9FFF, 奇数)
			t.bankData[uint(t.bank & 0x07)] = data
		}
	case 0xA000 <= address && address <= 0xBFFF:
		if address & 0x01 == 0 {
			// ミラーリング ($A000~$BFFE, 偶数)
			if data & 0x01 == 0 {
				t.Mirroring = MIRRORING_VERTICAL
			} else {
				t.Mirroring = MIRRORING_HORIZONTAL
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
			if address & 0x01 == 0 {
			// IRQ ラッチ ($C000~$DFFE, 偶数)
			t.irqLatch = data
		} else {
			// IRQ リロード ($C001~$DFFF, 奇数)
			t.irqReload = true
		}
	case 0xE000 <= address && address <= PRG_ROM_END:
		if address & 0x01 == 0 {
			// IRQ 無効化 ($E000~$FFFE, 偶数)
			t.irqEnable = false
		} else {
			// IRQ 有効化 ($E001~$FFFD, 奇数)
			t.irqEnable = true
		}
	}
}

// MARK: プログラムROMの読み取り
func (t *TxROM) ReadProgramROM(address uint16) uint8 {
	/*
		mode         0      1
		$8000~$9FFF: R6    (-2)
		$A000~$BFFF: R7     R7
		$C000~$DFFF: (-2)   R6
		$E000~$FFFF: (-1)  (-1)
	*/
	bankMax := uint(len(t.ProgramROM)) / BANK_SIZE

	mode := t.bank & 0x40

	lastBank1 := uint(bankMax - 1)
	lastBank2 := uint(bankMax - 2)
	r6Bank := uint(t.bankData[6])
	r7Bank := uint(t.bankData[7])

	switch mode {
	case 0:
		switch {
		case PRG_ROM_START <= address && address <= 0x9FFF:
			return t.ProgramROM[uint(address - PRG_ROM_START) + r6Bank * BANK_SIZE]
		case 0xA000 <= address && address <= 0xBFFF:
			return t.ProgramROM[uint(address - 0xA000) + r7Bank * BANK_SIZE]
		case 0xC000 <= address && address <= 0xDFFF:
			return t.ProgramROM[uint(address - 0xC000) + lastBank2 * BANK_SIZE]
		case 0xE000 <= address && address <= PRG_ROM_END:
			return t.ProgramROM[uint(address - 0xE000) + lastBank1 * BANK_SIZE]
		default:
			panic(fmt.Sprintf("Error: unexpected program rom read: $%04X", address))
		}
	case 1:
		switch {
		case PRG_ROM_START <= address && address <= 0x9FFF:
			return t.ProgramROM[uint(address - PRG_ROM_START) + lastBank2 * BANK_SIZE]
		case 0xA000 <= address && address <= 0xBFFF:
			return t.ProgramROM[uint(address - 0xA000) + r7Bank * BANK_SIZE]
		case 0xC000 <= address && address <= 0xDFFF:
			return t.ProgramROM[uint(address - 0xC000) + r6Bank * BANK_SIZE]
		case 0xE000 <= address && address <= PRG_ROM_END:
			return t.ProgramROM[uint(address - 0xE000) + lastBank1 * BANK_SIZE]
		default:
			panic(fmt.Sprintf("Error: unexpected program rom read: $%04X", address))
		}
	default:
		panic(fmt.Sprintf("Error: unexpected read mode %02X", mode))
	}
}

// MARK: キャラクタROMのアドレス計算
func (t *TxROM) getCharacterROMAddress(address uint16) uint {
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

	r0Bank := uint(t.bankData[0])
	r1Bank := uint(t.bankData[1])
	r2Bank := uint(t.bankData[2])
	r3Bank := uint(t.bankData[3])
	r4Bank := uint(t.bankData[4])
	r5Bank := uint(t.bankData[5])


	switch mode {
	case 0:
		switch {
		case CHR_BANK_START <= address && address <= 0x07FF:
			return uint(address - CHR_BANK_START) + r0Bank * CHR_BANK_SIZE
		case 0x0800 <= address && address <= 0x0FFF:
			return uint(address - 0x0800) + r1Bank * CHR_BANK_SIZE
		case 0x1000 <= address && address <= 0x13FF:
			return uint(address - 0x1000) + r2Bank * CHR_BANK_SIZE
		case 0x1400 <= address && address <= 0x17FF:
			return uint(address - 0x1400) + r3Bank * CHR_BANK_SIZE
		case 0x1800 <= address && address <= 0x1BFF:
			return uint(address - 0x1800) + r4Bank * CHR_BANK_SIZE
		case 0x1C00 <= address && address <= CHR_BANK_END:
			return uint(address - 0x1C00) + r5Bank * CHR_BANK_SIZE
		default:
			panic(fmt.Sprintf("Error: unexpected character rom address: $%04X", address))
		}
	case 1:
		switch {
		case CHR_BANK_START <= address && address <= 0x03FF:
			return uint(address - CHR_BANK_START) + r2Bank * CHR_BANK_SIZE
		case 0x0400 <= address && address <= 0x07FF:
			return uint(address - 0x0400) + r3Bank * CHR_BANK_SIZE
		case 0x0800 <= address && address <= 0x0BFF:
			return uint(address - 0x0800) + r4Bank * CHR_BANK_SIZE
		case 0x0C00 <= address && address <= 0x0FFF:
			return uint(address - 0x0C00) + r5Bank * CHR_BANK_SIZE
		case 0x1000 <= address && address <= 0x17FF:
			return uint(address - 0x1000) + r0Bank * CHR_BANK_SIZE
		case 0x1800 <= address && address <= CHR_BANK_END:
			return uint(address - 0x1800) + r1Bank * CHR_BANK_SIZE
		default:
			panic(fmt.Sprintf("Error: unexpected character rom address: $%04X", address))
		}
	default:
		panic(fmt.Sprintf("Error: unexpected read mode %02X", mode))
	}
}

// MARK: キャラクタROMの読み取り
func (t *TxROM) ReadCharacterROM(address uint16) uint8 {
	return t.CharacterROM[t.getCharacterROMAddress(address)]
}

// MARK: キャラクタROMへの書き込み
func (t *TxROM) WriteToCharacterROM(address uint16, data uint8) {
	t.CharacterROM[t.getCharacterROMAddress(address)] = data
}

// MARK: プログラムRAMの読み取り
func (t *TxROM) ReadProgramRAM(address uint16) uint8 {
	return t.ProgramRAM[address-PRG_RAM_START]
}

// MARK: プログラムRAMへの書き込み
func (t *TxROM) WriteToProgramRAM(address uint16, data uint8) {
	t.ProgramRAM[address-PRG_RAM_START] = data
}

// MARK: ミラーリングの取得
func (t *TxROM) GetMirroring() Mirroring {
	return t.Mirroring
}

// MARK: キャラクタRAMを使用するかどうかを取得
func (t *TxROM) GetIsCharacterRAM() bool {
	return t.IsCharacterRAM
}

// MARK: プログラムROMの取得
func (t *TxROM) GetProgramROM() []uint8 {
	return t.ProgramROM
}

// MARK: キャラクタROMの取得
func (t *TxROM) GetCharacterROM() []uint8 {
	return t.CharacterROM
}

// MARK: マッパー名の取得
func (t *TxROM) GetMapperInfo() string {
	return "TxROM (Mapper 4)"
}