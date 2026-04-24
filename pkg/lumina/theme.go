package lumina

import (
	"sync"
)

// Theme represents a complete visual theme with semantic color tokens.
type Theme struct {
	Name    string            `json:"name"`
	Colors  map[string]string `json:"colors"`  // semantic color tokens
	Space   map[string]int    `json:"space"`   // spacing scale
	Spacing map[string]int    `json:"spacing"` // alias for Space (compat)
	Borders map[string]string `json:"borders"` // border style tokens
}

// Built-in themes

// CatppuccinMocha is the default dark theme (Catppuccin Mocha palette).
var CatppuccinMocha = Theme{
	Name: "catppuccin-mocha",
	Colors: map[string]string{
		"background":  "#1E1E2E",
		"foreground":  "#CDD6F4",
		"primary":     "#89B4FA",
		"secondary":   "#A6ADC8",
		"accent":      "#F5C2E7",
		"destructive": "#F38BA8",
		"muted":       "#313244",
		"border":      "#45475A",
		"ring":        "#89B4FA",
		"card":        "#1E1E2E",
		"popover":     "#1E1E2E",
		"success":     "#A6E3A1",
		"warning":     "#F9E2AF",
		"info":        "#89DCEB",
	},
	Space:   map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
	Spacing: map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
	Borders: map[string]string{"none": "", "rounded": "rounded", "double": "double", "single": "single"},
}

// CatppuccinLatte is a light theme (Catppuccin Latte palette).
var CatppuccinLatte = Theme{
	Name: "catppuccin-latte",
	Colors: map[string]string{
		"background":  "#EFF1F5",
		"foreground":  "#4C4F69",
		"primary":     "#1E66F5",
		"secondary":   "#6C6F85",
		"accent":      "#EA76CB",
		"destructive": "#D20F39",
		"muted":       "#E6E9EF",
		"border":      "#BCC0CC",
		"ring":        "#1E66F5",
		"card":        "#EFF1F5",
		"popover":     "#EFF1F5",
		"success":     "#40A02B",
		"warning":     "#DF8E1D",
		"info":        "#04A5E5",
	},
	Space:   map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
	Spacing: map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
	Borders: map[string]string{"none": "", "rounded": "rounded", "double": "double", "single": "single"},
}

// TokyoNight is a dark blue theme.
var TokyoNight = Theme{
	Name: "tokyo-night",
	Colors: map[string]string{
		"background":  "#1A1B26",
		"foreground":  "#C0CAF5",
		"primary":     "#7AA2F7",
		"secondary":   "#9AA5CE",
		"accent":      "#BB9AF7",
		"destructive": "#F7768E",
		"muted":       "#292E42",
		"border":      "#3B4261",
		"ring":        "#7AA2F7",
		"card":        "#1A1B26",
		"popover":     "#1A1B26",
		"success":     "#9ECE6A",
		"warning":     "#E0AF68",
		"info":        "#7DCFFF",
	},
	Space:   map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
	Spacing: map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
	Borders: map[string]string{"none": "", "rounded": "rounded", "double": "double", "single": "single"},
}

// Nord is a cool arctic theme.
var Nord = Theme{
	Name: "nord",
	Colors: map[string]string{
		"background":  "#2E3440",
		"foreground":  "#ECEFF4",
		"primary":     "#88C0D0",
		"secondary":   "#D8DEE9",
		"accent":      "#B48EAD",
		"destructive": "#BF616A",
		"muted":       "#3B4252",
		"border":      "#4C566A",
		"ring":        "#88C0D0",
		"card":        "#2E3440",
		"popover":     "#2E3440",
		"success":     "#A3BE8C",
		"warning":     "#EBCB8B",
		"info":        "#81A1C1",
	},
	Space:   map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
	Spacing: map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
	Borders: map[string]string{"none": "", "rounded": "rounded", "double": "double", "single": "single"},
}

// builtinThemes maps theme names to theme objects.
var builtinThemes = map[string]Theme{
	"catppuccin-mocha": CatppuccinMocha,
	"catppuccin-latte": CatppuccinLatte,
	"tokyo-night":      TokyoNight,
	"nord":             Nord,
}

// ThemeManager manages the current theme and style definitions.
type ThemeManager struct {
	mu      sync.RWMutex
	current Theme
	styles  map[string]map[string]string // className -> property -> value
	version int                          // incremented on theme change
}

var globalThemeManager = &ThemeManager{
	current: CatppuccinMocha,
	styles:  make(map[string]map[string]string),
}

// GetThemeManager returns the global theme manager.
func GetThemeManager() *ThemeManager {
	return globalThemeManager
}

// SetTheme sets the current theme by name or custom theme.
func (tm *ThemeManager) SetTheme(theme Theme) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.current = theme
	tm.version++
}

// SetThemeByName sets the current theme by built-in name.
func (tm *ThemeManager) SetThemeByName(name string) bool {
	if t, ok := builtinThemes[name]; ok {
		tm.SetTheme(t)
		return true
	}
	return false
}

// GetTheme returns the current theme.
func (tm *ThemeManager) GetTheme() Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.current
}

// GetColor resolves a color token from the current theme.
func (tm *ThemeManager) GetColor(token string) string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if c, ok := tm.current.Colors[token]; ok {
		return c
	}
	return token // return as-is if not a token (e.g. raw hex)
}

// GetSpace resolves a spacing token from the current theme.
func (tm *ThemeManager) GetSpace(token string) int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if s, ok := tm.current.Space[token]; ok {
		return s
	}
	return 0
}

// Version returns the current theme version (for change detection).
func (tm *ThemeManager) Version() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.version
}

// DefineStyles registers CSS-like style classes.
func (tm *ThemeManager) DefineStyles(styles map[string]map[string]string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for name, props := range styles {
		tm.styles[name] = props
	}
}

// GetStyle returns the style properties for a className.
func (tm *ThemeManager) GetStyle(className string) map[string]string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.styles[className]
}

// ResolveStyle resolves theme tokens in a style map.
// E.g., "primary" -> "#89B4FA" (from current theme colors).
func (tm *ThemeManager) ResolveStyle(style map[string]string) map[string]string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	resolved := make(map[string]string, len(style))
	for k, v := range style {
		if k == "background" || k == "foreground" || k == "border_color" {
			if c, ok := tm.current.Colors[v]; ok {
				resolved[k] = c
				continue
			}
		}
		resolved[k] = v
	}
	return resolved
}

// ResetTheme resets to default theme (for testing).
func (tm *ThemeManager) ResetTheme() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.current = CatppuccinMocha
	tm.styles = make(map[string]map[string]string)
	tm.version = 0
}

// ListThemes returns available built-in theme names.
func ListThemes() []string {
	names := make([]string, 0, len(builtinThemes))
	for name := range builtinThemes {
		names = append(names, name)
	}
	return names
}

// --- Bridge functions for style_api.go and hooks.go compatibility ---

// RegisterTheme registers a custom theme by name.
func RegisterTheme(name string, theme *Theme) {
	builtinThemes[name] = *theme
}

// GetThemeByName returns a theme by name, or nil if not found.
func GetThemeByName(name string) *Theme {
	if t, ok := builtinThemes[name]; ok {
		return &t
	}
	return nil
}

// SetTheme sets the current theme (pointer version for style_api compat).
func SetTheme(theme *Theme) {
	globalThemeManager.SetTheme(*theme)
}

// DefaultTheme returns the default theme (Catppuccin Mocha).
func DefaultTheme() *Theme {
	t := CatppuccinMocha
	return &t
}

// GetCurrentTheme returns the current theme as a pointer.
func GetCurrentTheme() *Theme {
	t := globalThemeManager.GetTheme()
	return &t
}
