package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func resetInspector() {
	globalInspector.enabled = false
	globalInspector.highlightID = ""
	globalInspector.selectedID = ""
	globalInspector.scrollY = 0
}

func TestDevTools_Toggle(t *testing.T) {
	resetInspector()

	if IsInspectorVisible() {
		t.Error("inspector should be off initially")
	}

	ToggleInspector()
	if !IsInspectorVisible() {
		t.Error("inspector should be on after toggle")
	}

	ToggleInspector()
	if IsInspectorVisible() {
		t.Error("inspector should be off after second toggle")
	}
}

func TestDevTools_ElementHighlight(t *testing.T) {
	resetInspector()
	ToggleInspector()

	// Create a frame with an owner node
	frame := NewFrame(40, 10)
	node := &VNode{
		Type: "text",
		Props: map[string]any{"id": "test-elem"},
		X: 5, Y: 2, W: 10, H: 3,
	}
	// Set owner node on some cells
	for y := 2; y < 5; y++ {
		for x := 5; x < 15; x++ {
			frame.Cells[y][x].OwnerNode = node
			frame.Cells[y][x].Char = 'X'
		}
	}

	// Hit test should find the node
	hit := HitTestFrame(frame, 7, 3)
	if hit == nil {
		t.Fatal("HitTestFrame returned nil")
	}
	if id, ok := hit.Props["id"].(string); !ok || id != "test-elem" {
		t.Errorf("hit ID = %v, want test-elem", hit.Props["id"])
	}

	// Set highlight
	SetInspectorHighlight("test-elem")
	if globalInspector.highlightID != "test-elem" {
		t.Errorf("highlightID = %q, want test-elem", globalInspector.highlightID)
	}

	// Render highlight should modify border cells
	RenderHighlight(frame, node)
	// Top-left corner cell should have blue highlight
	if frame.Cells[2][5].Background != "#3B82F6" {
		t.Errorf("highlight bg = %q, want #3B82F6", frame.Cells[2][5].Background)
	}

	resetInspector()
}

func TestDevTools_SelectElement(t *testing.T) {
	resetInspector()
	ToggleInspector()

	SetInspectorSelected("my-button")
	if globalInspector.selectedID != "my-button" {
		t.Errorf("selectedID = %q, want my-button", globalInspector.selectedID)
	}

	resetInspector()
}

func TestDevTools_TreeLines(t *testing.T) {
	resetInspector()

	root := &VNode{
		Type: "vbox",
		Props: map[string]any{"id": "root"},
		W: 80, H: 24,
		Children: []*VNode{
			{
				Type: "text",
				Content: "Hello",
				Props: map[string]any{"id": "greeting"},
				W: 80, H: 1,
			},
			{
				Type: "hbox",
				Props: map[string]any{"id": "row"},
				W: 80, H: 5,
				Children: []*VNode{
					{
						Type: "text",
						Content: "Left",
						Props: map[string]any{"id": "left"},
						W: 40, H: 5,
					},
					{
						Type: "text",
						Content: "Right",
						Props: map[string]any{"id": "right"},
						W: 40, H: 5,
					},
				},
			},
		},
	}

	lines := buildElementTree(root, 0)
	if len(lines) != 5 {
		t.Fatalf("expected 5 tree lines, got %d", len(lines))
	}

	// First line should be root
	if lines[0].id != "root" {
		t.Errorf("line 0 id = %q, want root", lines[0].id)
	}
	// Second line should be greeting
	if lines[1].id != "greeting" {
		t.Errorf("line 1 id = %q, want greeting", lines[1].id)
	}
	// Third line should be row
	if lines[2].id != "row" {
		t.Errorf("line 2 id = %q, want row", lines[2].id)
	}
}

func TestDevTools_StyleInfo(t *testing.T) {
	resetInspector()

	root := &VNode{
		Type: "vbox",
		Props: map[string]any{"id": "styled"},
		Style: Style{
			Width: 40, Height: 10,
			Border: "rounded",
			Foreground: "#FFFFFF",
			Background: "#1a1a2e",
			Padding: 2,
		},
		W: 40, H: 10,
	}

	SetInspectorSelected("styled")
	lines := buildStyleInspector(root)

	if len(lines) == 0 {
		t.Fatal("style info should not be empty")
	}

	// Should contain element type
	found := false
	for _, line := range lines {
		if line == " Element: vbox" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("style info should contain element type, got: %v", lines)
	}

	// Should contain border info
	foundBorder := false
	for _, line := range lines {
		if line == "  border: rounded" {
			foundBorder = true
			break
		}
	}
	if !foundBorder {
		t.Errorf("style info should contain border, got: %v", lines)
	}

	resetInspector()
}

func TestDevTools_PanelOverlay(t *testing.T) {
	resetInspector()
	ToggleInspector()

	root := &VNode{
		Type: "vbox",
		Props: map[string]any{"id": "app-root"},
		W: 80, H: 24,
		Children: []*VNode{
			{Type: "text", Content: "Hello", Props: map[string]any{"id": "txt"}, W: 80, H: 1},
		},
	}

	panelVNode := BuildInspectorVNode(root, 80, 24)
	if panelVNode == nil {
		t.Fatal("panel VNode should not be nil when inspector is enabled")
	}
	if panelVNode.Type != "vbox" {
		t.Errorf("panel type = %q, want vbox", panelVNode.Type)
	}
	if len(panelVNode.Children) == 0 {
		t.Error("panel should have children")
	}

	// Panel should render without panic
	computeFlexLayout(panelVNode, 40, 0, 40, 24)
	frame := NewFrame(80, 24)
	clip := Rect{X: 40, Y: 0, W: 40, H: 24}
	renderVNode(frame, panelVNode, clip)

	// Check panel area has content
	hasContent := false
	for y := 0; y < 24 && !hasContent; y++ {
		for x := 40; x < 80; x++ {
			if frame.Cells[y][x].Char != ' ' && frame.Cells[y][x].Char != 0 && !frame.Cells[y][x].Transparent {
				hasContent = true
				break
			}
		}
	}
	if !hasContent {
		t.Error("panel area should have visible content")
	}

	// When inspector is disabled, panel should be nil
	resetInspector()
	panelVNode = BuildInspectorVNode(root, 80, 24)
	if panelVNode != nil {
		t.Error("panel should be nil when inspector is disabled")
	}
}

func TestDevTools_LuaAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalInspector.enabled = false
	globalInspector.highlightID = ""
	globalInspector.selectedID = ""
	L := lua.NewState()
	Open(L)
	defer L.Close()
	globalInspector.enabled = false

	// Test toggleInspector on
	err := L.DoString(`
		local lumina = require("lumina")
		lumina.devtools.toggleInspector()
	`)
	if err != nil {
		t.Fatalf("toggleInspector: %v", err)
	}
	if !globalInspector.enabled {
		t.Error("expected enabled after toggleInspector")
	}

	// Test selectElement
	err = L.DoString(`
		local lumina = require("lumina")
		lumina.devtools.selectElement("my-element")
	`)
	if err != nil {
		t.Fatalf("selectElement: %v", err)
	}
	if globalInspector.selectedID != "my-element" {
		t.Errorf("selectedID = %q, want my-element", globalInspector.selectedID)
	}

	// Test isInspectorEnabled
	err = L.DoString(`
		local lumina = require("lumina")
		assert(lumina.devtools.isInspectorEnabled() == true, "should be enabled")
	`)
	if err != nil {
		t.Fatalf("isInspectorEnabled: %v", err)
	}

	// Test toggleInspector off
	err = L.DoString(`
		local lumina = require("lumina")
		lumina.devtools.toggleInspector()
	`)
	if err != nil {
		t.Fatalf("toggleInspector off: %v", err)
	}
	if globalInspector.enabled {
		t.Error("expected disabled after second toggleInspector")
	}

	resetInspector()
}
