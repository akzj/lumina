// Package devtools provides a Chrome-style DevTools panel for Lumina v2.
// Press F12 to toggle. Shows Elements (VNode tree) and Perf (frame stats) tabs.
package devtools

import (
	"runtime"
	"time"

	"github.com/akzj/lumina/pkg/lumina/v2/layout"
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
	ID        string
	Name      string
	X, Y      int
	W, H      int
	ZIndex    int
	VNodeTree *layout.VNode
}

// Panel is the DevTools panel state.
type Panel struct {
	Visible   bool
	ActiveTab Tab
	Width     int
	Height    int // panel height (bottom portion of screen)

	// Data sources
	tracker    *perf.Tracker
	components []ComponentInfo

	// Runtime metrics
	LuaCPUTime time.Duration
	LuaMemBytes uint64
	GoMemStats  runtime.MemStats
}

// NewPanel creates a new DevTools panel backed by the given perf tracker.
func NewPanel(tracker *perf.Tracker) *Panel {
	return &Panel{
		tracker: tracker,
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

// UpdateComponents replaces the component snapshot list.
func (p *Panel) UpdateComponents(infos []ComponentInfo) {
	p.components = infos
}

// UpdateLuaMetrics updates Lua runtime metrics.
func (p *Panel) UpdateLuaMetrics(cpuTime time.Duration, memBytes uint64) {
	p.LuaCPUTime = cpuTime
	p.LuaMemBytes = memBytes
}
