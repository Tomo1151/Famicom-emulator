package ui

import (
	"Famicom-emulator/config"

	"github.com/veandco/go-sdl2/sdl"
)

// OptionWindow renders a simple configuration panel in a separate window.
type OptionWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	config   *config.Config
	onClose  func(id uint32)
}

// NewOptionWindow creates a window for tweaking runtime options.
func NewOptionWindow(cfg *config.Config, onClose func(id uint32)) (*OptionWindow, error) {
	w, err := sdl.CreateWindow(
		"Options",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		360,
		240,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, err
	}

	r, err := sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		w.Destroy()
		return nil, err
	}

	return &OptionWindow{window: w, renderer: r, config: cfg, onClose: onClose}, nil
}

// ID returns the SDL identifier for the window.
func (o *OptionWindow) ID() uint32 {
	id, _ := o.window.GetID()
	return id
}

// HandleEvent closes the window on relevant events.
func (o *OptionWindow) HandleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_CLOSE {
			o.requestClose()
		}
	case *sdl.KeyboardEvent:
		if e.State == sdl.PRESSED {
			switch e.Keysym.Sym {
			case sdl.K_ESCAPE, sdl.K_ASTERISK, sdl.K_KP_MULTIPLY:
				o.requestClose()
			}
		}
	}
}

// Update currently performs no state updates.
func (o *OptionWindow) Update() {}

// Render draws a placeholder UI representing the configuration.
func (o *OptionWindow) Render() {
	opacity := uint8(180)
	if o.renderer == nil {
		return
	}
	o.renderer.SetDrawColor(30, 30, 30, opacity)
	o.renderer.Clear()

	// Draw simple bars to visualize current config state (placeholder UI).
	o.renderer.SetDrawColor(80, 160, 255, opacity)
	scaleWidth := int32(float32(300) * float32(o.config.ScaleFactor) / 6.0)
	optionRect := sdl.Rect{X: 30, Y: 40, W: scaleWidth, H: 20}
	o.renderer.FillRect(&optionRect)

	o.renderer.SetDrawColor(255, 200, 80, opacity)
	volumeWidth := int32(float32(300) * o.config.SoundVolume)
	volumeRect := sdl.Rect{X: 30, Y: 100, W: volumeWidth, H: 20}
	o.renderer.FillRect(&volumeRect)

	if o.config.ShowFPS {
		o.renderer.SetDrawColor(120, 255, 120, opacity)
		fpsRect := sdl.Rect{X: 30, Y: 160, W: 300, H: 20}
		o.renderer.FillRect(&fpsRect)
	}

	o.renderer.Present()
}

// Close frees SDL resources.
func (o *OptionWindow) Close() {
	if o.renderer != nil {
		o.renderer.Destroy()
	}
	if o.window != nil {
		o.window.Destroy()
	}
	// Prevent multiple invocations.
	o.renderer = nil
	o.window = nil
}

func (o *OptionWindow) requestClose() {
	if o.onClose != nil {
		o.onClose(o.ID())
	}
}
