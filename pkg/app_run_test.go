package v2

import (
	"os"
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
)

// --- NewApp tests ---

func TestNewApp_CreatesApp(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewApp(L, 40, 10, ta)
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.luaState != L {
		t.Error("luaState not set")
	}
	if app.engine == nil {
		t.Error("engine not set")
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

func TestNewApp_RegistersLuminaGlobal(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	_ = NewApp(L, 40, 10, ta)

	// Verify "lumina" global table exists with expected functions.
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		t.Fatal("lumina global is not a table")
	}
	tblIdx := L.AbsIndex(-1)

	for _, name := range []string{
		"createComponent", "createElement", "defineComponent",
		"useState", "quit",
		"setInterval", "setTimeout", "clearInterval", "clearTimeout",
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

	app := NewApp(L, 40, 10, ta)

	err := app.RunString(`
		lumina.createComponent({
			id = "test-comp",
			render = function(props)
				return lumina.createElement("box", {id = "root"})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	// Verify component is registered via engine.
	comp := app.engine.GetComponent("test-comp")
	if comp == nil {
		t.Fatal("component 'test-comp' not registered")
	}
}

func TestLuaCreateComponent_RenderProducesOutput(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewApp(L, 40, 10, ta)

	err := app.RunString(`
		lumina.createComponent({
			id = "hello",
			render = function(props)
				return lumina.createElement("text", {content = "Hello"})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	// Render and check output was produced.
	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("no screen output after RenderAll")
	}
}

// --- Event loop tests ---

func TestEventLoop_StopViaQuit(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewApp(L, 40, 10, ta)

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

	app := NewApp(L, 40, 10, ta)

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

	app := NewApp(L, 40, 10, ta)

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

	app := NewApp(L, 40, 10, ta)

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

func TestHandleInputEvent_Resize(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewApp(L, 40, 10, ta)

	app.handleInputEvent(InputEvent{
		Type: "resize",
		X:    80,
		Y:    24,
	})

	if app.width != 80 || app.height != 24 {
		t.Errorf("size = %dx%d, want 80x24", app.width, app.height)
	}
}

func TestHandleInputEvent_Keydown(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewApp(L, 40, 10, ta)

	// Dispatch a keydown event — should not panic even without components.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
}

func TestHandleInputEvent_Mouse(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewApp(L, 40, 10, ta)

	// Dispatch a mouse event — should not panic even without components.
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 3})
}

// --- Accessors ---

func TestApp_Accessors(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewApp(L, 40, 10, ta)

	if app.Engine() == nil {
		t.Error("Engine() returned nil")
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

	app := NewApp(L, 40, 10, ta)

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

	app := NewApp(L, 40, 10, ta)

	// Create a temp Lua script.
	dir := t.TempDir()
	scriptPath := dir + "/app.lua"
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "comp1",
			render = function(props)
				return lumina.createElement("text", {content = "version1"})
			end
		})
	`)

	// Load and render.
	if err := app.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	app.RenderAll()

	comp := app.engine.GetComponent("comp1")
	if comp == nil {
		t.Fatal("comp1 not registered after initial load")
	}

	// Rewrite the script with different content.
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "comp1",
			render = function(props)
				return lumina.createElement("text", {content = "version2"})
			end
		})
	`)

	// Trigger hot reload.
	app.reloadScript(scriptPath)

	// Verify component still exists after reload.
	comp = app.engine.GetComponent("comp1")
	if comp == nil {
		t.Fatal("comp1 not registered after reload")
	}
}

func TestApp_HotReload_ScriptError_DoesNotCrash(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()

	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	scriptPath := dir + "/app.lua"
	writeFile(t, scriptPath, `
		lumina.createComponent({
			id = "comp1",
			render = function(props)
				return lumina.createElement("text", {content = "good"})
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
}

// writeFile is a test helper that writes content to a file.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", path, err)
	}
}
