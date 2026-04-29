package perf

import (
	"fmt"
	"sort"
	"testing"
)

// FrameCheck is a predicate on FrameStats.
type FrameCheck func(FrameStats) (ok bool, msg string)

// CheckMetric returns a check that asserts a specific metric equals expected.
func CheckMetric(m Metric, expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(m)
		if got != expected {
			return false, fmt.Sprintf("metric %d: got %d, want %d", m, got, expected)
		}
		return true, ""
	}
}

// CheckRenderComponents asserts that exactly the given component IDs were recorded
// via RecordComponent (order-independent).
func CheckRenderComponents(expected ...string) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := make([]string, len(f.RenderComponents))
		copy(got, f.RenderComponents)
		want := make([]string, len(expected))
		copy(want, expected)
		sort.Strings(got)
		sort.Strings(want)
		if len(got) != len(want) {
			return false, fmt.Sprintf("RenderComponents: got %v, want %v", got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				return false, fmt.Sprintf("RenderComponents: got %v, want %v", got, want)
			}
		}
		return true, ""
	}
}

// CheckComponentsRendered asserts the render-count counter (ComponentsRendered) equals expected.
func CheckComponentsRendered(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(ComponentsRendered)
		if got != expected {
			return false, fmt.Sprintf("render count: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckPaintCells asserts the PaintCells counter equals expected.
func CheckPaintCells(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(PaintCells)
		if got != expected {
			return false, fmt.Sprintf("PaintCells: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckPaintCellsMax asserts the PaintCells counter is at most max.
func CheckPaintCellsMax(max int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(PaintCells)
		if got > max {
			return false, fmt.Sprintf("PaintCells: got %d, want ≤%d (overdraw!)", got, max)
		}
		return true, ""
	}
}

// CheckPaintClearCells asserts the PaintClearCells counter equals expected.
func CheckPaintClearCells(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(PaintClearCells)
		if got != expected {
			return false, fmt.Sprintf("PaintClearCells: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckDirtyRectArea asserts the DirtyRectArea counter equals expected.
func CheckDirtyRectArea(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(DirtyRectArea)
		if got != expected {
			return false, fmt.Sprintf("DirtyRectArea: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckDirtyRectAreaMax asserts the DirtyRectArea counter is at most max.
func CheckDirtyRectAreaMax(max int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(DirtyRectArea)
		if got > max {
			return false, fmt.Sprintf("DirtyRectArea: got %d, want ≤%d (overdraw!)", got, max)
		}
		return true, ""
	}
}

// CheckComponentsRenderedMax asserts the render-count counter is at most max.
func CheckComponentsRenderedMax(max int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(ComponentsRendered)
		if got > max {
			return false, fmt.Sprintf("render count: got %d, want ≤%d", got, max)
		}
		return true, ""
	}
}

// AssertLastFrame asserts all checks against the last completed frame.
func (t *Tracker) AssertLastFrame(tb testing.TB, checks ...FrameCheck) {
	tb.Helper()
	f := t.LastFrame()
	for _, check := range checks {
		if ok, msg := check(f); !ok {
			tb.Errorf("perf assertion failed: %s", msg)
		}
	}
}
