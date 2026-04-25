package lumina_test

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

// ─── Event Bridge Tests ──────────────────────────────────────────────

func TestEventBridge_OnClickRegistered(t *testing.T) {
	// Render a component with onClick, verify handler is registered in EventBus
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		_G._clicked = false
		local App = lumina.defineComponent({
			name = "ClickTest",
			render = function(self)
				return {
					type = "button",
					id = "btn1",
					content = "Click Me",
					onClick = function()
						_G._clicked = true
					end,
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Render triggers the bridge
	app.RenderOnce()

	// Verify btn1 is registered as focusable
	ids := lumina.GetGlobalEventBus().GetFocusableIDs()
	found := false
	for _, id := range ids {
		if id == "btn1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("btn1 should be registered as focusable, got IDs: %v", ids)
	}
}

func TestEventBridge_FocusableRegistered(t *testing.T) {
	// Render button/input elements, verify they're registered as focusable
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "FocusTest",
			render = function(self)
				return {
					type = "vbox",
					children = {
						{ type = "button", id = "btn_a", content = "Button A" },
						{ type = "input", id = "input_b", content = "" },
						{ type = "text", id = "label_c", content = "Not focusable" },
						{ type = "select", id = "sel_d", content = "Select" },
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	eb := lumina.GetGlobalEventBus()

	// button, input, select should be focusable
	if !eb.IsFocusable("btn_a") {
		t.Error("btn_a (button) should be focusable")
	}
	if !eb.IsFocusable("input_b") {
		t.Error("input_b (input) should be focusable")
	}
	if !eb.IsFocusable("sel_d") {
		t.Error("sel_d (select) should be focusable")
	}
	// text should NOT be focusable (unless it has an onClick)
	if eb.IsFocusable("label_c") {
		t.Error("label_c (text) should NOT be focusable")
	}
}

func TestEventBridge_KeyBindingDispatched(t *testing.T) {
	// Register onKey binding, simulate keypress via emitKeyEvent, verify called
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		_G._key_pressed = false

		lumina.onKey("x", function()
			_G._key_pressed = true
		end)

		local App = lumina.defineComponent({
			name = "KeyTest",
			render = function(self)
				return { type = "text", content = "Press x" }
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	// Simulate key press via Lua emitKeyEvent
	err = app.L.DoString(`
		local lumina = require("lumina")
		lumina.emitKeyEvent("x")
	`)
	if err != nil {
		t.Fatalf("emitKeyEvent failed: %v", err)
	}

	// Process the lua_callback event that was posted
	app.ProcessPendingEvents()

	// Check if the key handler was called
	err = app.L.DoString(`
		assert(_G._key_pressed == true, "key handler should have been called, got: " .. tostring(_G._key_pressed))
	`)
	if err != nil {
		t.Fatalf("key binding was not dispatched: %v", err)
	}
}

func TestEventBridge_ClearOnRerender(t *testing.T) {
	// Re-render, verify old bridged handlers are cleared and new ones registered
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "RerenderTest",
			render = function(self)
				return {
					type = "vbox",
					children = {
						{ type = "button", id = "btn_rerender", content = "Click",
						  onClick = function() end },
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// First render
	app.RenderOnce()

	eb := lumina.GetGlobalEventBus()
	if !eb.IsFocusable("btn_rerender") {
		t.Fatal("btn_rerender should be focusable after first render")
	}

	// Second render (should clear and re-register)
	app.RenderOnce()

	if !eb.IsFocusable("btn_rerender") {
		t.Error("btn_rerender should still be focusable after re-render")
	}
}

func TestEventBridge_AutoGenerateID(t *testing.T) {
	// Elements with onClick but no ID should get auto-generated IDs
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "AutoIDTest",
			render = function(self)
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "Clickable",
						  onClick = function() end },
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	// Should have at least one auto-generated focusable ID
	eb := lumina.GetGlobalEventBus()
	ids := eb.GetFocusableIDs()
	hasAuto := false
	for _, id := range ids {
		if len(id) > 5 && id[:5] == "auto_" {
			hasAuto = true
			break
		}
	}
	if !hasAuto {
		t.Errorf("expected auto-generated focusable ID, got: %v", ids)
	}
}

func TestEventBridge_LuaFuncRefStored(t *testing.T) {
	// Verify that Lua functions in VNode props are stored as LuaFuncRef, not nil
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "FuncRefTest",
			render = function(self)
				return {
					type = "button",
					id = "funcref_btn",
					content = "Test",
					onClick = function() end,
					onChange = function() end,
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	// The button should be focusable (onClick was registered)
	eb := lumina.GetGlobalEventBus()
	if !eb.IsFocusable("funcref_btn") {
		t.Error("funcref_btn should be focusable (onClick handler should be registered)")
	}
}

func TestEventBridge_ClickEmitTriggersHandler(t *testing.T) {
	// Full flow: render with onClick → emit click event → verify Lua callback runs
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		_G._click_count = 0
		local App = lumina.defineComponent({
			name = "ClickEmitTest",
			render = function(self)
				return {
					type = "button",
					id = "emit_btn",
					content = "Click Me",
					onClick = function()
						_G._click_count = _G._click_count + 1
					end,
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	// Emit a click event targeting the button
	err = app.L.DoString(`
		local lumina = require("lumina")
		lumina.emit("emit_btn", "click")
	`)
	if err != nil {
		t.Fatalf("emitEvent failed: %v", err)
	}

	// Process the lua_callback event
	app.ProcessPendingEvents()

	// Verify the click handler was called
	err = app.L.DoString(`
		assert(_G._click_count == 1, "click handler should have been called once, got: " .. tostring(_G._click_count))
	`)
	if err != nil {
		t.Fatalf("click handler was not called: %v", err)
	}
}
