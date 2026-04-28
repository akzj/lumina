package widget

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
