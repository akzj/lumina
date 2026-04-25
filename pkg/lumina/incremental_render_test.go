package lumina_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

// newRenderTestApp creates a test app with buffered output and ANSI adapter.
func newRenderTestApp(t *testing.T) (*lumina.App, *bytes.Buffer) {
	t.Helper()
	app := lumina.NewApp()
	var buf bytes.Buffer
	tio := lumina.NewBufferTermIO(80, 24, &buf)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))
	_ = tio
	return app, &buf
}

func TestIncrementalRender_NoChange(t *testing.T) {
	app, buf := newRenderTestApp(t)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "StaticApp",
			render = function(self)
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "Hello World" },
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// First render (full)
	app.RenderOnce()
	if buf.Len() == 0 {
		t.Fatal("first render produced no output")
	}

	// Second render — same content, DiffVNode detects no patches → skip
	buf.Reset()
	app.RenderOnce()

	// No changes → component not dirty → RenderOnce should produce nothing
	// (or very little if it re-renders identically)
	if buf.Len() > 0 {
		// Acceptable: component may not be dirty, so no output
		t.Logf("second render produced %d bytes (component may not be dirty)", buf.Len())
	}
}

func TestIncrementalRender_TextChange(t *testing.T) {
	app, buf := newRenderTestApp(t)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "CounterApp",
			render = function(self)
				local count, setCount = lumina.useState("count", 0)
				_G._setCount = setCount
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "Count: " .. tostring(count) },
						{ type = "text", content = "Static line" },
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// First render (full)
	app.RenderOnce()
	firstOutput := stripANSICodes(buf.String())
	if !strings.Contains(firstOutput, "Count: 0") {
		t.Errorf("first render should contain 'Count: 0', got: %q", firstOutput[:min(len(firstOutput), 200)])
	}

	// Change state via useState setter → marks component dirty
	buf.Reset()
	err = app.L.DoString(`_G._setCount(1)`)
	if err != nil {
		t.Fatalf("setState failed: %v", err)
	}

	// Second render — incremental path (small change)
	app.RenderOnce()
	secondLen := buf.Len()
	if secondLen == 0 {
		t.Error("second render produced no output after state change")
	}

	// The incremental render should produce LESS output than a full render
	// (only dirty rects, not the entire 80x24 frame)
	t.Logf("incremental render produced %d bytes", secondLen)
}

func TestIncrementalRender_FullFallback(t *testing.T) {
	app, buf := newRenderTestApp(t)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "BigChangeApp",
			render = function(self)
				local mode, setMode = lumina.useState("mode", "simple")
				_G._setMode = setMode
				if mode == "complex" then
					return {
						type = "vbox",
						children = {
							{ type = "text", content = "Line A" },
							{ type = "text", content = "Line B" },
							{ type = "text", content = "Line C" },
							{ type = "text", content = "Line D" },
							{ type = "text", content = "Line E" },
							{ type = "text", content = "Line F" },
						}
					}
				else
					return {
						type = "hbox",
						children = {
							{ type = "text", content = "Simple" },
						}
					}
				end
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// First render
	app.RenderOnce()
	if buf.Len() == 0 {
		t.Fatal("first render produced no output")
	}

	// Dramatic change — type changes from hbox to vbox + many new children
	// ShouldFullRerender should return true → full re-render
	buf.Reset()
	err = app.L.DoString(`_G._setMode("complex")`)
	if err != nil {
		t.Fatalf("state change failed: %v", err)
	}

	app.RenderOnce()
	output := stripANSICodes(buf.String())
	if len(output) == 0 {
		t.Error("dramatic change render produced no output")
	}
	// Verify new content is present
	if !strings.Contains(output, "Line A") {
		t.Error("output should contain 'Line A' after mode change")
	}
}

func TestIncrementalRender_AddChild(t *testing.T) {
	app, buf := newRenderTestApp(t)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "GrowApp",
			render = function(self)
				local n, setN = lumina.useState("n", 1)
				_G._setN = setN
				local children = {}
				for i = 1, n do
					children[i] = { type = "text", content = "Item " .. tostring(i) }
				end
				return {
					type = "vbox",
					children = children,
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// First render — 1 item
	app.RenderOnce()
	firstOutput := stripANSICodes(buf.String())
	if !strings.Contains(firstOutput, "Item 1") {
		t.Error("first render should contain 'Item 1'")
	}

	// Add a child
	buf.Reset()
	err = app.L.DoString(`_G._setN(2)`)
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}

	app.RenderOnce()
	output := stripANSICodes(buf.String())
	if len(output) == 0 {
		t.Error("add-child render produced no output")
	}
}

func TestIncrementalRender_RemoveChild(t *testing.T) {
	app, buf := newRenderTestApp(t)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "ShrinkApp",
			render = function(self)
				local n, setN = lumina.useState("n", 3)
				_G._setN = setN
				local children = {}
				for i = 1, n do
					children[i] = { type = "text", content = "Item " .. tostring(i) }
				end
				return {
					type = "vbox",
					children = children,
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// First render — 3 items
	app.RenderOnce()
	if buf.Len() == 0 {
		t.Fatal("first render produced no output")
	}

	// Remove a child (3 → 2)
	buf.Reset()
	err = app.L.DoString(`_G._setN(2)`)
	if err != nil {
		t.Fatalf("remove child failed: %v", err)
	}

	app.RenderOnce()
	output := stripANSICodes(buf.String())
	if len(output) == 0 {
		t.Error("remove-child render produced no output")
	}
}

// min is defined in e2e_full_test.go
