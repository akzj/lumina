package lumina

import (
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestDevToolsEnableDisable(t *testing.T) {
	dt := GetDevTools()
	dt.Reset()

	if dt.IsEnabled() {
		t.Fatal("expected disabled initially")
	}
	dt.Enable()
	if !dt.IsEnabled() {
		t.Fatal("expected enabled after Enable()")
	}
	dt.Disable()
	if dt.IsEnabled() {
		t.Fatal("expected disabled after Disable()")
	}
	dt.Reset()
}

func TestDevToolsToggle(t *testing.T) {
	dt := GetDevTools()
	dt.Reset()

	dt.Toggle()
	if !dt.IsEnabled() {
		t.Fatal("expected enabled after Toggle()")
	}
	if !dt.IsVisible() {
		t.Fatal("expected visible after Toggle()")
	}
	dt.Toggle()
	if !dt.IsEnabled() {
		t.Fatal("expected still enabled after second Toggle()")
	}
	if dt.IsVisible() {
		t.Fatal("expected hidden after second Toggle()")
	}
	dt.Reset()
}

func TestDevToolsRecordRender(t *testing.T) {
	dt := GetDevTools()
	dt.Reset()
	dt.Enable()

	dt.RecordRender("MyComponent", 5*time.Millisecond)
	dt.RecordRender("MyComponent", 3*time.Millisecond)
	dt.RecordRender("OtherComponent", 1*time.Millisecond)

	if c := dt.GetRenderCount("MyComponent"); c != 2 {
		t.Fatalf("expected 2 renders, got %d", c)
	}
	if c := dt.GetRenderCount("OtherComponent"); c != 1 {
		t.Fatalf("expected 1 render, got %d", c)
	}
	if d := dt.GetRenderTime("MyComponent"); d != 3*time.Millisecond {
		t.Fatalf("expected 3ms, got %v", d)
	}
	dt.Reset()
}

func TestDevToolsComponentTree(t *testing.T) {
	dt := GetDevTools()
	dt.Reset()
	dt.Enable()

	tree := []*DevToolsNode{
		{
			ID: "app-1", Name: "App",
			Props: map[string]any{"title": "My App"},
			State: map[string]any{"count": 0},
			Children: []*DevToolsNode{
				{ID: "header-1", Name: "Header", Props: map[string]any{}, State: map[string]any{}},
				{ID: "content-1", Name: "Content", Props: map[string]any{}, State: map[string]any{}},
			},
		},
	}
	dt.UpdateTree(tree)

	got := dt.GetTree()
	if len(got) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(got))
	}
	if got[0].Name != "App" {
		t.Fatalf("expected App, got %s", got[0].Name)
	}
	if len(got[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(got[0].Children))
	}

	// Find by ID
	node := dt.GetNodeByID("header-1")
	if node == nil {
		t.Fatal("expected to find header-1")
	}
	if node.Name != "Header" {
		t.Fatalf("expected Header, got %s", node.Name)
	}
	dt.Reset()
}

func TestDevToolsInspector(t *testing.T) {
	dt := GetDevTools()
	dt.Reset()
	dt.Enable()

	tree := []*DevToolsNode{
		{
			ID: "comp-1", Name: "Counter",
			Props: map[string]any{"initial": 5},
			State: map[string]any{"count": 10},
			Hooks: []DevToolsHook{
				{Type: "useState", Value: 10},
				{Type: "useEffect", Value: nil},
			},
		},
	}
	dt.UpdateTree(tree)
	dt.RecordRender("Counter", 2*time.Millisecond)

	dt.SetSelected("comp-1")
	inspector := dt.RenderInspector()
	if inspector == "" {
		t.Fatal("expected non-empty inspector output")
	}
	// Should contain component name
	if !containsStr(inspector, "Counter") {
		t.Fatal("expected inspector to contain 'Counter'")
	}
	if !containsStr(inspector, "useState") {
		t.Fatal("expected inspector to contain 'useState'")
	}
	dt.Reset()
}

func TestDevToolsRenderTree(t *testing.T) {
	dt := GetDevTools()
	dt.Reset()
	dt.Enable()

	tree := []*DevToolsNode{
		{
			ID: "root", Name: "App",
			Children: []*DevToolsNode{
				{ID: "child1", Name: "Sidebar"},
				{ID: "child2", Name: "Content"},
			},
		},
	}
	dt.UpdateTree(tree)
	dt.RecordRender("App", time.Millisecond)

	output := dt.RenderTree()
	if !containsStr(output, "App") {
		t.Fatal("expected tree to contain 'App'")
	}
	if !containsStr(output, "Sidebar") {
		t.Fatal("expected tree to contain 'Sidebar'")
	}
	dt.Reset()
}

func TestDevToolsSummary(t *testing.T) {
	dt := GetDevTools()
	dt.Reset()
	dt.Enable()

	dt.RecordRender("A", time.Millisecond)
	dt.RecordRender("B", time.Millisecond)
	dt.RecordRender("A", time.Millisecond)

	summary := dt.Summary()
	if summary["enabled"] != true {
		t.Fatal("expected enabled=true")
	}
	if summary["total_renders"] != 3 {
		t.Fatalf("expected 3 total renders, got %v", summary["total_renders"])
	}
	dt.Reset()
}

func TestLuaDevToolsAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalDevTools.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		assert(lumina.devtools ~= nil, "devtools should exist")
		assert(lumina.devtools.isEnabled() == false, "should be disabled initially")

		lumina.devtools.enable()
		assert(lumina.devtools.isEnabled() == true, "should be enabled")

		lumina.devtools.toggle()
		assert(lumina.devtools.isVisible() == true, "should be visible")

		lumina.devtools.toggle()
		assert(lumina.devtools.isVisible() == false, "should be hidden")

		local tree = lumina.devtools.getTree()
		assert(type(tree) == "string", "tree should be string")

		local summary = lumina.devtools.summary()
		assert(type(summary) == "table", "summary should be table")
		assert(summary.enabled == true, "summary.enabled should be true")
	`)
	if err != nil {
		t.Fatalf("Lua devtools API: %v", err)
	}
	globalDevTools.Reset()
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
