package widget

import (
	"fmt"
	"strings"

	"github.com/akzj/lumina/pkg/render"
)

// TableColumn describes a single column in a Table widget.
type TableColumn struct {
	Header string
	Key    string
	Width  int
}

// TableState is the internal state for a Table widget.
type TableState struct {
	SelectedRow  int // -1 = none selected
	HoveredRow   int // -1 = none hovered
	ScrollOffset int // for future scrolling support
}

// readColumns extracts columns from props["columns"].
// Columns come from Lua as []any where each element is map[string]any
// with "header", "key", and optional "width" keys.
func readColumns(props map[string]any) []TableColumn {
	raw, ok := props["columns"]
	if !ok {
		return nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	cols := make([]TableColumn, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		header, _ := m["header"].(string)
		key, _ := m["key"].(string)
		width := 10 // default width
		if w, ok := m["width"].(float64); ok && w > 0 {
			width = int(w)
		}
		cols = append(cols, TableColumn{Header: header, Key: key, Width: width})
	}
	return cols
}

// readRows extracts rows from props["rows"].
// Rows come from Lua as []any where each element is map[string]any.
func readRows(props map[string]any) []map[string]any {
	raw, ok := props["rows"]
	if !ok {
		return nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	rows := make([]map[string]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rows = append(rows, m)
	}
	return rows
}

// padRight pads or truncates a string to exactly width characters.
func padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// Table is the built-in Table widget.
// Props: columns ([]map[string]any with "header"/"key"/"width"), rows ([]map[string]any),
//
//	selectable (bool), striped (bool)
//
// onChange fires with the selected row index (int).
var Table = &Widget{
	Name: "Table",
	NewState: func() any {
		return &TableState{SelectedRow: -1, HoveredRow: -1}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme
		s := state.(*TableState)
		columns := readColumns(props)
		rows := readRows(props)
		selectable, _ := props["selectable"].(bool)
		striped, _ := props["striped"].(bool)

		children := make([]*render.Node, 0, len(rows)+2)

		// Header row
		headerCells := make([]*render.Node, len(columns))
		for i, col := range columns {
			text := padRight(col.Header, col.Width)
			headerCells[i] = &render.Node{
				Type:    "text",
				Content: text,
				Style: render.Style{
					Foreground: t.Primary,
					Bold:       true,
					Width:      col.Width,
					Right:      -1,
					Bottom:     -1,
				},
			}
		}
		children = append(children, &render.Node{
			Type:     "hbox",
			Children: headerCells,
			Style:    render.Style{Right: -1, Bottom: -1},
		})

		// Separator
		totalWidth := 0
		for _, col := range columns {
			totalWidth += col.Width
		}
		children = append(children, &render.Node{
			Type:    "text",
			Content: strings.Repeat("─", totalWidth),
			Style: render.Style{
				Foreground: t.Surface1,
				Dim:        true,
				Right:      -1,
				Bottom:     -1,
			},
		})

		// Data rows
		for i, row := range rows {
			cells := make([]*render.Node, len(columns))
			for j, col := range columns {
				value := ""
				if v, ok := row[col.Key]; ok {
					value = fmt.Sprintf("%v", v)
				}
				text := padRight(value, col.Width)
				cells[j] = &render.Node{
					Type:    "text",
					Content: text,
					Style: render.Style{
						Foreground: t.Text,
						Width:      col.Width,
						Right:      -1,
						Bottom:     -1,
					},
				}
			}

			rowBg := ""
			if selectable && i == s.SelectedRow {
				rowBg = t.Primary
				// Change text color for selected row
				for _, cell := range cells {
					cell.Style.Foreground = t.PrimaryDark
				}
			} else if i == s.HoveredRow {
				rowBg = t.Surface1
			} else if striped && i%2 == 1 {
				rowBg = t.Surface0
			}

			children = append(children, &render.Node{
				Type:     "hbox",
				Children: cells,
				Style: render.Style{
					Background: rowBg,
					Right:      -1,
					Bottom:     -1,
				},
			})
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
		s := state.(*TableState)
		selectable, _ := props["selectable"].(bool)
		rows := readRows(props)

		switch event.Type {
		case "keydown":
			if !selectable || len(rows) == 0 {
				return false
			}
			switch event.Key {
			case "ArrowDown", "j":
				s.SelectedRow++
				if s.SelectedRow >= len(rows) {
					s.SelectedRow = 0
				}
				event.FireOnChange = s.SelectedRow
				return true
			case "ArrowUp", "k":
				s.SelectedRow--
				if s.SelectedRow < 0 {
					s.SelectedRow = len(rows) - 1
				}
				event.FireOnChange = s.SelectedRow
				return true
			case "Enter":
				if s.SelectedRow >= 0 && s.SelectedRow < len(rows) {
					event.FireOnChange = s.SelectedRow
				}
				return false
			}
		case "mouseleave":
			if s.HoveredRow >= 0 {
				s.HoveredRow = -1
				return true
			}
		}
		return false
	},
}
