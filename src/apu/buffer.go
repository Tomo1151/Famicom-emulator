package apu

import (
	"sync"
)

// リングバッファ構造体
type RingBuffer struct {
	buffer   [BUFFER_SIZE]float32
	writePos int
	readPos  int
	mutex    sync.RWMutex
}

func (rb *RingBuffer) Write(data []float32) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	// 連続領域にできるだけ一括コピー
	n := len(data)
	if n == 0 {
		return
	}

	// 書き込み位置から末尾までの連続領域サイズ
	contiguous := BUFFER_SIZE - rb.writePos
	if n <= contiguous {
		// 一括コピー
		copy(rb.buffer[rb.writePos:], data)
		rb.writePos = (rb.writePos + n) % BUFFER_SIZE
	} else {
		// 分割コピー
		copy(rb.buffer[rb.writePos:], data[:contiguous])
		copy(rb.buffer[:], data[contiguous:])
		rb.writePos = n - contiguous
	}
}

func (rb *RingBuffer) Read(data []float32) int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	// 読み取り可能なデータ量を計算
	available := 0
	if rb.writePos >= rb.readPos {
		available = rb.writePos - rb.readPos
	} else {
		available = BUFFER_SIZE - rb.readPos + rb.writePos
	}

	if available == 0 {
		// データがない場合は無音（0.0）を返す
		for i := range data {
			data[i] = 0.0
		}
		return 0
	}

	// 読み取るデータ量（要求された量か利用可能な量の少ない方）
	n := min(len(data), available)

	// 連続領域からできるだけ一括コピー
	contiguous := BUFFER_SIZE - rb.readPos
	if n <= contiguous {
		// 一括コピー
		copy(data, rb.buffer[rb.readPos:rb.readPos+n])
		rb.readPos = (rb.readPos + n) % BUFFER_SIZE
	} else {
		// 分割コピー
		copy(data[:contiguous], rb.buffer[rb.readPos:])
		copy(data[contiguous:n], rb.buffer[:n-contiguous])
		rb.readPos = n - contiguous
	}

	// 残りを無音で埋める
	for i := n; i < len(data); i++ {
		data[i] = 0.0
	}

	return n
}

func (rb *RingBuffer) Available() int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	if rb.writePos >= rb.readPos {
		return rb.writePos - rb.readPos
	}
	return BUFFER_SIZE - rb.readPos + rb.writePos
}

// リングバッファ初期化メソッドを追加
func (rb *RingBuffer) Init() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	// バッファを無音（128）で初期化
	for i := range rb.buffer {
		rb.buffer[i] = 0.0
	}
	rb.writePos = 0
	rb.readPos = 0
}

type BlipBuffer struct {
	sampleRate float32
	tickRate   float32 // Tick 単位のクロック (CPUClock / 2)
	lastTime   uint64
	accum      float32
	samples    []float32
}

func (b *BlipBuffer) Init(sampleRate uint) {
	b.sampleRate = float32(sampleRate)
	b.tickRate = CPUClock / 2
	b.samples = make([]float32, 0, 4096)
}

// Tick単位の音量差分を追加
func (b *BlipBuffer) addDelta(time uint64, delta float32) {
	// 前回の時間からサンプルに変換
	dt := max(time-b.lastTime, 0)

	// Tickからサンプルレート
	n := int(float32(dt) * b.sampleRate / b.tickRate)
	if n == 0 {
		// Tick内での変化は累積
		b.accum += float32(delta)
		return
	}

	// 累積分を加えてn個のサンプルを生成
	value := float32(b.accum) + delta
	for range n {
		b.samples = append(b.samples, value)
	}

	// 累積をクリア
	b.accum = 0
	b.lastTime = time
}

// Blip_Buffer の読み取り
func (b *BlipBuffer) Read(out []float32, count int) int {
	n := min(len(b.samples), count)
	copy(out, b.samples[:n])
	b.samples = b.samples[n:]
	for i := n; i < count && i < len(out); i++ {
		out[i] = 0
	}
	return n
}

// バッファのフラッシュ
func (b *BlipBuffer) endFrame() {
	if b.accum != 0 {
		b.samples = append(b.samples, b.accum)
		b.accum = 0
	}
}
