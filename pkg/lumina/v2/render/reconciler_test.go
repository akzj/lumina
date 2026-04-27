package render

import "testing"

func TestReconcile_NoChange(t *testing.T) {
	// Create a node, reconcile with identical descriptor → no dirty flags
	node := NewNode("box")
	node.Style = Style{Width: 10, Height: 5, Background: "#333", Right: -1, Bottom: -1}
	node.PaintDirty = false
	node.LayoutDirty = false

	desc := Descriptor{Type: "box", Style: Style{Width: 10, Height: 5, Background: "#333", Right: -1, Bottom: -1}}
	changed := Reconcile(node, desc)

	if changed {
		t.Error("expected no change")
	}
	if node.PaintDirty || node.LayoutDirty {
		t.Error("expected no dirty flags")
	}
}

func TestReconcile_StylePaintOnly(t *testing.T) {
	// Change only color → PaintDirty but NOT LayoutDirty
	node := NewNode("box")
	node.Style = Style{Width: 10, Height: 5, Background: "#333", Right: -1, Bottom: -1}
	node.PaintDirty = false
	node.LayoutDirty = false

	desc := Descriptor{Type: "box", Style: Style{Width: 10, Height: 5, Background: "#F00", Right: -1, Bottom: -1}}
	Reconcile(node, desc)

	if !node.PaintDirty {
		t.Error("expected PaintDirty")
	}
	if node.LayoutDirty {
		t.Error("did not expect LayoutDirty for color-only change")
	}
}

func TestReconcile_StyleLayoutChange(t *testing.T) {
	// Change width → both LayoutDirty and PaintDirty
	node := NewNode("box")
	node.Style = Style{Width: 10, Height: 5, Right: -1, Bottom: -1}
	node.PaintDirty = false
	node.LayoutDirty = false

	desc := Descriptor{Type: "box", Style: Style{Width: 20, Height: 5, Right: -1, Bottom: -1}}
	Reconcile(node, desc)

	if !node.PaintDirty {
		t.Error("expected PaintDirty")
	}
	if !node.LayoutDirty {
		t.Error("expected LayoutDirty for width change")
	}
}

func TestReconcile_ContentChange(t *testing.T) {
	node := NewNode("text")
	node.Content = "hello"

	desc := Descriptor{Type: "text", Content: "world"}
	Reconcile(node, desc)

	if node.Content != "world" {
		t.Errorf("expected 'world', got %q", node.Content)
	}
	if !node.PaintDirty {
		t.Error("expected PaintDirty")
	}
}

func TestReconcile_ChildrenSameKeys(t *testing.T) {
	// Same children, same order → update in place
	parent := NewNode("vbox")
	child1 := NewNode("text")
	child1.Key = "a"
	child1.Content = "AAA"
	child2 := NewNode("text")
	child2.Key = "b"
	child2.Content = "BBB"
	parent.AddChild(child1)
	parent.AddChild(child2)

	descs := []Descriptor{
		{Type: "text", Key: "a", Content: "AAA-updated"},
		{Type: "text", Key: "b", Content: "BBB"},
	}
	desc := Descriptor{Type: "vbox", Children: descs}
	Reconcile(parent, desc)

	if parent.Children[0].Content != "AAA-updated" {
		t.Error("child 0 not updated")
	}
	if parent.Children[0] != child1 {
		t.Error("child 0 should be same pointer (reused)")
	}
	if parent.Children[1].Content != "BBB" {
		t.Error("child 1 should be unchanged")
	}
}

func TestReconcile_ChildAdded(t *testing.T) {
	parent := NewNode("vbox")
	child1 := NewNode("text")
	child1.Key = "a"
	parent.AddChild(child1)

	descs := []Descriptor{
		{Type: "text", Key: "a", Content: "A"},
		{Type: "text", Key: "b", Content: "B"},
	}
	desc := Descriptor{Type: "vbox", Children: descs}
	Reconcile(parent, desc)

	if len(parent.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.Children))
	}
	if parent.Children[1].Content != "B" {
		t.Error("new child not created")
	}
	if !parent.LayoutDirty {
		t.Error("expected LayoutDirty when children added")
	}
}

func TestReconcile_ChildRemoved(t *testing.T) {
	parent := NewNode("vbox")
	child1 := NewNode("text")
	child1.Key = "a"
	child2 := NewNode("text")
	child2.Key = "b"
	parent.AddChild(child1)
	parent.AddChild(child2)

	descs := []Descriptor{
		{Type: "text", Key: "a", Content: "A"},
	}
	desc := Descriptor{Type: "vbox", Children: descs}
	Reconcile(parent, desc)

	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0] != child1 {
		t.Error("remaining child should be same pointer")
	}
}

func TestReconcile_ChildReorder(t *testing.T) {
	parent := NewNode("vbox")
	child1 := NewNode("text")
	child1.Key = "a"
	child1.Content = "A"
	child2 := NewNode("text")
	child2.Key = "b"
	child2.Content = "B"
	parent.AddChild(child1)
	parent.AddChild(child2)

	// Reverse order
	descs := []Descriptor{
		{Type: "text", Key: "b", Content: "B"},
		{Type: "text", Key: "a", Content: "A"},
	}
	desc := Descriptor{Type: "vbox", Children: descs}
	Reconcile(parent, desc)

	if parent.Children[0] != child2 {
		t.Error("first child should now be 'b'")
	}
	if parent.Children[1] != child1 {
		t.Error("second child should now be 'a'")
	}
}

func TestReconcile_EventHandlerUpdate(t *testing.T) {
	node := NewNode("box")
	node.OnClick = 42

	desc := Descriptor{Type: "box", OnClick: 99}
	Reconcile(node, desc)

	if node.OnClick != 99 {
		t.Errorf("expected OnClick=99, got %d", node.OnClick)
	}
}

func TestReconcile_DeepTree(t *testing.T) {
	// Build a 3-level tree, change a deep node
	root := NewNode("vbox")
	row := NewNode("hbox")
	row.Key = "row0"
	cell := NewNode("text")
	cell.Key = "cell0"
	cell.Content = "·"
	row.AddChild(cell)
	root.AddChild(row)

	desc := Descriptor{
		Type: "vbox",
		Children: []Descriptor{
			{Type: "hbox", Key: "row0", Children: []Descriptor{
				{Type: "text", Key: "cell0", Content: "█"},
			}},
		},
	}
	Reconcile(root, desc)

	if root.Children[0].Children[0].Content != "█" {
		t.Error("deep node not updated")
	}
	if root.Children[0].Children[0] != cell {
		t.Error("deep node should be same pointer (reused)")
	}
}

func TestComponent_SetState(t *testing.T) {
	comp := NewComponent("c1", "Cell", "Cell")
	comp.Dirty = false

	comp.SetState("hovered", true)
	if !comp.Dirty {
		t.Error("expected Dirty after SetState")
	}

	// Same value → no change
	comp.Dirty = false
	comp.SetState("hovered", true)
	if comp.Dirty {
		t.Error("expected no Dirty for same value")
	}
}

func TestComponent_FindChild(t *testing.T) {
	parent := NewComponent("p", "Grid", "Grid")
	child := NewComponent("c1", "Cell", "Cell")
	child.ID = "cell-1"
	parent.ChildMap["Cell:cell-1"] = child

	found := parent.FindChild("Cell", "cell-1")
	if found != child {
		t.Error("FindChild failed")
	}

	notFound := parent.FindChild("Cell", "cell-999")
	if notFound != nil {
		t.Error("expected nil for missing child")
	}
}

func TestMarkLayoutDirty_Propagation(t *testing.T) {
	root := NewNode("vbox")
	root.Style.Width = 80
	root.Style.Height = 24
	child := NewNode("hbox")
	grandchild := NewNode("text")
	root.AddChild(child)
	child.AddChild(grandchild)

	// Clear all dirty flags
	root.LayoutDirty = false
	child.LayoutDirty = false
	grandchild.LayoutDirty = false

	// Mark grandchild dirty → should propagate to child, then stop at root (fixed size)
	grandchild.MarkLayoutDirty()

	if !grandchild.LayoutDirty {
		t.Error("grandchild should be layout dirty")
	}
	if !child.LayoutDirty {
		t.Error("child should be layout dirty (propagated)")
	}
	if !root.LayoutDirty {
		t.Error("root should be layout dirty (boundary)")
	}
}

func TestReconcile_UnkeyedChildren(t *testing.T) {
	// Children without keys — falls through to map-based matching
	parent := NewNode("vbox")
	child1 := NewNode("text")
	child1.Content = "A"
	child2 := NewNode("text")
	child2.Content = "B"
	parent.AddChild(child1)
	parent.AddChild(child2)

	// Add a third unkeyed child
	descs := []Descriptor{
		{Type: "text", Content: "A"},
		{Type: "text", Content: "B"},
		{Type: "text", Content: "C"},
	}
	desc := Descriptor{Type: "vbox", Children: descs}
	Reconcile(parent, desc)

	if len(parent.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(parent.Children))
	}
	if parent.Children[2].Content != "C" {
		t.Error("third child not created")
	}
}

func TestReconcileCollectRefs_UpdatedHandlers(t *testing.T) {
	// When event handlers change, old refs should be collected
	node := NewNode("box")
	node.OnClick = 10
	node.OnMouseEnter = 20

	desc := Descriptor{Type: "box", OnClick: 30, OnMouseEnter: 40}

	var freedRefs []int64
	ReconcileCollectRefs(node, desc, &freedRefs)

	if node.OnClick != 30 {
		t.Errorf("expected OnClick=30, got %d", node.OnClick)
	}
	if node.OnMouseEnter != 40 {
		t.Errorf("expected OnMouseEnter=40, got %d", node.OnMouseEnter)
	}

	// Old refs (10, 20) should be in freedRefs
	if len(freedRefs) != 2 {
		t.Fatalf("expected 2 freed refs, got %d: %v", len(freedRefs), freedRefs)
	}
	found10, found20 := false, false
	for _, r := range freedRefs {
		if r == 10 {
			found10 = true
		}
		if r == 20 {
			found20 = true
		}
	}
	if !found10 || !found20 {
		t.Errorf("expected freed refs to contain 10 and 20, got %v", freedRefs)
	}
}

func TestReconcileCollectRefs_RemovedChildren(t *testing.T) {
	// When children are removed, their refs should be collected
	parent := NewNode("vbox")
	child := NewNode("box")
	child.Key = "a"
	child.OnClick = 50
	child.OnKeyDown = 60
	parent.AddChild(child)

	desc := Descriptor{Type: "vbox", Children: nil} // remove all children

	var freedRefs []int64
	ReconcileCollectRefs(parent, desc, &freedRefs)

	// Refs from removed child should be collected
	if len(freedRefs) != 2 {
		t.Fatalf("expected 2 freed refs from removed child, got %d: %v", len(freedRefs), freedRefs)
	}
	found50, found60 := false, false
	for _, r := range freedRefs {
		if r == 50 {
			found50 = true
		}
		if r == 60 {
			found60 = true
		}
	}
	if !found50 || !found60 {
		t.Errorf("expected freed refs to contain 50 and 60, got %v", freedRefs)
	}
}

func TestReconcileCollectRefs_NilDoesNotPanic(t *testing.T) {
	// Reconcile without freedRefs (nil) should still work
	node := NewNode("box")
	node.OnClick = 10
	desc := Descriptor{Type: "box", OnClick: 20}
	Reconcile(node, desc) // uses nil freedRefs internally
	if node.OnClick != 20 {
		t.Errorf("expected OnClick=20, got %d", node.OnClick)
	}
}
