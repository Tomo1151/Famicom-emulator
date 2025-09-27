package apu

import "sync"

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
	n := len(data)
	if n > available {
		n = available
	}

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
		rb.buffer[i] = 128
	}
	rb.writePos = 0
	rb.readPos = 0
}
