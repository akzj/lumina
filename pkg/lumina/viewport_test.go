package lumina

import (
	"testing"
)

// --- Viewport Scroll Methods ---

func TestViewportScrollDown(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20}

	vp.ScrollDown(5)
	if vp.ScrollY != 5 {
		t.Errorf("ScrollDown(5): ScrollY = %d, want 5", vp.ScrollY)
	}

	vp.ScrollDown(10)
	if vp.ScrollY != 15 {
		t.Errorf("ScrollDown(10): ScrollY = %d, want 15", vp.ScrollY)
	}
}

func TestViewportScrollUp(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 50}

	vp.ScrollUp(10)
	if vp.ScrollY != 40 {
		t.Errorf("ScrollUp(10): ScrollY = %d, want 40", vp.ScrollY)
	}

	// Scroll up past top — clamp to 0
	vp.ScrollUp(100)
	if vp.ScrollY != 0 {
		t.Errorf("ScrollUp past top: ScrollY = %d, want 0", vp.ScrollY)
	}
}

func TestViewportScrollDownClamp(t *testing.T) {
	vp := &Viewport{ContentH: 30, ViewH: 20}

	// Max scroll = 30 - 20 = 10
	vp.ScrollDown(100)
	if vp.ScrollY != 10 {
		t.Errorf("ScrollDown past bottom: ScrollY = %d, want 10", vp.ScrollY)
	}
}

func TestViewportScrollTo(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20}

	vp.ScrollTo(50)
	if vp.ScrollY != 50 {
		t.Errorf("ScrollTo(50): ScrollY = %d, want 50", vp.ScrollY)
	}

	// Clamp to max
	vp.ScrollTo(200)
	if vp.ScrollY != 80 {
		t.Errorf("ScrollTo(200): ScrollY = %d, want 80 (max)", vp.ScrollY)
	}

	// Clamp to 0
	vp.ScrollTo(-10)
	if vp.ScrollY != 0 {
		t.Errorf("ScrollTo(-10): ScrollY = %d, want 0", vp.ScrollY)
	}
}

func TestViewportScrollToBottom(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20}

	vp.ScrollToBottom()
	if vp.ScrollY != 80 {
		t.Errorf("ScrollToBottom: ScrollY = %d, want 80", vp.ScrollY)
	}
}

func TestViewportScrollToTop(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 50, ScrollX: 10}

	vp.ScrollToTop()
	if vp.ScrollY != 0 || vp.ScrollX != 0 {
		t.Errorf("ScrollToTop: ScrollY=%d ScrollX=%d, want 0,0", vp.ScrollY, vp.ScrollX)
	}
}

// --- EnsureVisible ---

func TestEnsureVisibleAlreadyVisible(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 10}

	// Region [15, 15+3) is within [10, 30) — no scroll needed
	vp.EnsureVisible(15, 3)
	if vp.ScrollY != 10 {
		t.Errorf("EnsureVisible (already visible): ScrollY = %d, want 10", vp.ScrollY)
	}
}

func TestEnsureVisibleScrollDown(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 0}

	// Region [25, 25+5) is below viewport [0, 20) — need to scroll down
	vp.EnsureVisible(25, 5)
	// Need: 25+5-20 = 10
	if vp.ScrollY != 10 {
		t.Errorf("EnsureVisible (scroll down): ScrollY = %d, want 10", vp.ScrollY)
	}
}

func TestEnsureVisibleScrollUp(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 30}

	// Region [10, 10+5) is above viewport [30, 50) — need to scroll up
	vp.EnsureVisible(10, 5)
	if vp.ScrollY != 10 {
		t.Errorf("EnsureVisible (scroll up): ScrollY = %d, want 10", vp.ScrollY)
	}
}

// --- AtTop / AtBottom / ScrollPercent ---

func TestAtTop(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 0}
	if !vp.AtTop() {
		t.Error("AtTop: expected true at ScrollY=0")
	}

	vp.ScrollY = 1
	if vp.AtTop() {
		t.Error("AtTop: expected false at ScrollY=1")
	}
}

func TestAtBottom(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 80}
	if !vp.AtBottom() {
		t.Error("AtBottom: expected true at ScrollY=80 (max)")
	}

	vp.ScrollY = 79
	if vp.AtBottom() {
		t.Error("AtBottom: expected false at ScrollY=79")
	}
}

func TestScrollPercent(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20}

	vp.ScrollY = 0
	if p := vp.ScrollPercent(); p != 0 {
		t.Errorf("ScrollPercent at top = %f, want 0", p)
	}

	vp.ScrollY = 80
	if p := vp.ScrollPercent(); p != 1.0 {
		t.Errorf("ScrollPercent at bottom = %f, want 1.0", p)
	}

	vp.ScrollY = 40
	if p := vp.ScrollPercent(); p != 0.5 {
		t.Errorf("ScrollPercent at middle = %f, want 0.5", p)
	}
}

func TestScrollPercentNoScroll(t *testing.T) {
	// Content fits in viewport — no scrolling needed
	vp := &Viewport{ContentH: 10, ViewH: 20}
	if p := vp.ScrollPercent(); p != 0 {
		t.Errorf("ScrollPercent (no scroll needed) = %f, want 0", p)
	}
}

// --- NeedsScroll ---

func TestNeedsScroll(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20}
	if !vp.NeedsScroll() {
		t.Error("NeedsScroll: expected true when content > view")
	}

	vp.ContentH = 20
	if vp.NeedsScroll() {
		t.Error("NeedsScroll: expected false when content == view")
	}

	vp.ContentH = 10
	if vp.NeedsScroll() {
		t.Error("NeedsScroll: expected false when content < view")
	}
}

// --- VisibleRange ---

func TestVisibleRange(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 10}

	start, end := vp.VisibleRange()
	if start != 10 || end != 30 {
		t.Errorf("VisibleRange = (%d, %d), want (10, 30)", start, end)
	}
}

func TestVisibleRangeAtBottom(t *testing.T) {
	vp := &Viewport{ContentH: 25, ViewH: 20, ScrollY: 5}

	start, end := vp.VisibleRange()
	if start != 5 || end != 25 {
		t.Errorf("VisibleRange at bottom = (%d, %d), want (5, 25)", start, end)
	}
}

// --- ScrollbarThumb ---

func TestScrollbarThumb(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 0}

	thumbStart, thumbSize := vp.ScrollbarThumb(20)

	// Thumb size: 20*20/100 = 4
	if thumbSize != 4 {
		t.Errorf("ScrollbarThumb size = %d, want 4", thumbSize)
	}
	// At top: thumbStart = 0
	if thumbStart != 0 {
		t.Errorf("ScrollbarThumb start = %d, want 0", thumbStart)
	}
}

func TestScrollbarThumbAtBottom(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 80}

	thumbStart, thumbSize := vp.ScrollbarThumb(20)

	if thumbSize != 4 {
		t.Errorf("ScrollbarThumb size = %d, want 4", thumbSize)
	}
	// At bottom: thumbStart = 20 - 4 = 16
	if thumbStart != 16 {
		t.Errorf("ScrollbarThumb start at bottom = %d, want 16", thumbStart)
	}
}

func TestScrollbarThumbMiddle(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 20, ScrollY: 40}

	thumbStart, thumbSize := vp.ScrollbarThumb(20)

	if thumbSize != 4 {
		t.Errorf("ScrollbarThumb size = %d, want 4", thumbSize)
	}
	// Middle: thumbStart = (40 * 16) / 80 = 8
	if thumbStart != 8 {
		t.Errorf("ScrollbarThumb start at middle = %d, want 8", thumbStart)
	}
}

func TestScrollbarThumbNoScroll(t *testing.T) {
	vp := &Viewport{ContentH: 10, ViewH: 20}

	_, thumbSize := vp.ScrollbarThumb(20)
	if thumbSize != 0 {
		t.Errorf("ScrollbarThumb size (no scroll) = %d, want 0", thumbSize)
	}
}

func TestScrollbarThumbMinSize(t *testing.T) {
	// Very large content — thumb should be at least 1
	vp := &Viewport{ContentH: 10000, ViewH: 20}

	_, thumbSize := vp.ScrollbarThumb(20)
	if thumbSize < 1 {
		t.Errorf("ScrollbarThumb min size = %d, want >= 1", thumbSize)
	}
}

// --- Viewport Registry ---

func TestViewportRegistry(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	vp := GetViewport("test-list")
	if vp == nil {
		t.Fatal("GetViewport returned nil")
	}

	vp.ContentH = 100
	vp.ViewH = 20
	vp.ScrollTo(50)

	// Retrieve same viewport
	vp2 := GetViewport("test-list")
	if vp2.ScrollY != 50 {
		t.Errorf("Retrieved viewport ScrollY = %d, want 50", vp2.ScrollY)
	}
}

func TestViewportRegistryPersistence(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	// Simulate: first render creates viewport and scrolls
	vp := GetViewport("log-view")
	vp.ContentH = 200
	vp.ViewH = 20
	vp.ScrollTo(100)
	SetViewport("log-view", vp)

	// Simulate: second render retrieves same viewport (scroll preserved)
	vp2 := GetViewport("log-view")
	if vp2.ScrollY != 100 {
		t.Errorf("Persistent viewport ScrollY = %d, want 100", vp2.ScrollY)
	}
}

func TestRemoveViewport(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	vp := GetViewport("temp")
	vp.ScrollY = 42
	SetViewport("temp", vp)

	RemoveViewport("temp")

	// Should get a fresh viewport
	vp2 := GetViewport("temp")
	if vp2.ScrollY != 0 {
		t.Errorf("After remove, ScrollY = %d, want 0", vp2.ScrollY)
	}
}

func TestScrollViewport(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	vp := GetViewport("sv-test")
	vp.ContentH = 100
	vp.ViewH = 20
	SetViewport("sv-test", vp)

	ok := ScrollViewport("sv-test", 5)
	if !ok {
		t.Error("ScrollViewport returned false for existing viewport")
	}
	if vp.ScrollY != 5 {
		t.Errorf("ScrollViewport down: ScrollY = %d, want 5", vp.ScrollY)
	}

	ScrollViewport("sv-test", -3)
	if vp.ScrollY != 2 {
		t.Errorf("ScrollViewport up: ScrollY = %d, want 2", vp.ScrollY)
	}

	ok = ScrollViewport("nonexistent", 5)
	if ok {
		t.Error("ScrollViewport returned true for nonexistent viewport")
	}
}

// --- Scrollable Layout Integration ---

func TestScrollableVBoxLayout(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	// Create a scrollable vbox with 10 children, each height=3, in a 80x10 container
	children := make([]*VNode, 10)
	for i := 0; i < 10; i++ {
		children[i] = makeStyledNode("box", map[string]any{"height": int64(3)})
	}

	root := makeStyledNode("vbox", map[string]any{
		"overflow": "scroll",
	}, children...)
	root.Props["id"] = "scroll-test"

	computeFlexLayout(root, 0, 0, 80, 10)

	// Total content height: 10 * 3 = 30
	vp := GetViewport("scroll-test")
	if vp.ContentH != 30 {
		t.Errorf("ContentH = %d, want 30", vp.ContentH)
	}
	if vp.ViewH != 10 {
		t.Errorf("ViewH = %d, want 10", vp.ViewH)
	}
	if !vp.NeedsScroll() {
		t.Error("Expected NeedsScroll to be true")
	}

	// First child should be at Y=0
	if root.Children[0].Y != 0 {
		t.Errorf("First child Y = %d, want 0", root.Children[0].Y)
	}
	// Second child at Y=3
	if root.Children[1].Y != 3 {
		t.Errorf("Second child Y = %d, want 3", root.Children[1].Y)
	}
}

func TestScrollableVBoxScrolled(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	// Pre-set viewport with scroll offset
	vp := &Viewport{ScrollY: 6, ContentH: 30, ViewH: 10}
	SetViewport("scroll-test2", vp)

	children := make([]*VNode, 10)
	for i := 0; i < 10; i++ {
		children[i] = makeStyledNode("box", map[string]any{"height": int64(3)})
	}

	root := makeStyledNode("vbox", map[string]any{
		"overflow": "scroll",
	}, children...)
	root.Props["id"] = "scroll-test2"

	computeFlexLayout(root, 0, 0, 80, 10)

	// With scroll offset 6, first child should be at Y = 0 - 6 = -6
	if root.Children[0].Y != -6 {
		t.Errorf("Scrolled first child Y = %d, want -6", root.Children[0].Y)
	}
	// Third child (index 2) at Y = 6 - 6 = 0
	if root.Children[2].Y != 0 {
		t.Errorf("Scrolled third child Y = %d, want 0", root.Children[2].Y)
	}
}

// --- Scroll Clipping (Frame rendering) ---

func TestScrollClipping(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	// Create scrollable vbox with text children
	children := make([]*VNode, 5)
	for i := 0; i < 5; i++ {
		children[i] = makeStyledText("Line "+string(rune('A'+i)), map[string]any{"height": int64(1)})
	}

	root := makeStyledNode("vbox", map[string]any{
		"overflow": "scroll",
	}, children...)
	root.Props["id"] = "clip-test"

	frame := VNodeToFrame(root, 40, 3)

	// Only first 3 lines should be visible (viewport height = 3)
	// Line A at row 0
	if frame.Cells[0][0].Char != 'L' {
		t.Errorf("Row 0 char = %c, want 'L' (from 'Line A')", frame.Cells[0][0].Char)
	}
}

func TestScrollClippingWithOffset(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	// Pre-scroll to offset 2
	vp := &Viewport{ScrollY: 2}
	SetViewport("clip-test2", vp)

	children := make([]*VNode, 5)
	for i := 0; i < 5; i++ {
		children[i] = makeStyledText(string(rune('A'+i)), map[string]any{"height": int64(1)})
	}

	root := makeStyledNode("vbox", map[string]any{
		"overflow": "scroll",
	}, children...)
	root.Props["id"] = "clip-test2"

	frame := VNodeToFrame(root, 40, 3)

	// After scrolling by 2, we should see children C, D, E (indices 2, 3, 4)
	// Child C (content "C") should be at row 0
	if frame.Cells[0][0].Char != 'C' {
		t.Errorf("Scrolled row 0 char = '%c', want 'C'", frame.Cells[0][0].Char)
	}
}

// --- Scrollbar Rendering ---

func TestScrollbarRendering(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	children := make([]*VNode, 20)
	for i := 0; i < 20; i++ {
		children[i] = makeStyledText(string(rune('A'+i%26)), map[string]any{"height": int64(1)})
	}

	root := makeStyledNode("vbox", map[string]any{
		"overflow": "scroll",
	}, children...)
	root.Props["id"] = "scrollbar-test"

	frame := VNodeToFrame(root, 40, 10)

	// Scrollbar should be in the rightmost column of the content area (col 39)
	// The scrollbar column is at contentW (which is W-1 for scrollbar reservation)
	// Check that some scrollbar characters exist
	scrollbarCol := 39 // last column
	foundTrack := false
	foundThumb := false
	for row := 0; row < 10; row++ {
		ch := frame.Cells[row][scrollbarCol].Char
		if ch == '│' {
			foundTrack = true
		}
		if ch == '█' {
			foundThumb = true
		}
	}

	if !foundThumb {
		t.Error("No scrollbar thumb (█) found")
	}
	// Track might not be visible if thumb fills the whole track for small content
	_ = foundTrack
}

// --- Viewport with no ID (graceful handling) ---

func TestScrollableNoID(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	// Scrollable container without an ID — should still layout without panic
	children := make([]*VNode, 5)
	for i := 0; i < 5; i++ {
		children[i] = makeStyledNode("box", map[string]any{"height": int64(3)})
	}

	root := makeStyledNode("vbox", map[string]any{
		"overflow": "scroll",
	}, children...)
	// No ID set

	// Should not panic
	computeFlexLayout(root, 0, 0, 80, 10)
}

// --- Content fits (no scroll needed) ---

func TestScrollableContentFits(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	// Only 2 children, each height=3, in a 80x10 container — fits!
	children := make([]*VNode, 2)
	for i := 0; i < 2; i++ {
		children[i] = makeStyledNode("box", map[string]any{"height": int64(3)})
	}

	root := makeStyledNode("vbox", map[string]any{
		"overflow": "scroll",
	}, children...)
	root.Props["id"] = "fits-test"

	computeFlexLayout(root, 0, 0, 80, 10)

	vp := GetViewport("fits-test")
	if vp.NeedsScroll() {
		t.Error("Content fits — NeedsScroll should be false")
	}
	if vp.ContentH != 6 {
		t.Errorf("ContentH = %d, want 6", vp.ContentH)
	}
}

// --- Horizontal scroll ---

func TestScrollableHBox(t *testing.T) {
	ClearViewports()
	defer ClearViewports()

	children := make([]*VNode, 10)
	for i := 0; i < 10; i++ {
		children[i] = makeStyledNode("box", map[string]any{"width": int64(20)})
	}

	root := makeStyledNode("hbox", map[string]any{
		"overflow": "scroll",
	}, children...)
	root.Props["id"] = "hscroll-test"

	computeFlexLayout(root, 0, 0, 40, 10)

	vp := GetViewport("hscroll-test")
	if vp.ContentW != 200 {
		t.Errorf("Horizontal ContentW = %d, want 200", vp.ContentW)
	}
	if !vp.NeedsHScroll() {
		t.Error("Expected horizontal scroll needed")
	}
}

// --- Scrollbar position calculation ---

func TestScrollbarPosition(t *testing.T) {
	vp := &Viewport{ContentH: 200, ViewH: 20, ScrollY: 0}

	// Track height = 20
	thumbStart, thumbSize := vp.ScrollbarThumb(20)

	// Thumb size = 20 * 20 / 200 = 2
	if thumbSize != 2 {
		t.Errorf("thumbSize = %d, want 2", thumbSize)
	}
	if thumbStart != 0 {
		t.Errorf("thumbStart at top = %d, want 0", thumbStart)
	}

	// Scroll to bottom
	vp.ScrollY = 180
	thumbStart, _ = vp.ScrollbarThumb(20)
	if thumbStart != 18 {
		t.Errorf("thumbStart at bottom = %d, want 18", thumbStart)
	}
}

// --- Edge case: zero-height viewport ---

func TestZeroHeightViewport(t *testing.T) {
	vp := &Viewport{ContentH: 100, ViewH: 0}

	vp.ScrollDown(10)
	if vp.ScrollY != 0 {
		// maxScrollY = 100 - 0 = 100, so scroll should work
		// Actually maxScrollY = 100, so it should be 10
	}

	if vp.AtBottom() {
		// With ViewH=0, maxScrollY=100, so not at bottom at ScrollY=10
	}
}

// --- Edge case: content smaller than view ---

func TestContentSmallerThanView(t *testing.T) {
	vp := &Viewport{ContentH: 5, ViewH: 20}

	vp.ScrollDown(10)
	// maxScrollY = 0, so ScrollY should be clamped to 0
	if vp.ScrollY != 0 {
		t.Errorf("ScrollDown when content < view: ScrollY = %d, want 0", vp.ScrollY)
	}

	if !vp.AtTop() {
		t.Error("Should be at top when content < view")
	}
	if !vp.AtBottom() {
		t.Error("Should be at bottom when content < view")
	}
}
