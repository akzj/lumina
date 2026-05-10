package render

// Theme defines the color palette for all widgets.
// Semantic tokens, not raw colors.
type Theme struct {
	// Base colors (backward-compatible keys: base, surface0, surface1, surface2)
	Base     string // main background
	Surface0 string // elevated surface
	Surface1 string // secondary surface, borders
	Surface2 string // subtle borders

	// Extended surface colors (rich theme support)
	BgDark   string // deeper background (bgDark)
	Surface3 string // tertiary surface (surface3)

	// Text colors (backward-compatible: text, muted)
	Text         string // default foreground
	Muted        string // disabled/placeholder text
	TextBright   string // bright/highlighted text
	TextDisabled string // fully disabled text

	// Accent colors (backward-compatible: primary, primaryDark, hover, pressed)
	Primary     string // primary accent
	PrimaryDark string // dark text on primary bg
	PrimaryDim  string // dimmed primary
	Hover       string // hover state
	Pressed     string // pressed state

	// Secondary & accent colors
	Secondary    string // secondary accent
	SecondaryDim string // dimmed secondary
	Accent       string // accent highlight
	AccentDim    string // dimmed accent

	// Border colors
	Border      string // default border
	BorderLight string // light border
	BorderDim   string // dim/subtle border

	// Glow colors (for highlights, focus rings, etc.)
	GlowCyan   string
	GlowPurple string

	// Semantic colors (for Alert/Badge components)
	Success string // green
	Warning string // yellow
	Error   string // red
	Info    string // blue/info

	// Scrollbar colors
	ScrollbarThumb string
	ScrollbarTrack string
}

// DefaultTheme is Catppuccin Mocha.
var DefaultTheme = &Theme{
	Base:     "#1E1E2E",
	Surface0: "#313244",
	Surface1: "#45475A",
	Surface2: "#585B70",

	BgDark:   "#181825",
	Surface3: "#11111B",

	Text:         "#CDD6F4",
	Muted:        "#6C7086",
	TextBright:   "#F2F2F7",
	TextDisabled: "#4C4F69",

	Primary:     "#89B4FA",
	PrimaryDark: "#1E1E2E",
	PrimaryDim:  "#3B5998",
	Hover:       "#B4BEFE",
	Pressed:     "#74C7EC",

	Secondary:    "#CBA6F7",
	SecondaryDim: "#6C3F99",
	Accent:       "#94E2D5",
	AccentDim:    "#3B7A6E",

	Border:      "#585B70",
	BorderLight: "#6C7086",
	BorderDim:   "#45475A",

	GlowCyan:   "#89DCEB",
	GlowPurple: "#CBA6F7",

	Success: "#A6E3A1",
	Warning: "#F9E2AF",
	Error:   "#F38BA8",
	Info:    "#89B4FA",

	ScrollbarThumb: "#585B70",
	ScrollbarTrack: "#313244",
}

// CurrentTheme is the active theme. Widgets read from this.
// Can be changed at runtime via lumina.setTheme().
var CurrentTheme = DefaultTheme

// LatteTheme is Catppuccin Latte (light).
var LatteTheme = &Theme{
	Base:     "#EFF1F5",
	Surface0: "#CCD0DA",
	Surface1: "#BCC0CC",
	Surface2: "#ACB0BE",

	BgDark:   "#E6E9EF",
	Surface3: "#DCE0E8",

	Text:         "#4C4F69",
	Muted:        "#8C8FA1",
	TextBright:   "#1E1E2E",
	TextDisabled: "#BCC0CC",

	Primary:     "#1E66F5",
	PrimaryDark: "#EFF1F5",
	PrimaryDim:  "#1A4FB8",
	Hover:       "#7287FD",
	Pressed:     "#209FB5",

	Secondary:    "#EA76CB",
	SecondaryDim: "#B84DA0",
	Accent:       "#179299",
	AccentDim:    "#126B6F",

	Border:      "#ACB0BE",
	BorderLight: "#8C8FA1",
	BorderDim:   "#BCC0CC",

	GlowCyan:   "#04A5E5",
	GlowPurple: "#8839EF",

	Success: "#40A02B",
	Warning: "#DF8E1D",
	Error:   "#D20F39",
	Info:    "#1E66F5",

	ScrollbarThumb: "#ACB0BE",
	ScrollbarTrack: "#CCD0DA",
}

// NordTheme is the Nord color scheme.
var NordTheme = &Theme{
	Base:     "#2E3440",
	Surface0: "#3B4252",
	Surface1: "#434C5E",
	Surface2: "#4C566A",

	BgDark:   "#242933",
	Surface3: "#2A303C",

	Text:         "#ECEFF4",
	Muted:        "#7B88A1",
	TextBright:   "#FFFFFF",
	TextDisabled: "#4C566A",

	Primary:     "#88C0D0",
	PrimaryDark: "#2E3440",
	PrimaryDim:  "#5A8A9A",
	Hover:       "#8FBCBB",
	Pressed:     "#81A1C1",

	Secondary:    "#B48EAD",
	SecondaryDim: "#7B5E7A",
	Accent:       "#A3BE8C",
	AccentDim:    "#6B8A5A",

	Border:      "#4C566A",
	BorderLight: "#7B88A1",
	BorderDim:   "#434C5E",

	GlowCyan:   "#88C0D0",
	GlowPurple: "#B48EAD",

	Success: "#A3BE8C",
	Warning: "#EBCB8B",
	Error:   "#BF616A",
	Info:    "#81A1C1",

	ScrollbarThumb: "#4C566A",
	ScrollbarTrack: "#3B4252",
}

// DraculaTheme is the Dracula color scheme.
var DraculaTheme = &Theme{
	Base:     "#282A36",
	Surface0: "#44475A",
	Surface1: "#6272A4",
	Surface2: "#7C85A3",

	BgDark:   "#21222C",
	Surface3: "#1A1B24",

	Text:         "#F8F8F2",
	Muted:        "#6272A4",
	TextBright:   "#FFFFFF",
	TextDisabled: "#44475A",

	Primary:     "#BD93F9",
	PrimaryDark: "#282A36",
	PrimaryDim:  "#7B5EA8",
	Hover:       "#CAA9FA",
	Pressed:     "#FF79C6",

	Secondary:    "#FF79C6",
	SecondaryDim: "#B84DA0",
	Accent:       "#50FA7B",
	AccentDim:    "#36B85A",

	Border:      "#6272A4",
	BorderLight: "#7C85A3",
	BorderDim:   "#44475A",

	GlowCyan:   "#8BE9FD",
	GlowPurple: "#BD93F9",

	Success: "#50FA7B",
	Warning: "#F1FA8C",
	Error:   "#FF5555",
	Info:    "#8BE9FD",

	ScrollbarThumb: "#6272A4",
	ScrollbarTrack: "#44475A",
}

// AmberTheme is a warm amber/golden hour theme — yellow-dominant with rich earthy tones.
var AmberTheme = &Theme{
	Base:     "#1a1410",
	Surface0: "#241c16",
	Surface1: "#2e241c",
	Surface2: "#1e1814",

	BgDark:   "#120d0a",
	Surface3: "#3d2a0a",

	Text:         "#e8dcc8",
	Muted:        "#9c8b74",
	TextBright:   "#f5efe0",
	TextDisabled: "#6b5d4a",

	Primary:     "#f59e0b",
	PrimaryDark: "#1a1410",
	PrimaryDim:  "#92400e",
	Hover:       "#fbbf24",
	Pressed:     "#d97706",

	Secondary:    "#f97316",
	SecondaryDim: "#c2410c",
	Accent:       "#fbbf24",
	AccentDim:    "#b45309",

	Border:      "#5c4010",
	BorderLight: "#7c5a18",
	BorderDim:   "#3d2a0a",

	GlowCyan:   "#fbbf24",
	GlowPurple: "#f97316",

	Success: "#4ade80",
	Warning: "#fbbf24",
	Error:   "#ef4444",
	Info:    "#38bdf8",

	ScrollbarThumb: "#f59e0b",
	ScrollbarTrack: "#241c16",
}

// CrimsonTheme is a deep crimson/ruby red theme — bold reds with pink accents.
var CrimsonTheme = &Theme{
	Base:     "#1a1018",
	Surface0: "#241820",
	Surface1: "#2e1e28",
	Surface2: "#1e141c",

	BgDark:   "#120a10",
	Surface3: "#3d1020",

	Text:         "#e8d0d8",
	Muted:        "#9c7c88",
	TextBright:   "#f5e0e8",
	TextDisabled: "#6b5460",

	Primary:     "#e11d48",
	PrimaryDark: "#1a1018",
	PrimaryDim:  "#881337",
	Hover:       "#f43f5e",
	Pressed:     "#be123c",

	Secondary:    "#ec4899",
	SecondaryDim: "#be185d",
	Accent:       "#f43f5e",
	AccentDim:    "#9f1239",

	Border:      "#5c1a30",
	BorderLight: "#7c2440",
	BorderDim:   "#3d1020",

	GlowCyan:   "#fb7185",
	GlowPurple: "#f472b6",

	Success: "#4ade80",
	Warning: "#fbbf24",
	Error:   "#f87171",
	Info:    "#38bdf8",

	ScrollbarThumb: "#e11d48",
	ScrollbarTrack: "#241820",
}

// ForestTheme is a deep forest emerald green theme — rich greens with lime accents.
var ForestTheme = &Theme{
	Base:     "#0f1a14",
	Surface0: "#162418",
	Surface1: "#1e2e20",
	Surface2: "#121e16",

	BgDark:   "#0a120e",
	Surface3: "#103d20",

	Text:         "#c8e8d0",
	Muted:        "#7c9c84",
	TextBright:   "#e0f5e8",
	TextDisabled: "#546b5c",

	Primary:     "#10b981",
	PrimaryDark: "#0f1a14",
	PrimaryDim:  "#064e3b",
	Hover:       "#34d399",
	Pressed:     "#059669",

	Secondary:    "#84cc16",
	SecondaryDim: "#65a30d",
	Accent:       "#34d399",
	AccentDim:    "#047857",

	Border:      "#1a5c30",
	BorderLight: "#247c40",
	BorderDim:   "#103d20",

	GlowCyan:   "#6ee7b7",
	GlowPurple: "#a3e635",

	Success: "#4ade80",
	Warning: "#fbbf24",
	Error:   "#ef4444",
	Info:    "#38bdf8",

	ScrollbarThumb: "#10b981",
	ScrollbarTrack: "#162418",
}

// BuiltinThemes maps theme names to Theme pointers.
var BuiltinThemes = map[string]*Theme{
	"mocha":   DefaultTheme,
	"latte":   LatteTheme,
	"nord":    NordTheme,
	"dracula": DraculaTheme,
	"amber":   AmberTheme,
	"crimson": CrimsonTheme,
	"forest":  ForestTheme,
}

// ThemeToMap converts a Theme struct to a map[string]string for the Lua API.
func ThemeToMap(t *Theme) map[string]string {
	return map[string]string{
		// Base colors
		"base":     t.Base,
		"surface0": t.Surface0,
		"surface1": t.Surface1,
		"surface2": t.Surface2,
		"bgDark":   t.BgDark,
		"surface3": t.Surface3,

		// Text colors
		"text":         t.Text,
		"muted":        t.Muted,
		"textBright":   t.TextBright,
		"textDisabled": t.TextDisabled,

		// Accent colors
		"primary":     t.Primary,
		"primaryDark": t.PrimaryDark,
		"primaryDim":  t.PrimaryDim,
		"hover":       t.Hover,
		"pressed":     t.Pressed,

		// Secondary & accent
		"secondary":    t.Secondary,
		"secondaryDim": t.SecondaryDim,
		"accent":       t.Accent,
		"accentDim":    t.AccentDim,

		// Borders
		"border":      t.Border,
		"borderLight": t.BorderLight,
		"borderDim":   t.BorderDim,

		// Glow
		"glowCyan":   t.GlowCyan,
		"glowPurple": t.GlowPurple,

		// Semantic
		"success": t.Success,
		"warning": t.Warning,
		"error":   t.Error,
		"info":    t.Info,

		// Scrollbar
		"scrollbarThumb": t.ScrollbarThumb,
		"scrollbarTrack": t.ScrollbarTrack,
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
