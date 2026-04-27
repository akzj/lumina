package output

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
)

// tuiAdapter renders a buffer to an ANSI terminal.
type tuiAdapter struct {
	w *bufio.Writer
}

// NewTUIAdapter creates a TUI adapter that writes ANSI escape sequences to w.
func NewTUIAdapter(w io.Writer) Adapter {
	return &tuiAdapter{w: bufio.NewWriter(w)}
}

// WriteFull writes the entire screen buffer as ANSI output.
func (t *tuiAdapter) WriteFull(screen *buffer.Buffer) error {
	t.w.WriteString("\033[0m") // reset at start
	var curFg, curBg string
	var curBold, curDim, curUnderline bool

	for y := 0; y < screen.Height(); y++ {
		// Move cursor to start of row (1-based).
		fmt.Fprintf(t.w, "\033[%d;%dH", y+1, 1)
		for x := 0; x < screen.Width(); x++ {
			c := screen.Get(x, y)
			t.writeCell(c, &curFg, &curBg, &curBold, &curDim, &curUnderline)
			if c.Wide {
				x++ // skip the padding cell — terminal already advanced cursor
			}
		}
	}
	return nil
}

// WriteDirty writes only the cells within the dirty rects.
func (t *tuiAdapter) WriteDirty(screen *buffer.Buffer, dirtyRects []buffer.Rect) error {
	t.w.WriteString("\033[0m") // reset at start
	var curFg, curBg string
	var curBold, curDim, curUnderline bool

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
				t.writeCell(c, &curFg, &curBg, &curBold, &curDim, &curUnderline)
				if c.Wide {
					x++ // skip the padding cell — terminal already advanced cursor
				}
			}
		}
	}
	return nil
}

// writeCell emits ANSI escape sequences for a single cell, optimizing by
// tracking current state and only emitting changes.
func (t *tuiAdapter) writeCell(c buffer.Cell, curFg, curBg *string, curBold, curDim, curUnderline *bool) {
	// Check if attributes changed — if bold/dim/underline turned OFF, we need a reset.
	needReset := (!c.Bold && *curBold) || (!c.Dim && *curDim) || (!c.Underline && *curUnderline)
	if needReset {
		t.w.WriteString("\033[0m")
		*curFg = ""
		*curBg = ""
		*curBold = false
		*curDim = false
		*curUnderline = false
	}

	// Set foreground color if changed.
	if c.Foreground != *curFg {
		if c.Foreground == "" {
			// Only reset fg if we're not already at default (after a full reset).
			if *curFg != "" {
				t.w.WriteString("\033[39m")
			}
		} else {
			r, g, b := parseHexColor(c.Foreground)
			fmt.Fprintf(t.w, "\033[38;2;%d;%d;%dm", r, g, b)
		}
		*curFg = c.Foreground
	}

	// Set background color if changed.
	if c.Background != *curBg {
		if c.Background == "" {
			if *curBg != "" {
				t.w.WriteString("\033[49m")
			}
		} else {
			r, g, b := parseHexColor(c.Background)
			fmt.Fprintf(t.w, "\033[48;2;%d;%d;%dm", r, g, b)
		}
		*curBg = c.Background
	}

	// Set bold if changed.
	if c.Bold && !*curBold {
		t.w.WriteString("\033[1m")
		*curBold = true
	}

	// Set dim if changed.
	if c.Dim && !*curDim {
		t.w.WriteString("\033[2m")
		*curDim = true
	}

	// Set underline if changed.
	if c.Underline && !*curUnderline {
		t.w.WriteString("\033[4m")
		*curUnderline = true
	}

	// Write the character. Zero cells render as space.
	ch := c.Char
	if ch == 0 {
		ch = ' '
	}
	t.w.WriteRune(ch)
}

// Flush flushes buffered output.
func (t *tuiAdapter) Flush() error {
	return t.w.Flush()
}

// Close flushes and is a no-op otherwise (we don't own the writer).
func (t *tuiAdapter) Close() error {
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
