package v2

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/output"
	"github.com/akzj/lumina/pkg/render"
)

func TestHoverHitTestAccuracy(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 80, 24, ta)

	err := app.RunScript("../examples/breathing_light.lua")
	if err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	app.RenderAll()
	eng := app.Engine()

	// Find hover nodes
	var hoverNodes []*render.Node
	var walkNodes func(n *render.Node)
	walkNodes = func(n *render.Node) {
		if n == nil {
			return
		}
		if n.HoverStyle != nil {
			hoverNodes = append(hoverNodes, n)
		}
		for _, c := range n.Children {
			walkNodes(c)
		}
	}
	walkNodes(eng.Root().RootNode)

	// For each hover node, check what hitTest returns at its center
	for _, node := range hoverNodes {
		midX := node.X + node.W/2
		midY := node.Y
		hit := eng.HitTestScreen(midX, midY)
		if hit == nil {
			t.Errorf("HitTest at (%d,%d) for %q returned nil!", midX, midY, node.Content)
			continue
		}
		if hit != node {
			t.Errorf("HitTest at (%d,%d) for %q returned DIFFERENT node: type=%s content=%q hasHoverStyle=%v",
				midX, midY, node.Content, hit.Type, hit.Content, hit.HoverStyle != nil)
			// Walk up to see if our node is an ancestor
			for p := hit.Parent; p != nil; p = p.Parent {
				if p == node {
					t.Logf("  → Our node IS an ancestor of hit (hit is deeper)")
					break
				}
			}
		} else {
			t.Logf("✅ HitTest at (%d,%d) correctly returns %q", midX, midY, node.Content)
		}
	}

	// Also test: what happens between the text nodes (in the gap)?
	if len(hoverNodes) >= 2 {
		gapX := hoverNodes[0].X + hoverNodes[0].W + 1 // in the gap
		gapY := hoverNodes[0].Y
		hit := eng.HitTestScreen(gapX, gapY)
		if hit != nil {
			t.Logf("Gap at (%d,%d): hit type=%s content=%q hasHoverStyle=%v", gapX, gapY, hit.Type, hit.Content, hit.HoverStyle != nil)
		} else {
			t.Logf("Gap at (%d,%d): hit=nil", gapX, gapY)
		}
	}
}
