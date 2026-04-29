package widget

import (
	"github.com/akzj/lumina/pkg/render"
)

// ScrollViewState is the internal state for a ScrollView widget.
type ScrollViewState struct {
	// Reserved for future use (scrollbar dragging, etc.)
}

// ScrollView is the built-in scrollable container widget.
// It wraps children in a vbox with overflow:"scroll", provides keyboard scrolling
// (ArrowUp/ArrowDown/Up/Down/j/k, PageUp/PageDown, Home/End), and the engine paints
// a scrollbar in the
// reserved right column.
//
// Props:
//   - style: merged into the root node style (overflow is forced to "scroll")
//   - showScrollbar: bool (default true) — whether to show the scrollbar
//   - _childNodes: []*render.Node — children (set automatically by the engine)
//
// OnEvent consumes keydown events for scrolling and sets ScrollBy on the
// WidgetEvent for the engine to apply.
var ScrollView = &Widget{
	Name: "ScrollView",
	NewState: func() any {
		return &ScrollViewState{}
	},
	Render: func(props map[string]any, state any) any {
		// Build style from props, forcing overflow = "scroll"
		style := render.Style{
			Overflow: "scroll",
		}

		// Merge user-provided style
		if userStyle, ok := props["style"].(map[string]any); ok {
			mergeStyle(&style, userStyle)
		}
		// Force overflow to scroll regardless of user style
		style.Overflow = "scroll"

		// Collect children from _childNodes
		var children []*render.Node
		if childNodes, ok := props["_childNodes"].([]*render.Node); ok {
			children = childNodes
		}

		root := &render.Node{
			Type:      "vbox",
			Children:  children,
			Style:     style,
			Focusable: true,
		}

		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		if event.Type != "keydown" {
			return false
		}

		// Calculate visible height for page scrolling
		visibleH := event.WidgetH
		if visibleH < 1 {
			visibleH = 1
		}

		switch event.Key {
		case "ArrowUp", "Up", "k":
			event.ScrollBy = -1
			return true
		case "ArrowDown", "Down", "j":
			event.ScrollBy = 1
			return true
		case "PageUp":
			event.ScrollBy = -visibleH
			return true
		case "PageDown":
			event.ScrollBy = visibleH
			return true
		case "Home":
			// Scroll to top: use a large negative value, engine will clamp to 0
			event.ScrollBy = -999999
			return true
		case "End":
			// Scroll to bottom: use a large positive value, engine will clamp to max
			event.ScrollBy = 999999
			return true
		}

		return false
	},
}

// mergeStyle applies Lua-style map values onto a render.Style struct.
func mergeStyle(s *render.Style, m map[string]any) {
	if v, ok := m["width"]; ok {
		s.Width = intFromAny(v)
	}
	if v, ok := m["height"]; ok {
		s.Height = intFromAny(v)
	}
	if v, ok := m["flex"]; ok {
		s.Flex = intFromAny(v)
	}
	if v, ok := m["border"]; ok {
		if str, ok := v.(string); ok {
			s.Border = str
		}
	}
	if v, ok := m["background"]; ok {
		if str, ok := v.(string); ok {
			s.Background = str
		}
	}
	if v, ok := m["foreground"]; ok {
		if str, ok := v.(string); ok {
			s.Foreground = str
		}
	}
	if v, ok := m["paddingTop"]; ok {
		s.PaddingTop = intFromAny(v)
	}
	if v, ok := m["paddingBottom"]; ok {
		s.PaddingBottom = intFromAny(v)
	}
	if v, ok := m["paddingLeft"]; ok {
		s.PaddingLeft = intFromAny(v)
	}
	if v, ok := m["paddingRight"]; ok {
		s.PaddingRight = intFromAny(v)
	}
	if v, ok := m["position"]; ok {
		if str, ok := v.(string); ok {
			s.Position = str
		}
	}
	if v, ok := m["left"]; ok {
		s.Left = intFromAny(v)
	}
	if v, ok := m["top"]; ok {
		s.Top = intFromAny(v)
	}
}

// intFromAny converts a Lua number (float64 or int64) to int.
func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int64:
		return int(n)
	case int:
		return n
	}
	return 0
}
