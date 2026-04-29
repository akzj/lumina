package v2

import (
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

// TestSlotFactory_Basic tests that slot factories produce correct descriptor tables.
func TestSlotFactory_Basic(t *testing.T) {
	app, _, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local Slot = require("lux.slot")
		local Title = Slot("title")

		local result = Title { "Hello World" }
		assert(result.type == "_slot", "type should be _slot, got: " .. tostring(result.type))
		assert(result._slotName == "title", "_slotName should be title")
		assert(#result.children == 1, "should have 1 child")
		assert(result.children[1] == "Hello World", "child should be 'Hello World'")
	`)
	if err != nil {
		t.Fatalf("Slot factory test failed: %v", err)
	}
}

// TestSlotFactory_WithProps tests that slot factories handle mixed props + children.
func TestSlotFactory_WithProps(t *testing.T) {
	app, _, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local Slot = require("lux.slot")
		local Actions = Slot("actions")

		local child1 = { type = "text", content = "btn1" }
		local child2 = { type = "text", content = "btn2" }
		local result = Actions { align = "right", child1, child2 }

		assert(result.type == "_slot", "type should be _slot")
		assert(result._slotName == "actions", "_slotName should be actions")
		assert(#result.children == 2, "should have 2 children, got: " .. tostring(#result.children))
		assert(result.props.align == "right", "props.align should be right")
	`)
	if err != nil {
		t.Fatalf("Slot factory with props test failed: %v", err)
	}
}

// TestComposableDialog_BasicSlots tests the composable Dialog with Title, Content, and Actions slots.
func TestComposableDialog_BasicSlots(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local lux = require("lux")
		local Dialog = lux.Dialog
		local Title, Content, Actions = Dialog.Title, Dialog.Content, Dialog.Actions

		lumina.createComponent({
			id = "dlg-test",
			name = "DlgTest",
			x = 0, y = 0, w = 80, h = 24,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 24},
				},
					Dialog {
						open = true,
						width = 40,

						Title { "Confirm Delete" },
						Content { "Are you sure?" },
						Actions {
							lumina.createElement("text", {}, "[Cancel]"),
							lumina.createElement("text", {}, "[Delete]"),
						},
					}
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	// Debug: dump screen
	if !screenHasString(ta, "Confirm Delete") {
		for y := 0; y < 24; y++ {
			line := readScreenLine(ta, y, 80)
			if line != "" {
				t.Logf("screen[%d]: %q", y, line)
			}
		}
		t.Error("expected 'Confirm Delete' title on screen")
	}
	if !screenHasString(ta, "Are you sure?") {
		t.Error("expected 'Are you sure?' content on screen")
	}
	if !screenHasString(ta, "[Cancel]") {
		t.Error("expected '[Cancel]' action on screen")
	}
	if !screenHasString(ta, "[Delete]") {
		t.Error("expected '[Delete]' action on screen")
	}
}

func findTextNodeContaining(n *render.Node, sub string) *render.Node {
	if n == nil {
		return nil
	}
	if n.Type == "text" && strings.Contains(n.Content, sub) {
		return n
	}
	for _, c := range n.Children {
		if hit := findTextNodeContaining(c, sub); hit != nil {
			return hit
		}
	}
	return nil
}

// TestComposableDialog_ActionsTextOnClickNode asserts the OK action text node
// carries a Lua onClick ref after render (regression for empty-key reconcile / hit-test).
func TestComposableDialog_ActionsTextOnClickNode(t *testing.T) {
	app, _, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local lux = require("lux")
		local Dialog = lux.Dialog
		local Title, Content, Actions = Dialog.Title, Dialog.Content, Dialog.Actions

		lumina.createComponent({
			id = "dlg-okref",
			name = "DlgOkRef",
			x = 0, y = 0, w = 80, h = 24,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 24},
				},
					Dialog {
						open = true,
						width = 40,
						Title { "T" },
						Content { "Body" },
						Actions {
							lumina.createElement("text", {
								foreground = "#89B4FA",
								bold = true,
								onClick = function() end,
							}, "  [ OK ]  "),
						},
					}
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	root := app.Engine().Root()
	if root == nil || root.RootNode == nil {
		t.Fatal("missing root RootNode")
	}
	tn := findTextNodeContaining(root.RootNode, "[ OK ]")
	if tn == nil {
		t.Fatal("expected a text node containing [ OK ]")
	}
	if tn.OnClick == 0 {
		t.Fatalf("OK action text node should have OnClick ref, got 0 (content=%q)", tn.Content)
	}
}

// TestComposableDialog_TitlePropFallback tests that the title prop is used when no Title slot.
func TestComposableDialog_TitlePropFallback(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local lux = require("lux")
		local Dialog = lux.Dialog

		lumina.createComponent({
			id = "dlg-fallback",
			name = "DlgFallback",
			x = 0, y = 0, w = 80, h = 24,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 24},
				},
					Dialog {
						open = true,
						title = "Simple Dialog",
						message = "Hello World",
						width = 40,
					}
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "Simple Dialog") {
		for y := 0; y < 24; y++ {
			line := readScreenLine(ta, y, 80)
			if line != "" {
				t.Logf("screen[%d]: %q", y, line)
			}
		}
		t.Error("expected 'Simple Dialog' title from prop")
	}
	if !screenHasString(ta, "Hello World") {
		t.Error("expected 'Hello World' message from prop")
	}
}

// TestComposableDialog_Closed tests that Dialog with open=false does not render.
func TestComposableDialog_Closed(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local lux = require("lux")
		local Dialog = lux.Dialog

		lumina.createComponent({
			id = "dlg-closed",
			name = "DlgClosed",
			x = 0, y = 0, w = 80, h = 24,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 24},
				},
					lumina.createElement("text", {}, "Background"),
					Dialog {
						open = false,
						title = "Hidden",
					}
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "Background") {
		t.Error("expected 'Background' text on screen")
	}
	if screenHasString(ta, "Hidden") {
		t.Error("Dialog should not be visible when open=false")
	}
}

// TestComposableDialog_RequireDialog tests that require("lux.dialog") works directly.
func TestComposableDialog_RequireDialog(t *testing.T) {
	app, _, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local Dialog = require("lux.dialog")
		assert(Dialog ~= nil, "require('lux.dialog') returned nil")
		assert(Dialog.Title ~= nil, "Dialog.Title missing")
		assert(Dialog.Content ~= nil, "Dialog.Content missing")
		assert(Dialog.Actions ~= nil, "Dialog.Actions missing")
	`)
	if err != nil {
		t.Fatalf("require lux.dialog failed: %v", err)
	}
}

// TestComposableDialog_ContentOnly tests Dialog with only Content slot (no Title/Actions).
func TestComposableDialog_ContentOnly(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local Dialog = require("lux.dialog")
		local Content = Dialog.Content

		lumina.createComponent({
			id = "dlg-content-only",
			name = "DlgContentOnly",
			x = 0, y = 0, w = 80, h = 24,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 24},
				},
					Dialog {
						open = true,
						title = "Info",
						width = 40,
						Content { "Some information here" },
					}
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "Info") {
		for y := 0; y < 24; y++ {
			line := readScreenLine(ta, y, 80)
			if line != "" {
				t.Logf("screen[%d]: %q", y, line)
			}
		}
		t.Error("expected 'Info' title on screen")
	}
	if !screenHasString(ta, "Some information here") {
		t.Error("expected content text on screen")
	}
}
