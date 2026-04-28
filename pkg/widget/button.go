package widget

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
		bg := "#89B4FA" // primary blue
		fg := "#1E1E2E" // dark text
		switch variant {
		case "outline":
			bg = ""
			fg = "#89B4FA"
		case "ghost":
			bg = ""
			fg = "#CDD6F4"
		case "secondary":
			bg = "#45475A"
			fg = "#CDD6F4"
		}

		if disabled {
			bg = "#313244"
			fg = "#6C7086"
		} else if s.Pressed {
			bg = "#74C7EC" // lighter when pressed
		} else if s.Hovered {
			bg = "#B4BEFE" // lighter when hovered
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
