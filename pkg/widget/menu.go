package widget

import (
	"strings"

	"github.com/akzj/lumina/pkg/render"
)

// MenuItem represents a single item in a Menu or Dropdown.
type MenuItem struct {
	Label    string
	Icon     string
	Disabled bool
	Divider  bool
}

// readMenuItems extracts menu items from props["items"].
// Items come from Lua as []any where each element is map[string]any.
func readMenuItems(props map[string]any) []MenuItem {
	raw, ok := props["items"]
	if !ok {
		return nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	items := make([]MenuItem, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		label, _ := m["label"].(string)
		icon, _ := m["icon"].(string)
		disabled, _ := m["disabled"].(bool)
		divider, _ := m["divider"].(bool)
		items = append(items, MenuItem{Label: label, Icon: icon, Disabled: disabled, Divider: divider})
	}
	return items
}

// findNextMenuItem finds the next selectable item in direction (+1 or -1), wrapping around.
// Returns current if no valid item found.
func findNextMenuItem(items []MenuItem, current, direction int) int {
	n := len(items)
	if n == 0 {
		return current
	}
	for i := 1; i <= n; i++ {
		idx := ((current + direction*i) % n + n) % n
		if !items[idx].Divider && !items[idx].Disabled {
			return idx
		}
	}
	return current // no valid item found
}

// MenuState is the internal state for a Menu widget.
type MenuState struct {
	SelectedIndex int // currently selected/active item
	HoveredIndex  int // mouse hover (-1 = none)
}

// Menu is the built-in Menu widget.
// Props: items ([]any — maps with "label", "icon", "disabled", "divider"),
//
//	selected (float64 — controlled selected index), compact (bool — no border)
//
// onChange fires with the selected item index (int).
var Menu = &Widget{
	Name: "Menu",
	NewState: func() any {
		return &MenuState{SelectedIndex: 0, HoveredIndex: -1}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		s := state.(*MenuState)
		items := readMenuItems(props)

		// Controlled mode
		selected := s.SelectedIndex
		if v, ok := props["selected"].(float64); ok {
			selected = int(v)
		}
		compact, _ := props["compact"].(bool)

		children := make([]*render.Node, len(items))
		for i, item := range items {
			if item.Divider {
				children[i] = &render.Node{
					Type:    "text",
					Content: strings.Repeat("─", 14),
					Style: render.Style{
						Foreground: t.Surface1,
						Dim:        true,
						Right:      -1,
						Bottom:     -1,
					},
				}
				continue
			}

			prefix := "  "
			fg := t.Text
			bg := ""

			if item.Disabled {
				fg = t.Muted
			} else if i == selected {
				fg = t.Primary
				bg = t.Surface0
				prefix = "▸ "
			} else if i == s.HoveredIndex {
				bg = t.Surface1
			}

			content := prefix
			if item.Icon != "" {
				content += item.Icon + " "
			}
			content += item.Label

			children[i] = &render.Node{
				Type:    "text",
				Content: content,
				Style: render.Style{
					Foreground: fg,
					Background: bg,
					Right:      -1,
					Bottom:     -1,
				},
			}
		}

		border := "rounded"
		if compact {
			border = ""
		}

		root := &render.Node{
			Type:      "vbox",
			Focusable: true,
			Children:  children,
			Style: render.Style{
				Border: border,
				Right:  -1,
				Bottom: -1,
			},
		}

		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		s := state.(*MenuState)
		items := readMenuItems(props)

		switch event.Type {
		case "keydown":
			if len(items) == 0 {
				return false
			}
			switch event.Key {
			case "ArrowDown", "j":
				next := findNextMenuItem(items, s.SelectedIndex, 1)
				if next != s.SelectedIndex {
					s.SelectedIndex = next
					event.FireOnChange = s.SelectedIndex + 1
					return true
				}
			case "ArrowUp", "k":
				next := findNextMenuItem(items, s.SelectedIndex, -1)
				if next != s.SelectedIndex {
					s.SelectedIndex = next
					event.FireOnChange = s.SelectedIndex + 1
					return true
				}
			case "Enter":
				if s.SelectedIndex >= 0 && s.SelectedIndex < len(items) {
					if !items[s.SelectedIndex].Disabled && !items[s.SelectedIndex].Divider {
						event.FireOnChange = s.SelectedIndex + 1
					}
				}
				return false
			}
		case "mouseleave":
			if s.HoveredIndex >= 0 {
				s.HoveredIndex = -1
				return true
			}
		}
		return false
	},
}
