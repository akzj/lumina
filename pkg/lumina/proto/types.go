package proto

import "time"

// Frame represents a complete or partial terminal frame for MCP protocol.
type Frame struct {
	Timestamp     int64    `json:"timestamp"`
	Patches       []Patch  `json:"patches,omitempty"`
	Cells         [][]Cell `json:"cells,omitempty"`
	Events        []string `json:"events,omitempty"`
	ComponentTree string   `json:"component_tree,omitempty"`
}

// Patch represents a single cell update.
type Patch struct {
	Op          PatchOp `json:"op"`
	X           int     `json:"x"`
	Y           int     `json:"y"`
	Cell        Cell    `json:"c"`
	ComponentID string  `json:"component_id,omitempty"`
}

// PatchOp is the patch operation type.
type PatchOp int

const (
	PatchSet   PatchOp = 0
	PatchClear PatchOp = 1
)

// Cell represents a terminal cell.
type Cell struct {
	Char       string `json:"char"`
	Foreground string `json:"fg"`
	Background string `json:"bg"`
	Bold       bool   `json:"bold,omitempty"`
	Dim        bool   `json:"dim,omitempty"`
	Underline  bool   `json:"underline,omitempty"`
}

// ComponentRequest creates or updates a component.
type ComponentRequest struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Props    map[string]string `json:"props"`
	ParentID string            `json:"parent_id,omitempty"`
}

// ComponentResponse is the response to a component request.
type ComponentResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// EventNotification is sent when a component fires an event.
type EventNotification struct {
	ComponentID string            `json:"component_id"`
	EventType   string            `json:"event_type"`
	EventData   map[string]string `json:"event_data"`
	Timestamp   int64             `json:"timestamp"`
}

// MCPRequest wraps all MCP requests.
type MCPRequest struct {
	Method    string            `json:"method"`
	RequestID string            `json:"request_id"`
	Component *ComponentRequest `json:"component,omitempty"`
	Params    map[string]string `json:"params,omitempty"`
}

// MCPResponse wraps all MCP responses.
type MCPResponse struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Data      []byte `json:"data,omitempty"`
}

// NewFrame creates a new frame with current timestamp.
func NewFrame() *Frame {
	return &Frame{
		Timestamp: time.Now().UnixMilli(),
		Patches:   make([]Patch, 0),
		Cells:     make([][]Cell, 0),
	}
}

// AddPatch adds a patch to the frame.
func (f *Frame) AddPatch(x, y int, cell Cell) {
	f.Patches = append(f.Patches, Patch{
		Op:   PatchSet,
		X:    x,
		Y:    y,
		Cell: cell,
	})
}

// NewEventNotification creates an event notification with current timestamp.
func NewEventNotification(compID, eventType string, data map[string]string) *EventNotification {
	return &EventNotification{
		ComponentID: compID,
		EventType:   eventType,
		EventData:   data,
		Timestamp:   time.Now().UnixMilli(),
	}
}

// NewMCPResponse creates a response for a request.
func NewMCPResponse(requestID string, success bool, errMsg string) *MCPResponse {
	return &MCPResponse{
		RequestID: requestID,
		Success:   success,
		Error:     errMsg,
	}
}
