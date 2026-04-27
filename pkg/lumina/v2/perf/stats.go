package perf

import "time"

// FrameStats holds metrics for a single frame (or cumulative totals).
type FrameStats struct {
	StartTime        time.Time
	Duration         time.Duration
	Counters         [metricCount]int
	EventsByType     map[string]int
	RenderComponents []string

	// Cumulative fields (only meaningful in TotalStats).
	Frames           int
	TotalDuration    time.Duration
	MaxFrameDuration time.Duration
}

// Get returns the value of a counter.
func (s FrameStats) Get(m Metric) int { return s.Counters[m] }
