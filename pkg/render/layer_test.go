package render

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func newLayerTestEngine(t *testing.T) (*Engine, *lua.State) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	e := NewEngine(L, 40, 10)
	e.RegisterLuaAPI()
	return e, L
}

func TestLayerCreation(t *testing.T) {
	e, L := newLayerTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("box", {
					style = {background = "#000"},
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// After RenderAll, syncMainLayer should have created layer 0
	if len(e.Layers()) == 0 {
		t.Fatal("expected at least 1 layer after RenderAll")
	}
	if e.Layers()[0].ID != "_main" {
		t.Errorf("expected layer 0 ID '_main', got %q", e.Layers()[0].ID)
	}
	if e.Layers()[0].Root == nil {
		t.Fatal("layer 0 Root is nil")
	}
}

func TestLayerPaintOrder(t *testing.T) {
	e, L := newLayerTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {
					style = {foreground = "#FFF", background = "#000"},
				}, "AAAA")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Verify main layer painted
	c := e.Buffer().Get(0, 0)
	if c.Ch != 'A' {
		t.Errorf("expected 'A' at (0,0), got %c (%d)", c.Ch, c.Ch)
	}

	// Create overlay layer that covers (0,0) with 'B'
	overlayRoot := NewNode("text")
	overlayRoot.Content = "BBBB"
	overlayRoot.Style.Foreground = "#0F0"
	overlayRoot.Style.Background = "#00F"
	overlayRoot.Style.Width = 10
	overlayRoot.Style.Height = 1
	overlayRoot.Style.Left = 0
	overlayRoot.Style.Top = 0

	e.CreateLayer("overlay1", overlayRoot, false)
	e.RenderDirty()

	// Now (0,0) should have 'B' from the overlay
	c = e.Buffer().Get(0, 0)
	if c.Ch != 'B' {
		t.Errorf("expected 'B' at (0,0) after overlay, got %c (%d)", c.Ch, c.Ch)
	}
	if c.FG != "#0F0" {
		t.Errorf("expected FG '#0F0', got %q", c.FG)
	}
}

func TestLayerEventDispatch(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	// Create main layer manually (no Lua needed for this test)
	mainRoot := NewNode("box")
	mainRoot.X = 0
	mainRoot.Y = 0
	mainRoot.W = 40
	mainRoot.H = 10
	e.syncMainLayer()
	e.Layers()[0].Root = mainRoot

	// Create overlay layer at (5,2) size 10x3
	overlayRoot := NewNode("box")
	overlayRoot.Style.Left = 5
	overlayRoot.Style.Top = 2
	overlayRoot.Style.Width = 10
	overlayRoot.Style.Height = 3
	overlayRoot.X = 5
	overlayRoot.Y = 2
	overlayRoot.W = 10
	overlayRoot.H = 3

	e.CreateLayer("overlay", overlayRoot, false)

	// Hit test at (7, 3) — should hit overlay
	node, layer := e.hitTestLayers(7, 3)
	if node == nil {
		t.Fatal("expected hit in overlay area")
	}
	if layer == nil || layer.ID != "overlay" {
		t.Errorf("expected overlay layer, got %v", layer)
	}

	// Hit test at (0, 0) — should hit main layer
	node, layer = e.hitTestLayers(0, 0)
	if node == nil {
		t.Fatal("expected hit in main layer area")
	}
	if layer == nil || layer.ID != "_main" {
		t.Errorf("expected main layer, got %v", layer)
	}
}

func TestLayerModal(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	// Create main layer
	mainRoot := NewNode("box")
	mainRoot.X = 0
	mainRoot.Y = 0
	mainRoot.W = 40
	mainRoot.H = 10
	e.syncMainLayer()
	e.Layers()[0].Root = mainRoot

	// Create modal overlay at (10,3) size 10x4
	overlayRoot := NewNode("box")
	overlayRoot.X = 10
	overlayRoot.Y = 3
	overlayRoot.W = 10
	overlayRoot.H = 4

	e.CreateLayer("dialog", overlayRoot, true) // modal=true

	// Hit test at (0, 0) — outside modal, should be blocked
	node, layer := e.hitTestLayers(0, 0)
	if node != nil {
		t.Error("expected nil node for click outside modal")
	}
	if layer == nil || layer.ID != "dialog" {
		t.Errorf("expected dialog layer (modal block), got %v", layer)
	}

	// Hit test at (12, 4) — inside modal, should hit
	node, layer = e.hitTestLayers(12, 4)
	if node == nil {
		t.Fatal("expected hit inside modal area")
	}
	if layer == nil || layer.ID != "dialog" {
		t.Errorf("expected dialog layer, got %v", layer)
	}
}

func TestLayerRemove(t *testing.T) {
	e, L := newLayerTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {
					style = {foreground = "#FFF", background = "#000"},
				}, "MAIN")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Create overlay
	overlayRoot := NewNode("text")
	overlayRoot.Content = "OVER"
	overlayRoot.Style.Foreground = "#0F0"
	overlayRoot.Style.Background = "#00F"
	overlayRoot.Style.Width = 10
	overlayRoot.Style.Height = 1
	overlayRoot.Style.Left = 0
	overlayRoot.Style.Top = 0

	e.CreateLayer("temp", overlayRoot, false)
	e.RenderDirty()

	// Verify overlay is visible
	c := e.Buffer().Get(0, 0)
	if c.Ch != 'O' {
		t.Errorf("expected 'O' at (0,0) with overlay, got %c", c.Ch)
	}

	// Remove overlay
	e.RemoveLayer("temp")

	if len(e.Layers()) != 1 {
		t.Errorf("expected 1 layer after removal, got %d", len(e.Layers()))
	}

	// After RenderDirty, main content should be restored
	e.RenderDirty()
	c = e.Buffer().Get(0, 0)
	if c.Ch != 'M' {
		t.Errorf("expected 'M' at (0,0) after overlay removal, got %c (%d)", c.Ch, c.Ch)
	}
}

func TestLayerRemoveMainBlocked(t *testing.T) {
	e, _ := newLayerTestEngine(t)
	e.syncMainLayer()

	// Try to remove the main layer — should be blocked
	e.RemoveLayer("_main")
	if len(e.Layers()) != 1 {
		t.Errorf("main layer should not be removable, got %d layers", len(e.Layers()))
	}
}

func TestLayerBringToFront(t *testing.T) {
	e, _ := newLayerTestEngine(t)
	e.syncMainLayer()

	root1 := NewNode("box")
	root1.X = 0
	root1.Y = 0
	root1.W = 10
	root1.H = 5
	root2 := NewNode("box")
	root2.X = 0
	root2.Y = 0
	root2.W = 10
	root2.H = 5

	e.CreateLayer("layer1", root1, false)
	e.CreateLayer("layer2", root2, false)

	// layer2 is on top
	if e.Layers()[2].ID != "layer2" {
		t.Errorf("expected layer2 on top, got %s", e.Layers()[2].ID)
	}

	// Bring layer1 to front
	e.BringToFront("layer1")
	if e.Layers()[2].ID != "layer1" {
		t.Errorf("expected layer1 on top after BringToFront, got %s", e.Layers()[2].ID)
	}
}

func TestSingleLayerBackwardCompat(t *testing.T) {
	// With only the main layer (no overlays), behavior should be identical to before
	e, L := newLayerTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "app",
			render = function(props)
				return lumina.createElement("box", {
					style = {background = "#1E1E2E"},
				},
					lumina.createElement("text", {
						id = "greeting",
						style = {foreground = "#CDD6F4"},
					}, "Hello World")
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Verify rendering works
	comp := e.GetComponent("app")
	if comp == nil {
		t.Fatal("component not registered")
	}
	if comp.RootNode == nil {
		t.Fatal("RootNode nil")
	}

	// Verify layer 0 exists and points to root
	if len(e.Layers()) != 1 {
		t.Errorf("expected 1 layer, got %d", len(e.Layers()))
	}
	if e.Layers()[0].Root != comp.RootNode {
		t.Error("layer 0 Root does not match root component's RootNode")
	}

	// Verify events still work (HandleMouseMove should not panic)
	e.HandleMouseMove(5, 5)
	e.HandleClick(5, 5)
	e.HandleKeyDown("a")
	e.HandleScroll(5, 5, 1)
}

func TestLayerFocusCycling(t *testing.T) {
	e, _ := newLayerTestEngine(t)

	// Create main layer with a focusable node
	mainRoot := NewNode("box")
	mainRoot.X = 0
	mainRoot.Y = 0
	mainRoot.W = 40
	mainRoot.H = 10

	mainInput := NewNode("input")
	mainInput.Focusable = true
	mainInput.X = 0
	mainInput.Y = 0
	mainInput.W = 10
	mainInput.H = 1
	mainRoot.AddChild(mainInput)

	e.syncMainLayer()
	e.Layers()[0].Root = mainRoot

	// Create modal overlay with its own focusable node
	overlayRoot := NewNode("box")
	overlayRoot.X = 10
	overlayRoot.Y = 3
	overlayRoot.W = 20
	overlayRoot.H = 4

	overlayInput := NewNode("input")
	overlayInput.Focusable = true
	overlayInput.X = 11
	overlayInput.Y = 4
	overlayInput.W = 10
	overlayInput.H = 1
	overlayRoot.AddChild(overlayInput)

	e.CreateLayer("modal-dialog", overlayRoot, true)

	// FocusNext should only cycle within the modal layer
	e.FocusNext()
	if e.FocusedNode() != overlayInput {
		t.Error("expected focus on overlay input in modal layer")
	}
}
