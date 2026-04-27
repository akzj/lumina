package perf

import "time"

// Metric identifies a performance counter.
type Metric int

const (
	// Render pipeline.
	Renders  Metric = iota // renderFn calls
	Layouts                // ComputeLayout calls (includes re-layout for moves)
	Paints                 // Paint calls

	// Compositor.
	OcclusionBuilds  // OcclusionMap.Build (SetLayers) calls
	OcclusionUpdates // UpdateDirtyRegions calls
	ComposeFull      // ComposeAll calls
	ComposeDirty     // ComposeDirty calls
	ComposeRects     // ComposeRects calls (for moves)
	DirtyRectsOut    // number of dirty rects output

	// Structure.
	HitTesterRebuilds // rebuildHitTester calls
	HandlerFullSyncs  // syncHandlers (full) calls
	HandlerDirtySyncs // syncDirtyHandlers calls

	// Events.
	EventsDispatched // events that hit a handler
	EventsMissed     // events with no handler target

	// Component operations.
	ComponentsRegistered   // RegisterComponent calls
	ComponentsUnregistered // UnregisterComponent calls
	MovesPositionOnly      // Move with same size
	MovesWithResize        // Move with size change
	StateSets              // SetState calls

	// Screen output.
	WriteDirtyCalls // adapter.WriteDirty calls
	WriteFullCalls  // adapter.WriteFull calls
	FlushCalls      // adapter.Flush calls

	metricCount // sentinel for array sizing
)

// Tracker records performance metrics per frame.
// Zero value is disabled. Call Enable() to start recording.
type Tracker struct {
	enabled  bool
	current  FrameStats
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

// BeginFrame starts a new frame. Resets per-frame counters.
func (t *Tracker) BeginFrame() {
	if !t.enabled {
		return
	}
	t.current = FrameStats{
		StartTime:    time.Now(),
		EventsByType: make(map[string]int),
	}
}

// EndFrame finalizes the current frame, records to history, accumulates totals.
func (t *Tracker) EndFrame() {
	if !t.enabled {
		return
	}
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
func (t *Tracker) Record(m Metric, delta int) {
	if !t.enabled {
		return
	}
	t.current.Counters[m] += delta
}

// RecordComponent records a render for a specific component.
func (t *Tracker) RecordComponent(compID string) {
	if !t.enabled {
		return
	}
	t.current.Counters[Renders]++
	t.current.RenderComponents = append(t.current.RenderComponents, compID)
}

// RecordEvent records an event by type.
func (t *Tracker) RecordEvent(eventType string, dispatched bool) {
	if !t.enabled {
		return
	}
	if t.current.EventsByType == nil {
		t.current.EventsByType = make(map[string]int)
	}
	t.current.EventsByType[eventType]++
	if dispatched {
		t.current.Counters[EventsDispatched]++
	} else {
		t.current.Counters[EventsMissed]++
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

// TotalStats returns cumulative stats across all frames.
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
