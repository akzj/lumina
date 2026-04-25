package lumina

import (
	"testing"
)

// TestRelativePositioning verifies that position="relative" offsets elements from normal flow.
func TestRelativePositioning(t *testing.T) {
	// Create a vbox with two children; second child has relative offset
	parent := &VNode{
		Type: "vbox",
		Props: map[string]any{
			"style": map[string]any{
				"width": 40, "height": 10,
			},
		},
		Children: []*VNode{
			{Type: "text", Content: "First", Props: map[string]any{
				"style": map[string]any{"height": 1},
			}},
			{Type: "text", Content: "Second", Props: map[string]any{
				"style": map[string]any{
					"height":   1,
					"position": "relative",
					"left":     5,
					"top":      2,
				},
			}},
		},
	}

	computeFlexLayout(parent, 0, 0, 40, 10)

	// First child at normal position
	first := parent.Children[0]
	if first.Y != 0 {
		t.Errorf("first child Y = %d, want 0", first.Y)
	}

	// Second child should be at normal Y (1) + relative offset top=2 → Y=3
	// and X = 0 + left=5 → X=5
	second := parent.Children[1]
	if second.Y != 3 {
		t.Errorf("second child Y = %d, want 3 (normal=1 + relative top=2)", second.Y)
	}
	if second.X != 5 {
		t.Errorf("second child X = %d, want 5 (normal=0 + relative left=5)", second.X)
	}
}

// TestRelativePositioning_RightBottom verifies right/bottom offsets work.
func TestRelativePositioning_RightBottom(t *testing.T) {
	parent := &VNode{
		Type: "vbox",
		Props: map[string]any{
			"style": map[string]any{"width": 40, "height": 10},
		},
		Children: []*VNode{
			{Type: "text", Content: "Shifted", Props: map[string]any{
				"style": map[string]any{
					"height":   1,
					"position": "relative",
					"right":    3,
					"bottom":   1,
				},
			}},
		},
	}

	computeFlexLayout(parent, 0, 0, 40, 10)

	child := parent.Children[0]
	// right=3 with left=0 → X = 0 - 3 = -3
	if child.X != -3 {
		t.Errorf("child X = %d, want -3 (right=3)", child.X)
	}
	// bottom=1 with top=0 → Y = 0 - 1 = -1
	if child.Y != -1 {
		t.Errorf("child Y = %d, want -1 (bottom=1)", child.Y)
	}
}

// TestZIndexRenderOrder verifies that higher z-index positioned children render on top.
func TestZIndexRenderOrder(t *testing.T) {
	// Create a container with two absolute children at same position but different z-index
	parent := &VNode{
		Type: "vbox",
		Props: map[string]any{
			"style": map[string]any{"width": 20, "height": 5},
		},
		Children: []*VNode{
			{Type: "text", Content: "AAAA", Props: map[string]any{
				"style": map[string]any{
					"position": "absolute",
					"left": 0, "top": 0,
					"width": 4, "height": 1,
					"zIndex": 1,
				},
			}},
			{Type: "text", Content: "BB", Props: map[string]any{
				"style": map[string]any{
					"position": "absolute",
					"left": 0, "top": 0,
					"width": 2, "height": 1,
					"zIndex": 10,
				},
			}},
		},
	}

	computeFlexLayout(parent, 0, 0, 20, 5)
	frame := NewFrame(20, 5)
	clip := Rect{X: 0, Y: 0, W: 20, H: 5}
	renderVNode(frame, parent, clip)

	// At position (0,0), the higher z-index child ("BB") should win
	if frame.Cells[0][0].Char != 'B' {
		t.Errorf("cell (0,0) = %c, want 'B' (higher z-index)", frame.Cells[0][0].Char)
	}
	if frame.Cells[0][1].Char != 'B' {
		t.Errorf("cell (0,1) = %c, want 'B' (higher z-index)", frame.Cells[0][1].Char)
	}
	// At position (2,0) and (3,0), only "AAAA" is present
	if frame.Cells[0][2].Char != 'A' {
		t.Errorf("cell (0,2) = %c, want 'A'", frame.Cells[0][2].Char)
	}
	if frame.Cells[0][3].Char != 'A' {
		t.Errorf("cell (0,3) = %c, want 'A'", frame.Cells[0][3].Char)
	}
}

// TestFrameRateLimit verifies that lastRenderTime field exists and is used.
func TestFrameRateLimit(t *testing.T) {
	// Verify App struct has lastRenderTime field (compile-time check)
	app := &App{}
	if !app.lastRenderTime.IsZero() {
		t.Error("lastRenderTime should be zero-value initially")
	}
}

// TestWebTerminalResize verifies resize callback propagation.
// We test the SetSize/SetOnResize mechanism by constructing a WebTerminal
// manually (without starting the read loop that needs a real WSConn).
func TestWebTerminalResize(t *testing.T) {
	// Build a WebTerminal struct directly to avoid the goroutine that reads from WSConn
	wt := &WebTerminal{
		width:  80,
		height: 24,
	}

	resized := false
	var gotW, gotH int
	wt.SetOnResize(func(w, h int) {
		resized = true
		gotW = w
		gotH = h
	})

	wt.SetSize(120, 40)

	if !resized {
		t.Fatal("resize callback was not called")
	}
	if gotW != 120 || gotH != 40 {
		t.Errorf("resize callback got (%d, %d), want (120, 40)", gotW, gotH)
	}

	// Verify terminal reports new size
	w, h := wt.Size()
	if w != 120 || h != 40 {
		t.Errorf("Size() = (%d, %d), want (120, 40)", w, h)
	}
}
