package lumina

import (
	"encoding/json"
	"os"

	"github.com/akzj/lumina/pkg/lumina/proto"
)

// JSON encoder for MCP output
var jsonEncoder = json.NewEncoder(os.Stdout)

// ToMCPFrame converts internal Frame to proto.Frame.
func ToMCPFrame(frame *Frame) *proto.Frame {
	pf := &proto.Frame{
		Timestamp: frame.Timestamp,
	}

	// Convert cells to proto format
	for _, row := range frame.Cells {
		protoRow := make([]proto.Cell, len(row))
		for i, cell := range row {
			protoRow[i] = proto.Cell{
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

// WriteJSONFrame writes a frame as JSON to stdout.
func WriteJSONFrame(frame *Frame) error {
	pf := ToMCPFrame(frame)
	return jsonEncoder.Encode(pf)
}

// IsJSONMode returns true if output mode is JSON.
func IsJSONMode() bool {
	return GetOutputMode() == ModeJSON
}
