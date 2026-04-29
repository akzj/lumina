package widget

import (
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestWindowRenderBasic(t *testing.T) {
	state := Window.NewState()
	props := map[string]any{
		"title":  "Test Window",
		"x":      10,
		"y":      5,
		"width":  40,
		"height": 15,
	}
	result := Window.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Window.Render returned nil or non-*Node")
	}
	if node.Type != "vbox" {
		t.Errorf("expected root type 'vbox', got %q", node.Type)
	}
	if node.Style.Border != "rounded" {
		t.Errorf("expected border 'rounded', got %q", node.Style.Border)
	}
	if node.Style.Position != "absolute" {
		t.Errorf("expected position 'absolute', got %q", node.Style.Position)
	}
	if node.Style.Left != 10 {
		t.Errorf("expected left=10, got %d", node.Style.Left)
	}
	if node.Style.Top != 5 {
		t.Errorf("expected top=5, got %d", node.Style.Top)
	}
	if node.Style.Width != 40 {
		t.Errorf("expected width=40, got %d", node.Style.Width)
	}
	if node.Style.Height != 15 {
		t.Errorf("expected height=15, got %d", node.Style.Height)
	}

	// Should have: title hbox + divider + resize handle = 3 children (no Lua children)
	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children (title_hbox+divider+resize), got %d", len(node.Children))
	}

	// Title bar is an hbox with title text + close button
	titleBar := node.Children[0]
	if titleBar.Type != "hbox" {
		t.Errorf("expected title bar type 'hbox', got %q", titleBar.Type)
	}
	if len(titleBar.Children) != 2 {
		t.Fatalf("title bar should have 2 children (title+close), got %d", len(titleBar.Children))
	}
	titleTextNode := titleBar.Children[0]
	if !strings.Contains(titleTextNode.Content, "Test Window") {
		t.Errorf("title text should contain 'Test Window', got %q", titleTextNode.Content)
	}
	if !titleTextNode.Style.Bold {
		t.Error("title text should be bold")
	}
	closeNode := titleBar.Children[1]
	if !strings.Contains(closeNode.Content, "✕") {
		t.Errorf("close button should contain '✕', got %q", closeNode.Content)
	}

	// Resize handle should contain ◢
	resizeNode := node.Children[2]
	if !strings.Contains(resizeNode.Content, "◢") {
		t.Errorf("resize handle should contain '◢', got %q", resizeNode.Content)
	}
}

func TestWindowRenderWithChildren(t *testing.T) {
	state := Window.NewState()
	childNodes := []*render.Node{
		{Type: "text", Content: "Hello from child"},
	}
	props := map[string]any{
		"title":       "With Children",
		"x":           0,
		"y":           0,
		"width":       30,
		"height":      10,
		"_childNodes": childNodes,
	}
	result := Window.Render(props, state)
	node := result.(*render.Node)

	// Should have: title + divider + content_box + resize handle = 4 children
	if len(node.Children) != 4 {
		t.Fatalf("expected 4 children (title+divider+content_box+resize), got %d", len(node.Children))
	}
	// Children are wrapped in a content box (vbox with flex=1, overflow=hidden)
	contentBox := node.Children[2]
	if contentBox.Type != "vbox" {
		t.Fatalf("expected content box to be vbox, got %q", contentBox.Type)
	}
	if len(contentBox.Children) != 1 {
		t.Fatalf("expected 1 child in content box, got %d", len(contentBox.Children))
	}
	if contentBox.Children[0].Content != "Hello from child" {
		t.Errorf("child content: got %q, want 'Hello from child'", contentBox.Children[0].Content)
	}
}

func TestWindowDragMovesWindow(t *testing.T) {
	state := Window.NewState()
	props := map[string]any{
		"title":  "Draggable",
		"x":      10,
		"y":      5,
		"width":  40,
		"height": 15,
	}

	// Mousedown on title bar (relY=1 = title row inside border)
	evt := &render.WidgetEvent{
		Type:    "mousedown",
		X:       15,
		Y:       6, // widgetY=5, relY=1 = title bar
		WidgetX: 10, WidgetY: 5, WidgetW: 40, WidgetH: 15,
	}
	changed := Window.OnEvent(props, state, evt)
	if !changed {
		t.Error("mousedown on title bar should return true (dirty)")
	}
	if !evt.CaptureMouse {
		t.Error("mousedown on title bar should set CaptureMouse")
	}

	ws := state.(*WindowState)
	if !ws.Dragging {
		t.Error("state should be Dragging after title bar mousedown")
	}

	// Mousemove to new position
	evt2 := &render.WidgetEvent{
		Type:    "mousemove",
		X:       20, // moved 5 right
		Y:       8,  // moved 2 down
		WidgetX: 10, WidgetY: 5, WidgetW: 40, WidgetH: 15,
	}
	changed2 := Window.OnEvent(props, state, evt2)
	if !changed2 {
		t.Error("mousemove during drag should return true")
	}

	// Check FireOnChange has move data
	moveData, ok := evt2.FireOnChange.(map[string]any)
	if !ok {
		t.Fatalf("expected map FireOnChange, got %T", evt2.FireOnChange)
	}
	if moveData["type"] != "move" {
		t.Errorf("expected type='move', got %v", moveData["type"])
	}
	// Original pos (10,5) + delta (5,2) = (15,7)
	if moveData["x"] != 15 {
		t.Errorf("expected x=15, got %v", moveData["x"])
	}
	if moveData["y"] != 7 {
		t.Errorf("expected y=7, got %v", moveData["y"])
	}

	// Mouseup releases drag
	evt3 := &render.WidgetEvent{Type: "mouseup", X: 20, Y: 8}
	changed3 := Window.OnEvent(props, state, evt3)
	if !changed3 {
		t.Error("mouseup after drag should return true")
	}
	if ws.Dragging {
		t.Error("state should not be Dragging after mouseup")
	}
}

func TestWindowResizeHandle(t *testing.T) {
	state := Window.NewState()
	props := map[string]any{
		"title":  "Resizable",
		"x":      0,
		"y":      0,
		"width":  40,
		"height": 15,
	}

	// Mousedown on resize handle (bottom-right corner)
	// relY = 14 (>= h-2=13), relX = 39 (>= w-3=37)
	evt := &render.WidgetEvent{
		Type:    "mousedown",
		X:       39,
		Y:       14,
		WidgetX: 0, WidgetY: 0, WidgetW: 40, WidgetH: 15,
	}
	changed := Window.OnEvent(props, state, evt)
	if !changed {
		t.Error("mousedown on resize handle should return true")
	}
	if !evt.CaptureMouse {
		t.Error("mousedown on resize handle should set CaptureMouse")
	}

	ws := state.(*WindowState)
	if !ws.Resizing {
		t.Error("state should be Resizing after resize handle mousedown")
	}

	// Mousemove to resize
	evt2 := &render.WidgetEvent{
		Type:    "mousemove",
		X:       44, // moved 5 right
		Y:       17, // moved 3 down
		WidgetX: 0, WidgetY: 0, WidgetW: 40, WidgetH: 15,
	}
	changed2 := Window.OnEvent(props, state, evt2)
	if !changed2 {
		t.Error("mousemove during resize should return true")
	}

	resizeData, ok := evt2.FireOnChange.(map[string]any)
	if !ok {
		t.Fatalf("expected map FireOnChange, got %T", evt2.FireOnChange)
	}
	if resizeData["type"] != "resize" {
		t.Errorf("expected type='resize', got %v", resizeData["type"])
	}
	// Original (40,15) + delta (5,3) = (45,18)
	if resizeData["width"] != 45 {
		t.Errorf("expected width=45, got %v", resizeData["width"])
	}
	if resizeData["height"] != 18 {
		t.Errorf("expected height=18, got %v", resizeData["height"])
	}
}

func TestWindowMinimumSize(t *testing.T) {
	state := Window.NewState()
	props := map[string]any{
		"title":  "Small",
		"x":      0,
		"y":      0,
		"width":  40,
		"height": 15,
	}

	// Start resize
	evt := &render.WidgetEvent{
		Type:    "mousedown",
		X:       39,
		Y:       14,
		WidgetX: 0, WidgetY: 0, WidgetW: 40, WidgetH: 15,
	}
	Window.OnEvent(props, state, evt)

	// Try to resize very small (delta = -35, -15 → would be 5, 0)
	evt2 := &render.WidgetEvent{
		Type:    "mousemove",
		X:       4,  // delta = -35
		Y:       -1, // delta = -15
		WidgetX: 0, WidgetY: 0, WidgetW: 40, WidgetH: 15,
	}
	Window.OnEvent(props, state, evt2)

	resizeData := evt2.FireOnChange.(map[string]any)
	// min width=10, min height=5
	if resizeData["width"] != 10 {
		t.Errorf("expected min width=10, got %v", resizeData["width"])
	}
	if resizeData["height"] != 5 {
		t.Errorf("expected min height=5, got %v", resizeData["height"])
	}
}

func TestWindowCloseButton(t *testing.T) {
	state := Window.NewState()
	props := map[string]any{
		"title":  "Closable",
		"x":      10,
		"y":      5,
		"width":  40,
		"height": 15,
	}

	// Click on close button area (top-right corner, relY=1, relX >= widgetW-3)
	// Test via "click" event (what HandleClick dispatches)
	evt := &render.WidgetEvent{
		Type:    "click",
		X:       48, // widgetX=10, widgetW=40, relX=38 >= 37
		Y:       6,  // widgetY=5, relY=1 = title bar
		WidgetX: 10, WidgetY: 5, WidgetW: 40, WidgetH: 15,
	}
	Window.OnEvent(props, state, evt)

	if evt.FireOnChange != "close" {
		t.Errorf("expected FireOnChange='close', got %v", evt.FireOnChange)
	}
	if evt.CaptureMouse {
		t.Error("close button should NOT set CaptureMouse")
	}

	// Also verify mousedown on close button fires close
	evt2 := &render.WidgetEvent{
		Type:    "mousedown",
		X:       48,
		Y:       6,
		WidgetX: 10, WidgetY: 5, WidgetW: 40, WidgetH: 15,
	}
	Window.OnEvent(props, state, evt2)
	if evt2.FireOnChange != "close" {
		t.Errorf("mousedown on close: expected FireOnChange='close', got %v", evt2.FireOnChange)
	}
}
