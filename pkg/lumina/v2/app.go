// Package v2 provides the composition root for Lumina v2.
// App ties together the render engine, event handling, devtools, and output
// into a single render-loop orchestrator.
package v2

import (
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/animation"
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/devtools"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
	"github.com/akzj/lumina/pkg/lumina/v2/render"
	"github.com/akzj/lumina/pkg/lumina/v2/router"
)

// App is the composition root — ties all v2 modules together.
type App struct {
	width   int
	height  int
	adapter output.Adapter
	tracker *perf.Tracker

	// DevTools panel
	devtools *devtools.Panel

	// Runtime (populated by NewApp / Run)
	luaState  *lua.State
	animMgr   *animation.Manager
	routerMgr *router.Router
	timerMgr  *timerManager
	quit      chan struct{}
	running   bool

	// Render engine — the single rendering path.
	engine *render.Engine

	// DevTools refresh throttle
	devtoolsLastRefresh time.Time
}

// NewApp creates a new App with the V2 render engine.
func NewApp(L *lua.State, w, h int, adapter output.Adapter) *App {
	t := perf.NewTracker(60)

	eng := render.NewEngine(L, w, h)
	eng.RegisterLuaAPI()
	eng.SetTracker(t)

	a := &App{
		width:     w,
		height:    h,
		adapter:   adapter,
		tracker:   t,
		devtools:  devtools.NewPanel(t),
		luaState:  L,
		animMgr:   animation.NewManager(),
		routerMgr: router.New(),
		timerMgr:  newTimerManager(),
		quit:      make(chan struct{}),
		engine:    eng,
	}

	// Register app-level APIs that the engine doesn't provide:
	// quit, setInterval, setTimeout, clearInterval, clearTimeout
	a.registerAppLuaAPIs()

	return a
}

// NewTestApp creates an App with a TestAdapter for testing.
func NewTestApp(w, h int) (*App, *output.TestAdapter) {
	ta := output.NewTestAdapter()
	L := lua.NewState()
	app := NewApp(L, w, h, ta)
	return app, ta
}

// Tracker returns the performance tracker. Call Enable() to start recording.
func (a *App) Tracker() *perf.Tracker {
	return a.tracker
}

// Engine returns the render engine.
func (a *App) Engine() *render.Engine {
	return a.engine
}

// DevTools returns the DevTools panel (for testing/inspection).
func (a *App) DevTools() *devtools.Panel {
	return a.devtools
}

// RenderAll performs a full render using the engine.
func (a *App) RenderAll() {
	a.tracker.BeginFrame()

	a.engine.RenderAll()

	screen := a.engine.ToBuffer()
	_ = a.adapter.WriteFull(screen)
	a.tracker.Record(perf.WriteFullCalls, 1)
	_ = a.adapter.Flush()
	a.tracker.Record(perf.FlushCalls, 1)

	a.tracker.EndFrame()
}

// RenderDirty renders only dirty components and outputs changed regions.
func (a *App) RenderDirty() {
	a.tracker.BeginFrame()

	a.engine.RenderDirty()

	// Paint devtools overlay on top of the rendered content if visible.
	if a.devtools.Visible {
		panelH := a.devtools.Height
		panelY := a.height - panelH
		paintDevToolsOverlay(a.engine.Buffer(), a.devtools, panelY, a.width, panelH)
	}

	dirtyRect := a.engine.DirtyRect()
	// When devtools is visible, expand dirty rect to include the panel area.
	if a.devtools.Visible {
		panelH := a.devtools.Height
		panelY := a.height - panelH
		panelRect := buffer.Rect{X: 0, Y: panelY, W: a.width, H: panelH}
		dirtyRect = unionRect(dirtyRect, panelRect)
	}

	if dirtyRect.W > 0 && dirtyRect.H > 0 {
		screen := a.engine.ToBuffer()
		_ = a.adapter.WriteDirty(screen, []buffer.Rect{dirtyRect})
		a.tracker.Record(perf.DirtyRectsOut, 1)
		a.tracker.Record(perf.WriteDirtyCalls, 1)
		_ = a.adapter.Flush()
		a.tracker.Record(perf.FlushCalls, 1)
	}

	a.tracker.EndFrame()
}

// unionRect returns the bounding rect containing both a and b.
func unionRect(a, b buffer.Rect) buffer.Rect {
	if a.W == 0 || a.H == 0 {
		return b
	}
	if b.W == 0 || b.H == 0 {
		return a
	}
	x1 := a.X
	if b.X < x1 {
		x1 = b.X
	}
	y1 := a.Y
	if b.Y < y1 {
		y1 = b.Y
	}
	x2 := a.X + a.W
	if b.X+b.W > x2 {
		x2 = b.X + b.W
	}
	y2 := a.Y + a.H
	if b.Y+b.H > y2 {
		y2 = b.Y + b.H
	}
	return buffer.Rect{X: x1, Y: y1, W: x2 - x1, H: y2 - y1}
}

// HandleEvent dispatches an input event through the engine.
// F12 and DevTools tab-switching keys are intercepted before normal dispatch.
func (a *App) HandleEvent(e *event.Event) {
	if e.Type == "keydown" {
		if e.Key == "F12" {
			a.toggleDevToolsV2()
			return
		}
		// Tab switching when devtools is visible.
		if a.devtools.Visible {
			switch e.Key {
			case "1":
				a.devtools.SetTab(devtools.TabElements)
				a.refreshDevToolsV2()
				return
			case "2":
				a.devtools.SetTab(devtools.TabPerf)
				a.refreshDevToolsV2()
				return
			}
		}
	}

	switch e.Type {
	case "click", "mousedown":
		a.engine.HandleClick(e.X, e.Y)
	case "mousemove":
		a.engine.HandleMouseMove(e.X, e.Y)
	case "keydown":
		a.engine.HandleKeyDown(e.Key)
	case "scroll":
		// Scroll direction is in e.Key ("up"/"down"), convert to delta
		delta := 1
		if e.Key == "up" {
			delta = -1
		}
		a.engine.HandleScroll(e.X, e.Y, delta)
	}
}

// Screen returns the current screen buffer.
func (a *App) Screen() *buffer.Buffer {
	return a.engine.ToBuffer()
}

// FocusedID returns the currently focused VNode ID.
// Delegates to the engine's focus tracking.
func (a *App) FocusedID() string {
	// The engine doesn't expose focus tracking yet.
	// TODO: Add focus tracking to the engine.
	return ""
}

// Resize resizes the screen.
func (a *App) Resize(w, h int) {
	a.width = w
	a.height = h
	a.engine.Resize(w, h)
}

// SetState updates a component's state (marks it dirty for re-render).
func (a *App) SetState(compID string, key string, value any) {
	a.engine.SetState(compID, key, value)
}

// tickDevTools is called every frame tick from the event loop.
func (a *App) tickDevTools() {
	a.tickDevToolsV2()
}
