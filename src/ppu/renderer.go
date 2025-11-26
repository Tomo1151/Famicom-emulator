package ppu

// MARK: 定数定義
const (
	SCREEN_WIDTH  uint = 256
	SCREEN_HEIGHT uint = 240
	CANVAS_WIDTH  uint = SCREEN_WIDTH
	CANVAS_HEIGHT uint = SCREEN_HEIGHT
)

const (
	PIXEL_TYPE_BACKGROUND PixelType = 0
	PIXEL_TYPE_SPRITE     PixelType = 1
)

// MARK: PixelTypeの定義
type PixelType byte

// MARK: Pixelの定義
type Pixel struct {
	priority            uint8
	backgroundValue     [3]uint8
	spriteValue         [3]uint8
	isBgTransparent     bool
	isSpriteTransparent bool
}

func (p *Pixel) Value(ppu *PPU) [3]uint8 {
	var color [3]uint8
	if p.isBgTransparent && p.isSpriteTransparent {
		color = PALETTE[ppu.paletteTable[0]]
	} else if p.isBgTransparent && !p.isSpriteTransparent {
		color = p.spriteValue
	} else if !p.isBgTransparent && p.isSpriteTransparent {
		color = p.backgroundValue
	} else {
		if p.priority == 0 {
			color = p.spriteValue
		} else {
			color = p.backgroundValue
		}
	}
	return color
}

// MARK: Canvasの定義
type Canvas struct {
	Width   uint
	Height  uint
	Buffers [2][uint(SCREEN_WIDTH) * uint(SCREEN_HEIGHT) * 3]byte // ダブルバッファリング
	index   int                                                   // 現在表示されていバッファのインデックス
}

// MARK: キャンバスの初期化メソッド
func (c *Canvas) Init() {
	c.Width = SCREEN_WIDTH
	c.Height = SCREEN_HEIGHT
	c.index = 0
}

// MARK: キャンバスの指定した座標に色をセット
func (c *Canvas) setPixelAt(x uint, y uint, palette [3]uint8) {
	if x >= c.Width || y >= c.Height {
		return
	}

	// 裏側のバッファに描画を行う
	back := 1 - c.index
	basePtr := int((y*c.Width + x) * 3)
	c.Buffers[back][basePtr+0] = palette[0] // R
	c.Buffers[back][basePtr+1] = palette[1] // G
	c.Buffers[back][basePtr+2] = palette[2] // B
}

// 現在描画しているバッファと入れ替える
func (c *Canvas) Swap() {
	c.index = 1 - c.index
}

// 現在描画しているバッファの先頭のポインタを返す
func (c *Canvas) FrontBuffer() *[uint(SCREEN_WIDTH) * uint(SCREEN_HEIGHT) * 3]byte {
	return &c.Buffers[c.index]
}

// MARK: 指定したスキャンラインをキャンバスに描画
func RenderScanlineToCanvas(ppu *PPU, canvas *Canvas, scanline uint16) {
	ppu.ClearLineBuffer()
	ppu.CalculateScanlineBackground(canvas, scanline)
	ppu.CalculateScanlineSprite(canvas, scanline)

	for x := range SCREEN_WIDTH {
		canvas.setPixelAt(x, uint(scanline), ppu.lineBuffer[x].Value(ppu))
	}
}
