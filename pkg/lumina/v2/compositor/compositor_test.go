package compositor

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
)

// helper: create a buffer filled with a single character.
func filledBuffer(w, h int, ch rune, fg string) *buffer.Buffer {
	b := buffer.New(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b.Set(x, y, buffer.Cell{Char: ch, Foreground: fg})
		}
	}
	return b
}

// helper: create a buffer with specific cells set (rest are zero/transparent).
func sparseBuffer(w, h int, cells map[[2]int]buffer.Cell) *buffer.Buffer {
	b := buffer.New(w, h)
	for pos, cell := range cells {
		b.Set(pos[0], pos[1], cell)
	}
	return b
}

func TestOcclusion_SingleLayer(t *testing.T) {
	buf := filledBuffer(10, 5, 'A', "#ff0000")
	layer := &Layer{
		ID:     "a",
		Buffer: buf,
		Rect:   buffer.Rect{X: 0, Y: 0, W: 10, H: 5},
		ZIndex: 0,
	}

	om := NewOcclusionMap(10, 5)
	om.Build([]*Layer{layer})

	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			if got := om.Owner(x, y); got != "a" {
				t.Fatalf("(%d,%d): expected owner 'a', got %q", x, y, got)
			}
		}
	}
}

func TestOcclusion_TwoLayers_Overlap(t *testing.T) {
	// A covers entire 10x10 at z=0
	bufA := filledBuffer(10, 10, 'A', "#ff0000")
	layerA := &Layer{
		ID: "a", Buffer: bufA,
		Rect: buffer.Rect{X: 0, Y: 0, W: 10, H: 10}, ZIndex: 0,
	}

	// B covers 3x3 at (2,2) at z=100
	bufB := filledBuffer(3, 3, 'B', "#00ff00")
	layerB := &Layer{
		ID: "b", Buffer: bufB,
		Rect: buffer.Rect{X: 2, Y: 2, W: 3, H: 3}, ZIndex: 100,
	}

	om := NewOcclusionMap(10, 10)
	om.Build([]*Layer{layerA, layerB})

	// B should own the overlap area.
	for y := 2; y < 5; y++ {
		for x := 2; x < 5; x++ {
			if got := om.Owner(x, y); got != "b" {
				t.Errorf("(%d,%d): expected 'b', got %q", x, y, got)
			}
		}
	}
	// A should own cells outside B's rect.
	if got := om.Owner(0, 0); got != "a" {
		t.Errorf("(0,0): expected 'a', got %q", got)
	}
	if got := om.Owner(9, 9); got != "a" {
		t.Errorf("(9,9): expected 'a', got %q", got)
	}
}

func TestOcclusion_ThreeLayers(t *testing.T) {
	bufA := filledBuffer(10, 10, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 10, H: 10}, ZIndex: 0}

	bufB := filledBuffer(5, 5, 'B', "#00ff00")
	layerB := &Layer{ID: "b", Buffer: bufB, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 100}

	bufC := filledBuffer(2, 2, 'C', "#0000ff")
	layerC := &Layer{ID: "c", Buffer: bufC, Rect: buffer.Rect{X: 1, Y: 1, W: 2, H: 2}, ZIndex: 200}

	om := NewOcclusionMap(10, 10)
	om.Build([]*Layer{layerA, layerB, layerC})

	// C owns (1,1)-(2,2)
	if got := om.Owner(1, 1); got != "c" {
		t.Errorf("(1,1): expected 'c', got %q", got)
	}
	if got := om.Owner(2, 2); got != "c" {
		t.Errorf("(2,2): expected 'c', got %q", got)
	}
	// B owns (0,0) and (4,4) but not where C is.
	if got := om.Owner(0, 0); got != "b" {
		t.Errorf("(0,0): expected 'b', got %q", got)
	}
	if got := om.Owner(4, 4); got != "b" {
		t.Errorf("(4,4): expected 'b', got %q", got)
	}
	// A owns (9,9) — outside B and C.
	if got := om.Owner(9, 9); got != "a" {
		t.Errorf("(9,9): expected 'a', got %q", got)
	}
}

func TestOcclusion_NoOverlap(t *testing.T) {
	bufA := filledBuffer(5, 5, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 0}

	bufB := filledBuffer(5, 5, 'B', "#00ff00")
	layerB := &Layer{ID: "b", Buffer: bufB, Rect: buffer.Rect{X: 5, Y: 0, W: 5, H: 5}, ZIndex: 0}

	om := NewOcclusionMap(10, 5)
	om.Build([]*Layer{layerA, layerB})

	if got := om.Owner(0, 0); got != "a" {
		t.Errorf("(0,0): expected 'a', got %q", got)
	}
	if got := om.Owner(4, 4); got != "a" {
		t.Errorf("(4,4): expected 'a', got %q", got)
	}
	if got := om.Owner(5, 0); got != "b" {
		t.Errorf("(5,0): expected 'b', got %q", got)
	}
	if got := om.Owner(9, 4); got != "b" {
		t.Errorf("(9,4): expected 'b', got %q", got)
	}
}

func TestCompositor_ComposeAll(t *testing.T) {
	bufA := filledBuffer(10, 10, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 10, H: 10}, ZIndex: 0}

	bufB := filledBuffer(3, 3, 'B', "#00ff00")
	layerB := &Layer{ID: "b", Buffer: bufB, Rect: buffer.Rect{X: 2, Y: 2, W: 3, H: 3}, ZIndex: 100}

	comp := NewCompositor(10, 10)
	comp.SetLayers([]*Layer{layerA, layerB})
	screen := comp.ComposeAll()

	// (0,0) should be 'A'
	if c := screen.Get(0, 0); c.Char != 'A' {
		t.Errorf("(0,0): expected 'A', got %q", c.Char)
	}
	// (2,2) should be 'B' (higher z)
	if c := screen.Get(2, 2); c.Char != 'B' {
		t.Errorf("(2,2): expected 'B', got %q", c.Char)
	}
	// (9,9) should be 'A'
	if c := screen.Get(9, 9); c.Char != 'A' {
		t.Errorf("(9,9): expected 'A', got %q", c.Char)
	}
}

func TestCompositor_ComposeDirty_SingleCell(t *testing.T) {
	bufA := filledBuffer(5, 5, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 0}

	comp := NewCompositor(5, 5)
	comp.SetLayers([]*Layer{layerA})
	comp.ComposeAll()

	// Now change cell (1,1) in the buffer.
	bufA.Set(1, 1, buffer.Cell{Char: 'X', Foreground: "#ffffff"})
	dr := buffer.Rect{X: 1, Y: 1, W: 1, H: 1}
	layerA.DirtyRect = &dr

	rects := comp.ComposeDirty([]*Layer{layerA})
	if len(rects) != 1 {
		t.Fatalf("expected 1 dirty rect, got %d", len(rects))
	}

	// Screen should now have 'X' at (1,1).
	if c := comp.Screen().Get(1, 1); c.Char != 'X' {
		t.Errorf("(1,1): expected 'X', got %q", c.Char)
	}
	// (0,0) should still be 'A'.
	if c := comp.Screen().Get(0, 0); c.Char != 'A' {
		t.Errorf("(0,0): expected 'A', got %q", c.Char)
	}
}

func TestCompositor_ComposeDirty_SubRect(t *testing.T) {
	bufA := filledBuffer(10, 10, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 10, H: 10}, ZIndex: 0}

	comp := NewCompositor(10, 10)
	comp.SetLayers([]*Layer{layerA})
	comp.ComposeAll()

	// Change a 3x3 sub-region.
	for y := 2; y < 5; y++ {
		for x := 2; x < 5; x++ {
			bufA.Set(x, y, buffer.Cell{Char: 'Z', Foreground: "#ffffff"})
		}
	}
	dr := buffer.Rect{X: 2, Y: 2, W: 3, H: 3}
	layerA.DirtyRect = &dr

	rects := comp.ComposeDirty([]*Layer{layerA})
	if len(rects) != 1 {
		t.Fatalf("expected 1 dirty rect, got %d", len(rects))
	}
	if rects[0] != (buffer.Rect{X: 2, Y: 2, W: 3, H: 3}) {
		t.Errorf("dirty rect: expected {2,2,3,3}, got %+v", rects[0])
	}

	// Verify screen content.
	if c := comp.Screen().Get(2, 2); c.Char != 'Z' {
		t.Errorf("(2,2): expected 'Z', got %q", c.Char)
	}
	if c := comp.Screen().Get(0, 0); c.Char != 'A' {
		t.Errorf("(0,0): expected 'A', got %q", c.Char)
	}
}

func TestCompositor_ComposeDirty_Occluded(t *testing.T) {
	// A is at z=0, B is at z=100 covering the same area.
	bufA := filledBuffer(5, 5, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 0}

	bufB := filledBuffer(5, 5, 'B', "#00ff00")
	layerB := &Layer{ID: "b", Buffer: bufB, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 100}

	comp := NewCompositor(5, 5)
	comp.SetLayers([]*Layer{layerA, layerB})
	comp.ComposeAll()

	// Now mark A as dirty — but B occludes it entirely.
	bufA.Set(2, 2, buffer.Cell{Char: 'X', Foreground: "#ffffff"})
	dr := buffer.Rect{X: 2, Y: 2, W: 1, H: 1}
	layerA.DirtyRect = &dr

	comp.ComposeDirty([]*Layer{layerA})

	// Screen should still show 'B' at (2,2) because B occludes A.
	if c := comp.Screen().Get(2, 2); c.Char != 'B' {
		t.Errorf("(2,2): expected 'B' (occluded), got %q", c.Char)
	}
}

func TestCompositor_ComposeRects_WindowMove(t *testing.T) {
	bufA := filledBuffer(10, 10, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 10, H: 10}, ZIndex: 0}

	bufB := filledBuffer(3, 3, 'B', "#00ff00")
	layerB := &Layer{ID: "b", Buffer: bufB, Rect: buffer.Rect{X: 0, Y: 0, W: 3, H: 3}, ZIndex: 100}

	comp := NewCompositor(10, 10)
	comp.SetLayers([]*Layer{layerA, layerB})
	comp.ComposeAll()

	// Verify initial state.
	if c := comp.Screen().Get(0, 0); c.Char != 'B' {
		t.Fatalf("(0,0) before move: expected 'B', got %q", c.Char)
	}

	// Simulate window move: B moves from (0,0) to (5,5).
	oldRect := layerB.Rect
	layerB.Rect = buffer.Rect{X: 5, Y: 5, W: 3, H: 3}
	comp.SetLayers([]*Layer{layerA, layerB}) // rebuild occlusion map

	// Recompose old and new rects.
	rects := comp.ComposeRects([]buffer.Rect{oldRect, layerB.Rect})
	if len(rects) != 2 {
		t.Fatalf("expected 2 rects, got %d", len(rects))
	}

	// Old position should now show 'A' (background layer).
	if c := comp.Screen().Get(0, 0); c.Char != 'A' {
		t.Errorf("(0,0) after move: expected 'A', got %q", c.Char)
	}
	// New position should show 'B'.
	if c := comp.Screen().Get(5, 5); c.Char != 'B' {
		t.Errorf("(5,5) after move: expected 'B', got %q", c.Char)
	}
}

func TestCompositor_ManyLayers(t *testing.T) {
	layers := make([]*Layer, 100)
	for i := 0; i < 100; i++ {
		buf := filledBuffer(10, 10, rune('A'+i%26), "#ff0000")
		layers[i] = &Layer{
			ID:     string(rune('0' + i)),
			Buffer: buf,
			Rect:   buffer.Rect{X: 0, Y: 0, W: 10, H: 10},
			ZIndex: i,
		}
	}

	comp := NewCompositor(10, 10)
	comp.SetLayers(layers)
	screen := comp.ComposeAll()

	// The highest z-index layer (index 99) should own all cells.
	expected := rune('A' + 99%26)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if c := screen.Get(x, y); c.Char != expected {
				t.Fatalf("(%d,%d): expected %q, got %q", x, y, expected, c.Char)
			}
		}
	}
}

func TestCompositor_TransparentCells(t *testing.T) {
	// Bottom layer: all 'A'.
	bufA := filledBuffer(5, 5, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 0}

	// Top layer: only (1,1) and (3,3) are filled, rest are zero (transparent).
	bufB := sparseBuffer(5, 5, map[[2]int]buffer.Cell{
		{1, 1}: {Char: 'B', Foreground: "#00ff00"},
		{3, 3}: {Char: 'B', Foreground: "#00ff00"},
	})
	layerB := &Layer{ID: "b", Buffer: bufB, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 100}

	comp := NewCompositor(5, 5)
	comp.SetLayers([]*Layer{layerA, layerB})
	screen := comp.ComposeAll()

	// (1,1) should be 'B' — top layer has content.
	if c := screen.Get(1, 1); c.Char != 'B' {
		t.Errorf("(1,1): expected 'B', got %q", c.Char)
	}
	// (3,3) should be 'B'.
	if c := screen.Get(3, 3); c.Char != 'B' {
		t.Errorf("(3,3): expected 'B', got %q", c.Char)
	}
	// (0,0) should be 'A' — top layer is transparent here.
	if c := screen.Get(0, 0); c.Char != 'A' {
		t.Errorf("(0,0): expected 'A', got %q", c.Char)
	}
	// (2,2) should be 'A'.
	if c := screen.Get(2, 2); c.Char != 'A' {
		t.Errorf("(2,2): expected 'A', got %q", c.Char)
	}
}

func TestCompositor_CachedSort(t *testing.T) {
	// This test verifies that the occlusion map correctly rebuilds
	// when layers are added or removed.
	bufA := filledBuffer(5, 5, 'A', "#ff0000")
	layerA := &Layer{ID: "a", Buffer: bufA, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 0}

	comp := NewCompositor(5, 5)
	comp.SetLayers([]*Layer{layerA})
	comp.ComposeAll()

	if c := comp.Screen().Get(0, 0); c.Char != 'A' {
		t.Fatalf("before add: expected 'A', got %q", c.Char)
	}

	// Add a new layer on top.
	bufB := filledBuffer(5, 5, 'B', "#00ff00")
	layerB := &Layer{ID: "b", Buffer: bufB, Rect: buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, ZIndex: 100}
	comp.SetLayers([]*Layer{layerA, layerB})
	comp.ComposeAll()

	if c := comp.Screen().Get(0, 0); c.Char != 'B' {
		t.Errorf("after add: expected 'B', got %q", c.Char)
	}

	// Remove the top layer.
	comp.SetLayers([]*Layer{layerA})
	comp.ComposeAll()

	if c := comp.Screen().Get(0, 0); c.Char != 'A' {
		t.Errorf("after remove: expected 'A', got %q", c.Char)
	}
}
