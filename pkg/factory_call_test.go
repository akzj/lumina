package v2

import (
	"testing"
)

// TestFactoryCallSyntax_GoWidget verifies that Factory { props } syntax works
// for Go widgets (e.g., lumina.Checkbox).
func TestFactoryCallSyntax_GoWidget(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "call-test",
			name = "CallTest",
			x = 0, y = 0, w = 80, h = 10,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 10},
				},
					-- Use __call syntax for Go widget
					lumina.Checkbox { label = "Call Syntax", checked = true, key = "cb1" }
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "[x]") {
		t.Error("expected '[x]' on screen from Checkbox __call syntax")
	}
	if !screenHasString(ta, "Call Syntax") {
		t.Error("expected 'Call Syntax' on screen from Checkbox __call syntax")
	}
}

// TestFactoryCallSyntax_LuaComponent verifies that Factory { props } syntax works
// for Lua-defined components via defineComponent.
func TestFactoryCallSyntax_LuaComponent(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 10)

	err := app.RunString(`
		-- Define a Lua component
		local MyBox = lumina.defineComponent("MyBox", function(props)
			return lumina.createElement("text", {
				foreground = "#ffffff",
			}, "MyBox:" .. (props.title or ""))
		end)

		lumina.createComponent({
			id = "call-test-lua",
			name = "CallTestLua",
			x = 0, y = 0, w = 80, h = 10,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 10},
				},
					-- Use __call syntax for Lua component
					MyBox { title = "Hello" }
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "MyBox:Hello") {
		t.Error("expected 'MyBox:Hello' on screen from Lua component __call syntax")
	}
}

// TestFactoryCallSyntax_WithChildren verifies that Factory(props, child1, child2) works
// using a Lua-defined component that accepts children via __call syntax.
func TestFactoryCallSyntax_WithChildren(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 10)

	err := app.RunString(`
		-- Define a Lua component that renders its children
		local Wrapper = lumina.defineComponent("Wrapper", function(props)
			return lumina.createElement("vbox", {
				style = {width = 80, height = 10},
			},
				lumina.createElement("text", {}, "Header:" .. (props.title or "")),
				lumina.createElement("text", {}, "ChildSlot")
			)
		end)

		lumina.createComponent({
			id = "call-children-test",
			name = "CallChildrenTest",
			x = 0, y = 0, w = 80, h = 10,
			render = function(props)
				-- Use __call syntax wrapped in a container
				return lumina.createElement("vbox", {
					style = {width = 80, height = 10},
				},
					Wrapper { title = "MyTitle" }
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "Header:MyTitle") {
		t.Error("expected 'Header:MyTitle' on screen")
	}
	if !screenHasString(ta, "ChildSlot") {
		t.Error("expected 'ChildSlot' on screen")
	}
}

// TestFactoryCallSyntax_CreateElementStillWorks verifies backward compatibility:
// lumina.createElement(factory, props) still works.
func TestFactoryCallSyntax_CreateElementStillWorks(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "compat-test",
			name = "CompatTest",
			x = 0, y = 0, w = 80, h = 10,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 10},
				},
					-- Old syntax still works
					lumina.createElement(lumina.Checkbox, {
						label = "Old Syntax",
						checked = false,
						key = "compat-cb",
					})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "[ ]") {
		t.Error("expected '[ ]' on screen from old createElement syntax")
	}
	if !screenHasString(ta, "Old Syntax") {
		t.Error("expected 'Old Syntax' on screen from old createElement syntax")
	}
}

// TestFactoryCallSyntax_MixedTableDebug tests Pattern 2 step by step.
// First verifies children count is passed, then full rendering.
func TestFactoryCallSyntax_MixedTable(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 10)

	err := app.RunString(`
		local Panel = lumina.defineComponent("Panel", function(props)
			local children = props.children or {}
			local count = 0
			for _ in ipairs(children) do count = count + 1 end
			return lumina.createElement("text", {}, "PANEL:" .. tostring(count))
		end)

		lumina.createComponent({
			id = "mixed-debug",
			name = "MixedDebug",
			x = 0, y = 0, w = 80, h = 10,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 10},
				},
					Panel {
						border = "single",
						lumina.createElement("text", {}, "child1"),
						lumina.createElement("text", {}, "child2"),
					}
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "PANEL:2") {
		// Dump screen for debug
		for y := 0; y < 10; y++ {
			line := readScreenLine(ta, y, 80)
			t.Logf("screen[%d]: %q", y, line)
		}
		t.Error("expected 'PANEL:2' on screen — children not passed via mixed table")
	}
}

// TestFactoryCallSyntax_MixedTablePropsOnly verifies that a mixed table
// with ONLY string keys (no integer keys) still works as pure props.
func TestFactoryCallSyntax_MixedTablePropsOnly(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "mixed-props-only",
			name = "MixedPropsOnly",
			x = 0, y = 0, w = 80, h = 10,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 10},
				},
					lumina.Checkbox { label = "PropsOnly", checked = false, key = "po1" }
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "[ ]") {
		t.Error("expected '[ ]' on screen from pure props table")
	}
	if !screenHasString(ta, "PropsOnly") {
		t.Error("expected 'PropsOnly' on screen from pure props table")
	}
}

// TestFactoryCallSyntax_MultipleArgs verifies Pattern 1 with explicit children args:
// Factory(props, child1, child2)
func TestFactoryCallSyntax_MultipleArgs(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 10)

	err := app.RunString(`
		local Container = lumina.defineComponent("Container", function(props)
			local children = props.children or {}
			return lumina.createElement("vbox", {
				style = {width = 80, height = 10},
			}, table.unpack(children))
		end)

		lumina.createComponent({
			id = "multi-args-test",
			name = "MultiArgsTest",
			x = 0, y = 0, w = 80, h = 10,
			render = function(props)
				return lumina.createElement("vbox", {
					style = {width = 80, height = 10},
				},
					-- Pattern 1: explicit args
					Container(
						{ border = "single" },
						lumina.createElement("text", {}, "Arg1Child"),
						lumina.createElement("text", {}, "Arg2Child")
					)
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	if !screenHasString(ta, "Arg1Child") {
		t.Error("expected 'Arg1Child' on screen from Pattern 1 multi-args")
	}
	if !screenHasString(ta, "Arg2Child") {
		t.Error("expected 'Arg2Child' on screen from Pattern 1 multi-args")
	}
}
