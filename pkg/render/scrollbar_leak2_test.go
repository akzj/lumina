package render

import (
	"strings"
	"testing"
)

// TestScrollbarLeakDebugCoords checks the actual coordinates after layout
// to understand if any node inside the right panel has X=0.
func TestScrollbarLeakDebugCoords(t *testing.T) {
	bufW, bufH := 214, 50 // wider, like real terminal

	// Left panel
	leftRows := make([]*Node, 5)
	for i := range leftRows {
		leftRows[i] = &Node{
			Type:    "text",
			Content: "▶ Agent " + string(rune('A'+i)),
			Style:   Style{Foreground: "#ffffff", WidthPercent: 100},
		}
	}
	scrollView1 := &Node{Type: "vbox", ID: "agent-scroll", Style: Style{
		Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#221d2e",
	}, Children: leftRows}
	scrollComp1 := &Node{Type: "component", Children: []*Node{scrollView1}}
	leftSidebar := &Node{Type: "vbox", Style: Style{
		WidthPercent: 100, HeightPercent: 100, Background: "#120d1a",
	}, Children: []*Node{scrollComp1}}
	leftSidebarComp := &Node{Type: "component", Children: []*Node{leftSidebar}}
	pane1 := &Node{Type: "vbox", Style: Style{Overflow: "hidden", Width: 33, HeightPercent: 100}}
	pane1.Children = []*Node{leftSidebarComp}

	// Right panel: nested deeper (like real app: component > vbox > component > scrollView)
	bigText := strings.Repeat("这是一段很长的中文文本，包含宽字符。", 500) // Chinese wide chars
	rightRows := make([]*Node, 15)
	for i := range rightRows {
		if i == 7 {
			rightRows[i] = &Node{
				Type:    "text",
				Content: bigText,
				Style:   Style{Foreground: "#ffffff", WidthPercent: 100},
			}
		} else {
			rightRows[i] = &Node{
				Type:    "text",
				Content: "消息 " + string(rune('0'+i)),
				Style:   Style{Foreground: "#ffffff", WidthPercent: 100},
			}
		}
	}

	scrollView2 := &Node{Type: "vbox", ID: "chat-scroll", Style: Style{
		Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#1a1626",
		ScrollbarThumbColor: "#cdd6f4",
		ScrollbarTrackColor: "#313244",
	}, Children: rightRows}
	scrollComp2 := &Node{Type: "component", Children: []*Node{scrollView2}}
	centerPanel := &Node{Type: "vbox", Style: Style{
		WidthPercent: 100, HeightPercent: 100, Background: "#1a1626",
	}, Children: []*Node{scrollComp2}}
	centerPanelComp := &Node{Type: "component", Children: []*Node{centerPanel}}
	pane2 := &Node{Type: "vbox", Style: Style{Overflow: "hidden", Flex: 1, HeightPercent: 100}}
	pane2.Children = []*Node{centerPanelComp}

	// Root
	root := &Node{Type: "hbox", Style: Style{WidthPercent: 100, HeightPercent: 100}, Children: []*Node{pane1, pane2}}

	var setParents func(n *Node)
	setParents = func(n *Node) {
		for _, c := range n.Children {
			c.Parent = n
			setParents(c)
		}
	}
	setParents(root)

	layoutViewportW = bufW
	layoutViewportH = bufH
	LayoutFull(root, 0, 0, bufW, bufH)

	t.Logf("root: X=%d W=%d", root.X, root.W)
	t.Logf("pane1: X=%d W=%d", pane1.X, pane1.W)
	t.Logf("pane2: X=%d W=%d", pane2.X, pane2.W)
	t.Logf("scrollView2: X=%d W=%d ScrollHeight=%d", scrollView2.X, scrollView2.W, scrollView2.ScrollHeight)

	// Check all nodes inside pane2 — none should have X < 33
	var checkCoords func(n *Node, path string)
	checkCoords = func(n *Node, path string) {
		if n.W > 0 && n.H > 0 {
			if n.X < 33 {
				t.Errorf("Node %s type=%q id=%q has X=%d (< 33)! W=%d",
					path, n.Type, n.ID, n.X, n.W)
			}
		}
		for i, c := range n.Children {
			checkCoords(c, path+"/"+c.Type+"["+string(rune('0'+i))+"]")
		}
	}
	checkCoords(pane2, "pane2")

	// Now do the paint and check for leaks
	buf := NewCellBuffer(bufW, bufH)
	root.PaintDirty = true
	PaintDirty(buf, root)

	scrollbarTrack := "#313244"
	scrollbarThumb := "#cdd6f4"
	rightPanelBG := "#1a1626"

	// Check X=0..32 for any right-panel content
	leakCount := 0
	for x := 0; x < 33; x++ {
		for y := 0; y < bufH; y++ {
			c := buf.Get(x, y)
			if c.BG == scrollbarTrack || c.BG == scrollbarThumb || c.FG == scrollbarThumb {
				if leakCount < 5 {
					t.Errorf("LEAK: X=%d Y=%d BG=%q FG=%q Ch=%c", x, y, c.BG, c.FG, c.Ch)
				}
				leakCount++
			}
			if c.BG == rightPanelBG && x < 33 {
				// right panel bg in left panel area — could be leak
				if leakCount < 5 {
					t.Logf("INFO: X=%d Y=%d has right panel BG=%q Ch=%c", x, y, c.BG, c.Ch)
				}
			}
		}
	}
	if leakCount > 0 {
		t.Errorf("Total leaks found: %d", leakCount)
	}

	// Now scroll and test dirty paint
	maxScroll := computeMaxScrollY(scrollView2)
	t.Logf("maxScrollY=%d", maxScroll)

	// Dirty paint scenarios
	scenarios := []struct {
		name    string
		scrollY int
		dirty   []*Node
	}{
		{"scroll_mid", maxScroll / 2, []*Node{scrollView2}},
		{"scroll_max", maxScroll, []*Node{scrollView2}},
		{"scroll+pane2", maxScroll / 2, []*Node{scrollView2, pane2}},
		{"scroll+root", maxScroll, []*Node{root}},
	}

	for _, sc := range scenarios {
		scrollView2.ScrollY = sc.scrollY
		for _, n := range sc.dirty {
			n.PaintDirty = true
		}
		PaintDirty(buf, root)

		for x := 0; x < 33; x++ {
			for y := 0; y < bufH; y++ {
				c := buf.Get(x, y)
				if c.BG == scrollbarTrack || c.BG == scrollbarThumb || c.FG == scrollbarThumb {
					t.Errorf("[%s] LEAK: X=%d Y=%d BG=%q FG=%q Ch=%c", sc.name, x, y, c.BG, c.FG, c.Ch)
					goto nextScenario
				}
			}
		}
	nextScenario:
	}
}
