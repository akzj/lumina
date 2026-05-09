package render

import (
	"testing"
)

// TestWidthPercentInOverflowHidden verifies that width:100% resolves
// relative to the overflow:hidden parent, not the root.
func TestWidthPercentInOverflowHidden(t *testing.T) {
	// Structure: root > pane1(overflow:hidden, W=32) > component > child(width:100%)
	child := &Node{Type: "vbox", Style: Style{WidthPercent: 100, HeightPercent: 100, Background: "#120d1a"}}
	comp := &Node{Type: "component", Children: []*Node{child}}
	pane1 := &Node{Type: "vbox", Style: Style{Width: 32, HeightPercent: 100, Overflow: "hidden"}, Children: []*Node{comp}}
	pane2 := &Node{Type: "vbox", Style: Style{Flex: 1, HeightPercent: 100, Overflow: "hidden"}}
	root := &Node{Type: "hbox", Style: Style{WidthPercent: 100, HeightPercent: 100}, Children: []*Node{pane1, pane2}}

	var setParents func(n *Node)
	setParents = func(n *Node) {
		for _, c := range n.Children {
			c.Parent = n
			setParents(c)
		}
	}
	setParents(root)

	layoutViewportW = 214
	layoutViewportH = 55
	LayoutFull(root, 0, 0, 214, 55)

	t.Logf("root: X=%d W=%d", root.X, root.W)
	t.Logf("pane1: X=%d W=%d", pane1.X, pane1.W)
	t.Logf("comp: X=%d W=%d", comp.X, comp.W)
	t.Logf("child: X=%d W=%d", child.X, child.W)

	if child.W != 32 {
		t.Errorf("child.W = %d, want 32 (should be constrained by pane1's width)", child.W)
	}
	if child.X != 0 {
		t.Errorf("child.X = %d, want 0", child.X)
	}
}
