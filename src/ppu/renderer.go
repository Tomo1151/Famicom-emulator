package ppu

import "github.com/veandco/go-sdl2/sdl"

// MARK: 定数定義
const (
	SCREEN_WIDTH  uint = 256
	SCREEN_HEIGHT uint = 240

	SCALE_FACTOR uint   = 3
	FPS          uint64 = 60
)

// MARK: Canvasの定義
type Canvas struct {
	width  uint
	height uint
	buffer [SCREEN_WIDTH * SCREEN_HEIGHT * 3]uint8
}

// MARK: Canvasのコンストラクタ
func NewCanvas() Canvas {
	return Canvas{
		width:  SCREEN_WIDTH,
		height: SCREEN_HEIGHT,
	}
}

// MARK: 指定した座標へ色を書き込み
func (c *Canvas) SetPixel(x, y uint, color sdl.Color) {
	basePos := (y*SCREEN_WIDTH + x) * 3 // 基準インデックスの計算
	c.buffer[basePos+0] = color.R
	c.buffer[basePos+1] = color.G
	c.buffer[basePos+2] = color.B
}

// MARK: バッファの取得メソッド
func (c *Canvas) Buffer() *[SCREEN_WIDTH * SCREEN_HEIGHT * 3]uint8 {
	return &c.buffer
}
