package widget

import "github.com/akzj/lumina/pkg/render"

// Label is the built-in Label widget.
// Props: text (string)
// Label renders simple text with no interactive behavior.
// The "for" association (focusing a related input) is handled by
// the user in Lua via an onClick callback, not by the widget itself.
var Label = &Widget{
	Name: "Label",
	NewState: func() any {
		return nil // Label has no state
	},
	Render: func(props map[string]any, state any) any {
		text, _ := props["text"].(string)
		if text == "" {
			text = "Label"
		}

		fg := "#CDD6F4"

		node := &render.Node{
			Type:    "text",
			Content: text,
			Style: render.Style{
				Foreground: fg,
				Right:      -1,
				Bottom:     -1,
			},
		}
		return node
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		// Label has no interactive behavior
		return false
	},
}
