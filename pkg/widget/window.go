package widget

import (
	"strings"

	"github.com/akzj/lumina/pkg/render"
)

// WindowState is the internal state for a Window widget (drag/resize tracking).
type WindowState struct {
	Dragging   bool
	Resizing   bool
	DragStartX int
	DragStartY int
	OrigX      int
	OrigY      int
	OrigW      int
	OrigH      int
}

// Window is the built-in Window widget with drag and resize support.
// Props: title (string), x/y (int), width/height (int)
// OnChange fires with "close", or {type="move", x=N, y=N}, or {type="resize", width=N, height=N}
var Window = &Widget{
	Name: "Window",
	NewState: func() any {
		return &WindowState{}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme

		title, _ := props["title"].(string)
		if title == "" {
			title = "Window"
		}

		x := intProp(props, "x", 0)
		y := intProp(props, "y", 0)
		w := intProp(props, "width", 40)
		h := intProp(props, "height", 15)

		children := make([]*render.Node, 0, 4)

		// Title bar: use hbox with title on left, close button on right
		// Content width inside border = w - 2 (border left + right)
		contentW := w - 2
		children = append(children, &render.Node{
			Type: "hbox",
			Style: render.Style{
				Background: t.Primary,
				Height:     1,
				Width:      contentW,
				Right:      -1, Bottom: -1,
			},
			Children: []*render.Node{
				{
					Type:    "text",
					Content: title,
					Style: render.Style{
						Foreground: t.PrimaryDark,
						Bold:       true,
						Flex:       1,
						Right:      -1, Bottom: -1,
					},
				},
				{
					Type:    "text",
					Content: "✕",
					Style: render.Style{
						Foreground: t.PrimaryDark,
						Width:      1,
						Right:      -1, Bottom: -1,
					},
				},
			},
		})

		// Divider
		divW := contentW
		if divW < 1 {
			divW = 1
		}
		children = append(children, &render.Node{
			Type:    "text",
			Content: strings.Repeat("─", divW),
			Style: render.Style{
				Foreground: t.Surface1,
				Dim:        true,
				Right:      -1, Bottom: -1,
			},
		})

		// Content area: wrap Lua children in a flex box with overflow hidden
		// so content is clipped to available space and resize handle stays at bottom
		if childNodes, ok := props["_childNodes"].([]*render.Node); ok {
			contentBox := &render.Node{
				Type:     "vbox",
				Children: childNodes,
				Style: render.Style{
					Flex:     1,
					Overflow: "hidden",
					Right:    -1, Bottom: -1,
				},
			}
			children = append(children, contentBox)
		}

		// Resize handle at bottom-right
		handlePad := contentW - 1
		if handlePad < 0 {
			handlePad = 0
		}
		children = append(children, &render.Node{
			Type:    "text",
			Content: strings.Repeat(" ", handlePad) + "◢",
			Style: render.Style{
				Foreground: t.Muted,
				Right:      -1, Bottom: -1,
			},
		})

		root := &render.Node{
			Type:     "vbox",
			Children: children,
			Style: render.Style{
				Position:   "absolute",
				Left:       x,
				Top:        y,
				Width:      w,
				Height:     h,
				Border:     "rounded",
				Background: t.Base,
				Right:      -1, Bottom: -1,
			},
		}

		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		s := state.(*WindowState)

		w := intProp(props, "width", 40)
		h := intProp(props, "height", 15)

		switch event.Type {
		case "click":
			// Close button: top-right corner of widget (title bar row)
			relY := event.Y - event.WidgetY
			relX := event.X - event.WidgetX
			if (relY == 0 || relY == 1) && relX >= event.WidgetW-3 {
				event.FireOnChange = "close"
				return false
			}
			// Any other click activates the window (bring to front)
			event.FireOnChange = "activate"
			return false

		case "mousedown":
			relY := event.Y - event.WidgetY
			relX := event.X - event.WidgetX

			if relY == 0 || relY == 1 { // title bar area (border + title)
				// Check for close button (last 2 chars of widget)
				if relX >= event.WidgetW-3 {
					event.FireOnChange = "close"
					return false
				}
				// Start drag
				s.Dragging = true
				s.DragStartX = event.X
				s.DragStartY = event.Y
				s.OrigX = event.WidgetX
				s.OrigY = event.WidgetY
				event.CaptureMouse = true
				return true
			}

			// Check if click is on resize handle (bottom-right corner)
			if relY >= h-2 && relX >= w-3 {
				s.Resizing = true
				s.DragStartX = event.X
				s.DragStartY = event.Y
				s.OrigW = w
				s.OrigH = h
				event.CaptureMouse = true
				return true
			}

		case "mousemove":
			if s.Dragging {
				dx := event.X - s.DragStartX
				dy := event.Y - s.DragStartY
				newX := s.OrigX + dx
				newY := s.OrigY + dy
				event.FireOnChange = map[string]any{
					"type": "move",
					"x":    newX,
					"y":    newY,
				}
				return true
			}
			if s.Resizing {
				dx := event.X - s.DragStartX
				dy := event.Y - s.DragStartY
				newW := max(10, s.OrigW+dx)
				newH := max(5, s.OrigH+dy)
				event.FireOnChange = map[string]any{
					"type":   "resize",
					"width":  newW,
					"height": newH,
				}
				return true
			}

		case "mouseup":
			if s.Dragging || s.Resizing {
				s.Dragging = false
				s.Resizing = false
				return true
			}
		}
		return false
	},
}

// intProp extracts an integer property from props with a default value.
// Handles float64 (Lua numbers), int64, and int types.
func intProp(props map[string]any, key string, def int) int {
	if v, ok := props[key].(float64); ok {
		return int(v)
	}
	if v, ok := props[key].(int64); ok {
		return int(v)
	}
	if v, ok := props[key].(int); ok {
		return v
	}
	return def
}
