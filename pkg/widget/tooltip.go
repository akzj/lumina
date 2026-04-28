package widget

import "github.com/akzj/lumina/pkg/render"

// TooltipState is the internal state for a Tooltip widget.
type TooltipState struct {
	Hovered bool
}

// Tooltip is the built-in Tooltip widget.
// Props: label (string), text (string)
// Shows tooltip text above the trigger label on hover.
var Tooltip = &Widget{
	Name: "Tooltip",
	NewState: func() any {
		return &TooltipState{}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		s := state.(*TooltipState)

		label, _ := props["label"].(string)
		if label == "" {
			label = "hover me"
		}
		tipText, _ := props["text"].(string)
		if tipText == "" {
			tipText = "Tooltip"
		}

		children := make([]*render.Node, 0, 2)

		// Tooltip text (shown above trigger when hovered)
		if s.Hovered {
			children = append(children, &render.Node{
				Type:    "text",
				Content: " " + tipText + " ",
				Style: render.Style{
					Background: t.Surface1,
					Foreground: t.Text,
					Right:      -1,
					Bottom:     -1,
				},
			})
		}

		// Trigger text
		triggerFg := t.Text
		if s.Hovered {
			triggerFg = t.Primary
		}
		children = append(children, &render.Node{
			Type:    "text",
			Content: label,
			Style: render.Style{
				Foreground: triggerFg,
				Right:      -1,
				Bottom:     -1,
			},
		})

		root := &render.Node{
			Type:     "vbox",
			Children: children,
			Style: render.Style{
				Right:  -1,
				Bottom: -1,
			},
		}

		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		s := state.(*TooltipState)
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
		}
		return false
	},
}
