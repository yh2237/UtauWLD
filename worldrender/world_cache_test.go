package worldrender

import (
	"context"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type recordingWorldEngine struct {
	analyzeCalls int
	inputF0Lens  []int
}

func (e *recordingWorldEngine) Close() error { return nil }

func (e *recordingWorldEngine) Analyze(samples []float64, sampleRate int, inputF0 []float64) (worldFeatures, error) {
	e.analyzeCalls++
	e.inputF0Lens = append(e.inputF0Lens, len(inputF0))
	hopSize := int(math.Round(worldFramePeriodMS * float64(sampleRate) / 1000))
	frames := len(samples)/hopSize + 1
	const fftSize = 16
	bins := fftSize/2 + 1
	features := worldFeatures{Frames: frames, FFTSize: fftSize, F0: make([]float64, frames),
		Spectrum: make([]float64, frames*bins), Aperiodicity: make([]float64, frames*bins)}
	for frame := range features.F0 {
		features.F0[frame] = 200
	}
	for index := range features.Spectrum {
		features.Spectrum[index] = 1
		features.Aperiodicity[index] = .1
	}
	return features, nil
}

func (e *recordingWorldEngine) Synthesize(features worldFeatures, sampleRate int) ([]float64, error) {
	return nil, nil
}

func writeTestFRQ(t *testing.T, path string, hopSize, length int, f0 float64) {
	t.Helper()
	data := make([]byte, 40+length*16)
	copy(data, "FREQ0003")
	binary.LittleEndian.PutUint32(data[8:12], uint32(hopSize))
	binary.LittleEndian.PutUint32(data[36:40], uint32(length))
	for index := 0; index < length; index++ {
		binary.LittleEndian.PutUint64(data[40+index*16:48+index*16], math.Float64bits(f0))
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareWorldUnitsCacheKeyIncludesFRQPath(t *testing.T) {
	const sampleRate = 16000
	dir := t.TempDir()
	source := filepath.Join(dir, "unit.wav")
	samples := make([]float32, sampleRate/5)
	for index := range samples {
		samples[index] = .1 * float32(math.Sin(2*math.Pi*200*float64(index)/float64(sampleRate)))
	}
	if err := writePCM16(source, sampleRate, samples); err != nil {
		t.Fatal(err)
	}
	frqPath := filepath.Join(dir, "unit.frq")
	writeTestFRQ(t, frqPath, 160, 21, 200)
	engine := &recordingWorldEngine{}
	input := phraseInput{SampleRate: sampleRate, Units: []Unit{
		{Source: source, OffsetMS: 0, CutoffMS: 0, LengthMS: 200, RequiredLengthMS: 200, FRQPath: frqPath},
		{Source: source, OffsetMS: 0, CutoffMS: 0, LengthMS: 200, RequiredLengthMS: 200},
	}}
	prepared, err := prepareWorldUnits(context.Background(), engine, input, newWorldFeatureCache(4), 1)
	if err != nil {
		t.Fatal(err)
	}
	if engine.analyzeCalls != 2 {
		t.Fatalf("analyze calls = %d, want 2 (FRQPath must be part of the fallback cache key)", engine.analyzeCalls)
	}
	if len(engine.inputF0Lens) != 2 || engine.inputF0Lens[0] == 0 || engine.inputF0Lens[1] != 0 {
		t.Fatalf("input F0 lengths = %v, want [non-zero 0]", engine.inputF0Lens)
	}
	if len(prepared) != 2 {
		t.Fatalf("prepared units = %d, want 2", len(prepared))
	}
}
