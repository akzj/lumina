package widget

import (
	"strings"

	"github.com/akzj/lumina/pkg/render"
)

// DropdownState is the internal state for a Dropdown widget.
type DropdownState struct {
	Open        bool
	Hovered     bool
	Highlighted int // keyboard-highlighted item in dropdown (-1 = none)
}

// Dropdown is the built-in Dropdown widget.
// A trigger button that opens a dropdown menu for actions.
// Props: label (string — trigger text), items ([]any — same format as Menu items)
// onChange fires with the selected item index (int).
var Dropdown = &Widget{
	Name: "Dropdown",
	NewState: func() any {
		return &DropdownState{Highlighted: -1}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		s := state.(*DropdownState)

		label, _ := props["label"].(string)
		if label == "" {
			label = "Menu"
		}
		items := readMenuItems(props)

		arrow := "▼"
		if s.Open {
			arrow = "▲"
		}

		triggerBg := t.Surface0
		triggerFg := t.Text
		if s.Hovered && !s.Open {
			triggerBg = t.Surface1
		}

		// Trigger button
		trigger := &render.Node{
			Type: "hbox",
			Children: []*render.Node{
				{
					Type:    "text",
					Content: " " + label + " ",
					Style: render.Style{
						Foreground: triggerFg,
						Right:      -1,
						Bottom:     -1,
					},
				},
				{
					Type:    "text",
					Content: arrow + " ",
					Style: render.Style{
						Foreground: t.Muted,
						Right:      -1,
						Bottom:     -1,
					},
				},
			},
			Style: render.Style{
				Background: triggerBg,
				Border:     "rounded",
				Right:      -1,
				Bottom:     -1,
			},
		}

		rootChildren := []*render.Node{trigger}

		// Dropdown menu (when open)
		if s.Open && len(items) > 0 {
			optionNodes := make([]*render.Node, len(items))
			for i, item := range items {
				if item.Divider {
					optionNodes[i] = &render.Node{
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

				fg := t.Text
				bg := ""
				if item.Disabled {
					fg = t.Muted
				} else if i == s.Highlighted {
					bg = t.Surface1
				}

				content := " "
				if item.Icon != "" {
					content += item.Icon + " "
				}
				content += item.Label + " "

				optionNodes[i] = &render.Node{
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

			dropdown := &render.Node{
				Type:     "vbox",
				Children: optionNodes,
				Style: render.Style{
					Background: t.Surface0,
					Border:     "single",
					Position:   "absolute",
					Top:        3, // below trigger (border top + content + border bottom)
					Left:       0,
					Right:      -1,
					Bottom:     -1,
				},
			}
			rootChildren = append(rootChildren, dropdown)
		}

		root := &render.Node{
			Type:      "box",
			Focusable: true,
			Children:  rootChildren,
			Style: render.Style{
				Right:  -1,
				Bottom: -1,
			},
		}

		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		s := state.(*DropdownState)
		items := readMenuItems(props)

		switch event.Type {
		case "click":
			if s.Open {
				// Check if click is on a dropdown item
				triggerHeight := 3 // border + content + border
				optionY := event.Y - (event.WidgetY + triggerHeight)
				if optionY >= 0 && optionY < len(items) {
					item := items[optionY]
					if !item.Disabled && !item.Divider {
						s.Open = false
						s.Highlighted = -1
						event.FireOnChange = optionY
						return true
					}
				}
				// Click elsewhere → close
				s.Open = false
				s.Highlighted = -1
				return true
			}
			// Open dropdown
			s.Open = true
			s.Highlighted = findNextMenuItem(items, -1, 1)
			return true

		case "keydown":
			if !s.Open {
				if event.Key == "Enter" || event.Key == " " {
					s.Open = true
					s.Highlighted = findNextMenuItem(items, -1, 1)
					return true
				}
				return false
			}
			switch event.Key {
			case "ArrowDown", "j":
				s.Highlighted = findNextMenuItem(items, s.Highlighted, 1)
				return true
			case "ArrowUp", "k":
				s.Highlighted = findNextMenuItem(items, s.Highlighted, -1)
				return true
			case "Enter":
				if s.Highlighted >= 0 && s.Highlighted < len(items) {
					item := items[s.Highlighted]
					if !item.Disabled && !item.Divider {
						s.Open = false
						event.FireOnChange = s.Highlighted
						s.Highlighted = -1
						return true
					}
				}
			case "Escape":
				s.Open = false
				s.Highlighted = -1
				return true
			}

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
		case "blur":
			if s.Open {
				s.Open = false
				s.Highlighted = -1
				return true
			}
		}
		return false
	},
}
