package widget

import (
	"fmt"

	"github.com/akzj/lumina/pkg/render"
)

// ListState is the internal state for a List widget.
type ListState struct {
	SelectedIndex int // -1 = none
	HoveredIndex  int // -1 = none
}

// readListItems extracts items from props["items"].
// Items come from Lua as []any where each element is either a string
// or a map[string]any with a "label" key.
func readListItems(props map[string]any) []string {
	raw, ok := props["items"]
	if !ok {
		return nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	items := make([]string, 0, len(list))
	for _, item := range list {
		switch v := item.(type) {
		case string:
			items = append(items, v)
		case map[string]any:
			label, _ := v["label"].(string)
			items = append(items, label)
		}
	}
	return items
}

// List is the built-in List widget.
// Props: items ([]any — strings or maps with "label"), selectable (bool, default true),
//
//	showIndex (bool, default false)
//
// onChange fires with the selected item index (int).
var List = &Widget{
	Name: "List",
	NewState: func() any {
		return &ListState{SelectedIndex: -1, HoveredIndex: -1}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		s := state.(*ListState)
		items := readListItems(props)
		selectable := true
		if v, ok := props["selectable"].(bool); ok {
			selectable = v
		}
		showIndex, _ := props["showIndex"].(bool)

		children := make([]*render.Node, len(items))
		for i, item := range items {
			prefix := "  "
			fg := t.Text
			bg := ""

			if selectable && i == s.SelectedIndex {
				prefix = "● "
				fg = t.Primary
			}
			if i == s.HoveredIndex {
				bg = t.Surface1
			}

			label := item
			if showIndex {
				label = fmt.Sprintf("%d. %s", i+1, item)
			}

			children[i] = &render.Node{
				Type:    "text",
				Content: prefix + label,
				Style: render.Style{
					Foreground: fg,
					Background: bg,
					Right:      -1,
					Bottom:     -1,
				},
			}
		}

		root := &render.Node{
			Type:      "vbox",
			Focusable: selectable,
			Children:  children,
			Style: render.Style{
				Border: "rounded",
				Right:  -1,
				Bottom: -1,
			},
		}

		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		s := state.(*ListState)
		items := readListItems(props)
		selectable := true
		if v, ok := props["selectable"].(bool); ok {
			selectable = v
		}

		switch event.Type {
		case "keydown":
			if !selectable || len(items) == 0 {
				return false
			}
			switch event.Key {
			case "ArrowDown", "j":
				s.SelectedIndex++
				if s.SelectedIndex >= len(items) {
					s.SelectedIndex = 0
				}
				event.FireOnChange = s.SelectedIndex + 1 // 1-based for Lua onChange
				return true
			case "ArrowUp", "k":
				s.SelectedIndex--
				if s.SelectedIndex < 0 {
					s.SelectedIndex = len(items) - 1
				}
				event.FireOnChange = s.SelectedIndex + 1
				return true
			case "Enter":
				if s.SelectedIndex >= 0 && s.SelectedIndex < len(items) {
					event.FireOnChange = s.SelectedIndex + 1
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
