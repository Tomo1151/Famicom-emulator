package apu

import "sync"

// リングバッファ構造体
type RingBuffer struct {
	buffer    [BUFFER_SIZE]float32
	writePos  int
	readPos   int
	mutex     sync.RWMutex
}

func (rb *RingBuffer) Write(data []float32) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	for _, sample := range data {
		rb.buffer[rb.writePos] = sample
		rb.writePos = (rb.writePos + 1) % BUFFER_SIZE
	}
}

func (rb *RingBuffer) Read(data []float32) int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	
	readCount := 0
	for i := range data {
		if rb.readPos == rb.writePos {
			// バッファが空の場合は無音（128はU8フォーマットの中央値）
			data[i] = 128
		} else {
			data[i] = rb.buffer[rb.readPos]
			rb.readPos = (rb.readPos + 1) % BUFFER_SIZE
			readCount++
		}
	}
	return readCount
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

