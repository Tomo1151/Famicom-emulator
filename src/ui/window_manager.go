package ui

import "github.com/veandco/go-sdl2/sdl"

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
