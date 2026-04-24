// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"sync"
	"sync/atomic"
)

// Theme represents a Lumina theme with colors, spacing, borders.
type Theme struct {
	Name    string
	Colors  map[string]string // "primary" → "cyan"
	Spacing map[string]int    // "sm" → 4, "md" → 8
	Borders map[string]string // "rounded" → "round"
}

// DefaultTheme returns the built-in dark theme.
func DefaultTheme() *Theme {
	return &Theme{
		Name: "dark",
		Colors: map[string]string{
			"primary":    "cyan",
			"secondary":  "white",
			"background": "black",
			"text":       "white",
			"surface":    "gray",
			"error":      "red",
			"success":    "green",
			"warning":    "yellow",
		},
		Spacing: map[string]int{
			"xs": 2, "sm": 4, "md": 8, "lg": 16, "xl": 32,
		},
		Borders: map[string]string{
			"rounded": "round",
			"single":  "single",
			"double":  "double",
			"none":    "none",
		},
	}
}

// LightTheme returns the built-in light theme.
func LightTheme() *Theme {
	return &Theme{
		Name: "light",
		Colors: map[string]string{
			"primary":    "blue",
			"secondary":  "black",
			"background": "white",
			"text":       "black",
			"surface":    "lightgray",
			"error":      "red",
			"success":    "green",
			"warning":    "yellow",
		},
		Spacing: map[string]int{
			"xs": 2, "sm": 4, "md": 8, "lg": 16, "xl": 32,
		},
		Borders: map[string]string{
			"rounded": "round",
			"single":  "single",
			"double":  "double",
			"none":    "none",
		},
	}
}

// Theme registry
var (
	themeRegistry   = make(map[string]*Theme)
	themeRegistryMu sync.RWMutex
)

// RegisterTheme registers a named theme.
func RegisterTheme(name string, t *Theme) {
	themeRegistryMu.Lock()
	defer themeRegistryMu.Unlock()
	themeRegistry[name] = t
}

// GetThemeByName returns a registered theme by name.
func GetThemeByName(name string) *Theme {
	themeRegistryMu.RLock()
	defer themeRegistryMu.RUnlock()
	return themeRegistry[name]
}

func init() {
	// Register built-in themes
	RegisterTheme("dark", DefaultTheme())
	RegisterTheme("light", LightTheme())
}

// Current theme management
var currentTheme atomic.Value // stores *Theme

// SetTheme sets the current active theme.
func SetTheme(t *Theme) {
	if t != nil {
		currentTheme.Store(t)
	}
}

// GetCurrentTheme returns the currently active theme.
func GetCurrentTheme() *Theme {
	if t := currentTheme.Load(); t != nil {
		return t.(*Theme)
	}
	return DefaultTheme()
}

// StyleResolver resolves theme values to concrete styles.
type StyleResolver struct {
	theme *Theme
}

// NewStyleResolver creates a new style resolver for a theme.
func NewStyleResolver(t *Theme) *StyleResolver {
	if t == nil {
		t = DefaultTheme()
	}
	return &StyleResolver{theme: t}
}

// ResolveColor resolves a color name to ANSI color name.
func (s *StyleResolver) ResolveColor(name string) string {
	if color, ok := s.theme.Colors[name]; ok {
		return color
	}
	return s.theme.Colors["text"]
}

// ResolveSpacing resolves a spacing name to pixels.
func (s *StyleResolver) ResolveSpacing(name string) int {
	if space, ok := s.theme.Spacing[name]; ok {
		return space
	}
	return 4 // default sm
}
