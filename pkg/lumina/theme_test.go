package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestSetThemeByName(t *testing.T) {
	tm := GetThemeManager()
	tm.ResetTheme()

	if !tm.SetThemeByName("tokyo-night") {
		t.Fatal("expected tokyo-night to be found")
	}
	theme := tm.GetTheme()
	if theme.Name != "tokyo-night" {
		t.Fatalf("expected tokyo-night, got %s", theme.Name)
	}
	if theme.Colors["background"] != "#1A1B26" {
		t.Fatalf("expected #1A1B26, got %s", theme.Colors["background"])
	}
	tm.ResetTheme()
}

func TestSetThemeCustom(t *testing.T) {
	tm := GetThemeManager()
	tm.ResetTheme()

	custom := Theme{
		Name:    "my-theme",
		Colors:  map[string]string{"primary": "#FF0000", "background": "#000000"},
		Space:   map[string]int{"xs": 0, "sm": 1},
		Spacing: map[string]int{"xs": 0, "sm": 1},
		Borders: map[string]string{"none": ""},
	}
	tm.SetTheme(custom)
	got := tm.GetTheme()
	if got.Name != "my-theme" {
		t.Fatalf("expected my-theme, got %s", got.Name)
	}
	if got.Colors["primary"] != "#FF0000" {
		t.Fatalf("expected #FF0000, got %s", got.Colors["primary"])
	}
	tm.ResetTheme()
}

func TestGetColor(t *testing.T) {
	tm := GetThemeManager()
	tm.ResetTheme()

	// Known token
	c := tm.GetColor("primary")
	if c != "#89B4FA" {
		t.Fatalf("expected #89B4FA, got %s", c)
	}

	// Unknown token returns as-is
	c = tm.GetColor("#AABBCC")
	if c != "#AABBCC" {
		t.Fatalf("expected #AABBCC, got %s", c)
	}
}

func TestThemeTokenResolution(t *testing.T) {
	tm := GetThemeManager()
	tm.ResetTheme()

	style := map[string]string{
		"background": "primary",
		"foreground": "foreground",
		"other":      "some-value",
	}
	resolved := tm.ResolveStyle(style)
	if resolved["background"] != "#89B4FA" {
		t.Fatalf("expected #89B4FA, got %s", resolved["background"])
	}
	if resolved["foreground"] != "#CDD6F4" {
		t.Fatalf("expected #CDD6F4, got %s", resolved["foreground"])
	}
	if resolved["other"] != "some-value" {
		t.Fatalf("expected some-value, got %s", resolved["other"])
	}
}

func TestThemeVersionIncrement(t *testing.T) {
	tm := GetThemeManager()
	tm.ResetTheme()

	v0 := tm.Version()
	tm.SetThemeByName("nord")
	v1 := tm.Version()
	if v1 <= v0 {
		t.Fatalf("expected version to increment: %d -> %d", v0, v1)
	}
	tm.ResetTheme()
}

func TestMultipleBuiltinThemes(t *testing.T) {
	names := ListThemes()
	if len(names) < 4 {
		t.Fatalf("expected at least 4 themes, got %d", len(names))
	}
	// Verify each can be set
	tm := GetThemeManager()
	for _, name := range names {
		if !tm.SetThemeByName(name) {
			t.Fatalf("failed to set theme: %s", name)
		}
		got := tm.GetTheme()
		if got.Name != name {
			t.Fatalf("expected %s, got %s", name, got.Name)
		}
	}
	tm.ResetTheme()
}

func TestDefineAndGetStyles(t *testing.T) {
	tm := GetThemeManager()
	tm.ResetTheme()

	tm.DefineStyles(map[string]map[string]string{
		"button": {
			"background": "primary",
			"foreground": "foreground",
			"padding":    "sm",
		},
		"button:focused": {
			"background": "accent",
			"border":     "double",
		},
	})

	s := tm.GetStyle("button")
	if s == nil {
		t.Fatal("expected button style")
	}
	if s["background"] != "primary" {
		t.Fatalf("expected primary, got %s", s["background"])
	}

	sf := tm.GetStyle("button:focused")
	if sf == nil {
		t.Fatal("expected button:focused style")
	}
	if sf["background"] != "accent" {
		t.Fatalf("expected accent, got %s", sf["background"])
	}
	tm.ResetTheme()
}

func TestLuaSetThemeByName(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalThemeManager.ResetTheme()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.setTheme("tokyo-night")
	`)
	if err != nil {
		t.Fatalf("setTheme: %v", err)
	}
	theme := globalThemeManager.GetTheme()
	if theme.Name != "tokyo-night" {
		t.Fatalf("expected tokyo-night, got %s", theme.Name)
	}
	globalThemeManager.ResetTheme()
}

func TestLuaSetThemeCustom(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalThemeManager.ResetTheme()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.setTheme({
			name = "my-custom",
			colors = { primary = "#FF0000", background = "#111111" },
		})
	`)
	if err != nil {
		t.Fatalf("setTheme custom: %v", err)
	}
	theme := globalThemeManager.GetTheme()
	if theme.Name != "my-custom" {
		t.Fatalf("expected my-custom, got %s", theme.Name)
	}
	if theme.Colors["primary"] != "#FF0000" {
		t.Fatalf("expected #FF0000, got %s", theme.Colors["primary"])
	}
	globalThemeManager.ResetTheme()
}

func TestLuaUseTheme(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalThemeManager.ResetTheme()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local theme = lumina.useTheme()
		assert(theme.name == "catppuccin-mocha", "expected catppuccin-mocha, got " .. tostring(theme.name))
		assert(theme.colors.primary == "#89B4FA", "expected #89B4FA, got " .. tostring(theme.colors.primary))
		assert(theme.space.md == 2, "expected 2, got " .. tostring(theme.space.md))
	`)
	if err != nil {
		t.Fatalf("useTheme: %v", err)
	}
}

func TestLuaDefineStyles(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalThemeManager.ResetTheme()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.defineStyles({
			myButton = {
				background = "primary",
				foreground = "foreground",
				padding = "sm",
			},
		})
	`)
	if err != nil {
		t.Fatalf("defineStyles: %v", err)
	}

	s := globalThemeManager.GetStyle("myButton")
	if s == nil {
		t.Fatal("expected myButton style")
	}
	if s["background"] != "primary" {
		t.Fatalf("expected primary, got %s", s["background"])
	}
	globalThemeManager.ResetTheme()
}

func TestLuaGetThemeColor(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalThemeManager.ResetTheme()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local c = lumina.getThemeColor("primary")
		assert(c == "#89B4FA", "expected #89B4FA, got " .. tostring(c))
		local raw = lumina.getThemeColor("#AABBCC")
		assert(raw == "#AABBCC", "expected #AABBCC, got " .. tostring(raw))
	`)
	if err != nil {
		t.Fatalf("getThemeColor: %v", err)
	}
}
