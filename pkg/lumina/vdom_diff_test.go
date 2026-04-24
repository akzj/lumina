package lumina

import (
	"testing"
)

// helper to build a VNode with type, content, and optional props.
func vn(nodeType, content string, props map[string]any, children ...*VNode) *VNode {
	n := NewVNode(nodeType)
	n.Content = content
	if props != nil {
		for k, v := range props {
			n.Props[k] = v
		}
	}
	for _, c := range children {
		n.AddChild(c)
	}
	return n
}

func TestDiffVNode_BothNil(t *testing.T) {
	patches := DiffVNode(nil, nil)
	if len(patches) != 0 {
		t.Fatalf("expected 0 patches, got %d", len(patches))
	}
}

func TestDiffVNode_SameNode(t *testing.T) {
	a := vn("text", "hello", nil)
	b := vn("text", "hello", nil)
	patches := DiffVNode(a, b)
	if len(patches) != 0 {
		t.Fatalf("expected 0 patches for identical nodes, got %d: %+v", len(patches), patches)
	}
}

func TestDiffVNode_NilToNew(t *testing.T) {
	b := vn("text", "hello", nil)
	patches := DiffVNode(nil, b)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Type != PatchReplace {
		t.Fatalf("expected PatchReplace, got %s", patches[0].Type)
	}
	if patches[0].NewNode != b {
		t.Fatal("expected NewNode to be b")
	}
}

func TestDiffVNode_OldToNil(t *testing.T) {
	a := vn("text", "hello", nil)
	patches := DiffVNode(a, nil)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Type != PatchRemove {
		t.Fatalf("expected PatchRemove, got %s", patches[0].Type)
	}
}

func TestDiffVNode_DifferentType(t *testing.T) {
	a := vn("text", "hello", nil)
	b := vn("box", "", nil)
	patches := DiffVNode(a, b)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Type != PatchReplace {
		t.Fatalf("expected PatchReplace, got %s", patches[0].Type)
	}
}

func TestDiffVNode_PropsChanged(t *testing.T) {
	a := vn("box", "", map[string]any{"color": "red"})
	b := vn("box", "", map[string]any{"color": "blue"})
	patches := DiffVNode(a, b)

	found := false
	for _, p := range patches {
		if p.Type == PatchUpdate {
			found = true
		}
	}
	if !found {
		t.Fatal("expected PatchUpdate for changed props")
	}
}

func TestDiffVNode_TextChanged(t *testing.T) {
	a := vn("text", "hello", nil)
	b := vn("text", "world", nil)
	patches := DiffVNode(a, b)

	found := false
	for _, p := range patches {
		if p.Type == PatchText {
			found = true
		}
	}
	if !found {
		t.Fatal("expected PatchText for changed content")
	}
}

func TestDiffVNode_ChildAdded(t *testing.T) {
	a := vn("box", "", nil,
		vn("text", "a", nil),
	)
	b := vn("box", "", nil,
		vn("text", "a", nil),
		vn("text", "b", nil),
	)
	patches := DiffVNode(a, b)

	found := false
	for _, p := range patches {
		if p.Type == PatchInsert && p.Index == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected PatchInsert at index 1, got patches: %+v", patches)
	}
}

func TestDiffVNode_ChildRemoved(t *testing.T) {
	a := vn("box", "", nil,
		vn("text", "a", nil),
		vn("text", "b", nil),
	)
	b := vn("box", "", nil,
		vn("text", "a", nil),
	)
	patches := DiffVNode(a, b)

	found := false
	for _, p := range patches {
		if p.Type == PatchRemove && p.Index == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected PatchRemove at index 1, got patches: %+v", patches)
	}
}

func TestDiffVNode_KeyedReorder(t *testing.T) {
	a := vn("box", "", nil,
		vn("text", "a", map[string]any{"key": "ka"}),
		vn("text", "b", map[string]any{"key": "kb"}),
		vn("text", "c", map[string]any{"key": "kc"}),
	)
	// Reverse order.
	b := vn("box", "", nil,
		vn("text", "c", map[string]any{"key": "kc"}),
		vn("text", "b", map[string]any{"key": "kb"}),
		vn("text", "a", map[string]any{"key": "ka"}),
	)
	patches := DiffVNode(a, b)

	reorders := 0
	for _, p := range patches {
		if p.Type == PatchReorder {
			reorders++
		}
	}
	if reorders == 0 {
		t.Fatal("expected PatchReorder patches for reordered keyed children")
	}
}

func TestDiffVNode_KeyedInsert(t *testing.T) {
	a := vn("box", "", nil,
		vn("text", "a", map[string]any{"key": "ka"}),
		vn("text", "c", map[string]any{"key": "kc"}),
	)
	b := vn("box", "", nil,
		vn("text", "a", map[string]any{"key": "ka"}),
		vn("text", "b", map[string]any{"key": "kb"}),
		vn("text", "c", map[string]any{"key": "kc"}),
	)
	patches := DiffVNode(a, b)

	found := false
	for _, p := range patches {
		if p.Type == PatchInsert {
			found = true
		}
	}
	if !found {
		t.Fatal("expected PatchInsert for new keyed child")
	}
}

func TestDiffVNode_KeyedRemove(t *testing.T) {
	a := vn("box", "", nil,
		vn("text", "a", map[string]any{"key": "ka"}),
		vn("text", "b", map[string]any{"key": "kb"}),
		vn("text", "c", map[string]any{"key": "kc"}),
	)
	b := vn("box", "", nil,
		vn("text", "a", map[string]any{"key": "ka"}),
		vn("text", "c", map[string]any{"key": "kc"}),
	)
	patches := DiffVNode(a, b)

	found := false
	for _, p := range patches {
		if p.Type == PatchRemove {
			found = true
		}
	}
	if !found {
		t.Fatal("expected PatchRemove for removed keyed child")
	}
}

func TestDiffVNode_DeepNested(t *testing.T) {
	a := vn("box", "", nil,
		vn("box", "", nil,
			vn("text", "deep", nil),
		),
	)
	b := vn("box", "", nil,
		vn("box", "", nil,
			vn("text", "changed", nil),
		),
	)
	patches := DiffVNode(a, b)

	found := false
	for _, p := range patches {
		if p.Type == PatchText {
			found = true
			// Path should be [0, 0] (first child of first child).
			if len(p.Path) != 2 || p.Path[0] != 0 || p.Path[1] != 0 {
				t.Fatalf("expected path [0,0], got %v", p.Path)
			}
		}
	}
	if !found {
		t.Fatal("expected PatchText for deep nested change")
	}
}

func TestDiffVNode_MixedChanges(t *testing.T) {
	a := vn("box", "", nil,
		vn("text", "unchanged", nil),
		vn("text", "will-change", nil),
		vn("text", "will-remove", nil),
	)
	b := vn("box", "", nil,
		vn("text", "unchanged", nil),
		vn("text", "changed!", nil),
		vn("box", "", nil), // new type at index 2
		vn("text", "added", nil),
	)
	patches := DiffVNode(a, b)

	types := map[PatchType]int{}
	for _, p := range patches {
		types[p.Type]++
	}

	if types[PatchText] < 1 {
		t.Fatal("expected at least 1 PatchText")
	}
	if types[PatchReplace] < 1 {
		t.Fatal("expected at least 1 PatchReplace (type change at index 2)")
	}
	if types[PatchInsert] < 1 {
		t.Fatal("expected at least 1 PatchInsert (new child at index 3)")
	}
}

func TestShouldFullRerender(t *testing.T) {
	root := vn("box", "", nil,
		vn("text", "a", nil),
		vn("text", "b", nil),
	)
	// 3 total nodes. >50% = 2+ patches.
	smallPatches := []Patch{{Type: PatchText}}
	if ShouldFullRerender(smallPatches, root) {
		t.Fatal("should not full-rerender with 1 patch on 3 nodes")
	}

	bigPatches := []Patch{{}, {}, {}}
	if !ShouldFullRerender(bigPatches, root) {
		t.Fatal("should full-rerender with 3 patches on 3 nodes")
	}
}

func TestCountNodes(t *testing.T) {
	root := vn("box", "", nil,
		vn("text", "a", nil),
		vn("box", "", nil,
			vn("text", "b", nil),
		),
	)
	if n := countNodes(root); n != 4 {
		t.Fatalf("expected 4 nodes, got %d", n)
	}
}

func TestPropsEqual(t *testing.T) {
	a := map[string]any{"color": "red", "bold": true}
	b := map[string]any{"color": "red", "bold": true}
	if !propsEqual(a, b) {
		t.Fatal("expected equal")
	}

	c := map[string]any{"color": "blue", "bold": true}
	if propsEqual(a, c) {
		t.Fatal("expected not equal")
	}

	// Key and children should be ignored.
	d := map[string]any{"color": "red", "bold": true, "key": "k1", "children": "ignored"}
	if !propsEqual(a, d) {
		t.Fatal("expected equal (key and children ignored)")
	}
}

func TestClearRegion(t *testing.T) {
	frame := NewFrame(10, 5)
	frame.Cells[1][2] = Cell{Char: 'X', Foreground: "#FF0000"}
	frame.Cells[1][3] = Cell{Char: 'Y'}

	node := &VNode{X: 2, Y: 1, W: 2, H: 1}
	clearRegion(frame, node)

	if frame.Cells[1][2].Char != ' ' || frame.Cells[1][2].Foreground != "" {
		t.Fatal("expected cell (2,1) to be cleared")
	}
	if frame.Cells[1][3].Char != ' ' {
		t.Fatal("expected cell (3,1) to be cleared")
	}
}
