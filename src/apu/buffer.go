package apu

import "sync"

// リングバッファ構造体
type RingBuffer struct {
	buffer    [BUFFER_SIZE]uint8
	writePos  int
	readPos   int
	mutex     sync.RWMutex
}

func (rb *RingBuffer) Write(data []uint8) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	for _, sample := range data {
		rb.buffer[rb.writePos] = sample
		rb.writePos = (rb.writePos + 1) % BUFFER_SIZE
	}
}

func (rb *RingBuffer) Read(data []uint8) int {
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

