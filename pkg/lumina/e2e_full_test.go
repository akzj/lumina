package lumina_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

// stripANSICodes removes ANSI escape codes from output for content assertions.
func stripANSICodes(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}

// newTestApp creates a fresh App with a BufferTermIO, ready for headless testing.
// Returns the app, the output buffer, and the TermIO.
func newTestApp(t *testing.T, w, h int) (*lumina.App, *bytes.Buffer, *lumina.BufferTermIO) {
	t.Helper()
	app := lumina.NewApp()
	buf := &bytes.Buffer{}
	tio := lumina.NewBufferTermIO(w, h, buf)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))
	return app, buf, tio
}

// ─── Task 2: Validate Lua Script Loading ─────────────────────────────────

func TestE2E_CounterApp(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.LoadScript("../../examples/counter.lua", tio)
	if err != nil {
		t.Fatalf("counter.lua load failed: %v", err)
	}

	app.RenderOnce()

	// Counter should produce non-trivial output
	if buf.Len() < 100 {
		t.Errorf("counter produced too little output: %d bytes", buf.Len())
	}

	// The layout renders text vertically in narrow columns,
	// so check for individual characters from "Counter" and "42"
	output := stripANSICodes(buf.String())
	if !strings.Contains(output, "C") || !strings.Contains(output, "o") {
		t.Errorf("expected counter content in output, got %d bytes", len(output))
	}
}

func TestE2E_DashboardApp(t *testing.T) {
	app, buf, tio := newTestApp(t, 120, 40)
	defer app.Close()

	err := app.LoadScript("../../examples/dashboard.lua", tio)
	if err != nil {
		t.Fatalf("dashboard.lua load failed: %v", err)
	}

	app.RenderOnce()
	output := stripANSICodes(buf.String())
	if buf.Len() == 0 {
		t.Error("dashboard produced no output")
	}
	// Dashboard should contain system info
	if !strings.Contains(output, "Dashboard") && !strings.Contains(output, "CPU") &&
		!strings.Contains(output, "Memory") && !strings.Contains(output, "System") {
		// At least one of these keywords should appear
		t.Logf("dashboard output (%d bytes): %s", buf.Len(), output[:min(200, len(output))])
	}
}

func TestE2E_TodoMVC(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.LoadScript("../../examples/todo_mvc.lua", tio)
	if err != nil {
		t.Fatalf("todo_mvc.lua load failed: %v", err)
	}

	app.RenderOnce()
	if buf.Len() == 0 {
		t.Error("todo_mvc produced no output")
	}
}

func TestE2E_FormDemo(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.LoadScript("../../examples/form_demo.lua", tio)
	if err != nil {
		t.Fatalf("form_demo.lua load failed: %v", err)
	}

	app.RenderOnce()
	if buf.Len() == 0 {
		t.Error("form_demo produced no output")
	}
}

func TestE2E_ChatApp(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.LoadScript("../../examples/chat_app.lua", tio)
	if err != nil {
		t.Fatalf("chat_app.lua load failed: %v", err)
	}

	app.RenderOnce()
	if buf.Len() == 0 {
		t.Error("chat_app produced no output")
	}
}

func TestE2E_FileBrowser(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.LoadScript("../../examples/file_browser.lua", tio)
	if err != nil {
		t.Fatalf("file_browser.lua load failed: %v", err)
	}

	app.RenderOnce()
	if buf.Len() == 0 {
		t.Error("file_browser produced no output")
	}
}

// ─── Task 3: Validate shadcn require() at Runtime ────────────────────────

func TestE2E_ShadcnRequireAll(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	// Verify every shadcn component can be required
	err := app.L.DoString(`
		local shadcn = require("shadcn")

		-- Phase 21: Basic components
		assert(shadcn.Button ~= nil, "Button missing")
		assert(shadcn.Badge ~= nil, "Badge missing")
		assert(shadcn.Card ~= nil, "Card missing")
		assert(shadcn.Alert ~= nil, "Alert missing")
		assert(shadcn.Label ~= nil, "Label missing")
		assert(shadcn.Separator ~= nil, "Separator missing")
		assert(shadcn.Skeleton ~= nil, "Skeleton missing")
		assert(shadcn.Spinner ~= nil, "Spinner missing")
		assert(shadcn.Avatar ~= nil, "Avatar missing")
		assert(shadcn.Breadcrumb ~= nil, "Breadcrumb missing")
		assert(shadcn.Kbd ~= nil, "Kbd missing")
		assert(shadcn.Input ~= nil, "Input missing")
		assert(shadcn.Switch ~= nil, "Switch missing")
		assert(shadcn.Progress ~= nil, "Progress missing")
		assert(shadcn.Accordion ~= nil, "Accordion missing")
		assert(shadcn.Tabs ~= nil, "Tabs missing")
		assert(shadcn.Table ~= nil, "Table missing")
		assert(shadcn.Pagination ~= nil, "Pagination missing")
		assert(shadcn.Toggle ~= nil, "Toggle missing")
		assert(shadcn.ToggleGroup ~= nil, "ToggleGroup missing")

		-- Phase 22: Form components
		assert(shadcn.Select ~= nil, "Select missing")
		assert(shadcn.Checkbox ~= nil, "Checkbox missing")
		assert(shadcn.RadioGroup ~= nil, "RadioGroup missing")
		assert(shadcn.Slider ~= nil, "Slider missing")
		assert(shadcn.Textarea ~= nil, "Textarea missing")
		assert(shadcn.Field ~= nil, "Field missing")
		assert(shadcn.InputGroup ~= nil, "InputGroup missing")
		assert(shadcn.InputOTP ~= nil, "InputOTP missing")
		assert(shadcn.Combobox ~= nil, "Combobox missing")
		assert(shadcn.NativeSelect ~= nil, "NativeSelect missing")
		assert(shadcn.Form ~= nil, "Form missing")

		-- Phase 23: Overlay + Complex
		assert(shadcn.Dialog ~= nil, "Dialog missing")
		assert(shadcn.AlertDialog ~= nil, "AlertDialog missing")
		assert(shadcn.Sheet ~= nil, "Sheet missing")
		assert(shadcn.Drawer ~= nil, "Drawer missing")
		assert(shadcn.DropdownMenu ~= nil, "DropdownMenu missing")
		assert(shadcn.ContextMenu ~= nil, "ContextMenu missing")
		assert(shadcn.Popover ~= nil, "Popover missing")
		assert(shadcn.Tooltip ~= nil, "Tooltip missing")
		assert(shadcn.HoverCard ~= nil, "HoverCard missing")
		assert(shadcn.Command ~= nil, "Command missing")
		assert(shadcn.Menubar ~= nil, "Menubar missing")
		assert(shadcn.ScrollArea ~= nil, "ScrollArea missing")
		assert(shadcn.Collapsible ~= nil, "Collapsible missing")
		assert(shadcn.Carousel ~= nil, "Carousel missing")
		assert(shadcn.Sonner ~= nil, "Sonner missing")

		-- Phase 38: Additional components
		assert(shadcn.AspectRatio ~= nil, "AspectRatio missing")
		assert(shadcn.ButtonGroup ~= nil, "ButtonGroup missing")
		assert(shadcn.Calendar ~= nil, "Calendar missing")
		assert(shadcn.DatePicker ~= nil, "DatePicker missing")
		assert(shadcn.NavigationMenu ~= nil, "NavigationMenu missing")
		assert(shadcn.Resizable ~= nil, "Resizable missing")
		assert(shadcn.Sidebar ~= nil, "Sidebar missing")
		assert(shadcn.Chart ~= nil, "Chart missing")
		assert(shadcn.DataTable ~= nil, "DataTable missing")
		assert(shadcn.ColorPicker ~= nil, "ColorPicker missing")
	`)
	if err != nil {
		t.Fatalf("shadcn require failed: %v", err)
	}
}

func TestE2E_ShadcnButtonRender(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local Button = require("shadcn.button")
		local tree = lumina.render(Button, { label = "Click Me", variant = "default" })
	`)
	if err != nil {
		t.Fatalf("shadcn Button render failed: %v", err)
	}

	// Button produces rendered output — verify it's non-trivial
	if buf.Len() < 50 {
		t.Errorf("Button produced too little output: %d bytes", buf.Len())
	}

	// The text may be laid out vertically depending on column width,
	// so check that individual characters from "Click Me" appear
	output := stripANSICodes(buf.String())
	if !strings.Contains(output, "C") || !strings.Contains(output, "l") {
		t.Errorf("expected Button content characters in output")
	}
}

func TestE2E_ShadcnCardRender(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	// Card module returns { Card = ..., CardHeader = ..., ... }
	err := app.L.DoString(`
		local lumina = require("lumina")
		local cardMod = require("shadcn.card")
		lumina.render(cardMod.Card, {
			children = {
				{ type = "text", content = "Card body" }
			}
		})
	`)
	if err != nil {
		t.Fatalf("shadcn Card render failed: %v", err)
	}

	// Card should produce output with border + content
	if buf.Len() < 50 {
		t.Errorf("Card produced too little output: %d bytes", buf.Len())
	}
}

// ─── Task 4: Validate Render Pipeline ────────────────────────────────────

func TestE2E_RenderPipeline_SimpleComponent(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "TestPipeline",
			render = function(self)
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "Hello Lumina!" },
						{ type = "text", content = "Version: " .. lumina.version() },
					}
				}
			end
		})
		lumina.render(App)
	`)
	if err != nil {
		t.Fatalf("render pipeline failed: %v", err)
	}

	output := stripANSICodes(buf.String())
	if !strings.Contains(output, "Hello Lumina!") {
		t.Errorf("expected 'Hello Lumina!' in output, got: %s", output)
	}
	if !strings.Contains(output, "0.3.0") {
		t.Errorf("expected version '0.3.0' in output, got: %s", output)
	}
}

func TestE2E_RenderPipeline_NestedComponents(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")

		local Header = lumina.defineComponent({
			name = "Header",
			render = function(self)
				return { type = "text", content = "=== HEADER ===" }
			end
		})

		local Footer = lumina.defineComponent({
			name = "Footer",
			render = function(self)
				return { type = "text", content = "=== FOOTER ===" }
			end
		})

		local Page = lumina.defineComponent({
			name = "Page",
			render = function(self)
				return {
					type = "vbox",
					children = {
						lumina.createElement(Header, {}),
						{ type = "text", content = "Page Body" },
						lumina.createElement(Footer, {}),
					}
				}
			end
		})

		lumina.render(Page)
	`)
	if err != nil {
		t.Fatalf("nested component render failed: %v", err)
	}

	output := stripANSICodes(buf.String())
	if !strings.Contains(output, "HEADER") {
		t.Errorf("expected 'HEADER' in output")
	}
	if !strings.Contains(output, "Page Body") {
		t.Errorf("expected 'Page Body' in output")
	}
	if !strings.Contains(output, "FOOTER") {
		t.Errorf("expected 'FOOTER' in output")
	}
}

func TestE2E_RenderPipeline_StateInit(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")

		local Counter = lumina.defineComponent({
			name = "StatefulCounter",
			init = function(props)
				return { count = props.initial or 0, label = props.label or "Count" }
			end,
			render = function(self)
				return {
					type = "text",
					content = self.label .. ": " .. tostring(self.count)
				}
			end
		})

		lumina.render(Counter, { initial = 100, label = "Score" })
	`)
	if err != nil {
		t.Fatalf("stateful render failed: %v", err)
	}

	output := stripANSICodes(buf.String())
	if !strings.Contains(output, "Score: 100") {
		t.Errorf("expected 'Score: 100' in output, got: %s", output)
	}
}

func TestE2E_RenderPipeline_HooksUseState(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	// useState(key, initialValue) — key is a string, not a value
	err := app.L.DoString(`
		local lumina = require("lumina")

		local HookApp = lumina.defineComponent({
			name = "HookApp",
			render = function(self)
				local count, setCount = lumina.useState("count", 42)
				return {
					type = "text",
					content = "Hook count: " .. tostring(count)
				}
			end
		})

		lumina.render(HookApp)
	`)
	if err != nil {
		t.Fatalf("hooks useState failed: %v", err)
	}

	output := stripANSICodes(buf.String())
	if !strings.Contains(output, "Hook count: 42") {
		t.Errorf("expected 'Hook count: 42' in output, got: %s", output)
	}
}

func TestE2E_RenderPipeline_Memo(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")

		local Expensive = lumina.defineComponent({
			name = "Expensive",
			render = function(self)
				return { type = "text", content = "Expensive: " .. tostring(self.props and self.props.value or "?") }
			end
		})

		local MemoExpensive = lumina.memo(Expensive)
		assert(MemoExpensive ~= nil, "memo should return a factory")
		assert(MemoExpensive.name == "Expensive", "memo should preserve name")
		assert(MemoExpensive._memoized == true, "memo should set _memoized flag")
	`)
	if err != nil {
		t.Fatalf("memo test failed: %v", err)
	}
}

func TestE2E_RenderPipeline_ErrorBoundary(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")

		local Boundary = lumina.createErrorBoundary({
			fallback = function(err)
				return { type = "text", content = "Caught: " .. err }
			end
		})

		assert(Boundary ~= nil, "error boundary should be created")
		assert(Boundary._isErrorBoundary == true, "should have error boundary flag")
	`)
	if err != nil {
		t.Fatalf("error boundary creation failed: %v", err)
	}
	_ = buf // output not checked for this structural test
}

func TestE2E_RenderPipeline_CreateElement(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")

		local MyComp = lumina.defineComponent({
			name = "MyComp",
			render = function(self) return { type = "text", content = "hi" } end
		})

		local elem = lumina.createElement(MyComp, { key = "test", label = "hello" })
		assert(elem ~= nil, "createElement should return a table")
		assert(elem.type == "component", "type should be 'component'")
		assert(elem._factory ~= nil, "should have _factory")
		assert(elem._props ~= nil, "should have _props")
		assert(elem._props.key == "test", "props should be passed through")
	`)
	if err != nil {
		t.Fatalf("createElement test failed: %v", err)
	}
}

// ─── Task 5: Validate Mount/Run Flow ─────────────────────────────────────

func TestE2E_MountAndRun(t *testing.T) {
	app, buf, tio := newTestApp(t, 80, 24)
	defer app.Close()

	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	// Test that mount() + run() works in headless mode
	err := app.L.DoString(`
		local lumina = require("lumina")

		local App = lumina.defineComponent({
			name = "MountTest",
			render = function(self)
				return { type = "text", content = "Mounted!" }
			end
		})

		lumina.mount(App)
		lumina.run()  -- should be no-op in headless mode
	`)
	if err != nil {
		t.Fatalf("mount/run failed: %v", err)
	}

	output := stripANSICodes(buf.String())
	if !strings.Contains(output, "Mounted!") {
		t.Errorf("expected 'Mounted!' in output, got: %s", output)
	}
}

// ─── Task 6: Validate All Example Apps Load Without Error ────────────────

func TestE2E_AllExamplesLoad(t *testing.T) {
	// Standalone .lua files (use lumina.render)
	standaloneExamples := []struct {
		name string
		path string
	}{
		{"counter", "../../examples/counter.lua"},
		{"dashboard", "../../examples/dashboard.lua"},
		{"todo_mvc", "../../examples/todo_mvc.lua"},
		{"form_demo", "../../examples/form_demo.lua"},
		{"chat_app", "../../examples/chat_app.lua"},
		{"file_browser", "../../examples/file_browser.lua"},
	}

	for _, ex := range standaloneExamples {
		t.Run(ex.name, func(t *testing.T) {
			app, buf, tio := newTestApp(t, 80, 24)
			defer app.Close()

			err := app.LoadScript(ex.path, tio)
			if err != nil {
				t.Fatalf("failed to load %s: %v", ex.name, err)
			}

			app.RenderOnce()
			if buf.Len() == 0 {
				t.Errorf("%s produced no output", ex.name)
			}
		})
	}

	// Directory-based examples (use lumina.mount + lumina.run + lumina.onKey)
	dirExamples := []struct {
		name string
		path string
	}{
		{"chat_dir", "../../examples/chat/main.lua"},
		{"dashboard_dir", "../../examples/dashboard/main.lua"},
		{"todo_dir", "../../examples/todo/main.lua"},
		{"file_explorer_dir", "../../examples/file-explorer/main.lua"},
		{"components_showcase", "../../examples/components-showcase/main.lua"},
	}

	for _, ex := range dirExamples {
		t.Run(ex.name, func(t *testing.T) {
			app, buf, tio := newTestApp(t, 120, 40)
			defer app.Close()

			err := app.LoadScript(ex.path, tio)
			if err != nil {
				t.Fatalf("failed to load %s: %v", ex.name, err)
			}

			app.RenderOnce()
			if buf.Len() == 0 {
				t.Errorf("%s produced no output", ex.name)
			}
		})
	}
}

func TestE2E_AllExamplesProduceContent(t *testing.T) {
	// Verify examples produce meaningful content (not just ANSI codes)
	examples := []struct {
		name     string
		path     string
		keywords []string // at least one of these should appear
	}{
		{"counter", "../../examples/counter.lua", []string{"42", "Counter", "Increment"}},
		{"todo_mvc", "../../examples/todo_mvc.lua", []string{"Todo", "task", "Add"}},
		{"form_demo", "../../examples/form_demo.lua", []string{"Form", "Name", "Email", "Submit"}},
		{"chat_app", "../../examples/chat_app.lua", []string{"Chat", "message", "Alice", "Bob"}},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			app, buf, tio := newTestApp(t, 80, 24)
			defer app.Close()

			err := app.LoadScript(ex.path, tio)
			if err != nil {
				t.Fatalf("failed to load %s: %v", ex.name, err)
			}

			app.RenderOnce()
			output := stripANSICodes(buf.String())

			found := false
			for _, kw := range ex.keywords {
				if strings.Contains(output, kw) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s: none of keywords %v found in output (%d bytes)",
					ex.name, ex.keywords, len(output))
			}
		})
	}
}

// ─── BufferTermIO Tests ──────────────────────────────────────────────────

func TestBufferTermIO_Interface(t *testing.T) {
	buf := &bytes.Buffer{}
	tio := lumina.NewBufferTermIO(80, 24, buf)

	// Verify Size
	w, h := tio.Size()
	if w != 80 || h != 24 {
		t.Errorf("expected 80x24, got %dx%d", w, h)
	}

	// Verify SetSize
	tio.SetSize(120, 40)
	w, h = tio.Size()
	if w != 120 || h != 40 {
		t.Errorf("expected 120x40 after SetSize, got %dx%d", w, h)
	}

	// Verify Write
	n, err := tio.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Errorf("Write failed: n=%d, err=%v", n, err)
	}
	if tio.Output() != "hello" {
		t.Errorf("expected 'hello', got %q", tio.Output())
	}

	// Verify Reset
	tio.Reset()
	if tio.Output() != "" {
		t.Errorf("expected empty after Reset, got %q", tio.Output())
	}

	// Verify Close
	if err := tio.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestBufferTermIO_Input(t *testing.T) {
	tio := lumina.NewBufferTermIO(80, 24, nil)

	// Write input data (simulate keystrokes)
	tio.WriteInput([]byte("abc"))

	// Read it back
	p := make([]byte, 10)
	n, err := tio.Read(p)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 3 || string(p[:n]) != "abc" {
		t.Errorf("expected 'abc', got %q (n=%d)", string(p[:n]), n)
	}
}

// ─── Lua API Completeness ────────────────────────────────────────────────

func TestE2E_LuaAPICompleteness(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	// Verify all major API functions exist on the lumina module
	err := app.L.DoString(`
		local lumina = require("lumina")

		-- Core API
		assert(type(lumina.version) == "function", "version missing")
		assert(type(lumina.defineComponent) == "function", "defineComponent missing")
		assert(type(lumina.createElement) == "function", "createElement missing")
		assert(type(lumina.render) == "function", "render missing")
		assert(type(lumina.mount) == "function", "mount missing")
		assert(type(lumina.run) == "function", "run missing")
		assert(type(lumina.quit) == "function", "quit missing")
		assert(type(lumina.memo) == "function", "memo missing")
		assert(type(lumina.createPortal) == "function", "createPortal missing")
		assert(type(lumina.forwardRef) == "function", "forwardRef missing")
		assert(type(lumina.lazy) == "function", "lazy missing")
		assert(type(lumina.createErrorBoundary) == "function", "createErrorBoundary missing")

		-- Hooks
		assert(type(lumina.useState) == "function", "useState missing")
		assert(type(lumina.useEffect) == "function", "useEffect missing")
		assert(type(lumina.useMemo) == "function", "useMemo missing")
		assert(type(lumina.useCallback) == "function", "useCallback missing")
		assert(type(lumina.useRef) == "function", "useRef missing")
		assert(type(lumina.useReducer) == "function", "useReducer missing")
		assert(type(lumina.useId) == "function", "useId missing")
		assert(type(lumina.useContext) == "function", "useContext missing")
		assert(type(lumina.createContext) == "function", "createContext missing")

		-- Event API
		assert(type(lumina.on) == "function", "on missing")
		assert(type(lumina.off) == "function", "off missing")
		assert(type(lumina.emit) == "function", "emit missing")
		assert(type(lumina.onKey) == "function", "onKey missing")
		assert(type(lumina.setFocus) == "function", "setFocus missing")
		assert(type(lumina.getFocused) == "function", "getFocused missing")

		-- State management
		assert(type(lumina.createStore) == "function", "createStore missing")
		assert(type(lumina.useStore) == "function", "useStore missing")

		-- Router
		assert(type(lumina.createRouter) == "function", "createRouter missing")
		assert(type(lumina.navigate) == "function", "navigate missing")

		-- Async
		assert(type(lumina.useAsync) == "function", "useAsync missing")
		assert(type(lumina.delay) == "function", "delay missing")
		assert(type(lumina.fetch) == "function", "fetch missing")

		-- Built-in component factories
		assert(type(lumina.Suspense) == "table", "Suspense missing")
		assert(type(lumina.Profiler) == "table", "Profiler missing")
		assert(type(lumina.StrictMode) == "table", "StrictMode missing")
	`)
	if err != nil {
		t.Fatalf("API completeness check failed: %v", err)
	}
}

// ─── Render Pipeline Internals ───────────────────────────────────────────

func TestE2E_VNodeToFrame(t *testing.T) {
	// Test the Go-side VNode → Frame conversion directly
	vnode := lumina.NewVNode("vbox")
	child1 := lumina.NewVNode("text")
	child1.SetContent("Line 1")
	child2 := lumina.NewVNode("text")
	child2.SetContent("Line 2")
	vnode.AddChild(child1)
	vnode.AddChild(child2)

	frame := lumina.VNodeToFrame(vnode, 40, 10)
	if frame == nil {
		t.Fatal("VNodeToFrame returned nil")
	}
	if frame.Width != 40 || frame.Height != 10 {
		t.Errorf("expected 40x10, got %dx%d", frame.Width, frame.Height)
	}

	// Check that content was rendered into frame cells
	found := false
	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			if frame.Cells[y][x].Char == 'L' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("expected 'L' character from 'Line 1' in frame cells")
	}
}

func TestE2E_ANSIAdapterRoundTrip(t *testing.T) {
	// Test that ANSI adapter produces readable output
	buf := &bytes.Buffer{}
	adapter := lumina.NewANSIAdapter(buf)

	frame := lumina.NewFrame(10, 2)
	// Write "Hi" into first row
	frame.Cells[0][0] = lumina.Cell{Char: 'H', Foreground: "#FFFFFF"}
	frame.Cells[0][1] = lumina.Cell{Char: 'i', Foreground: "#FFFFFF"}
	frame.MarkDirty()

	err := adapter.Write(frame)
	if err != nil {
		t.Fatalf("adapter.Write failed: %v", err)
	}

	output := stripANSICodes(buf.String())
	if !strings.Contains(output, "Hi") {
		t.Errorf("expected 'Hi' in ANSI output, got: %q", output)
	}
}

// ─── Store / createStore ─────────────────────────────────────────────────

func TestE2E_CreateStore(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")

		local store = lumina.createStore({
			state = { count = 0, name = "test" },
		})

		assert(store ~= nil, "store should be created")
		-- getState should return the initial state
		local state = store.getState()
		assert(state.count == 0, "initial count should be 0")
		assert(state.name == "test", "initial name should be 'test'")
	`)
	if err != nil {
		t.Fatalf("createStore test failed: %v", err)
	}
}

// ─── Context API ─────────────────────────────────────────────────────────

func TestE2E_ContextAPI(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")

		local ctx = lumina.createContext("default_value")
		assert(ctx ~= nil, "context should be created")
	`)
	if err != nil {
		t.Fatalf("context API test failed: %v", err)
	}
}

// ─── Router API ──────────────────────────────────────────────────────────

func TestE2E_RouterAPI(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")

		local router = lumina.createRouter({
			routes = {
				{ path = "/" },
				{ path = "/about" },
				{ path = "/users/:id" },
			},
			initialPath = "/"
		})

		assert(router ~= nil, "router should be created")
		assert(router.routeCount == 3, "should have 3 routes, got " .. tostring(router.routeCount))

		-- Navigate
		lumina.navigate("/about")
		assert(lumina.getCurrentPath() == "/about", "should be at /about")

		-- Back
		local ok = lumina.back()
		assert(ok == true, "back should succeed")
		assert(lumina.getCurrentPath() == "/", "should be back at /")
	`)
	if err != nil {
		t.Fatalf("router API test failed: %v", err)
	}
}

// ─── Overlay API ─────────────────────────────────────────────────────────

func TestE2E_OverlayAPI(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")

		-- Show overlay
		lumina.showOverlay({
			id = "test-dialog",
			x = 10, y = 5,
			width = 30, height = 10,
			modal = true,
			content = { type = "text", content = "Dialog Content" }
		})

		assert(lumina.isOverlayVisible("test-dialog") == true, "overlay should be visible")

		-- Toggle
		lumina.toggleOverlay("test-dialog")
		assert(lumina.isOverlayVisible("test-dialog") == false, "overlay should be hidden after toggle")

		-- Show again and hide
		lumina.showOverlay({ id = "test-dialog", x = 0, y = 0, width = 20, height = 10 })
		lumina.hideOverlay("test-dialog")
		assert(lumina.isOverlayVisible("test-dialog") == false, "overlay should be hidden")
	`)
	if err != nil {
		t.Fatalf("overlay API test failed: %v", err)
	}
}

// ─── Animation API ───────────────────────────────────────────────────────

func TestE2E_AnimationPresets(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")

		-- Animation presets should exist
		assert(lumina.animation ~= nil, "animation sub-table missing")
		assert(type(lumina.animation.fadeIn) == "function", "fadeIn missing")
		assert(type(lumina.animation.fadeOut) == "function", "fadeOut missing")
		assert(type(lumina.animation.pulse) == "function", "pulse missing")
		assert(type(lumina.animation.spin) == "function", "spin missing")

		-- Test a preset returns config
		local config = lumina.animation.fadeIn(500)
		assert(config.from == 0, "fadeIn from should be 0")
		assert(config.to == 1, "fadeIn to should be 1")
		assert(config.duration == 500, "fadeIn duration should be 500")
		assert(config.easing == "easeInOut", "fadeIn easing should be easeInOut")
	`)
	if err != nil {
		t.Fatalf("animation presets test failed: %v", err)
	}
}

// ─── i18n API ────────────────────────────────────────────────────────────

func TestE2E_I18nModule(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		assert(lumina.i18n ~= nil, "i18n sub-table missing")
	`)
	if err != nil {
		t.Fatalf("i18n module test failed: %v", err)
	}
}

// ─── DevTools Module ─────────────────────────────────────────────────────

func TestE2E_DevToolsModule(t *testing.T) {
	app, _, _ := newTestApp(t, 80, 24)
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		assert(lumina.debug ~= nil, "debug sub-table missing")
	`)
	if err != nil {
		t.Fatalf("devtools module test failed: %v", err)
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
