package ui

import (
	"Famicom-emulator/apu"
	"Famicom-emulator/ppu"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

// MARK: AudioWindow の定義
type AudioWindow struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	apu      *apu.APU
	onClose  func(id uint32)
	baseW    int
	baseH    int
	scale    int
}

// MARK: APUウィンドウの作成メソッド
func NewAudioWindow(a *apu.APU, scale int, onClose func(id uint32)) (*AudioWindow, error) {
	baseW := int(ppu.SCREEN_WIDTH * 2)
	baseH := int(ppu.SCREEN_HEIGHT)
	width := int32(baseW * scale)
	height := int32(baseH * scale)
	w, err := sdl.CreateWindow(
		"Audio Visualizer",
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

	return &AudioWindow{window: w, renderer: r, apu: a, onClose: onClose, baseW: baseW, baseH: baseH, scale: scale}, nil
}

// MARK: ウィンドウのID取得メソッド
func (aw *AudioWindow) ID() uint32 {
	id, _ := aw.window.GetID()
	return id
}

// MARK: イベント処理メソッド
func (aw *AudioWindow) HandleEvent(event sdl.Event) {
	// ウィンドウ固有のイベントを処理する（閉じる・スケール変更）
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
			case sdl.K_PLUS, sdl.K_SEMICOLON:
				aw.setScale(aw.scale + 1)
			case sdl.K_MINUS:
				aw.setScale(aw.scale - 1)
			}
		}
	}
}

// MARK: スケール設定メソッド
func (aw *AudioWindow) setScale(s int) {
	// 1 ~ 8 の間に設定
	s = min(max(s, 1), 8)
	if s == aw.scale {
		return
	}
	if s == aw.scale {
		return
	}
	aw.scale = s
	if aw.window != nil {
		aw.window.SetSize(int32(aw.baseW*aw.scale), int32(aw.baseH*aw.scale))
	}
}

// MARK: 更新メソッド
func (aw *AudioWindow) Update() {}

// MARK: 描画メソッド
func (aw *AudioWindow) Render() {
	if aw.renderer == nil {
		return
	}
	aw.renderer.SetDrawColor(16, 16, 16, 255)
	aw.renderer.Clear()

	// オーディオコールバックで渡されるサンプル数を取得し、同数分を参照する。
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

	thickness := int32(1)
	if aw.scale >= 2 {
		thickness = int32(aw.scale)
	}
	if thickness < 1 {
		thickness = 1
	}

	for ch := range 5 {
		// 各チャンネル領域の背景を塗る
		r := sdl.Rect{X: 0, Y: int32(ch * regionH), W: int32(width), H: int32(regionH)}
		if ch%2 == 0 {
			aw.renderer.SetDrawColor(24, 24, 24, 255)
		} else {
			aw.renderer.SetDrawColor(20, 20, 20, 255)
		}
		aw.renderer.FillRect(&r)

		// 中央の基準線を描画
		aw.renderer.SetDrawColor(80, 80, 80, 255)
		midY := int32(ch*regionH + regionH/2)
		aw.renderer.DrawLine(0, midY, int32(width), midY)

		// 波形をポリラインとして描画（太線化あり）
		aw.renderer.SetDrawColor(160, 220-uint8(ch*20), 80+uint8(ch*20), 255)
		channelSamples := samples[ch]
		if len(channelSamples) == 0 {
			continue
		}

		// X方向の各画素に対応するサンプル位置を求め、線形補間してY座標を算出する。
		var prevSet bool
		var prevX, prevY int32
		sampleCount := len(channelSamples)
		for x := 0; x < width; x++ {
			if sampleCount == 0 {
				break
			}
			idxF := 0.0
			if width > 1 {
				idxF = float64(x) * float64(sampleCount-1) / float64(width-1)
			}
			idx0 := max(int(math.Floor(idxF)), 0)
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

			// チャンネル毎の振幅レンジで正規化する。
			// pulse系(1-4ch)は0..15，DMCは0..127程度の幅
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
				// メインの線を引き、視認性のために平行オフセットを描いて太線化する。
				aw.renderer.DrawLine(prevX, prevY, cx, cy)
				for t := int32(1); t < thickness; t++ {
					aw.renderer.DrawLine(prevX, prevY+t, cx, cy+t)
					aw.renderer.DrawLine(prevX, prevY-t, cx, cy-t)
				}
			}
			// 各点に小さな長方形を描いて、細いセグメントも見えるようにする。
			pr := sdl.Rect{X: cx - thickness/2, Y: cy - thickness/2, W: thickness, H: thickness}
			aw.renderer.FillRect(&pr)
			prevX = cx
			prevY = cy
			prevSet = true
		}
	}

	aw.renderer.Present()
}

// MARK: SDLリソースの解放メソッド
func (aw *AudioWindow) Close() {
	if aw.renderer != nil {
		aw.renderer.Destroy()
	}
	if aw.window != nil {
		aw.window.Destroy()
	}
	aw.renderer = nil
	aw.window = nil
}

// MARK: ウィンドウを閉じるメソッド
func (aw *AudioWindow) requestClose() {
	if aw.onClose != nil {
		aw.onClose(aw.ID())
	}
}
