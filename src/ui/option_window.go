package ui

import (
	"Famicom-emulator/config"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: OptionWindow の定義
type OptionWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	config   *config.Config
	onClose  func(id uint32)
}

// MARK: OptionWindow の作成メソッド
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

// MARK: ウィンドウのID取得メソッド
func (o *OptionWindow) ID() uint32 {
	id, _ := o.window.GetID()
	return id
}

// MARK: イベント処理メソッド
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

// MARK: 更新メソッド
func (o *OptionWindow) Update() {}

// MARK: 描画メソッド
func (o *OptionWindow) Render() {
	opacity := uint8(180)
	if o.renderer == nil {
		return
	}
	o.renderer.SetDrawColor(30, 30, 30, opacity)
	o.renderer.Clear()

	o.renderer.SetDrawColor(80, 160, 255, opacity)
	scaleWidth := int32(float32(300) * float32(o.config.Render.SCALE_FACTOR) / 6.0)
	optionRect := sdl.Rect{X: 30, Y: 40, W: scaleWidth, H: 20}
	o.renderer.FillRect(&optionRect)

	o.renderer.SetDrawColor(255, 200, 80, opacity)
	volumeWidth := int32(float32(300) * o.config.APU.SOUND_VOLUME)
	volumeRect := sdl.Rect{X: 30, Y: 100, W: volumeWidth, H: 20}
	o.renderer.FillRect(&volumeRect)

	o.renderer.Present()
}

// MARK: SDLリソースの解放メソッド
func (o *OptionWindow) Close() {
	if o.renderer != nil {
		o.renderer.Destroy()
	}
	if o.window != nil {
		o.window.Destroy()
	}

	o.renderer = nil
	o.window = nil
}

// MARK: ウィンドウを閉じるメソッド
func (o *OptionWindow) requestClose() {
	if o.onClose != nil {
		o.onClose(o.ID())
	}
}
