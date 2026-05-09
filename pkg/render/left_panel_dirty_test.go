package render

import (
    "testing"
)

// TestLeftPanelDirtyRepaint tests that X=0 stays covered after dirty repaints
func TestLeftPanelDirtyRepaint(t *testing.T) {
    bufW, bufH := 214, 53  // Match user's actual terminal size
    
    rows := make([]*Node, 20)
    for i := range rows {
        rows[i] = &Node{Type: "text", Content: "Agent row content", Style: Style{Foreground: "#ffffff", WidthPercent: 100}}
    }
    
    scrollView := &Node{Type: "vbox", ID: "scrollview", Style: Style{Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#221d2e"}, Children: rows}
    scrollComp := &Node{Type: "component", Children: []*Node{scrollView}}
    agentTreeWrap := &Node{Type: "vbox", Style: Style{Flex: 2, WidthPercent: 100, Overflow: "hidden"}, Children: []*Node{scrollComp}}
    leftSidebar := &Node{Type: "vbox", Style: Style{WidthPercent: 100, HeightPercent: 100, Background: "#120d1a"}, Children: []*Node{agentTreeWrap}}
    leftSidebarComp := &Node{Type: "component", Children: []*Node{leftSidebar}}
    
    pane1 := &Node{Type: "vbox", Style: Style{Overflow: "hidden", Width: 32, HeightPercent: 100}}
    pane1.Children = []*Node{leftSidebarComp}
    
    // Right panel with its own scroll
    rightRows := make([]*Node, 100)
    for i := range rightRows {
        rightRows[i] = &Node{Type: "text", Content: "Chat message content that is quite long", Style: Style{Foreground: "#ffffff"}}
    }
    rightScroll := &Node{Type: "vbox", ID: "chat-scroll", Style: Style{Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#1a1626", ScrollAnchor: "top"}, Children: rightRows}
    rightScrollComp := &Node{Type: "component", Children: []*Node{rightScroll}}
    centerPanel := &Node{Type: "vbox", Style: Style{WidthPercent: 100, HeightPercent: 100, Background: "#1a1626"}, Children: []*Node{rightScrollComp}}
    centerPanelComp := &Node{Type: "component", Children: []*Node{centerPanel}}
    
    pane2 := &Node{Type: "vbox", Style: Style{Overflow: "hidden", Flex: 1}, Children: []*Node{centerPanelComp}}
    
    root := &Node{Type: "hbox", Style: Style{WidthPercent: 100, HeightPercent: 100}, Children: []*Node{pane1, pane2}}
    
    var setParents func(n *Node)
    setParents = func(n *Node) { for _, c := range n.Children { c.Parent = n; setParents(c) } }
    setParents(root)
    
    layoutViewportW = bufW
    layoutViewportH = bufH
    LayoutFull(root, 0, 0, bufW, bufH)
    
    // Initial full paint
    buf := NewCellBuffer(bufW, bufH)
    root.PaintDirty = true
    PaintDirty(buf, root)
    
    // Verify initial state
    checkX0 := func(label string) int {
        empty := 0
        for y := 0; y < bufH; y++ {
            c := buf.Get(0, y)
            if c.Ch == 0 && c.BG == "" {
                empty++
            }
        }
        if empty > 0 {
            t.Errorf("[%s] X=0 has %d empty cells out of %d", label, empty, bufH)
        }
        return empty
    }
    
    checkX0("initial")
    
    // Simulate: right panel scroll (mark right scroll dirty)
    rightScroll.ScrollY = 50
    rightScroll.PaintDirty = true
    PaintDirty(buf, root)
    checkX0("after right scroll")
    
    // Simulate: a row in left panel changes (mark one row dirty)
    rows[3].PaintDirty = true
    PaintDirty(buf, root)
    checkX0("after left row dirty")
    
    // Simulate: root gets dirty (e.g., from layout change)
    root.PaintDirty = true
    PaintDirty(buf, root)
    checkX0("after root dirty")
    
    // Simulate: scrollView in left panel gets dirty
    scrollView.PaintDirty = true
    PaintDirty(buf, root)
    checkX0("after left scrollView dirty")
    
    // Simulate: agentTreeWrap gets dirty  
    agentTreeWrap.PaintDirty = true
    PaintDirty(buf, root)
    checkX0("after agentTreeWrap dirty")
    
    // Simulate: leftSidebar gets dirty
    leftSidebar.PaintDirty = true
    PaintDirty(buf, root)
    checkX0("after leftSidebar dirty")
    
    // Simulate: pane1 gets dirty
    pane1.PaintDirty = true
    PaintDirty(buf, root)
    checkX0("after pane1 dirty")
}
