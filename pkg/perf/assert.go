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

// CheckRenders asserts the Renders counter equals expected.
func CheckRenders(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(Renders)
		if got != expected {
			return false, fmt.Sprintf("Renders: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckLayouts asserts the Layouts counter equals expected.
func CheckLayouts(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(Layouts)
		if got != expected {
			return false, fmt.Sprintf("Layouts: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckPaints asserts the Paints counter equals expected.
func CheckPaints(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(Paints)
		if got != expected {
			return false, fmt.Sprintf("Paints: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckOcclusionBuilds asserts the OcclusionBuilds counter equals expected.
func CheckOcclusionBuilds(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(OcclusionBuilds)
		if got != expected {
			return false, fmt.Sprintf("OcclusionBuilds: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckOcclusionUpdates asserts the OcclusionUpdates counter equals expected.
func CheckOcclusionUpdates(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(OcclusionUpdates)
		if got != expected {
			return false, fmt.Sprintf("OcclusionUpdates: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckHitTesterRebuilds asserts the HitTesterRebuilds counter equals expected.
func CheckHitTesterRebuilds(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(HitTesterRebuilds)
		if got != expected {
			return false, fmt.Sprintf("HitTesterRebuilds: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckEventsDispatched asserts the EventsDispatched counter equals expected.
func CheckEventsDispatched(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(EventsDispatched)
		if got != expected {
			return false, fmt.Sprintf("EventsDispatched: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckEventsMissed asserts the EventsMissed counter equals expected.
func CheckEventsMissed(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(EventsMissed)
		if got != expected {
			return false, fmt.Sprintf("EventsMissed: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckRenderComponents asserts that exactly the given component IDs were rendered
// (order-independent).
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

// CheckHandlerFullSyncs asserts the HandlerFullSyncs counter equals expected.
func CheckHandlerFullSyncs(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(HandlerFullSyncs)
		if got != expected {
			return false, fmt.Sprintf("HandlerFullSyncs: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckHandlerDirtySyncs asserts the HandlerDirtySyncs counter equals expected.
func CheckHandlerDirtySyncs(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(HandlerDirtySyncs)
		if got != expected {
			return false, fmt.Sprintf("HandlerDirtySyncs: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// --- V2 Render Engine checks ---

// CheckV2ComponentsRendered asserts the V2ComponentsRendered counter equals expected.
func CheckV2ComponentsRendered(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(V2ComponentsRendered)
		if got != expected {
			return false, fmt.Sprintf("V2ComponentsRendered: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckV2PaintCells asserts the V2PaintCells counter equals expected.
func CheckV2PaintCells(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(V2PaintCells)
		if got != expected {
			return false, fmt.Sprintf("V2PaintCells: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckV2PaintCellsMax asserts the V2PaintCells counter is at most max.
func CheckV2PaintCellsMax(max int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(V2PaintCells)
		if got > max {
			return false, fmt.Sprintf("V2PaintCells: got %d, want ≤%d (overdraw!)", got, max)
		}
		return true, ""
	}
}

// CheckV2PaintClearCells asserts the V2PaintClearCells counter equals expected.
func CheckV2PaintClearCells(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(V2PaintClearCells)
		if got != expected {
			return false, fmt.Sprintf("V2PaintClearCells: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckV2DirtyRectArea asserts the V2DirtyRectArea counter equals expected.
func CheckV2DirtyRectArea(expected int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(V2DirtyRectArea)
		if got != expected {
			return false, fmt.Sprintf("V2DirtyRectArea: got %d, want %d", got, expected)
		}
		return true, ""
	}
}

// CheckV2DirtyRectAreaMax asserts the V2DirtyRectArea counter is at most max.
func CheckV2DirtyRectAreaMax(max int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(V2DirtyRectArea)
		if got > max {
			return false, fmt.Sprintf("V2DirtyRectArea: got %d, want ≤%d (overdraw!)", got, max)
		}
		return true, ""
	}
}

// CheckV2ComponentsRenderedMax asserts the V2ComponentsRendered counter is at most max.
func CheckV2ComponentsRenderedMax(max int) FrameCheck {
	return func(f FrameStats) (bool, string) {
		got := f.Get(V2ComponentsRendered)
		if got > max {
			return false, fmt.Sprintf("V2ComponentsRendered: got %d, want ≤%d", got, max)
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
