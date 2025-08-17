package apu

import "fmt"

type SquareWaveRegister struct {
	toneVolume    uint8
	sweep         uint8
	freqLow       uint8
	freqHighKeyOn uint8
}

// MARK: 矩形波レジスタへ書き込むメソッド（1ch/2ch）
func (swr *SquareWaveRegister) write(address uint16, data uint8) {
	switch address {
	case 0x4000:
		swr.toneVolume = data
	case 0x4001:
		swr.sweep = data
	case 0x4002:
		swr.freqLow = data
	case 0x4003:
		swr.freqHighKeyOn = data
	default:
		panic(fmt.Sprintf("APU Error: Invalid write at: %04X", address))
	}
}

// MARK: レジスタからデューティ比を取得するメソッド
func (swr *SquareWaveRegister) duty() float32 {
	// 00: 12.5%, 01: 25.0%, 10: 50.0%, 11: 75.0%
	value := (swr.toneVolume & 0xC0) >> 6
	switch value {
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
	return (swr.toneVolume & 0x0F) > 0 // ボリュームが0より大きければ有効（テスト実装）
}

// MARK: レジスタからボリュームを取得するメソッド（1ch/2ch）
func (swr *SquareWaveRegister) volume() float32 {
	// 0が消音，15が最大 ※ スウィープ無効時のみ
	return float32(swr.toneVolume&0x0F) / 15.0
}

// MARK: レジスタから矩形波のピッチを取得するメソッド
func (swr *SquareWaveRegister) freq() float32 {
	value := ((uint16(swr.freqHighKeyOn) & 0x07) << 8) | uint16(swr.freqLow)
	return CPU_CLOCK / (16.0*float32(value) + 1.0)
}
