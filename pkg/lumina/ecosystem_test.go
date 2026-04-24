package lumina_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

// ─── Task 1: Real-World Example Apps Load Without Error ──────────────

func TestExampleMarkdownViewerLoads(t *testing.T) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(120, 40, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/markdown-viewer/main.lua", tio)
	if err != nil {
		t.Fatalf("markdown-viewer failed to load: %v", err)
	}

	output := tio.Output()
	if len(output) == 0 {
		t.Error("markdown-viewer produced no output")
	}
}

func TestExampleAPIClientLoads(t *testing.T) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(120, 40, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/api-client/main.lua", tio)
	if err != nil {
		t.Fatalf("api-client failed to load: %v", err)
	}

	output := tio.Output()
	if len(output) == 0 {
		t.Error("api-client produced no output")
	}
}

func TestExampleKanbanLoads(t *testing.T) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(120, 40, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/kanban/main.lua", tio)
	if err != nil {
		t.Fatalf("kanban failed to load: %v", err)
	}

	output := tio.Output()
	if len(output) == 0 {
		t.Error("kanban produced no output")
	}
}

func TestExampleAIAgentLoads(t *testing.T) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(120, 40, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/ai-agent/main.lua", tio)
	if err != nil {
		t.Fatalf("ai-agent failed to load: %v", err)
	}

	output := tio.Output()
	if len(output) == 0 {
		t.Error("ai-agent produced no output")
	}
}

// ─── Task 2: Example Content Validation ──────────────────────────────

func TestMarkdownViewerHasDemoContent(t *testing.T) {
	data, err := os.ReadFile("../../examples/markdown-viewer/main.lua")
	if err != nil {
		t.Fatalf("read markdown-viewer: %v", err)
	}
	content := string(data)

	checks := []string{
		"parseMD",
		"lumina.defineComponent",
		"lumina.mount",
		"shadcn.Card",
		"shadcn.Separator",
		"demoMarkdown",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("markdown-viewer should contain %q", check)
		}
	}
}

func TestAPIClientHasComponents(t *testing.T) {
	data, err := os.ReadFile("../../examples/api-client/main.lua")
	if err != nil {
		t.Fatalf("read api-client: %v", err)
	}
	content := string(data)

	checks := []string{
		"createStore",
		"RequestPanel",
		"ResponsePanel",
		"HistorySidebar",
		"GET", "POST", "PUT", "DELETE",
		"shadcn.Button",
		"shadcn.Input",
		"shadcn.Card",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("api-client should contain %q", check)
		}
	}
}

func TestKanbanHasComponents(t *testing.T) {
	data, err := os.ReadFile("../../examples/kanban/main.lua")
	if err != nil {
		t.Fatalf("read kanban: %v", err)
	}
	content := string(data)

	checks := []string{
		"createStore",
		"KanbanCard",
		"KanbanColumn",
		"NewCardDialog",
		"Todo", "In Progress", "Done",
		"shadcn.Button",
		"shadcn.Dialog",
		"lumina.onKey",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("kanban should contain %q", check)
		}
	}
}

func TestAIAgentHasComponents(t *testing.T) {
	data, err := os.ReadFile("../../examples/ai-agent/main.lua")
	if err != nil {
		t.Fatalf("read ai-agent: %v", err)
	}
	content := string(data)

	checks := []string{
		"createStore",
		"StatusBadge",
		"Message",
		"ToolCall",
		"MessagesPanel",
		"ToolsPanel",
		"CommandInput",
		"StatsBar",
		"MCP",
		"addMessage",
		"addToolCall",
		"setStatus",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("ai-agent should contain %q", check)
		}
	}
}

// ─── Task 3: All Examples Directory Listing ──────────────────────────

func TestAllNewExamplesExist(t *testing.T) {
	examples := []string{
		"../../examples/markdown-viewer/main.lua",
		"../../examples/api-client/main.lua",
		"../../examples/kanban/main.lua",
		"../../examples/ai-agent/main.lua",
	}

	for _, ex := range examples {
		name := filepath.Dir(ex)
		t.Run(filepath.Base(name), func(t *testing.T) {
			if _, err := os.Stat(ex); os.IsNotExist(err) {
				t.Fatalf("example not found: %s", ex)
			}
		})
	}
}

// ─── Store Integration ───────────────────────────────────────────────

func TestStoreDispatchInExamples(t *testing.T) {
	// Verify createStore + dispatch works for the patterns used in examples
	app := lumina.NewApp()
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")

		local store = lumina.createStore({
			state = {
				count = 0,
				items = {},
			},
			actions = {
				increment = function(state)
					state.count = state.count + 1
				end,
				addItem = function(state, item)
					table.insert(state.items, item)
				end,
			},
		})

		store.dispatch("increment")
		store.dispatch("increment")
		store.dispatch("addItem", "hello")
		store.dispatch("addItem", "world")

		local s = store.getState()
		assert(s.count == 2, "count should be 2, got " .. tostring(s.count))
		assert(#s.items == 2, "items should have 2 entries")
		assert(s.items[1] == "hello", "first item should be 'hello'")
	`)
	if err != nil {
		t.Fatalf("store dispatch test failed: %v", err)
	}
}

func TestDefineComponentWithStore(t *testing.T) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.LoadScript("", tio)
	if err != nil {
		// LoadScript with empty path may fail, try DoString instead
	}

	err = app.L.DoString(`
		local lumina = require("lumina")

		local store = lumina.createStore({
			state = { status = "idle" },
			actions = {
				setStatus = function(state, s)
					state.status = s
				end,
			},
		})

		local StatusDisplay = lumina.defineComponent({
			name = "StatusDisplay",
			render = function(self)
				local s = store.getState()
				return { type = "text", content = "Status: " .. s.status }
			end,
		})

		store.dispatch("setStatus", "running")
		local s = store.getState()
		assert(s.status == "running", "status should be running")
	`)
	if err != nil {
		t.Fatalf("component + store test failed: %v", err)
	}
}
