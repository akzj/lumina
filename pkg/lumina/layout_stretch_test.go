package lumina

import (
	"testing"
)

// TestLayoutVBox_ContainerStretchesWithoutFlex verifies that a container child
// (vbox/hbox) without explicit height or flex stretches to fill available space
// in a vbox parent. This is the fix for the dashboard right-side content bug
// where the HomePage vbox got H=1 instead of filling the parent.
func TestLayoutVBox_ContainerStretchesWithoutFlex(t *testing.T) {
	// Parent vbox with flex=1 (gets full height from its parent)
	parent := NewVNode("vbox")
	parent.Props["style"] = map[string]any{"flex": 1}

	// Child vbox with padding=1 but NO height and NO flex
	child := NewVNode("vbox")
	child.Props["style"] = map[string]any{"padding": 1}

	text := NewVNode("text")
	text.Content = "Hello World"
	child.AddChild(text)

	parent.AddChild(child)

	// Root hbox
	root := NewVNode("hbox")
	root.AddChild(parent)

	frame := VNodeToFrame(root, 80, 24)
	_ = frame

	// The child vbox should stretch to fill the parent's content area
	if child.H != 24 {
		t.Errorf("child container should stretch to H=24, got H=%d", child.H)
	}
}

// TestLayoutVBox_TextGetsMinimumHeight verifies that text nodes still get
// minimum height of 1 (not implicit flex) in a vbox.
func TestLayoutVBox_TextGetsMinimumHeight(t *testing.T) {
	parent := NewVNode("vbox")

	text1 := NewVNode("text")
	text1.Content = "Line 1"
	text2 := NewVNode("text")
	text2.Content = "Line 2"

	parent.AddChild(text1)
	parent.AddChild(text2)

	frame := VNodeToFrame(parent, 80, 24)
	_ = frame

	// Text nodes should get H=1, not stretch
	if text1.H != 1 {
		t.Errorf("text1 should get H=1, got H=%d", text1.H)
	}
	if text2.H != 1 {
		t.Errorf("text2 should get H=1, got H=%d", text2.H)
	}
}

// TestLayoutHBox_ContainerStretchesWithoutFlex verifies that a container child
// without explicit width or flex stretches to fill available space in an hbox.
func TestLayoutHBox_ContainerStretchesWithoutFlex(t *testing.T) {
	// Sidebar with fixed width
	sidebar := NewVNode("vbox")
	sidebar.Props["style"] = map[string]any{"width": 20}

	// Content area without width or flex — should stretch
	content := NewVNode("vbox")
	text := NewVNode("text")
	text.Content = "Content"
	content.AddChild(text)

	root := NewVNode("hbox")
	root.AddChild(sidebar)
	root.AddChild(content)

	frame := VNodeToFrame(root, 120, 40)
	_ = frame

	if sidebar.W != 20 {
		t.Errorf("sidebar should get W=20, got W=%d", sidebar.W)
	}
	// Content should get remaining width (120 - 20 = 100)
	if content.W != 100 {
		t.Errorf("content container should stretch to W=100, got W=%d", content.W)
	}
}

// TestLayoutVBox_MixedFlexAndContainer verifies that explicit flex children
// take priority over implicit container stretch.
func TestLayoutVBox_MixedFlexAndContainer(t *testing.T) {
	header := NewVNode("text")
	header.Content = "Header"

	// Explicit flex child
	body := NewVNode("vbox")
	body.Props["style"] = map[string]any{"flex": 2}
	bodyText := NewVNode("text")
	bodyText.Content = "Body"
	body.AddChild(bodyText)

	// Container without flex — gets implicit flex=1
	footer := NewVNode("vbox")
	footerText := NewVNode("text")
	footerText.Content = "Footer"
	footer.AddChild(footerText)

	root := NewVNode("vbox")
	root.AddChild(header)
	root.AddChild(body)
	root.AddChild(footer)

	frame := VNodeToFrame(root, 80, 30)
	_ = frame

	// Header: H=1 (text)
	if header.H != 1 {
		t.Errorf("header should get H=1, got H=%d", header.H)
	}

	// Remaining = 30 - 1 = 29, split as flex 2:1
	// body = 29 * 2/3 = 19
	// footer = 29 * 1/3 = 9
	if body.H < 15 {
		t.Errorf("body (flex=2) should get most space, got H=%d", body.H)
	}
	if footer.H < 5 {
		t.Errorf("footer (implicit flex=1) should get some space, got H=%d", footer.H)
	}
}

// TestLayoutDashboard_RightSideContent is the regression test for the dashboard bug.
// Verifies that in an hbox with sidebar(width=20) + content(flex=1),
// the content area renders its children correctly.
func TestLayoutDashboard_RightSideContent(t *testing.T) {
	sidebar := NewVNode("vbox")
	sidebar.Props["style"] = map[string]any{
		"width":      20,
		"border":     "single",
		"background": "#181825",
	}
	sideText := NewVNode("text")
	sideText.Content = "Home"
	sidebar.AddChild(sideText)

	contentText := NewVNode("text")
	contentText.Content = "Dashboard Content Here"

	innerVBox := NewVNode("vbox")
	innerVBox.Props["style"] = map[string]any{"padding": 1}
	innerVBox.AddChild(contentText)

	contentVBox := NewVNode("vbox")
	contentVBox.Props["style"] = map[string]any{"flex": 1}
	contentVBox.AddChild(innerVBox)

	root := NewVNode("hbox")
	root.AddChild(sidebar)
	root.AddChild(contentVBox)

	frame := VNodeToFrame(root, 120, 40)

	// Sidebar should be 20 wide
	if sidebar.W != 20 {
		t.Errorf("sidebar width: got %d, want 20", sidebar.W)
	}

	// Content vbox should get remaining width
	if contentVBox.W != 100 {
		t.Errorf("contentVBox width: got %d, want 100", contentVBox.W)
	}

	// Inner vbox should stretch to fill content area (not H=1!)
	if innerVBox.H < 30 {
		t.Errorf("innerVBox should stretch to fill parent, got H=%d (want >=30)", innerVBox.H)
	}

	// Content text should be visible in the frame at x >= 21 (past sidebar + padding)
	hasContent := false
	for x := 21; x < 120; x++ {
		cell := frame.Cells[1][x] // row 1 (after padding top)
		if cell.Char != ' ' && cell.Char != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Dashboard content text not rendered in right-side area")
	}
}
