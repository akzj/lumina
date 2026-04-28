package widget

import "github.com/akzj/lumina/pkg/render"

// SelectOption is a single option in a Select dropdown.
type SelectOption struct {
	Label string
	Value string
}

// readOptions extracts options from props["options"].
// Options come from Lua as []any where each element is map[string]any
// with "label" and "value" keys.
func readOptions(props map[string]any) []SelectOption {
	raw, ok := props["options"]
	if !ok {
		return nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	opts := make([]SelectOption, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		label, _ := m["label"].(string)
		value, _ := m["value"].(string)
		opts = append(opts, SelectOption{Label: label, Value: value})
	}
	return opts
}

// SelectState is the internal state for a Select widget.
type SelectState struct {
	Open        bool // dropdown is visible
	Selected    int  // currently selected option index (-1 = none)
	Highlighted int  // keyboard-highlighted option index
	Hovered     bool // mouse hover on the select trigger
}

// findSelectedIndex returns the index of the option matching value, or -1.
func findSelectedIndex(options []SelectOption, value string) int {
	for i, opt := range options {
		if opt.Value == value {
			return i
		}
	}
	return -1
}

// setParents recursively sets Parent pointers on all children.
func setParents(node *render.Node) {
	for _, ch := range node.Children {
		ch.Parent = node
		setParents(ch)
	}
}

// Select is the built-in Select (dropdown) widget.
// Props: options ([]map[string]any with "label"/"value"), value (string, controlled),
//
//	placeholder (string), disabled (bool)
//
// onChange fires with the selected option's value (string).
// The dropdown is rendered using absolute positioning within the widget's node tree.
var Select = &Widget{
	Name: "Select",
	NewState: func() any {
		return &SelectState{
			Selected:    -1,
			Highlighted: 0,
		}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		s := state.(*SelectState)
		disabled, _ := props["disabled"].(bool)
		placeholder, _ := props["placeholder"].(string)
		if placeholder == "" {
			placeholder = "Select..."
		}
		options := readOptions(props)

		// Controlled mode: if "value" prop is set, override selected index
		selected := s.Selected
		if v, ok := props["value"]; ok {
			if vs, ok := v.(string); ok {
				selected = findSelectedIndex(options, vs)
			}
		}

		// Determine display text
		displayText := placeholder
		if selected >= 0 && selected < len(options) {
			displayText = options[selected].Label
		}

		// Arrow indicator
		arrow := "▼"
		if s.Open {
			arrow = "▲"
		}

		// Colors
		bg := t.Surface0
		fg := t.Text
		borderColor := t.Surface1
		if disabled {
			bg = t.Base
			fg = t.Muted
			borderColor = t.Surface0
		} else if s.Hovered || s.Open {
			borderColor = t.Primary
		}

		arrowFg := t.Primary
		if disabled {
			arrowFg = t.Muted
		}

		// Placeholder color
		textFg := fg
		if selected < 0 || selected >= len(options) {
			textFg = t.Muted
		}

		textNode := &render.Node{
			Type:    "text",
			Content: displayText,
			Style: render.Style{
				Foreground: textFg,
				Flex:       1,
				Right:      -1,
				Bottom:     -1,
			},
		}

		arrowNode := &render.Node{
			Type:    "text",
			Content: arrow,
			Style: render.Style{
				Foreground: arrowFg,
				Right:      -1,
				Bottom:     -1,
			},
		}

		trigger := &render.Node{
			Type:     "hbox",
			Children: []*render.Node{textNode, arrowNode},
			Style: render.Style{
				Right:  -1,
				Bottom: -1,
			},
		}

		rootChildren := []*render.Node{trigger}

		// Dropdown (only when open)
		if s.Open && len(options) > 0 {
			optionNodes := make([]*render.Node, len(options))
			for i, opt := range options {
				optBg := ""
				optFg := t.Text
				content := opt.Label

				if i == s.Highlighted {
					optBg = t.Surface1
				}
				if i == selected {
					content = content + " ✓"
					optFg = t.Primary
				}

				optionNodes[i] = &render.Node{
					Type:    "text",
					Content: content,
					Style: render.Style{
						Background: optBg,
						Foreground: optFg,
						Right:      -1,
						Bottom:     -1,
					},
				}
			}

			dropdown := &render.Node{
				Type:     "vbox",
				Children: optionNodes,
				Style: render.Style{
					Position:   "absolute",
					Top:        1, // below trigger row
					Left:       0,
					Border:     "single",
					Background: t.Base,
					Foreground: borderColor,
					Right:      -1,
					Bottom:     -1,
				},
			}
			rootChildren = append(rootChildren, dropdown)
		}

		root := &render.Node{
			Type:      "box",
			Children:  rootChildren,
			Focusable: !disabled,
			Disabled:  disabled,
			Style: render.Style{
				Border:       "rounded",
				Background:   bg,
				Foreground:   borderColor,
				PaddingLeft:  1,
				PaddingRight: 1,
				Right:        -1,
				Bottom:       -1,
			},
		}

		_ = borderColor // used in style above
		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		s := state.(*SelectState)
		disabled, _ := props["disabled"].(bool)
		if disabled {
			return false
		}
		options := readOptions(props)

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
			if s.Open {
				// The trigger area is 1 row (row 0 relative to widget).
				// The dropdown starts at row 1 (relative to widget content area).
				// With border, the widget's WidgetY is the top border row.
				// Content starts at WidgetY+1 (border top). Trigger is at WidgetY+1.
				// Dropdown options start at WidgetY+2 (trigger row + 1) + 1 (dropdown border).
				// Each option is 1 row.
				triggerHeight := 3 // 1 border top + 1 trigger row + 1 border bottom
				dropdownBorderTop := 1
				optionY := event.Y - (event.WidgetY + triggerHeight + dropdownBorderTop)
				if optionY >= 0 && optionY < len(options) {
					// Clicked on an option
					s.Selected = optionY
					s.Highlighted = optionY
					s.Open = false
					event.FireOnChange = options[optionY].Value
					return true
				}
				// Click elsewhere → close
				s.Open = false
				return true
			}
			// Open dropdown
			s.Open = true
			if s.Selected >= 0 {
				s.Highlighted = s.Selected
			} else {
				s.Highlighted = 0
			}
			return true

		case "keydown":
			switch event.Key {
			case " ", "Enter":
				if !s.Open {
					// Open dropdown
					s.Open = true
					if s.Selected >= 0 {
						s.Highlighted = s.Selected
					} else {
						s.Highlighted = 0
					}
					return true
				}
				// Select highlighted item
				if len(options) > 0 && s.Highlighted >= 0 && s.Highlighted < len(options) {
					s.Selected = s.Highlighted
					s.Open = false
					event.FireOnChange = options[s.Selected].Value
					return true
				}
			case "Escape":
				if s.Open {
					s.Open = false
					return true
				}
			case "ArrowUp":
				if s.Open && len(options) > 0 {
					s.Highlighted--
					if s.Highlighted < 0 {
						s.Highlighted = len(options) - 1
					}
					return true
				}
			case "ArrowDown":
				if s.Open && len(options) > 0 {
					s.Highlighted++
					if s.Highlighted >= len(options) {
						s.Highlighted = 0
					}
					return true
				}
			}

		case "blur":
			// Close dropdown when losing focus
			if s.Open {
				s.Open = false
				return true
			}
		}
		return false
	},
}
