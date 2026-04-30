package render

import "testing"

func TestEngine_FocusPrev(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	root := NewNode("box")
	root.W = 40
	root.H = 10

	a := NewNode("input")
	a.ID = "a"
	a.Focusable = true
	a.W = 10
	a.H = 1
	root.AddChild(a)

	b := NewNode("input")
	b.ID = "b"
	b.Focusable = true
	b.W = 10
	b.H = 1
	root.AddChild(b)

	c := NewNode("input")
	c.ID = "c"
	c.Focusable = true
	c.W = 10
	c.H = 1
	root.AddChild(c)

	e.syncMainLayer()
	e.Layers()[0].Root = root

	// FocusPrev with nothing focused → wraps to last (c)
	e.FocusPrev()
	if e.FocusedNode() != c {
		t.Errorf("expected focus on c, got %v", e.FocusedNode())
	}

	// FocusPrev → b
	e.FocusPrev()
	if e.FocusedNode() != b {
		t.Errorf("expected focus on b, got %v", e.FocusedNode())
	}

	// FocusPrev → a
	e.FocusPrev()
	if e.FocusedNode() != a {
		t.Errorf("expected focus on a, got %v", e.FocusedNode())
	}

	// FocusPrev → wraps to c
	e.FocusPrev()
	if e.FocusedNode() != c {
		t.Errorf("expected focus to wrap to c, got %v", e.FocusedNode())
	}
}

func TestEngine_FocusPrev_ModalLayer(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	// Main layer with focusable nodes
	mainRoot := NewNode("box")
	mainRoot.W = 40
	mainRoot.H = 10
	mainInput := NewNode("input")
	mainInput.ID = "main-input"
	mainInput.Focusable = true
	mainInput.W = 10
	mainInput.H = 1
	mainRoot.AddChild(mainInput)
	e.syncMainLayer()
	e.Layers()[0].Root = mainRoot

	// Modal layer with its own focusable node
	modalRoot := NewNode("box")
	modalRoot.W = 20
	modalRoot.H = 5
	modalInput := NewNode("input")
	modalInput.ID = "modal-input"
	modalInput.Focusable = true
	modalInput.W = 10
	modalInput.H = 1
	modalRoot.AddChild(modalInput)
	e.CreateLayer("modal", modalRoot, true)

	// FocusPrev should only cycle within the modal layer
	e.FocusPrev()
	if e.FocusedNode() != modalInput {
		t.Error("expected focus on modal input, not main input")
	}
}

func TestEngine_FindNodeByID(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	root := NewNode("box")
	root.ID = "root"
	root.W = 40
	root.H = 10

	child := NewNode("text")
	child.ID = "child-1"
	root.AddChild(child)

	grandchild := NewNode("input")
	grandchild.ID = "deep-node"
	child.AddChild(grandchild)

	e.syncMainLayer()
	e.Layers()[0].Root = root

	// Find root
	if found := e.FindNodeByID("root"); found != root {
		t.Error("expected to find root node")
	}

	// Find nested node
	if found := e.FindNodeByID("deep-node"); found != grandchild {
		t.Error("expected to find deep-node")
	}

	// Not found
	if found := e.FindNodeByID("nonexistent"); found != nil {
		t.Error("expected nil for nonexistent ID")
	}

	// Empty ID
	if found := e.FindNodeByID(""); found != nil {
		t.Error("expected nil for empty ID")
	}
}

func TestEngine_FocusByID(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	root := NewNode("box")
	root.W = 40
	root.H = 10

	a := NewNode("input")
	a.ID = "input-a"
	a.Focusable = true
	a.W = 10
	a.H = 1
	root.AddChild(a)

	b := NewNode("input")
	b.ID = "input-b"
	b.Focusable = true
	b.W = 10
	b.H = 1
	root.AddChild(b)

	disabled := NewNode("input")
	disabled.ID = "input-disabled"
	disabled.Focusable = true
	disabled.Disabled = true
	disabled.W = 10
	disabled.H = 1
	root.AddChild(disabled)

	e.syncMainLayer()
	e.Layers()[0].Root = root

	// Focus by ID
	if ok := e.FocusByID("input-b"); !ok {
		t.Error("expected FocusByID to succeed")
	}
	if e.FocusedNode() != b {
		t.Error("expected focused node to be input-b")
	}

	// Focus a different node
	if ok := e.FocusByID("input-a"); !ok {
		t.Error("expected FocusByID to succeed for input-a")
	}
	if e.FocusedNode() != a {
		t.Error("expected focused node to be input-a")
	}

	// Disabled node should not be focusable
	if ok := e.FocusByID("input-disabled"); ok {
		t.Error("expected FocusByID to fail for disabled node")
	}
	// Focus should remain on a
	if e.FocusedNode() != a {
		t.Error("expected focus to remain on input-a after failed FocusByID")
	}

	// Nonexistent ID
	if ok := e.FocusByID("nonexistent"); ok {
		t.Error("expected FocusByID to fail for nonexistent ID")
	}
}

func TestEngine_FocusableIDs(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	root := NewNode("box")
	root.W = 40
	root.H = 10

	a := NewNode("input")
	a.ID = "a"
	a.Focusable = true
	a.W = 10
	a.H = 1
	root.AddChild(a)

	b := NewNode("input")
	b.ID = "b"
	b.Focusable = true
	b.W = 10
	b.H = 1
	root.AddChild(b)

	// No ID — focusable but should be excluded from IDs list
	noID := NewNode("input")
	noID.Focusable = true
	noID.W = 10
	noID.H = 1
	root.AddChild(noID)

	c := NewNode("input")
	c.ID = "c"
	c.Focusable = true
	c.Disabled = true // disabled — not in focusable list
	c.W = 10
	c.H = 1
	root.AddChild(c)

	e.syncMainLayer()
	e.Layers()[0].Root = root

	ids := e.FocusableIDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 focusable IDs, got %d: %v", len(ids), ids)
	}
	if ids[0] != "a" || ids[1] != "b" {
		t.Errorf("expected [a, b], got %v", ids)
	}
}

func TestEngine_FocusNextAndPrev_Symmetry(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	root := NewNode("box")
	root.W = 40
	root.H = 10

	nodes := make([]*Node, 3)
	for i := range nodes {
		n := NewNode("input")
		n.ID = string(rune('a' + i))
		n.Focusable = true
		n.W = 10
		n.H = 1
		root.AddChild(n)
		nodes[i] = n
	}

	e.syncMainLayer()
	e.Layers()[0].Root = root

	// FocusNext 3 times should cycle: a → b → c
	e.FocusNext()
	if e.FocusedNode() != nodes[0] {
		t.Error("FocusNext: expected a")
	}
	e.FocusNext()
	if e.FocusedNode() != nodes[1] {
		t.Error("FocusNext: expected b")
	}
	e.FocusNext()
	if e.FocusedNode() != nodes[2] {
		t.Error("FocusNext: expected c")
	}

	// FocusPrev 3 times should cycle: b → a → c (wrap)
	e.FocusPrev()
	if e.FocusedNode() != nodes[1] {
		t.Error("FocusPrev: expected b")
	}
	e.FocusPrev()
	if e.FocusedNode() != nodes[0] {
		t.Error("FocusPrev: expected a")
	}
	e.FocusPrev()
	if e.FocusedNode() != nodes[2] {
		t.Error("FocusPrev: expected c (wrap)")
	}
}
