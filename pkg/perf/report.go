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
	fmt.Fprintf(&b, "Output: dirtyRects=%d writeDirty=%d writeFull=%d flush=%d\n",
		f.Get(DirtyRectsOut), f.Get(WriteDirtyCalls), f.Get(WriteFullCalls), f.Get(FlushCalls))
	if f.Get(ComponentsRendered) > 0 || f.Get(PaintCells) > 0 {
		fmt.Fprintf(&b, "Render: renderCount=%d paintCells=%d clearCells=%d dirtyArea=%d\n",
			f.Get(ComponentsRendered), f.Get(PaintCells), f.Get(PaintClearCells), f.Get(DirtyRectArea))
	}
	if len(f.RenderComponents) > 0 {
		fmt.Fprintf(&b, "RecordComponent: [%s]\n", strings.Join(f.RenderComponents, ", "))
	}
	if len(f.EventsByType) > 0 {
		fmt.Fprint(&b, "EventsByType: ")
		for k, v := range f.EventsByType {
			fmt.Fprintf(&b, "%s=%d ", k, v)
		}
		fmt.Fprintln(&b)
	}
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
	fmt.Fprintf(&b, "Output: dirtyRects=%d writeDirty=%d writeFull=%d flush=%d\n",
		s.Get(DirtyRectsOut), s.Get(WriteDirtyCalls), s.Get(WriteFullCalls), s.Get(FlushCalls))
	if s.Get(ComponentsRendered) > 0 || s.Get(PaintCells) > 0 {
		fmt.Fprintf(&b, "Render: renderCount=%d paintCells=%d clearCells=%d dirtyArea=%d\n",
			s.Get(ComponentsRendered), s.Get(PaintCells), s.Get(PaintClearCells), s.Get(DirtyRectArea))
	}
	return b.String()
}
