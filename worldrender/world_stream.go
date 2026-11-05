package worldrender

import (
	"errors"
	"fmt"
	"math"
)

// WorldStreamは、区切って足したパラメータを、区切りでもつながった波形にする（WORLDの逐次合成）。
type WorldStream struct {
	stream     worldStream
	bufferSize int
	bins       int
}

type worldStream interface {
	add(f0, spectrum, aperiodicity []float64) bool
	synthesize(output []float64) bool
	close()
}

type worldStreamer interface {
	openStream(sampleRate int, frameMS float64, fftSize, bufferSize, ringSize int) (worldStream, error)
}

type worldF0Analyzer interface {
	f0(samples []float64, sampleRate int, frameMS float64) ([]float64, error)
}

var errWorldStreamUnavailable = errors.New("WORLD streaming synthesis is unavailable on this platform")

// OpenStreamは逐次合成を始める。bufferSizeサンプルずつ取り出す。
func (w *World) OpenStream(sampleRate int, frameMS float64, fftSize, bufferSize int) (*WorldStream, error) {
	if sampleRate <= 0 || sampleRate > maxWorldInt32 || fftSize <= 0 || fftSize > maxWorldInt32 ||
		bufferSize <= 0 || bufferSize > maxWorldInt32 {
		return nil, fmt.Errorf("invalid WORLD stream shape: rate=%d fft=%d buffer=%d", sampleRate, fftSize, bufferSize)
	}
	if math.IsNaN(frameMS) || math.IsInf(frameMS, 0) || frameMS <= 0 {
		return nil, fmt.Errorf("invalid WORLD stream frame period: %v", frameMS)
	}
	streamer, ok := w.engine.(worldStreamer)
	if !ok {
		return nil, errWorldStreamUnavailable
	}
	stream, err := streamer.openStream(sampleRate, frameMS, fftSize, bufferSize, 64)
	if err != nil {
		return nil, err
	}
	return &WorldStream{stream: stream, bufferSize: bufferSize, bins: fftSize/2 + 1}, nil
}

// Addはフレームのパラメータを足す。spectrumとaperiodicityはフレームごとにfftSize/2+1個並ぶ。
// たまっていて足せなければfalseを返す。先にSynthesizeで取り出す。
func (s *WorldStream) Add(f0, spectrum, aperiodicity []float64) bool {
	if len(f0) == 0 || len(spectrum) != len(f0)*s.bins || len(aperiodicity) != len(spectrum) {
		return false
	}
	return s.stream.add(f0, spectrum, aperiodicity)
}

// Synthesizeは、合成できるだけ取り出す。
func (s *WorldStream) Synthesize() []float64 {
	var out []float64
	buffer := make([]float64, s.bufferSize)
	for s.stream.synthesize(buffer) {
		out = append(out, buffer...)
	}
	return out
}

func (s *WorldStream) Close() { s.stream.close() }

// F0はHarvestのF0だけを求める。Analyzeよりスペクトルの分だけ速い。
func (w *World) F0(samples []float64, sampleRate int, frameMS float64) ([]float64, error) {
	if sampleRate <= 0 || sampleRate > maxWorldInt32 {
		return nil, fmt.Errorf("WORLD sample rate is out of range: %d", sampleRate)
	}
	if len(samples) > maxWorldInt32 {
		return nil, fmt.Errorf("WORLD sample count is out of range: %d", len(samples))
	}
	if math.IsNaN(frameMS) || math.IsInf(frameMS, 0) || frameMS <= 0 {
		return nil, fmt.Errorf("WORLD frame period must be positive and finite: %v", frameMS)
	}
	analyzer, ok := w.engine.(worldF0Analyzer)
	if !ok {
		f0, _, _, _, err := w.Analyze(samples, sampleRate, frameMS, 0)
		return f0, err
	}
	return analyzer.f0(samples, sampleRate, frameMS)
}
