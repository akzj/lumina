package widget

import (
	"strings"

	"github.com/akzj/lumina/pkg/render"
)

// DialogState is the internal state for a Dialog widget.
type DialogState struct {
	Open bool
}

// Dialog is the built-in Dialog widget.
// Props: open (bool), title (string), message (string), width (float64/int)
// OnChange fires with "close" when Escape is pressed.
var Dialog = &Widget{
	Name: "Dialog",
	NewState: func() any {
		return &DialogState{}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		s := state.(*DialogState)
		open, _ := props["open"].(bool)

		if !open {
			s.Open = false
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

		title, _ := props["title"].(string)
		if title == "" {
			title = "Dialog"
		}
		message, _ := props["message"].(string)
		width := 40
		if w, ok := props["width"].(float64); ok && w > 0 {
			width = int(w)
		}

		children := make([]*render.Node, 0, 4)

		// Title
		children = append(children, &render.Node{
			Type:    "text",
			Content: title,
			Style: render.Style{
				Foreground: t.Primary,
				Bold:       true,
				Right:      -1,
				Bottom:     -1,
			},
		})

		// Divider
		dividerWidth := width - 4 // account for border + padding
		if dividerWidth < 1 {
			dividerWidth = 1
		}
		children = append(children, &render.Node{
			Type:    "text",
			Content: strings.Repeat("─", dividerWidth),
			Style: render.Style{
				Foreground: t.Surface1,
				Dim:        true,
				Right:      -1,
				Bottom:     -1,
			},
		})

		// Message
		if message != "" {
			children = append(children, &render.Node{
				Type:    "text",
				Content: message,
				Style: render.Style{
					Foreground: t.Text,
					Right:      -1,
					Bottom:     -1,
					MarginTop:  1,
				},
			})
		}

		root := &render.Node{
			Type:      "vbox",
			Focusable: true,
			Children:  children,
			Style: render.Style{
				Background:   t.Surface0,
				Border:       "rounded",
				Padding:      1,
				Width:        width,
				Right:        -1,
				Bottom:       -1,
			},
		}

		setParents(root)
		s.Open = open
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		open, _ := props["open"].(bool)
		if !open {
			return false
		}

		if event.Type == "keydown" && event.Key == "Escape" {
			event.FireOnChange = "close"
			return false
		}
		return false
	},
}
