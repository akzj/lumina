package perf

import "time"

// Metric identifies a performance counter (render engine + adapter output only).
type Metric int

const (
	DirtyRectsOut Metric = iota // number of dirty rects passed to WriteDirty this frame
	WriteDirtyCalls
	WriteFullCalls
	FlushCalls

	// Render engine (from engine.RenderAll / RenderDirty after paint).
	ComponentsRendered // render count: renderComponent calls this frame
	PaintCells         // CellBuffer cell writes (SetChar / Set, etc.)
	PaintClearCells    // cells cleared (ClearRect / Clear) before repaint
	DirtyRectArea      // paint dirty bbox area: Stats.DirtyW * DirtyH

	metricCount // sentinel for array sizing
)

// Tracker records performance metrics per frame.
// Zero value is disabled. Call Enable() to start recording.
type Tracker struct {
	enabled  bool
	current  FrameStats
	pending  FrameStats // accumulates inter-frame records (between EndFrame and BeginFrame)
	inFrame  bool       // true between BeginFrame/EndFrame
	history  []FrameStats
	histPos  int
	histSize int
	total    FrameStats
	alertFn  func(FrameStats)
}

// NewTracker creates a new tracker with a ring buffer of historySize frames.
// Pass 0 for default (60 frames).
func NewTracker(historySize int) *Tracker {
	if historySize <= 0 {
		historySize = 60
	}
	return &Tracker{
		history:  make([]FrameStats, historySize),
		histSize: historySize,
	}
}

// Enable starts recording metrics.
func (t *Tracker) Enable() { t.enabled = true }

// Disable stops recording metrics.
func (t *Tracker) Disable() { t.enabled = false }

// Enabled returns whether the tracker is recording.
func (t *Tracker) Enabled() bool { return t.enabled }

// BeginFrame starts a new frame. Merges any inter-frame records (from
// SetState, MoveComponent, etc. called between EndFrame and BeginFrame)
// into the new frame, then resets per-frame counters.
func (t *Tracker) BeginFrame() {
	if !t.enabled {
		return
	}
	// Start with pending inter-frame counters.
	t.current = FrameStats{
		StartTime:        time.Now(),
		EventsByType:     make(map[string]int),
		RenderComponents: append([]string(nil), t.pending.RenderComponents...),
	}
	// Merge pending counters.
	for i := 0; i < int(metricCount); i++ {
		t.current.Counters[i] = t.pending.Counters[i]
	}
	// Merge pending events.
	for k, v := range t.pending.EventsByType {
		t.current.EventsByType[k] += v
	}
	// Reset pending.
	t.pending = FrameStats{}
	t.inFrame = true
}

// EndFrame finalizes the current frame, records to history, accumulates totals.
func (t *Tracker) EndFrame() {
	if !t.enabled {
		return
	}
	t.inFrame = false
	t.current.Duration = time.Since(t.current.StartTime)

	// Save to history ring buffer.
	t.history[t.histPos] = t.current
	t.histPos = (t.histPos + 1) % t.histSize

	// Accumulate to total.
	for i := 0; i < int(metricCount); i++ {
		t.total.Counters[i] += t.current.Counters[i]
	}
	t.total.Frames++
	t.total.TotalDuration += t.current.Duration
	if t.current.Duration > t.total.MaxFrameDuration {
		t.total.MaxFrameDuration = t.current.Duration
	}

	// Alert callback.
	if t.alertFn != nil {
		t.alertFn(t.current)
	}
}

// Record increments a counter by delta.
// If called between frames (outside BeginFrame/EndFrame), the record is
// accumulated into a pending buffer and merged into the next BeginFrame.
func (t *Tracker) Record(m Metric, delta int) {
	if !t.enabled {
		return
	}
	if t.inFrame {
		t.current.Counters[m] += delta
	} else {
		t.pending.Counters[m] += delta
	}
}

// RecordComponent records a test/instrumentation hook: increments ComponentsRendered
// and appends compID to RenderComponents (same slice shape as before; not used by the engine).
func (t *Tracker) RecordComponent(compID string) {
	if !t.enabled {
		return
	}
	if t.inFrame {
		t.current.Counters[ComponentsRendered]++
		t.current.RenderComponents = append(t.current.RenderComponents, compID)
	} else {
		t.pending.Counters[ComponentsRendered]++
		t.pending.RenderComponents = append(t.pending.RenderComponents, compID)
	}
}

// RecordEvent records an event type for debugging (EventsByType map only; no separate counters).
func (t *Tracker) RecordEvent(eventType string, dispatched bool) {
	if !t.enabled {
		return
	}
	_ = dispatched
	if t.inFrame {
		if t.current.EventsByType == nil {
			t.current.EventsByType = make(map[string]int)
		}
		t.current.EventsByType[eventType]++
	} else {
		if t.pending.EventsByType == nil {
			t.pending.EventsByType = make(map[string]int)
		}
		t.pending.EventsByType[eventType]++
	}
}

// SetAlert sets a callback invoked after each EndFrame.
func (t *Tracker) SetAlert(fn func(FrameStats)) {
	t.alertFn = fn
}

// --- Query API ---

// LastFrame returns the most recent completed frame's stats.
func (t *Tracker) LastFrame() FrameStats {
	idx := (t.histPos - 1 + t.histSize) % t.histSize
	return t.history[idx]
}

// CurrentFrame returns the in-progress frame (between BeginFrame/EndFrame).
func (t *Tracker) CurrentFrame() FrameStats {
	return t.current
}

// TotalStats returns cumulative stats across all frames since Enable (or last Reset).
func (t *Tracker) TotalStats() FrameStats {
	return t.total
}

// History returns recorded frames oldest-first.
func (t *Tracker) History() []FrameStats {
	result := make([]FrameStats, 0, t.histSize)
	for i := 0; i < t.histSize; i++ {
		idx := (t.histPos + i) % t.histSize
		if t.history[idx].StartTime.IsZero() {
			continue // unused slot
		}
		result = append(result, t.history[idx])
	}
	return result
}

// Reset clears all stats.
func (t *Tracker) Reset() {
	t.current = FrameStats{}
	t.total = FrameStats{}
	t.histPos = 0
	for i := range t.history {
		t.history[i] = FrameStats{}
	}
}
