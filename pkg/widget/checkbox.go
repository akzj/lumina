package widget

// Prefer lua/lux/checkbox.lua for new code (pure Lua, no Go dependency).
// This Go widget remains functional for backward compatibility and direct usage
// via lumina.Checkbox, but new applications should use the Lua version.

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
		t := CurrentTheme
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
		fg := t.Text
		indicatorFg := t.Primary
		if disabled {
			fg = t.Muted
			indicatorFg = t.Muted
		} else if s.Hovered {
			indicatorFg = t.Hover
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
			// Sync internal state from controlled prop before toggling
			if v, ok := props["checked"]; ok {
				if b, ok := v.(bool); ok {
					s.Checked = b
				}
			}
			s.Checked = !s.Checked
			event.FireOnChange = s.Checked
			return true
		case "keydown":
			if event.Key != " " && event.Key != "Enter" {
				return false
			}
			// Sync internal state from controlled prop before toggling
			if v, ok := props["checked"]; ok {
				if b, ok := v.(bool); ok {
					s.Checked = b
				}
			}
			s.Checked = !s.Checked
			event.FireOnChange = s.Checked
			return true
		}
		return false
	},
}
