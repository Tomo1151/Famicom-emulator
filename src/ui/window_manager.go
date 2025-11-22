package ui

import "github.com/veandco/go-sdl2/sdl"

// Window defines the behaviour each SDL window wrapper must implement.
type Window interface {
	ID() uint32
	HandleEvent(event sdl.Event)
	Update()
	Render()
	Close()
}

// WindowManager keeps track of every open SDL window.
type WindowManager struct {
	windows map[uint32]Window
}

// NewWindowManager returns an initialized WindowManager.
func NewWindowManager() *WindowManager {
	return &WindowManager{windows: make(map[uint32]Window)}
}

// Add registers a window to be managed.
func (wm *WindowManager) Add(w Window) {
	wm.windows[w.ID()] = w
}

// Remove closes and unregisters the window with the given ID.
func (wm *WindowManager) Remove(id uint32) {
	if w, ok := wm.windows[id]; ok {
		w.Close()
		delete(wm.windows, id)
	}
}

// Get retrieves a window reference by ID.
func (wm *WindowManager) Get(id uint32) Window {
	return wm.windows[id]
}

// HandleEvent routes an SDL event to the appropriate window.
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
		// dispatch to all windows when no specific target exists (e.g. controller events)
		for _, w := range wm.windows {
			w.HandleEvent(event)
		}
	}
}

// RenderAll updates and renders every managed window.
func (wm *WindowManager) RenderAll() {
	for _, w := range wm.windows {
		w.Update()
		w.Render()
	}
}

// CloseAll releases every window.
func (wm *WindowManager) CloseAll() {
	for id := range wm.windows {
		wm.Remove(id)
	}
}

func (wm *WindowManager) dispatch(windowID uint32, event sdl.Event) {
	if windowID == 0 {
		return
	}
	if w, ok := wm.windows[windowID]; ok {
		w.HandleEvent(event)
	}
}
