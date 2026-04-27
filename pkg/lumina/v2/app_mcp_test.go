package v2

import (
	"encoding/json"
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/mcp"
)

// TestApp_MCPInspectTree creates a real app with components and inspects the tree.
func TestApp_MCPInspectTree(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	app.RegisterComponent("counter", "Counter", buffer.Rect{X: 0, Y: 0, W: 40, H: 10}, 0,
		func(state map[string]any, props map[string]any) *layout.VNode {
			return layout.NewVNode("box")
		})
	app.RegisterComponent("timer", "Timer", buffer.Rect{X: 0, Y: 10, W: 40, H: 10}, 0,
		func(state map[string]any, props map[string]any) *layout.VNode {
			return layout.NewVNode("box")
		})
	app.RenderAll()

	tree := app.MCPInspectTree()
	if len(tree) != 2 {
		t.Fatalf("expected 2 components, got %d", len(tree))
	}

	// Verify component info.
	ids := make(map[string]bool)
	for _, c := range tree {
		ids[c.ID] = true
	}
	if !ids["counter"] {
		t.Error("missing 'counter' component")
	}
	if !ids["timer"] {
		t.Error("missing 'timer' component")
	}
}

// TestApp_MCPInspectComponent verifies component detail retrieval.
func TestApp_MCPInspectComponent(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	app.RegisterComponent("counter", "Counter", buffer.Rect{X: 5, Y: 3, W: 40, H: 10}, 2,
		func(state map[string]any, props map[string]any) *layout.VNode {
			return layout.NewVNode("box")
		})
	app.SetState("counter", "count", 42)
	app.RenderAll()

	detail, err := app.MCPInspectComponent("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.ID != "counter" {
		t.Errorf("expected ID 'counter', got %q", detail.ID)
	}
	if detail.Name != "Counter" {
		t.Errorf("expected Name 'Counter', got %q", detail.Name)
	}
	if detail.Rect != [4]int{5, 3, 40, 10} {
		t.Errorf("unexpected rect: %v", detail.Rect)
	}
	if detail.ZIndex != 2 {
		t.Errorf("expected ZIndex 2, got %d", detail.ZIndex)
	}
	if detail.State["count"] != 42 {
		t.Errorf("expected state count=42, got %v", detail.State["count"])
	}

	// Not found.
	_, err = app.MCPInspectComponent("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent component")
	}
}

// TestApp_MCPGetSetState verifies state get/set round-trip.
func TestApp_MCPGetSetState(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	app.RegisterComponent("counter", "Counter", buffer.Rect{X: 0, Y: 0, W: 40, H: 10}, 0,
		func(state map[string]any, props map[string]any) *layout.VNode {
			return layout.NewVNode("box")
		})
	app.RenderAll()

	// Set state via MCP.
	err := app.MCPSetState("counter", "count", 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get state via MCP.
	val, err := app.MCPGetState("counter", "count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 99 {
		t.Errorf("expected count=99, got %v", val)
	}

	// Get all state.
	allState, err := app.MCPGetState("counter", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stateMap, ok := allState.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", allState)
	}
	if stateMap["count"] != 99 {
		t.Errorf("expected count=99 in full state, got %v", stateMap["count"])
	}

	// Error cases.
	_, err = app.MCPGetState("nonexistent", "count")
	if err == nil {
		t.Error("expected error for nonexistent component")
	}
	_, err = app.MCPGetState("counter", "nonexistent_key")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

// TestApp_MCPSimulateClick verifies click dispatch via MCP.
func TestApp_MCPSimulateClick(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	clicked := false
	app.RegisterComponent("btn", "Button", buffer.Rect{X: 0, Y: 0, W: 20, H: 5}, 0,
		func(state map[string]any, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("box")
			vn.ID = "btn-root"
			vn.Props = map[string]any{
				"onClick": event.EventHandler(func(e *event.Event) {
					clicked = true
				}),
			}
			return vn
		})
	app.RenderAll()

	// Simulate click via MCP.
	err := app.MCPSimulateClick("btn-root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !clicked {
		t.Error("expected click handler to be called")
	}
}

// TestApp_MCPGetScreenText verifies screen text extraction.
func TestApp_MCPGetScreenText(t *testing.T) {
	app, _ := NewTestApp(10, 3)

	app.RegisterComponent("hello", "Hello", buffer.Rect{X: 0, Y: 0, W: 10, H: 3}, 0,
		func(state map[string]any, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Props = map[string]any{"content": "Hi"}
			return vn
		})
	app.RenderAll()

	text := app.MCPGetScreenText()
	if text == "" {
		t.Fatal("expected non-empty screen text")
	}
	// Should have 3 lines (3 rows + newlines).
	lines := 0
	for _, c := range text {
		if c == '\n' {
			lines++
		}
	}
	if lines != 3 {
		t.Errorf("expected 3 lines, got %d", lines)
	}
}

// TestApp_MCPFocus verifies focus navigation via MCP.
func TestApp_MCPFocus(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	app.RegisterComponent("form", "Form", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state map[string]any, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "form-root"

			input1 := layout.NewVNode("input")
			input1.ID = "input1"
			input1.Props = map[string]any{"focusable": true}

			input2 := layout.NewVNode("input")
			input2.ID = "input2"
			input2.Props = map[string]any{"focusable": true}

			root.Children = []*layout.VNode{input1, input2}
			return root
		})
	app.RenderAll()

	// Should have focusable IDs.
	ids := app.MCPGetFocusableIDs()
	if len(ids) < 2 {
		t.Fatalf("expected at least 2 focusable IDs, got %d: %v", len(ids), ids)
	}

	// Focus next.
	focused := app.MCPFocusNext()
	if focused == "" {
		t.Error("expected non-empty focused ID after FocusNext")
	}

	// Set focus.
	app.MCPSetFocus("input2")
	if app.MCPGetFocusedID() != "input2" {
		t.Errorf("expected focused 'input2', got %q", app.MCPGetFocusedID())
	}
}

// TestApp_MCPVersion verifies version string.
func TestApp_MCPVersion(t *testing.T) {
	app, _ := NewTestApp(80, 24)
	ver := app.MCPGetVersion()
	if ver != "lumina-v2" {
		t.Errorf("expected 'lumina-v2', got %q", ver)
	}
}

// TestApp_MCPHandler_Integration tests the full MCP handler wired to a real App.
func TestApp_MCPHandler_Integration(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	app.RegisterComponent("counter", "Counter", buffer.Rect{X: 0, Y: 0, W: 40, H: 10}, 0,
		func(state map[string]any, props map[string]any) *layout.VNode {
			return layout.NewVNode("box")
		})
	app.SetState("counter", "count", 0)
	app.RenderAll()

	h := mcp.NewHandler(app)

	// inspectTree via handler.
	resp := h.Handle(mcp.Request{ID: 1, Method: "inspectTree"})
	if resp.Error != nil {
		t.Fatalf("inspectTree error: %v", resp.Error)
	}

	// setState via handler.
	resp = h.Handle(mcp.Request{
		ID:     2,
		Method: "setState",
		Params: json.RawMessage(`{"id":"counter","key":"count","value":42}`),
	})
	if resp.Error != nil {
		t.Fatalf("setState error: %v", resp.Error)
	}

	// getState via handler — verify round-trip.
	resp = h.Handle(mcp.Request{
		ID:     3,
		Method: "getState",
		Params: json.RawMessage(`{"id":"counter","key":"count"}`),
	})
	if resp.Error != nil {
		t.Fatalf("getState error: %v", resp.Error)
	}

	// Tools list.
	tools := h.Tools()
	if len(tools) == 0 {
		t.Error("expected non-empty tool list")
	}
}
