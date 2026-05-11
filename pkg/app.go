// Package v2 provides the composition root for Lumina v2.
// App ties together the render engine, event handling, devtools, and output
// into a single render-loop orchestrator.
package v2

import (
	"fmt"
	"os"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/animation"
	"github.com/akzj/lumina/pkg/buffer"
	"github.com/akzj/lumina/pkg/devtools"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/hotreload"
	"github.com/akzj/lumina/pkg/output"
	"github.com/akzj/lumina/pkg/perf"
	"github.com/akzj/lumina/pkg/render"
	"github.com/akzj/lumina/pkg/router"
	"github.com/akzj/lumina/pkg/store"
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
	watcher    *hotreload.Watcher // hot reload file watcher (nil when --watch is off)
	scriptPath string             // path of the loaded script (for hot reload via MCP)

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

	// Mouse click tracking: synthesize click only when mouseup at same position
	mouseDownX, mouseDownY int

	// FPS tracking: true when RenderDirty produced visible output this tick
	lastFrameRendered bool
}

// NewApp creates a new App with the V2 render engine.
func NewApp(L *lua.State, w, h int, adapter output.Adapter) *App {
	t := perf.NewTracker(60)

	eng := render.NewEngine(L, w, h)
	eng.RegisterLuaAPI()
	eng.ThemeGetter = func() map[string]string {
		return render.ThemeToMap(render.CurrentTheme)
	}
	eng.ThemeSetter = func(name string) bool {
		return render.SetThemeByName(name)
	}
	registerLuxModules(L)
	eng.SetTracker(t)

	sched := lua.NewScheduler(L)
	sched.OnError = func(err error) {
		fmt.Fprintf(os.Stderr, "coroutine error: %v\n", err)
		// Notify Lua error handler
		if eng.GetOnErrorRef() != 0 {
			L.RawGetI(lua.RegistryIndex, eng.GetOnErrorRef())
			L.PushString(err.Error())
			L.PushString("coroutine")
			L.PCall(2, 0, 0)
		}
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

	// Wire animation manager into engine for lumina.useAnimation() hook
	eng.AnimManager = a.animMgr
	eng.NowMs = func() int64 {
		return time.Now().UnixMilli()
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
	a.setCursorFromEngine()
	_ = a.adapter.Flush()
	a.tracker.Record(perf.FlushCalls, 1)

	a.tracker.EndFrame()
}

// RenderDirty renders only dirty components and outputs changed regions.
func (a *App) RenderDirty() {
	a.lastFrameRendered = false
	a.tracker.BeginFrame()

	a.engine.RenderDirty()

	dirtyRect := a.engine.DirtyRect()
	if dirtyRect.W > 0 && dirtyRect.H > 0 {
		a.lastFrameRendered = true
		// If devtools visible, repaint overlay on dirty frames only
		if a.devtools.Visible {
			panelX, panelY, panelW, panelH := a.devtools.PanelRect(a.width, a.height)
			paintDevToolsOverlay(a.engine.Buffer(), a.devtools, panelX, panelY, panelW, panelH)
			panelRect := buffer.Rect{X: panelX, Y: panelY, W: panelW, H: panelH}
			dirtyRect = unionRect(dirtyRect, panelRect)
		}
		screen := a.engine.ToBuffer()
		_ = a.adapter.WriteDirty(screen, []buffer.Rect{dirtyRect})
		a.tracker.Record(perf.DirtyRectsOut, 1)
		a.tracker.Record(perf.WriteDirtyCalls, 1)
		a.setCursorFromEngine()
		_ = a.adapter.Flush()
		a.tracker.Record(perf.FlushCalls, 1)
	}
	// If no dirty rect (idle frame), do NOTHING — no WriteDirty, no Flush

	a.tracker.EndFrame()
}

// setCursorFromEngine positions the hardware cursor at the focused node's
// cursor hint position. This is essential for IME candidate window positioning.
// The cursor is hidden when no node is focused or no hint is set.
func (a *App) setCursorFromEngine() {
	node := a.engine.FocusedNode()
	if node == nil || node.CursorHintCol < 0 {
		a.adapter.SetCursor(0, 0, false)
		return
	}
	// Calculate absolute screen position from node position + cursor hint offset
	x := node.X + node.CursorHintCol
	y := node.Y + node.CursorHintRow
	a.adapter.SetCursor(x, y, true)
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
		if e.Key == "F5" {
			if a.scriptPath != "" {
				a.reloadScript(a.scriptPath)
			}
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
			case "3":
				a.devtools.CycleAnchor(a.width, a.height)
				a.syncEngineViewport()
				a.paintDevToolsV2()
				return
			}
			// Elements tab: inspect pick, clear selection, scroll
			if a.devtools.ActiveTab == devtools.TabElements {
				switch e.Key {
				case "i":
					a.devtools.ArmElementsPick()
					a.refreshDevToolsV2()
					return
				case "0", "Escape":
					a.devtools.ClearElementsSelection()
					a.refreshDevToolsV2()
					return
				case "ArrowUp", "Up":
					a.devtools.ScrollElements(-1)
					a.refreshDevToolsV2()
					return
				case "ArrowDown", "Down":
					a.devtools.ScrollElements(1)
					a.refreshDevToolsV2()
					return
				case "PageUp":
					a.devtools.ScrollElements(-a.devtools.ElementsPageScrollLines())
					a.refreshDevToolsV2()
					return
				case "PageDown":
					a.devtools.ScrollElements(a.devtools.ElementsPageScrollLines())
					a.refreshDevToolsV2()
					return
				}
			}
		}
	}

	// Panel mouse routing: intercept events inside/on the devtools panel before engine.
	if a.devtools.Visible {
		switch e.Type {
		case "mousedown":
			// Elements inspect pick: click disarms pick (selection was already set by hover).
			if a.devtools.ElementsPickArmed() {
				a.devtools.ClearElementsPickArm()
				a.mouseDownX = -1
				a.mouseDownY = -1
				a.refreshDevToolsV2()
				return
			}
			// Resize border drag start.
			if a.devtools.HitResizeBorder(e.X, e.Y, a.width, a.height) {
				coord := e.Y
				if a.devtools.Anchor == devtools.AnchorRight {
					coord = e.X
				}
				a.devtools.StartResize(coord)
				return
			}
			// Click inside panel content.
			if a.devtools.ContainsPoint(e.X, e.Y, a.width, a.height) {
				a.mouseDownX = -1
				a.mouseDownY = -1
				a.handleDevToolsPanelClick(e.X, e.Y)
				return
			}

		case "mousemove":
			if a.devtools.IsResizing() {
				coord := e.Y
				if a.devtools.Anchor == devtools.AnchorRight {
					coord = e.X
				}
				a.devtools.UpdateResize(coord, a.width, a.height)
				a.syncEngineViewport()
				a.refreshDevToolsV2()
				return
			}
			// Elements inspect pick: hover over app area to preview selection in real-time.
			if a.devtools.ElementsPickArmed() && a.devtools.ActiveTab == devtools.TabElements {
				inAppArea := !a.devtools.ContainsPoint(e.X, e.Y, a.width, a.height)
				if inAppArea {
					r := a.engine.Root()
					if r != nil && r.RootNode != nil {
						if hit := a.engine.HitTestScreen(e.X, e.Y); hit != nil {
							if idx := preorderIndexOfHit(r.RootNode, hit); idx >= 0 {
								if idx != a.devtools.ElementsSelectedIdx() {
									a.devtools.SetElementsSelection(idx)
									a.devtools.ExpandToNode(idx)
									a.paintDevToolsV2()
								}
							}
						}
					}
				}
				return // suppress engine hover events during pick mode
			}

		case "mouseup":
			if a.devtools.IsResizing() {
				a.devtools.EndResize()
				return
			}

		case "scroll":
			if a.devtools.ContainsPoint(e.X, e.Y, a.width, a.height) {
				if a.devtools.ActiveTab == devtools.TabElements {
					delta := 1
					if e.Key == "up" {
						delta = -1
					}
					a.devtools.ScrollElements(delta)
					a.refreshDevToolsV2()
				}
				return
			}
		}
	}

	switch e.Type {
	case "mousedown":
		a.mouseDownX = e.X
		a.mouseDownY = e.Y
		a.engine.HandleMouseDown(e.X, e.Y, e.Button)
	case "mouseup":
		a.engine.HandleMouseUp(e.X, e.Y)
		// Synthesize click only if mouseup at same position as mousedown
		// AND mousedown handler did not call preventDefault
		if e.X == a.mouseDownX && e.Y == a.mouseDownY && !a.engine.ClickPrevented() {
			a.engine.HandleClick(e.X, e.Y, e.Button)
		}
	case "click":
		// Explicit click event (e.g. from WebSocket adapter)
		a.engine.HandleClick(e.X, e.Y, e.Button)
	case "mousemove":
		a.engine.HandleMouseMove(e.X, e.Y)
	case "keydown":
		a.engine.HandleKeyDown(e.Key)
	case "scroll":
		// Horizontal wheel (xterm SGR 66/67) uses Key "left"/"right"; Shift+vertical
		// wheel uses "up"/"down" with Shift set.
		if e.Key == "left" || e.Key == "right" {
			delta := 1
			if e.Key == "left" {
				delta = -1
			}
			a.engine.HandleScrollH(e.X, e.Y, delta)
			break
		}
		// Vertical wheel: Key is "up"/"down"
		delta := 1
		if e.Key == "up" {
			delta = -1
		}
		// Panel scroll is handled in the routing block above; skip engine scroll if in panel.
		if a.devtools.Visible && a.devtools.ContainsPoint(e.X, e.Y, a.width, a.height) {
			return
		}
		// Windows Terminal / ConPTY often omit Shift on wheel in SGR reports; many
		// terminals still set Alt or Ctrl — treat any of them like Shift+wheel for
		// horizontal scroll (iTerm/macOS typically sets Shift correctly).
		modWheel := e.Shift || e.Alt || e.Ctrl
		if modWheel {
			a.engine.HandleScrollH(e.X, e.Y, delta)
		} else {
			a.engine.HandleScroll(e.X, e.Y, delta)
		}
	}
}

// Screen returns the current screen buffer.
func (a *App) Screen() *buffer.Buffer {
	return a.engine.ToBuffer()
}

// FocusedID returns the currently focused VNode ID.
// Delegates to the engine's focus tracking.
func (a *App) FocusedID() string {
	node := a.engine.FocusedNode()
	if node != nil {
		return node.ID
	}
	return ""
}

// Resize resizes the screen (terminal size change).
func (a *App) Resize(w, h int) {
	a.width = w
	a.height = h
	// Always resize the buffer to the full terminal size.
	a.engine.Resize(w, h)
	if a.devtools.Visible {
		a.devtools.InitSizeForAnchor(w, h) // re-clamp panel size on terminal resize
		a.syncEngineViewport()
	}
}

// syncEngineViewport constrains the engine layout to the app content area (excluding
// the devtools panel). The CellBuffer stays at full terminal size so the panel can
// be painted on top.
func (a *App) syncEngineViewport() {
	if a.devtools.Visible {
		appW, appH := a.devtools.AppRect(a.width, a.height)
		a.engine.SetLayoutBounds(appW, appH)
	} else {
		a.engine.SetLayoutBounds(a.width, a.height)
	}
}

// SetState updates a component's state (marks it dirty for re-render).
func (a *App) SetState(compID string, key string, value any) {
	a.engine.SetState(compID, key, value)
}

// tickDevTools is called every frame tick from the event loop.
func (a *App) tickDevTools(rendered bool) {
	a.tickDevToolsV2(rendered)
}

// handleDevToolsPanelClick handles a mousedown inside the panel area.
func (a *App) handleDevToolsPanelClick(mx, my int) {
	// Tab bar click → switch tab.
	if a.devtools.HitTabBar(mx, my, a.width, a.height) {
		if tab := a.devtools.TabAtPoint(mx, my, a.width, a.height); tab >= 0 {
			a.devtools.SetTab(tab)
			a.refreshDevToolsV2()
		}
		return
	}
	// Elements content area click → select tree node.
	if a.devtools.ActiveTab == devtools.TabElements {
		a.handleDevToolsElementsClick(mx, my)
	}
}

// handleDevToolsElementsClick maps a mouse click inside the Elements content area
// to a visible tree row. Clicking the fold icon toggles collapse; clicking elsewhere selects.
func (a *App) handleDevToolsElementsClick(mx, my int) {
	panelX, panelY, _, _ := a.devtools.PanelRect(a.width, a.height)
	contentStartY := panelY + a.devtools.ContentStartRow()
	row := my - contentStartY
	if row < 0 {
		return
	}
	// Account for scroll-up indicator row.
	scrollY := a.devtools.ElementsScrollY()
	if scrollY > 0 {
		row-- // first content row is the "↑ N more above" indicator
	}
	if row < 0 {
		return
	}
	vi := scrollY + row
	visIndices := a.devtools.VisibleIndices()
	if vi < 0 || vi >= len(visIndices) {
		return
	}
	ni := visIndices[vi]
	nodes := a.devtools.NodeTree()
	if ni < 0 || ni >= len(nodes) {
		return
	}
	node := nodes[ni]

	// Content column offset: right anchor has a resize column before content.
	contentStartX := panelX
	if a.devtools.Anchor == devtools.AnchorRight {
		contentStartX = panelX + devtools.ResizeHandleThick
	}
	// Icon is at: contentStartX + 2 (leading spaces) + 2*depth (indent)
	iconCol := contentStartX + 2 + 2*node.Depth
	if mx == iconCol && a.devtools.NodeHasChildren(ni) {
		a.devtools.ToggleCollapse(vi)
		a.refreshDevToolsV2()
		return
	}

	a.devtools.SetElementsSelection(ni)
	a.refreshDevToolsV2()
}
