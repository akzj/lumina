package lumina

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// JSONAdapter implements OutputAdapter for JSON/MCP output.
type JSONAdapter struct {
	writer  io.Writer
	mu      sync.Mutex
	verbose bool // Full frame output for debugging
}

// NewJSONAdapter creates a new JSON output adapter.
func NewJSONAdapter(w io.Writer) *JSONAdapter {
	if w == nil {
		w = os.Stdout
	}
	return &JSONAdapter{writer: w}
}

// Write outputs a frame as JSON MCP format.
func (a *JSONAdapter) Write(frame *Frame) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Convert Frame to MCP format
	mcpFrame := a.toMCPFrame(frame)

	output, err := json.Marshal(mcpFrame)
	if err != nil {
		return err
	}

	_, err = a.writer.Write(append(output, '\n'))
	return err
}

// ToMCPFrame converts internal Frame to proto.Frame (for backwards compatibility).
func ToMCPFrame(frame *Frame) *ProtoFrame {
	pf := &ProtoFrame{
		Timestamp: frame.Timestamp,
	}

	// Convert cells to proto format
	for _, row := range frame.Cells {
		protoRow := make([]ProtoCell, len(row))
		for i, cell := range row {
			protoRow[i] = ProtoCell{
				Char:       string(cell.Char),
				Foreground: cell.Foreground,
				Background: cell.Background,
				Bold:       cell.Bold,
				Dim:        cell.Dim,
				Underline:  cell.Underline,
			}
		}
		pf.Cells = append(pf.Cells, protoRow)
	}

	return pf
}

// ProtoFrame mirrors proto.Frame for internal use.
type ProtoFrame struct {
	Timestamp int64
	Cells     [][]ProtoCell
}

// ProtoCell mirrors proto.Cell for internal use.
type ProtoCell struct {
	Char       string
	Foreground string
	Background string
	Bold       bool
	Dim        bool
	Underline  bool
}

// toMCPFrame converts a Frame to MCP protocol format.
func (a *JSONAdapter) toMCPFrame(frame *Frame) map[string]any {
	result := map[string]any{
		"type":      "frame",
		"timestamp": frame.Timestamp,
	}

	if a.verbose {
		// Full frame output for debugging
		cells := make([][]map[string]any, len(frame.Cells))
		for y, row := range frame.Cells {
			cells[y] = make([]map[string]any, len(row))
			for x, cell := range row {
				cells[y][x] = map[string]any{
					"char": string(cell.Char),
					"fg":   cell.Foreground,
					"bg":   cell.Background,
					"bold": cell.Bold,
					"dim":  cell.Dim,
				}
			}
		}
		result["cells"] = cells
		result["width"] = frame.Width
		result["height"] = frame.Height
	} else {
		// Patch-based output for efficiency
		patches := make([]map[string]any, 0)
		for _, rect := range frame.DirtyRects {
			for y := rect.Y; y < rect.Y+rect.H && y < len(frame.Cells); y++ {
				for x := rect.X; x < rect.X+rect.W && x < len(frame.Cells[y]); x++ {
					cell := frame.Cells[y][x]
					patches = append(patches, map[string]any{
						"op": "set",
						"x":  x,
						"y":  y,
						"c": map[string]any{
							"char": string(cell.Char),
							"fg":   cell.Foreground,
							"bg":   cell.Background,
							"bold": cell.Bold,
						},
					})
				}
			}
		}
		result["patches"] = patches
	}

	return result
}

// Flush implements OutputAdapter.
func (a *JSONAdapter) Flush() error {
	return nil
}

// Close implements OutputAdapter.
func (a *JSONAdapter) Close() error {
	return nil
}

// Mode implements OutputAdapter.
func (a *JSONAdapter) Mode() OutputMode {
	return ModeJSON
}

// SetVerbose sets the verbose mode.
func (a *JSONAdapter) SetVerbose(v bool) {
	a.verbose = v
}

// JSONFrame represents a frame in JSON format.
type JSONFrame struct {
	Type      string             `json:"type"`
	Timestamp int64              `json:"timestamp"`
	Patches   []map[string]any   `json:"patches,omitempty"`
	Cells     [][]map[string]any `json:"cells,omitempty"`
	Width     int                `json:"width,omitempty"`
	Height    int                `json:"height,omitempty"`
}

// ToJSON converts a Frame to JSON bytes.
func (f *Frame) ToJSON() ([]byte, error) {
	adapter := &JSONAdapter{verbose: true}
	return json.Marshal(adapter.toMCPFrame(f))
}

// ToJSONCompact converts a Frame to compact JSON bytes (patch-based).
func (f *Frame) ToJSONCompact() ([]byte, error) {
	adapter := &JSONAdapter{verbose: false}
	return json.Marshal(adapter.toMCPFrame(f))
}

// RecordJSONFrame records a frame for MCP inspection.
func RecordJSONFrame(frame *Frame) {
	// Record frame for later inspection/diff
	RecordFrame(frame)
}

// WriteJSONFrame writes a frame as JSON to a writer.
func WriteJSONFrame(w io.Writer, frame *Frame) error {
	return NewJSONAdapter(w).Write(frame)
}

// CurrentTimestamp returns current time in milliseconds.
func CurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}
