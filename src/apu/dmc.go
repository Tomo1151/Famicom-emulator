package apu

var (
	dmcFrequencyTable = [16]uint16{
		0x1AC, 0x17C, 0x154, 0x140,
		0x11E, 0x0FE, 0x0E2, 0x0D6,
		0x0BE, 0x0A0, 0x08E, 0x080,
		0x06A, 0x054, 0x048, 0x036,
	}
)

// MARK: DMCの定義
type DMCWaveChannel struct {
	register DMCRegister
	cpuRead  CpuBusReader
	enabled  bool
	irq      bool

	// DAC
	deltaCounter uint8 // 7bit DAC (0-127)

	// タイマー
	timerPeriod uint16
	timerValue  uint16

	// サンプル処理
	byteCount   uint16
	baseAddress uint16
	sample      uint8
	bitsLeft    uint8
	bytesLeft   uint16

	buffer BlipBuffer
}

// MARK: DMCの初期化メソッド
func (dwc *DMCWaveChannel) Init(reader CpuBusReader) {
	dwc.register = DMCRegister{}
	dwc.register.Init()
	dwc.cpuRead = reader
	dwc.baseAddress = 0xC000
	dwc.byteCount = 1
	dwc.buffer.Init()
}

// MARK: DMCのタイマーを進めるメソッド
func (dwc *DMCWaveChannel) tick(cycles uint) {
	if dwc.timerValue == 0 || !dwc.enabled {
		return
	}

	// タイマーが0になるまで待機
	if dwc.timerValue <= uint16(cycles) {
		dwc.timerValue = 0
	} else {
		dwc.timerValue -= uint16(cycles)
		return
	}

	dwc.timerValue = dwc.timerPeriod
	if dwc.bitsLeft == 0 {
		// 次のサンプルをフェッチ
		if dwc.bytesLeft > 0 {
			dwc.sample = dwc.cpuRead(dwc.baseAddress)
			dwc.baseAddress++

			// オーバーフロー
			if dwc.baseAddress == 0 {
				dwc.baseAddress = 0x8000
			}
			dwc.bytesLeft--
			dwc.bitsLeft = 8

			if dwc.bytesLeft == 0 {
				// サンプル終了
				if dwc.register.loop {
					dwc.restart()
				} else if dwc.register.irqEnabled {
					dwc.irq = true
				}
			}
		} else {
			// サンプルがない場合
			return
		}
	}

	// 1ビット処理
	if (dwc.sample & 0x01) == 1 {
		if dwc.deltaCounter < 127 {
			dwc.deltaCounter += 2
		}
	} else {
		if dwc.deltaCounter > 1 {
			dwc.deltaCounter -= 2
		}
	}
	dwc.sample >>= 1
	dwc.bitsLeft--
}

// MARK: DMCの出力メソッド
func (dwc *DMCWaveChannel) output() float32 {
	return float32(dwc.deltaCounter)
}

// MARK: DMCの再生再開メソッド
func (dwc *DMCWaveChannel) restart() {
	dwc.baseAddress = (uint16(dwc.register.sampleStartAddress) << 6) + 0xC000
	dwc.bytesLeft = (uint16(dwc.register.byteCount) << 4) + 1
}

// チャンネルの有効/無効設定メソッド
func (dwc *DMCWaveChannel) setEnabled(enabled bool) {
	dwc.enabled = enabled
	if !enabled {
		// 無効化されたときに再生中のサンプルを止める
		dwc.bytesLeft = 0
	} else {
		// 有効化されたときに再生が終わっていれば再開
		if dwc.bytesLeft == 0 {
			dwc.restart()
			dwc.timerValue = dwc.timerPeriod
		}
	}
}
