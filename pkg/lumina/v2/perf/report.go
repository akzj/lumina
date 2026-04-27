package perf

import (
	"fmt"
	"strings"
)

// Report generates a human-readable report of the last completed frame.
func (t *Tracker) Report() string {
	f := t.LastFrame()
	var b strings.Builder
	fmt.Fprintf(&b, "=== Frame Report (%.2fms) ===\n", float64(f.Duration.Microseconds())/1000)
	fmt.Fprintf(&b, "Renders: %d", f.Get(Renders))
	if len(f.RenderComponents) > 0 {
		fmt.Fprintf(&b, " [%s]", strings.Join(f.RenderComponents, ", "))
	}
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Layouts: %d, Paints: %d\n", f.Get(Layouts), f.Get(Paints))
	fmt.Fprintf(&b, "Occlusion: builds=%d updates=%d\n", f.Get(OcclusionBuilds), f.Get(OcclusionUpdates))
	fmt.Fprintf(&b, "Compose: full=%d dirty=%d rects=%d\n", f.Get(ComposeFull), f.Get(ComposeDirty), f.Get(ComposeRects))
	fmt.Fprintf(&b, "HitTester rebuilds: %d\n", f.Get(HitTesterRebuilds))
	fmt.Fprintf(&b, "Handler syncs: full=%d dirty=%d\n", f.Get(HandlerFullSyncs), f.Get(HandlerDirtySyncs))
	fmt.Fprintf(&b, "Events: dispatched=%d missed=%d\n", f.Get(EventsDispatched), f.Get(EventsMissed))
	if len(f.EventsByType) > 0 {
		fmt.Fprintf(&b, "  by type: ")
		for k, v := range f.EventsByType {
			fmt.Fprintf(&b, "%s=%d ", k, v)
		}
		fmt.Fprintln(&b)
	}
	fmt.Fprintf(&b, "Components: reg=%d unreg=%d moves(pos)=%d moves(resize)=%d setState=%d\n",
		f.Get(ComponentsRegistered), f.Get(ComponentsUnregistered),
		f.Get(MovesPositionOnly), f.Get(MovesWithResize), f.Get(StateSets))
	fmt.Fprintf(&b, "Output: dirtyRects=%d writeDirty=%d writeFull=%d flush=%d\n",
		f.Get(DirtyRectsOut), f.Get(WriteDirtyCalls), f.Get(WriteFullCalls), f.Get(FlushCalls))
	return b.String()
}

// TotalReport generates a cumulative report across all frames.
func (t *Tracker) TotalReport() string {
	s := t.TotalStats()
	var b strings.Builder
	fmt.Fprintf(&b, "=== Total Report (%d frames, %.2fms total, %.2fms max) ===\n",
		s.Frames,
		float64(s.TotalDuration.Microseconds())/1000,
		float64(s.MaxFrameDuration.Microseconds())/1000)
	fmt.Fprintf(&b, "Renders: %d, Layouts: %d, Paints: %d\n",
		s.Get(Renders), s.Get(Layouts), s.Get(Paints))
	fmt.Fprintf(&b, "Occlusion: builds=%d updates=%d\n",
		s.Get(OcclusionBuilds), s.Get(OcclusionUpdates))
	fmt.Fprintf(&b, "Compose: full=%d dirty=%d rects=%d\n",
		s.Get(ComposeFull), s.Get(ComposeDirty), s.Get(ComposeRects))
	fmt.Fprintf(&b, "HitTester rebuilds: %d\n", s.Get(HitTesterRebuilds))
	fmt.Fprintf(&b, "Handler syncs: full=%d dirty=%d\n",
		s.Get(HandlerFullSyncs), s.Get(HandlerDirtySyncs))
	fmt.Fprintf(&b, "Events: dispatched=%d missed=%d\n",
		s.Get(EventsDispatched), s.Get(EventsMissed))
	fmt.Fprintf(&b, "Components: reg=%d unreg=%d moves(pos)=%d moves(resize)=%d setState=%d\n",
		s.Get(ComponentsRegistered), s.Get(ComponentsUnregistered),
		s.Get(MovesPositionOnly), s.Get(MovesWithResize), s.Get(StateSets))
	fmt.Fprintf(&b, "Output: dirtyRects=%d writeDirty=%d writeFull=%d flush=%d\n",
		s.Get(DirtyRectsOut), s.Get(WriteDirtyCalls), s.Get(WriteFullCalls), s.Get(FlushCalls))
	return b.String()
}
