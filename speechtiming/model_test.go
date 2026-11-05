package speechtiming

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"
)

func safetensorsWithShape(t *testing.T, shape []int) []byte {
	t.Helper()
	header, err := json.Marshal(map[string]any{
		"x": map[string]any{"dtype": "F32", "shape": shape, "data_offsets": []int{0, 0}},
	})
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 8+len(header))
	binary.LittleEndian.PutUint64(data[:8], uint64(len(header)))
	copy(data[8:], header)
	return data
}

func TestLoadTCNRejectsInvalidTensorShapes(t *testing.T) {
	for name, shape := range map[string][]int{
		"negative dimension": {-math.MaxInt/2 - 1},
		"zero dimension":     {0},
		"overflowing size":   {math.MaxInt/2 + 1},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadTCN(safetensorsWithShape(t, shape)); err == nil {
				t.Fatal("invalid tensor shape was accepted")
			}
		})
	}
}
