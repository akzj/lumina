package widget

import "github.com/akzj/lumina/pkg/render"

// RadioState is the internal state for a Radio widget.
type RadioState struct {
	Checked bool
	Hovered bool
}

// Radio is the built-in Radio button widget.
// Props: label (string), value (string), checked (bool, controlled), disabled (bool)
// onChange fires with the "value" prop (string) on click or Space keydown.
// Radio never toggles off — it only selects. Group exclusivity is managed by
// the parent component (Lua side) via controlled mode.
var Radio = &Widget{
	Name: "Radio",
	NewState: func() any {
		return &RadioState{}
	},
	Render: func(props map[string]any, state any) any {
		s := state.(*RadioState)
		label, _ := props["label"].(string)
		disabled, _ := props["disabled"].(bool)

		// Controlled mode: if "checked" prop is explicitly set, use it
		checked := s.Checked
		if v, ok := props["checked"]; ok {
			if b, ok := v.(bool); ok {
				checked = b
			}
		}

		// Determine radio visual
		indicator := "( )"
		if checked {
			indicator = "(●)"
		}

		// Colors
		fg := "#CDD6F4"
		indicatorFg := "#89B4FA"
		if disabled {
			fg = "#6C7086"
			indicatorFg = "#6C7086"
		} else if s.Hovered {
			indicatorFg = "#B4BEFE"
		}

		indicatorNode := &render.Node{
			Type:    "text",
			Content: indicator,
			Style: render.Style{
				Foreground: indicatorFg,
				Right:      -1,
				Bottom:     -1,
			},
		}

		children := []*render.Node{indicatorNode}

		if label != "" {
			labelNode := &render.Node{
				Type:    "text",
				Content: " " + label,
				Style: render.Style{
					Foreground: fg,
					Right:      -1,
					Bottom:     -1,
				},
			}
			if disabled {
				labelNode.Style.Dim = true
			}
			children = append(children, labelNode)
		}

		boxNode := &render.Node{
			Type:      "hbox",
			Children:  children,
			Focusable: !disabled,
			Disabled:  disabled,
			Style: render.Style{
				Right:  -1,
				Bottom: -1,
			},
		}
		for _, ch := range children {
			ch.Parent = boxNode
		}

		return boxNode
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		s := state.(*RadioState)
		disabled, _ := props["disabled"].(bool)
		if disabled {
			return false
		}

		switch event.Type {
		case "mouseenter":
			if !s.Hovered {
				s.Hovered = true
				return true
			}
		case "mouseleave":
			if s.Hovered {
				s.Hovered = false
				return true
			}
		case "click":
			// Radio only selects, never deselects
			if !s.Checked {
				s.Checked = true
			}
			// Always fire onChange with value prop (even if already checked)
			value, _ := props["value"].(string)
			event.FireOnChange = value
			return true
		case "keydown":
			if event.Key != " " && event.Key != "Enter" {
				return false
			}
			if !s.Checked {
				s.Checked = true
			}
			value, _ := props["value"].(string)
			event.FireOnChange = value
			return true
		}
		return false
	},
}
