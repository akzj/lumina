// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"bytes"
	"fmt"
	"io"
)

// ANSIAdapter implements OutputAdapter for ANSI terminal output.
type ANSIAdapter struct {
	writer io.Writer
	buf    bytes.Buffer
	width  int
	height int
}

// NewANSIAdapter creates a new ANSIAdapter writing to the given writer.
func NewANSIAdapter(w io.Writer) *ANSIAdapter {
	return &ANSIAdapter{
		writer: w,
		buf:    bytes.Buffer{},
		width:  80, // default terminal width
		height: 24, // default terminal height
	}
}

// SetSize sets the terminal dimensions.
func (a *ANSIAdapter) SetSize(width, height int) {
	a.width = width
	a.height = height
}

// Write writes a frame to the terminal using ANSI escape codes.
func (a *ANSIAdapter) Write(frame *Frame) error {
	a.buf.Reset()

	if len(frame.DirtyRects) == 0 {
		// No dirty regions, nothing to write
		return nil
	}

	// For each dirty rect, output the cells
	for _, rect := range frame.DirtyRects {
		for y := rect.Y; y < rect.Y+rect.H && y < frame.Height; y++ {
			// Move to start of line
			a.buf.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, rect.X+1))

			// Output cells in this row
			var lastStyle string
			for x := rect.X; x < rect.X+rect.W && x < frame.Width; x++ {
				cell := frame.Cells[y][x]
				style := a.styleCodes(&cell)

				// Only emit style change if different from last
				if style != lastStyle {
					a.buf.WriteString(style)
					lastStyle = style
				}

				// Handle special characters
				switch cell.Char {
				case '\\':
					a.buf.WriteString("\\\\")
				case '\x1b':
					a.buf.WriteString("\\e")
				case '\r':
					a.buf.WriteString("\\r")
				case '\n':
					a.buf.WriteString("\\n")
				default:
					a.buf.WriteRune(cell.Char)
				}
			}

			// Clear to end of line and reset style
			a.buf.WriteString("\x1b[0m")
		}
	}

	// Move cursor to bottom-left (standard terminal position)
	a.buf.WriteString("\x1b[999;1H")

	// Flush to writer
	_, err := a.writer.Write(a.buf.Bytes())
	return err
}

// Flush flushes the buffered output.
func (a *ANSIAdapter) Flush() error {
	if f, ok := a.writer.(interface{ Flush() error }); ok {
		return f.Flush()
	}
	return nil
}

// Close closes the adapter.
func (a *ANSIAdapter) Close() error {
	// Write final position and reset style
	a.buf.Reset()
	a.buf.WriteString("\x1b[0m\x1b[999;1H")
	_, err := a.writer.Write(a.buf.Bytes())
	return err
}

// Mode returns ModeANSI.
func (a *ANSIAdapter) Mode() OutputMode {
	return ModeANSI
}

// styleCodes generates the ANSI escape codes for a cell's style.
func (a *ANSIAdapter) styleCodes(cell *Cell) string {
	codes := ""

	// Reset
	codes += "\x1b[0m"

	// Foreground color
	if cell.Foreground != "" {
		if len(cell.Foreground) == 7 && cell.Foreground[0] == '#' {
			// Parse hex color
			r := cell.Foreground[1:3]
			g := cell.Foreground[3:5]
			b := cell.Foreground[5:7]
			codes += fmt.Sprintf("\x1b[38;2;%s;%s;%sm", hexToDec(r), hexToDec(g), hexToDec(b))
		} else {
			// Named color
			codes += a.namedColor("fg", cell.Foreground)
		}
	}

	// Background color
	if cell.Background != "" {
		if len(cell.Background) == 7 && cell.Background[0] == '#' {
			r := cell.Background[1:3]
			g := cell.Background[3:5]
			b := cell.Background[5:7]
			codes += fmt.Sprintf("\x1b[48;2;%s;%s;%sm", hexToDec(r), hexToDec(g), hexToDec(b))
		} else {
			codes += a.namedColor("bg", cell.Background)
		}
	}

	// Text attributes
	if cell.Bold {
		codes += "\x1b[1m"
	}
	if cell.Dim {
		codes += "\x1b[2m"
	}
	if cell.Underline {
		codes += "\x1b[4m"
	}
	if cell.Reverse {
		codes += "\x1b[7m"
	}
	if cell.Blink {
		codes += "\x1b[5m"
	}

	return codes
}

// namedColor converts a named color to ANSI code.
func (a *ANSIAdapter) namedColor(prefix string, name string) string {
	colors := map[string]int{
		"black":   30,
		"red":     31,
		"green":   32,
		"yellow":  33,
		"blue":    34,
		"magenta": 35,
		"cyan":    36,
		"white":   37,
		"default": 39,
		// Bright variants
		"bright_black":   90,
		"bright_red":     91,
		"bright_green":   92,
		"bright_yellow":  93,
		"bright_blue":    94,
		"bright_magenta": 95,
		"bright_cyan":    96,
		"bright_white":   97,
	}

	if code, ok := colors[name]; ok {
		return fmt.Sprintf("\x1b[%dm", code)
	}

	// BG colors
	bgColors := map[string]int{
		"bg_black":   40,
		"bg_red":     41,
		"bg_green":   42,
		"bg_yellow":  43,
		"bg_blue":    44,
		"bg_magenta": 45,
		"bg_cyan":    46,
		"bg_white":   47,
		"bg_default": 49,
	}

	if code, ok := bgColors[name]; ok {
		return fmt.Sprintf("\x1b[%dm", code)
	}

	return ""
}

// hexToDec converts a hex byte to decimal string.
func hexToDec(hex string) string {
	var n int
	for _, c := range hex {
		n *= 16
		if c >= '0' && c <= '9' {
			n += int(c - '0')
		} else if c >= 'a' && c <= 'f' {
			n += int(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			n += int(c - 'A' + 10)
		}
	}
	return fmt.Sprintf("%d", n)
}

// WriteString writes a plain string to the terminal with ANSI formatting.
func (a *ANSIAdapter) WriteString(s string, cell *Cell) error {
	a.buf.Reset()
	a.buf.WriteString(a.styleCodes(cell))
	a.buf.WriteString(s)
	a.buf.WriteString("\x1b[0m")
	_, err := a.writer.Write(a.buf.Bytes())
	return err
}

// ClearScreen clears the entire terminal screen.
func (a *ANSIAdapter) ClearScreen() error {
	_, err := a.writer.Write([]byte("\x1b[2J"))
	return err
}

// ClearLine clears the current line.
func (a *ANSIAdapter) ClearLine() error {
	_, err := a.writer.Write([]byte("\x1b[2K"))
	return err
}

// HideCursor hides the cursor.
func (a *ANSIAdapter) HideCursor() error {
	_, err := a.writer.Write([]byte("\x1b[?25l"))
	return err
}

// ShowCursor shows the cursor.
func (a *ANSIAdapter) ShowCursor() error {
	_, err := a.writer.Write([]byte("\x1b[?25h"))
	return err
}

// MoveCursor moves the cursor to the specified position.
func (a *ANSIAdapter) MoveCursor(x, y int) error {
	_, err := a.writer.Write([]byte(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1)))
	return err
}
