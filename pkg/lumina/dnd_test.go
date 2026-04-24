package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestDragStateCreation(t *testing.T) {
	ds := &DragState{}
	if ds.Dragging() {
		t.Fatal("should not be dragging initially")
	}
}

func TestDragStateStartDrag(t *testing.T) {
	ds := &DragState{}
	ds.StartDrag("card-1", "card", map[string]any{"id": 1})
	if !ds.Dragging() {
		t.Fatal("should be dragging after StartDrag")
	}
	if ds.GetDragType() != "card" {
		t.Fatalf("expected type 'card', got '%s'", ds.GetDragType())
	}
}

func TestDragStateEndDrag(t *testing.T) {
	ds := &DragState{}
	ds.StartDrag("card-1", "card", "data123")
	ds.SetDropTarget("zone-1")
	src, tgt, data := ds.EndDrag()
	if src != "card-1" {
		t.Fatalf("expected source 'card-1', got '%s'", src)
	}
	if tgt != "zone-1" {
		t.Fatalf("expected target 'zone-1', got '%s'", tgt)
	}
	if data != "data123" {
		t.Fatalf("expected data 'data123', got '%v'", data)
	}
	if ds.Dragging() {
		t.Fatal("should not be dragging after EndDrag")
	}
}

func TestDropZoneAccept(t *testing.T) {
	dz := &DropZone{Accept: []string{"card", "task"}}
	if !dz.CanDrop("card") {
		t.Fatal("should accept 'card'")
	}
	if !dz.CanDrop("task") {
		t.Fatal("should accept 'task'")
	}
	if dz.CanDrop("image") {
		t.Fatal("should reject 'image'")
	}
}

func TestDropZoneRejectInvalid(t *testing.T) {
	dz := &DropZone{Accept: []string{"file"}}
	if dz.CanDrop("card") {
		t.Fatal("should reject 'card' when only 'file' accepted")
	}
}

func TestDropZoneDisabled(t *testing.T) {
	dz := &DropZone{Accept: []string{"card"}, Disabled: true}
	if dz.CanDrop("card") {
		t.Fatal("disabled drop zone should reject everything")
	}
}

func TestDragPosition(t *testing.T) {
	ds := &DragState{}
	ds.StartDrag("item-1", "item", nil)
	ds.UpdatePosition(10, 20)
	ds.mu.Lock()
	x, y := ds.PositionX, ds.PositionY
	ds.mu.Unlock()
	if x != 10 || y != 20 {
		t.Fatalf("expected (10,20), got (%d,%d)", x, y)
	}
}

func TestDropZoneAcceptAll(t *testing.T) {
	dz := &DropZone{} // no Accept list = accept all
	if !dz.CanDrop("anything") {
		t.Fatal("empty accept list should accept all types")
	}
}

func TestLuaDragDropAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalDragState.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local dropped = false
		local droppedData = nil

		local drag = lumina.useDrag({
			type = "card",
			data = { id = 42, title = "Task" },
		})

		local drop = lumina.useDrop({
			accept = { "card" },
			onDrop = function(data)
				dropped = true
				droppedData = data
			end,
		})

		assert(drag.isDragging() == false, "should not be dragging initially")
		assert(drop.canDrop() == false, "can't drop when not dragging")

		-- Start drag
		drag.start("card-42")
		assert(drag.isDragging() == true, "should be dragging")
		assert(drop.canDrop() == true, "should accept card type")

		-- Drop
		local ok = drop.drop()
		assert(ok == true, "drop should succeed")
		assert(dropped == true, "onDrop should have been called")
		assert(drag.isDragging() == false, "should not be dragging after drop")
	`)
	if err != nil {
		t.Fatalf("Lua DnD: %v", err)
	}
	globalDragState.Reset()
}
