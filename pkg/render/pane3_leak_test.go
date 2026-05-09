package render

import (
	"testing"
)

// TestPane3ContentLeakToPane1 tests whether right panel (pane3) content
// can leak to left panel (pane1) area when paintDirtyWalk processes
// a dirty node inside pane3.
func TestPane3ContentLeakToPane1(t *testing.T) {
	bufW, bufH := 214, 55

	// === PANE 1 (left sidebar) ===
	leftContent := &Node{Type: "text", Content: "Left sidebar", Style: Style{
		Foreground: "#ffffff", Height: 1, WidthPercent: 100,
	}}
	leftVbox := &Node{Type: "vbox", Style: Style{
		WidthPercent: 100, HeightPercent: 100, Background: "#120d1a",
	}, Children: []*Node{leftContent}}
	leftComp := &Node{Type: "component", Children: []*Node{leftVbox}}
	pane1 := &Node{Type: "vbox", ID: "split-pane-1", Style: Style{
		Width: 32, HeightPercent: 100, Overflow: "hidden",
	}, Children: []*Node{leftComp}}

	// === DIVIDER 1 ===
	divider1 := &Node{Type: "vbox", ID: "split-div-1", Style: Style{
		Width: 1, Background: "#38605c",
	}}

	// === PANE 2 (center) ===
	centerContent := &Node{Type: "text", Content: "Center content", Style: Style{
		Foreground: "#ffffff", Height: 1, WidthPercent: 100,
	}}
	centerVbox := &Node{Type: "vbox", Style: Style{
		WidthPercent: 100, HeightPercent: 100, Background: "#1a1626",
	}, Children: []*Node{centerContent}}
	centerComp := &Node{Type: "component", Children: []*Node{centerVbox}}
	pane2 := &Node{Type: "vbox", ID: "split-pane-2", Style: Style{
		Flex: 1, HeightPercent: 100, Overflow: "hidden",
	}, Children: []*Node{centerComp}}

	// === DIVIDER 2 ===
	divider2 := &Node{Type: "vbox", ID: "split-div-2", Style: Style{
		Width: 1, Background: "#38605c",
	}}

	// === PANE 3 (right panel - Config) ===
	// Simulating ConfigPanel: vbox with padding=1, text items
	configItems := []*Node{
		{Type: "text", Content: "LLM Configuration", Style: Style{
			Foreground: "#cdd6f4", Height: 1,
		}},
		{Type: "text", Content: "Provider      : anthropic", Style: Style{
			Foreground: "#cdd6f4", Height: 1,
		}},
		{Type: "text", Content: "Model         : claude-3", Style: Style{
			Foreground: "#cdd6f4", Height: 1,
		}},
		{Type: "text", Content: "Temperature   : 0.7", Style: Style{
			Foreground: "#cdd6f4", Height: 1,
		}},
	}
	configVbox := &Node{Type: "vbox", Style: Style{
		WidthPercent: 100, PaddingLeft: 1, PaddingRight: 1, PaddingTop: 1, PaddingBottom: 1,
	}, Children: configItems}
	configComp := &Node{Type: "component", Children: []*Node{configVbox}}

	// Right tab bar
	rightTabBar := &Node{Type: "hbox", Style: Style{
		Height: 1, WidthPercent: 100,
	}, Children: []*Node{
		{Type: "text", Content: "[ Config ]", Style: Style{Height: 1, Foreground: "#cdd6f4"}},
	}}

	rightPanel := &Node{Type: "vbox", Style: Style{
		WidthPercent: 100, HeightPercent: 100,
	}, Children: []*Node{rightTabBar, configComp}}
	rightPanelComp := &Node{Type: "component", Children: []*Node{rightPanel}}

	pane3 := &Node{Type: "vbox", ID: "split-pane-3", Style: Style{
		Width: 30, HeightPercent: 100, Overflow: "hidden",
	}, Children: []*Node{rightPanelComp}}

	// === ROOT (SplitPane hbox) ===
	root := &Node{Type: "hbox", Style: Style{Flex: 1, WidthPercent: 100, HeightPercent: 100},
		Children: []*Node{pane1, divider1, pane2, divider2, pane3}}

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

	t.Logf("pane1: X=%d W=%d", pane1.X, pane1.W)
	t.Logf("pane2: X=%d W=%d", pane2.X, pane2.W)
	t.Logf("pane3: X=%d W=%d", pane3.X, pane3.W)
	t.Logf("rightPanel: X=%d W=%d", rightPanel.X, rightPanel.W)
	t.Logf("configVbox: X=%d W=%d", configVbox.X, configVbox.W)
	t.Logf("configItems[3] (Temperature): X=%d W=%d", configItems[3].X, configItems[3].W)

	// Check: all nodes in pane3 should have X >= pane3.X
	pane3X := pane3.X
	var checkCoords func(n *Node, path string)
	checkCoords = func(n *Node, path string) {
		if n.W > 0 && n.H > 0 && n.X < pane3X {
			t.Errorf("LAYOUT BUG: %s type=%q X=%d (< %d) W=%d", path, n.Type, n.X, pane3X, n.W)
		}
		for i, c := range n.Children {
			checkCoords(c, path+"/"+string(rune('0'+i)))
		}
	}
	checkCoords(pane3, "pane3")

	// Initial full paint
	buf := NewCellBuffer(bufW, bufH)
	root.PaintDirty = true
	PaintDirty(buf, root)

	// Check X=0..31 for pane3 content
	checkLeak := func(label string) {
		for x := 0; x < 32; x++ {
			for y := 0; y < bufH; y++ {
				c := buf.Get(x, y)
				// "Temperature" text should NOT appear in left panel
				if c.Ch == 'T' && x < 5 {
					// Check if it's "Temperature"
					word := ""
					for dx := 0; dx < 11 && x+dx < bufW; dx++ {
						cc := buf.Get(x+dx, y)
						if cc.Ch == 0 {
							break
						}
						word += string(cc.Ch)
					}
					if len(word) >= 4 && word[:4] == "Temp" {
						t.Errorf("[%s] LEAK: 'Temperature' found at X=%d Y=%d", label, x, y)
						return
					}
				}
			}
		}
	}

	checkLeak("initial")

	// Now simulate dirty paint scenarios:
	// 1. Only configItems[3] dirty (Temperature text re-renders)
	configItems[3].PaintDirty = true
	PaintDirty(buf, root)
	checkLeak("temperature_dirty")

	// 2. configComp dirty
	configComp.PaintDirty = true
	PaintDirty(buf, root)
	checkLeak("configComp_dirty")

	// 3. rightPanel dirty
	rightPanel.PaintDirty = true
	PaintDirty(buf, root)
	checkLeak("rightPanel_dirty")

	// 4. rightPanelComp dirty
	rightPanelComp.PaintDirty = true
	PaintDirty(buf, root)
	checkLeak("rightPanelComp_dirty")

	// 5. pane3 dirty
	pane3.PaintDirty = true
	PaintDirty(buf, root)
	checkLeak("pane3_dirty")

	// 6. Multiple dirty simultaneously
	configItems[3].PaintDirty = true
	pane3.PaintDirty = true
	PaintDirty(buf, root)
	checkLeak("temp+pane3_dirty")

	// 7. Root dirty
	root.PaintDirty = true
	PaintDirty(buf, root)
	checkLeak("root_dirty")
}
