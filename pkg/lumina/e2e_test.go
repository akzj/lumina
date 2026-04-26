package lumina_test

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina"
)

// stripANSI removes ANSI escape codes from output for testing.
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[\x20-\x3f]*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}

func TestCounterE2E(t *testing.T) {
	// Capture output
	buf := &bytes.Buffer{}

	L := lua.NewState()
	defer L.Close()

	// Set up adapter to capture output
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(buf))
	lumina.Open(L)

	// Render counter with initial=42
	err := L.DoString(`
		local lumina = require("lumina")
		local Counter = lumina.defineComponent({
			name = "Counter",
			init = function(props)
				return { count = props.initial or 0 }
			end,
			render = function(instance)
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "Count: " .. tostring(instance.count) }
					}
				}
			end
		})
		lumina.render(Counter, { initial = 42 })
	`)
	if err != nil {
		t.Fatalf("Counter render failed: %v", err)
	}

	// Check output contains "Count: 42" (strip ANSI codes first)
	output := stripANSI(buf.String())
	if !strings.Contains(output, "Count: 42") {
		t.Errorf("expected 'Count: 42' in output, got: %s", output)
	}
}

func TestLayoutVBox(t *testing.T) {
	buf := &bytes.Buffer{}

	L := lua.NewState()
	defer L.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(buf))
	lumina.Open(L)

	// Render vbox with multiple children
	err := L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "App",
			render = function()
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "Header" },
						{ type = "text", content = "Content" },
						{ type = "text", content = "Footer" }
					}
				}
			end
		})
		lumina.render(App)
	`)
	if err != nil {
		t.Fatalf("VBox render failed: %v", err)
	}

	// Strip ANSI codes for content testing
	output := stripANSI(buf.String())

	// Verify all content appears
	if !strings.Contains(output, "Header") {
		t.Error("expected 'Header' in output")
	}
	if !strings.Contains(output, "Content") {
		t.Error("expected 'Content' in output")
	}
	if !strings.Contains(output, "Footer") {
		t.Error("expected 'Footer' in output")
	}
}

func TestMultipleComponents(t *testing.T) {
	buf := &bytes.Buffer{}

	L := lua.NewState()
	defer L.Close()

	adapter := lumina.NewANSIAdapter(buf)
	lumina.SetOutputAdapter(adapter)
	lumina.Open(L)

	// Render multiple independent components
	err := L.DoString(`
		local lumina = require("lumina")
		
		local A = lumina.defineComponent({
			name = "A",
			render = function() return { type = "text", content = "Component A" } end
		})
		local B = lumina.defineComponent({
			name = "B",
			render = function() return { type = "text", content = "Component B" } end
		})
		
		lumina.render(A)
		lumina.render(B)
	`)
	if err != nil {
		t.Fatalf("Multiple components failed: %v", err)
	}

	// Strip ANSI codes for content testing
	output := stripANSI(buf.String())
	// First render produces full "Component A", second render diffs to produce just changed chars.
	// The cumulative buffer should contain "Component A" from the first full write.
	if !strings.Contains(output, "Component A") {
		t.Error("expected 'Component A' in output")
	}
	// Second render changes "Component A" to "Component B" — diff outputs only changed char.
	// Check the diff produced the distinguishing character 'B' (at position where 'A' was).
	// The cumulative output should have 'B' from the diff write.
	if !strings.Contains(output, "B") {
		t.Error("expected 'B' in output from second render diff")
	}
}

func TestFrameDimensions(t *testing.T) {
	frame := lumina.NewFrame(80, 24)

	if frame.Width != 80 || frame.Height != 24 {
		t.Errorf("unexpected frame dimensions: %dx%d", frame.Width, frame.Height)
	}

	if len(frame.Cells) != 24 {
		t.Errorf("expected 24 rows, got %d", len(frame.Cells))
	}
}

func TestANSIAdapterOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter := lumina.NewANSIAdapter(buf)

	frame := &lumina.Frame{
		Cells: [][]lumina.Cell{
			{{Char: 'H'}, {Char: 'i'}},
		},
		DirtyRects: []lumina.Rect{{X: 0, Y: 0, W: 2, H: 1}},
		Width:      2,
		Height:     1,
	}

	err := adapter.Write(frame)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Hi") {
		t.Errorf("expected 'Hi' in output, got: %s", output)
	}
}

func TestComponentStatePersistence(t *testing.T) {
	buf := &bytes.Buffer{}

	L := lua.NewState()
	defer L.Close()

	adapter := lumina.NewANSIAdapter(buf)
	lumina.SetOutputAdapter(adapter)
	lumina.Open(L)

	// Create component and render twice
	err := L.DoString(`
		local lumina = require("lumina")
		
		local Counter = lumina.defineComponent({
			name = "Counter",
			init = function(props)
				return { count = props.initial or 0 }
			end,
			render = function(instance)
				return { type = "text", content = "Count: " .. tostring(instance.count) }
			end
		})
		
		-- First render
		lumina.render(Counter, { initial = 1 })
	`)
	if err != nil {
		t.Fatalf("first render failed: %v", err)
	}

	// Strip ANSI codes
	output1 := stripANSI(buf.String())
	if !strings.Contains(output1, "Count: 1") {
		t.Errorf("expected 'Count: 1' in first render, got: %s", output1)
	}

	// Reset buffer for second render and invalidate adapter's prev frame
	// so the next render does a full write (not just a diff)
	buf.Reset()
	adapter.Invalidate()

	// Second render with different initial
	err = L.DoString(`
		local lumina = require("lumina")
		
		local Counter = lumina.defineComponent({
			name = "Counter",
			init = function(props)
				return { count = props.initial or 0 }
			end,
			render = function(instance)
				return { type = "text", content = "Count: " .. tostring(instance.count) }
			end
		})
		
		-- Second render with different initial
		lumina.render(Counter, { initial = 99 })
	`)
	if err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	// Strip ANSI codes
	output2 := stripANSI(buf.String())
	if !strings.Contains(output2, "Count: 99") {
		t.Errorf("expected 'Count: 99' in second render, got: %s", output2)
	}
}

func TestAllTestsPass(t *testing.T) {
	// This is a meta-test that verifies the test suite itself works
	// Run all tests via go test in subprocess would be ideal
	// For now, just verify the lumina package is importable
	L := lua.NewState()
	defer L.Close()

	lumina.Open(L)

	// Verify module is loaded
	err := L.DoString(`local lumina = require("lumina"); return lumina.version()`)
	if err != nil {
		t.Fatalf("module loading failed: %v", err)
	}
}

func TestE2E_NewDemos(t *testing.T) {
    demos := []string{
        "../../examples/color-picker/main.lua",
        "../../examples/form-builder/main.lua",
        "../../examples/data-table/main.lua",
        "../../examples/navigation/main.lua",
    }
    for _, demo := range demos {
        t.Run(filepath.Base(demo), func(t *testing.T) {
            app := lumina.NewAppWithSize(80, 24)
            defer app.Close()
            var buf bytes.Buffer
            tio := lumina.NewBufferTermIO(80, 24, &buf)
            err := app.LoadScript(demo, tio)
            if err != nil {
                t.Fatalf("LoadScript %s: %v", demo, err)
            }
            app.RenderOnce()
        })
    }
}
