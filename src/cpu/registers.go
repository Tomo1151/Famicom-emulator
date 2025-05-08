package cpu

type registers struct {
	A  uint8          // アキュムレータ: 8bit
	X  uint8          // インデックスレジスタ: 8bit
	Y  uint8          // インデックスレジスタ: 8bit
	SP uint8          // スタックポインタ: 8bit
	PC uint16         // プログラムカウンタ: 16bit
	P  statusRegister // ステータスレジスタ: 8bit
}

type statusRegister struct {
	Negative  bool // N: 演算結果の最上位ビットが1の時にセット
	Overflow  bool // V: P演算結果がオーバーフローした時にセット
	Reserved  bool // R: 予約済み，常にセット
	Break     bool // B: BRK発生時にセットされIRQ発生時にクリア
	Decimal   bool // D: 0 -> デフォルト， 1 -> BCDモード (NESではBCDモードは未実装)
	Interrupt bool // I: 0 -> IRQ許可， 1 -> IRQ禁止
	Zero      bool // Z: 演算結果が0の時にセット
	Carry     bool // C: キャリー発生時にセット
}