package lumina

import (
	"testing"
)

func TestFlexRowBasic(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 80, H: 24}
	items := []FlexItem{
		{Basis: 10},
		{Basis: 20},
		{Basis: 15},
	}
	content := []Rect{{}, {}, {}}
	layout := FlexLayout{Direction: "row"}

	rects := CalculateFlexLayout(container, items, content, layout)
	if len(rects) != 3 {
		t.Fatalf("expected 3 rects, got %d", len(rects))
	}
	// Items should be placed left to right
	if rects[0].X != 0 {
		t.Fatalf("expected X=0, got %d", rects[0].X)
	}
	if rects[0].W != 10 {
		t.Fatalf("expected W=10, got %d", rects[0].W)
	}
	if rects[1].X != 10 {
		t.Fatalf("expected X=10, got %d", rects[1].X)
	}
	if rects[1].W != 20 {
		t.Fatalf("expected W=20, got %d", rects[1].W)
	}
	if rects[2].X != 30 {
		t.Fatalf("expected X=30, got %d", rects[2].X)
	}
}

func TestFlexColumnBasic(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 80, H: 24}
	items := []FlexItem{
		{Basis: 5},
		{Basis: 10},
	}
	content := []Rect{{}, {}}
	layout := FlexLayout{Direction: "column"}

	rects := CalculateFlexLayout(container, items, content, layout)
	if len(rects) != 2 {
		t.Fatalf("expected 2 rects, got %d", len(rects))
	}
	// Items should be placed top to bottom
	if rects[0].Y != 0 {
		t.Fatalf("expected Y=0, got %d", rects[0].Y)
	}
	if rects[0].H != 5 {
		t.Fatalf("expected H=5, got %d", rects[0].H)
	}
	if rects[1].Y != 5 {
		t.Fatalf("expected Y=5, got %d", rects[1].Y)
	}
	if rects[1].H != 10 {
		t.Fatalf("expected H=10, got %d", rects[1].H)
	}
}

func TestFlexGrowDistribution(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 100, H: 10}
	items := []FlexItem{
		{Grow: 1, Basis: 10},
		{Grow: 2, Basis: 10},
		{Grow: 1, Basis: 10},
	}
	content := []Rect{{}, {}, {}}
	layout := FlexLayout{Direction: "row"}

	rects := CalculateFlexLayout(container, items, content, layout)
	// Free space = 100 - 30 = 70, distributed 1:2:1
	// Item 0: 10 + 17 = 27 (or 28 with rounding)
	// Item 1: 10 + 35 = 45
	// Item 2: 10 + 17 = 27
	total := rects[0].W + rects[1].W + rects[2].W
	if total != 100 {
		t.Fatalf("expected total width 100, got %d", total)
	}
	// Item 1 should be roughly twice the grow of items 0 and 2
	if rects[1].W <= rects[0].W {
		t.Fatalf("item 1 (grow=2) should be wider than item 0 (grow=1): %d vs %d", rects[1].W, rects[0].W)
	}
}

func TestFlexShrink(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 50, H: 10}
	items := []FlexItem{
		{Shrink: 1, Basis: 30},
		{Shrink: 1, Basis: 30},
	}
	content := []Rect{{}, {}}
	layout := FlexLayout{Direction: "row"}

	rects := CalculateFlexLayout(container, items, content, layout)
	// Total basis = 60, container = 50, deficit = 10
	// Each shrinks by 5
	if rects[0].W != 25 {
		t.Fatalf("expected W=25, got %d", rects[0].W)
	}
	if rects[1].W != 25 {
		t.Fatalf("expected W=25, got %d", rects[1].W)
	}
}

func TestFlexJustifySpaceBetween(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 100, H: 10}
	items := []FlexItem{
		{Basis: 10},
		{Basis: 10},
		{Basis: 10},
	}
	content := []Rect{{}, {}, {}}
	layout := FlexLayout{Direction: "row", JustifyContent: "space-between"}

	rects := CalculateFlexLayout(container, items, content, layout)
	// Free space = 70, 2 gaps → 35 each
	if rects[0].X != 0 {
		t.Fatalf("expected first item at X=0, got %d", rects[0].X)
	}
	// Last item should be at right edge
	lastRight := rects[2].X + rects[2].W
	if lastRight != 100 {
		t.Fatalf("expected last item right edge at 100, got %d", lastRight)
	}
	// Middle item should be centered
	if rects[1].X <= rects[0].X+rects[0].W {
		t.Fatalf("middle item should be after first item")
	}
}

func TestFlexJustifyCenter(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 100, H: 10}
	items := []FlexItem{
		{Basis: 20},
	}
	content := []Rect{{}}
	layout := FlexLayout{Direction: "row", JustifyContent: "center"}

	rects := CalculateFlexLayout(container, items, content, layout)
	// Item should be centered: (100-20)/2 = 40
	if rects[0].X != 40 {
		t.Fatalf("expected X=40, got %d", rects[0].X)
	}
}

func TestFlexAlignItemsCenter(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 100, H: 20}
	items := []FlexItem{
		{Basis: 30},
	}
	content := []Rect{{W: 30, H: 4}}
	layout := FlexLayout{Direction: "row", AlignItems: "center"}

	rects := CalculateFlexLayout(container, items, content, layout)
	// Cross size = 20, item height = 4, centered: (20-4)/2 = 8
	if rects[0].Y != 8 {
		t.Fatalf("expected Y=8, got %d", rects[0].Y)
	}
	if rects[0].H != 4 {
		t.Fatalf("expected H=4, got %d", rects[0].H)
	}
}

func TestFlexGap(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 100, H: 10}
	items := []FlexItem{
		{Basis: 10},
		{Basis: 10},
		{Basis: 10},
	}
	content := []Rect{{}, {}, {}}
	layout := FlexLayout{Direction: "row", Gap: 5}

	rects := CalculateFlexLayout(container, items, content, layout)
	if rects[0].X != 0 {
		t.Fatalf("expected X=0, got %d", rects[0].X)
	}
	if rects[1].X != 15 { // 10 + 5 gap
		t.Fatalf("expected X=15, got %d", rects[1].X)
	}
	if rects[2].X != 30 { // 15 + 10 + 5 gap
		t.Fatalf("expected X=30, got %d", rects[2].X)
	}
}

func TestFlexWrap(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 30, H: 20}
	items := []FlexItem{
		{Basis: 15},
		{Basis: 15},
		{Basis: 15},
	}
	content := []Rect{{W: 15, H: 5}, {W: 15, H: 5}, {W: 15, H: 5}}
	layout := FlexLayout{Direction: "row", Wrap: "wrap"}

	rects := CalculateFlexLayout(container, items, content, layout)
	// First line: items 0 and 1 (15+15=30 fits in 30)
	// Second line: item 2 (wraps)
	if rects[0].Y != 0 {
		t.Fatalf("expected item 0 Y=0, got %d", rects[0].Y)
	}
	if rects[1].Y != 0 {
		t.Fatalf("expected item 1 Y=0, got %d", rects[1].Y)
	}
	if rects[2].Y == 0 {
		t.Fatalf("expected item 2 to wrap to next line, but Y=0")
	}
	if rects[2].Y != 5 {
		t.Fatalf("expected item 2 Y=5 (after first line), got %d", rects[2].Y)
	}
}

func TestFlexNestedContainers(t *testing.T) {
	// Outer container
	outer := Rect{X: 0, Y: 0, W: 80, H: 24}
	outerItems := []FlexItem{
		{Grow: 1},
		{Grow: 3},
	}
	outerContent := []Rect{{W: 1, H: 24}, {W: 1, H: 24}}
	outerLayout := FlexLayout{Direction: "row"}

	outerRects := CalculateFlexLayout(outer, outerItems, outerContent, outerLayout)
	// Left panel: 1 + (80-2)*1/4 = 1 + 19 = 20
	// Right panel: 1 + (80-2)*3/4 = 1 + 58 = 59 (or 60 with rounding)
	if outerRects[0].W+outerRects[1].W != 80 {
		t.Fatalf("expected total 80, got %d", outerRects[0].W+outerRects[1].W)
	}

	// Inner flex in right panel
	innerItems := []FlexItem{
		{Basis: 5},
		{Basis: 5},
	}
	innerContent := []Rect{{W: 5, H: 1}, {W: 5, H: 1}}
	innerLayout := FlexLayout{Direction: "column"}
	innerRects := CalculateFlexLayout(outerRects[1], innerItems, innerContent, innerLayout)

	if innerRects[0].X != outerRects[1].X {
		t.Fatalf("inner item X should match parent X: %d vs %d", innerRects[0].X, outerRects[1].X)
	}
	if innerRects[0].Y != outerRects[1].Y {
		t.Fatalf("inner item Y should match parent Y: %d vs %d", innerRects[0].Y, outerRects[1].Y)
	}
}

func TestFlexBasis(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 100, H: 10}
	items := []FlexItem{
		{Basis: 40},
		{Basis: 60},
	}
	content := []Rect{{W: 5, H: 1}, {W: 5, H: 1}} // content sizes ignored when basis set
	layout := FlexLayout{Direction: "row"}

	rects := CalculateFlexLayout(container, items, content, layout)
	if rects[0].W != 40 {
		t.Fatalf("expected W=40 (basis), got %d", rects[0].W)
	}
	if rects[1].W != 60 {
		t.Fatalf("expected W=60 (basis), got %d", rects[1].W)
	}
}

func TestFlexEmptyItems(t *testing.T) {
	container := Rect{X: 0, Y: 0, W: 80, H: 24}
	result := CalculateFlexLayout(container, nil, nil, FlexLayout{})
	if result != nil {
		t.Fatalf("expected nil for empty items, got %v", result)
	}
}
