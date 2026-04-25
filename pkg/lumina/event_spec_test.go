package lumina

import (
	"testing"
)

func TestEventBubbles(t *testing.T) {
	// Verify bubble policy
	bubbling := []string{"click", "mousedown", "mouseup", "mousemove", "keydown", "keyup",
		"input", "change", "submit", "scroll", "wheel", "contextmenu"}
	nonBubbling := []string{"mouseenter", "mouseleave", "focus", "blur", "resize"}

	for _, et := range bubbling {
		if !eventBubbles(et) {
			t.Errorf("%s should bubble", et)
		}
	}
	for _, et := range nonBubbling {
		if eventBubbles(et) {
			t.Errorf("%s should NOT bubble", et)
		}
	}
}

func TestCurrentTarget(t *testing.T) {
	app := NewAppWithSize(20, 5)
	_ = NewMockTermIO(20, 5)

	// Build a VNode tree: parent → child
	parent := NewVNode("box")
	parent.Props["id"] = "parent-box"
	child := NewVNode("box")
	child.Props["id"] = "child-box"
	parent.AddChild(child)

	// Build VNode tree for event bubbling
	tree := BuildVNodeTree(parent)
	globalEventBus.SetVNodeTree(tree)

	// Register handler on parent for "click"
	var capturedCurrentTarget string
	globalEventBus.On("click", "parent-box", func(e *Event) {
		capturedCurrentTarget = e.CurrentTarget
	})

	// Emit click with Target = child
	globalEventBus.Emit(&Event{
		Type:    "click",
		Target:  "child-box",
		Bubbles: true,
	})

	// During parent's handler, CurrentTarget should be "parent-box"
	if capturedCurrentTarget != "parent-box" {
		t.Errorf("Expected CurrentTarget='parent-box', got %q", capturedCurrentTarget)
	}

	_ = app // keep app alive for cleanup
}

func TestNonBubblingEventStopsAtTarget(t *testing.T) {
	app := NewAppWithSize(20, 5)
	_ = NewMockTermIO(20, 5)

	// Build a VNode tree: parent → child
	parent := NewVNode("box")
	parent.Props["id"] = "parent-box"
	child := NewVNode("box")
	child.Props["id"] = "child-box"
	parent.AddChild(child)

	tree := BuildVNodeTree(parent)
	globalEventBus.SetVNodeTree(tree)

	// Register "mouseenter" handler on BOTH parent and child
	childFired := false
	parentFired := false

	globalEventBus.On("mouseenter", "child-box", func(e *Event) {
		childFired = true
	})
	globalEventBus.On("mouseenter", "parent-box", func(e *Event) {
		parentFired = true
	})

	// Emit mouseenter with Target = child (non-bubbling)
	globalEventBus.Emit(&Event{
		Type:    "mouseenter",
		Target:  "child-box",
		Bubbles: false,
	})

	// Child's handler should fire (target phase)
	if !childFired {
		t.Error("Expected child's mouseenter handler to fire")
	}

	// Parent's handler should NOT fire (no bubble for mouseenter)
	if parentFired {
		t.Error("Parent's mouseenter handler should NOT fire (non-bubbling event)")
	}

	_ = app
}
