package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/stretchr/testify/assert"
)

func TestThemeDefault(t *testing.T) {
	theme := DefaultTheme()
	assert.Equal(t, "dark", theme.Name)
	assert.Equal(t, "cyan", theme.Colors["primary"])
	assert.Equal(t, "white", theme.Colors["text"])
	assert.Equal(t, "black", theme.Colors["background"])
	assert.Equal(t, 4, theme.Spacing["sm"])
	assert.Equal(t, 8, theme.Spacing["md"])
}

func TestThemeLight(t *testing.T) {
	theme := LightTheme()
	assert.Equal(t, "light", theme.Name)
	assert.Equal(t, "blue", theme.Colors["primary"])
	assert.Equal(t, "black", theme.Colors["text"])
	assert.Equal(t, "white", theme.Colors["background"])
}

func TestThemeRegistry(t *testing.T) {
	RegisterTheme("test", DefaultTheme())
	theme := GetThemeByName("test")
	assert.NotNil(t, theme)
	assert.Equal(t, "dark", theme.Name)

	// Non-existent theme
	theme = GetThemeByName("nonexistent")
	assert.Nil(t, theme)
}

func TestUseThemeHook(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local theme = lumina.hooks.useTheme()
		assert(theme.name == "dark", "default theme should be dark, got: " .. tostring(theme.name))
		assert(theme.colors ~= nil, "colors should exist")
		assert(theme.colors.primary == "cyan", "primary should be cyan")
		assert(theme.spacing ~= nil, "spacing should exist")
		assert(theme.spacing.md == 8, "md spacing should be 8")
	`)
	assert.NoError(t, err)
}

func TestDefineStyle(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		
		-- Define a style
		lumina.defineStyle("button", {
			padding = 8,
			color = "primary"
		})
		
		-- Get the style back
		local style = lumina.getStyle("button")
		assert(style ~= nil, "style should exist")
		assert(style.padding == 8, "padding should be 8")
		assert(style.color == "primary", "color should be primary")
	`)
	assert.NoError(t, err)
}

func TestDefineGlobalStyles(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		
		-- Define multiple styles at once
		lumina.defineGlobalStyles({
			primary = { color = "blue" },
			secondary = { color = "gray" }
		})
		
		-- Get styles back
		local p = lumina.getStyle("primary")
		assert(p ~= nil and p.color == "blue")
		
		local s = lumina.getStyle("secondary")
		assert(s ~= nil and s.color == "gray")
	`)
	assert.NoError(t, err)
}

func TestSetTheme(t *testing.T) {
	// Test setting to light theme
	SetTheme(LightTheme())
	theme := GetCurrentTheme()
	assert.Equal(t, "light", theme.Name)

	// Test setting to dark theme
	SetTheme(DefaultTheme())
	theme = GetCurrentTheme()
	assert.Equal(t, "dark", theme.Name)
}

func TestSetThemeLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		
		-- Set to light theme
		lumina.setTheme("light")
		
		-- Verify via useTheme
		local theme = lumina.hooks.useTheme()
		assert(theme.name == "light", "theme should be light")
		
		-- Reset to dark
		lumina.setTheme("dark")
		theme = lumina.hooks.useTheme()
		assert(theme.name == "dark", "theme should be dark")
	`)
	assert.NoError(t, err)
}

func TestDefineTheme(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		
		-- Define a custom theme
		local theme = lumina.defineTheme("custom", {
			colors = {
				primary = "red",
				background = "black"
			},
			spacing = {
				sm = 2,
				md = 4
			}
		})
		
		assert(theme ~= nil, "theme should be returned")
		assert(theme.name == "custom")
		assert(theme.colors.primary == "red")
		
		-- Use the theme
		lumina.setTheme("custom")
		local current = lumina.hooks.useTheme()
		assert(current.name == "custom")
	`)
	assert.NoError(t, err)
}

func TestStyleResolver(t *testing.T) {
	theme := DefaultTheme()
	resolver := NewStyleResolver(theme)

	assert.Equal(t, "cyan", resolver.ResolveColor("primary"))
	assert.Equal(t, "white", resolver.ResolveColor("unknown_color"))

	assert.Equal(t, 4, resolver.ResolveSpacing("sm"))
	assert.Equal(t, 8, resolver.ResolveSpacing("md"))
	assert.Equal(t, 4, resolver.ResolveSpacing("unknown_spacing")) // default
}

func TestStyleResolverNilTheme(t *testing.T) {
	resolver := NewStyleResolver(nil)
	assert.NotNil(t, resolver)
	assert.Equal(t, "cyan", resolver.ResolveColor("primary")) // from default
}
