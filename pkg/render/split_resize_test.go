package render

import (
	"strings"
	"testing"
)

func TestSplitPaneResizeLargeTextMatchesFullRepaint(t *testing.T) {
	const bufW, bufH = 180, 40

	build := func(rightW int) (*Node, *Node) {
		leftRows := make([]*Node, 12)
		for i := range leftRows {
			leftRows[i] = &Node{
				Type:    "text",
				Content: "● agent-row-" + string(rune('a'+i)) + " · leaf",
				Style:   Style{Foreground: "#ffffff", WidthPercent: 100, Height: 1},
			}
		}
		leftScroll := &Node{Type: "vbox", Style: Style{
			Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#120d1a",
		}, Children: leftRows}
		pane1 := &Node{Type: "vbox", ID: "split-pane-1", Style: Style{
			Width: 32, HeightPercent: 100, Overflow: "hidden", Background: "#120d1a",
		}, Children: []*Node{leftScroll}}

		divider1 := &Node{Type: "vbox", Style: Style{Width: 1, Background: "#38605c"}}
		divider2 := &Node{Type: "vbox", Style: Style{Width: 1, Background: "#38605c"}}

		bigText := strings.Repeat("Rewrite a1c08a842bb505d407544f8bbbd6b0121ccd309 (179/893) (4 seconds passed, remaining 15 predicted)    ", 700)
		rows := make([]*Node, 18)
		for i := range rows {
			content := "normal row"
			if i == 8 {
				content = bigText
			}
			rows[i] = &Node{Type: "text", Content: content, Style: Style{Foreground: "#cdd6f4", WidthPercent: 100}}
		}
		chatScroll := &Node{Type: "vbox", ID: "chat-scroll", Style: Style{
			Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#1a1626",
		}, Children: rows}
		pane2 := &Node{Type: "vbox", ID: "split-pane-2", Style: Style{
			Flex: 1, HeightPercent: 100, Overflow: "hidden", Background: "#1a1626",
		}, Children: []*Node{chatScroll}}

		pane3 := &Node{Type: "vbox", ID: "split-pane-3", Style: Style{
			Width: rightW, HeightPercent: 100, Overflow: "hidden", Background: "#120d1a",
		}, Children: []*Node{
			{Type: "text", Content: "LLM Configuration", Style: Style{Height: 1, WidthPercent: 100}},
			{Type: "text", Content: "Provider : anthropic", Style: Style{Height: 1, WidthPercent: 100}},
		}}

		root := &Node{Type: "hbox", Style: Style{WidthPercent: 100, HeightPercent: 100},
			Children: []*Node{pane1, divider1, pane2, divider2, pane3}}
		setParentsRecursive(root)
		return root, pane3
	}

	root, pane3 := build(36)
	LayoutFull(root, 0, 0, bufW, bufH)
	chatScroll := findSplitResizeNodeByID(root, "chat-scroll")
	if chatScroll == nil {
		t.Fatal("missing chat scroll")
	}
	chatScroll.ScrollY = computeMaxScrollY(chatScroll) / 2
	root.PaintDirty = true
	buf := NewCellBuffer(bufW, bufH)
	PaintDirty(buf, root)
	terminal := cloneCellBuffer(buf)
	buf.ResetStats()

	pane3.Style.Width = 58
	pane3.MarkLayoutDirty()
	LayoutIncremental(root)
	PaintDirty(buf, root)
	actualScrollY := chatScroll.ScrollY
	stats := buf.Stats()
	applyDirtyRect(terminal, buf, stats.DirtyX, stats.DirtyY, stats.DirtyW, stats.DirtyH)

	expectedRoot, _ := build(58)
	LayoutFull(expectedRoot, 0, 0, bufW, bufH)
	expectedScroll := findSplitResizeNodeByID(expectedRoot, "chat-scroll")
	if expectedScroll == nil {
		t.Fatal("missing expected chat scroll")
	}
	expectedScroll.ScrollY = actualScrollY
	expected := NewCellBuffer(bufW, bufH)
	expectedRoot.PaintDirty = true
	PaintDirty(expected, expectedRoot)

	for y := 0; y < bufH; y++ {
		for x := 0; x < bufW; x++ {
			got, want := terminal.Get(x, y), expected.Get(x, y)
			if got != want {
				t.Fatalf("cell mismatch at (%d,%d): got %+v want %+v", x, y, got, want)
			}
		}
	}
}

func cloneCellBuffer(src *CellBuffer) *CellBuffer {
	dst := NewCellBuffer(src.Width(), src.Height())
	for y := 0; y < src.Height(); y++ {
		for x := 0; x < src.Width(); x++ {
			dst.Set(x, y, src.Get(x, y))
		}
	}
	dst.ResetStats()
	return dst
}

func applyDirtyRect(dst, src *CellBuffer, x, y, w, h int) {
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			dst.Set(col, row, src.Get(col, row))
		}
	}
	dst.ResetStats()
}

func findSplitResizeNodeByID(node *Node, id string) *Node {
	if node == nil {
		return nil
	}
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := findSplitResizeNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}
