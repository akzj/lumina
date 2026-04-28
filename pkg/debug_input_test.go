package v2

import (
	"fmt"
	"testing"

	"github.com/akzj/lumina/pkg/event"
)

func TestDebugInputMode(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../examples/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	app.RenderAll()

	dumpScreen := func(label string, maxLines int) {
		fmt.Printf("\n=== %s ===\n", label)
		if ta.LastScreen == nil { fmt.Println("  (nil screen)"); return }
		for y := 0; y < maxLines && y < ta.LastScreen.Height(); y++ {
			line := ""
			for x := 0; x < ta.LastScreen.Width(); x++ {
				ch := ta.LastScreen.Get(x, y).Char
				if ch == 0 { line += " " } else { line += string(ch) }
			}
			fmt.Printf("  line %d: [%s]\n", y, line)
		}
	}

	// Check initial state
	t.Logf("Initial has '+ [a] Add new todo': %v", screenHasString(ta, "+ [a] Add new todo"))
	dumpScreen("Initial", 6)

	// Press 'a' to enter input mode
	app.HandleEvent(&event.Event{Type: "keydown", Key: "a"})
	app.RenderDirty()

	t.Logf("After 'a': has placeholder 'Type a new todo': %v", screenHasString(ta, "Type a new todo"))
	t.Logf("After 'a': still has '+ [a] Add new todo': %v", screenHasString(ta, "+ [a] Add new todo"))

	// Check focused node
	eng := app.Engine()
	focused := eng.FocusedNode()
	if focused != nil {
		t.Logf("Focused node: type=%s id=%s content=%q autoFocus=%v focusable=%v",
			focused.Type, focused.ID, focused.Content, focused.AutoFocus, focused.Focusable)
	} else {
		t.Log("WARNING: No focused node! Input won't receive keystrokes")
	}

	dumpScreen("After 'a'", 6)

	// Try typing 'h'
	app.HandleEvent(&event.Event{Type: "keydown", Key: "h"})
	app.RenderDirty()

	focused = eng.FocusedNode()
	if focused != nil {
		t.Logf("After 'h': focused content=%q cursorPos=%d", focused.Content, focused.CursorPos)
	} else {
		t.Log("After 'h': still no focused node")
	}

	dumpScreen("After typing 'h'", 6)

	// Try typing Chinese
	app.HandleEvent(&event.Event{Type: "keydown", Key: "你"})
	app.RenderDirty()

	focused = eng.FocusedNode()
	if focused != nil {
		t.Logf("After '你': focused content=%q cursorPos=%d", focused.Content, focused.CursorPos)
	}

	dumpScreen("After typing '你'", 6)
}
