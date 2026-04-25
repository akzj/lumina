// Package lumina — window manager for multi-window compositing.
package lumina

import (
	"sort"
	"sync"
)

// Window represents a managed window with position, size, and content.
type Window struct {
	ID        string
	Title     string
	X, Y      int // position on screen
	W, H      int // dimensions (including chrome)
	MinW, MinH int // minimum size
	ZIndex    int
	Visible   bool
	Focused   bool
	Minimized bool
	Maximized bool
	// Pre-maximize geometry for restore
	RestoreX, RestoreY, RestoreW, RestoreH int

	VNode  *VNode // content VNode rendered by Lua component
	CompID string // component ID that renders this window's content
}

// dragState tracks an in-progress window drag.
type dragState struct {
	active   bool
	windowID string
	offsetX  int // cursor offset from window top-left
	offsetY  int
}

// resizeState tracks an in-progress window resize.
type resizeState struct {
	active   bool
	windowID string
	startX   int // cursor position at start
	startY   int
	startW   int // window size at start
	startH   int
}

// WindowManager manages multiple windows with z-ordering, drag, and resize.
type WindowManager struct {
	mu      sync.RWMutex
	windows map[string]*Window
	order   []string // z-order: first = back, last = front
	nextZ   int
	screenW int
	screenH int
	drag    dragState
	resize  resizeState
}

// NewWindowManager creates a new WindowManager.
func NewWindowManager(screenW, screenH int) *WindowManager {
	return &WindowManager{
		windows: make(map[string]*Window),
		screenW: screenW,
		screenH: screenH,
		nextZ:   100, // start above normal overlays
	}
}

// globalWindowManager is the singleton window manager.
var globalWindowManager = NewWindowManager(80, 24)

// SetScreenSize updates the screen dimensions.
func (wm *WindowManager) SetScreenSize(w, h int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.screenW = w
	wm.screenH = h
}

// CreateWindow creates and registers a new window.
func (wm *WindowManager) CreateWindow(id, title string, x, y, w, h int) *Window {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if w < 10 {
		w = 10
	}
	if h < 4 {
		h = 4 // minimum: title bar + 2 content rows + border
	}

	win := &Window{
		ID:      id,
		Title:   title,
		X:       x,
		Y:       y,
		W:       w,
		H:       h,
		MinW:    10,
		MinH:    4,
		ZIndex:  wm.nextZ,
		Visible: true,
		Focused: true,
	}
	wm.nextZ++

	// Unfocus all other windows
	for _, other := range wm.windows {
		other.Focused = false
	}

	wm.windows[id] = win
	wm.order = append(wm.order, id)
	return win
}

// CloseWindow removes a window.
func (wm *WindowManager) CloseWindow(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.windows, id)
	for i, oid := range wm.order {
		if oid == id {
			wm.order = append(wm.order[:i], wm.order[i+1:]...)
			break
		}
	}
	// Focus the next top window
	if len(wm.order) > 0 {
		topID := wm.order[len(wm.order)-1]
		if w, ok := wm.windows[topID]; ok {
			w.Focused = true
		}
	}
}

// FocusWindow brings a window to the front and marks it focused.
func (wm *WindowManager) FocusWindow(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	win, ok := wm.windows[id]
	if !ok {
		return
	}

	// Unfocus all
	for _, w := range wm.windows {
		w.Focused = false
	}
	win.Focused = true
	win.ZIndex = wm.nextZ
	wm.nextZ++

	// Move to end of order
	for i, oid := range wm.order {
		if oid == id {
			wm.order = append(wm.order[:i], wm.order[i+1:]...)
			break
		}
	}
	wm.order = append(wm.order, id)
}

// MoveWindow updates a window's position.
func (wm *WindowManager) MoveWindow(id string, x, y int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if win, ok := wm.windows[id]; ok {
		win.X = x
		win.Y = y
	}
}

// ResizeWindow updates a window's dimensions, respecting minimum size.
func (wm *WindowManager) ResizeWindow(id string, w, h int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if win, ok := wm.windows[id]; ok {
		if w < win.MinW {
			w = win.MinW
		}
		if h < win.MinH {
			h = win.MinH
		}
		win.W = w
		win.H = h
	}
}

// MinimizeWindow hides a window (minimized state).
func (wm *WindowManager) MinimizeWindow(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if win, ok := wm.windows[id]; ok {
		win.Minimized = true
		win.Focused = false
		// Focus next visible window
		for i := len(wm.order) - 1; i >= 0; i-- {
			if w, ok2 := wm.windows[wm.order[i]]; ok2 && !w.Minimized && w.ID != id {
				w.Focused = true
				break
			}
		}
	}
}

// MaximizeWindow fills the screen (saves current geometry for restore).
func (wm *WindowManager) MaximizeWindow(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if win, ok := wm.windows[id]; ok {
		if win.Maximized {
			return // already maximized
		}
		// Save current geometry
		win.RestoreX = win.X
		win.RestoreY = win.Y
		win.RestoreW = win.W
		win.RestoreH = win.H
		// Fill screen
		win.X = 0
		win.Y = 0
		win.W = wm.screenW
		win.H = wm.screenH
		win.Maximized = true
	}
}

// RestoreWindow un-maximizes or un-minimizes a window.
func (wm *WindowManager) RestoreWindow(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if win, ok := wm.windows[id]; ok {
		if win.Maximized {
			win.X = win.RestoreX
			win.Y = win.RestoreY
			win.W = win.RestoreW
			win.H = win.RestoreH
			win.Maximized = false
		}
		if win.Minimized {
			win.Minimized = false
		}
	}
}

// GetVisible returns all visible, non-minimized windows sorted by ZIndex.
func (wm *WindowManager) GetVisible() []*Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	var result []*Window
	for _, id := range wm.order {
		if win, ok := wm.windows[id]; ok && win.Visible && !win.Minimized {
			result = append(result, win)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ZIndex < result[j].ZIndex
	})
	return result
}

// GetFocused returns the currently focused window, or nil.
func (wm *WindowManager) GetFocused() *Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	for _, win := range wm.windows {
		if win.Focused {
			return win
		}
	}
	return nil
}

// GetWindow returns a window by ID.
func (wm *WindowManager) GetWindow(id string) *Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.windows[id]
}

// WindowAtPoint returns the topmost visible window containing the point (x, y).
func (wm *WindowManager) WindowAtPoint(x, y int) *Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	// Iterate in reverse z-order (front to back)
	for i := len(wm.order) - 1; i >= 0; i-- {
		win, ok := wm.windows[wm.order[i]]
		if !ok || !win.Visible || win.Minimized {
			continue
		}
		if x >= win.X && x < win.X+win.W && y >= win.Y && y < win.Y+win.H {
			return win
		}
	}
	return nil
}

// Count returns the total number of windows.
func (wm *WindowManager) Count() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.windows)
}

// --- Drag support ---

// StartDrag begins dragging a window from the title bar.
func (wm *WindowManager) StartDrag(windowID string, offsetX, offsetY int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.drag = dragState{
		active:   true,
		windowID: windowID,
		offsetX:  offsetX,
		offsetY:  offsetY,
	}
}

// UpdateDrag moves the dragged window to follow the cursor.
func (wm *WindowManager) UpdateDrag(cursorX, cursorY int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if !wm.drag.active {
		return
	}
	if win, ok := wm.windows[wm.drag.windowID]; ok {
		win.X = cursorX - wm.drag.offsetX
		win.Y = cursorY - wm.drag.offsetY
	}
}

// StopDrag ends the current drag operation.
func (wm *WindowManager) StopDrag() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.drag.active = false
}

// IsDragging returns true if a drag is in progress.
func (wm *WindowManager) IsDragging() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.drag.active
}

// --- Resize support ---

// StartResize begins resizing a window.
func (wm *WindowManager) StartResize(windowID string, cursorX, cursorY int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	win, ok := wm.windows[windowID]
	if !ok {
		return
	}
	wm.resize = resizeState{
		active:   true,
		windowID: windowID,
		startX:   cursorX,
		startY:   cursorY,
		startW:   win.W,
		startH:   win.H,
	}
}

// UpdateResize adjusts the window size based on cursor movement.
func (wm *WindowManager) UpdateResize(cursorX, cursorY int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if !wm.resize.active {
		return
	}
	win, ok := wm.windows[wm.resize.windowID]
	if !ok {
		return
	}
	newW := wm.resize.startW + (cursorX - wm.resize.startX)
	newH := wm.resize.startH + (cursorY - wm.resize.startY)
	if newW < win.MinW {
		newW = win.MinW
	}
	if newH < win.MinH {
		newH = win.MinH
	}
	win.W = newW
	win.H = newH
}

// StopResize ends the current resize operation.
func (wm *WindowManager) StopResize() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.resize.active = false
}

// IsResizing returns true if a resize is in progress.
func (wm *WindowManager) IsResizing() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.resize.active
}

// --- Tiling layouts ---

// TileHorizontal arranges all visible windows side-by-side.
func (wm *WindowManager) TileHorizontal() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	visible := wm.visibleWindows()
	if len(visible) == 0 {
		return
	}
	w := wm.screenW / len(visible)
	for i, win := range visible {
		win.X = i * w
		win.Y = 0
		win.W = w
		win.H = wm.screenH
		win.Maximized = false
		win.Minimized = false
	}
}

// TileVertical arranges all visible windows stacked top-to-bottom.
func (wm *WindowManager) TileVertical() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	visible := wm.visibleWindows()
	if len(visible) == 0 {
		return
	}
	h := wm.screenH / len(visible)
	for i, win := range visible {
		win.X = 0
		win.Y = i * h
		win.W = wm.screenW
		win.H = h
		win.Maximized = false
		win.Minimized = false
	}
}

// TileGrid arranges visible windows in a 2xN grid.
func (wm *WindowManager) TileGrid() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	visible := wm.visibleWindows()
	n := len(visible)
	if n == 0 {
		return
	}
	cols := 2
	if n == 1 {
		cols = 1
	}
	rows := (n + cols - 1) / cols
	cellW := wm.screenW / cols
	cellH := wm.screenH / rows
	for i, win := range visible {
		col := i % cols
		row := i / cols
		win.X = col * cellW
		win.Y = row * cellH
		win.W = cellW
		win.H = cellH
		win.Maximized = false
		win.Minimized = false
	}
}

// visibleWindows returns non-minimized windows in z-order (must be called with lock held).
func (wm *WindowManager) visibleWindows() []*Window {
	var result []*Window
	for _, id := range wm.order {
		if win, ok := wm.windows[id]; ok && win.Visible && !win.Minimized {
			result = append(result, win)
		}
	}
	return result
}

// BuildWindowVNode creates a VNode for the window chrome (border + title bar + content).
func BuildWindowVNode(win *Window) *VNode {
	titleBg := "#333333"
	if win.Focused {
		titleBg = "#3B82F6"
	}

	// Title text (left-aligned, flex=1)
	titleText := &VNode{
		Type:    "text",
		Content: " " + win.Title,
		Props: map[string]any{
			"style": map[string]any{
				"flex":       1,
				"bold":       true,
				"foreground": "#FFFFFF",
			},
		},
	}

	// Window control buttons: _ □ X
	controls := &VNode{
		Type:    "text",
		Content: " _ □ X ",
		Props: map[string]any{
			"id": win.ID + "-controls",
			"style": map[string]any{
				"foreground": "#FFFFFF",
			},
		},
	}

	// Title bar (1 row)
	titleBar := &VNode{
		Type: "hbox",
		Props: map[string]any{
			"style": map[string]any{
				"height":     1,
				"background": titleBg,
				"foreground": "#FFFFFF",
			},
		},
		Children: []*VNode{titleText, controls},
	}

	// Content area — use the window's VNode or a placeholder
	content := win.VNode
	if content == nil {
		content = &VNode{
			Type:    "text",
			Content: "",
		}
	}

	// Wrap content with flex=1 to fill remaining space
	contentWrap := &VNode{
		Type: "vbox",
		Props: map[string]any{
			"style": map[string]any{
				"flex": 1,
			},
		},
		Children: []*VNode{content},
	}

	// Window container with border
	return &VNode{
		Type: "vbox",
		Props: map[string]any{
			"style": map[string]any{
				"border":     "rounded",
				"width":      win.W,
				"height":     win.H,
				"background": "#1a1a2e",
				"foreground": "#FFFFFF",
			},
		},
		Children: []*VNode{titleBar, contentWrap},
	}
}

// Clear removes all windows.
func (wm *WindowManager) Clear() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.windows = make(map[string]*Window)
	wm.order = nil
	wm.drag.active = false
	wm.resize.active = false
}
