package lumina

import (
	"sync"
)

// DragState tracks the current drag operation.
type DragState struct {
	mu         sync.Mutex
	IsDragging bool
	DragSource string // ID of dragged element
	DragType   string // type of drag (for accept filtering)
	DragData   any    // data attached to drag
	DropTarget string // ID of current drop target
	PositionX  int    // current drag position X
	PositionY  int    // current drag position Y
}

var globalDragState = &DragState{}

// GetDragState returns the global drag state.
func GetDragState() *DragState {
	return globalDragState
}

// StartDrag begins a drag operation.
func (ds *DragState) StartDrag(sourceID, dragType string, data any) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.IsDragging = true
	ds.DragSource = sourceID
	ds.DragType = dragType
	ds.DragData = data
	ds.DropTarget = ""
}

// UpdatePosition updates the drag position.
func (ds *DragState) UpdatePosition(x, y int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.PositionX = x
	ds.PositionY = y
}

// SetDropTarget sets the current drop target.
func (ds *DragState) SetDropTarget(targetID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.DropTarget = targetID
}

// EndDrag ends the drag operation and returns the drop result.
func (ds *DragState) EndDrag() (sourceID, targetID string, data any) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	sourceID = ds.DragSource
	targetID = ds.DropTarget
	data = ds.DragData
	ds.IsDragging = false
	ds.DragSource = ""
	ds.DragType = ""
	ds.DragData = nil
	ds.DropTarget = ""
	ds.PositionX = 0
	ds.PositionY = 0
	return
}

// Dragging returns whether a drag is in progress.
func (ds *DragState) Dragging() bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return ds.IsDragging
}

// GetDragType returns the current drag type.
func (ds *DragState) GetDragType() string {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return ds.DragType
}

// GetDragData returns the current drag data.
func (ds *DragState) GetDragData() any {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return ds.DragData
}

// Reset clears drag state (for testing).
func (ds *DragState) Reset() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.IsDragging = false
	ds.DragSource = ""
	ds.DragType = ""
	ds.DragData = nil
	ds.DropTarget = ""
	ds.PositionX = 0
	ds.PositionY = 0
}

// Draggable represents a draggable element configuration.
type Draggable struct {
	ID       string
	Type     string
	Data     any
	Disabled bool
}

// DropZone represents a drop target configuration.
type DropZone struct {
	ID       string
	Accept   []string // accepted drag types
	Disabled bool
	OnDrop   func(data any)
}

// CanDrop checks if the drop zone accepts the given drag type.
func (dz *DropZone) CanDrop(dragType string) bool {
	if dz.Disabled {
		return false
	}
	if len(dz.Accept) == 0 {
		return true // accept all
	}
	for _, a := range dz.Accept {
		if a == dragType {
			return true
		}
	}
	return false
}
