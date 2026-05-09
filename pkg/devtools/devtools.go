// Package devtools provides a Chrome-style DevTools panel for Lumina v2.
// Press F12 to toggle. Shows Elements (VNode tree) and Perf (frame stats) tabs.
package devtools

import (
	"runtime"
	"time"

	"github.com/akzj/lumina/pkg/perf"
)

// Tab identifies a DevTools tab.
type Tab int

const (
	TabElements Tab = iota
	TabPerf
)

// Elements walk limits (DevTools snapshot). Deep enough for real apps; capped for safety.
const (
	ElementsWalkMaxDepth = 48
	ElementsWalkMaxNodes = 4000
)

// ElementsPanelOverheadLines is the number of rows consumed by the resize handle,
// tab bar, and blank separator before the tree content begins.
// Exported so app.go can compute the click-to-row mapping.
const ElementsPanelOverheadLines = 4

// elementsPanelOverheadLines is the internal alias used within the package.
const elementsPanelOverheadLines = ElementsPanelOverheadLines

// Panel anchor positions.
const (
	AnchorBottom = "bottom"
	AnchorRight  = "right"
)

// ResizeHandleThick is the number of rows (bottom) or columns (right) used by the resize handle.
const ResizeHandleThick = 1

// resizeHandleThick is the internal alias used within the package.
const resizeHandleThick = ResizeHandleThick

// Full selection block: separator + detail lines (see paintElementsTab).
const elementsDetailIdealLines = 7

// When a row is selected, keep at least this many lines for the tree before giving space to detail.
const elementsTreeMinLinesWhenDetail = 2

// ComponentInfo holds snapshot data about a registered component.
type ComponentInfo struct {
	ID     string
	Name   string
	X, Y   int
	W, H   int
	ZIndex int
}

// PerfSnapshot holds a frozen copy of perf data, captured before the devtools
// component renders so the devtools' own render does not pollute the numbers.
type PerfSnapshot struct {
	Last  perf.FrameStats
	Total perf.FrameStats
	FPS   int
}

// Panel is the DevTools panel state.
type Panel struct {
	Visible   bool
	ActiveTab Tab
	Width     int
	Height    int // panel height (AnchorBottom) or panel width (AnchorRight)

	// last known screen dimensions — updated by InitSizeForAnchor / UpdateResize
	screenW int
	screenH int

	// Position anchor
	Anchor string // AnchorBottom | AnchorRight

	// Resize drag state
	resizing         bool
	resizeStartMouse int // mouseY (bottom) or mouseX (right) at drag start
	resizeStartSize  int // panel Height (bottom) or Width (right) at drag start

	// Data sources
	tracker             *perf.Tracker
	components          []ComponentInfo
	nodeTree            []NodeInfo   // flattened node tree for Elements tab
	elementsScrollY     int          // scroll offset for Elements tab
	elementsPickArmed   bool         // next mousedown above panel picks a node (Elements tab)
	elementsSelectedIdx int          // flat preorder index in nodeTree, or -1
	perfSnap            PerfSnapshot // frozen perf data for display

	// FPS tracking (updated by TickFPS, called from event loop)
	fpsFrameCount int
	fpsLastTime   time.Time
	fps           int // smoothed FPS (EMA)

	// Runtime metrics
	LuaCPUTime  time.Duration
	LuaMemBytes uint64
	GoMemStats  runtime.MemStats
}

// NewPanel creates a new DevTools panel backed by the given perf tracker.
func NewPanel(tracker *perf.Tracker) *Panel {
	return &Panel{
		tracker:             tracker,
		fpsLastTime:         time.Now(),
		elementsSelectedIdx: -1,
		Anchor:              AnchorBottom,
	}
}

// Toggle flips the panel's visibility.
func (p *Panel) Toggle() {
	p.Visible = !p.Visible
}

// SetTab switches the active tab.
func (p *Panel) SetTab(tab Tab) {
	p.ActiveTab = tab
	if tab != TabElements {
		p.elementsPickArmed = false
	}
}

// FPS returns the current smoothed frames-per-second value.
func (p *Panel) FPS() int {
	return p.fps
}

// TickFPS should be called once per frame tick (from the event loop).
// It uses exponential moving average (EMA) to smooth the FPS measurement,
// updating every 300ms like the v1 implementation.
func (p *Panel) TickFPS() {
	p.fpsFrameCount++
	now := time.Now()
	elapsed := now.Sub(p.fpsLastTime)
	if elapsed >= 300*time.Millisecond {
		rawFPS := float64(p.fpsFrameCount) / elapsed.Seconds()
		if p.fps == 0 {
			p.fps = int(rawFPS + 0.5)
		} else {
			smoothed := 0.3*rawFPS + 0.7*float64(p.fps)
			p.fps = int(smoothed + 0.5)
		}
		p.fpsFrameCount = 0
		p.fpsLastTime = now
	}
}

// SnapshotPerf captures the current perf tracker state so the devtools render
// can display it without being affected by its own render/layout/paint cycle.
// Call this BEFORE marking the devtools component dirty.
func (p *Panel) SnapshotPerf() {
	p.perfSnap = PerfSnapshot{
		Last:  p.tracker.LastFrame(),
		Total: p.tracker.TotalStats(),
		FPS:   p.fps,
	}
}

// UpdateComponents replaces the component snapshot list.
func (p *Panel) UpdateComponents(infos []ComponentInfo) {
	p.components = infos
}

// Components returns the current component snapshot list.
func (p *Panel) Components() []ComponentInfo {
	return p.components
}

// NodeInfo holds snapshot data about a render node for the Elements tree view.
type NodeInfo struct {
	Type    string // "box", "hbox", "vbox", "text", "input", "textarea", "component"
	X, Y    int
	W, H    int
	BG      string
	FG      string
	Content string // for text nodes
	Depth   int    // indentation level

	// Identity & layout (for inspect detail)
	ID      string
	Key     string
	Path    string // preorder path e.g. "0/2/1"
	ScrollY int
	Overflow string
	Border   string
	Flex     int
	ComponentID      string // non-empty if node.Component != nil
	ComponentFactory string // Component.Type (Lua factory / Go widget name)
}

// UpdateNodeTree replaces the node tree snapshot for the Elements tab.
func (p *Panel) UpdateNodeTree(infos []NodeInfo) {
	p.nodeTree = infos
}

// NodeTree returns the current node tree snapshot.
func (p *Panel) NodeTree() []NodeInfo {
	return p.nodeTree
}

// elementsPanelContentLines is rows available under the tab bar for Elements (tree + optional detail).
func (p *Panel) elementsPanelContentLines() int {
	n := p.panelVisibleRows() - elementsPanelOverheadLines
	if n < 0 {
		return 0
	}
	return n
}

// elementsDetailReserved returns extra lines reserved below the tree when a node is selected.
// On short panels it shrinks so the tree keeps at least elementsTreeMinLinesWhenDetail rows.
func (p *Panel) elementsDetailReserved() int {
	if p.ActiveTab != TabElements {
		return 0
	}
	if p.elementsSelectedIdx < 0 || p.elementsSelectedIdx >= len(p.nodeTree) {
		return 0
	}
	avail := p.elementsPanelContentLines()
	if avail <= 0 {
		return 0
	}
	maxDetail := avail - elementsTreeMinLinesWhenDetail
	if maxDetail < 0 {
		maxDetail = 0
	}
	if maxDetail == 0 {
		return 0
	}
	if maxDetail > elementsDetailIdealLines {
		return elementsDetailIdealLines
	}
	return maxDetail
}

// elementsTreeVisibleLines is how many tree lines fit in the panel (excluding tab bar and detail block).
func (p *Panel) elementsTreeVisibleLines() int {
	lines := p.panelVisibleRows() - elementsPanelOverheadLines - p.elementsDetailReserved()
	if lines < 1 {
		lines = 1
	}
	return lines
}

func (p *Panel) clampElementsScroll() {
	maxScroll := len(p.nodeTree) - p.elementsTreeVisibleLines()
	if maxScroll < 0 {
		maxScroll = 0
	}
	if p.elementsScrollY > maxScroll {
		p.elementsScrollY = maxScroll
	}
	if p.elementsScrollY < 0 {
		p.elementsScrollY = 0
	}
}

func (p *Panel) ensureElementsScrollShowsSelection() {
	if p.elementsSelectedIdx < 0 || len(p.nodeTree) == 0 {
		return
	}
	vis := p.elementsTreeVisibleLines()
	if vis < 1 {
		return
	}
	if p.elementsScrollY > p.elementsSelectedIdx {
		p.elementsScrollY = p.elementsSelectedIdx
	}
	if p.elementsSelectedIdx >= p.elementsScrollY+vis {
		p.elementsScrollY = p.elementsSelectedIdx - vis + 1
	}
	p.clampElementsScroll()
}

// ArmElementsPick arms one-shot inspect: next mousedown above the panel picks a node.
func (p *Panel) ArmElementsPick() { p.elementsPickArmed = true }

// ClearElementsPickArm clears the inspect arm without selecting.
func (p *Panel) ClearElementsPickArm() { p.elementsPickArmed = false }

// ElementsPickArmed reports whether the next mousedown will run inspect pick.
func (p *Panel) ElementsPickArmed() bool { return p.elementsPickArmed }

// SetElementsSelection sets the selected flat-tree index and scrolls it into view.
func (p *Panel) SetElementsSelection(idx int) {
	p.elementsSelectedIdx = idx
	p.ensureElementsScrollShowsSelection()
}

// ClearElementsSelection clears the selected node.
func (p *Panel) ClearElementsSelection() {
	p.elementsSelectedIdx = -1
	p.clampElementsScroll()
}

// ElementsSelectedIdx returns the selected flat index or -1.
func (p *Panel) ElementsSelectedIdx() int { return p.elementsSelectedIdx }

// ElementsPageScrollLines returns scroll delta magnitude for PageUp/PageDown in the tree.
func (p *Panel) ElementsPageScrollLines() int { return p.elementsTreeVisibleLines() }

// ElementsTreeVisibleLines is how many tree rows fit (after tab bar and optional detail block).
func (p *Panel) ElementsTreeVisibleLines() int { return p.elementsTreeVisibleLines() }

// ElementsDetailReservedLines is the number of lines reserved below the tree for selection details.
func (p *Panel) ElementsDetailReservedLines() int { return p.elementsDetailReserved() }

// OnElementsTreeRebuilt clamps selection after a new snapshot (e.g. shallower tree).
func (p *Panel) OnElementsTreeRebuilt(newLen int) {
	if p.elementsSelectedIdx >= newLen {
		if newLen == 0 {
			p.elementsSelectedIdx = -1
		} else {
			p.elementsSelectedIdx = newLen - 1
		}
	}
	p.ensureElementsScrollShowsSelection()
	p.clampElementsScroll()
}

// ResetElementsInspect clears pick arm and selection (e.g. when closing DevTools).
func (p *Panel) ResetElementsInspect() {
	p.elementsPickArmed = false
	p.elementsSelectedIdx = -1
}

// ScrollElements adjusts the Elements tab scroll offset by delta lines.
func (p *Panel) ScrollElements(delta int) {
	p.elementsScrollY += delta
	p.clampElementsScroll()
}

// ElementsScrollY returns the current Elements tab scroll offset.
func (p *Panel) ElementsScrollY() int {
	return p.elementsScrollY
}

// Snapshot returns the frozen perf snapshot for display.
func (p *Panel) Snapshot() PerfSnapshot {
	return p.perfSnap
}

// UpdateLuaMetrics updates Lua runtime metrics.
func (p *Panel) UpdateLuaMetrics(cpuTime time.Duration, memBytes uint64) {
	p.LuaCPUTime = cpuTime
	p.LuaMemBytes = memBytes
}

// ---------------------------------------------------------------------------
// Geometry helpers
// ---------------------------------------------------------------------------

// InitSizeForAnchor initialises panel Height (bottom) or Width (right) when zero,
// and records the current screen dimensions for content-line calculations.
func (p *Panel) InitSizeForAnchor(screenW, screenH int) {
	p.screenW = screenW
	p.screenH = screenH
	if p.Anchor == AnchorRight {
		if p.Width <= 0 {
			p.Width = screenW * 4 / 10
			if p.Width < 20 {
				p.Width = 20
			}
		}
	} else {
		if p.Height <= 0 {
			p.Height = screenH * 4 / 10
			if p.Height < 8 {
				p.Height = 8
			}
		}
	}
}

// panelVisibleRows returns the number of terminal rows the panel occupies.
func (p *Panel) panelVisibleRows() int {
	if p.Anchor == AnchorRight {
		if p.screenH > 0 {
			return p.screenH
		}
		return p.Height // fallback
	}
	return p.Height
}

// PanelRect returns the panel's screen rectangle (x, y, w, h).
func (p *Panel) PanelRect(screenW, screenH int) (x, y, w, h int) {
	if p.Anchor == AnchorRight {
		pw := p.Width
		if pw <= 0 {
			pw = screenW * 4 / 10
		}
		return screenW - pw, 0, pw, screenH
	}
	ph := p.Height
	if ph <= 0 {
		ph = screenH * 4 / 10
	}
	return 0, screenH - ph, screenW, ph
}

// AppRect returns the app content area dimensions when the panel is docked.
// AnchorBottom shrinks height; AnchorRight shrinks width.
func (p *Panel) AppRect(screenW, screenH int) (appW, appH int) {
	_, _, pw, ph := p.PanelRect(screenW, screenH)
	if p.Anchor == AnchorRight {
		return screenW - pw, screenH
	}
	return screenW, screenH - ph
}

// ContainsPoint reports whether (mx, my) lies inside the panel rectangle.
func (p *Panel) ContainsPoint(mx, my, screenW, screenH int) bool {
	px, py, pw, ph := p.PanelRect(screenW, screenH)
	return mx >= px && mx < px+pw && my >= py && my < py+ph
}

// HitResizeBorder reports whether (mx, my) is on the resize drag handle.
// Bottom anchor: top row of the panel.
// Right anchor: left column of the panel.
func (p *Panel) HitResizeBorder(mx, my, screenW, screenH int) bool {
	px, py, pw, ph := p.PanelRect(screenW, screenH)
	if p.Anchor == AnchorRight {
		return mx == px && my >= py && my < py+ph
	}
	return my == py && mx >= px && mx < px+pw
}

// HitTabBar reports whether (mx, my) is on the tab bar row.
func (p *Panel) HitTabBar(mx, my, screenW, screenH int) bool {
	px, py, pw, _ := p.PanelRect(screenW, screenH)
	tabRow := py + resizeHandleThick
	if p.Anchor == AnchorRight {
		tabRow = py
		return my == tabRow && mx >= px+resizeHandleThick && mx < px+pw
	}
	return my == tabRow && mx >= px && mx < px+pw
}

// tabBarElementsStart is the column (relative to tab bar start) where "Elements" label begins.
const tabBarElementsStart = 1  // after leading space
const tabBarElementsEnd = 10
const tabBarPerfStart = 12
const tabBarPerfEnd = 17

// TabAtPoint returns which tab was clicked, or -1 if neither.
func (p *Panel) TabAtPoint(mx, my, screenW, screenH int) Tab {
	if !p.HitTabBar(mx, my, screenW, screenH) {
		return -1
	}
	px, _, _, _ := p.PanelRect(screenW, screenH)
	tabStartX := px
	if p.Anchor == AnchorRight {
		tabStartX = px + resizeHandleThick
	}
	rel := mx - tabStartX
	if rel >= tabBarElementsStart && rel < tabBarElementsEnd {
		return TabElements
	}
	if rel >= tabBarPerfStart && rel < tabBarPerfEnd {
		return TabPerf
	}
	return -1
}

// ---------------------------------------------------------------------------
// Resize drag
// ---------------------------------------------------------------------------

// StartResize begins a resize drag. mouseCoord is Y (bottom) or X (right).
func (p *Panel) StartResize(mouseCoord int) {
	p.resizing = true
	p.resizeStartMouse = mouseCoord
	if p.Anchor == AnchorRight {
		p.resizeStartSize = p.Width
	} else {
		p.resizeStartSize = p.Height
	}
}

// UpdateResize updates panel size while dragging.
func (p *Panel) UpdateResize(mouseCoord, screenW, screenH int) {
	if !p.resizing {
		return
	}
	p.screenW = screenW
	p.screenH = screenH
	delta := p.resizeStartMouse - mouseCoord
	if p.Anchor == AnchorRight {
		// dragging left border right → delta negative → panel narrows
		p.Width = clampInt(p.resizeStartSize+delta, 20, screenW-10)
	} else {
		// dragging top border down → delta negative → panel shrinks
		p.Height = clampInt(p.resizeStartSize+delta, 8, screenH-3)
	}
}

// EndResize finishes the resize drag.
func (p *Panel) EndResize() { p.resizing = false }

// IsResizing reports whether a resize drag is in progress.
func (p *Panel) IsResizing() bool { return p.resizing }

// ---------------------------------------------------------------------------
// Anchor management
// ---------------------------------------------------------------------------

// CycleAnchor toggles between AnchorBottom and AnchorRight.
func (p *Panel) CycleAnchor(screenW, screenH int) {
	if p.Anchor == AnchorRight {
		p.Anchor = AnchorBottom
	} else {
		p.Anchor = AnchorRight
	}
	p.InitSizeForAnchor(screenW, screenH)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
