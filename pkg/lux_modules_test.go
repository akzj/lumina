package v2

import (
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
