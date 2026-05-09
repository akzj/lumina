package render

import (
	"strings"
	"testing"
)

// TestScrollbarLeakSplitPane reproduces the bug with exact SplitPane structure.
// The key difference from previous tests: SplitPane uses dividers (1-col vbox)
// between panes, and the tab bar has overflow:scroll.
func TestScrollbarLeakSplitPane(t *testing.T) {
	bufW, bufH := 214, 55

	// === LEFT SIDEBAR (pane1) ===
	leftRows := make([]*Node, 8)
	for i := range leftRows {
		leftRows[i] = &Node{
			Type:    "text",
			Content: "▶ ◉ agent-" + string(rune('a'+i)) + " · leaf",
			Style:   Style{Foreground: "#ffffff", WidthPercent: 100, Height: 1},
		}
	}
	agentScroll := &Node{Type: "vbox", ID: "agent-scroll", Style: Style{
		Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#1a1626",
		ScrollbarThumbColor: "#cdd6f4",
		ScrollbarTrackColor: "#313244",
	}, Children: leftRows}
	agentScrollComp := &Node{Type: "component", Children: []*Node{agentScroll}}
	agentTreeWrap := &Node{Type: "vbox", Style: Style{
		Flex: 2, WidthPercent: 100, Overflow: "hidden",
	}, Children: []*Node{agentScrollComp}}
	agentTreeWrapComp := &Node{Type: "component", Children: []*Node{agentTreeWrap}}

	leftSidebar := &Node{Type: "vbox", Style: Style{
		WidthPercent: 100, HeightPercent: 100, Background: "#120d1a",
	}, Children: []*Node{agentTreeWrapComp}}
	leftSidebarComp := &Node{Type: "component", Children: []*Node{leftSidebar}}

	// SplitPane pane1 wrapper
	pane1 := &Node{Type: "vbox", ID: "split-pane-1", Style: Style{
		Width: 32, HeightPercent: 100, Overflow: "hidden",
	}, Children: []*Node{leftSidebarComp}}

	// === DIVIDER 1 ===
	divider1 := &Node{Type: "vbox", ID: "split-div-1", Style: Style{
		Width: 1, Background: "#38605c",
	}}

	// === CENTER PANEL (pane2) ===
	// Tab bar (overflow:scroll, scrollbar:none)
	tabTexts := make([]*Node, 5)
	for i := range tabTexts {
		tabTexts[i] = &Node{
			Type:    "text",
			Content: "[ ● tab-" + string(rune('0'+i)) + " ]",
			Style:   Style{Height: 1, Foreground: "#cdd6f4"},
		}
	}
	tabBar := &Node{Type: "hbox", ID: "editor-tab-bar", Style: Style{
		Height: 1, WidthPercent: 100, Overflow: "scroll", Scrollbar: "none",
	}, Children: tabTexts}

	// Chat messages with one 64K text
	bigText := strings.Repeat("Rewrite abc123def456 (354/893) (8 seconds passed, remaining 12 predicted)    ", 800)
	chatRows := make([]*Node, 20)
	for i := range chatRows {
		if i == 10 {
			chatRows[i] = &Node{
				Type:    "text",
				Content: bigText,
				Style:   Style{Foreground: "#cdd6f4", WidthPercent: 100},
			}
		} else {
			chatRows[i] = &Node{
				Type:    "text",
				Content: "Normal message " + string(rune('0'+i%10)),
				Style:   Style{Foreground: "#cdd6f4", WidthPercent: 100, Height: 1},
			}
		}
	}
	chatScroll := &Node{Type: "vbox", ID: "chat-scroll", Style: Style{
		Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#1a1626",
		ScrollbarThumbColor: "#cdd6f4",
		ScrollbarTrackColor: "#313244",
	}, Children: chatRows}
	chatScrollComp := &Node{Type: "component", Children: []*Node{chatScroll}}

	chatPanel := &Node{Type: "vbox", Style: Style{
		Flex: 1, WidthPercent: 100,
	}, Children: []*Node{chatScrollComp}}
	chatPanelComp := &Node{Type: "component", Children: []*Node{chatPanel}}

	centerPanel := &Node{Type: "vbox", Style: Style{
		WidthPercent: 100, HeightPercent: 100,
	}, Children: []*Node{tabBar, chatPanelComp}}
	centerPanelComp := &Node{Type: "component", Children: []*Node{centerPanel}}

	// SplitPane pane2 wrapper
	pane2 := &Node{Type: "vbox", ID: "split-pane-2", Style: Style{
		Flex: 1, HeightPercent: 100, Overflow: "hidden",
	}, Children: []*Node{centerPanelComp}}

	// === ROOT (SplitPane hbox) ===
	root := &Node{Type: "hbox", Style: Style{Flex: 1, WidthPercent: 100, HeightPercent: 100},
		Children: []*Node{pane1, divider1, pane2}}

	// Set parent pointers
	var setParents func(n *Node)
	setParents = func(n *Node) {
		for _, c := range n.Children {
			c.Parent = n
			setParents(c)
		}
	}
	setParents(root)

	// Layout
	layoutViewportW = bufW
	layoutViewportH = bufH
	LayoutFull(root, 0, 0, bufW, bufH)

	t.Logf("pane1: X=%d W=%d H=%d", pane1.X, pane1.W, pane1.H)
	t.Logf("divider1: X=%d W=%d", divider1.X, divider1.W)
	t.Logf("pane2: X=%d W=%d H=%d", pane2.X, pane2.W, pane2.H)
	t.Logf("chatScroll: X=%d W=%d H=%d ScrollHeight=%d", chatScroll.X, chatScroll.W, chatScroll.H, chatScroll.ScrollHeight)
	t.Logf("tabBar: X=%d W=%d H=%d ScrollWidth=%d", tabBar.X, tabBar.W, tabBar.H, tabBar.ScrollWidth)

	// Initial full paint
	buf := NewCellBuffer(bufW, bufH)
	root.PaintDirty = true
	PaintDirty(buf, root)

	scrollbarTrack := "#313244"
	scrollbarThumb := "#cdd6f4"

	checkLeak := func(label string) bool {
		leaked := false
		for x := 0; x < 33; x++ {
			for y := 0; y < bufH; y++ {
				c := buf.Get(x, y)
				if c.BG == scrollbarTrack || c.BG == scrollbarThumb {
					if !leaked {
						t.Errorf("[%s] LEAK at X=%d Y=%d BG=%q FG=%q Ch=%c", label, x, y, c.BG, c.FG, c.Ch)
					}
					leaked = true
				}
				// Also check for scrollbar characters with scrollbar FG
				if c.FG == scrollbarThumb && (c.Ch == '█' || c.Ch == '░') {
					if !leaked {
						t.Errorf("[%s] LEAK at X=%d Y=%d BG=%q FG=%q Ch=%c", label, x, y, c.BG, c.FG, c.Ch)
					}
					leaked = true
				}
			}
		}
		return leaked
	}

	checkLeak("initial")

	// Scroll chat to the big text area
	maxScroll := computeMaxScrollY(chatScroll)
	t.Logf("chatScroll maxScrollY=%d", maxScroll)

	// Test various scroll positions and dirty combinations
	testCases := []struct {
		name       string
		scrollY    int
		dirtyNodes []*Node
	}{
		{"scroll_10", maxScroll / 10, []*Node{chatScroll}},
		{"scroll_30", maxScroll * 3 / 10, []*Node{chatScroll}},
		{"scroll_50", maxScroll / 2, []*Node{chatScroll}},
		{"scroll_70", maxScroll * 7 / 10, []*Node{chatScroll}},
		{"scroll_90", maxScroll * 9 / 10, []*Node{chatScroll}},
		{"scroll_max", maxScroll, []*Node{chatScroll}},
		// Dirty pane2 simultaneously (simulates layout change)
		{"scroll+pane2_dirty", maxScroll / 2, []*Node{chatScroll, pane2}},
		// Dirty root (full repaint path)
		{"scroll+root_dirty", maxScroll / 2, []*Node{root}},
		// Dirty chatScroll + parent component
		{"scroll+comp_dirty", maxScroll / 2, []*Node{chatScroll, chatPanelComp}},
		// Multiple scrolls in sequence without clearing
		{"rapid_scroll_1", maxScroll / 4, []*Node{chatScroll}},
		{"rapid_scroll_2", maxScroll / 2, []*Node{chatScroll}},
		{"rapid_scroll_3", maxScroll * 3 / 4, []*Node{chatScroll}},
	}

	for _, tc := range testCases {
		chatScroll.ScrollY = tc.scrollY
		for _, n := range tc.dirtyNodes {
			n.PaintDirty = true
		}
		PaintDirty(buf, root)
		if checkLeak(tc.name) {
			// Print what's at the scrollbar position for reference
			sbX := chatScroll.X + chatScroll.W - 1
			t.Logf("  Expected scrollbar at X=%d", sbX)
			break
		}
	}

	// Test: incremental layout (simulates component re-render)
	chatScroll.ScrollY = maxScroll / 2
	chatScroll.LayoutDirty = true
	chatScroll.PaintDirty = true
	LayoutIncremental(root)
	PaintDirty(buf, root)
	checkLeak("incremental_layout")

	// Test: tab bar scroll (horizontal scroll in sibling)
	tabBar.ScrollX = 10
	tabBar.PaintDirty = true
	chatScroll.ScrollY = maxScroll / 2
	chatScroll.PaintDirty = true
	PaintDirty(buf, root)
	checkLeak("tab_scroll+chat_scroll")
}
