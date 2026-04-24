package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestGridFixedColumns(t *testing.T) {
	layout := GridLayout{
		Columns: []GridTrack{
			{Type: "fixed", Value: 10},
			{Type: "fixed", Value: 20},
			{Type: "fixed", Value: 10},
		},
	}
	items := []GridItem{{}, {}, {}}
	rects := CalculateGridLayout(40, 10, layout, items)
	if len(rects) != 3 {
		t.Fatalf("expected 3 rects, got %d", len(rects))
	}
	if rects[0].Width != 10 {
		t.Fatalf("expected width 10, got %d", rects[0].Width)
	}
	if rects[1].Width != 20 {
		t.Fatalf("expected width 20, got %d", rects[1].Width)
	}
	if rects[1].X != 10 {
		t.Fatalf("expected x=10, got %d", rects[1].X)
	}
}

func TestGridFrUnits(t *testing.T) {
	layout := GridLayout{
		Columns: []GridTrack{
			{Type: "fr", Value: 1},
			{Type: "fr", Value: 2},
			{Type: "fr", Value: 1},
		},
	}
	items := []GridItem{{}, {}, {}}
	rects := CalculateGridLayout(40, 10, layout, items)
	if len(rects) != 3 {
		t.Fatalf("expected 3 rects, got %d", len(rects))
	}
	// 40 / 4 fr = 10 per fr
	if rects[0].Width != 10 {
		t.Fatalf("expected width 10, got %d", rects[0].Width)
	}
	if rects[1].Width != 20 {
		t.Fatalf("expected width 20, got %d", rects[1].Width)
	}
}

func TestGridMixedFixedFr(t *testing.T) {
	layout := GridLayout{
		Columns: []GridTrack{
			{Type: "fixed", Value: 10},
			{Type: "fr", Value: 1},
			{Type: "fr", Value: 1},
		},
	}
	items := []GridItem{{}, {}, {}}
	rects := CalculateGridLayout(30, 10, layout, items)
	// 30 - 10 fixed = 20 for 2fr = 10 each
	if rects[0].Width != 10 {
		t.Fatalf("expected fixed width 10, got %d", rects[0].Width)
	}
	if rects[1].Width != 10 {
		t.Fatalf("expected fr width 10, got %d", rects[1].Width)
	}
}

func TestGridColumnSpan(t *testing.T) {
	layout := GridLayout{
		Columns: []GridTrack{
			{Type: "fr", Value: 1},
			{Type: "fr", Value: 1},
			{Type: "fr", Value: 1},
		},
	}
	items := []GridItem{
		{ColumnStart: 1, ColumnEnd: 3, RowStart: 1}, // spans 2 cols
		{},                                            // auto-placed
	}
	rects := CalculateGridLayout(30, 10, layout, items)
	if rects[0].Width != 20 {
		t.Fatalf("expected span width 20, got %d", rects[0].Width)
	}
}

func TestGridRowSpan(t *testing.T) {
	layout := GridLayout{
		Columns: []GridTrack{
			{Type: "fr", Value: 1},
			{Type: "fr", Value: 1},
		},
		Rows: []GridTrack{
			{Type: "fixed", Value: 5},
			{Type: "fixed", Value: 5},
		},
	}
	items := []GridItem{
		{ColumnStart: 1, RowStart: 1, RowSpan: 2}, // spans 2 rows
		{},
	}
	rects := CalculateGridLayout(20, 10, layout, items)
	if rects[0].Height != 10 {
		t.Fatalf("expected span height 10, got %d", rects[0].Height)
	}
}

func TestGridGap(t *testing.T) {
	layout := GridLayout{
		Columns: []GridTrack{
			{Type: "fr", Value: 1},
			{Type: "fr", Value: 1},
		},
		Gap: 2,
	}
	items := []GridItem{{}, {}}
	rects := CalculateGridLayout(22, 10, layout, items)
	// 22 - 2 gap = 20, 20/2 = 10 each
	if rects[0].Width != 10 {
		t.Fatalf("expected width 10, got %d", rects[0].Width)
	}
	if rects[1].X != 12 {
		t.Fatalf("expected x=12 (10+2gap), got %d", rects[1].X)
	}
}

func TestGridAutoPlacement(t *testing.T) {
	layout := GridLayout{
		Columns: []GridTrack{
			{Type: "fr", Value: 1},
			{Type: "fr", Value: 1},
			{Type: "fr", Value: 1},
		},
	}
	// 5 items in 3 columns = 2 rows
	items := []GridItem{{}, {}, {}, {}, {}}
	rects := CalculateGridLayout(30, 10, layout, items)
	if len(rects) != 5 {
		t.Fatalf("expected 5 rects, got %d", len(rects))
	}
	// Item 4 (index 3) should be on row 2
	if rects[3].Y != rects[0].Y {
		// Actually row 2 starts after row 1
	}
	// Item 3 should be in column 0, row 1
	if rects[3].X != 0 {
		t.Fatalf("expected item 4 at x=0, got %d", rects[3].X)
	}
}

// Virtual Scroll tests

func TestVirtualListVisibleRange(t *testing.T) {
	vl := NewVirtualList(100, 1)
	vl.SetBuffer(0)
	start, end := vl.VisibleRange(10)
	if start != 0 || end != 10 {
		t.Fatalf("expected 0-10, got %d-%d", start, end)
	}
}

func TestVirtualListScrollUpdatesRange(t *testing.T) {
	vl := NewVirtualList(100, 1)
	vl.SetBuffer(0)
	vl.ScrollTo(50)
	start, end := vl.VisibleRange(10)
	if start != 50 || end != 60 {
		t.Fatalf("expected 50-60, got %d-%d", start, end)
	}
}

func TestVirtualListBuffer(t *testing.T) {
	vl := NewVirtualList(100, 1)
	vl.SetBuffer(5)
	vl.ScrollTo(50)
	start, end := vl.VisibleRange(10)
	if start != 45 {
		t.Fatalf("expected start 45, got %d", start)
	}
	if end != 65 {
		t.Fatalf("expected end 65, got %d", end)
	}
}

func TestVirtualListLargeDataset(t *testing.T) {
	vl := NewVirtualList(1000000, 1)
	vl.SetBuffer(3)
	vl.ScrollTo(500000)
	start, end := vl.VisibleRange(20)
	if start != 499997 {
		t.Fatalf("expected start 499997, got %d", start)
	}
	if end != 500023 {
		t.Fatalf("expected end 500023, got %d", end)
	}
}

func TestVirtualListItemOffset(t *testing.T) {
	vl := NewVirtualList(100, 3)
	offset := vl.ItemOffset(10)
	if offset != 30 {
		t.Fatalf("expected offset 30, got %d", offset)
	}
}

func TestVirtualListTotalHeight(t *testing.T) {
	vl := NewVirtualList(100, 2)
	if vl.TotalHeight() != 200 {
		t.Fatalf("expected 200, got %d", vl.TotalHeight())
	}
}

func TestLuaVirtualListAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local vl = lumina.createVirtualList({
			totalItems = 1000,
			itemHeight = 1,
			buffer = 3,
		})
		assert(vl ~= nil, "virtual list should exist")

		local start, stop = vl.visibleRange(20)
		assert(start == 0, "start should be 0, got " .. tostring(start))
		assert(stop == 23, "end should be 23, got " .. tostring(stop))

		vl.scrollTo(500)
		local s2, e2 = vl.visibleRange(20)
		assert(s2 == 497, "start should be 497, got " .. tostring(s2))
		assert(e2 == 523, "end should be 523, got " .. tostring(e2))

		local h = vl.totalHeight()
		assert(h == 1000, "total height should be 1000, got " .. tostring(h))

		local off = vl.itemOffset(10)
		assert(off == 10, "offset should be 10, got " .. tostring(off))
	`)
	if err != nil {
		t.Fatalf("Lua VirtualList: %v", err)
	}
}
