// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"bytes"
	"fmt"
	"io"
)

// ANSIAdapter implements OutputAdapter for ANSI terminal output.
type ANSIAdapter struct {
	writer            io.Writer
	buf               bytes.Buffer
	width             int
	height            int
	colorMode         ColorMode // detected or configured color capability
	prevFrame         *Frame    // previous frame for cell-level diff
	DefaultBackground string    // fallback bg for cells with empty Background
	lastStyle         string    // last emitted style codes (skip reset when unchanged)
}

// NewANSIAdapter creates a new ANSIAdapter writing to the given writer.
// It auto-detects the terminal's color capability and reads the default
// background from the current theme.
func NewANSIAdapter(w io.Writer) *ANSIAdapter {
	bg := "#1E1E2E" // Catppuccin Mocha base (fallback)
	if theme := GetCurrentTheme(); theme != nil {
		if tbg, ok := theme.Colors["background"]; ok && tbg != "" {
			bg = tbg
		}
	}
	return &ANSIAdapter{
		writer:            w,
		buf:               bytes.Buffer{},
		width:             80,
		height:            24,
		colorMode:         DetectColorMode(),
		DefaultBackground: bg,
	}
}

// NewANSIAdapterWithColorMode creates an ANSIAdapter with a specific color mode.
func NewANSIAdapterWithColorMode(w io.Writer, mode ColorMode) *ANSIAdapter {
	bg := "#1E1E2E" // Catppuccin Mocha base (fallback)
	if theme := GetCurrentTheme(); theme != nil {
		if tbg, ok := theme.Colors["background"]; ok && tbg != "" {
			bg = tbg
		}
	}
	return &ANSIAdapter{
		writer:            w,
		buf:               bytes.Buffer{},
		width:             80,
		height:            24,
		colorMode:         mode,
		DefaultBackground: bg,
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
// Uses double buffering: compares against the previous frame and only writes
// changed cells. All output is buffered and written in a single Write call
// to eliminate tearing.
func (a *ANSIAdapter) Write(frame *Frame) error {
	a.buf.Reset()

	// If no dirty regions, nothing to write
	if len(frame.DirtyRects) == 0 {
		return nil
	}

	// Hide cursor during render to prevent flicker
	a.buf.WriteString("\x1b[?25l")

	if a.prevFrame == nil || a.prevFrame.Width != frame.Width || a.prevFrame.Height != frame.Height {
		// First frame or size changed — full write
		a.writeFullFrame(frame)
	} else {
		// Cell-level diff — only write changed cells
		a.writeDiffFrame(frame, a.prevFrame)
	}

	// Move cursor to safe position
	a.buf.WriteString("\x1b[999;1H")

	// Atomic write — single write call to terminal
	_, err := a.writer.Write(a.buf.Bytes())

	// Store current frame as previous for next diff
	a.prevFrame = frame.Clone()

	return err
}

// writeFullFrame writes every cell in the frame (used for first render or resize).
func (a *ANSIAdapter) writeFullFrame(frame *Frame) {
	for y := 0; y < frame.Height; y++ {
		// Move to start of line
		a.buf.WriteString(fmt.Sprintf("\x1b[%d;1H", y+1))
		var lastStyle string
		for x := 0; x < frame.Width; x++ {
			cell := frame.Cells[y][x]

			// Skip padding cells after wide characters
			if cell.Char == 0 {
				continue
			}

			style := a.styleCodes(&cell)
			if style != lastStyle {
				a.buf.WriteString(style)
				lastStyle = style
			}

			a.writeChar(cell.Char)
		}
		a.buf.WriteString("\x1b[0m")
	}
}

// writeDiffFrame writes only cells that differ from the previous frame.
// When dirty rects are available, only those regions are scanned.
func (a *ANSIAdapter) writeDiffFrame(newFrame, oldFrame *Frame) {
	if len(newFrame.DirtyRects) > 0 && !isFullFrameDirty(newFrame) {
		// Scan only dirty regions
		for _, rect := range newFrame.DirtyRects {
			a.writeDiffRegion(newFrame, oldFrame, rect)
		}
		return
	}
	// Full scan fallback
	a.writeDiffRegion(newFrame, oldFrame, Rect{X: 0, Y: 0, W: newFrame.Width, H: newFrame.Height})
}

// isFullFrameDirty returns true if the dirty rects cover the entire frame
// (single rect at 0,0 with full width/height). In that case, full scan is simpler.
func isFullFrameDirty(f *Frame) bool {
	if len(f.DirtyRects) == 1 {
		r := f.DirtyRects[0]
		return r.X == 0 && r.Y == 0 && r.W == f.Width && r.H == f.Height
	}
	return false
}

// writeDiffRegion writes changed cells within a specific rectangular region.
func (a *ANSIAdapter) writeDiffRegion(newFrame, oldFrame *Frame, rect Rect) {
	lastX, lastY := -1, -1
	var lastStyle string

	endY := rect.Y + rect.H
	if endY > newFrame.Height {
		endY = newFrame.Height
	}
	endX := rect.X + rect.W
	if endX > newFrame.Width {
		endX = newFrame.Width
	}
	startY := rect.Y
	if startY < 0 {
		startY = 0
	}
	startX := rect.X
	if startX < 0 {
		startX = 0
	}

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			newCell := newFrame.Cells[y][x]

			// Skip padding cells after wide characters
			if newCell.Char == 0 {
				continue
			}

			oldCell := oldFrame.Cells[y][x]

			// Skip unchanged cells
			if cellEqual(newCell, oldCell) {
				continue
			}

			// Move cursor if not sequential
			if x != lastX+1 || y != lastY {
				a.buf.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
			}

			// Write style + char
			style := a.styleCodes(&newCell)
			if style != lastStyle {
				a.buf.WriteString(style)
				lastStyle = style
			}

			a.writeChar(newCell.Char)

			lastX = x
			lastY = y
		}
	}
	if lastStyle != "" {
		a.buf.WriteString("\x1b[0m") // reset at end
	}
}

// writeChar writes a single character, escaping special characters.
func (a *ANSIAdapter) writeChar(ch rune) {
	switch ch {
	case '\\':
		a.buf.WriteByte('\\') // write backslash as-is (no double-escape)
	case '\x1b':
		// Skip raw escape characters — they'd be interpreted by terminal
	case '\r', '\n':
		// Skip control characters that would break cell grid
	default:
		a.buf.WriteRune(ch)
	}
}

// cellEqual returns true if two cells are visually identical.
func cellEqual(a, b Cell) bool {
	return a.Char == b.Char &&
		a.Foreground == b.Foreground &&
		a.Background == b.Background &&
		a.Bold == b.Bold &&
		a.Dim == b.Dim &&
		a.Underline == b.Underline &&
		a.Reverse == b.Reverse &&
		a.Blink == b.Blink &&
		a.Transparent == b.Transparent
}

// Invalidate clears the previous frame, forcing the next Write to do a full rewrite.
// Useful for tests that reset the output buffer between renders.
func (a *ANSIAdapter) Invalidate() {
	a.prevFrame = nil
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

	// Background color — use default background when empty
	bg := cell.Background
	if bg == "" {
		bg = a.DefaultBackground
	}
	if bg != "" && a.colorMode != ColorNone {
		codes += a.colorCode("bg", bg)
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
	// Map to the 6×6×6 color cube (indices 16-231).
	cubeVals := [6]int{0, 0x5f, 0x87, 0xaf, 0xd7, 0xff}
	ri := (r * 5 + 127) / 255
	gi := (g * 5 + 127) / 255
	bi := (b * 5 + 127) / 255
	cubeIdx := 16 + 36*ri + 6*gi + bi
	cr, cg, cb := cubeVals[ri], cubeVals[gi], cubeVals[bi]
	cubeDist := colorDist(r, g, b, cr, cg, cb)

	// Map to grayscale ramp (indices 232-255, values 8, 18, 28, ..., 238).
	gray := (r + g + b) / 3
	var grayIdx int
	if gray < 4 {
		grayIdx = 232
	} else if gray > 243 {
		grayIdx = 255
	} else {
		grayIdx = 232 + (gray-3)/10
	}
	grayVal := 8 + (grayIdx-232)*10
	grayDist := colorDist(r, g, b, grayVal, grayVal, grayVal)

	// Return whichever is closer.
	if grayDist < cubeDist {
		return grayIdx
	}
	return cubeIdx
}

// colorDist returns the squared Euclidean distance between two RGB colors.
func colorDist(r1, g1, b1, r2, g2, b2 int) int {
	dr := r1 - r2
	dg := g1 - g2
	db := b1 - b2
	return dr*dr + dg*dg + db*db
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
