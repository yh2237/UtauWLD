//go:build js

package worldrender

import (
	"syscall/js"
	"testing"
)

func TestWasmSynthesizeValidatesShapeBeforeJS(t *testing.T) {
	called := false
	api := js.Global().Get("Object").New()
	synthesize := js.FuncOf(func(this js.Value, args []js.Value) any {
		called = true
		return js.ValueOf(map[string]any{"ok": true, "samples": js.Global().Get("Float64Array").New(0)})
	})
	defer synthesize.Release()
	api.Set("synthesize", synthesize)
	js.Global().Set("utauttsWorld", api)
	defer js.Global().Set("utauttsWorld", js.Undefined())

	engine := &wasmWorldEngine{}
	if _, err := engine.Synthesize(worldFeatures{Frames: 2, FFTSize: 0}, 16000); err == nil {
		t.Fatal("invalid feature shape was accepted")
	}
	if called {
		t.Fatal("JS synthesize was called with an invalid shape")
	}
}

func TestWasmSynthesizeCallsJSWithValidShape(t *testing.T) {
	called := false
	api := js.Global().Get("Object").New()
	synthesize := js.FuncOf(func(this js.Value, args []js.Value) any {
		called = true
		if len(args) != 7 {
			t.Errorf("JS synthesize arguments = %d, want 7", len(args))
		}
		result := js.Global().Get("Object").New()
		result.Set("ok", true)
		result.Set("samples", js.Global().Get("Float64Array").New(3))
		return result
	})
	defer synthesize.Release()
	api.Set("synthesize", synthesize)
	js.Global().Set("utauttsWorld", api)
	defer js.Global().Set("utauttsWorld", js.Undefined())

	engine := &wasmWorldEngine{}
	features := worldFeatures{Frames: 2, FFTSize: 2, F0: []float64{200, 200},
		Spectrum: []float64{1, 1, 1, 1}, Aperiodicity: []float64{0, 0, 0, 0}}
	samples, err := engine.Synthesize(features, 16000)
	if err != nil {
		t.Fatal(err)
	}
	if !called || len(samples) != 3 {
		t.Fatalf("called=%v samples=%d, want true/3", called, len(samples))
	}
}
