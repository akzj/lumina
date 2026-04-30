package v2

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestV2E2E_RequireLumina(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 5)
	err := app.RunString(`
		local l = require("lumina")
		assert(l ~= nil, "require('lumina') returned nil")
		assert(l.createElement ~= nil, "lumina.createElement missing")
		assert(l.defineComponent ~= nil, "lumina.defineComponent missing")

		l.createComponent({
			id = "req-lumina-test",
			render = function()
				return l.createElement("text", {}, "lumina-ok")
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()
	if !screenHasString(ta, "lumina-ok") {
		t.Error("expected 'lumina-ok' on screen")
	}
}

func TestV2E2E_RequireLuxCard(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)
	err := app.RunString(`
		local Card = require("lux.card")
		assert(Card ~= nil, "require('lux.card') returned nil")

		lumina.createComponent({
			id = "card-test",
			render = function()
				return lumina.createElement("vbox", {},
					lumina.createElement(Card, {title = "Hello", id = "card1"})
				)
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()
	if !screenHasString(ta, "Hello") {
		t.Error("expected 'Hello' from Card component")
	}
}

func TestV2E2E_RequireLux(t *testing.T) {
	app, _, _ := newV2App(t, 40, 10)
	err := app.RunString(`
		local lux = require("lux")
		assert(lux ~= nil, "require('lux') returned nil")
		assert(lux.Card ~= nil, "lux.Card missing")
		assert(lux.Badge ~= nil, "lux.Badge missing")
		assert(lux.Divider ~= nil, "lux.Divider missing")
		assert(lux.Progress ~= nil, "lux.Progress missing")
		assert(lux.Spinner ~= nil, "lux.Spinner missing")
		assert(lux.ListView ~= nil, "lux.ListView missing")
		assert(lux.Atlantis ~= nil, "lux.Atlantis missing")
		assert(lux.Atlantis.applyTheme ~= nil, "lux.Atlantis.applyTheme missing")
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
}

func TestV2E2E_RequireLuxBadge(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)
	err := app.RunString(`
		local Badge = require("lux.badge")
		lumina.createComponent({
			id = "badge-test",
			render = function()
				return lumina.createElement("vbox", {},
					lumina.createElement(Badge, {label = "NEW", variant = "success", id = "badge1"})
				)
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()
	if !screenHasString(ta, "NEW") {
		t.Error("expected 'NEW' from Badge component")
	}
}

func TestV2E2E_LuminaGetTheme(t *testing.T) {
	app, _, _ := newV2App(t, 40, 10)
	err := app.RunString(`
		local t = lumina.getTheme()
		assert(t ~= nil, "getTheme returned nil")
		assert(t.primary ~= nil, "theme.primary missing")
		assert(t.base ~= nil, "theme.base missing")
		assert(t.text ~= nil, "theme.text missing")
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
}

func TestV2E2E_RequireTheme(t *testing.T) {
	app, _, _ := newV2App(t, 40, 10)
	err := app.RunString(`
		local theme = require("theme")
		assert(theme ~= nil, "require('theme') returned nil")
		assert(theme.current ~= nil, "theme.current missing")
		local t = theme.current()
		assert(t ~= nil, "theme.current() returned nil")
		assert(t.primary ~= nil, "theme.primary missing")
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
}

func TestV2E2E_RequireLuxProgress(t *testing.T) {
	app, ta, _ := newV2App(t, 60, 10)
	err := app.RunString(`
		local Progress = require("lux.progress")
		lumina.createComponent({
			id = "progress-test",
			render = function()
				return lumina.createElement("vbox", {},
					lumina.createElement(Progress, {value = 50, width = 10, id = "prog1"})
				)
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()
	if !screenHasString(ta, "50%") {
		t.Error("expected '50%' from Progress component")
	}
}

func TestV2E2E_RequireLuxDivider(t *testing.T) {
	app, ta, _ := newV2App(t, 60, 10)
	err := app.RunString(`
		local Divider = require("lux.divider")
		lumina.createComponent({
			id = "divider-test",
			render = function()
				return lumina.createElement("vbox", {},
					lumina.createElement(Divider, {width = 10, id = "div1"})
				)
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()
	// Divider renders repeated "─" chars
	if !screenHasChar(ta, '─') {
		t.Error("expected '─' from Divider component")
	}
}

func TestV2E2E_RequireLuxListView(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 12)
	err := app.RunString(`
		local ListView = require("lux.list")
		lumina.createComponent({
			id = "listview-test",
			render = function()
				return lumina.createElement("vbox", { id = "root" },
					lumina.createElement(ListView, {
						id = "lv",
						rows = { { t = "RowA" }, { t = "RowB" } },
						rowHeight = 1,
						height = 6,
						selectedIndex = 1,
						renderRow = function(row, i, ctx)
							return lumina.createElement("text", {
								style = { height = 1 },
								bold = ctx.selected,
							}, row.t)
						end,
					})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()
	if !screenHasString(ta, "RowA") || !screenHasString(ta, "RowB") {
		t.Errorf("expected RowA and RowB on screen, got: %q", readScreenLine(ta, 0, 40))
	}
}

func TestV2E2E_AtlantisComponentsMainLua(t *testing.T) {
	app, ta, _ := newV2App(t, 100, 40)
	path := filepath.Join("..", "examples", "components", "main.lua")
	if err := app.RunScript(path); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.RenderAll()
	if !screenHasString(ta, "ATLANTIS") {
		t.Error("expected ATLANTIS brand on screen")
	}
	var flat strings.Builder
	for y := 0; y < ta.LastScreen.Height(); y++ {
		flat.WriteString(readScreenLine(ta, y, 200))
		flat.WriteByte('\n')
	}
	screen := flat.String()
	if !strings.Contains(screen, "Vertical") {
		t.Logf("screen dump:\n%s", screen)
		t.Error("expected Vertical form card title somewhere on screen")
	}
	if !strings.Contains(screen, "Form Layout") {
		t.Error("expected breadcrumb Form Layout segment")
	}
}

func TestV2E2E_AtlantisShellMinimal(t *testing.T) {
	app, ta, _ := newV2App(t, 100, 30)
	err := app.RunString(`
		local lux = require("lux")
		local A = lux.Atlantis
		lumina.app({
			id = "atlantis-min",
			render = function()
				return A.Shell({
					sidebar = A.SideNav({
						brand = "ATLANTIS",
						items = { { id = "a", label = "Nav" } },
						activeId = "a",
					}),
					mainChildren = {
						lumina.createElement("text", { id = "mark" }, "Vertical"),
					},
				})
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString: %v", err)
	}
	app.RenderAll()
	if !screenHasString(ta, "Vertical") {
		t.Error("expected Vertical marker from minimal Shell")
	}
	if !screenHasString(ta, "ATLANTIS") {
		t.Error("expected ATLANTIS in minimal Shell")
	}
}
