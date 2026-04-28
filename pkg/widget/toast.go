package widget

import "github.com/akzj/lumina/pkg/render"

// ToastState is the internal state for a Toast widget.
type ToastState struct {
	Visible bool
}

// Toast is the built-in Toast widget.
// Props: message (string), variant (string: "default"/"success"/"warning"/"error"),
//
//	visible (bool)
//
// onChange fires with "dismiss" when clicked or Escape is pressed.
var Toast = &Widget{
	Name: "Toast",
	NewState: func() any {
		return &ToastState{}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		visible, _ := props["visible"].(bool)

		if !visible {
			return &render.Node{
				Type: "box",
				Style: render.Style{
					Width:  0,
					Height: 0,
					Right:  -1,
					Bottom: -1,
				},
			}
		}

		message, _ := props["message"].(string)
		if message == "" {
			message = "Notification"
		}
		variant, _ := props["variant"].(string)

		// Icon + color based on variant
		var icon, fg string
		switch variant {
		case "success":
			icon = "✓ "
			fg = t.Success
		case "warning":
			icon = "⚠ "
			fg = t.Warning
		case "error":
			icon = "✗ "
			fg = t.Error
		default:
			icon = "ℹ "
			fg = t.Primary
		}

		children := []*render.Node{
			{
				Type:    "text",
				Content: icon,
				Style: render.Style{
					Foreground: fg,
					Bold:       true,
					Right:      -1,
					Bottom:     -1,
				},
			},
			{
				Type:    "text",
				Content: message,
				Style: render.Style{
					Foreground: t.Text,
					Flex:       1,
					Right:      -1,
					Bottom:     -1,
				},
			},
			{
				Type:    "text",
				Content: " ✕",
				Style: render.Style{
					Foreground: t.Muted,
					Right:      -1,
					Bottom:     -1,
				},
			},
		}

		root := &render.Node{
			Type:     "hbox",
			Children: children,
			Style: render.Style{
				Background:   t.Surface0,
				Border:       "rounded",
				PaddingLeft:  1,
				PaddingRight: 1,
				Right:        -1,
				Bottom:       -1,
			},
		}

		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		visible, _ := props["visible"].(bool)
		if !visible {
			return false
		}

		switch event.Type {
		case "click":
			event.FireOnChange = "dismiss"
			return false
		case "keydown":
			if event.Key == "Escape" {
				event.FireOnChange = "dismiss"
				return false
			}
		}
		return false
	},
}
