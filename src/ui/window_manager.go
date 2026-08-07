package ui

import (
	"Famicom-emulator/apu"
	"Famicom-emulator/ppu"

	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: Window インターフェースの定義
type Window interface {
	ID() uint32
	HandleEvent(event sdl.Event)
	Update()
	Render()
	Close()
	SetIcon(*sdl.Surface)
}

// MARK: WindowManager の定義
type WindowManager struct {
	icon    *sdl.Surface
	windows map[uint32]Window
}

// MARK: WindowManager の作成メソッド
func NewWindowManager() *WindowManager {
	// アイコン画像のロード
	file, err := os.Open("../icon.png")
	if err != nil {
		fmt.Printf("Error: icon file not found: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error: icon image decode failed: %v", err)
	}
	icon, err := createSurfaceFromImage(img)

	return &WindowManager{icon: icon, windows: make(map[uint32]Window)}
}

// MARK: ウィンドウの登録メソッド
func (wm *WindowManager) Add(w Window) {
	// アイコンとウィンドウのセット
	w.SetIcon(wm.icon)
	wm.windows[w.ID()] = w
}

// MARK: ID指定でウィンドウを削除するメソッド
func (wm *WindowManager) Remove(id uint32) {
	if w, ok := wm.windows[id]; ok {
		w.Close()
		delete(wm.windows, id)
	}
}

// MARK: ID指定でウィンドウを取得するメソッド
func (wm *WindowManager) Get(id uint32) Window {
	return wm.windows[id]
}

// MARK: イベント処理メソッド
func (wm *WindowManager) HandleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		wm.dispatch(e.WindowID, event)
		if e.Event == sdl.WINDOWEVENT_CLOSE {
			wm.Remove(e.WindowID)
		}
	case *sdl.KeyboardEvent:
		wm.dispatch(e.WindowID, event)
	case *sdl.TextInputEvent:
		wm.dispatch(e.WindowID, event)
	case *sdl.MouseButtonEvent:
		wm.dispatch(e.WindowID, event)
	case *sdl.MouseMotionEvent:
		wm.dispatch(e.WindowID, event)
	case *sdl.MouseWheelEvent:
		wm.dispatch(e.WindowID, event)
	default:
		// イベントのウィンドウ指定がなければすべてにイベントを投げる
		for _, w := range wm.windows {
			w.HandleEvent(event)
		}
	}
}

// MARK: すべてのウィンドウを描画するメソッド
func (wm *WindowManager) RenderAll() {
	for _, w := range wm.windows {
		w.Update()
		w.Render()
	}
}

// MARK: すべてのウィンドウを閉じるメソッド
func (wm *WindowManager) CloseAll() {
	for id := range wm.windows {
		wm.Remove(id)
	}

	// アイコン画像の解放
	wm.icon.Free()
}

// MARK: ID指定でイベントをウィンドウに投げるメソッド
func (wm *WindowManager) dispatch(windowID uint32, event sdl.Event) {
	if windowID == 0 {
		return
	}
	if w, ok := wm.windows[windowID]; ok {
		w.HandleEvent(event)
	}
}

// MARK: NameTableWindow の表示/非表示切り替えメソッド
func (wm *WindowManager) ToggleNameTableWindow(p *ppu.PPU, scale int) (uint32, error) {
	for _, w := range wm.windows {
		if nw, ok := w.(*NameTableWindow); ok {
			id := nw.ID()
			wm.Remove(id)
			return 0, nil
		}
	}
	// NameTable viewer は4画面あり大きいため SCALE_FACTOR - 1 を使用する
	desiredScale := max(scale-1, 1)
	nw, err := NewNameTableWindow(p, desiredScale, func(id uint32) { wm.Remove(id) })
	if err != nil {
		return 0, err
	}
	wm.Add(nw)
	return nw.ID(), nil
}

// MARK: OptionWindow の表示/非表示切り替えメソッド
func (wm *WindowManager) ToggleCharacterWindow(p *ppu.PPU, scale int) (uint32, error) {
	for _, w := range wm.windows {
		if cw, ok := w.(*CharacterWindow); ok {
			id := cw.ID()
			wm.Remove(id)
			return 0, nil
		}
	}
	cw, err := NewCharacterWindow(p, scale, func(id uint32) { wm.Remove(id) })
	if err != nil {
		return 0, err
	}
	wm.Add(cw)
	return cw.ID(), nil
}

// MARK: AudioWindow の表示/非表示切り替えメソッド
func (wm *WindowManager) ToggleAudioWindow(a *apu.APU, scale int) (uint32, error) {
	for _, w := range wm.windows {
		if aw, ok := w.(*AudioWindow); ok {
			id := aw.ID()
			wm.Remove(id)
			return 0, nil
		}
	}
	aw, err := NewAudioWindow(a, scale, func(id uint32) { wm.Remove(id) })
	if err != nil {
		return 0, err
	}
	wm.Add(aw)
	return aw.ID(), nil
}

// MARK: OAMWindow の表示/非表示切り替えメソッド
func (wm *WindowManager) ToggleOAMWindow(p *ppu.PPU, scale int) (uint32, error) {
	for _, w := range wm.windows {
		if ow, ok := w.(*OAMWindow); ok {
			id := ow.ID()
			wm.Remove(id)
			return 0, nil
		}
	}
	ow, err := NewOAMWindow(p, scale, func(id uint32) { wm.Remove(id) })
	if err != nil {
		return 0, err
	}
	wm.Add(ow)
	return ow.ID(), nil
}

// MARK: 画像からSDL_Surfaceへ変換するメソッド
func createSurfaceFromImage(img image.Image) (*sdl.Surface, error) {
	bounds := img.Bounds()
	width := int32(bounds.Dx())
	height := int32(bounds.Dy())

	surface, err := sdl.CreateRGBSurface(
		0,
		width,
		height,
		32,
		0x000000FF,
		0x0000FF00,
		0x00FF0000,
		0xFF000000,
	)

	if err != nil {
		return nil, err
	}

	pixels := surface.Pixels()
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[idx+0] = byte(r >> 8)
			pixels[idx+1] = byte(g >> 8)
			pixels[idx+2] = byte(b >> 8)
			pixels[idx+3] = byte(a >> 8)
			idx += 4
		}
	}

	return surface, err
}
