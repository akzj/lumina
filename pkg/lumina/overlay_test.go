package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// -----------------------------------------------------------------------
// OverlayManager tests
// -----------------------------------------------------------------------

func TestOverlayManager_ShowHide(t *testing.T) {
	om := NewOverlayManager()

	ov := &Overlay{ID: "dialog1", X: 5, Y: 3, W: 20, H: 10, ZIndex: 100}
	om.Show(ov)

	if !om.IsVisible("dialog1") {
		t.Fatal("expected dialog1 to be visible after Show")
	}
	if om.Count() != 1 {
		t.Fatalf("expected 1 overlay, got %d", om.Count())
	}

	om.Hide("dialog1")
	if om.IsVisible("dialog1") {
		t.Fatal("expected dialog1 to be hidden after Hide")
	}

	// Overlay still exists, just hidden
	if om.Count() != 1 {
		t.Fatalf("expected 1 overlay after Hide, got %d", om.Count())
	}

	// Remove entirely
	om.Remove("dialog1")
	if om.Count() != 0 {
		t.Fatalf("expected 0 overlays after Remove, got %d", om.Count())
	}
}

func TestOverlayManager_MultipleOverlaysSortedByZIndex(t *testing.T) {
	om := NewOverlayManager()

	om.Show(&Overlay{ID: "low", ZIndex: 10, Visible: true})
	om.Show(&Overlay{ID: "high", ZIndex: 100, Visible: true})
	om.Show(&Overlay{ID: "mid", ZIndex: 50, Visible: true})

	visible := om.GetVisible()
	if len(visible) != 3 {
		t.Fatalf("expected 3 visible overlays, got %d", len(visible))
	}
	if visible[0].ID != "low" || visible[1].ID != "mid" || visible[2].ID != "high" {
		t.Fatalf("expected order [low, mid, high], got [%s, %s, %s]",
			visible[0].ID, visible[1].ID, visible[2].ID)
	}
}

func TestOverlayManager_ModalBackdrop(t *testing.T) {
	om := NewOverlayManager()

	// Create a modal overlay
	content := NewVNode("box")
	content.AddChild(&VNode{Type: "text", Content: "Dialog", Props: make(map[string]any)})

	om.Show(&Overlay{
		ID: "modal1", VNode: content,
		X: 5, Y: 3, W: 20, H: 10,
		ZIndex: 100, Modal: true,
	})

	// Render with overlays
	base := NewVNode("box")
	base.AddChild(&VNode{Type: "text", Content: "Base content", Props: make(map[string]any)})

	frame := VNodeToFrameWithOverlays(base, 40, 20, om.GetVisible())

	// The backdrop should dim the frame
	// Check that cells are dimmed (modal overlay renders backdrop)
	if !frame.Cells[0][0].Dim {
		t.Fatal("expected cells to be dimmed by modal backdrop")
	}
}

// -----------------------------------------------------------------------
// Position absolute tests
// -----------------------------------------------------------------------

func TestPositionAbsolute_RelativeToParent(t *testing.T) {
	// Parent box at (0,0) 40x20, child with position:absolute at top:2, left:5
	parent := NewVNode("box")
	parent.Props = map[string]any{
		"style": map[string]any{
			"width": int64(40), "height": int64(20),
		},
	}

	child := NewVNode("text")
	child.Content = "Floating"
	child.Props = map[string]any{
		"style": map[string]any{
			"position": "absolute",
			"top":      int64(2),
			"left":     int64(5),
			"width":    int64(10),
			"height":   int64(3),
		},
	}
	parent.AddChild(child)

	computeFlexLayout(parent, 0, 0, 40, 20)

	if child.X != 5 {
		t.Fatalf("expected child X=5, got %d", child.X)
	}
	if child.Y != 2 {
		t.Fatalf("expected child Y=2, got %d", child.Y)
	}
	if child.W != 10 {
		t.Fatalf("expected child W=10, got %d", child.W)
	}
}

func TestPositionFixed_RelativeToScreen(t *testing.T) {
	// Parent at (10, 5), child with position:fixed should be relative to screen (0,0)
	parent := NewVNode("box")
	parent.Props = map[string]any{}

	child := NewVNode("text")
	child.Content = "Fixed"
	child.Props = map[string]any{
		"style": map[string]any{
			"position": "fixed",
			"top":      int64(0),
			"left":     int64(0),
			"width":    int64(15),
			"height":   int64(1),
		},
	}
	parent.AddChild(child)

	computeFlexLayout(parent, 10, 5, 30, 15)

	// Fixed position should be relative to screen, not parent
	if child.X != 0 {
		t.Fatalf("expected child X=0 (fixed), got %d", child.X)
	}
	if child.Y != 0 {
		t.Fatalf("expected child Y=0 (fixed), got %d", child.Y)
	}
}

func TestOverlay_VNodeRendering(t *testing.T) {
	// Create a base frame, then render an overlay on top
	base := NewVNode("box")
	base.Props = map[string]any{}
	baseText := &VNode{Type: "text", Content: "Background", Props: make(map[string]any)}
	base.AddChild(baseText)

	// Create overlay content
	ovContent := NewVNode("box")
	ovContent.Props = map[string]any{}
	ovText := &VNode{Type: "text", Content: "Popup!", Props: make(map[string]any)}
	ovContent.AddChild(ovText)

	overlays := []*Overlay{
		{ID: "popup", VNode: ovContent, X: 5, Y: 3, W: 10, H: 5, ZIndex: 10, Visible: true},
	}

	frame := VNodeToFrameWithOverlays(base, 40, 20, overlays)

	// Check that the overlay content appears at the expected position
	// The overlay text "Popup!" should be at row 3, col 5
	if frame.Height < 4 || frame.Width < 15 {
		t.Fatalf("frame too small: %dx%d", frame.Width, frame.Height)
	}

	// Check that "Popup!" appears at the overlay position
	got := ""
	for x := 5; x < 11 && x < frame.Width; x++ {
		if frame.Cells[3][x].Char != 0 {
			got += string(frame.Cells[3][x].Char)
		}
	}
	if got != "Popup!" {
		t.Fatalf("expected 'Popup!' at overlay position, got %q", got)
	}
}

func TestAbsoluteChildren_ExcludedFromFlexFlow(t *testing.T) {
	// A vbox with 3 children: normal, absolute, normal
	// The absolute child should not affect the layout of the other two
	parent := NewVNode("vbox")
	parent.Props = map[string]any{
		"style": map[string]any{"width": int64(20), "height": int64(10)},
	}

	child1 := &VNode{Type: "text", Content: "A", Props: map[string]any{
		"style": map[string]any{"height": int64(2)},
	}}
	absChild := &VNode{Type: "text", Content: "Float", Props: map[string]any{
		"style": map[string]any{
			"position": "absolute", "top": int64(0), "left": int64(0),
			"width": int64(5), "height": int64(1),
		},
	}}
	child2 := &VNode{Type: "text", Content: "B", Props: map[string]any{
		"style": map[string]any{"height": int64(2)},
	}}

	parent.AddChild(child1)
	parent.AddChild(absChild)
	parent.AddChild(child2)

	computeFlexLayout(parent, 0, 0, 20, 10)

	// child1 should be at Y=0, child2 should be at Y=2 (right after child1)
	// absChild should NOT create a gap between them
	if child1.Y != 0 {
		t.Fatalf("expected child1 Y=0, got %d", child1.Y)
	}
	if child2.Y != 2 {
		t.Fatalf("expected child2 Y=2 (no gap from absolute child), got %d", child2.Y)
	}
}

func TestTopLeftRightBottom_Positioning(t *testing.T) {
	parent := NewVNode("box")
	parent.Props = map[string]any{
		"style": map[string]any{"width": int64(40), "height": int64(20)},
	}

	// Child anchored to bottom-right
	child := &VNode{Type: "text", Content: "BR", Props: map[string]any{
		"style": map[string]any{
			"position": "absolute",
			"right":    int64(2),
			"bottom":   int64(1),
			"width":    int64(5),
			"height":   int64(3),
		},
	}}
	parent.AddChild(child)

	computeFlexLayout(parent, 0, 0, 40, 20)

	// right=2, width=5 → X = 40 - 5 - 2 = 33
	// bottom=1, height=3 → Y = 20 - 3 - 1 = 16
	if child.X != 33 {
		t.Fatalf("expected child X=33 (right-anchored), got %d", child.X)
	}
	if child.Y != 16 {
		t.Fatalf("expected child Y=16 (bottom-anchored), got %d", child.Y)
	}
}

// -----------------------------------------------------------------------
// Lua API tests
// -----------------------------------------------------------------------

func TestLua_ShowOverlayAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalOverlayManager.Clear()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.showOverlay({
			id = "test-overlay",
			content = { type = "text", content = "Hello Overlay" },
			x = 10, y = 5,
			width = 30, height = 10,
			zIndex = 50,
			modal = false,
		})
	`)
	if err != nil {
		t.Fatalf("showOverlay error: %v", err)
	}

	if !globalOverlayManager.IsVisible("test-overlay") {
		t.Fatal("expected test-overlay to be visible")
	}

	ov := globalOverlayManager.Get("test-overlay")
	if ov == nil {
		t.Fatal("expected overlay to exist")
	}
	if ov.X != 10 || ov.Y != 5 {
		t.Fatalf("expected position (10,5), got (%d,%d)", ov.X, ov.Y)
	}
	if ov.W != 30 || ov.H != 10 {
		t.Fatalf("expected size (30,10), got (%d,%d)", ov.W, ov.H)
	}
	if ov.ZIndex != 50 {
		t.Fatalf("expected zIndex=50, got %d", ov.ZIndex)
	}
	if ov.Modal {
		t.Fatal("expected modal=false")
	}
	globalOverlayManager.Clear()
}

func TestLua_HideOverlayAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalOverlayManager.Clear()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.showOverlay({
			id = "hide-test",
			content = { type = "text", content = "Temp" },
			x = 0, y = 0, width = 10, height = 5,
		})
		lumina.hideOverlay("hide-test")
	`)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if globalOverlayManager.IsVisible("hide-test") {
		t.Fatal("expected hide-test to be hidden after hideOverlay")
	}
	globalOverlayManager.Clear()
}

func TestLua_ToggleOverlayAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalOverlayManager.Clear()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.showOverlay({
			id = "toggle-test",
			content = { type = "text", content = "Toggle" },
			x = 0, y = 0, width = 10, height = 5,
		})
		_vis1 = lumina.isOverlayVisible("toggle-test")
		lumina.toggleOverlay("toggle-test")
		_vis2 = lumina.isOverlayVisible("toggle-test")
		lumina.toggleOverlay("toggle-test")
		_vis3 = lumina.isOverlayVisible("toggle-test")
	`)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	L.GetGlobal("_vis1")
	if !L.ToBoolean(-1) {
		t.Fatal("expected visible after show")
	}
	L.Pop(1)

	L.GetGlobal("_vis2")
	if L.ToBoolean(-1) {
		t.Fatal("expected hidden after first toggle")
	}
	L.Pop(1)

	L.GetGlobal("_vis3")
	if !L.ToBoolean(-1) {
		t.Fatal("expected visible after second toggle")
	}
	L.Pop(1)
	globalOverlayManager.Clear()
}

// -----------------------------------------------------------------------
// Modal input routing tests
// -----------------------------------------------------------------------

func TestModalOverlay_EscapeCloses(t *testing.T) {
	globalOverlayManager.Clear()

	content := NewVNode("box")
	content.Props = make(map[string]any)
	globalOverlayManager.Show(&Overlay{
		ID: "esc-modal", VNode: content,
		X: 0, Y: 0, W: 20, H: 10,
		ZIndex: 100, Modal: true,
	})

	if !globalOverlayManager.IsVisible("esc-modal") {
		t.Fatal("expected modal to be visible")
	}

	// Simulate Escape key — the app.handleEvent would call this
	topModal := globalOverlayManager.GetTopModal()
	if topModal == nil {
		t.Fatal("expected top modal")
	}
	if topModal.ID != "esc-modal" {
		t.Fatalf("expected top modal ID 'esc-modal', got %q", topModal.ID)
	}

	// Escape closes the modal
	globalOverlayManager.Hide(topModal.ID)

	if globalOverlayManager.IsVisible("esc-modal") {
		t.Fatal("expected modal to be hidden after Escape")
	}
	globalOverlayManager.Clear()
}

func TestModalOverlay_GetTopModal(t *testing.T) {
	globalOverlayManager.Clear()

	// No modals → nil
	if globalOverlayManager.GetTopModal() != nil {
		t.Fatal("expected nil when no modals")
	}

	// Add non-modal overlay
	globalOverlayManager.Show(&Overlay{
		ID: "popup", X: 0, Y: 0, W: 10, H: 5, ZIndex: 50,
	})
	if globalOverlayManager.GetTopModal() != nil {
		t.Fatal("expected nil when only non-modal overlays")
	}

	// Add modal overlay
	globalOverlayManager.Show(&Overlay{
		ID: "modal1", X: 0, Y: 0, W: 10, H: 5, ZIndex: 100, Modal: true,
	})
	top := globalOverlayManager.GetTopModal()
	if top == nil || top.ID != "modal1" {
		t.Fatal("expected modal1 as top modal")
	}

	// Add higher-ZIndex modal
	globalOverlayManager.Show(&Overlay{
		ID: "modal2", X: 0, Y: 0, W: 10, H: 5, ZIndex: 200, Modal: true,
	})
	top = globalOverlayManager.GetTopModal()
	if top == nil || top.ID != "modal2" {
		t.Fatalf("expected modal2 as top modal, got %v", top)
	}

	globalOverlayManager.Clear()
}
