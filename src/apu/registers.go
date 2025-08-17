package apu

import "fmt"

const (
	NOISE_MODE_SHORT NoiseRegisterMode = iota
	NOISE_MODE_LONG
)

type NoiseRegisterMode uint8

// MARK: 矩形波レジスタ
type SquareWaveRegister struct {
	// 0x4000 | 0x4004
	volume uint8
	envelope bool
	keyOffCounter bool
	duty uint8

	// 0x4001 | 0x4005
	sweepShift uint8
	sweepDirection uint8
	sweepPeriod uint8
	sweepEnabled uint8

	// 0x4002 | 0x4006
	frequency uint16

	// 0x4003 | 0x4007
	keyOffCount uint8
}

// MARK: 矩形波レジスタの初期化メソッド
func (swr *SquareWaveRegister) Init() {
	swr.volume = 0x00
	swr.envelope = false
	swr.keyOffCounter = false
	swr.duty = 0x00
	swr.sweepShift = 0x00
	swr.sweepDirection = 0x00
	swr.sweepPeriod = 0x00
	swr.sweepEnabled = 0x00
	swr.frequency = 0x00
	swr.keyOffCount = 0x00
}

// MARK: 矩形波レジスタの書き込みメソッド（1ch/2ch）
func (swr *SquareWaveRegister) write(address uint16, data uint8) {
	switch address {
	case 0x4000, 0x4004:
		swr.volume = data & 0x0F
		swr.envelope = (data & 0x10) == 0
		swr.keyOffCounter = (data & 0x20) == 0
		swr.duty = (data & 0xC0) >> 6
	case 0x4001, 0x4005:
		swr.sweepShift = data & 0x07
		swr.sweepDirection = (data & 0x08) >> 3
		swr.sweepPeriod = (data & 0x70) >> 4
		swr.sweepEnabled = (data & 0x80) >> 7
	case 0x4002, 0x4006:
		swr.frequency = (swr.frequency & 0x0700) | uint16(data)
	case 0x4003, 0x4007:
		swr.frequency = (swr.frequency & 0x00FF) | uint16(data & 0x07) << 8
		swr.keyOffCount = (data & 0xF8) >> 3
	default:
		panic(fmt.Sprintf("APU Error: Invalid write at: %04X", address))
	}
}

// MARK: レジスタからデューティ比を取得するメソッド
func (swr *SquareWaveRegister) getDuty() float32 {
	// 00: 12.5%, 01: 25.0%, 10: 50.0%, 11: 75.0%
	switch swr.duty {
	case 0b00:
		return 0.125
	case 0b01:
		return 0.25
	case 0b10:
		return 0.50
	case 0b11:
		return 0.75
	default:
		return 0.0
	}
}

// MARK: そのチャンネルの矩形波を鳴らすかどうかを取得するメソッド
func (swr *SquareWaveRegister) isEnabled() bool {
	return (swr.volume & 0x0F) > 0 // ボリュームが0より大きければ有効（テスト実装）
}

// MARK: レジスタからボリュームを取得するメソッド（1ch/2ch）
func (swr *SquareWaveRegister) getVolume() float32 {
	// 0が消音，15が最大 ※ スウィープ無効時のみ
	return float32(swr.volume) / 15.0
}

// MARK: レジスタから矩形波のピッチを取得するメソッド
func (swr *SquareWaveRegister) getFrequency() float32 {
	return CPU_CLOCK / (16.0*float32(swr.frequency) + 1.0)
}


// MARK: 三角波レジスタ
type TriangleWaveRegister struct {
	// 0x4008
	length uint8
	keyOffCounter bool

	// 0x400A, 0x400B
	frequency uint16
	keyOffCount uint8
}

// 三角波レジスタの書き込みメソッド
func (twr *TriangleWaveRegister) Init() {
	twr.length = 0x00
	twr.keyOffCounter = false
	twr.frequency = 0x0000
	twr.keyOffCount = 0x00
}

// 三角波レジスタの書き込みメソッド（3ch）
func (twr *TriangleWaveRegister) write(address uint16, data uint8) {
	switch address {
	case 0x4008:
		twr.length = data & 0x7F
		twr.keyOffCounter = (data & 0x80) == 0
	case 0x400A:
		twr.frequency = (twr.frequency & 0x0700) | uint16(data)
	case 0x400B:
		twr.frequency = (twr.frequency & 0x00FF) | uint16(data & 0x07) << 8
		twr.keyOffCount = (data & 0xF8) >> 3
	default:
		panic(fmt.Sprintf("APU Error: Invalid write at: %04X", address))
	}
}

// MARK: レジスタから三角波のピッチを取得するメソッド
func (twr *TriangleWaveRegister) getFrequency() float32 {
	return CPU_CLOCK / (32.0*float32(twr.frequency) + 1.0)
}


// MARK: ノイズレジスタ
type NoiseWaveRegister struct {
	// 0x400C
	volume uint8
	envelope bool
	keyOffCounter bool

	// 0x400E
	frequency uint8
	mode NoiseRegisterMode

	// 0x400F
	keyOffCount uint8
}

// MARK: ノイズレジスタの初期化メソッド
func (nwr *NoiseWaveRegister) Init() {
	nwr.volume = 0x00
	nwr.envelope = false
	nwr.keyOffCounter = false
	nwr.frequency = 0x00
	nwr.mode = NOISE_MODE_LONG
	nwr.keyOffCount = 0x00
}

// MARK: ノイズレジスタの書き込みメソッド（4ch）
func (nwr *NoiseWaveRegister) write(adress uint16, data uint8) {
	switch adress {
	case 0x400C:
		nwr.volume = data & 0x0F
		nwr.envelope = (data & 0x10) == 0
		nwr.keyOffCounter = (data & 0x20) == 0
	case 0x400E:
		nwr.frequency = data & 0x0F
		mode := data & 0x80

		if (mode != 0){
			nwr.mode = NOISE_MODE_SHORT
		} else {
			nwr.mode = NOISE_MODE_LONG
		}
	case 0x400F:
		nwr.keyOffCount = (data & 0xF8) >> 3
	default:
		panic(fmt.Sprintf("APU Error: unexpected write at %04X", adress))
	}
}

// MARK: ノイズレジスタから4chの音量を取得するメソッド
func (nwr *NoiseWaveRegister) getVolume() float32 {
	return float32(nwr.volume) / 15.0
}

// MARK: ノイズレジスタから4chのモードを取得するメソッド
func (nwr *NoiseWaveRegister) getMode() NoiseRegisterMode {
	return nwr.mode
}

// MARK: ノイズレジスタからノイズのピッチを取得するメソッド
func (nwr *NoiseWaveRegister) getFrequency() float32 {
	noiseFrequencyTable := [16]uint16{
		0x0002,
		0x0004,
		0x0008,
		0x0010,
		0x0020,
		0x0030,
		0x0040,
		0x0050,
		0x0065,
		0x007F,
		0x00BE,
		0x00FE,
		0x017D,
		0x01FC,
		0x03F9,
		0x07F2,
	}

	return CPU_CLOCK / float32(noiseFrequencyTable[nwr.frequency])
}


// MARK: ノイズシフトレジスタ
type NoiseShiftRegister struct {
	mode   NoiseRegisterMode
	value  uint16
}

func (nsr *NoiseShiftRegister) InitWithLongMode() {
	nsr.mode = NOISE_MODE_LONG
	nsr.value = 1
}

func (nsr *NoiseShiftRegister) InitWithShortMode() {
	nsr.mode = NOISE_MODE_SHORT
	nsr.value = 1
}

func (nsr *NoiseShiftRegister) next() bool {
	/*
		タイマーによってシフトレジスタが励起されるたびに1ビット右シフト
		もしショートモードなら
			ビット0とビット6のXOR
		ロングモードなら
			ビット0とビット1のXOR
		が入る
	*/
	var shiftBit uint16

	switch nsr.mode {
	case NOISE_MODE_LONG:
		shiftBit = 1
	case NOISE_MODE_SHORT:
		shiftBit = 6
	default:
		panic(fmt.Sprintf("APU Error: unexpected noise shift register mode: %04X", shiftBit))
	}

	value := (nsr.value & 0x01) ^ ((nsr.value >> shiftBit) & 0x01)
	nsr.value >>= 1
	nsr.value = (nsr.value & 0b011_1111_1111_1111) | value<<14

	// シフトレジスタのビット0が1であればチャンネルの出力が0になる
	result := nsr.value&0x01 != 0

	// fmt.Printf("NoiseShift: mode=%d, value=0x%04X, result=%t\n", nsr.mode, nsr.value, result)
	return result
}