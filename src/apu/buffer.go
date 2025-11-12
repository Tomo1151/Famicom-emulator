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

func (rb *RingBuffer) Buffer() [BUFFER_SIZE]float32 { return rb.buffer }
func (rb *RingBuffer) testBuffer() {
	for i := range 100 {
		rb.buffer[i] = 1.0
	}
}

type BlipBuffer struct {
	sampleRate float64
	tickRate   float64
	lastTime   uint64
	lastLevel  float32 // 絶対レベルを保持
	frac       float64 // サンプル生成の余り
	samples    []float32
	mutex      sync.Mutex // RWMutexからMutexに変更
}

func (b *BlipBuffer) Init() {
	b.sampleRate = float64(SAMPLE_RATE)
	b.tickRate = float64(CPU_CLOCK)
	b.samples = make([]float32, 0, 8192)
}

// time: APUクロック, delta: レベルの差分
func (b *BlipBuffer) addDelta(time uint64, delta float32) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 前回イベントから今回イベントまでの時間を、前回レベルで埋める
	dt := time - b.lastTime
	if dt > 0 {
		// 生成すべきサンプル数を計算
		samplesToGen := b.frac + (float64(dt) * b.sampleRate / b.tickRate)
		count := int(samplesToGen)

		if count > 0 {
			// ゼロ次ホールド：この区間はずっと同じレベル
			for i := 0; i < count; i++ {
				b.samples = append(b.samples, b.lastLevel)
			}
			b.frac = samplesToGen - float64(count)
		} else {
			b.frac = samplesToGen
		}
	}

	b.lastTime = time
	b.lastLevel += delta // レベルを更新
}

// Blip_Buffer の読み取り
func (b *BlipBuffer) Read(out []float32, count int) int {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	n := min(len(b.samples), count)
	if n > 0 {
		copy(out, b.samples[:n])
		b.samples = b.samples[n:]
	}

	// 足りない分を0で埋める
	for i := n; i < count && i < len(out); i++ {
		out[i] = 0
	}
	return n
}

// バッファのフラッシュ
func (b *BlipBuffer) endFrame(time uint64) {
	// 最後のイベントから現在時刻までをフラッシュ
	b.addDelta(time, 0)
}
