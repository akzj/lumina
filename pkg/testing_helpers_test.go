package v2

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/buffer"
	"github.com/akzj/lumina/pkg/output"
)

// nopAdapter is a no-op output adapter for benchmarks (avoids output overhead).
type nopAdapter struct{}

func (nopAdapter) WriteFull(_ *buffer.Buffer) error                   { return nil }
func (nopAdapter) WriteDirty(_ *buffer.Buffer, _ []buffer.Rect) error { return nil }
func (nopAdapter) Flush() error                                       { return nil }
func (nopAdapter) Close() error                                       { return nil }

// newLuaApp creates a NewApp with a fresh Lua state and TestAdapter.
func newLuaApp(t testing.TB, w, h int) (*App, *output.TestAdapter, *lua.State) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app := NewApp(L, w, h, ta)
	return app, ta, L
}

// readScreenLine reads up to maxLen rune characters from screen row y,
// stopping at the first zero-char cell.
func readScreenLine(ta *output.TestAdapter, y, maxLen int) string {
	var line []rune
	for x := 0; x < maxLen; x++ {
		c := ta.LastScreen.Get(x, y)
		if c.Char == 0 {
			break
		}
		line = append(line, c.Char)
	}
	return string(line)
}

// screenHasChar returns true if the screen contains the given rune anywhere.
func screenHasChar(ta *output.TestAdapter, r rune) bool {
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if ta.LastScreen.Get(x, y).Char == r {
				return true
			}
		}
	}
	return false
}

// screenHasString returns true if the string appears starting at some (x, y).
func screenHasString(ta *output.TestAdapter, s string) bool {
	runes := []rune(s)
	if len(runes) == 0 {
		return true
	}
	h := ta.LastScreen.Height()
	w := ta.LastScreen.Width()
	for y := 0; y < h; y++ {
		for x := 0; x <= w-len(runes); x++ {
			match := true
			for i, r := range runes {
				if ta.LastScreen.Get(x+i, y).Char != r {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}
