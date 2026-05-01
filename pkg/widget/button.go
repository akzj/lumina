package widget

// Deprecated: This Go widget is superseded by lua/lux/button.lua which provides
// richer features (severity, appearance, group, split) without Go dependency.
// The Go widget remains for backward compatibility but will be removed in a future version.

import "github.com/akzj/lumina/pkg/render"

// ButtonState is the internal state for a Button widget.
type ButtonState struct {
	Hovered bool
	Pressed bool
}

// Button is the built-in Button widget.
var Button = &Widget{
	Name: "Button",
	NewState: func() any {
		return &ButtonState{}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		s := state.(*ButtonState)
		label, _ := props["label"].(string)
		if label == "" {
			label = "Button"
		}
		disabled, _ := props["disabled"].(bool)
		variant, _ := props["variant"].(string)
		if variant == "" {
			variant = "primary"
		}

		// Determine colors based on variant + state
		bg := t.Primary
		fg := t.PrimaryDark
		switch variant {
		case "outline":
			bg = ""
			fg = t.Primary
		case "ghost":
			bg = ""
			fg = t.Text
		case "secondary":
			bg = t.Surface1
			fg = t.Text
		}

		if disabled {
			bg = t.Surface0
			fg = t.Muted
		} else if s.Pressed {
			bg = t.Pressed
		} else if s.Hovered {
			bg = t.Hover
		}

		// Determine border style
		border := "rounded"
		if variant == "ghost" {
			border = "none"
		}

		textNode := &render.Node{
			Type:    "text",
			Content: label,
			Style: render.Style{
				Foreground: fg,
				Right:      -1,
				Bottom:     -1,
			},
		}
		if disabled {
			textNode.Style.Dim = true
		}

		boxNode := &render.Node{
			Type:      "hbox",
			Children:  []*render.Node{textNode},
			Focusable: !disabled,
			Disabled:  disabled,
			Style: render.Style{
				Border:       border,
				Background:   bg,
				PaddingLeft:  1,
				PaddingRight: 1,
				Justify:      "center",
				Align:        "center",
				Right:        -1,
				Bottom:       -1,
			},
		}
		textNode.Parent = boxNode

		return boxNode
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		s := state.(*ButtonState)
		disabled, _ := props["disabled"].(bool)
		if disabled {
			return false
		}

		switch event.Type {
		case "mouseenter":
			if !s.Hovered {
				s.Hovered = true
				return true // state changed, needs re-render
			}
		case "mouseleave":
			if s.Hovered || s.Pressed {
				s.Hovered = false
				s.Pressed = false
				return true
			}
		case "mousedown":
			if !s.Pressed {
				s.Pressed = true
				return true
			}
		case "mouseup":
			if s.Pressed {
				s.Pressed = false
				return true
			}
		}
		return false
	},
}
