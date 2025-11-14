package apu

import (
	"sync"
)

// BlipBuffer の定義
type BlipBuffer struct {
	sampleRate float64
	tickRate   float64
	lastTime   uint64
	lastLevel  float32
	frac       float64
	samples    []float32
	mutex      sync.Mutex
}

// MARK: BlipBufferの初期化メソッド
func (b *BlipBuffer) Init() {
	b.sampleRate = float64(SAMPLE_RATE)
	b.tickRate = float64(CPU_CLOCK)
	b.samples = make([]float32, 0, 8192)
}

// MARK: レベルの差分を追加するメソッド
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
			for range count {
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

// MARK: BlipBuffer の読み取りメソッド
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

// MARK: バッファのフラッシュをするメソッド
func (b *BlipBuffer) endFrame(time uint64) {
	// 最後のイベントから現在時刻までをフラッシュ
	b.addDelta(time, 0)
}
