package worldrender

import (
	"math"
	"testing"
)

type boundaryWorldEngine struct {
	analyzeCalls int
	synthCalls   int
	f0Calls      int
	openCalls    int
}

func (e *boundaryWorldEngine) Close() error { return nil }

func (e *boundaryWorldEngine) Analyze(samples []float64, sampleRate int, inputF0 []float64) (worldFeatures, error) {
	e.analyzeCalls++
	return worldFeatures{}, nil
}

func (e *boundaryWorldEngine) AnalyzeWith(samples []float64, sampleRate int, inputF0 []float64, frameMS, d4cThreshold float64) (worldFeatures, error) {
	e.analyzeCalls++
	return worldFeatures{}, nil
}

func (e *boundaryWorldEngine) Synthesize(features worldFeatures, sampleRate int) ([]float64, error) {
	e.synthCalls++
	return make([]float64, worldSynthesisLengthAt(features.Frames, sampleRate, features.periodMS())), nil
}

func (e *boundaryWorldEngine) f0(samples []float64, sampleRate int, frameMS float64) ([]float64, error) {
	e.f0Calls++
	return nil, nil
}

func (e *boundaryWorldEngine) openStream(sampleRate int, frameMS float64, fftSize, bufferSize, ringSize int) (worldStream, error) {
	e.openCalls++
	return &boundaryWorldStream{}, nil
}

type boundaryWorldStream struct{}

func (s *boundaryWorldStream) add(f0, spectrum, aperiodicity []float64) bool { return true }
func (s *boundaryWorldStream) synthesize(output []float64) bool              { return false }
func (s *boundaryWorldStream) close()                                        {}

func TestWorldSynthesizeRejectsInvalidRateAndFramePeriod(t *testing.T) {
	engine := &boundaryWorldEngine{}
	world := &World{engine: engine}
	f0 := []float64{200, 200}
	spectrum := []float64{1, 1, 1, 1}
	aperiodicity := []float64{0, 0, 0, 0}
	for name, tc := range map[string]struct {
		sampleRate int
		frameMS    float64
	}{
		"zero sample rate":      {0, 10},
		"negative sample rate":  {-1, 10},
		"huge sample rate":      {1_000_000_000_000_000_000, 10},
		"NaN frame period":      {16000, math.NaN()},
		"infinite frame period": {16000, math.Inf(1)},
		"huge frame period":     {16000, 1e300},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := world.Synthesize(f0, spectrum, aperiodicity, 2, tc.sampleRate, tc.frameMS); err == nil {
				t.Fatal("invalid synthesis options were accepted")
			}
		})
	}
	if engine.synthCalls != 0 {
		t.Fatalf("engine synthesis calls = %d, want 0", engine.synthCalls)
	}
}

func TestWorldSynthesizeKeepsDefaultFramePeriod(t *testing.T) {
	engine := &boundaryWorldEngine{}
	world := &World{engine: engine}
	if _, err := world.Synthesize([]float64{200, 200}, []float64{1, 1, 1, 1}, []float64{0, 0, 0, 0}, 2, 16000, 0); err != nil {
		t.Fatalf("zero frame period should use the default: %v", err)
	}
	if engine.synthCalls != 1 {
		t.Fatalf("engine synthesis calls = %d, want 1", engine.synthCalls)
	}
}

func TestWorldF0RejectsInvalidRateAndFramePeriod(t *testing.T) {
	engine := &boundaryWorldEngine{}
	world := &World{engine: engine}
	for name, tc := range map[string]struct {
		sampleRate int
		frameMS    float64
	}{
		"zero sample rate":      {0, 10},
		"negative sample rate":  {-1, 10},
		"huge sample rate":      {1_000_000_000_000_000_000, 10},
		"zero frame period":     {16000, 0},
		"negative frame period": {16000, -10},
		"NaN frame period":      {16000, math.NaN()},
		"infinite frame period": {16000, math.Inf(1)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := world.F0([]float64{0, 0}, tc.sampleRate, tc.frameMS); err == nil {
				t.Fatal("invalid F0 options were accepted")
			}
		})
	}
	if engine.f0Calls != 0 {
		t.Fatalf("engine F0 calls = %d, want 0", engine.f0Calls)
	}
}

func TestWorldAnalyzeRejectsInvalidOptions(t *testing.T) {
	engine := &boundaryWorldEngine{}
	world := &World{engine: engine}
	for name, tc := range map[string]struct {
		sampleRate int
		frameMS    float64
		d4c        float64
	}{
		"zero sample rate":      {0, 10, 0},
		"negative sample rate":  {-1, 10, 0},
		"huge sample rate":      {1_000_000_000_000_000_000, 10, 0},
		"zero frame period":     {16000, 0, 0},
		"NaN frame period":      {16000, math.NaN(), 0},
		"infinite frame period": {16000, math.Inf(1), 0},
		"NaN threshold":         {16000, 10, math.NaN()},
		"infinite threshold":    {16000, 10, math.Inf(1)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, _, _, err := world.Analyze([]float64{0, 0}, tc.sampleRate, tc.frameMS, tc.d4c); err == nil {
				t.Fatal("invalid analysis options were accepted")
			}
		})
	}
	if engine.analyzeCalls != 0 {
		t.Fatalf("engine analysis calls = %d, want 0", engine.analyzeCalls)
	}
}

func TestWorldOpenStreamRejectsInvalidShape(t *testing.T) {
	engine := &boundaryWorldEngine{}
	world := &World{engine: engine}
	for name, tc := range map[string]struct {
		sampleRate int
		frameMS    float64
		fftSize    int
		bufferSize int
	}{
		"zero sample rate":  {0, 10, 1024, 256},
		"zero frame period": {16000, 0, 1024, 256},
		"NaN frame period":  {16000, math.NaN(), 1024, 256},
		"zero fft size":     {16000, 10, 0, 256},
		"zero buffer size":  {16000, 10, 1024, 0},
		"huge fft size":     {16000, 10, 1_000_000_000_000_000_000, 256},
		"huge buffer size":  {16000, 10, 1024, 1_000_000_000_000_000_000},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := world.OpenStream(tc.sampleRate, tc.frameMS, tc.fftSize, tc.bufferSize); err == nil {
				t.Fatal("invalid stream shape was accepted")
			}
		})
	}
	if engine.openCalls != 0 {
		t.Fatalf("engine stream opens = %d, want 0", engine.openCalls)
	}
}

func TestWorldOpenStreamAcceptsValidShape(t *testing.T) {
	engine := &boundaryWorldEngine{}
	world := &World{engine: engine}
	stream, err := world.OpenStream(16000, 10, 1024, 256)
	if err != nil {
		t.Fatal(err)
	}
	if stream == nil || engine.openCalls != 1 {
		t.Fatalf("stream = %v, opens = %d", stream, engine.openCalls)
	}
	stream.Close()
}

func TestValidateWorldSynthesisRejectsInvalidRate(t *testing.T) {
	features := worldFeatures{Frames: 2, FFTSize: 2, F0: []float64{200, 200},
		Spectrum: []float64{1, 1, 1, 1}, Aperiodicity: []float64{0, 0, 0, 0}}
	for _, sampleRate := range []int{0, -1, 1_000_000_000_000_000_000} {
		if err := validateWorldSynthesis(features, sampleRate); err == nil {
			t.Fatalf("sample rate %d was accepted", sampleRate)
		}
	}
	if err := validateWorldSynthesis(features, 16000); err != nil {
		t.Fatalf("valid synthesis features rejected: %v", err)
	}
	oneBin := worldFeatures{Frames: 2, FFTSize: 1, F0: []float64{200, 200},
		Spectrum: []float64{1, 1}, Aperiodicity: []float64{0, 0}}
	if err := validateWorldSynthesis(oneBin, 16000); err == nil {
		t.Fatal("FFT size 1 was accepted")
	}
}

func TestValidateWorldFeaturesRejectsOutOfRangeShape(t *testing.T) {
	features := worldFeatures{Frames: 2, FFTSize: 2, F0: []float64{200, 200},
		Spectrum: []float64{1, 1, 1, 1}, Aperiodicity: []float64{0, 0, 0, 0}}
	features.FFTSize = 1_000_000_000_000_000_000
	if err := validateWorldFeatures(features); err == nil {
		t.Fatal("out-of-range FFT size was accepted")
	}
	features.FFTSize = 2
	features.Frames = 1_000_000_000_000_000_000
	if err := validateWorldFeatures(features); err == nil {
		t.Fatal("out-of-range frame count was accepted")
	}
}
