package ui

import (
	"Famicom-emulator/apu"
	"Famicom-emulator/ppu"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

// APUWindow renders per-channel waveforms in a square window.
// The window is vertically split into 5 equal regions for ch1..ch5.
type APUWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	apu      *apu.APU
	onClose  func(id uint32)
	baseW    int
	baseH    int
	scale    int
}

func NewAPUWindow(a *apu.APU, scale int, onClose func(id uint32)) (*APUWindow, error) {
	// Use two-screen width (like CHR viewer) and scale factor.
	// Width = 2 * SCREEN_WIDTH * scale, Height = SCREEN_HEIGHT * scale
	baseW := int(ppu.SCREEN_WIDTH * 2)
	baseH := int(ppu.SCREEN_HEIGHT)
	width := int32(baseW * scale)
	height := int32(baseH * scale)
	w, err := sdl.CreateWindow(
		"APU Viewer",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		width,
		height,
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

	return &APUWindow{window: w, renderer: r, apu: a, onClose: onClose, baseW: baseW, baseH: baseH, scale: scale}, nil
}

func (aw *APUWindow) ID() uint32 {
	id, _ := aw.window.GetID()
	return id
}

func (aw *APUWindow) HandleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_CLOSE {
			aw.requestClose()
		}
	case *sdl.KeyboardEvent:
		if e.State == sdl.PRESSED {
			switch e.Keysym.Sym {
			case sdl.K_ESCAPE:
				aw.requestClose()
			case sdl.K_UP:
				aw.setScale(aw.scale + 1)
			case sdl.K_DOWN:
				aw.setScale(aw.scale - 1)
			}
		}
	}
}

func (aw *APUWindow) setScale(s int) {
	if s < 1 {
		s = 1
	}
	if s > 8 {
		s = 8
	}
	if s == aw.scale {
		return
	}
	aw.scale = s
	if aw.window != nil {
		aw.window.SetSize(int32(aw.baseW*aw.scale), int32(aw.baseH*aw.scale))
	}
}

func (aw *APUWindow) Update() {
	// No stateful updates required; samples are read during Render().
}

func (aw *APUWindow) Render() {
	if aw.renderer == nil {
		return
	}
	aw.renderer.SetDrawColor(16, 16, 16, 255)
	aw.renderer.Clear()

	// Sample count used by the audio callback
	n := apu.AudioCallbackSampleCount()
	samples := apu.GetRecentChannelSamples(n)
	if samples == nil {
		aw.renderer.Present()
		return
	}

	w, h := aw.window.GetSize()
	width := int(w)
	height := int(h)
	regionH := height / 5
	// thickness in pixels scaled by UI scale (approx)
	thickness := int32(1)
	if aw.scale >= 2 {
		thickness = int32(aw.scale)
	}
	if thickness < 1 {
		thickness = 1
	}

	for ch := 0; ch < 5; ch++ {
		// background for region
		r := sdl.Rect{X: 0, Y: int32(ch * regionH), W: int32(width), H: int32(regionH)}
		if ch%2 == 0 {
			aw.renderer.SetDrawColor(24, 24, 24, 255)
		} else {
			aw.renderer.SetDrawColor(20, 20, 20, 255)
		}
		aw.renderer.FillRect(&r)

		// draw mid-line
		aw.renderer.SetDrawColor(80, 80, 80, 255)
		midY := int32(ch*regionH + regionH/2)
		aw.renderer.DrawLine(0, midY, int32(width), midY)

		// draw waveform as a connected polyline (thickened)
		aw.renderer.SetDrawColor(160, 220-uint8(ch*20), 80+uint8(ch*20), 255)
		channelSamples := samples[ch]
		if len(channelSamples) == 0 {
			continue
		}
		// We'll compute a y for each x across the window and connect with DrawLine.
		var prevSet bool
		var prevX, prevY int32
		sampleCount := len(channelSamples)
		for x := 0; x < width; x++ {
			// Map x -> fractional sample index and linearly interpolate
			if sampleCount == 0 {
				break
			}
			idxF := 0.0
			if width > 1 {
				idxF = float64(x) * float64(sampleCount-1) / float64(width-1)
			}
			idx0 := int(math.Floor(idxF))
			if idx0 < 0 {
				idx0 = 0
			}
			if idx0 >= sampleCount {
				idx0 = sampleCount - 1
			}
			frac := idxF - float64(idx0)
			s0 := float64(channelSamples[idx0])
			s1 := s0
			if idx0+1 < sampleCount {
				s1 = float64(channelSamples[idx0+1])
			}
			val := s0*(1.0-frac) + s1*frac
			var denom float64 = 1.0
			if ch >= 0 && ch <= 3 {
				denom = 15.0
			} else {
				denom = 127.0
			}
			if denom != 0 {
				val = val / denom
			}
			if val > 1.0 {
				val = 1.0
			} else if val < -1.0 {
				val = -1.0
			}
			yf := float64(regionH) / 2.0 * val
			y := int32(ch*regionH) + int32(regionH/2) - int32(yf)
			cx := int32(x)
			cy := y

			if prevSet {
				// draw main line
				aw.renderer.DrawLine(prevX, prevY, cx, cy)
				// thicken by drawing parallel horizontal offsets
				for t := int32(1); t < thickness; t++ {
					aw.renderer.DrawLine(prevX, prevY+t, cx, cy+t)
					aw.renderer.DrawLine(prevX, prevY-t, cx, cy-t)
				}
			}
			// draw a small square at the point to ensure visibility on thin segments
			pr := sdl.Rect{X: cx - thickness/2, Y: cy - thickness/2, W: thickness, H: thickness}
			aw.renderer.FillRect(&pr)
			prevX = cx
			prevY = cy
			prevSet = true
		}
	}

	aw.renderer.Present()
}

func (aw *APUWindow) Close() {
	if aw.renderer != nil {
		aw.renderer.Destroy()
	}
	if aw.window != nil {
		aw.window.Destroy()
	}
	aw.renderer = nil
	aw.window = nil
}

func (aw *APUWindow) requestClose() {
	if aw.onClose != nil {
		aw.onClose(aw.ID())
	}
}
