package speechtiming

import (
	"math"
	"reflect"
	"testing"
)

func warpTestFeatures() Features {
	const frames, fftSize = 40, 64
	bins := fftSize/2 + 1
	features := Features{Frames: frames, FFTSize: fftSize, SampleRate: 16000,
		F0: make([]float64, frames), Spectrum: make([]float64, frames*bins), Aperiodicity: make([]float64, frames*bins)}
	for frame := range features.F0 {
		features.F0[frame] = 200
	}
	for index := range features.Spectrum {
		features.Spectrum[index] = 1
		features.Aperiodicity[index] = .1
	}
	return features
}

func warpTestTimeline() Timeline {
	return Timeline{
		Spans:  []Span{{Label: "a", Start: 0, End: .2}, {Label: "i", Start: .2, End: .4}},
		Starts: []float64{0, .2},
		Ends:   []float64{.4},
	}
}

func TestWarpRejectsNonFiniteStrength(t *testing.T) {
	model, err := DefaultTarget()
	if err != nil {
		t.Fatal(err)
	}
	for _, strength := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := Warp(model, warpTestTimeline(), warpTestFeatures(), strength); err == nil {
			t.Fatalf("strength %v was accepted", strength)
		}
	}
}

func TestWarpZeroStrengthReturnsInput(t *testing.T) {
	model, err := DefaultTarget()
	if err != nil {
		t.Fatal(err)
	}
	input := warpTestFeatures()
	output, err := Warp(model, warpTestTimeline(), input, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(output, input) {
		t.Fatal("zero strength changed features")
	}
}
