package render

import (
	"strings"
	"testing"
)

// TestScrollbarLeakToLeftPanel reproduces the bug where scrolling a large text
// in the right panel causes scrollbar colors to appear at X=0 in the left panel.
//
// Setup:
//   root (hbox, W=80, H=40)
//     pane1 (vbox, overflow:hidden, W=32)
//       leftSidebar (vbox, bg="#120d1a", width=100%, height=100%)
//         scrollView1 (vbox, overflow:scroll, flex=1, width=100%, bg="#221d2e")
//           10 agent rows (text nodes)
//     pane2 (vbox, overflow:hidden, flex=1)
//       centerPanel (vbox, bg="#1a1626", width=100%, height=100%)
//         scrollView2 (vbox, overflow:scroll, flex=1, width=100%, bg="#1a1626")
//           normal messages + one 64K text message
func TestScrollbarLeakToLeftPanel(t *testing.T) {
	bufW, bufH := 80, 40

	// Left panel: agent tree with wide char icons
	leftRows := make([]*Node, 10)
	for i := range leftRows {
		// Use wide character "▶" at the start (like real agent tree icons)
		leftRows[i] = &Node{
			Type:    "text",
			Content: "▶ Agent " + string(rune('A'+i)) + " · running",
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

	pane1 := &Node{Type: "vbox", Style: Style{Overflow: "hidden", Width: 32, HeightPercent: 100}}
	pane1.Children = []*Node{leftSidebarComp}

	// Right panel: chat with a 64K text message
	// Create a very long text (simulating 64K content)
	bigText := strings.Repeat("This is a very long line of text that should cause scrolling. ", 1000)
	rightRows := make([]*Node, 20)
	for i := range rightRows {
		if i == 10 {
			// The 64K message
			rightRows[i] = &Node{
				Type:    "text",
				Content: bigText,
				Style:   Style{Foreground: "#ffffff", WidthPercent: 100},
			}
		} else {
			rightRows[i] = &Node{
				Type:    "text",
				Content: "Normal chat message " + string(rune('0'+i)),
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

	// Root hbox
	root := &Node{Type: "hbox", Style: Style{WidthPercent: 100, HeightPercent: 100}, Children: []*Node{pane1, pane2}}

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
	t.Logf("pane2: X=%d W=%d H=%d", pane2.X, pane2.W, pane2.H)
	t.Logf("scrollView1: X=%d W=%d H=%d ScrollHeight=%d", scrollView1.X, scrollView1.W, scrollView1.H, scrollView1.ScrollHeight)
	t.Logf("scrollView2: X=%d W=%d H=%d ScrollHeight=%d", scrollView2.X, scrollView2.W, scrollView2.H, scrollView2.ScrollHeight)

	// Initial paint
	buf := NewCellBuffer(bufW, bufH)
	root.PaintDirty = true
	PaintDirty(buf, root)

	// Record initial state of X=0 column
	leftPanelBG := "#120d1a"   // leftSidebar bg
	scrollView1BG := "#221d2e" // left scrollView bg
	
	// Verify initial state is clean
	checkLeftPanel := func(label string) {
		for y := 0; y < bufH; y++ {
			c := buf.Get(0, y)
			// X=0 should only have left panel colors
			if c.BG != "" && c.BG != leftPanelBG && c.BG != scrollView1BG {
				t.Errorf("[%s] X=0 Y=%d has unexpected BG=%q Ch=%c (expected %q or %q)",
					label, y, c.BG, c.Ch, leftPanelBG, scrollView1BG)
			}
		}
		// Also check X=0..31 for right panel colors leaking in
		rightPanelBG := "#1a1626"
		scrollbarTrack := "#313244"
		scrollbarThumb := "#cdd6f4"
		for x := 0; x < 32; x++ {
			for y := 0; y < bufH; y++ {
				c := buf.Get(x, y)
				if c.BG == scrollbarTrack || c.BG == scrollbarThumb || c.FG == scrollbarThumb {
					t.Errorf("[%s] X=%d Y=%d has scrollbar color BG=%q FG=%q Ch=%c",
						label, x, y, c.BG, c.FG, c.Ch)
					return
				}
				_ = rightPanelBG
			}
		}
	}

	checkLeftPanel("initial")

	// Now scroll the right panel's ScrollView to show the big text
	// Simulate scrolling down significantly
	maxScroll := computeMaxScrollY(scrollView2)
	t.Logf("Right scroll maxScrollY=%d", maxScroll)
	
	if maxScroll <= 0 {
		t.Log("WARNING: maxScroll <= 0, big text may not be long enough to trigger scrolling")
	}

	// Scroll to middle (where the big text is)
	scrollView2.ScrollY = maxScroll / 2
	scrollView2.PaintDirty = true
	PaintDirty(buf, root)
	checkLeftPanel("after scroll to middle")

	// Scroll to near the big text
	scrollView2.ScrollY = maxScroll / 3
	scrollView2.PaintDirty = true
	PaintDirty(buf, root)
	checkLeftPanel("after scroll to 1/3")

	// Scroll to max
	scrollView2.ScrollY = maxScroll
	scrollView2.PaintDirty = true
	PaintDirty(buf, root)
	checkLeftPanel("after scroll to max")

	// Also test: mark pane2 dirty simultaneously (race condition)
	scrollView2.ScrollY = maxScroll / 2
	scrollView2.PaintDirty = true
	pane2.PaintDirty = true
	PaintDirty(buf, root)
	checkLeftPanel("after scroll + pane2 dirty")

	// Test: mark root dirty (full repaint)
	scrollView2.ScrollY = maxScroll
	root.PaintDirty = true
	PaintDirty(buf, root)
	checkLeftPanel("after root dirty with scroll at max")

	// Test: incremental layout + paint (simulating component re-render)
	scrollView2.ScrollY = maxScroll / 2
	scrollView2.LayoutDirty = true
	scrollView2.PaintDirty = true
	LayoutIncremental(root)
	PaintDirty(buf, root)
	checkLeftPanel("after incremental layout + scroll")
}
