package widget

import "github.com/akzj/lumina/pkg/render"

// SwitchState is the internal state for a Switch widget.
type SwitchState struct {
	Checked bool
	Hovered bool
}

// Switch is the built-in Switch (toggle) widget.
// Props: label (string), checked (bool, controlled), disabled (bool)
// onChange fires with the new checked value (bool) on click or Space keydown.
var Switch = &Widget{
	Name: "Switch",
	NewState: func() any {
		return &SwitchState{}
	},
	Render: func(props map[string]any, state any) any {
		s := state.(*SwitchState)
		label, _ := props["label"].(string)
		disabled, _ := props["disabled"].(bool)

		// Controlled mode: if "checked" prop is explicitly set, use it
		checked := s.Checked
		if v, ok := props["checked"]; ok {
			if b, ok := v.(bool); ok {
				checked = b
			}
		}

		// Determine switch track visual
		track := "[○  ]"
		trackBg := "#45475A"
		if checked {
			track = "[  ●]"
			trackBg = "#89B4FA"
		}

		// Colors
		fg := "#CDD6F4"
		trackFg := "#CDD6F4"
		if disabled {
			fg = "#6C7086"
			trackFg = "#6C7086"
			trackBg = "#313244"
		} else if s.Hovered {
			if checked {
				trackBg = "#B4BEFE"
			} else {
				trackBg = "#585B70"
			}
		}

		trackNode := &render.Node{
			Type:    "text",
			Content: track,
			Style: render.Style{
				Foreground: trackFg,
				Background: trackBg,
				Right:      -1,
				Bottom:     -1,
			},
		}

		children := []*render.Node{trackNode}

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
		s := state.(*SwitchState)
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
			s.Checked = !s.Checked
			event.FireOnChange = s.Checked
			return true
		case "keydown":
			if event.Key != " " && event.Key != "Enter" {
				return false
			}
			s.Checked = !s.Checked
			event.FireOnChange = s.Checked
			return true
		}
		return false
	},
}
