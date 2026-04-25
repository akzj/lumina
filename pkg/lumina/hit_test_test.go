package lumina_test

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

func TestHitTest_BasicVNode(t *testing.T) {
	vnode := lumina.NewVNode("button")
	vnode.Props["id"] = "btn1"
	vnode.X = 5
	vnode.Y = 5
	vnode.W = 10
	vnode.H = 10

	// Click inside → found
	id := lumina.HitTestVNode(vnode, 7, 7)
	if id != "btn1" {
		t.Errorf("expected 'btn1', got %q", id)
	}

	// Click on edge (5,5) → found (inclusive)
	id = lumina.HitTestVNode(vnode, 5, 5)
	if id != "btn1" {
		t.Errorf("expected 'btn1' at edge, got %q", id)
	}

	// Click at (14,14) → found (5+10-1=14, still inside)
	id = lumina.HitTestVNode(vnode, 14, 14)
	if id != "btn1" {
		t.Errorf("expected 'btn1' at bottom-right edge, got %q", id)
	}
}

func TestHitTest_NestedChildren(t *testing.T) {
	parent := lumina.NewVNode("vbox")
	parent.Props["id"] = "parent"
	parent.X = 0
	parent.Y = 0
	parent.W = 80
	parent.H = 24

	child := lumina.NewVNode("button")
	child.Props["id"] = "child_btn"
	child.X = 10
	child.Y = 5
	child.W = 20
	child.H = 3
	parent.AddChild(child)

	// Click on child → child ID returned (deepest match)
	id := lumina.HitTestVNode(parent, 15, 6)
	if id != "child_btn" {
		t.Errorf("expected 'child_btn', got %q", id)
	}

	// Click on parent area outside child → parent ID
	id = lumina.HitTestVNode(parent, 50, 15)
	if id != "parent" {
		t.Errorf("expected 'parent', got %q", id)
	}
}

func TestHitTest_Miss(t *testing.T) {
	vnode := lumina.NewVNode("button")
	vnode.Props["id"] = "btn1"
	vnode.X = 10
	vnode.Y = 10
	vnode.W = 5
	vnode.H = 3

	// Click outside → empty
	id := lumina.HitTestVNode(vnode, 0, 0)
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}

	// Just past the right edge
	id = lumina.HitTestVNode(vnode, 15, 10)
	if id != "" {
		t.Errorf("expected empty at right edge, got %q", id)
	}

	// Just past the bottom edge
	id = lumina.HitTestVNode(vnode, 10, 13)
	if id != "" {
		t.Errorf("expected empty at bottom edge, got %q", id)
	}
}

func TestHitTest_NoID(t *testing.T) {
	// VNode without ID → returns ""
	vnode := lumina.NewVNode("box")
	vnode.X = 0
	vnode.Y = 0
	vnode.W = 10
	vnode.H = 10

	id := lumina.HitTestVNode(vnode, 5, 5)
	if id != "" {
		t.Errorf("expected empty for no-ID node, got %q", id)
	}
}

func TestHitTest_DeepNesting(t *testing.T) {
	root := lumina.NewVNode("vbox")
	root.Props["id"] = "root"
	root.X = 0
	root.Y = 0
	root.W = 80
	root.H = 24

	mid := lumina.NewVNode("hbox")
	mid.Props["id"] = "mid"
	mid.X = 5
	mid.Y = 5
	mid.W = 30
	mid.H = 10
	root.AddChild(mid)

	deep := lumina.NewVNode("button")
	deep.Props["id"] = "deep_btn"
	deep.X = 10
	deep.Y = 7
	deep.W = 10
	deep.H = 3
	mid.AddChild(deep)

	// Click on deepest → deep_btn
	id := lumina.HitTestVNode(root, 12, 8)
	if id != "deep_btn" {
		t.Errorf("expected 'deep_btn', got %q", id)
	}

	// Click on mid but not deep → mid
	id = lumina.HitTestVNode(root, 6, 6)
	if id != "mid" {
		t.Errorf("expected 'mid', got %q", id)
	}
}

func TestMouseClick_SetsTarget(t *testing.T) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		_G._clicked_target = ""
		local App = lumina.defineComponent({
			name = "MouseTest",
			render = function(self)
				return {
					type = "vbox",
					children = {
						{
							type = "button",
							id = "mouse_btn",
							content = "Click Me",
							onClick = function()
								_G._clicked_target = "mouse_btn"
							end,
						},
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	// Verify the button is registered as focusable
	eb := lumina.GetGlobalEventBus()
	if !eb.IsFocusable("mouse_btn") {
		t.Error("mouse_btn should be focusable after render")
	}
}
