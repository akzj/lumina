package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestAriaAttributes(t *testing.T) {
	attrs := AriaAttributes{
		Role:     "button",
		Label:    "Submit form",
		Expanded: BoolPtr(false),
		Disabled: BoolPtr(true),
	}
	if attrs.Role != "button" {
		t.Fatalf("expected role 'button', got '%s'", attrs.Role)
	}
	if attrs.Label != "Submit form" {
		t.Fatalf("expected label 'Submit form', got '%s'", attrs.Label)
	}
	if *attrs.Expanded != false {
		t.Fatal("expected expanded=false")
	}
	if *attrs.Disabled != true {
		t.Fatal("expected disabled=true")
	}
}

func TestParseAriaFromMap(t *testing.T) {
	m := map[string]any{
		"role":     "dialog",
		"label":    "Confirm dialog",
		"expanded": true,
		"live":     "polite",
	}
	attrs := ParseAriaFromMap(m)
	if attrs.Role != "dialog" {
		t.Fatalf("expected 'dialog', got '%s'", attrs.Role)
	}
	if attrs.Label != "Confirm dialog" {
		t.Fatalf("expected 'Confirm dialog', got '%s'", attrs.Label)
	}
	if attrs.Expanded == nil || *attrs.Expanded != true {
		t.Fatal("expected expanded=true")
	}
	if attrs.Live != "polite" {
		t.Fatalf("expected 'polite', got '%s'", attrs.Live)
	}
}

func TestAriaToMap(t *testing.T) {
	attrs := AriaAttributes{
		Role:    "navigation",
		Label:   "Main nav",
		Pressed: BoolPtr(true),
	}
	m := AriaToMap(attrs)
	if m["role"] != "navigation" {
		t.Fatalf("expected 'navigation', got '%v'", m["role"])
	}
	if m["label"] != "Main nav" {
		t.Fatalf("expected 'Main nav', got '%v'", m["label"])
	}
	if m["pressed"] != true {
		t.Fatalf("expected pressed=true, got '%v'", m["pressed"])
	}
}

func TestAnnouncerQueue(t *testing.T) {
	a := GetAnnouncer()
	a.Reset()

	a.Announce("Form submitted", "polite")
	a.Announce("Error occurred", "assertive")

	if a.Pending() != 2 {
		t.Fatalf("expected 2 pending, got %d", a.Pending())
	}

	msgs := a.Drain()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Message != "Form submitted" {
		t.Fatalf("expected 'Form submitted', got '%s'", msgs[0].Message)
	}
	if msgs[0].Priority != "polite" {
		t.Fatalf("expected 'polite', got '%s'", msgs[0].Priority)
	}
	if msgs[1].Priority != "assertive" {
		t.Fatalf("expected 'assertive', got '%s'", msgs[1].Priority)
	}

	if a.Pending() != 0 {
		t.Fatalf("expected 0 pending after drain, got %d", a.Pending())
	}
	a.Reset()
}

func TestAnnouncerDefaultPriority(t *testing.T) {
	a := GetAnnouncer()
	a.Reset()
	a.Announce("test", "")
	msgs := a.Drain()
	if len(msgs) != 1 {
		t.Fatal("expected 1 message")
	}
	if msgs[0].Priority != "polite" {
		t.Fatalf("expected default 'polite', got '%s'", msgs[0].Priority)
	}
	a.Reset()
}

func TestTestRendererBasic(t *testing.T) {
	tr := NewTestRenderer()
	tr.Render(map[string]any{
		"type":    "box",
		"content": "Hello World",
	})

	root := tr.Root()
	if root == nil {
		t.Fatal("expected root node")
	}
	if root.Type != "box" {
		t.Fatalf("expected 'box', got '%s'", root.Type)
	}
	if root.Content != "Hello World" {
		t.Fatalf("expected 'Hello World', got '%s'", root.Content)
	}
}

func TestTestRendererGetByText(t *testing.T) {
	tr := NewTestRenderer()
	tr.Render(map[string]any{
		"type": "vbox",
		"children": []any{
			map[string]any{"type": "text", "content": "Title"},
			map[string]any{"type": "text", "content": "Description here"},
		},
	})

	node := tr.GetByText("Title")
	if node == nil {
		t.Fatal("expected to find 'Title'")
	}
	if node.Content != "Title" {
		t.Fatalf("expected 'Title', got '%s'", node.Content)
	}

	node2 := tr.GetByText("Description")
	if node2 == nil {
		t.Fatal("expected to find 'Description'")
	}

	node3 := tr.GetByText("Nonexistent")
	if node3 != nil {
		t.Fatal("expected nil for nonexistent text")
	}
}

func TestTestRendererGetByRole(t *testing.T) {
	tr := NewTestRenderer()
	tr.Render(map[string]any{
		"type": "vbox",
		"children": []any{
			map[string]any{
				"type": "hbox",
				"aria": map[string]any{"role": "button", "label": "Submit"},
			},
			map[string]any{
				"type": "hbox",
				"aria": map[string]any{"role": "dialog"},
			},
		},
	})

	btn := tr.GetByRole("button")
	if btn == nil {
		t.Fatal("expected to find button role")
	}
	if btn.Aria.Label != "Submit" {
		t.Fatalf("expected label 'Submit', got '%s'", btn.Aria.Label)
	}

	dlg := tr.GetByRole("dialog")
	if dlg == nil {
		t.Fatal("expected to find dialog role")
	}
}

func TestTestRendererFireEvent(t *testing.T) {
	tr := NewTestRenderer()
	tr.Render(map[string]any{"type": "box"})
	tr.FireEvent("button-1", "click", nil)
	tr.FireEvent("input-1", "change", nil)

	events := tr.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Target != "button-1" || events[0].EventType != "click" {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestRenderToString(t *testing.T) {
	node := &TestVNode{
		Type: "vbox",
		Children: []*TestVNode{
			{Type: "text", Content: "Hello "},
			{Type: "text", Content: "World"},
		},
	}
	result := RenderToString(node)
	if result != "Hello World" {
		t.Fatalf("expected 'Hello World', got '%s'", result)
	}
}

func TestLuaAnnounceAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalAnnouncer.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.announce("Task completed", "polite")
		lumina.announce("Error!", "assertive")
	`)
	if err != nil {
		t.Fatalf("Lua announce: %v", err)
	}

	msgs := globalAnnouncer.Drain()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 announcements, got %d", len(msgs))
	}
	if msgs[0].Message != "Task completed" {
		t.Fatalf("expected 'Task completed', got '%s'", msgs[0].Message)
	}
	globalAnnouncer.Reset()
}

func TestLuaTestRendererAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local renderer = lumina.createTestRenderer()
		assert(renderer ~= nil, "renderer should exist")
		assert(type(renderer.render) == "function", "render should be function")
		assert(type(renderer.getByText) == "function", "getByText should be function")
		assert(type(renderer.getByRole) == "function", "getByRole should be function")
		assert(type(renderer.fireEvent) == "function", "fireEvent should be function")
		assert(type(renderer.tostring) == "function", "tostring should be function")

		-- Render a simple tree
		renderer.render({
			type = "vbox",
			children = {
				{ type = "text", content = "Hello" },
				{ type = "text", content = "World" },
			}
		})

		-- getByText
		local node = renderer.getByText("Hello")
		assert(node ~= nil, "should find 'Hello'")
		assert(node.content == "Hello", "content should be 'Hello'")

		-- tostring
		local text = renderer.tostring()
		assert(type(text) == "string", "tostring should return string")
	`)
	if err != nil {
		t.Fatalf("Lua TestRenderer: %v", err)
	}
}
