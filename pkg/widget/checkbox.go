package widget

import "github.com/akzj/lumina/pkg/render"

// CheckboxState is the internal state for a Checkbox widget.
type CheckboxState struct {
	Checked bool
	Hovered bool
}

// Checkbox is the built-in Checkbox widget.
// Props: label (string), checked (bool, controlled), disabled (bool)
// onChange fires with the new checked value (bool) on click or Space keydown.
var Checkbox = &Widget{
	Name: "Checkbox",
	NewState: func() any {
		return &CheckboxState{}
	},
	Render: func(props map[string]any, state any) any {
		s := state.(*CheckboxState)
		label, _ := props["label"].(string)
		disabled, _ := props["disabled"].(bool)

		// Controlled mode: if "checked" prop is explicitly set, use it
		checked := s.Checked
		if v, ok := props["checked"]; ok {
			if b, ok := v.(bool); ok {
				checked = b
			}
		}

		// Determine checkbox visual
		indicator := "[ ]"
		if checked {
			indicator = "[x]"
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
		s := state.(*CheckboxState)
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
