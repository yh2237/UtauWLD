package speechtiming

type Span struct {
	Label string
	Start float64
	End   float64
}

// Timelineはフレーズ全体の音素区間と伸縮の固定点（秒）。
type Timeline struct {
	Spans  []Span
	Starts []float64
	Ends   []float64
}

// PhoneTimelineはモーラ列からフレーズの音素区間と固定点を作る。
// Mora.Spansは余白込みのフレーズ時刻（秒）としてそのまま使う。
