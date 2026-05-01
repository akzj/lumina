package widget

// Deprecated: This Go widget is superseded by lua/lux/pagination.lua (pure Lua).
// The Go widget remains for backward compatibility but will be removed in a future version.

import (
	"fmt"

	"github.com/akzj/lumina/pkg/render"
)

// PaginationState is the internal state for a Pagination widget.
type PaginationState struct {
	HoveredButton string // "prev", "next", or page number string
}

// calcPageRange computes the visible page range given current page, total pages,
// and maximum visible buttons.
func calcPageRange(current, total, maxVisible int) (int, int) {
	if total <= maxVisible {
		return 1, total
	}
	half := maxVisible / 2
	start := current - half
	end := current + half
	if maxVisible%2 == 0 {
		end--
	}

	if start < 1 {
		start = 1
		end = maxVisible
	}
	if end > total {
		end = total
		start = total - maxVisible + 1
		if start < 1 {
			start = 1
		}
	}
	return start, end
}

// Pagination is the built-in Pagination widget.
// Props: page (float64, 1-based), totalPages (float64), maxVisible (float64, default 5)
// onChange fires with the new page number (int).
var Pagination = &Widget{
	Name: "Pagination",
	NewState: func() any {
		return &PaginationState{}
	},
	Render: func(props map[string]any, state any) any {
		t := CurrentTheme

		page := 1
		if p, ok := props["page"].(float64); ok && p > 0 {
			page = int(p)
		}
		totalPages := 1
		if tp, ok := props["totalPages"].(float64); ok && tp > 0 {
			totalPages = int(tp)
		}
		maxVisible := 5
		if mv, ok := props["maxVisible"].(float64); ok && mv > 0 {
			maxVisible = int(mv)
		}

		children := make([]*render.Node, 0, maxVisible+2)

		// Prev button
		prevFg := t.Text
		if page <= 1 {
			prevFg = t.Muted
		}
		children = append(children, &render.Node{
			Type:    "text",
			Content: " ◀ ",
			Style: render.Style{
				Foreground: prevFg,
				Right:      -1,
				Bottom:     -1,
			},
		})

		// Page numbers
		startPage, endPage := calcPageRange(page, totalPages, maxVisible)
		for i := startPage; i <= endPage; i++ {
			fg := t.Text
			bg := ""
			bold := false
			if i == page {
				fg = t.PrimaryDark
				bg = t.Primary
				bold = true
			}
			children = append(children, &render.Node{
				Type:    "text",
				Content: fmt.Sprintf(" %d ", i),
				Style: render.Style{
					Foreground: fg,
					Background: bg,
					Bold:       bold,
					Right:      -1,
					Bottom:     -1,
				},
			})
		}

		// Next button
		nextFg := t.Text
		if page >= totalPages {
			nextFg = t.Muted
		}
		children = append(children, &render.Node{
			Type:    "text",
			Content: " ▶ ",
			Style: render.Style{
				Foreground: nextFg,
				Right:      -1,
				Bottom:     -1,
			},
		})

		root := &render.Node{
			Type:      "hbox",
			Focusable: true,
			Children:  children,
			Style: render.Style{
				Right:  -1,
				Bottom: -1,
			},
		}

		setParents(root)
		return root
	},
	OnEvent: func(props map[string]any, state any, event *render.WidgetEvent) bool {
		page := 1
		if p, ok := props["page"].(float64); ok && p > 0 {
			page = int(p)
		}
		totalPages := 1
		if tp, ok := props["totalPages"].(float64); ok && tp > 0 {
			totalPages = int(tp)
		}

		switch event.Type {
		case "keydown":
			switch event.Key {
			case "ArrowLeft", "h":
				if page > 1 {
					event.FireOnChange = page - 1
					return true
				}
			case "ArrowRight", "l":
				if page < totalPages {
					event.FireOnChange = page + 1
					return true
				}
			}
		}
		return false
	},
}
