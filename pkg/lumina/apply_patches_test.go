package lumina

import (
    "fmt"
    "testing"
)

func TestApplyPatchesWithStyles(t *testing.T) {
    width, height := 20, 5
    
    // Frame 1: hover at row 2, col 5 (with styles like mouse_test.lua)
    tree1 := &VNode{
        Type: "vbox",
        Props: map[string]interface{}{"style": map[string]interface{}{"background": "#1E1E2E"}},
        Children: []*VNode{
            {Type: "text", Content: "Status: (5,2)     ", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#89B4FA", "background": "#181825"}}},
            {Type: "text", Content: "....................", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
            // Row 2: hbox with hover at col 5
            {Type: "hbox", Children: []*VNode{
                {Type: "text", Content: ".....", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
                {Type: "text", Content: "X", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#A6E3A1", "background": "#313244"}}},
                {Type: "text", Content: "..............", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
            }},
            {Type: "text", Content: "....................", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
            {Type: "text", Content: "....................", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
        },
    }
    
    computeFlexLayout(tree1, 0, 0, width, height)
    frame1 := VNodeToFrame(tree1, width, height)
    
    // Frame 2: hover at row 3, col 10
    tree2 := &VNode{
        Type: "vbox",
        Props: map[string]interface{}{"style": map[string]interface{}{"background": "#1E1E2E"}},
        Children: []*VNode{
            {Type: "text", Content: "Status: (10,3)    ", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#89B4FA", "background": "#181825"}}},
            {Type: "text", Content: "....................", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
            {Type: "text", Content: "....................", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
            // Row 3: hbox with hover at col 10
            {Type: "hbox", Children: []*VNode{
                {Type: "text", Content: "..........", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
                {Type: "text", Content: "X", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#A6E3A1", "background": "#313244"}}},
                {Type: "text", Content: ".........", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
            }},
            {Type: "text", Content: "....................", Props: map[string]interface{}{"style": map[string]interface{}{"foreground": "#585B70", "background": "#1E1E2E"}}},
        },
    }
    
    patches := DiffVNode(tree1, tree2)
    fmt.Printf("Patches: %d\n", len(patches))
    for i, p := range patches {
        fmt.Printf("  [%d] %s path=%v\n", i, p.Type, p.Path)
        if p.OldNode != nil { fmt.Printf("       old: type=%s pos=(%d,%d,%d,%d)\n", p.OldNode.Type, p.OldNode.X, p.OldNode.Y, p.OldNode.W, p.OldNode.H) }
        if p.NewNode != nil { fmt.Printf("       new: type=%s pos=(%d,%d,%d,%d)\n", p.NewNode.Type, p.NewNode.X, p.NewNode.Y, p.NewNode.W, p.NewNode.H) }
    }
    
    computeFlexLayout(tree2, 0, 0, width, height)
    ApplyPatches(frame1, tree2, patches, width, height)
    
    frame2 := VNodeToFrame(tree2, width, height)
    
    // Compare
    diffs := 0
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            c1 := frame1.Cells[y][x]
            c2 := frame2.Cells[y][x]
            if c1.Char != c2.Char || c1.Foreground != c2.Foreground || c1.Background != c2.Background {
                fmt.Printf("  DIFF (%d,%d): inc char='%c' fg=%s bg=%s | full char='%c' fg=%s bg=%s\n", 
                    x, y, c1.Char, c1.Foreground, c1.Background, c2.Char, c2.Foreground, c2.Background)
                diffs++
            }
        }
    }
    
    if diffs == 0 {
        t.Log("PASS — frames identical with styles")
    } else {
        t.Errorf("FAIL — %d differences", diffs)
    }
}

// Test with rapid successive hover movements (3 frames)
func TestApplyPatchesMultiFrame(t *testing.T) {
    width, height := 30, 6
    
    makeTree := func(hoverX, hoverY int) *VNode {
        children := make([]*VNode, 0, height)
        status := fmt.Sprintf("Mouse: (%d,%d)", hoverX, hoverY)
        for len(status) < width { status += " " }
        children = append(children, &VNode{Type: "text", Content: status})
        
        for y := 1; y < height; y++ {
            if y == hoverY {
                // Build hbox with segments
                before := ""
                for i := 0; i < hoverX; i++ { before += "." }
                after := ""
                for i := hoverX+1; i < width; i++ { after += "." }
                children = append(children, &VNode{
                    Type: "hbox",
                    Children: []*VNode{
                        {Type: "text", Content: before},
                        {Type: "text", Content: "X"},
                        {Type: "text", Content: after},
                    },
                })
            } else {
                row := ""
                for i := 0; i < width; i++ { row += "." }
                children = append(children, &VNode{Type: "text", Content: row})
            }
        }
        return &VNode{Type: "vbox", Children: children}
    }
    
    // Frame 1
    tree1 := makeTree(5, 2)
    computeFlexLayout(tree1, 0, 0, width, height)
    frame := VNodeToFrame(tree1, width, height)
    
    // Simulate 10 rapid mouse movements
    positions := [][2]int{{6,2},{7,2},{8,3},{9,3},{10,4},{11,4},{12,3},{13,2},{14,2},{15,3}}
    
    for step, pos := range positions {
        prevTree := tree1
        tree1 = makeTree(pos[0], pos[1])
        
        patches := DiffVNode(prevTree, tree1)
        computeFlexLayout(tree1, 0, 0, width, height)
        ApplyPatches(frame, tree1, patches, width, height)
        
        // Compare with full render
        fullFrame := VNodeToFrame(tree1, width, height)
        diffs := 0
        for y := 0; y < height; y++ {
            for x := 0; x < width; x++ {
                if frame.Cells[y][x].Char != fullFrame.Cells[y][x].Char {
                    diffs++
                }
            }
        }
        if diffs > 0 {
            t.Errorf("Step %d: hover(%d,%d) — %d char differences", step, pos[0], pos[1], diffs)
            fmt.Println("  Incremental:")
            for y := 0; y < height; y++ {
                fmt.Print("  ")
                for x := 0; x < width; x++ {
                    ch := frame.Cells[y][x].Char
                    if ch == 0 { ch = ' ' }
                    fmt.Printf("%c", ch)
                }
                fmt.Println()
            }
            fmt.Println("  Full:")
            for y := 0; y < height; y++ {
                fmt.Print("  ")
                for x := 0; x < width; x++ {
                    ch := fullFrame.Cells[y][x].Char
                    if ch == 0 { ch = ' ' }
                    fmt.Printf("%c", ch)
                }
                fmt.Println()
            }
            return
        }
    }
    t.Log("PASS — 10 rapid movements, all frames identical")
}
