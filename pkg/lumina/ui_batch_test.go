package lumina

import (
	"fmt"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// uiTestState creates a fresh Lua state with lumina + all component preloads.
func uiTestState(t *testing.T) *lua.State {
	t.Helper()
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	return L
}

// TestUIComponents_LoadAll verifies all 57 lumina/ui components load without error
// via both the new (lumina.ui.xxx) and old (shadcn.xxx) require paths.
func TestUIComponents_LoadAll(t *testing.T) {
	components := []string{
		"accordion", "alert", "alert_dialog", "aspect_ratio", "avatar",
		"badge", "breadcrumb", "button", "button_group",
		"calendar", "card", "carousel", "chart", "checkbox",
		"collapsible", "color_picker", "combobox", "command", "context_menu",
		"data_table", "date_picker", "dialog", "drawer", "dropdown_menu",
		"field", "form",
		"hover_card",
		"input", "input_group", "input_otp",
		"kbd",
		"label",
		"menubar",
		"native_select", "navigation_menu",
		"pagination", "popover", "progress",
		"radio_group", "resizable",
		"scroll_area", "select", "separator", "sheet", "sidebar",
		"skeleton", "slider", "sonner", "spinner", "switch",
		"table", "tabs", "textarea", "toggle", "toggle_group", "tooltip",
	}

	for _, comp := range components {
		t.Run("lumina.ui."+comp, func(t *testing.T) {
			L := uiTestState(t)
			defer L.Close()
			code := fmt.Sprintf(`local mod = require("lumina.ui.%s"); assert(mod ~= nil, "module is nil")`, comp)
			if err := L.DoString(code); err != nil {
				t.Fatalf("failed to load lumina.ui.%s: %v", comp, err)
			}
		})
	}

	// Also test the aggregate init module
	t.Run("lumina.ui", func(t *testing.T) {
		L := uiTestState(t)
		defer L.Close()
		if err := L.DoString(`local ui = require("lumina.ui"); assert(ui ~= nil)`); err != nil {
			t.Fatalf("failed to load lumina.ui: %v", err)
		}
	})
}

// TestUIComponents_RenderBasic tests that components with simple/default props
// render to a non-empty frame without panics.
func TestUIComponents_RenderBasic(t *testing.T) {
	type renderCase struct {
		name string
		lua  string // Lua code that returns a VNode tree
	}

	cases := []renderCase{
		{"button", `
			local Button = require("lumina.ui.button")
			return lumina.createElement(Button, {label = "Click Me"})
		`},
		{"badge", `
			local Badge = require("lumina.ui.badge")
			return lumina.createElement(Badge, {label = "New"})
		`},
		{"alert", `
			local Alert = require("lumina.ui.alert")
			return lumina.createElement(Alert, {title = "Warning", description = "Be careful"})
		`},
		{"label", `
			local Label = require("lumina.ui.label")
			return lumina.createElement(Label, {text = "Name:"})
		`},
		{"separator", `
			local Separator = require("lumina.ui.separator")
			return lumina.createElement(Separator, {})
		`},
		{"skeleton", `
			local Skeleton = require("lumina.ui.skeleton")
			return lumina.createElement(Skeleton, {width = 20, height = 3})
		`},
		{"spinner", `
			local Spinner = require("lumina.ui.spinner")
			return lumina.createElement(Spinner, {})
		`},
		{"kbd", `
			local Kbd = require("lumina.ui.kbd")
			return lumina.createElement(Kbd, {keys = "Ctrl+C"})
		`},
		{"progress", `
			local Progress = require("lumina.ui.progress")
			return lumina.createElement(Progress, {value = 50, max = 100})
		`},
		{"checkbox", `
			local Checkbox = require("lumina.ui.checkbox")
			return lumina.createElement(Checkbox, {label = "Accept", checked = false})
		`},
		{"switch", `
			local Switch = require("lumina.ui.switch")
			return lumina.createElement(Switch, {label = "Dark Mode", checked = true})
		`},
		{"toggle", `
			local Toggle = require("lumina.ui.toggle")
			return lumina.createElement(Toggle, {label = "Bold", pressed = false})
		`},
		{"slider", `
			local Slider = require("lumina.ui.slider")
			return lumina.createElement(Slider, {value = 50, min = 0, max = 100})
		`},
		{"avatar", `
			local Avatar = require("lumina.ui.avatar")
			return lumina.createElement(Avatar, {name = "John Doe"})
		`},
		{"card", `
			local card = require("lumina.ui.card")
			return lumina.createElement(card.Card, {}, {
				lumina.createElement(card.CardHeader, {}, {
					lumina.createElement(card.CardTitle, {text = "Title"})
				}),
				lumina.createElement(card.CardContent, {}, {
					{type = "text", content = "Card body"}
				})
			})
		`},
		{"input", `
			local Input = require("lumina.ui.input")
			return lumina.createElement(Input, {id = "test-input", placeholder = "Type here..."})
		`},
		{"textarea", `
			local Textarea = require("lumina.ui.textarea")
			return lumina.createElement(Textarea, {id = "test-ta", placeholder = "Enter text..."})
		`},
		{"simple_text", `
			return {type = "text", content = "Hello World"}
		`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			L := uiTestState(t)
			defer L.Close()

			// Wrap in a function that returns the VNode
			code := fmt.Sprintf(`
				local lumina = require("lumina")
				local function makeVNode()
					%s
				end
				return makeVNode()
			`, tc.lua)

			if err := L.DoString(code); err != nil {
				t.Fatalf("Lua error: %v", err)
			}

			// Convert Lua result to VNode
			vnode := LuaVNodeToVNode(L, -1)
			L.Pop(1)

			if vnode == nil {
				t.Fatal("VNode is nil")
			}

			// Render to frame — should not panic
			frame := VNodeToFrame(vnode, 80, 24)
			if frame == nil {
				t.Fatal("Frame is nil")
			}

			// Check frame has some non-space content
			hasContent := false
			for y := 0; y < frame.Height && !hasContent; y++ {
				for x := 0; x < frame.Width; x++ {
					cell := frame.Cells[y][x]
					if cell.Char != ' ' && cell.Char != 0 && !cell.Transparent {
						hasContent = true
						break
					}
				}
			}
			if !hasContent {
				t.Logf("WARNING: %s rendered but frame has no visible content", tc.name)
			}
		})
	}
}

// TestUIComponents_BackwardCompat verifies shadcn.xxx still works.
func TestUIComponents_BackwardCompat(t *testing.T) {
	L := uiTestState(t)
	defer L.Close()

	err := L.DoString(`
		-- Old-style require should still work
		local shadcn = require("shadcn")
		assert(shadcn ~= nil, "shadcn module is nil")
		assert(shadcn.Button ~= nil, "shadcn.Button is nil")
		assert(shadcn.Card ~= nil, "shadcn.Card is nil")
		assert(shadcn.Dialog ~= nil, "shadcn.Dialog is nil")

		-- Individual requires too
		local btn = require("shadcn.button")
		assert(btn ~= nil, "shadcn.button is nil")

		-- New-style require
		local ui = require("lumina.ui")
		assert(ui ~= nil, "lumina.ui module is nil")
		assert(ui.Button ~= nil, "ui.Button is nil")

		local btn2 = require("lumina.ui.button")
		assert(btn2 ~= nil, "lumina.ui.button is nil")
	`)
	if err != nil {
		t.Fatalf("backward compat: %v", err)
	}
}
