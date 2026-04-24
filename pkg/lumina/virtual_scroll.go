package lumina

// VirtualList manages virtual scrolling for large lists.
// Only items within the visible viewport (plus buffer) are rendered.
type VirtualList struct {
	TotalItems   int
	ItemHeight   int // height of each item in rows
	ScrollOffset int // current scroll position in rows
	Buffer       int // extra items to render above/below viewport
}

// NewVirtualList creates a new VirtualList.
func NewVirtualList(totalItems, itemHeight int) *VirtualList {
	if itemHeight < 1 {
		itemHeight = 1
	}
	return &VirtualList{
		TotalItems: totalItems,
		ItemHeight: itemHeight,
		Buffer:     3,
	}
}

// VisibleRange returns the start and end indices of items to render.
func (vl *VirtualList) VisibleRange(viewportHeight int) (start, end int) {
	if vl.TotalItems == 0 || viewportHeight <= 0 {
		return 0, 0
	}

	// How many items fit in the viewport
	itemsPerPage := viewportHeight / vl.ItemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	// First visible item
	firstVisible := vl.ScrollOffset / vl.ItemHeight
	if firstVisible < 0 {
		firstVisible = 0
	}

	start = firstVisible - vl.Buffer
	if start < 0 {
		start = 0
	}

	end = firstVisible + itemsPerPage + vl.Buffer
	if end > vl.TotalItems {
		end = vl.TotalItems
	}

	return start, end
}

// ItemOffset returns the Y offset for a given item index.
func (vl *VirtualList) ItemOffset(index int) int {
	return index * vl.ItemHeight
}

// TotalHeight returns the total scrollable height.
func (vl *VirtualList) TotalHeight() int {
	return vl.TotalItems * vl.ItemHeight
}

// ScrollTo sets the scroll offset to show a specific item.
func (vl *VirtualList) ScrollTo(index int) {
	if index < 0 {
		index = 0
	}
	if index >= vl.TotalItems {
		index = vl.TotalItems - 1
	}
	vl.ScrollOffset = index * vl.ItemHeight
}

// ScrollBy adjusts the scroll offset by delta rows.
func (vl *VirtualList) ScrollBy(delta int) {
	vl.ScrollOffset += delta
	if vl.ScrollOffset < 0 {
		vl.ScrollOffset = 0
	}
	maxOffset := vl.TotalHeight()
	if vl.ScrollOffset > maxOffset {
		vl.ScrollOffset = maxOffset
	}
}

// SetBuffer sets the number of buffer items.
func (vl *VirtualList) SetBuffer(buffer int) {
	if buffer < 0 {
		buffer = 0
	}
	vl.Buffer = buffer
}

// IsVisible returns whether an item at the given index is in the visible range.
func (vl *VirtualList) IsVisible(index, viewportHeight int) bool {
	start, end := vl.VisibleRange(viewportHeight)
	return index >= start && index < end
}

// PageUp scrolls up by one viewport height.
func (vl *VirtualList) PageUp(viewportHeight int) {
	vl.ScrollBy(-viewportHeight)
}

// PageDown scrolls down by one viewport height.
func (vl *VirtualList) PageDown(viewportHeight int) {
	vl.ScrollBy(viewportHeight)
}
