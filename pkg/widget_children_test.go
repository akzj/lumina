package v2

import "testing"

func TestWidgetChildren_DialogWithChildren(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local Dialog = require("lux.dialog")
		lumina.createComponent({
			id = "dialog-children",
			name = "DialogChildren",
			x = 0, y = 0, w = 80, h = 24,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 24},
				},
					Dialog {
						open = true,
						title = "Confirm",
						width = 40,
						key = "dlg1",
						Dialog.Content {
							lumina.createElement("text", {}, "Are you sure?"),
							lumina.createElement("text", {}, "This cannot be undone."),
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

	if !screenHasString(ta, "Confirm") {
		t.Error("expected 'Confirm' title on screen")
	}
	if !screenHasString(ta, "Are you sure?") {
		t.Error("expected 'Are you sure?' child on screen")
	}
	if !screenHasString(ta, "This cannot be undone.") {
		t.Error("expected 'This cannot be undone.' child on screen")
	}
}

func TestWidgetChildren_DialogWithoutChildren(t *testing.T) {
	// Verify existing Dialog API still works (message prop, no children)
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local Dialog = require("lux.dialog")
		lumina.createComponent({
			id = "dialog-msg",
			name = "DialogMsg",
			x = 0, y = 0, w = 80, h = 24,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 24},
				},
					Dialog {
						open = true,
						title = "Info",
						message = "Hello World",
						width = 40,
						key = "dlg2",
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
		t.Error("expected 'Info' title on screen")
	}
	if !screenHasString(ta, "Hello World") {
		t.Error("expected 'Hello World' message on screen")
	}
}

func TestWidgetChildren_DialogChildrenReplaceMessage(t *testing.T) {
	// When Content slot is provided, it should replace the message prop
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		local Dialog = require("lux.dialog")
		lumina.createComponent({
			id = "dialog-replace",
			name = "DialogReplace",
			x = 0, y = 0, w = 80, h = 24,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 24},
				},
					Dialog {
						open = true,
						title = "Replace Test",
						message = "Should NOT appear",
						width = 40,
						key = "dlg3",
						Dialog.Content {
							lumina.createElement("text", {}, "Custom Content"),
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

	if !screenHasString(ta, "Replace Test") {
		t.Error("expected 'Replace Test' title on screen")
	}
	if !screenHasString(ta, "Custom Content") {
		t.Error("expected 'Custom Content' child on screen")
	}
	if screenHasString(ta, "Should NOT appear") {
		t.Error("message prop should be replaced by Content slot")
	}
}
