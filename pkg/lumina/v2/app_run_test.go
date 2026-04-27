package v2

import (
	"os"
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
)

// --- NewAppWithLua tests ---

func TestNewAppWithLua_CreatesApp(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)
	if app == nil {
		t.Fatal("NewAppWithLua returned nil")
	}
	if app.luaState != L {
		t.Error("luaState not set")
	}
	if app.bridge == nil {
		t.Error("bridge not set")
	}
	if app.animMgr == nil {
		t.Error("animMgr not set")
	}
	if app.routerMgr == nil {
		t.Error("routerMgr not set")
	}
	if app.quit == nil {
		t.Error("quit channel not set")
	}
}

func TestNewAppWithLua_RegistersLuminaGlobal(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	_ = NewAppWithLua(L, 40, 10, ta)

	// Verify "lumina" global table exists with expected functions.
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		t.Fatal("lumina global is not a table")
	}
	tblIdx := L.AbsIndex(-1)

	for _, name := range []string{
		"createComponent", "removeComponent", "quit",
		"useState", "useEffect", "useMemo", "createElement",
		"useCallback", "useRef", "useReducer", "useId",
		"useLayoutEffect", "useAnimation",
		"navigate", "back", "useRoute",
	} {
		L.GetField(tblIdx, name)
		if !L.IsFunction(-1) {
			t.Errorf("lumina.%s is not a function (type: %s)", name, L.TypeName(L.Type(-1)))
		}
		L.Pop(1)
	}
	L.Pop(1) // pop lumina table
}

// --- createComponent via Lua ---

func TestLuaCreateComponent_RegistersComponent(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	err := app.RunString(`
		lumina.createComponent({
			id = "test-comp",
			name = "TestComp",
			x = 0, y = 0, w = 20, h = 5,
			zIndex = 0,
			render = function(state, props)
				return lumina.createElement("box", {id = "root"})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	// Verify component is registered.
	comp := app.manager.Get("test-comp")
	if comp == nil {
		t.Fatal("component 'test-comp' not registered")
	}
	if comp.Name() != "TestComp" {
		t.Errorf("name = %q, want %q", comp.Name(), "TestComp")
	}
	if comp.Rect().W != 20 || comp.Rect().H != 5 {
		t.Errorf("rect = %v, want W=20 H=5", comp.Rect())
	}
}

func TestLuaCreateComponent_RenderProducesVNode(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	err := app.RunString(`
		lumina.createComponent({
			id = "hello",
			x = 0, y = 0, w = 20, h = 3,
			render = function(state, props)
				return lumina.createElement("text", {id = "msg", content = "Hello"})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	// Render and check VNode tree.
	app.RenderAll()

	comp := app.manager.Get("hello")
	if comp == nil {
		t.Fatal("component not found")
	}
	vn := comp.VNodeTree()
	if vn == nil {
		t.Fatal("VNodeTree is nil after render")
	}
	if vn.Type != "text" {
		t.Errorf("VNode.Type = %q, want %q", vn.Type, "text")
	}
	if vn.Content != "Hello" {
		t.Errorf("VNode.Content = %q, want %q", vn.Content, "Hello")
	}
}

func TestLuaCreateComponent_MissingId_Errors(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	_ = NewAppWithLua(L, 40, 10, ta)

	err := L.DoString(`
		lumina.createComponent({
			x = 0, y = 0, w = 10, h = 5,
			render = function(state, props)
				return lumina.createElement("box", {})
			end
		})
	`)
	if err == nil {
		t.Error("expected error for missing id, got nil")
	}
}

func TestLuaCreateComponent_MissingRender_Errors(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	_ = NewAppWithLua(L, 40, 10, ta)

	err := L.DoString(`
		lumina.createComponent({
			id = "no-render",
			x = 0, y = 0, w = 10, h = 5,
		})
	`)
	if err == nil {
		t.Error("expected error for missing render, got nil")
	}
}

func TestLuaCreateComponent_InvalidDimensions_Errors(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	_ = NewAppWithLua(L, 40, 10, ta)

	err := L.DoString(`
		lumina.createComponent({
			id = "bad-dims",
			x = 0, y = 0, w = 0, h = 5,
			render = function(state, props)
				return lumina.createElement("box", {})
			end
		})
	`)
	if err == nil {
		t.Error("expected error for w=0, got nil")
	}
}

// --- removeComponent ---

func TestLuaRemoveComponent(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	err := app.RunString(`
		lumina.createComponent({
			id = "removeme",
			x = 0, y = 0, w = 10, h = 5,
			render = function(state, props)
				return lumina.createElement("box", {})
			end
		})
	`)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if app.manager.Get("removeme") == nil {
		t.Fatal("component should exist before removal")
	}

	err = app.RunString(`lumina.removeComponent("removeme")`)
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if app.manager.Get("removeme") != nil {
		t.Error("component should be nil after removal")
	}
}

// --- Event loop tests ---

func TestEventLoop_StopViaQuit(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	events := make(chan InputEvent)
	done := make(chan error, 1)

	go func() {
		done <- app.Run(RunConfig{
			Events:    events,
			FrameRate: 60,
		})
	}()

	// Give event loop time to start.
	time.Sleep(20 * time.Millisecond)

	if !app.IsRunning() {
		t.Error("expected app to be running")
	}

	app.Stop()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after Stop()")
	}

	if app.IsRunning() {
		t.Error("expected app to not be running after Stop")
	}
}

func TestEventLoop_StopViaChannelClose(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	events := make(chan InputEvent)
	done := make(chan error, 1)

	go func() {
		done <- app.Run(RunConfig{
			Events:    events,
			FrameRate: 60,
		})
	}()

	time.Sleep(20 * time.Millisecond)
	close(events)

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after closing events channel")
	}
}

func TestEventLoop_LuaQuit(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	events := make(chan InputEvent)
	done := make(chan error, 1)

	go func() {
		done <- app.Run(RunConfig{
			Events:    events,
			FrameRate: 60,
		})
	}()

	time.Sleep(20 * time.Millisecond)

	// Call lumina.quit() from Lua.
	err := L.DoString(`lumina.quit()`)
	if err != nil {
		t.Fatalf("lumina.quit() failed: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after lumina.quit()")
	}
}

func TestEventLoop_CtrlC_Quits(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	events := make(chan InputEvent, 1)
	done := make(chan error, 1)

	go func() {
		done <- app.Run(RunConfig{
			Events:    events,
			FrameRate: 60,
		})
	}()

	time.Sleep(20 * time.Millisecond)

	// Send Ctrl+C.
	events <- InputEvent{
		Type:      "keydown",
		Key:       "c",
		Modifiers: InputModifiers{Ctrl: true},
	}

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after Ctrl+C")
	}
}

// --- handleInputEvent tests ---

func TestHandleInputEvent_Keydown(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	// Track dispatched events.
	var dispatched []*event.Event
	app.dispatcher.RegisterHandlers("test-node", event.HandlerMap{
		"keydown": func(e *event.Event) {
			dispatched = append(dispatched, e)
		},
	})
	app.dispatcher.RegisterFocusable("test-node", 0)
	app.dispatcher.SetFocus("test-node")

	app.handleInputEvent(InputEvent{
		Type: "keydown",
		Key:  "Enter",
	})

	if len(dispatched) != 1 {
		t.Fatalf("expected 1 dispatched event, got %d", len(dispatched))
	}
	if dispatched[0].Key != "Enter" {
		t.Errorf("key = %q, want %q", dispatched[0].Key, "Enter")
	}
}

func TestHandleInputEvent_ShiftTab(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	// Register two focusables.
	app.dispatcher.RegisterFocusable("a", 0)
	app.dispatcher.RegisterFocusable("b", 1)
	app.dispatcher.SetFocus("b")

	app.handleInputEvent(InputEvent{
		Type:      "keydown",
		Key:       "Tab",
		Modifiers: InputModifiers{Shift: true},
	})

	// Shift+Tab should move focus backward.
	if app.FocusedID() != "a" {
		t.Errorf("focused = %q, want %q", app.FocusedID(), "a")
	}
}

func TestHandleInputEvent_Mouse(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	var dispatched []*event.Event
	app.dispatcher.RegisterHandlers("btn", event.HandlerMap{
		"mousedown": func(e *event.Event) {
			dispatched = append(dispatched, e)
		},
	})

	// No hit tester set, so the event won't hit "btn", but it should still
	// be processed without panic.
	app.handleInputEvent(InputEvent{
		Type: "mousedown",
		X:    5,
		Y:    3,
	})

	// No crash = pass. The event didn't hit any target because no hit tester.
	// This is expected — in real usage, RenderAll sets up the hit tester.
}

func TestHandleInputEvent_Resize(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	app.handleInputEvent(InputEvent{
		Type: "resize",
		X:    80,
		Y:    24,
	})

	if app.width != 80 || app.height != 24 {
		t.Errorf("size = %dx%d, want 80x24", app.width, app.height)
	}
}

// --- useState integration test ---

func TestLuaUseState_Integration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	err := app.RunString(`
		lumina.createComponent({
			id = "stateful",
			x = 0, y = 0, w = 20, h = 3,
			render = function(state, props)
				local count, setCount = lumina.useState("count", 42)
				return lumina.createElement("text", {
					id = "display",
					content = tostring(count),
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	comp := app.manager.Get("stateful")
	if comp == nil {
		t.Fatal("component not found")
	}
	vn := comp.VNodeTree()
	if vn == nil {
		t.Fatal("VNodeTree is nil")
	}
	if vn.Content != "42" {
		t.Errorf("content = %q, want %q", vn.Content, "42")
	}
}

// --- Accessors ---

func TestApp_Accessors(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	if app.Bridge() == nil {
		t.Error("Bridge() returned nil")
	}
	if app.AnimationManager() == nil {
		t.Error("AnimationManager() returned nil")
	}
	if app.RouterManager() == nil {
		t.Error("RouterManager() returned nil")
	}
	if app.IsRunning() {
		t.Error("should not be running before Run()")
	}
}

// --- Stop is idempotent ---

func TestStop_Idempotent(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	// Calling Stop multiple times should not panic.
	app.Stop()
	app.Stop()
	app.Stop()
}

// --- Hot Reload tests ---

func TestApp_HotReload_ReloadsScript(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	// Create a temp Lua script.
	dir := t.TempDir()
	scriptPath := dir + "/app.lua"
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "comp1",
			x = 0, y = 0, w = 20, h = 3,
			render = function(state, props)
				return lumina.createElement("text", {id = "t1", content = "version1"})
			end
		})
	`)

	// Load and render.
	if err := app.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	app.RenderAll()

	comp := app.manager.Get("comp1")
	if comp == nil {
		t.Fatal("comp1 not registered after initial load")
	}

	// Rewrite the script with different content.
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "comp1",
			x = 0, y = 0, w = 20, h = 3,
			render = function(state, props)
				return lumina.createElement("text", {id = "t1", content = "version2"})
			end
		})
	`)

	// Trigger hot reload.
	app.reloadScript(scriptPath)

	// Verify component still exists after reload.
	comp = app.manager.Get("comp1")
	if comp == nil {
		t.Fatal("comp1 not registered after reload")
	}

	// Verify the new render function is used by checking VNode content.
	vn := comp.VNodeTree()
	if vn == nil {
		t.Fatal("VNodeTree is nil after reload")
	}
	if vn.Content != "version2" {
		t.Errorf("expected content 'version2', got %q", vn.Content)
	}
}

func TestApp_HotReload_PreservesState(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	dir := t.TempDir()
	scriptPath := dir + "/counter.lua"
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "counter",
			x = 0, y = 0, w = 20, h = 3,
			render = function(state, props)
				local count, setCount = lumina.useState("count", 0)
				return lumina.createElement("text", {
					id = "display",
					content = "count=" .. tostring(count)
				})
			end
		})
	`)

	if err := app.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	app.RenderAll()

	// Set state to count=5.
	app.SetState("counter", "count", int64(5))
	app.RenderDirty()

	// Verify state was set.
	comp := app.manager.Get("counter")
	if comp == nil {
		t.Fatal("counter not registered")
	}
	if comp.State()["count"] != int64(5) {
		t.Fatalf("expected count=5, got %v", comp.State()["count"])
	}

	// Rewrite script (same structure, different render detail).
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "counter",
			x = 0, y = 0, w = 20, h = 3,
			render = function(state, props)
				local count, setCount = lumina.useState("count", 0)
				return lumina.createElement("text", {
					id = "display",
					content = "v2:count=" .. tostring(count)
				})
			end
		})
	`)

	// Hot reload.
	app.reloadScript(scriptPath)

	// Verify state was preserved.
	comp = app.manager.Get("counter")
	if comp == nil {
		t.Fatal("counter not registered after reload")
	}
	if comp.State()["count"] != int64(5) {
		t.Fatalf("expected count=5 after reload, got %v", comp.State()["count"])
	}

	// Verify new render function is used (content should start with "v2:").
	vn := comp.VNodeTree()
	if vn == nil {
		t.Fatal("VNodeTree is nil after reload")
	}
	if vn.Content != "v2:count=5" {
		t.Errorf("expected content 'v2:count=5', got %q", vn.Content)
	}
}

func TestApp_HotReload_ScriptError_DoesNotCrash(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 10, ta)

	dir := t.TempDir()
	scriptPath := dir + "/app.lua"
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "comp1",
			x = 0, y = 0, w = 20, h = 3,
			render = function(state, props)
				return lumina.createElement("text", {id = "t1", content = "good"})
			end
		})
	`)

	if err := app.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	app.RenderAll()

	// Write a script with syntax error.
	writeFile(t, scriptPath, `
		this is not valid lua!!!
	`)

	// reloadScript should not panic.
	app.reloadScript(scriptPath)

	// After a failed reload, components are gone (unregistered before re-exec).
	// This is expected — the user fixes the script and it reloads again.
}

func TestApp_HotReload_DevToolsPreserved(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewAppWithLua(L, 40, 20, ta)

	dir := t.TempDir()
	scriptPath := dir + "/app.lua"
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "mycomp",
			x = 0, y = 0, w = 20, h = 5,
			render = function(state, props)
				return lumina.createElement("box", {id = "root"})
			end
		})
	`)

	if err := app.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	app.RenderAll()

	// Open devtools.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"})

	// Verify devtools component exists.
	if app.manager.Get("__devtools") == nil {
		t.Fatal("devtools not registered after F12")
	}

	// Reload.
	app.reloadScript(scriptPath)

	// Devtools should still be registered.
	if app.manager.Get("__devtools") == nil {
		t.Fatal("devtools was removed during hot reload")
	}

	// App component should also exist.
	if app.manager.Get("mycomp") == nil {
		t.Fatal("mycomp not registered after reload")
	}
}

// writeFile is a test helper that writes content to a file.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", path, err)
	}
}
