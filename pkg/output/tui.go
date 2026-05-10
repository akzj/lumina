package output

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/akzj/lumina/pkg/buffer"
)

// tuiAdapter renders a buffer to an ANSI terminal.
type tuiAdapter struct {
	w      *bufio.Writer
	curX   int
	curY   int
	curVis bool
}

// tuiState tracks current ANSI attribute state for incremental output.
type tuiState struct {
	fg, bg                                                  string
	bold, dim, underline, italic, strikethrough, inverse bool
}

// NewTUIAdapter creates a TUI adapter that writes ANSI escape sequences to w.
func NewTUIAdapter(w io.Writer) Adapter {
	return &tuiAdapter{w: bufio.NewWriterSize(w, 256*1024)} // 256KB — enough for full screen
}

// WriteFull writes the entire screen buffer as ANSI output.
func (t *tuiAdapter) WriteFull(screen *buffer.Buffer) error {
	t.w.WriteString("\033[?2026h") // begin synchronized update
	t.w.WriteString("\033[0m")     // reset at start
	var st tuiState

	for y := 0; y < screen.Height(); y++ {
		// Move cursor to start of row (1-based).
		fmt.Fprintf(t.w, "\033[%d;%dH", y+1, 1)
		for x := 0; x < screen.Width(); x++ {
			c := screen.Get(x, y)
			t.writeCell(c, &st)
			if c.Wide {
				x++ // skip the next padding cell — terminal already advanced cursor by 2
			}
		}
	}
	t.w.WriteString("\033[?2026l") // end synchronized update
	return nil
}

// WriteDirty writes only the cells within the dirty rects.
func (t *tuiAdapter) WriteDirty(screen *buffer.Buffer, dirtyRects []buffer.Rect) error {
	t.w.WriteString("\033[?2026h") // begin synchronized update
	t.w.WriteString("\033[0m")     // reset at start
	var st tuiState

	bounds := buffer.Rect{X: 0, Y: 0, W: screen.Width(), H: screen.Height()}
	for _, dr := range dirtyRects {
		region := dr.Intersect(bounds)
		if region.W <= 0 || region.H <= 0 {
			continue
		}
		for y := region.Y; y < region.Y+region.H; y++ {
			// Move cursor to start of this dirty row segment.
			fmt.Fprintf(t.w, "\033[%d;%dH", y+1, region.X+1)
			for x := region.X; x < region.X+region.W; x++ {
				c := screen.Get(x, y)
				t.writeCell(c, &st)
				if c.Wide {
					x++ // skip the next padding cell — terminal already advanced cursor by 2
				}
			}
		}
	}
	t.w.WriteString("\033[?2026l") // end synchronized update
	return nil
}

// writeCell emits ANSI escape sequences for a single cell, optimizing by
// tracking current state and only emitting changes.
func (t *tuiAdapter) writeCell(c buffer.Cell, st *tuiState) {
	// Check if any attribute turned OFF — we need a full reset.
	needReset := (!c.Bold && st.bold) || (!c.Dim && st.dim) || (!c.Underline && st.underline) ||
		(!c.Italic && st.italic) || (!c.Strikethrough && st.strikethrough) || (!c.Inverse && st.inverse)
	if needReset {
		t.w.WriteString("\033[0m")
		*st = tuiState{} // reset all tracked state
	}

	// Set foreground color if changed.
	if c.Foreground != st.fg {
		if c.Foreground == "" {
			// Only reset fg if we're not already at default (after a full reset).
			if st.fg != "" {
				t.w.WriteString("\033[39m")
			}
		} else {
			r, g, b := parseHexColor(c.Foreground)
			fmt.Fprintf(t.w, "\033[38;2;%d;%d;%dm", r, g, b)
		}
		st.fg = c.Foreground
	}

	// Set background color if changed.
	if c.Background != st.bg {
		if c.Background == "" {
			if st.bg != "" {
				t.w.WriteString("\033[49m")
			}
		} else {
			r, g, b := parseHexColor(c.Background)
			fmt.Fprintf(t.w, "\033[48;2;%d;%d;%dm", r, g, b)
		}
		st.bg = c.Background
	}

	// Set bold if changed.
	if c.Bold && !st.bold {
		t.w.WriteString("\033[1m")
		st.bold = true
	}

	// Set dim if changed.
	if c.Dim && !st.dim {
		t.w.WriteString("\033[2m")
		st.dim = true
	}

	// Set italic if changed.
	if c.Italic && !st.italic {
		t.w.WriteString("\033[3m")
		st.italic = true
	}

	// Set underline if changed.
	if c.Underline && !st.underline {
		t.w.WriteString("\033[4m")
		st.underline = true
	}

	// Set inverse if changed.
	if c.Inverse && !st.inverse {
		t.w.WriteString("\033[7m")
		st.inverse = true
	}

	// Set strikethrough if changed.
	if c.Strikethrough && !st.strikethrough {
		t.w.WriteString("\033[9m")
		st.strikethrough = true
	}

	// Write the character. Control characters must not reach the terminal —
	// they would corrupt cursor positioning (e.g. \r resets to column 0).
	ch := c.Char
	if ch < 0x20 {
		// Control characters (including \0, \r, \n, \t, etc.) must not be
		// written to the terminal — they would corrupt cursor positioning.
		// Replace with space to preserve cell spacing.
		ch = ' '
	}
	t.w.WriteRune(ch)
}

// SetCursor positions the hardware cursor. Coordinates are 0-based.
// If visible is false, the cursor is hidden.
func (t *tuiAdapter) SetCursor(x, y int, visible bool) {
	t.curX = x
	t.curY = y
	t.curVis = visible
}

// Flush flushes buffered output, positioning the hardware cursor.
func (t *tuiAdapter) Flush() error {
	if t.curVis {
		// Position hardware cursor at the focused input's cursor location.
		// This is needed for IME candidate window positioning.
		// Keep cursor hidden — Lua textarea renders its own software cursor.
		fmt.Fprintf(t.w, "\033[%d;%dH", t.curY+1, t.curX+1)
		// NOTE: Do NOT show cursor (\033[?25h) — software cursor handles visibility
	} else {
		// No focused input — park cursor at top-left.
		t.w.WriteString("\033[1;1H")
	}
	// Always keep hardware cursor hidden — textarea renders its own blinking cursor
	t.w.WriteString("\033[?25l")
	return t.w.Flush()
}

// Close flushes and is a no-op otherwise (we don't own the writer).
func (t *tuiAdapter) Close() error {
	t.w.WriteString("\033[1;1H")
	return t.w.Flush()
}

// parseHexColor parses "#rrggbb" to r, g, b integers.
func parseHexColor(hex string) (r, g, b uint8) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0
	}
	rv, _ := strconv.ParseUint(hex[1:3], 16, 8)
	gv, _ := strconv.ParseUint(hex[3:5], 16, 8)
	bv, _ := strconv.ParseUint(hex[5:7], 16, 8)
	return uint8(rv), uint8(gv), uint8(bv)
}
