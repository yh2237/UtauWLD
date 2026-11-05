package worldrender

import (
	"container/list"
	"fmt"
	"math"
)

const worldFramePeriodMS = FramePeriodMS

// ネイティブのWORLD ABIはサンプル数・サンプルレート・フレーム数をint32で受ける。
const maxWorldInt32 = math.MaxInt32

// unitD4CThresholdは原音接続の解析に使うD4Cのしきい値。0にして有声判定を使わない。
// WORLDの既定の0.85では、足立レイの「い」のような定常の母音が無声と判定され、雑音になる。
const unitD4CThreshold = 0.0

type worldFeatures struct {
	Frames        int
	FFTSize       int
	F0            []float64
	Spectrum      []float64
	Aperiodicity  []float64
	FramePeriodMS float64
}

func (features worldFeatures) periodMS() float64 {
	if features.FramePeriodMS > 0 {
		return features.FramePeriodMS
	}
	return worldFramePeriodMS
}

type worldEngine interface {
	Close() error
	Analyze([]float64, int, []float64) (worldFeatures, error)
	Synthesize(worldFeatures, int) ([]float64, error)
}

type worldFeatureCacheEntry struct {
	key   string
	value cachedWorldUnit
	bytes int64
}

type worldFeatureCache struct {
	capacity int
	maxBytes int64
	bytes    int64
	order    *list.List
	entries  map[string]*list.Element
}

func newWorldFeatureCache(capacity int) *worldFeatureCache {
	return newWorldFeatureCacheWithLimit(capacity, 256<<20)
}

func newWorldFeatureCacheWithLimit(capacity int, maxBytes int64) *worldFeatureCache {
	return &worldFeatureCache{
		capacity: max(1, capacity), maxBytes: max(1, maxBytes),
		order: list.New(), entries: make(map[string]*list.Element),
	}
}

func (cache *worldFeatureCache) get(key string) (cachedWorldUnit, bool) {
	if cache == nil {
		return cachedWorldUnit{}, false
	}
	element, found := cache.entries[key]
	if !found {
		return cachedWorldUnit{}, false
	}
	cache.order.MoveToFront(element)
	return element.Value.(worldFeatureCacheEntry).value, true
}

func (cache *worldFeatureCache) put(key string, value cachedWorldUnit) {
	if cache == nil {
		return
	}
	if element, found := cache.entries[key]; found {
		previous := element.Value.(worldFeatureCacheEntry)
		entryBytes := worldFeatureCacheBytes(value)
		cache.bytes += entryBytes - previous.bytes
		element.Value = worldFeatureCacheEntry{key: key, value: value, bytes: entryBytes}
		cache.order.MoveToFront(element)
		cache.evict()
		return
	}
	entryBytes := worldFeatureCacheBytes(value)
	element := cache.order.PushFront(worldFeatureCacheEntry{key: key, value: value, bytes: entryBytes})
	cache.entries[key] = element
	cache.bytes += entryBytes
	cache.evict()
}

func (cache *worldFeatureCache) evict() {
	for cache.order.Len() > cache.capacity || cache.bytes > cache.maxBytes {
		oldest := cache.order.Back()
		if oldest == nil {
			return
		}
		entry := oldest.Value.(worldFeatureCacheEntry)
		cache.bytes -= entry.bytes
		delete(cache.entries, entry.key)
		cache.order.Remove(oldest)
	}
}

func worldFeatureCacheBytes(value cachedWorldUnit) int64 {
	features := value.features
	return int64(len(features.F0)+len(features.Spectrum)+len(features.Aperiodicity)) * 8
}

func worldSynthesisLengthAt(frames, sampleRate int, frameMS float64) int {
	length, err := worldSynthesisLength(frames, sampleRate, frameMS)
	if err != nil {
		return 0
	}
	return length
}

// worldSynthesisLengthはネイティブのsynthesis_lengthと同じ式で出力長を求める。
// 範囲外は0ではなくエラーにする。int変換のあふれをmakeのpanicにしないため。
func worldSynthesisLength(frames, sampleRate int, frameMS float64) (int, error) {
	if frames < 2 || frames > maxWorldInt32 || sampleRate <= 0 || sampleRate > maxWorldInt32 {
		return 0, fmt.Errorf("invalid WORLD synthesis shape or sample rate")
	}
	if math.IsNaN(frameMS) || math.IsInf(frameMS, 0) || frameMS <= 0 {
		return 0, fmt.Errorf("invalid WORLD frame period: %v", frameMS)
	}
	length := float64(frames-1) * frameMS / 1000 * float64(sampleRate)
	if math.IsNaN(length) || math.IsInf(length, 0) || length >= float64(maxWorldInt32) {
		return 0, fmt.Errorf("WORLD synthesis length is out of range")
	}
	return int(length) + 1, nil
}

func validateWorldFeatures(features worldFeatures) error {
	bins := features.FFTSize/2 + 1
	if features.Frames < 2 || features.Frames > maxWorldInt32 || features.FFTSize < 2 || features.FFTSize > maxWorldInt32 ||
		len(features.F0) != features.Frames ||
		len(features.Spectrum) != features.Frames*bins || len(features.Aperiodicity) != features.Frames*bins {
		return fmt.Errorf("invalid WORLD feature shape")
	}
	return nil
}

// validateWorldSynthesisは公開入口で特徴量の形・サンプルレート・フレーム周期を確かめる。
// 0以下のフレーム周期は従来どおり既定値として扱い、非有限値は拒否する。
func validateWorldSynthesis(features worldFeatures, sampleRate int) error {
	if err := validateWorldFeatures(features); err != nil {
		return err
	}
	if sampleRate <= 0 || sampleRate > maxWorldInt32 {
		return fmt.Errorf("WORLD sample rate is out of range: %d", sampleRate)
	}
	frameMS := features.FramePeriodMS
	if math.IsNaN(frameMS) || math.IsInf(frameMS, 0) {
		return fmt.Errorf("WORLD frame period must be finite: %v", frameMS)
	}
	if frameMS <= 0 {
		frameMS = worldFramePeriodMS
	}
	_, err := worldSynthesisLength(features.Frames, sampleRate, frameMS)
	return err
}
