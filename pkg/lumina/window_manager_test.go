package lumina

import (
	"testing"
)

func TestWindowManager_CreateClose(t *testing.T) {
	wm := NewWindowManager(80, 24)

	win := wm.CreateWindow("w1", "Test Window", 10, 5, 30, 15)
	if win == nil {
		t.Fatal("CreateWindow returned nil")
	}
	if win.ID != "w1" {
		t.Errorf("ID = %q, want w1", win.ID)
	}
	if wm.Count() != 1 {
		t.Errorf("Count = %d, want 1", wm.Count())
	}

	// Window should be visible and focused
	visible := wm.GetVisible()
	if len(visible) != 1 {
		t.Fatalf("GetVisible = %d, want 1", len(visible))
	}
	if !visible[0].Focused {
		t.Error("new window should be focused")
	}

	// Close it
	wm.CloseWindow("w1")
	if wm.Count() != 0 {
		t.Errorf("Count after close = %d, want 0", wm.Count())
	}
	if len(wm.GetVisible()) != 0 {
		t.Error("GetVisible should be empty after close")
	}
}

func TestWindowManager_Focus(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Window 1", 0, 0, 20, 10)
	wm.CreateWindow("w2", "Window 2", 10, 5, 20, 10)

	// w2 should be focused (last created)
	focused := wm.GetFocused()
	if focused == nil || focused.ID != "w2" {
		t.Errorf("focused = %v, want w2", focused)
	}

	// Focus w1 — should bring it to front
	wm.FocusWindow("w1")
	focused = wm.GetFocused()
	if focused == nil || focused.ID != "w1" {
		t.Errorf("after focus, focused = %v, want w1", focused)
	}

	// w1 should have higher z-index than w2
	w1 := wm.GetWindow("w1")
	w2 := wm.GetWindow("w2")
	if w1.ZIndex <= w2.ZIndex {
		t.Errorf("w1.ZIndex=%d should be > w2.ZIndex=%d", w1.ZIndex, w2.ZIndex)
	}

	// GetVisible should return both, w1 last (highest z-index)
	visible := wm.GetVisible()
	if len(visible) != 2 {
		t.Fatalf("GetVisible = %d, want 2", len(visible))
	}
	if visible[1].ID != "w1" {
		t.Errorf("last visible = %q, want w1 (highest z)", visible[1].ID)
	}
}

func TestWindowManager_Move(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Window", 10, 5, 30, 15)

	wm.MoveWindow("w1", 20, 10)
	win := wm.GetWindow("w1")
	if win.X != 20 || win.Y != 10 {
		t.Errorf("position = (%d, %d), want (20, 10)", win.X, win.Y)
	}
}

func TestWindowManager_Resize(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Window", 0, 0, 30, 15)

	wm.ResizeWindow("w1", 50, 20)
	win := wm.GetWindow("w1")
	if win.W != 50 || win.H != 20 {
		t.Errorf("size = (%d, %d), want (50, 20)", win.W, win.H)
	}

	// Resize below minimum
	wm.ResizeWindow("w1", 3, 2)
	win = wm.GetWindow("w1")
	if win.W != win.MinW || win.H != win.MinH {
		t.Errorf("size = (%d, %d), want min (%d, %d)", win.W, win.H, win.MinW, win.MinH)
	}
}

func TestWindowManager_MaximizeRestore(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Window", 10, 5, 30, 15)

	wm.MaximizeWindow("w1")
	win := wm.GetWindow("w1")
	if win.X != 0 || win.Y != 0 || win.W != 80 || win.H != 24 {
		t.Errorf("maximized = (%d,%d,%d,%d), want (0,0,80,24)", win.X, win.Y, win.W, win.H)
	}
	if !win.Maximized {
		t.Error("should be maximized")
	}

	// Restore
	wm.RestoreWindow("w1")
	win = wm.GetWindow("w1")
	if win.X != 10 || win.Y != 5 || win.W != 30 || win.H != 15 {
		t.Errorf("restored = (%d,%d,%d,%d), want (10,5,30,15)", win.X, win.Y, win.W, win.H)
	}
	if win.Maximized {
		t.Error("should not be maximized after restore")
	}
}

func TestWindowManager_Minimize(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Window 1", 0, 0, 20, 10)
	wm.CreateWindow("w2", "Window 2", 10, 5, 20, 10)

	// Minimize w2 — w1 should get focus
	wm.MinimizeWindow("w2")
	w2 := wm.GetWindow("w2")
	if !w2.Minimized {
		t.Error("w2 should be minimized")
	}

	// Only w1 should be visible
	visible := wm.GetVisible()
	if len(visible) != 1 || visible[0].ID != "w1" {
		t.Errorf("visible = %v, want [w1]", visible)
	}

	// Restore w2
	wm.RestoreWindow("w2")
	w2 = wm.GetWindow("w2")
	if w2.Minimized {
		t.Error("w2 should not be minimized after restore")
	}
	if len(wm.GetVisible()) != 2 {
		t.Error("both windows should be visible after restore")
	}
}

func TestWindowManager_Compositing(t *testing.T) {
	wm := NewWindowManager(80, 24)
	win := wm.CreateWindow("w1", "Hello", 5, 3, 20, 8)
	win.VNode = &VNode{
		Type:    "text",
		Content: "Window Content",
	}

	// Build the VNode for the window
	vnode := BuildWindowVNode(win)
	if vnode == nil {
		t.Fatal("BuildWindowVNode returned nil")
	}
	if vnode.Type != "vbox" {
		t.Errorf("root type = %q, want vbox", vnode.Type)
	}
	if len(vnode.Children) != 2 {
		t.Fatalf("children = %d, want 2 (title bar + content)", len(vnode.Children))
	}

	// Render the window VNode
	computeFlexLayout(vnode, 5, 3, 20, 8)
	frame := NewFrame(80, 24)
	clip := Rect{X: 5, Y: 3, W: 20, H: 8}
	renderVNode(frame, vnode, clip)

	// Check that content rendered somewhere in the window area
	hasContent := false
	for y := 3; y < 11; y++ {
		for x := 5; x < 25; x++ {
			if frame.Cells[y][x].Char != ' ' && frame.Cells[y][x].Char != 0 && !frame.Cells[y][x].Transparent {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}
	if !hasContent {
		t.Error("window area has no visible content")
	}
}

func TestWindowManager_TitleBarDrag(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Draggable", 10, 5, 30, 15)

	// Start drag from title bar at (15, 5) — offset from window = (5, 0)
	wm.StartDrag("w1", 5, 0)
	if !wm.IsDragging() {
		t.Fatal("should be dragging")
	}

	// Move cursor to (25, 10) — window should move to (20, 10)
	wm.UpdateDrag(25, 10)
	win := wm.GetWindow("w1")
	if win.X != 20 || win.Y != 10 {
		t.Errorf("after drag, position = (%d, %d), want (20, 10)", win.X, win.Y)
	}

	// Stop drag
	wm.StopDrag()
	if wm.IsDragging() {
		t.Error("should not be dragging after stop")
	}
}

func TestWindowManager_ResizeHandle(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Resizable", 0, 0, 30, 15)

	// Start resize from bottom-right corner
	wm.StartResize("w1", 29, 14)
	if !wm.IsResizing() {
		t.Fatal("should be resizing")
	}

	// Move cursor 10 right, 5 down → size should increase by same amount
	wm.UpdateResize(39, 19)
	win := wm.GetWindow("w1")
	if win.W != 40 || win.H != 20 {
		t.Errorf("after resize, size = (%d, %d), want (40, 20)", win.W, win.H)
	}

	wm.StopResize()
	if wm.IsResizing() {
		t.Error("should not be resizing after stop")
	}
}

func TestWindowManager_WindowAtPoint(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Back", 0, 0, 20, 10)
	wm.CreateWindow("w2", "Front", 5, 3, 20, 10)

	// Point (15, 5) is in w2 only
	win := wm.WindowAtPoint(15, 5)
	if win == nil || win.ID != "w2" {
		t.Errorf("at (15,5) = %v, want w2", win)
	}

	// Point (2, 1) is in w1 only
	win = wm.WindowAtPoint(2, 1)
	if win == nil || win.ID != "w1" {
		t.Errorf("at (2,1) = %v, want w1", win)
	}

	// Point (7, 5) is in both — should return w2 (higher z-index, front)
	win = wm.WindowAtPoint(7, 5)
	if win == nil || win.ID != "w2" {
		t.Errorf("at (7,5) = %v, want w2 (front)", win)
	}

	// Point (50, 20) is in neither
	win = wm.WindowAtPoint(50, 20)
	if win != nil {
		t.Errorf("at (50,20) = %v, want nil", win)
	}
}

func TestWindowManager_TileHorizontal(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Left", 0, 0, 20, 10)
	wm.CreateWindow("w2", "Right", 0, 0, 20, 10)

	wm.TileHorizontal()

	w1 := wm.GetWindow("w1")
	w2 := wm.GetWindow("w2")
	if w1.X != 0 || w1.W != 40 || w1.H != 24 {
		t.Errorf("w1 tile = (%d,%d,%d,%d), want (0,0,40,24)", w1.X, w1.Y, w1.W, w1.H)
	}
	if w2.X != 40 || w2.W != 40 || w2.H != 24 {
		t.Errorf("w2 tile = (%d,%d,%d,%d), want (40,0,40,24)", w2.X, w2.Y, w2.W, w2.H)
	}
}

func TestWindowManager_TileVertical(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "Top", 0, 0, 20, 10)
	wm.CreateWindow("w2", "Bottom", 0, 0, 20, 10)

	wm.TileVertical()

	w1 := wm.GetWindow("w1")
	w2 := wm.GetWindow("w2")
	if w1.Y != 0 || w1.H != 12 || w1.W != 80 {
		t.Errorf("w1 tile = (%d,%d,%d,%d), want (0,0,80,12)", w1.X, w1.Y, w1.W, w1.H)
	}
	if w2.Y != 12 || w2.H != 12 || w2.W != 80 {
		t.Errorf("w2 tile = (%d,%d,%d,%d), want (0,12,80,12)", w2.X, w2.Y, w2.W, w2.H)
	}
}

func TestWindowManager_TileGrid(t *testing.T) {
	wm := NewWindowManager(80, 24)
	wm.CreateWindow("w1", "TL", 0, 0, 20, 10)
	wm.CreateWindow("w2", "TR", 0, 0, 20, 10)
	wm.CreateWindow("w3", "BL", 0, 0, 20, 10)
	wm.CreateWindow("w4", "BR", 0, 0, 20, 10)

	wm.TileGrid()

	w1 := wm.GetWindow("w1")
	w2 := wm.GetWindow("w2")
	w3 := wm.GetWindow("w3")
	w4 := wm.GetWindow("w4")

	// 2x2 grid: each cell = 40x12
	if w1.X != 0 || w1.Y != 0 || w1.W != 40 || w1.H != 12 {
		t.Errorf("w1 = (%d,%d,%d,%d), want (0,0,40,12)", w1.X, w1.Y, w1.W, w1.H)
	}
	if w2.X != 40 || w2.Y != 0 || w2.W != 40 || w2.H != 12 {
		t.Errorf("w2 = (%d,%d,%d,%d), want (40,0,40,12)", w2.X, w2.Y, w2.W, w2.H)
	}
	if w3.X != 0 || w3.Y != 12 || w3.W != 40 || w3.H != 12 {
		t.Errorf("w3 = (%d,%d,%d,%d), want (0,12,40,12)", w3.X, w3.Y, w3.W, w3.H)
	}
	if w4.X != 40 || w4.Y != 12 || w4.W != 40 || w4.H != 12 {
		t.Errorf("w4 = (%d,%d,%d,%d), want (40,12,40,12)", w4.X, w4.Y, w4.W, w4.H)
	}
}
