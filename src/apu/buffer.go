package apu

import (
	"fmt"
	"math"
	"sync"
)

const (
	sincTapCount = 63
	sincCutoff   = 0.45 // 正規化カットオフ（Nyquist比）
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

	filterTaps  []float64
	filterState []float32
	filterIndex int
}

// MARK: BlipBufferの初期化メソッド
func (b *BlipBuffer) Init() {
	b.sampleRate = float64(SAMPLE_RATE)
	b.tickRate = float64(CPU_CLOCK)
	b.samples = make([]float32, 0, BUFFER_SIZE)
	b.filterTaps = designSincLowPass(sincTapCount, sincCutoff)
	b.filterState = make([]float32, len(b.filterTaps))
	b.filterIndex = 0
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

	var last float32 = b.lastLevel
	n := min(len(b.samples), count)

	if len(b.samples) < count {
		fmt.Printf("[BlipBuffer] Warning: couldn't read enough samples (want: %4d, got: %4d)\n", count, len(b.samples))
	}

	for i := range n {
		filtered := b.filterSample(b.samples[i])
		out[i] = filtered
		last = filtered
	}
	b.samples = b.samples[n:]

	// 足りない分を0で埋める
	for i := n; i < count && i < len(out); i++ {
		out[i] = last
	}
	return n
}

// MARK: バッファのフラッシュをするメソッド
func (b *BlipBuffer) endFrame(time uint64) {
	// 最後のイベントから現在時刻までをフラッシュ
	b.addDelta(time, 0)
}

// MARK: 時間のリセットをするメソッド
func (b *BlipBuffer) resetTime() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.lastTime = 0
	b.frac = 0
}

func (b *BlipBuffer) filterSample(sample float32) float32 {
	if len(b.filterTaps) == 0 {
		return sample
	}

	b.filterState[b.filterIndex] = sample

	sum := 0.0
	idx := b.filterIndex
	for i := range len(b.filterTaps) {
		sum += b.filterTaps[i] * float64(b.filterState[idx])
		idx--
		if idx < 0 {
			idx = len(b.filterState) - 1
		}
	}

	b.filterIndex++
	if b.filterIndex == len(b.filterState) {
		b.filterIndex = 0
	}
	return float32(sum)
}

func designSincLowPass(numTaps int, cutoff float64) []float64 {
	if numTaps%2 == 0 {
		numTaps++
	}
	taps := make([]float64, numTaps)
	mid := float64(numTaps-1) / 2
	var sum float64

	for n := range numTaps {
		x := float64(n) - mid
		var sinc float64
		if x == 0 {
			sinc = 1.0
		} else {
			sinc = math.Sin(2*math.Pi*cutoff*x) / (2 * math.Pi * cutoff * x)
		}
		window := 0.54 - 0.46*math.Cos(2*math.Pi*float64(n)/float64(numTaps-1))
		tap := 2 * cutoff * sinc * window
		taps[n] = tap
		sum += tap
	}

	for n := range taps {
		taps[n] /= sum
	}
	return taps
}

// MARK: ResamplingBufferの定義
type ResamplingBuffer struct {
	sampleRate float64
	tickRate   float64
	lastTime   uint64
	lastLevel  float32
	frac       float64
	samples    []float32
	mutex      sync.Mutex
}

// MARK: ResamplingBuffer初期化メソッド
func (b *ResamplingBuffer) Init() {
	b.sampleRate = float64(SAMPLE_RATE)
	b.tickRate = float64(CPU_CLOCK)
	b.samples = make([]float32, 0, BUFFER_SIZE)
}

// MARK: ResamplingBufferの書き込みメソッド
func (b *ResamplingBuffer) Write(time uint64, level float32) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	dt := time - b.lastTime
	if dt <= 0 {
		b.lastLevel = level
		return
	}

	samplesToGen := b.frac + (float64(dt) * b.sampleRate / b.tickRate)
	count := int(samplesToGen)

	if count > 0 {
		// 線形補間を行う
		step := (level - b.lastLevel) / float32(count)
		current := b.lastLevel

		for range count {
			b.samples = append(b.samples, current)
			current += step
		}
	}

	b.frac = samplesToGen - float64(count)
	b.lastTime = time
	b.lastLevel = level
}

// MARK: ResamplingBufferの読み取りメソッド
func (b *ResamplingBuffer) Read(out []float32, count int) int {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var last float32 = b.lastLevel
	n := min(len(b.samples), count)

	if len(b.samples) < count {
		fmt.Printf("[BlipBuffer] Warning: couldn't read enough samples (want: %4d, got: %4d)\n", count, len(b.samples))
	}

	if n > 0 {
		last = b.samples[n-1]
		copy(out, b.samples[:n])
		b.samples = b.samples[n:]
	}

	for i := n; i < count && i < len(out); i++ {
		out[i] = last
	}
	return n
}

// MARK: バッファのフラッシュをするメソッド
func (b *ResamplingBuffer) endFrame(time uint64) {
	b.Write(time, b.lastLevel)
}

// MARK: 時間のリセットをするメソッド
func (b *ResamplingBuffer) resetTime() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.lastTime = 0
	b.frac = 0
}
