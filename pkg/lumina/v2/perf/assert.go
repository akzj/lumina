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
