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
	Height    int // panel height (bottom portion of screen)

	// Data sources
	tracker          *perf.Tracker
	components       []ComponentInfo
	nodeTree         []NodeInfo   // flattened node tree for Elements tab
	elementsScrollY  int          // scroll offset for Elements tab
	elementsPickArmed bool        // next mousedown above panel picks a node (Elements tab)
	elementsSelectedIdx int       // flat preorder index in nodeTree, or -1
	perfSnap         PerfSnapshot // frozen perf data for display

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

// elementsDetailReserved returns extra lines reserved below the tree when a node is selected.
func (p *Panel) elementsDetailReserved() int {
	if p.ActiveTab != TabElements {
		return 0
	}
	if p.elementsSelectedIdx < 0 || p.elementsSelectedIdx >= len(p.nodeTree) {
		return 0
	}
	return 7 // separator + up to 5 detail lines + gap
}

// elementsTreeVisibleLines is how many tree lines fit in the panel (excluding tab bar and detail block).
func (p *Panel) elementsTreeVisibleLines() int {
	lines := p.Height - 3 - p.elementsDetailReserved()
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
