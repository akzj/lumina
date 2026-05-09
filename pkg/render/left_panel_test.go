package render

import (
    "testing"
)

// TestLeftPanelX0Coverage tests that X=0 column is properly covered
// when rendering through: hbox > pane1(overflow:hidden, no bg) > leftSidebar(vbox, bg) > scrollview(overflow:scroll) > rows
func TestLeftPanelX0Coverage(t *testing.T) {
    bufW, bufH := 80, 24
    
    // Build tree matching actual structure
    // Row items inside scroll
    rows := make([]*Node, 10)
    for i := range rows {
        rows[i] = &Node{Type: "text", Content: "Agent " + string(rune('A'+i)), Style: Style{Foreground: "#ffffff", WidthPercent: 100}}
    }
    
    scrollView := &Node{Type: "vbox", Style: Style{Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#221d2e"}, Children: rows}
    scrollComp := &Node{Type: "component", Children: []*Node{scrollView}}
    
    agentTreeWrap := &Node{Type: "vbox", Style: Style{Flex: 2, WidthPercent: 100, Overflow: "hidden"}, Children: []*Node{scrollComp}}
    
    leftSidebar := &Node{Type: "vbox", Style: Style{WidthPercent: 100, HeightPercent: 100, Background: "#120d1a"}, Children: []*Node{agentTreeWrap}}
    leftSidebarComp := &Node{Type: "component", Children: []*Node{leftSidebar}}
    
    // pane1: overflow:hidden, NO background
    pane1 := &Node{Type: "vbox", Style: Style{Overflow: "hidden", Width: 32, HeightPercent: 100}}
    pane1.Children = []*Node{leftSidebarComp}
    
    // pane2: overflow:hidden, flex
    pane2 := &Node{Type: "vbox", Style: Style{Overflow: "hidden", Flex: 1}, Children: []*Node{
        &Node{Type: "vbox", Style: Style{WidthPercent: 100, HeightPercent: 100, Background: "#1a1626"}},
    }}
    
    // root hbox
    root := &Node{Type: "hbox", Style: Style{WidthPercent: 100, HeightPercent: 100, Background: "#0a0a0a"}, Children: []*Node{pane1, pane2}}
    
    // Set parent pointers
    var setParents func(n *Node)
    setParents = func(n *Node) { for _, c := range n.Children { c.Parent = n; setParents(c) } }
    setParents(root)
    
    // Layout
    layoutViewportW = bufW
    layoutViewportH = bufH
    LayoutFull(root, 0, 0, bufW, bufH)
    
    t.Logf("pane1: X=%d W=%d H=%d", pane1.X, pane1.W, pane1.H)
    t.Logf("leftSidebar: X=%d W=%d H=%d bg=%q", leftSidebar.X, leftSidebar.W, leftSidebar.H, leftSidebar.Style.Background)
    t.Logf("agentTreeWrap: X=%d W=%d H=%d", agentTreeWrap.X, agentTreeWrap.W, agentTreeWrap.H)
    t.Logf("scrollView: X=%d W=%d H=%d", scrollView.X, scrollView.W, scrollView.H)
    
    // Paint
    buf := NewCellBuffer(bufW, bufH)
    // Simulate PaintDirty with root dirty
    buf.ClearRect(root.X, root.Y, root.W, root.H)
    paintNode(buf, root)
    
    // Check X=0 column: every cell should have been painted (Ch != 0 or BG != "")
    emptyCount := 0
    for y := 0; y < bufH; y++ {
        c := buf.Get(0, y)
        if c.Ch == 0 && c.BG == "" {
            if emptyCount < 5 {
                t.Errorf("X=0 Y=%d is EMPTY (Ch=0, BG=\"\") — not painted!", y)
            }
            emptyCount++
        }
    }
    if emptyCount > 0 {
        t.Errorf("Total empty cells at X=0: %d / %d", emptyCount, bufH)
    } else {
        t.Log("PASS: X=0 column fully covered")
    }
    
    // Also check a few other X positions in pane1
    for x := 0; x < 32; x++ {
        empty := 0
        for y := 0; y < bufH; y++ {
            c := buf.Get(x, y)
            if c.Ch == 0 && c.BG == "" {
                empty++
            }
        }
        if empty > 0 {
            t.Errorf("X=%d has %d empty cells", x, empty)
        }
    }
}
