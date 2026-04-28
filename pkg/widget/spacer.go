package widget

import "github.com/akzj/lumina/pkg/render"

// Spacer is the built-in Spacer widget — a flexible space filler.
// Props: size (float64 — fixed size in characters), direction (string — "horizontal" or "vertical", default "vertical")
var Spacer = &Widget{
	Name: "Spacer",
	NewState: func() any {
		return nil
	},
	Render: func(props map[string]any, state any) any {
		direction, _ := props["direction"].(string)

		style := render.Style{
			Flex:   1,
			Right:  -1,
			Bottom: -1,
		}

		if size, ok := props["size"].(float64); ok && size > 0 {
			s := int(size)
			style.Flex = 0
			if direction == "horizontal" {
				style.Width = s
			} else {
				style.Height = s
			}
		}

		return &render.Node{
			Type:  "box",
			Style: style,
		}
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		return false
	},
}
