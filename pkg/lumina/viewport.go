package lumina

import (
	"math"

	"github.com/akzj/go-lua/pkg/lua"
)

// Viewport tracks scroll state for a scrollable container.
type Viewport struct {
	ScrollX  int // horizontal scroll offset (pixels/cells from left)
	ScrollY  int // vertical scroll offset (rows from top)
	ContentW int // total content width
	ContentH int // total content height (sum of children + gaps)
	ViewW    int // visible width (container content area width)
	ViewH    int // visible height (container content area height)

	// Smooth scrolling fields
	ScrollYF  float64 // floating-point scroll position
	VelocityY float64 // scroll velocity (rows/frame)
	Damping   float64 // friction coefficient (0.92 = smooth deceleration)
	Animating bool    // true while inertia is active

	ScrollDirty bool // true when scroll position changed, cleared after re-layout
}

// ScrollDown scrolls down by n rows, clamped to valid range.
func (v *Viewport) ScrollDown(n int) {
	old := v.ScrollY
	v.ScrollY += n
	v.clampScroll()
	if v.ScrollY != old {
		v.ScrollDirty = true
	}
}

// ScrollUp scrolls up by n rows, clamped to valid range.
func (v *Viewport) ScrollUp(n int) {
	old := v.ScrollY
	v.ScrollY -= n
	v.clampScroll()
	if v.ScrollY != old {
		v.ScrollDirty = true
	}
}

// ScrollLeft scrolls left by n columns.
func (v *Viewport) ScrollLeft(n int) {
	old := v.ScrollX
	v.ScrollX -= n
	v.clampScroll()
	if v.ScrollX != old {
		v.ScrollDirty = true
	}
}

// ScrollRight scrolls right by n columns.
func (v *Viewport) ScrollRight(n int) {
	old := v.ScrollX
	v.ScrollX += n
	v.clampScroll()
	if v.ScrollX != old {
		v.ScrollDirty = true
	}
}

// ScrollTo sets the vertical scroll offset to y, clamped to valid range.
func (v *Viewport) ScrollTo(y int) {
	old := v.ScrollY
	v.ScrollY = y
	v.clampScroll()
	if v.ScrollY != old {
		v.ScrollDirty = true
	}
}

// ScrollToBottom scrolls to the very bottom of the content.
func (v *Viewport) ScrollToBottom() {
	old := v.ScrollY
	v.ScrollY = v.maxScrollY()
	if v.ScrollY != old {
		v.ScrollDirty = true
	}
}

// ScrollToTop scrolls to the very top.
func (v *Viewport) ScrollToTop() {
	old := v.ScrollY
	oldX := v.ScrollX
	v.ScrollY = 0
	v.ScrollX = 0
	if v.ScrollY != old || v.ScrollX != oldX {
		v.ScrollDirty = true
	}
}

// EnsureVisible scrolls the minimum amount needed to make the region
// [y, y+h) fully visible within the viewport.
func (v *Viewport) EnsureVisible(y, h int) {
	if y < v.ScrollY {
		// Region starts above viewport — scroll up
		v.ScrollY = y
	} else if y+h > v.ScrollY+v.ViewH {
		// Region ends below viewport — scroll down
		v.ScrollY = y + h - v.ViewH
	}
	v.clampScroll()
}

// AtTop returns true if scrolled to the very top.
func (v *Viewport) AtTop() bool {
	return v.ScrollY <= 0
}

// AtBottom returns true if scrolled to the very bottom.
func (v *Viewport) AtBottom() bool {
	return v.ScrollY >= v.maxScrollY()
}

// ScrollPercent returns the scroll position as a fraction [0.0, 1.0].
// Returns 0 if content fits in the viewport (no scrolling needed).
func (v *Viewport) ScrollPercent() float64 {
	maxY := v.maxScrollY()
	if maxY <= 0 {
		return 0
	}
	return float64(v.ScrollY) / float64(maxY)
}

// NeedsScroll returns true if the content exceeds the viewport.
func (v *Viewport) NeedsScroll() bool {
	return v.ContentH > v.ViewH
}

// NeedsHScroll returns true if the content exceeds the viewport horizontally.
func (v *Viewport) NeedsHScroll() bool {
	return v.ContentW > v.ViewW
}

// VisibleRange returns the range of content rows visible in the viewport.
// Returns (startRow, endRow) where endRow is exclusive.
func (v *Viewport) VisibleRange() (int, int) {
	start := v.ScrollY
	end := v.ScrollY + v.ViewH
	if end > v.ContentH {
		end = v.ContentH
	}
	return start, end
}

// maxScrollY returns the maximum valid ScrollY value.
func (v *Viewport) maxScrollY() int {
	max := v.ContentH - v.ViewH
	if max < 0 {
		return 0
	}
	return max
}

// maxScrollX returns the maximum valid ScrollX value.
func (v *Viewport) maxScrollX() int {
	max := v.ContentW - v.ViewW
	if max < 0 {
		return 0
	}
	return max
}

// clampScroll ensures scroll offsets are within valid bounds.
func (v *Viewport) clampScroll() {
	if v.ScrollY < 0 {
		v.ScrollY = 0
	}
	if maxY := v.maxScrollY(); v.ScrollY > maxY {
		v.ScrollY = maxY
	}
	if v.ScrollX < 0 {
		v.ScrollX = 0
	}
	if maxX := v.maxScrollX(); v.ScrollX > maxX {
		v.ScrollX = maxX
	}
}

// -----------------------------------------------------------------------
// Smooth scrolling
// -----------------------------------------------------------------------

// DefaultDamping is the default friction coefficient for smooth scrolling.
const DefaultDamping = 0.85

// ScrollSmooth adds velocity for smooth inertial scrolling.
func (v *Viewport) ScrollSmooth(deltaY float64) {
	v.VelocityY += deltaY
	v.Animating = true
	if v.Damping == 0 {
		v.Damping = DefaultDamping
	}
}

// Tick advances the smooth scroll animation by one frame.
// Returns true if the viewport was updated (needs re-render).
func (v *Viewport) Tick() bool {
	if !v.Animating {
		return false
	}

	v.ScrollYF += v.VelocityY
	v.VelocityY *= v.Damping

	// Clamp to bounds
	if v.ScrollYF < 0 {
		v.ScrollYF = 0
		v.VelocityY = 0
	}
	maxScroll := float64(v.maxScrollY())
	if maxScroll < 0 {
		maxScroll = 0
	}
	if v.ScrollYF > maxScroll {
		v.ScrollYF = maxScroll
		v.VelocityY = 0
	}

	// Update integer offset for rendering
	v.ScrollY = int(math.Round(v.ScrollYF))

	// Stop when velocity is negligible
	if math.Abs(v.VelocityY) < 0.1 {
		v.VelocityY = 0
		v.Animating = false
	}

	return true
}

// SyncFloatFromInt synchronizes ScrollYF from the integer ScrollY.
// Call this after any instant (non-smooth) scroll operation.
func (v *Viewport) SyncFloatFromInt() {
	v.ScrollYF = float64(v.ScrollY)
	v.VelocityY = 0
	v.Animating = false
}

// ScrollbarThumb calculates the scrollbar thumb position and size
// for a track of the given height. Returns (thumbStart, thumbSize).
// thumbSize is at least 1 if scrolling is needed.
func (v *Viewport) ScrollbarThumb(trackH int) (int, int) {
	if !v.NeedsScroll() || trackH <= 0 {
		return 0, 0
	}

	// Thumb size is proportional to viewport/content ratio
	thumbSize := (v.ViewH * trackH) / v.ContentH
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > trackH {
		thumbSize = trackH
	}

	// Thumb position is proportional to scroll position
	maxThumbStart := trackH - thumbSize
	thumbStart := 0
	if maxY := v.maxScrollY(); maxY > 0 {
		thumbStart = (v.ScrollY * maxThumbStart) / maxY
	}
	if thumbStart < 0 {
		thumbStart = 0
	}
	if thumbStart > maxThumbStart {
		thumbStart = maxThumbStart
	}

	return thumbStart, thumbSize
}

// --- Viewport Registry ---
// Viewports persist across re-renders so scroll position is maintained.

// All access is from the main thread (actor model) — no mutex needed.
var viewportRegistry = make(map[string]*Viewport)

// GetViewport returns the viewport for the given VNode ID, creating one if needed.
func GetViewport(id string) *Viewport {
	vp, ok := viewportRegistry[id]
	if ok {
		return vp
	}

	// Create new viewport
	// Double-check after acquiring write lock
	if vp, ok := viewportRegistry[id]; ok {
		return vp
	}
	vp = &Viewport{}
	viewportRegistry[id] = vp
	return vp
}

// SetViewport stores a viewport for the given VNode ID.
func SetViewport(id string, vp *Viewport) {
	viewportRegistry[id] = vp
}

// RemoveViewport removes a viewport from the registry.
func RemoveViewport(id string) {
	delete(viewportRegistry, id)
}

// ClearViewports removes all viewports (useful for testing).
func ClearViewports() {
	viewportRegistry = make(map[string]*Viewport)
}

// AnyViewportScrollDirty returns true if any viewport has been scrolled
// since the last render. Used to force re-render even when VNode diff is empty.
func AnyViewportScrollDirty() bool {
	for _, vp := range viewportRegistry {
		if vp.ScrollDirty {
			return true
		}
	}
	return false
}

// ClearAllScrollDirty clears the ScrollDirty flag on all viewports.
// Called after a successful re-render.
func ClearAllScrollDirty() {
	for _, vp := range viewportRegistry {
		vp.ScrollDirty = false
	}
}

// TickAllViewports ticks smooth scrolling for all viewports.
// Returns true if any viewport was updated (needs re-render).
func TickAllViewports() bool {
	updated := false
	for _, vp := range viewportRegistry {
		if vp.Tick() {
			updated = true
		}
	}
	return updated
}

// ScrollViewport scrolls the viewport for the given ID by dy rows.
// Positive dy = scroll down, negative dy = scroll up.
// Returns false if no viewport exists for the ID.
func ScrollViewport(id string, dy int) bool {
	vp, ok := viewportRegistry[id]
	if !ok {
		return false
	}
	if dy > 0 {
		vp.ScrollDown(dy)
	} else if dy < 0 {
		vp.ScrollUp(-dy)
	}
	return true
}

// findScrollableAncestor walks up the VNode tree to find the nearest
// scrollable container that contains the given (x, y) point.
// Returns the VNode ID and viewport, or ("", nil) if none found.
func findScrollableVNode(root *VNode, x, y int) (string, *Viewport) {
	if root == nil {
		return "", nil
	}

	// Check if this node contains the point
	if x < root.X || x >= root.X+root.W || y < root.Y || y >= root.Y+root.H {
		return "", nil
	}

	// Check children first (deepest match wins)
	for _, child := range root.Children {
		if id, vp := findScrollableVNode(child, x, y); id != "" {
			return id, vp
		}
	}

	// Check if this node is scrollable
	if root.Style.Overflow == "scroll" {
		if id, ok := root.Props["id"].(string); ok && id != "" {
			vp := GetViewport(id)
			return id, vp
		}
	}

	return "", nil
}

// --- Lua API ---

// luaScrollTo implements lumina.scrollTo(id, y)
// Scrolls the viewport for the given container ID to the specified row.
func luaScrollTo(L *lua.State) int {
	id := L.CheckString(1)
	y, ok := L.ToInteger(2)
	if !ok {
		L.PushString("scrollTo: expected integer for y position")
		L.Error()
		return 0
	}

	vp := GetViewport(id)
	vp.ScrollTo(int(y))
	return 0
}

// luaScrollToBottom implements lumina.scrollToBottom(id)
// Scrolls the viewport for the given container ID to the bottom.
func luaScrollToBottom(L *lua.State) int {
	id := L.CheckString(1)
	vp := GetViewport(id)
	vp.ScrollToBottom()
	return 0
}

// luaScrollToTop implements lumina.scrollToTop(id)
// Scrolls the viewport for the given container ID to the top.
func luaScrollToTop(L *lua.State) int {
	id := L.CheckString(1)
	vp := GetViewport(id)
	vp.ScrollToTop()
	return 0
}

// luaScrollBy implements lumina.scrollBy(id, dy)
// Scrolls the viewport by dy rows (positive = down, negative = up).
func luaScrollBy(L *lua.State) int {
	id := L.CheckString(1)
	dy, ok := L.ToInteger(2)
	if !ok {
		L.PushString("scrollBy: expected integer for dy")
		L.Error()
		return 0
	}

	vp := GetViewport(id)
	if dy > 0 {
		vp.ScrollDown(int(dy))
	} else if dy < 0 {
		vp.ScrollUp(int(-dy))
	}

	// Mark all components dirty to trigger re-render with new scroll position
	for _, comp := range globalRegistry.components {
		comp.Dirty.Store(true)
	}

	return 0
}

// luaGetScrollInfo implements lumina.getScrollInfo(id) -> table
// Returns scroll state for the given container.
func luaGetScrollInfo(L *lua.State) int {
	id := L.CheckString(1)

	vp, ok := viewportRegistry[id]

	if !ok {
		L.PushNil()
		return 1
	}

	L.NewTableFrom(map[string]any{
		"scrollX":    int64(vp.ScrollX),
		"scrollY":    int64(vp.ScrollY),
		"contentW":   int64(vp.ContentW),
		"contentH":   int64(vp.ContentH),
		"viewW":      int64(vp.ViewW),
		"viewH":      int64(vp.ViewH),
		"atTop":      vp.AtTop(),
		"atBottom":   vp.AtBottom(),
		"needsScroll": vp.NeedsScroll(),
	})
	return 1
}
