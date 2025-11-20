package main

import (
	"Famicom-emulator/apu"
	"Famicom-emulator/bus"
	"Famicom-emulator/cartridge"
	"Famicom-emulator/cpu"
	"Famicom-emulator/joypad"
	"Famicom-emulator/ppu"
	"log"
	"runtime"
	"time"

	// gotk3 (GTK3 bindings)
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	// HID (gamepad stub)
	"github.com/sstallion/go-hid"
)

// MARK: 定数定義
const (
	FRAME_PER_SECOND = 60
	// スケール設定
	SCALE_FACTOR = 3
)

// MARK: InputState の定義
type InputState struct {
	Left, Right, Up, Down bool
	A, B, Start, Select   bool
}
type Famicom struct {
	cpu       cpu.CPU
	ppu       ppu.PPU
	apu       apu.APU
	joypad1   joypad.JoyPad
	joypad2   joypad.JoyPad
	bus       bus.Bus
	cartridge cartridge.Cartridge

	keyboard1   InputState // 1P キーボード
	keyboard2   InputState // 2P キーボード
	controller1 InputState // 1P Gamepad (HID)
	controller2 InputState // 2P Gamepad (HID)

	// フレーム転送用バッファとフラグ (GTKメインスレッドでのみUI更新)
	frameBuf   []byte
	destPixbuf *gdk.Pixbuf
}

// MARK: Famicomの初期化メソッド
func (f *Famicom) Init(cartridge cartridge.Cartridge) {
	// ROMファイルのロード
	f.cartridge = cartridge
	err := f.cartridge.Load()
	if err != nil {
		log.Fatalf("Cartridge loading error: %v", err)
	}

	// 各コンポーネントの定義 / 接続
	f.cpu = cpu.CPU{}
	f.ppu = ppu.PPU{}
	f.apu = apu.APU{}
	f.joypad1 = joypad.JoyPad{}
	f.joypad2 = joypad.JoyPad{}
	f.bus = bus.Bus{}
	f.bus.ConnectComponents(
		&f.ppu,
		&f.apu,
		&f.cartridge,
		&f.joypad1,
		&f.joypad2,
	)

	// 入力データの定義
	f.keyboard1 = InputState{}
	f.keyboard2 = InputState{}
	f.controller1 = InputState{}
	f.controller2 = InputState{}
}

// MARK: Famicomの起動
func (f *Famicom) Start() {
	// GTKはメインスレッド固定が必要
	runtime.LockOSThread()
	// GTK 初期化 (gotk3)
	gtk.Init(nil)

	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatalf("GTK window error: %v", err)
	}
	window.SetTitle("Famicom emu (GTK3)")
	window.Connect("destroy", func() {
		f.bus.Shutdown()
		gtk.MainQuit()
	})

	// Image ウィジェットで表示 (Pixbuf を毎フレーム更新)
	image, err := gtk.ImageNew()
	if err != nil {
		log.Fatalf("GTK image error: %v", err)
	}
	window.Add(image)

	// キーイベント (GTK3)
	window.SetCanFocus(true)
	window.Connect("key-press-event", func(_ *gtk.Window, ev *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(ev)
		f.mapKey(uint(keyEvent.KeyVal()), true)
	})
	window.Connect("key-release-event", func(_ *gtk.Window, ev *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(ev)
		f.mapKey(uint(keyEvent.KeyVal()), false)
	})

	// HID 初期化 (ゲームパッド探索: 簡易スタブ)
	if err := hid.Init(); err == nil {
		defer hid.Exit()
		// TODO: enumerate devices & map buttons
	} else {
		log.Printf("HID init failed: %v", err)
	}

	// 表示サイズ設定
	originalWidth := int(ppu.SCREEN_WIDTH)
	originalHeight := int(ppu.SCREEN_HEIGHT)
	windowWidth := originalWidth * SCALE_FACTOR
	windowHeight := originalHeight * SCALE_FACTOR

	// 初期Pixbuf (黒画面)
	pixbuf, perr := gdk.PixbufNew(gdk.COLORSPACE_RGB, false, 8, windowWidth, windowHeight)
	if perr != nil {
		log.Fatalf("Pixbuf create error: %v", perr)
	}
	image.SetFromPixbuf(pixbuf)
	window.SetDefaultSize(windowWidth, windowHeight)
	image.SetSizeRequest(windowWidth, windowHeight)
	window.SetResizable(false)
	window.ShowAll()

	// フレームバッファ初期化 (元解像度256x240のRGBデータ)
	f.frameBuf = make([]byte, originalWidth*originalHeight*3)

	// destPixbuf 初期化 (スケーリング後の描画先)
	var perr2 error
	f.destPixbuf, perr2 = gdk.PixbufNew(gdk.COLORSPACE_RGB, false, 8, windowWidth, windowHeight)
	if perr2 != nil {
		log.Fatalf("destPixbuf create error: %v", perr2)
	}
	// 初期表示
	image.SetFromPixbuf(f.destPixbuf)

	// チャンネル作成
	renderChan := make(chan struct{})
	resumeChan := make(chan struct{})

	// NMI コールバック: フレーム生成のみ (UI操作しない)
	f.bus.Init(func(p *ppu.PPU, c *ppu.Canvas, j1 *joypad.JoyPad, j2 *joypad.JoyPad) {
		// JoyPad 状態更新 (入力は即時反映でOK)
		f.updateJoyPad(j1, &f.keyboard1, &f.controller1)
		f.updateJoyPad(j2, &f.keyboard2, &f.controller2)

		// 内部キャンバス 512x240 の左半分 (256x240) だけを表示用にコピー
		bufFull := c.Buffer[:]
		rowBytes := int(ppu.CANVAS_WIDTH) * 3
		visibleRowBytes := originalWidth * 3
		for y := range originalHeight {
			srcRowStart := y * rowBytes
			copy(f.frameBuf[y*visibleRowBytes:(y+1)*visibleRowBytes], bufFull[srcRowStart:srcRowStart+visibleRowBytes])
		}

		// 描画準備完了を通知
		renderChan <- struct{}{}

		// メインスレッドの描画完了を待機 (FPS制御)
		<-resumeChan
	})

	// FPS制御用タイマー (メインスレッド側で制御)
	go func() {
		ticker := time.NewTicker(time.Second / FRAME_PER_SECOND)
		defer ticker.Stop()
		for range ticker.C {
			glib.IdleAdd(func() bool {
				select {
				case <-renderChan:
					// 描画処理: 毎回 PixbufNewFromData で一時的な Pixbuf を作成し、destPixbuf へスケーリング転送する
					srcPixbuf, err := gdk.PixbufNewFromData(f.frameBuf, gdk.COLORSPACE_RGB, false, 8, originalWidth, originalHeight, originalWidth*3)
					if err == nil {
						// destPixbuf へスケーリングして描画
						srcPixbuf.Scale(f.destPixbuf, 0, 0, windowWidth, windowHeight, 0, 0, float64(SCALE_FACTOR), float64(SCALE_FACTOR), gdk.INTERP_NEAREST)
						// 再描画要求 (SetFromPixbuf を再度呼ぶことで更新を確実にする)
						image.SetFromPixbuf(f.destPixbuf)
					} else {
						log.Printf("Pixbuf create error: %v", err)
					}

					// CPU再開
					resumeChan <- struct{}{}
				default:
					// CPUがまだフレームを生成していない場合はスキップ
				}
				return false
			})
		}
	}()

	// CPU 初期化 & 実行
	f.cpu.Init(f.bus, false)
	go f.cpu.Run()

	gtk.Main()
}

// MARK: キーボードの状態を反映
func (f *Famicom) mapKey(keyVal uint, pressed bool) {
	// 1P
	switch keyVal {
	case gdk.KEY_k:
		f.keyboard1.A = pressed
	case gdk.KEY_j:
		f.keyboard1.B = pressed
	case gdk.KEY_w:
		f.keyboard1.Up = pressed
	case gdk.KEY_s:
		f.keyboard1.Down = pressed
	case gdk.KEY_a:
		f.keyboard1.Left = pressed
	case gdk.KEY_d:
		f.keyboard1.Right = pressed
	case gdk.KEY_Return:
		f.keyboard1.Start = pressed
	case gdk.KEY_BackSpace:
		f.keyboard1.Select = pressed
	// 2P
	case gdk.KEY_colon:
		f.keyboard2.A = pressed
	case gdk.KEY_semicolon:
		f.keyboard2.B = pressed
	case gdk.KEY_t:
		f.keyboard2.Up = pressed
	case gdk.KEY_g:
		f.keyboard2.Down = pressed
	case gdk.KEY_f:
		f.keyboard2.Left = pressed
	case gdk.KEY_h:
		f.keyboard2.Right = pressed
	case gdk.KEY_greater:
		f.keyboard2.Select = pressed
	case gdk.KEY_Tab:
		f.keyboard2.Start = pressed
	case gdk.KEY_Escape:
		if pressed {
			f.bus.Shutdown()
			gtk.MainQuit()
		}
	}
}

// MARK: コントローラーのボタン状態を検知
// HID Gamepad スタブ (未実装詳細)
func (f *Famicom) pollGamepads() {
	// TODO: 実際の HID デバイス列挙とボタン/軸のマッピング
	// devices, _ := hid.Enum() // 利用可能デバイス一覧
	// 今はノーオペレーション
}

// MARK: コントローラーのスティック状態を検知
// 軸入力スタブ
func (f *Famicom) updateAxisStub(c *InputState) {
	// TODO: HID 軸入力の反映
}

// MARK: JoyPadの状態を更新
func (f *Famicom) updateJoyPad(j *joypad.JoyPad, k *InputState, c *InputState) {
	// キーボードとコントローラの入力を統合
	buttonA := k.A || c.A
	buttonB := k.B || c.B
	buttonStart := k.Start || c.Start
	buttonSelect := k.Select || c.Select
	buttonUp := k.Up || c.Up
	buttonDown := k.Down || c.Down
	buttonLeft := k.Left || c.Left
	buttonRight := k.Right || c.Right

	// JoyPadの状態を更新
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_A_POSITION, buttonA)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_B_POSITION, buttonB)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_START_POSITION, buttonStart)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_SELECT_POSITION, buttonSelect)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_UP_POSITION, buttonUp)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_DOWN_POSITION, buttonDown)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_LEFT_POSITION, buttonLeft)
	j.SetButtonPressed(joypad.JOYPAD_BUTTON_RIGHT_POSITION, buttonRight)
}
