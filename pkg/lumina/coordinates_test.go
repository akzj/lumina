package lumina

import (
	"testing"
)

// TestHitTestFrame_CellBased verifies O(1) hit-test via Cell.OwnerNode.
func TestHitTestFrame_CellBased(t *testing.T) {
	frame := NewFrame(10, 5)

	// Create a VNode and assign it as owner of some cells
	vnode := NewVNode("box")
	vnode.Props["id"] = "target-box"
	vnode.X = 2
	vnode.Y = 1
	vnode.W = 4
	vnode.H = 3

	// Simulate rendering: mark cells as owned by vnode
	for y := 1; y < 4; y++ {
		for x := 2; x < 6; x++ {
			frame.Cells[y][x].OwnerNode = vnode
			frame.Cells[y][x].Char = '#'
			frame.Cells[y][x].Transparent = false
		}
	}

	// Hit inside the box
	result := HitTestFrame(frame, 3, 2)
	if result != vnode {
		t.Errorf("expected vnode at (3,2), got %v", result)
	}

	// Hit outside the box
	result = HitTestFrame(frame, 0, 0)
	if result != nil {
		t.Errorf("expected nil at (0,0), got %v", result)
	}

	// Hit out of bounds
	result = HitTestFrame(frame, -1, 0)
	if result != nil {
		t.Error("expected nil for negative x")
	}
	result = HitTestFrame(frame, 10, 0)
	if result != nil {
		t.Error("expected nil for x >= width")
	}

	// Nil frame
	result = HitTestFrame(nil, 0, 0)
	if result != nil {
		t.Error("expected nil for nil frame")
	}
}

// TestGlobalToLocal verifies coordinate transform from global to local.
func TestGlobalToLocal(t *testing.T) {
	vnode := NewVNode("box")
	vnode.X = 10
	vnode.Y = 5
	vnode.W = 20
	vnode.H = 10
	vnode.Style = Style{Border: "single", PaddingLeft: 1, PaddingTop: 1}

	// Global (12, 7) → local (0, 0) because:
	// border=1, paddingLeft=1 → content starts at X=12
	// border=1, paddingTop=1 → content starts at Y=7
	lx, ly := GlobalToLocal(vnode, 12, 7)
	if lx != 0 || ly != 0 {
		t.Errorf("GlobalToLocal(12,7) = (%d,%d), want (0,0)", lx, ly)
	}

	// Global (15, 9) → local (3, 2)
	lx, ly = GlobalToLocal(vnode, 15, 9)
	if lx != 3 || ly != 2 {
		t.Errorf("GlobalToLocal(15,9) = (%d,%d), want (3,2)", lx, ly)
	}

	// No border, no padding
	simple := NewVNode("box")
	simple.X = 5
	simple.Y = 3
	simple.W = 10
	simple.H = 5
	simple.Style = Style{}

	lx, ly = GlobalToLocal(simple, 7, 4)
	if lx != 2 || ly != 1 {
		t.Errorf("GlobalToLocal(7,4) on simple = (%d,%d), want (2,1)", lx, ly)
	}
}

// TestLocalToGlobal verifies coordinate transform from local to global.
func TestLocalToGlobal(t *testing.T) {
	vnode := NewVNode("box")
	vnode.X = 10
	vnode.Y = 5
	vnode.W = 20
	vnode.H = 10
	vnode.Style = Style{Border: "single", PaddingLeft: 1, PaddingTop: 1}

	// Local (0, 0) → global (12, 7)
	gx, gy := LocalToGlobal(vnode, 0, 0)
	if gx != 12 || gy != 7 {
		t.Errorf("LocalToGlobal(0,0) = (%d,%d), want (12,7)", gx, gy)
	}

	// Round-trip: global → local → global
	origX, origY := 15, 9
	lx, ly := GlobalToLocal(vnode, origX, origY)
	gx, gy = LocalToGlobal(vnode, lx, ly)
	if gx != origX || gy != origY {
		t.Errorf("round-trip failed: (%d,%d) → (%d,%d) → (%d,%d)", origX, origY, lx, ly, gx, gy)
	}
}

// TestContainsPoint verifies point-in-rect check.
func TestContainsPoint(t *testing.T) {
	vnode := NewVNode("box")
	vnode.X = 5
	vnode.Y = 3
	vnode.W = 10
	vnode.H = 5

	tests := []struct {
		x, y int
		want bool
	}{
		{5, 3, true},   // top-left corner
		{14, 7, true},  // bottom-right corner
		{10, 5, true},  // middle
		{4, 3, false},  // just left
		{15, 3, false}, // just right
		{5, 2, false},  // just above
		{5, 8, false},  // just below
	}

	for _, tt := range tests {
		got := ContainsPoint(vnode, tt.x, tt.y)
		if got != tt.want {
			t.Errorf("ContainsPoint(%d,%d) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

// TestContainsPointContent verifies point-in-content-area check.
func TestContainsPointContent(t *testing.T) {
	vnode := NewVNode("box")
	vnode.X = 0
	vnode.Y = 0
	vnode.W = 20
	vnode.H = 10
	vnode.Style = Style{Border: "single", PaddingLeft: 1, PaddingTop: 1, PaddingRight: 1, PaddingBottom: 1}

	// Content area: X=2, Y=2, W=16, H=6
	// (border=1 + paddingLeft=1 = 2, border=1 + paddingTop=1 = 2)
	// W = 20 - 2*1 - 1 - 1 = 16, H = 10 - 2*1 - 1 - 1 = 6

	if !ContainsPointContent(vnode, 2, 2) {
		t.Error("(2,2) should be inside content area")
	}
	if !ContainsPointContent(vnode, 17, 7) {
		t.Error("(17,7) should be inside content area")
	}
	if ContainsPointContent(vnode, 1, 2) {
		t.Error("(1,2) should be outside content area (in padding)")
	}
	if ContainsPointContent(vnode, 0, 0) {
		t.Error("(0,0) should be outside content area (in border)")
	}
}

// TestLocalCoordinates_InEvent verifies that events get correct LocalX/LocalY.
func TestLocalCoordinates_InEvent(t *testing.T) {
	// Create a frame with a VNode owning some cells
	frame := NewFrame(80, 24)

	vnode := NewVNode("box")
	vnode.Props["id"] = "test-target"
	vnode.X = 10
	vnode.Y = 5
	vnode.W = 20
	vnode.H = 10

	// Mark cells as owned
	for y := 5; y < 15; y++ {
		for x := 10; x < 30; x++ {
			if x < 80 && y < 24 {
				frame.Cells[y][x].OwnerNode = vnode
				frame.Cells[y][x].Transparent = false
			}
		}
	}

	// Simulate hit-test at (15, 8)
	targetNode := HitTestFrame(frame, 15, 8)
	if targetNode == nil {
		t.Fatal("expected non-nil target node")
	}

	// Create event with local coordinates
	e := &Event{
		Type: "mousedown",
		X:    15,
		Y:    8,
	}
	e.TargetNode = targetNode
	e.LocalX = e.X - targetNode.X
	e.LocalY = e.Y - targetNode.Y

	if e.LocalX != 5 {
		t.Errorf("LocalX = %d, want 5", e.LocalX)
	}
	if e.LocalY != 3 {
		t.Errorf("LocalY = %d, want 3", e.LocalY)
	}
}

// TestDragAndDrop_StateSequence tests the DnD state machine.
func TestDragAndDrop_StateSequence(t *testing.T) {
	ds := &DragState{}

	// Initially not dragging
	if ds.Dragging() {
		t.Error("should not be dragging initially")
	}

	// Start drag
	ds.StartDrag("source-1", "file", "test-data")
	if !ds.Dragging() {
		t.Error("should be dragging after StartDrag")
	}
	if ds.GetDragType() != "file" {
		t.Errorf("drag type = %q, want %q", ds.GetDragType(), "file")
	}
	if ds.GetDragData() != "test-data" {
		t.Errorf("drag data = %v, want %q", ds.GetDragData(), "test-data")
	}

	// Update position
	ds.UpdatePosition(50, 25)
	ds.mu.Lock()
	px, py := ds.PositionX, ds.PositionY
	ds.mu.Unlock()
	if px != 50 || py != 25 {
		t.Errorf("position = (%d,%d), want (50,25)", px, py)
	}

	// Set drop target
	ds.SetDropTarget("drop-zone-1")

	// End drag
	sourceID, targetID, data := ds.EndDrag()
	if sourceID != "source-1" {
		t.Errorf("source = %q, want %q", sourceID, "source-1")
	}
	if targetID != "drop-zone-1" {
		t.Errorf("target = %q, want %q", targetID, "drop-zone-1")
	}
	if data != "test-data" {
		t.Errorf("data = %v, want %q", data, "test-data")
	}

	// After end, not dragging
	if ds.Dragging() {
		t.Error("should not be dragging after EndDrag")
	}
}
