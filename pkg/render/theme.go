package render

// Theme defines the color palette for all widgets.
// Semantic tokens, not raw colors.
type Theme struct {
	// Base colors
	Base     string // main background
	Surface0 string // elevated surface
	Surface1 string // secondary surface, borders
	Surface2 string // subtle borders

	// Text colors
	Text  string // default foreground
	Muted string // disabled/placeholder text

	// Accent colors
	Primary     string // primary accent
	PrimaryDark string // dark text on primary bg
	Hover       string // hover state
	Pressed     string // pressed state

	// Semantic colors (for Alert/Badge components)
	Success string // green
	Warning string // yellow
	Error   string // red
}

// DefaultTheme is Catppuccin Mocha.
var DefaultTheme = &Theme{
	Base:        "#1E1E2E",
	Surface0:    "#313244",
	Surface1:    "#45475A",
	Surface2:    "#585B70",
	Text:        "#CDD6F4",
	Muted:       "#6C7086",
	Primary:     "#89B4FA",
	PrimaryDark: "#1E1E2E",
	Hover:       "#B4BEFE",
	Pressed:     "#74C7EC",
	Success:     "#A6E3A1",
	Warning:     "#F9E2AF",
	Error:       "#F38BA8",
}

// CurrentTheme is the active theme. Widgets read from this.
// Can be changed at runtime via lumina.setTheme().
var CurrentTheme = DefaultTheme

// LatteTheme is Catppuccin Latte (light).
var LatteTheme = &Theme{
	Base:        "#EFF1F5",
	Surface0:    "#CCD0DA",
	Surface1:    "#BCC0CC",
	Surface2:    "#ACB0BE",
	Text:        "#4C4F69",
	Muted:       "#8C8FA1",
	Primary:     "#1E66F5",
	PrimaryDark: "#EFF1F5",
	Hover:       "#7287FD",
	Pressed:     "#209FB5",
	Success:     "#40A02B",
	Warning:     "#DF8E1D",
	Error:       "#D20F39",
}

// NordTheme is the Nord color scheme.
var NordTheme = &Theme{
	Base:        "#2E3440",
	Surface0:    "#3B4252",
	Surface1:    "#434C5E",
	Surface2:    "#4C566A",
	Text:        "#ECEFF4",
	Muted:       "#7B88A1",
	Primary:     "#88C0D0",
	PrimaryDark: "#2E3440",
	Hover:       "#8FBCBB",
	Pressed:     "#81A1C1",
	Success:     "#A3BE8C",
	Warning:     "#EBCB8B",
	Error:       "#BF616A",
}

// DraculaTheme is the Dracula color scheme.
var DraculaTheme = &Theme{
	Base:        "#282A36",
	Surface0:    "#44475A",
	Surface1:    "#6272A4",
	Surface2:    "#7C85A3",
	Text:        "#F8F8F2",
	Muted:       "#6272A4",
	Primary:     "#BD93F9",
	PrimaryDark: "#282A36",
	Hover:       "#CAA9FA",
	Pressed:     "#FF79C6",
	Success:     "#50FA7B",
	Warning:     "#F1FA8C",
	Error:       "#FF5555",
}

// BuiltinThemes maps theme names to Theme pointers.
var BuiltinThemes = map[string]*Theme{
	"mocha":   DefaultTheme,
	"latte":   LatteTheme,
	"nord":    NordTheme,
	"dracula": DraculaTheme,
}

// ThemeToMap converts a Theme struct to a map[string]string for the Lua API.
func ThemeToMap(t *Theme) map[string]string {
	return map[string]string{
		"base":        t.Base,
		"surface0":    t.Surface0,
		"surface1":    t.Surface1,
		"surface2":    t.Surface2,
		"text":        t.Text,
		"muted":       t.Muted,
		"primary":     t.Primary,
		"primaryDark": t.PrimaryDark,
		"hover":       t.Hover,
		"pressed":     t.Pressed,
		"success":     t.Success,
		"warning":     t.Warning,
		"error":       t.Error,
	}
}

// SetThemeByName sets CurrentTheme to a built-in theme by name.
// Returns true if the theme was found, false otherwise.
func SetThemeByName(name string) bool {
	if t, ok := BuiltinThemes[name]; ok {
		CurrentTheme = t
		return true
	}
	return false
}
