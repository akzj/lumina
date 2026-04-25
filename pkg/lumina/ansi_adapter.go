// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"bytes"
	"fmt"
	"io"
)

// ANSIAdapter implements OutputAdapter for ANSI terminal output.
type ANSIAdapter struct {
	writer    io.Writer
	buf       bytes.Buffer
	width     int
	height    int
	colorMode ColorMode // detected or configured color capability
}

// NewANSIAdapter creates a new ANSIAdapter writing to the given writer.
// It auto-detects the terminal's color capability.
func NewANSIAdapter(w io.Writer) *ANSIAdapter {
	return &ANSIAdapter{
		writer:    w,
		buf:       bytes.Buffer{},
		width:     80,
		height:    24,
		colorMode: DetectColorMode(),
	}
}

// NewANSIAdapterWithColorMode creates an ANSIAdapter with a specific color mode.
func NewANSIAdapterWithColorMode(w io.Writer, mode ColorMode) *ANSIAdapter {
	return &ANSIAdapter{
		writer:    w,
		buf:       bytes.Buffer{},
		width:     80,
		height:    24,
		colorMode: mode,
	}
}

// ColorMode returns the adapter's current color mode.
func (a *ANSIAdapter) ColorMode() ColorMode {
	return a.colorMode
}

// SetColorMode sets the adapter's color mode.
func (a *ANSIAdapter) SetColorMode(mode ColorMode) {
	a.colorMode = mode
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

				// Skip padding cells after wide characters.
				// The terminal cursor already advances past them when the
				// wide char is emitted.
				if cell.Char == 0 {
					continue
				}

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
// It respects the adapter's color mode, downgrading colors as needed.
func (a *ANSIAdapter) styleCodes(cell *Cell) string {
	codes := "\x1b[0m" // reset

	// Foreground color
	if cell.Foreground != "" && a.colorMode != ColorNone {
		codes += a.colorCode("fg", cell.Foreground)
	}

	// Background color
	if cell.Background != "" && a.colorMode != ColorNone {
		codes += a.colorCode("bg", cell.Background)
	}

	// Text attributes (work in all modes except ColorNone)
	if a.colorMode != ColorNone {
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
	}

	return codes
}

// colorCode generates a foreground or background color escape code,
// respecting the adapter's color mode.
func (a *ANSIAdapter) colorCode(fgOrBg, color string) string {
	isHex := len(color) == 7 && color[0] == '#'

	if isHex {
		r, g, b := hexToRGB(color)
		switch a.colorMode {
		case ColorTrue:
			if fgOrBg == "fg" {
				return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
			}
			return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
		case Color256:
			idx := rgbTo256(r, g, b)
			if fgOrBg == "fg" {
				return fmt.Sprintf("\x1b[38;5;%dm", idx)
			}
			return fmt.Sprintf("\x1b[48;5;%dm", idx)
		case Color16:
			code := rgbTo16(r, g, b)
			if fgOrBg == "bg" {
				code += 10 // 30→40, 90→100
			}
			return fmt.Sprintf("\x1b[%dm", code)
		default:
			return ""
		}
	}

	// Named color — works in all color modes.
	return a.namedColor(fgOrBg, color)
}

// hexToRGB parses a "#RRGGBB" string to r, g, b components.
func hexToRGB(hex string) (r, g, b int) {
	r = hexByteToInt(hex[1:3])
	g = hexByteToInt(hex[3:5])
	b = hexByteToInt(hex[5:7])
	return
}

// hexByteToInt converts a 2-char hex string to an int.
func hexByteToInt(s string) int {
	var n int
	for _, c := range s {
		n *= 16
		if c >= '0' && c <= '9' {
			n += int(c - '0')
		} else if c >= 'a' && c <= 'f' {
			n += int(c-'a') + 10
		} else if c >= 'A' && c <= 'F' {
			n += int(c-'A') + 10
		}
	}
	return n
}

// rgbTo256 converts RGB to the nearest xterm-256 color index.
func rgbTo256(r, g, b int) int {
	// Check grayscale ramp first (indices 232-255).
	if r == g && g == b {
		if r < 8 {
			return 16 // black
		}
		if r > 248 {
			return 231 // white
		}
		return 232 + (r-8)*24/247
	}
	// Map to the 6×6×6 color cube (indices 16-231).
	ri := (r * 5 + 127) / 255
	gi := (g * 5 + 127) / 255
	bi := (b * 5 + 127) / 255
	return 16 + 36*ri + 6*gi + bi
}

// rgbTo16 converts RGB to the nearest ANSI 16-color foreground code (30-37, 90-97).
func rgbTo16(r, g, b int) int {
	// Simple luminance-based mapping.
	brightness := (r*299 + g*587 + b*114) / 1000
	bright := brightness > 128

	rBit := 0
	if r > 128 {
		rBit = 1
	}
	gBit := 0
	if g > 128 {
		gBit = 1
	}
	bBit := 0
	if b > 128 {
		bBit = 1
	}

	// ANSI color index: 0=black, 1=red, 2=green, 3=yellow, 4=blue, 5=magenta, 6=cyan, 7=white
	idx := bBit<<2 | gBit<<1 | rBit
	if bright {
		return 90 + idx // bright variant
	}
	return 30 + idx
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
