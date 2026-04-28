// Package v2 provides the composition root for Lumina v2.
// App ties together the render engine, event handling, devtools, and output
// into a single render-loop orchestrator.
package v2

import (
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/animation"
	"github.com/akzj/lumina/pkg/buffer"
	"github.com/akzj/lumina/pkg/devtools"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
	"github.com/akzj/lumina/pkg/perf"
	"github.com/akzj/lumina/pkg/render"
	"github.com/akzj/lumina/pkg/router"
	"github.com/akzj/lumina/pkg/store"
	"github.com/akzj/lumina/pkg/widget"
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
	scheduler *lua.Scheduler
	quit      chan struct{}
	running   bool

	// Render engine — the single rendering path.
	engine *render.Engine

	// Framework: global store
	store *store.Store

	// Framework: store key → component bindings (for useStore)
	storeBindings map[string][]storeBinding

	// Framework: components subscribed to route changes (for useRoute)
	routeBindings []routeBinding

	// Framework: global keybindings registered via lumina.app
	globalKeys []globalKeyBinding

	// DevTools refresh throttle
	devtoolsLastRefresh time.Time

	// Last error from key handlers or other async Lua calls
	lastError string
}

// NewApp creates a new App with the V2 render engine.
func NewApp(L *lua.State, w, h int, adapter output.Adapter) *App {
	t := perf.NewTracker(60)

	eng := render.NewEngine(L, w, h)
	for _, wgt := range widget.All() {
		eng.RegisterWidget(wgt)
	}
	eng.RegisterLuaAPI()
	eng.ThemeGetter = func() map[string]string {
		t := widget.CurrentTheme
		return map[string]string{
			"base":        t.Base,
			"surface0":    t.Surface0,
			"surface1":    t.Surface1,
			"surface2":    t.Surface2,
			"text":        t.Text,
			"muted":       t.Muted,
			"primary":     t.Primary,
			"primaryDark": t.PrimaryDark,
			"hover":       t.Hover,
			"pressed":     t.Pressed,
			"success":     t.Success,
			"warning":     t.Warning,
			"error":       t.Error,
		}
	}
	registerLuxModules(L)
	eng.SetTracker(t)

	sched := lua.NewScheduler(L)
	sched.OnError = func(err error) {
		// Log async coroutine errors (non-fatal)
		_ = err // TODO: wire to proper logger if available
	}

	eng.SetScheduler(sched)

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
		scheduler: sched,
		quit:      make(chan struct{}),
		engine:    eng,
		store:     store.New(nil),
	}

	// Register app-level APIs that the engine doesn't provide:
	// quit, setInterval, setTimeout, clearInterval, clearTimeout
	a.registerAppLuaAPIs()

	// Register framework APIs: lumina.store, lumina.router, lumina.useStore,
	// lumina.useRoute, lumina.app
	a.registerFrameworkAPIs()

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

// Store returns the global application store.
func (a *App) Store() *store.Store {
	return a.store
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

	dirtyRect := a.engine.DirtyRect()
	if dirtyRect.W > 0 && dirtyRect.H > 0 {
		// If devtools visible, repaint overlay on dirty frames only
		if a.devtools.Visible {
			panelH := a.devtools.Height
			panelY := a.height - panelH
			paintDevToolsOverlay(a.engine.Buffer(), a.devtools, panelY, a.width, panelH)
			panelRect := buffer.Rect{X: 0, Y: panelY, W: a.width, H: panelH}
			dirtyRect = unionRect(dirtyRect, panelRect)
		}
		screen := a.engine.ToBuffer()
		_ = a.adapter.WriteDirty(screen, []buffer.Rect{dirtyRect})
		a.tracker.Record(perf.DirtyRectsOut, 1)
		a.tracker.Record(perf.WriteDirtyCalls, 1)
		_ = a.adapter.Flush()
		a.tracker.Record(perf.FlushCalls, 1)
	}
	// If no dirty rect (idle frame), do NOTHING — no WriteDirty, no Flush

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
		// Tab switching and Elements scroll when devtools is visible.
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
			// Elements tab scroll: arrow keys and page up/down
			if a.devtools.ActiveTab == devtools.TabElements {
				switch e.Key {
				case "ArrowUp", "Up":
					a.devtools.ScrollElements(-1)
					a.refreshDevToolsV2()
					return
				case "ArrowDown", "Down":
					a.devtools.ScrollElements(1)
					a.refreshDevToolsV2()
					return
				case "PageUp":
					a.devtools.ScrollElements(-(a.devtools.Height - 3))
					a.refreshDevToolsV2()
					return
				case "PageDown":
					a.devtools.ScrollElements(a.devtools.Height - 3)
					a.refreshDevToolsV2()
					return
				}
			}
		}
	}

	switch e.Type {
	case "click", "mousedown":
		a.engine.HandleMouseDown(e.X, e.Y)
		a.engine.HandleClick(e.X, e.Y)
	case "mouseup":
		a.engine.HandleMouseUp(e.X, e.Y)
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
		// If scroll is in devtools panel area, scroll the Elements tab
		if a.devtools.Visible && a.devtools.ActiveTab == devtools.TabElements {
			panelY := a.height - a.devtools.Height
			if e.Y >= panelY {
				a.devtools.ScrollElements(delta)
				a.refreshDevToolsV2()
				return
			}
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
