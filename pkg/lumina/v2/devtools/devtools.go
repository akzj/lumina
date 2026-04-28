// Package devtools provides a Chrome-style DevTools panel for Lumina v2.
// Press F12 to toggle. Shows Elements (VNode tree) and Perf (frame stats) tabs.
package devtools

import (
	"runtime"
	"time"

	"github.com/akzj/lumina/pkg/lumina/v2/perf"
)

// Tab identifies a DevTools tab.
type Tab int

const (
	TabElements Tab = iota
	TabPerf
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
		tracker:     tracker,
		fpsLastTime: time.Now(),
	}
}

// Toggle flips the panel's visibility.
func (p *Panel) Toggle() {
	p.Visible = !p.Visible
}

// SetTab switches the active tab.
func (p *Panel) SetTab(tab Tab) {
	p.ActiveTab = tab
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
}

// UpdateNodeTree replaces the node tree snapshot for the Elements tab.
func (p *Panel) UpdateNodeTree(infos []NodeInfo) {
	p.nodeTree = infos
}

// NodeTree returns the current node tree snapshot.
func (p *Panel) NodeTree() []NodeInfo {
	return p.nodeTree
}

// ScrollElements adjusts the Elements tab scroll offset by delta lines.
func (p *Panel) ScrollElements(delta int) {
	p.elementsScrollY += delta
	if p.elementsScrollY < 0 {
		p.elementsScrollY = 0
	}
	maxScroll := len(p.nodeTree) - (p.Height - 3) // 3 for tab bar + margins
	if maxScroll < 0 {
		maxScroll = 0
	}
	if p.elementsScrollY > maxScroll {
		p.elementsScrollY = maxScroll
	}
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
