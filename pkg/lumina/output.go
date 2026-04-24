// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"os"
	"sync"
)

// OutputMode represents the output format mode.
type OutputMode int

const (
	// ModeANSI outputs ANSI escape codes for terminal display.
	ModeANSI OutputMode = iota
	// ModeJSON outputs JSON for AI agent parsing.
	ModeJSON
)

// String returns the string representation of OutputMode.
func (m OutputMode) String() string {
	switch m {
	case ModeJSON:
		return "json"
	default:
		return "ansi"
	}
}

// OutputAdapter is the interface for rendering frames to an output device.
type OutputAdapter interface {
	// Write writes a frame to the output.
	Write(frame *Frame) error
	// Flush flushes any buffered output.
	Flush() error
	// Close closes the adapter and releases resources.
	Close() error
	// Mode returns the output mode.
	Mode() OutputMode
}

// Frame represents a complete rendering frame.
type Frame struct {
	// Cells is the 2D grid of terminal cells.
	Cells [][]Cell
	// DirtyRects specifies regions that have changed since last frame.
	DirtyRects []Rect
	// Timestamp is the time when this frame was created.
	Timestamp int64
	// Width is the terminal width in cells.
	Width int
	// Height is the terminal height in cells.
	Height int
}

// Cell represents a single character cell in the terminal.
type Cell struct {
	// Char is the character to display.
	Char rune
	// Foreground is the foreground color in "#rrggbb" format.
	Foreground string
	// Background is the background color in "#rrggbb" format.
	Background string
	// Bold indicates bold text.
	Bold bool
	// Dim indicates dimmed text.
	Dim bool
	// Underline indicates underlined text.
	Underline bool
	// Reverse indicates reversed colors.
	Reverse bool
	// Blink indicates blinking text.
	Blink bool
}

// Rect represents a rectangular region in the terminal.
type Rect struct {
	X, Y int // Top-left corner
	W, H int // Width and height
}

// String returns a string representation of the rect.
func (r Rect) String() string {
	return fmt.Sprintf("Rect{%d,%d %dx%d}", r.X, r.Y, r.W, r.H)
}

// Contains checks if a point is within the rect.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// NewFrame creates a new frame with the given dimensions.
func NewFrame(width, height int) *Frame {
	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
		// Initialize with empty cells
		for x := 0; x < width; x++ {
			cells[y][x] = Cell{Char: ' '}
		}
	}
	return &Frame{
		Cells:  cells,
		Width:  width,
		Height: height,
	}
}

// MarkDirty marks the entire frame as dirty.
func (f *Frame) MarkDirty() {
	f.DirtyRects = []Rect{{X: 0, Y: 0, W: f.Width, H: f.Height}}
}

// AddDirtyRect adds a region to the dirty rects list.
func (f *Frame) AddDirtyRect(x, y, w, h int) {
	f.DirtyRects = append(f.DirtyRects, Rect{X: x, Y: y, W: w, H: h})
}

// GetCell returns the cell at the given coordinates.
func (f *Frame) GetCell(x, y int) Cell {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return Cell{}
	}
	return f.Cells[y][x]
}

// SetCell sets the cell at the given coordinates.
func (f *Frame) SetCell(x, y int, cell Cell) {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return
	}
	f.Cells[y][x] = cell
}

// NewCell creates a new cell with the given character.
func NewCell(ch rune) Cell {
	return Cell{Char: ch}
}

// SetForeground sets the foreground color.
func (c *Cell) SetForeground(color string) *Cell {
	c.Foreground = color
	return c
}

// SetBackground sets the background color.
func (c *Cell) SetBackground(color string) *Cell {
	c.Background = color
	return c
}

// SetBold sets the bold attribute.
func (c *Cell) SetBold(bold bool) *Cell {
	c.Bold = bold
	return c
}

// SetDim sets the dim attribute.
func (c *Cell) SetDim(dim bool) *Cell {
	c.Dim = dim
	return c
}

// SetUnderline sets the underline attribute.
func (c *Cell) SetUnderline(underline bool) *Cell {
	c.Underline = underline
	return c
}

// SetOutputAdapter sets the global output adapter.
func SetOutputAdapter(adapter OutputAdapter) {
	outputMu.Lock()
	defer outputMu.Unlock()
	outputAdapter = adapter
}

// GetOutputAdapter returns the global output adapter.
func GetOutputAdapter() OutputAdapter {
	outputMu.Lock()
	defer outputMu.Unlock()
	if outputAdapter == nil {
		// Lazy initialization with ANSI adapter
		outputAdapter = NewANSIAdapter(os.Stdout)
	}
	return outputAdapter
}

// NopAdapter is an OutputAdapter that discards all output.
type NopAdapter struct{}

// Write discards the frame.
func (a *NopAdapter) Write(frame *Frame) error {
	return nil
}

// Flush does nothing.
func (a *NopAdapter) Flush() error {
	return nil
}

// Close does nothing.
func (a *NopAdapter) Close() error {
	return nil
}

// Mode returns ModeANSI.
func (a *NopAdapter) Mode() OutputMode {
	return ModeANSI
}

// Output mode management
var (
	currentOutputMode OutputMode = ModeANSI
	outputAdapter     OutputAdapter
	outputMu          sync.Mutex // protects outputAdapter
)

// SetOutputMode sets the global output mode and configures the adapter.
func SetOutputMode(mode OutputMode) {
	outputMu.Lock()
	defer outputMu.Unlock()

	currentOutputMode = mode

	switch mode {
	case ModeJSON:
		if outputAdapter == nil || outputAdapter.Mode() != ModeJSON {
			outputAdapter = NewJSONAdapter(os.Stdout)
		}
	case ModeANSI:
		if outputAdapter == nil || outputAdapter.Mode() != ModeANSI {
			outputAdapter = NewANSIAdapter(os.Stdout)
		}
	}
}

// GetOutputMode returns the current output mode.
func GetOutputMode() OutputMode {
	return currentOutputMode
}

// OutputModeFromString converts a string to OutputMode.
func OutputModeFromString(s string) OutputMode {
	switch s {
	case "json":
		return ModeJSON
	default:
		return ModeANSI
	}
}
