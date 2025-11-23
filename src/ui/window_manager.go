package ui

import (
	"Famicom-emulator/apu"
	"Famicom-emulator/config"
	"Famicom-emulator/ppu"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: Window インターフェースの定義
type Window interface {
	ID() uint32
	HandleEvent(event sdl.Event)
	Update()
	Render()
	Close()
}

// MARK: WindowManager の定義
type WindowManager struct {
	windows map[uint32]Window
}

// MARK: WindowManager の作成メソッド
func NewWindowManager() *WindowManager {
	return &WindowManager{windows: make(map[uint32]Window)}
}

// MARK: ウィンドウの登録メソッド
func (wm *WindowManager) Add(w Window) {
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

// MARK: OptionWindow の表示/非表示切り替えメソッド
func (wm *WindowManager) ToggleOptionWindow(cfg *config.Config) (uint32, error) {
	// 既に開かれていれば閉じる
	for _, w := range wm.windows {
		if ow, ok := w.(*OptionWindow); ok {
			id := ow.ID()
			wm.Remove(id)
			return 0, nil
		}
	}

	ow, err := NewOptionWindow(cfg, func(id uint32) { wm.Remove(id) })
	if err != nil {
		return 0, err
	}
	wm.Add(ow)
	return ow.ID(), nil
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
