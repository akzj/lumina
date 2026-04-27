package bridge

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// ThemeManager manages named themes and the active theme.
type ThemeManager struct {
	themes map[string]map[string]string // name -> {key: value}
	active string
}

// NewThemeManager creates a ThemeManager with built-in dark/light themes.
func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		themes: make(map[string]map[string]string),
	}
	// Pre-register Catppuccin Mocha (dark) and Latte (light).
	tm.Define("dark", map[string]string{
		"bg": "#1E1E2E", "surface": "#313244", "fg": "#CDD6F4",
		"accent": "#89B4FA", "success": "#A6E3A1", "warning": "#F9E2AF",
		"error": "#F38BA8", "muted": "#6C7086", "border": "#585B70",
	})
	tm.Define("light", map[string]string{
		"bg": "#EFF1F5", "surface": "#CCD0DA", "fg": "#4C4F69",
		"accent": "#1E66F5", "success": "#40A02B", "warning": "#DF8E1D",
		"error": "#D20F39", "muted": "#9CA0B0", "border": "#BCC0CC",
	})
	tm.active = "dark"
	return tm
}

// Define registers a named theme with the given color map.
func (tm *ThemeManager) Define(name string, colors map[string]string) {
	tm.themes[name] = colors
}

// SetActive sets the active theme by name. Returns false if unknown.
func (tm *ThemeManager) SetActive(name string) bool {
	if _, ok := tm.themes[name]; ok {
		tm.active = name
		return true
	}
	return false
}

// Active returns the active theme's color map, or nil if none.
func (tm *ThemeManager) Active() map[string]string {
	if tm.active == "" {
		return nil
	}
	return tm.themes[tm.active]
}

// ActiveName returns the name of the active theme.
func (tm *ThemeManager) ActiveName() string {
	return tm.active
}

// --- Lua hook implementations ---

// luaDefineTheme implements lumina.defineTheme(name, colorsTable).
func (b *Bridge) luaDefineTheme(L *lua.State) int {
	name := L.CheckString(1)
	L.CheckType(2, lua.TypeTable)

	colors := make(map[string]string)
	L.ForEach(2, func(L *lua.State) bool {
		if L.Type(-2) == lua.TypeString {
			key, _ := L.ToString(-2)
			val, _ := L.ToString(-1)
			if key != "" && val != "" {
				colors[key] = val
			}
		}
		return true
	})

	b.themeMgr.Define(name, colors)
	return 0
}

// luaSetTheme implements lumina.setTheme(name).
// Marks all components dirty to trigger re-render with new theme.
func (b *Bridge) luaSetTheme(L *lua.State) int {
	name := L.CheckString(1)
	if !b.themeMgr.SetActive(name) {
		L.PushString("setTheme: unknown theme '" + name + "'")
		L.Error()
		return 0
	}
	// Mark all components dirty so they re-render with the new theme.
	if b.manager != nil {
		for _, comp := range b.manager.GetAll() {
			comp.MarkDirty()
		}
	}
	return 0
}

// luaUseTheme implements lumina.useTheme() → table of colors.
// Returns the active theme as a Lua table, or an empty table if none.
func (b *Bridge) luaUseTheme(L *lua.State) int {
	colors := b.themeMgr.Active()
	if colors == nil {
		L.NewTable()
		return 1
	}
	L.NewTable()
	idx := L.AbsIndex(-1)
	for k, v := range colors {
		L.PushString(v)
		L.SetField(idx, k)
	}
	return 1
}
