package lumina

import (
	"encoding/json"

	"github.com/akzj/go-lua/pkg/lua"
)

// DiffResult holds the diff between two frames.
type DiffResult struct {
	Before *Frame   `json:"before,omitempty"`
	After  *Frame   `json:"after,omitempty"`
	Patches []DiffPatch `json:"patches"`
	Stats  DiffStats    `json:"stats"`
}

// DiffPatch represents a single cell change.
type DiffPatch struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Before *Cell  `json:"before,omitempty"`
	After  *Cell  `json:"after,omitempty"`
	Op     string `json:"op"` // "set", "clear"
}

// DiffStats holds diff statistics.
type DiffStats struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
	Changed int `json:"changed"`
	Unchanged int `json:"unchanged"`
}

// DiffFrames compares two frames and returns the differences.
func DiffFrames(before, after *Frame) *DiffResult {
	result := &DiffResult{
		Patches: make([]DiffPatch, 0),
		Stats:   DiffStats{},
	}

	if before == nil || after == nil {
		return result
	}

	maxY := len(before.Cells)
	maxX := 0
	if len(after.Cells) < maxY {
		maxY = len(after.Cells)
	}

	for y := 0; y < maxY; y++ {
		if len(before.Cells[y]) > maxX {
			maxX = len(before.Cells[y])
		}
	}

	// Compare cells
	for y := 0; y < maxY; y++ {
		rowMaxX := len(before.Cells[y])
		if len(after.Cells[y]) < rowMaxX {
			rowMaxX = len(after.Cells[y])
		}

		for x := 0; x < rowMaxX; x++ {
			beforeCell := before.Cells[y][x]
			afterCell := after.Cells[y][x]

			if !cellsEqual(beforeCell, afterCell) {
				result.Patches = append(result.Patches, DiffPatch{
					X:      x,
					Y:      y,
					Before: &beforeCell,
					After:  &afterCell,
					Op:     "set",
				})
				result.Stats.Changed++
			} else {
				result.Stats.Unchanged++
			}
		}

		// Cells only in after
		if y < len(after.Cells) && len(after.Cells[y]) > len(before.Cells[y]) {
			for x := len(before.Cells[y]); x < len(after.Cells[y]); x++ {
				afterCell := after.Cells[y][x]
				result.Patches = append(result.Patches, DiffPatch{
					X:     x,
					Y:     y,
					After: &afterCell,
					Op:    "set",
				})
				result.Stats.Added++
			}
		}
	}

	return result
}

// cellsEqual compares two cells.
func cellsEqual(a, b Cell) bool {
	return a.Char == b.Char &&
		a.Foreground == b.Foreground &&
		a.Background == b.Background &&
		a.Bold == b.Bold &&
		a.Dim == b.Dim
}

// GetFrameHistory returns recent frames for diff.
func GetFrameHistoryN(n int) []*Frame {
	history := GetFrameHistory()
	if len(history) <= n {
		return history
	}
	return history[len(history)-n:]
}

// DiffLastFrames returns diff between the last two frames.
func DiffLastFrames() *DiffResult {
	history := GetFrameHistory()
	if len(history) < 2 {
		return &DiffResult{Patches: []DiffPatch{}, Stats: DiffStats{}}
	}
	before := history[len(history)-2]
	after := history[len(history)-1]
	return DiffFrames(before, after)
}

// Lua API

// diff() → JSON string of diff between last two frames
func diff(L *lua.State) int {
	result := DiffLastFrames()
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}

// diffFrames(n) → JSON string of diff between frames[len-n] and frames[len-1]
func diffFrames(L *lua.State) int {
	n := 2
	if L.GetTop() >= 1 {
		if c, ok := L.ToInteger(1); ok {
			n = int(c)
		}
		if n < 2 {
			n = 2
		}
		if n > 100 {
			n = 100
		}
	}

	history := GetFrameHistory()
	if len(history) < n {
		L.PushString(`{"error": "not enough frames"}`)
		return 1
	}

	before := history[len(history)-n]
	after := history[len(history)-1]
	result := DiffFrames(before, after)

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}
