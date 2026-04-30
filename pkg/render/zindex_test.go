package render

import "testing"

// TestZIndex_PaintOrder verifies that children with higher ZIndex paint on top
// (later in paint order), so their content overwrites lower-ZIndex siblings.
func TestZIndex_PaintOrder(t *testing.T) {
	// Three overlapping 3x1 children at the same position.
	// Array order: A(z=0), B(z=2), C(z=1)
	// Paint order (ascending ZIndex): A(z=0), C(z=1), B(z=2)
	// Last painted wins → cell shows "B"
	root := &Node{
		Type: "box",
		X: 0, Y: 0, W: 5, H: 3,
		Style: Style{},
	}
	childA := &Node{
		Type: "text", Content: "AAA",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 0, Foreground: "#ff0000"},
		Parent: root,
	}
	childB := &Node{
		Type: "text", Content: "BBB",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 2, Foreground: "#00ff00"},
		Parent: root,
	}
	childC := &Node{
		Type: "text", Content: "CCC",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 1, Foreground: "#0000ff"},
		Parent: root,
	}
	root.Children = []*Node{childA, childB, childC}

	buf := NewCellBuffer(5, 3)
	PaintFull(buf, root)

	// B has highest ZIndex (2), so it paints last → its chars should be visible
	for x := 0; x < 3; x++ {
		c := buf.Get(x, 0)
		if c.Ch != 'B' {
			t.Errorf("pos (%d,0): expected 'B' (highest ZIndex=2), got %q", x, c.Ch)
		}
	}
}

// TestZIndex_HitTest verifies that hit-test returns the highest ZIndex child.
func TestZIndex_HitTest(t *testing.T) {
	root := &Node{
		Type: "box",
		X: 0, Y: 0, W: 5, H: 3,
		Style: Style{},
	}
	childA := &Node{
		Type: "box", ID: "A",
		X: 0, Y: 0, W: 3, H: 3,
		Style:  Style{ZIndex: 0},
		Parent: root,
	}
	childB := &Node{
		Type: "box", ID: "B",
		X: 0, Y: 0, W: 3, H: 3,
		Style:  Style{ZIndex: 2},
		Parent: root,
	}
	childC := &Node{
		Type: "box", ID: "C",
		X: 0, Y: 0, W: 3, H: 3,
		Style:  Style{ZIndex: 1},
		Parent: root,
	}
	root.Children = []*Node{childA, childB, childC}

	hit := HitTest(root, 1, 1)
	if hit == nil {
		t.Fatal("expected hit, got nil")
	}
	if hit.ID != "B" {
		t.Errorf("expected hit on 'B' (highest ZIndex=2), got %q", hit.ID)
	}
}

// TestZIndex_DefaultPreservesOrder verifies that when all children have ZIndex=0,
// the existing behavior is preserved: last child in array paints on top and
// is hit-tested first.
func TestZIndex_DefaultPreservesOrder(t *testing.T) {
	root := &Node{
		Type: "box",
		X: 0, Y: 0, W: 5, H: 3,
		Style: Style{},
	}
	childA := &Node{
		Type: "text", Content: "AAA", ID: "A",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{Foreground: "#ff0000"},
		Parent: root,
	}
	childB := &Node{
		Type: "text", Content: "BBB", ID: "B",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{Foreground: "#00ff00"},
		Parent: root,
	}
	root.Children = []*Node{childA, childB}

	// Paint: B is last in array → paints on top → cells show "B"
	buf := NewCellBuffer(5, 3)
	PaintFull(buf, root)
	for x := 0; x < 3; x++ {
		c := buf.Get(x, 0)
		if c.Ch != 'B' {
			t.Errorf("pos (%d,0): expected 'B' (last in array, all ZIndex=0), got %q", x, c.Ch)
		}
	}

	// Hit-test: B is last in array → hit first (reverse order)
	// Need box type for hit-test bounds checking
	childA.Type = "box"
	childA.Content = ""
	childB.Type = "box"
	childB.Content = ""
	childA.W = 3
	childA.H = 3
	childB.W = 3
	childB.H = 3

	hit := HitTest(root, 1, 1)
	if hit == nil {
		t.Fatal("expected hit, got nil")
	}
	if hit.ID != "B" {
		t.Errorf("expected hit on 'B' (last in array, all ZIndex=0), got %q", hit.ID)
	}
}

// TestZIndex_Negative verifies that negative ZIndex paints below siblings with ZIndex=0.
func TestZIndex_Negative(t *testing.T) {
	root := &Node{
		Type: "box",
		X: 0, Y: 0, W: 5, H: 3,
		Style: Style{},
	}
	// childA has ZIndex=-1 (should be painted first/below)
	childA := &Node{
		Type: "text", Content: "AAA", ID: "A",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: -1, Foreground: "#ff0000"},
		Parent: root,
	}
	// childB has ZIndex=0 (should be painted on top)
	childB := &Node{
		Type: "text", Content: "BBB", ID: "B",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 0, Foreground: "#00ff00"},
		Parent: root,
	}
	root.Children = []*Node{childA, childB}

	buf := NewCellBuffer(5, 3)
	PaintFull(buf, root)

	// B has ZIndex=0 > A's ZIndex=-1, so B paints on top
	for x := 0; x < 3; x++ {
		c := buf.Get(x, 0)
		if c.Ch != 'B' {
			t.Errorf("pos (%d,0): expected 'B' (ZIndex=0 > -1), got %q", x, c.Ch)
		}
	}

	// Even if A is first in array, B should be hit first (higher ZIndex)
	childA.Type = "box"
	childA.Content = ""
	childB.Type = "box"
	childB.Content = ""
	childA.H = 3
	childB.H = 3

	hit := HitTest(root, 1, 1)
	if hit == nil {
		t.Fatal("expected hit, got nil")
	}
	if hit.ID != "B" {
		t.Errorf("expected hit on 'B' (ZIndex=0, higher than A's -1), got %q", hit.ID)
	}
}

// TestZIndex_EqualZIndex_ArrayOrderPreserved verifies that when multiple children
// have the same non-zero ZIndex, array order is preserved (stable sort).
func TestZIndex_EqualZIndex_ArrayOrderPreserved(t *testing.T) {
	root := &Node{
		Type: "box",
		X: 0, Y: 0, W: 5, H: 3,
		Style: Style{},
	}
	// A and B both have ZIndex=5, C has ZIndex=1
	// Paint order: C(z=1), A(z=5), B(z=5) — B paints last among equals
	childA := &Node{
		Type: "text", Content: "AAA", ID: "A",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 5, Foreground: "#ff0000"},
		Parent: root,
	}
	childB := &Node{
		Type: "text", Content: "BBB", ID: "B",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 5, Foreground: "#00ff00"},
		Parent: root,
	}
	childC := &Node{
		Type: "text", Content: "CCC", ID: "C",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 1, Foreground: "#0000ff"},
		Parent: root,
	}
	root.Children = []*Node{childA, childB, childC}

	buf := NewCellBuffer(5, 3)
	PaintFull(buf, root)

	// B is last among ZIndex=5 siblings → paints last → visible
	for x := 0; x < 3; x++ {
		c := buf.Get(x, 0)
		if c.Ch != 'B' {
			t.Errorf("pos (%d,0): expected 'B' (last among ZIndex=5), got %q", x, c.Ch)
		}
	}
}

// TestZIndex_ComponentChildren verifies z-index works through component placeholders.
func TestZIndex_ComponentChildren(t *testing.T) {
	root := &Node{
		Type: "component",
		X: 0, Y: 0, W: 5, H: 3,
		Style: Style{},
	}
	childA := &Node{
		Type: "text", Content: "AAA", ID: "A",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 1, Foreground: "#ff0000"},
		Parent: root,
	}
	childB := &Node{
		Type: "text", Content: "BBB", ID: "B",
		X: 0, Y: 0, W: 3, H: 1,
		Style:  Style{ZIndex: 10, Foreground: "#00ff00"},
		Parent: root,
	}
	root.Children = []*Node{childA, childB}

	buf := NewCellBuffer(5, 3)
	PaintFull(buf, root)

	// B has higher ZIndex → paints on top
	for x := 0; x < 3; x++ {
		c := buf.Get(x, 0)
		if c.Ch != 'B' {
			t.Errorf("pos (%d,0): expected 'B' (ZIndex=10 > 1), got %q", x, c.Ch)
		}
	}
}

// TestZIndex_HitTestComponent verifies z-index hit-test through component nodes.
func TestZIndex_HitTestComponent(t *testing.T) {
	root := &Node{
		Type: "component",
		X: 0, Y: 0, W: 5, H: 5,
		Style: Style{},
	}
	childA := &Node{
		Type: "box", ID: "A",
		X: 0, Y: 0, W: 5, H: 5,
		Style:  Style{ZIndex: 1},
		Parent: root,
	}
	childB := &Node{
		Type: "box", ID: "B",
		X: 0, Y: 0, W: 5, H: 5,
		Style:  Style{ZIndex: 10},
		Parent: root,
	}
	root.Children = []*Node{childA, childB}

	hit := HitTest(root, 2, 2)
	if hit == nil {
		t.Fatal("expected hit, got nil")
	}
	if hit.ID != "B" {
		t.Errorf("expected hit on 'B' (ZIndex=10 > 1), got %q", hit.ID)
	}
}

// TestZIndex_PaintOrderHelpers verifies the helper functions directly.
func TestZIndex_PaintOrderHelpers(t *testing.T) {
	t.Run("hasNonZeroZIndex_allZero", func(t *testing.T) {
		children := []*Node{
			{Style: Style{ZIndex: 0}},
			{Style: Style{ZIndex: 0}},
		}
		if hasNonZeroZIndex(children) {
			t.Error("expected false for all-zero ZIndex")
		}
	})

	t.Run("hasNonZeroZIndex_oneNonZero", func(t *testing.T) {
		children := []*Node{
			{Style: Style{ZIndex: 0}},
			{Style: Style{ZIndex: 5}},
		}
		if !hasNonZeroZIndex(children) {
			t.Error("expected true when one child has non-zero ZIndex")
		}
	})

	t.Run("paintOrderChildren_noSort", func(t *testing.T) {
		children := []*Node{
			{ID: "A", Style: Style{ZIndex: 0}},
			{ID: "B", Style: Style{ZIndex: 0}},
		}
		result := paintOrderChildren(children)
		// Should return original slice (same pointer)
		if &result[0] != &children[0] {
			t.Error("expected original slice when no z-index set")
		}
	})

	t.Run("paintOrderChildren_sorted", func(t *testing.T) {
		children := []*Node{
			{ID: "A", Style: Style{ZIndex: 3}},
			{ID: "B", Style: Style{ZIndex: 1}},
			{ID: "C", Style: Style{ZIndex: 2}},
		}
		result := paintOrderChildren(children)
		expected := []string{"B", "C", "A"}
		for i, id := range expected {
			if result[i].ID != id {
				t.Errorf("position %d: expected %q, got %q", i, id, result[i].ID)
			}
		}
		// Original slice should be unmodified
		if children[0].ID != "A" || children[1].ID != "B" || children[2].ID != "C" {
			t.Error("original children slice was modified")
		}
	})

	t.Run("hitTestOrderChildren_sorted", func(t *testing.T) {
		children := []*Node{
			{ID: "A", Style: Style{ZIndex: 3}},
			{ID: "B", Style: Style{ZIndex: 1}},
			{ID: "C", Style: Style{ZIndex: 2}},
		}
		result := hitTestOrderChildren(children)
		// Should be reverse of paint order: A(3), C(2), B(1)
		expected := []string{"A", "C", "B"}
		for i, id := range expected {
			if result[i].ID != id {
				t.Errorf("position %d: expected %q, got %q", i, id, result[i].ID)
			}
		}
	})

	t.Run("hitTestOrderChildren_nil_when_all_zero", func(t *testing.T) {
		children := []*Node{
			{ID: "A", Style: Style{ZIndex: 0}},
			{ID: "B", Style: Style{ZIndex: 0}},
		}
		result := hitTestOrderChildren(children)
		if result != nil {
			t.Error("expected nil when all ZIndex=0")
		}
	})
}
