package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// shadcnState creates a fresh Lua state with lumina + shadcn preloads.
func shadcnState(t *testing.T) *lua.State {
	t.Helper()
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	return L
}

func TestShadcn_Button(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Button = require("shadcn.button")
		assert(Button ~= nil, "Button should not be nil")
		assert(Button.isComponent == true, "Button should be a component")
		assert(Button.name == "ShadcnButton", "name should be ShadcnButton, got " .. tostring(Button.name))
	`)
	if err != nil {
		t.Fatalf("Button: %v", err)
	}
}

func TestShadcn_Badge(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Badge = require("shadcn.badge")
		assert(Badge ~= nil)
		assert(Badge.isComponent == true)
		assert(Badge.name == "ShadcnBadge")
	`)
	if err != nil {
		t.Fatalf("Badge: %v", err)
	}
}

func TestShadcn_Card(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local card = require("shadcn.card")
		assert(card ~= nil)
		assert(card.Card ~= nil, "Card should exist")
		assert(card.Card.isComponent == true)
		assert(card.CardHeader ~= nil)
		assert(card.CardTitle ~= nil)
		assert(card.CardDescription ~= nil)
		assert(card.CardContent ~= nil)
		assert(card.CardFooter ~= nil)
	`)
	if err != nil {
		t.Fatalf("Card: %v", err)
	}
}

func TestShadcn_Alert(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Alert = require("shadcn.alert")
		assert(Alert ~= nil)
		assert(Alert.isComponent == true)
		assert(Alert.name == "ShadcnAlert")
	`)
	if err != nil {
		t.Fatalf("Alert: %v", err)
	}
}

func TestShadcn_Label(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Label = require("shadcn.label")
		assert(Label ~= nil)
		assert(Label.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Label: %v", err)
	}
}

func TestShadcn_Separator(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Separator = require("shadcn.separator")
		assert(Separator ~= nil)
		assert(Separator.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Separator: %v", err)
	}
}

func TestShadcn_Skeleton(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Skeleton = require("shadcn.skeleton")
		assert(Skeleton ~= nil)
		assert(Skeleton.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Skeleton: %v", err)
	}
}

func TestShadcn_Spinner(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Spinner = require("shadcn.spinner")
		assert(Spinner ~= nil)
		assert(Spinner.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Spinner: %v", err)
	}
}

func TestShadcn_Avatar(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Avatar = require("shadcn.avatar")
		assert(Avatar ~= nil)
		assert(Avatar.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Avatar: %v", err)
	}
}

func TestShadcn_Breadcrumb(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Breadcrumb = require("shadcn.breadcrumb")
		assert(Breadcrumb ~= nil)
		assert(Breadcrumb.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Breadcrumb: %v", err)
	}
}

func TestShadcn_Kbd(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Kbd = require("shadcn.kbd")
		assert(Kbd ~= nil)
		assert(Kbd.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Kbd: %v", err)
	}
}

func TestShadcn_Input(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Input = require("shadcn.input")
		assert(Input ~= nil)
		assert(Input.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Input: %v", err)
	}
}

func TestShadcn_Switch(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Switch = require("shadcn.switch")
		assert(Switch ~= nil)
		assert(Switch.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Switch: %v", err)
	}
}

func TestShadcn_Progress(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Progress = require("shadcn.progress")
		assert(Progress ~= nil)
		assert(Progress.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
}

func TestShadcn_Accordion(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Accordion = require("shadcn.accordion")
		assert(Accordion ~= nil)
		assert(Accordion.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Accordion: %v", err)
	}
}

func TestShadcn_Tabs(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Tabs = require("shadcn.tabs")
		assert(Tabs ~= nil)
		assert(Tabs.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Tabs: %v", err)
	}
}

func TestShadcn_Table(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Table = require("shadcn.table")
		assert(Table ~= nil)
		assert(Table.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Table: %v", err)
	}
}

func TestShadcn_Pagination(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Pagination = require("shadcn.pagination")
		assert(Pagination ~= nil)
		assert(Pagination.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Pagination: %v", err)
	}
}

func TestShadcn_Toggle(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Toggle = require("shadcn.toggle")
		assert(Toggle ~= nil)
		assert(Toggle.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Toggle: %v", err)
	}
}

func TestShadcn_ToggleGroup(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local ToggleGroup = require("shadcn.toggle_group")
		assert(ToggleGroup ~= nil)
		assert(ToggleGroup.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("ToggleGroup: %v", err)
	}
}

func TestShadcn_InitModule(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local shadcn = require("shadcn")
		assert(shadcn ~= nil, "shadcn module should not be nil")
		assert(shadcn.Button ~= nil, "shadcn.Button should exist")
		assert(shadcn.Badge ~= nil, "shadcn.Badge should exist")
		assert(shadcn.Card ~= nil, "shadcn.Card should exist")
		assert(shadcn.Alert ~= nil, "shadcn.Alert should exist")
		assert(shadcn.Label ~= nil, "shadcn.Label should exist")
		assert(shadcn.Separator ~= nil, "shadcn.Separator should exist")
		assert(shadcn.Skeleton ~= nil, "shadcn.Skeleton should exist")
		assert(shadcn.Spinner ~= nil, "shadcn.Spinner should exist")
		assert(shadcn.Avatar ~= nil, "shadcn.Avatar should exist")
		assert(shadcn.Breadcrumb ~= nil, "shadcn.Breadcrumb should exist")
		assert(shadcn.Kbd ~= nil, "shadcn.Kbd should exist")
		assert(shadcn.Input ~= nil, "shadcn.Input should exist")
		assert(shadcn.Switch ~= nil, "shadcn.Switch should exist")
		assert(shadcn.Progress ~= nil, "shadcn.Progress should exist")
		assert(shadcn.Accordion ~= nil, "shadcn.Accordion should exist")
		assert(shadcn.Tabs ~= nil, "shadcn.Tabs should exist")
		assert(shadcn.Table ~= nil, "shadcn.Table should exist")
		assert(shadcn.Pagination ~= nil, "shadcn.Pagination should exist")
		assert(shadcn.Toggle ~= nil, "shadcn.Toggle should exist")
		assert(shadcn.ToggleGroup ~= nil, "shadcn.ToggleGroup should exist")
	`)
	if err != nil {
		t.Fatalf("Init module: %v", err)
	}
}

// TestShadcn_ButtonRender tests that a Button component can be rendered.
func TestShadcn_ButtonRender(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Button = require("shadcn.button")
		local tree = lumina.render(Button, { label = "Click me", variant = "default" })
		assert(tree ~= nil, "render should return a tree")
	`)
	if err != nil {
		t.Fatalf("ButtonRender: %v", err)
	}
}

// TestShadcn_BadgeRender tests Badge rendering with variant.
func TestShadcn_BadgeRender(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Badge = require("shadcn.badge")
		local tree = lumina.render(Badge, { label = "New", variant = "destructive" })
		assert(tree ~= nil, "render should return a tree")
	`)
	if err != nil {
		t.Fatalf("BadgeRender: %v", err)
	}
}

// TestShadcn_AlertRender tests Alert rendering.
func TestShadcn_AlertRender(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Alert = require("shadcn.alert")
		local tree = lumina.render(Alert, { title = "Warning", description = "Something happened", variant = "destructive" })
		assert(tree ~= nil)
	`)
	if err != nil {
		t.Fatalf("AlertRender: %v", err)
	}
}

// TestShadcn_ProgressRender tests Progress rendering.
func TestShadcn_ProgressRender(t *testing.T) {
	L := shadcnState(t)
	defer L.Close()
	err := L.DoString(`
		local Progress = require("shadcn.progress")
		local tree = lumina.render(Progress, { value = 75, max = 100, showLabel = true })
		assert(tree ~= nil)
	`)
	if err != nil {
		t.Fatalf("ProgressRender: %v", err)
	}
}

// Phase 38: Additional shadcn component tests

func TestShadcnAspectRatio(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`local c = require("shadcn.aspect_ratio"); assert(c ~= nil, "aspect_ratio nil")`)
	if err != nil {
		t.Fatalf("AspectRatio: %v", err)
	}
}

func TestShadcnButtonGroup(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`local c = require("shadcn.button_group"); assert(c ~= nil, "button_group nil")`)
	if err != nil {
		t.Fatalf("ButtonGroup: %v", err)
	}
}

func TestShadcnCalendar(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`local c = require("shadcn.calendar"); assert(c ~= nil, "calendar nil")`)
	if err != nil {
		t.Fatalf("Calendar: %v", err)
	}
}

func TestShadcnDatePicker(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`local c = require("shadcn.date_picker"); assert(c ~= nil, "date_picker nil")`)
	if err != nil {
		t.Fatalf("DatePicker: %v", err)
	}
}

func TestShadcnNavigationMenu(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`local c = require("shadcn.navigation_menu"); assert(c ~= nil, "navigation_menu nil")`)
	if err != nil {
		t.Fatalf("NavigationMenu: %v", err)
	}
}

func TestShadcnResizable(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`
		local r = require("shadcn.resizable")
		assert(r ~= nil, "resizable nil")
		assert(r.Panel ~= nil, "Panel nil")
		assert(r.Handle ~= nil, "Handle nil")
		assert(r.PanelGroup ~= nil, "PanelGroup nil")
	`)
	if err != nil {
		t.Fatalf("Resizable: %v", err)
	}
}

func TestShadcnSidebar(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`
		local s = require("shadcn.sidebar")
		assert(s ~= nil, "sidebar nil")
		assert(s.Sidebar ~= nil, "Sidebar nil")
		assert(s.Layout ~= nil, "Layout nil")
	`)
	if err != nil {
		t.Fatalf("Sidebar: %v", err)
	}
}

func TestShadcnChart(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`
		local c = require("shadcn.chart")
		assert(c ~= nil, "chart nil")
		assert(c.Bar ~= nil, "Bar nil")
		assert(c.Line ~= nil, "Line nil")
		assert(c.Pie ~= nil, "Pie nil")
	`)
	if err != nil {
		t.Fatalf("Chart: %v", err)
	}
}

func TestShadcnDataTable(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`local c = require("shadcn.data_table"); assert(c ~= nil, "data_table nil")`)
	if err != nil {
		t.Fatalf("DataTable: %v", err)
	}
}

func TestShadcnColorPicker(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`local c = require("shadcn.color_picker"); assert(c ~= nil, "color_picker nil")`)
	if err != nil {
		t.Fatalf("ColorPicker: %v", err)
	}
}

func TestShadcnInitAll57(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()
	err := L.DoString(`
		local shadcn = require("shadcn")
		assert(shadcn ~= nil, "shadcn nil")
		-- Verify new components accessible via init
		assert(shadcn.AspectRatio ~= nil, "AspectRatio nil in init")
		assert(shadcn.Calendar ~= nil, "Calendar nil in init")
		assert(shadcn.Chart ~= nil, "Chart nil in init")
		assert(shadcn.DataTable ~= nil, "DataTable nil in init")
		assert(shadcn.ColorPicker ~= nil, "ColorPicker nil in init")
		assert(shadcn.NavigationMenu ~= nil, "NavigationMenu nil in init")
		assert(shadcn.Resizable ~= nil, "Resizable nil in init")
		assert(shadcn.Sidebar ~= nil, "Sidebar nil in init")
	`)
	if err != nil {
		t.Fatalf("shadcn init all 57: %v", err)
	}
}
