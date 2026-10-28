package speechtiming

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
)

func TestDefaultTargetMatchesPyTorch(t *testing.T) {
	data, err := os.ReadFile("testdata/target-v1-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		IDs    [][3]int     `json:"ids"`
		Cont   [][4]float32 `json:"cont"`
		Output [][]float32  `json:"output"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	model, err := DefaultTarget()
	if err != nil {
		t.Fatal(err)
	}
	got, err := model.Predict(fixture.IDs, fixture.Cont)
	if err != nil {
		t.Fatal(err)
	}
	worst := 0.0
	for frame := range fixture.Output {
		for channel, want := range fixture.Output[frame] {
			worst = math.Max(worst, math.Abs(float64(got[frame][channel]-want)))
		}
	}
	if worst > 1e-3 {
		t.Fatalf("max abs difference from PyTorch = %g", worst)
	}
}

func TestPointwiseKeepsBiasOrderAndRemainingOutput(t *testing.T) {
	got := pointwise([]float32{-1e8, 1e8, 1}, 1, 3,
		[]float32{1, 1, 1, 1, 0, 1, 0, 1, 1, 0, 0, 2, 1, 1, 1},
		[]float32{1, 1, 1, 1, 1}, 5)
	want := []float32{1, -1e8, 1e8, 3, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pointwise = %v, want %v", got, want)
	}
}




func TestWarpKeepsIdenticalTargetAndProtectsEnds(t *testing.T) {
	frames := 60
	rows := make([][]float64, frames)
	for frame := range rows {
		rows[frame] = []float64{math.Sin(float64(frame) * 0.3), math.Cos(float64(frame) * 0.17)}
	}
	mapping := warpMap(rows, rows, []float64{20, 40}, frames)
	for frame, value := range mapping {
		if math.Abs(value-float64(frame)) > 1e-9 {
			t.Fatalf("identity mapping[%d] = %v", frame, value)
		}
	}
	weight := protectPhraseEnds(100, []float64{0.1, 0.3, 0.5}, []float64{0.62})
	if weight[20] != 1 || weight[38] != 0 || weight[64] != 0 || weight[65] != 1 {
		t.Fatalf("weights = %v", weight)
	}
	if weight[35] <= 0 || weight[35] >= 1 {
		t.Fatalf("ramp weight = %v", weight[35])
	}
}

