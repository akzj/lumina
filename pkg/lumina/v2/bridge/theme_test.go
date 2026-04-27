package bridge

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/paint"
)

// --- ThemeManager unit tests ---

func TestThemeManager_BuiltinThemes(t *testing.T) {
	tm := NewThemeManager()

	// Default active should be "dark".
	if tm.ActiveName() != "dark" {
		t.Errorf("ActiveName = %q, want %q", tm.ActiveName(), "dark")
	}

	dark := tm.Active()
	if dark == nil {
		t.Fatal("Active() returned nil for default dark theme")
	}

	// Spot-check a few Catppuccin Mocha colors.
	checks := map[string]string{
		"bg":     "#1E1E2E",
		"fg":     "#CDD6F4",
		"accent": "#89B4FA",
	}
	for key, want := range checks {
		if got := dark[key]; got != want {
			t.Errorf("dark[%q] = %q, want %q", key, got, want)
		}
	}

	// Light theme should also be pre-registered.
	if !tm.SetActive("light") {
		t.Fatal("SetActive(\"light\") returned false — light theme not registered")
	}
	light := tm.Active()
	if light == nil {
		t.Fatal("Active() returned nil for light theme")
	}
	if light["bg"] != "#EFF1F5" {
		t.Errorf("light[bg] = %q, want %q", light["bg"], "#EFF1F5")
	}
}

func TestThemeManager_Define(t *testing.T) {
	tm := NewThemeManager()

	custom := map[string]string{
		"bg": "#000000", "fg": "#FFFFFF",
	}
	tm.Define("custom", custom)

	if !tm.SetActive("custom") {
		t.Fatal("SetActive(\"custom\") returned false after Define")
	}
	colors := tm.Active()
	if colors["bg"] != "#000000" {
		t.Errorf("custom[bg] = %q, want %q", colors["bg"], "#000000")
	}
	if colors["fg"] != "#FFFFFF" {
		t.Errorf("custom[fg] = %q, want %q", colors["fg"], "#FFFFFF")
	}
}

func TestThemeManager_SetActive_Unknown(t *testing.T) {
	tm := NewThemeManager()

	if tm.SetActive("nonexistent") {
		t.Error("SetActive(\"nonexistent\") should return false")
	}
	// Active should still be "dark".
	if tm.ActiveName() != "dark" {
		t.Errorf("ActiveName = %q after failed SetActive, want %q", tm.ActiveName(), "dark")
	}
}

// --- Lua integration tests ---

func TestBridge_RegisterHooks_ThemeHooks(t *testing.T) {
	b := newTestBridge(t)
	L := b.L
	b.RegisterHooks()

	err := L.DoString(`
		assert(type(lumina.defineTheme) == "function", "defineTheme should be a function")
		assert(type(lumina.setTheme) == "function", "setTheme should be a function")
		assert(type(lumina.useTheme) == "function", "useTheme should be a function")
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBridge_UseTheme_DefaultDark(t *testing.T) {
	b := newTestBridge(t)
	L := b.L
	b.RegisterHooks()

	err := L.DoString(`
		local theme = lumina.useTheme()
		theme_bg = theme.bg
		theme_fg = theme.fg
		theme_accent = theme.accent
	`)
	if err != nil {
		t.Fatal(err)
	}

	checks := map[string]string{
		"theme_bg":     "#1E1E2E",
		"theme_fg":     "#CDD6F4",
		"theme_accent": "#89B4FA",
	}
	for global, want := range checks {
		L.GetGlobal(global)
		got, _ := L.ToString(-1)
		L.Pop(1)
		if got != want {
			t.Errorf("%s = %q, want %q", global, got, want)
		}
	}
}

func TestBridge_SetTheme_SwitchToLight(t *testing.T) {
	b := newTestBridge(t)
	L := b.L
	b.RegisterHooks()

	err := L.DoString(`
		lumina.setTheme("light")
		local theme = lumina.useTheme()
		theme_bg = theme.bg
		theme_fg = theme.fg
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("theme_bg")
	bg, _ := L.ToString(-1)
	L.Pop(1)
	if bg != "#EFF1F5" {
		t.Errorf("theme_bg = %q, want %q", bg, "#EFF1F5")
	}

	L.GetGlobal("theme_fg")
	fg, _ := L.ToString(-1)
	L.Pop(1)
	if fg != "#4C4F69" {
		t.Errorf("theme_fg = %q, want %q", fg, "#4C4F69")
	}
}

func TestBridge_SetTheme_Unknown(t *testing.T) {
	b := newTestBridge(t)
	L := b.L
	b.RegisterHooks()

	err := L.DoString(`
		local ok, err = pcall(function()
			lumina.setTheme("nonexistent")
		end)
		set_theme_error = not ok
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("set_theme_error")
	hasError := L.ToBoolean(-1)
	L.Pop(1)
	if !hasError {
		t.Error("setTheme with unknown theme should error")
	}
}

func TestBridge_DefineTheme_Custom(t *testing.T) {
	b := newTestBridge(t)
	L := b.L
	b.RegisterHooks()

	err := L.DoString(`
		lumina.defineTheme("ocean", {
			bg = "#0D1117",
			fg = "#C9D1D9",
			accent = "#58A6FF",
		})
		lumina.setTheme("ocean")
		local theme = lumina.useTheme()
		theme_bg = theme.bg
		theme_fg = theme.fg
		theme_accent = theme.accent
	`)
	if err != nil {
		t.Fatal(err)
	}

	checks := map[string]string{
		"theme_bg":     "#0D1117",
		"theme_fg":     "#C9D1D9",
		"theme_accent": "#58A6FF",
	}
	for global, want := range checks {
		L.GetGlobal(global)
		got, _ := L.ToString(-1)
		L.Pop(1)
		if got != want {
			t.Errorf("%s = %q, want %q", global, got, want)
		}
	}
}

func TestBridge_SetTheme_MarksDirty(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Create a manager with a component.
	p := paint.NewPainter()
	mgr := component.NewManager(p)
	b.SetManager(mgr)
	b.RegisterHooks()

	nopRender := func(state, props map[string]any) *layout.VNode { return nil }
	comp := component.NewComponent("c1", "c1", buffer.Rect{W: 10, H: 5}, 0, nopRender)
	mgr.Register(comp)

	// Clear dirty flag.
	mgr.ClearDirty()
	if comp.IsDirtyPaint() {
		t.Fatal("component should not be dirty after ClearDirty")
	}

	// setTheme should mark all components dirty.
	err := L.DoString(`lumina.setTheme("light")`)
	if err != nil {
		t.Fatal(err)
	}

	if !comp.IsDirtyPaint() {
		t.Error("component should be dirty after setTheme")
	}
}

func TestBridge_DefineTheme_OverrideBuiltin(t *testing.T) {
	b := newTestBridge(t)
	L := b.L
	b.RegisterHooks()

	// Override the built-in "dark" theme.
	err := L.DoString(`
		lumina.defineTheme("dark", {
			bg = "#000000",
			fg = "#FFFFFF",
		})
		local theme = lumina.useTheme()
		theme_bg = theme.bg
		theme_fg = theme.fg
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("theme_bg")
	bg, _ := L.ToString(-1)
	L.Pop(1)
	if bg != "#000000" {
		t.Errorf("overridden dark bg = %q, want %q", bg, "#000000")
	}

	L.GetGlobal("theme_fg")
	fg, _ := L.ToString(-1)
	L.Pop(1)
	if fg != "#FFFFFF" {
		t.Errorf("overridden dark fg = %q, want %q", fg, "#FFFFFF")
	}
}
