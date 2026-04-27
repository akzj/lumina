package v2

import (
	"fmt"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/animation"
	"github.com/akzj/lumina/pkg/lumina/v2/bridge"
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/render"
	"github.com/akzj/lumina/pkg/lumina/v2/router"
)

// NewAppWithLua creates an App wired to a Lua state with bridge, animation
// manager, and router. The Lua "lumina" global table is populated with all
// hooks (useState, useEffect, createElement, etc.) plus createComponent.
//
// Call Run() to start the event loop, or use the manual API (RegisterComponent,
// HandleEvent, RenderAll/RenderDirty) for testing.
func NewAppWithLua(L *lua.State, w, h int, adapter output.Adapter) *App {
	app := NewApp(w, h, adapter)

	app.luaState = L
	app.animMgr = animation.NewManager()
	app.routerMgr = router.New()
	app.timerMgr = newTimerManager()
	app.quit = make(chan struct{})

	// Create bridge and wire dependencies.
	b := bridge.NewBridge(L)
	b.SetManager(app.manager)
	b.SetAnimationManager(app.animMgr)
	b.SetRouter(app.routerMgr)
	app.bridge = b

	// Register all Lua hooks (useState, useEffect, createElement, etc.).
	b.RegisterHooks()

	// Register app-level Lua APIs (createComponent, removeComponent, quit).
	app.registerAppLuaAPIs()

	return app
}

// NewAppWithEngine creates an App using the new V2 render engine.
// This uses the persistent RenderNode tree with incremental reconcile/layout/paint.
func NewAppWithEngine(L *lua.State, w, h int, adapter output.Adapter) *App {
	app := NewApp(w, h, adapter)

	app.luaState = L
	app.animMgr = animation.NewManager()
	app.routerMgr = router.New()
	app.timerMgr = newTimerManager()
	app.quit = make(chan struct{})

	// Create the new render engine.
	eng := render.NewEngine(L, w, h)
	eng.RegisterLuaAPI()
	eng.SetTracker(app.tracker)
	app.engine = eng

	// Register app-level APIs that the engine doesn't provide:
	// quit, setInterval, setTimeout, clearInterval, clearTimeout
	app.registerAppLuaAPIs_V2()

	return app
}

// registerAppLuaAPIs_V2 registers app-level functions for the V2 engine.
// The engine already provides createComponent/createElement/useState/defineComponent.
// This adds: quit, setInterval, setTimeout, clearInterval, clearTimeout.
func (a *App) registerAppLuaAPIs_V2() {
	L := a.luaState

	// Get the existing "lumina" global table (already created by engine.RegisterLuaAPI).
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		L.Pop(1)
		return
	}
	tblIdx := L.AbsIndex(-1)

	// lumina.quit()
	L.PushFunction(a.luaQuit)
	L.SetField(tblIdx, "quit")

	// lumina.setInterval(fn, ms)
	L.PushFunction(a.luaSetInterval)
	L.SetField(tblIdx, "setInterval")

	// lumina.setTimeout(fn, ms)
	L.PushFunction(a.luaSetTimeout)
	L.SetField(tblIdx, "setTimeout")

	// lumina.clearInterval(id)
	L.PushFunction(a.luaClearTimer)
	L.SetField(tblIdx, "clearInterval")

	// lumina.clearTimeout(id)
	L.PushFunction(a.luaClearTimer)
	L.SetField(tblIdx, "clearTimeout")

	L.SetGlobal("lumina")
}

// registerAppLuaAPIs registers app-level functions on the "lumina" global
// table: createComponent, removeComponent, quit.
func (a *App) registerAppLuaAPIs() {
	L := a.luaState

	// Get or create the "lumina" global table.
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		L.Pop(1)
		L.NewTable()
	}
	tblIdx := L.AbsIndex(-1)

	// lumina.createComponent(config) — registers a component from Lua.
	L.PushFunction(a.luaCreateComponent)
	L.SetField(tblIdx, "createComponent")

	// lumina.removeComponent(id) — unregisters a component.
	L.PushFunction(a.luaRemoveComponent)
	L.SetField(tblIdx, "removeComponent")

	// lumina.quit() — signals the event loop to stop.
	L.PushFunction(a.luaQuit)
	L.SetField(tblIdx, "quit")

	// lumina.setInterval(fn, ms) — repeating timer, returns ID.
	L.PushFunction(a.luaSetInterval)
	L.SetField(tblIdx, "setInterval")

	// lumina.setTimeout(fn, ms) — one-shot timer, returns ID.
	L.PushFunction(a.luaSetTimeout)
	L.SetField(tblIdx, "setTimeout")

	// lumina.clearInterval(id) — cancel a timer.
	L.PushFunction(a.luaClearTimer)
	L.SetField(tblIdx, "clearInterval")

	// lumina.clearTimeout(id) — alias for clearInterval.
	L.PushFunction(a.luaClearTimer)
	L.SetField(tblIdx, "clearTimeout")

	L.SetGlobal("lumina")
}

// luaCreateComponent implements lumina.createComponent(config).
//
// Config is a Lua table:
//
//	{
//	    id     = "counter",       -- unique component ID (required)
//	    name   = "Counter",       -- display name (optional, defaults to id)
//	    x      = 0,  y = 0,      -- screen position
//	    w      = 40, h = 10,     -- dimensions
//	    zIndex = 0,               -- stacking order (optional, default 0)
//	    render = function(state, props) ... end,  -- render function (required)
//	}
func (a *App) luaCreateComponent(L *lua.State) int {
	L.CheckType(1, lua.TypeTable)
	absIdx := L.AbsIndex(1)

	// Read required fields.
	id := L.GetFieldString(absIdx, "id")
	if id == "" {
		L.PushString("createComponent: 'id' is required")
		L.Error()
		return 0
	}

	name := L.GetFieldString(absIdx, "name")
	if name == "" {
		name = id
	}

	x := int(L.GetFieldInt(absIdx, "x"))
	y := int(L.GetFieldInt(absIdx, "y"))
	w := int(L.GetFieldInt(absIdx, "w"))
	h := int(L.GetFieldInt(absIdx, "h"))
	zIndex := int(L.GetFieldInt(absIdx, "zIndex"))

	if w <= 0 || h <= 0 {
		L.PushString(fmt.Sprintf("createComponent: invalid dimensions w=%d h=%d", w, h))
		L.Error()
		return 0
	}

	// Get the render function and store as registry ref.
	L.GetField(absIdx, "render")
	if !L.IsFunction(-1) {
		L.Pop(1)
		L.PushString("createComponent: 'render' function is required")
		L.Error()
		return 0
	}
	renderRef := L.Ref(lua.RegistryIndex)

	rect := buffer.Rect{X: x, Y: y, W: w, H: h}
	b := a.bridge

	// Create a Go RenderFunc that calls the Lua render function via bridge.
	// The bridge handles: push function, push state/props, pcall, convert
	// returned Lua table to VNode tree.
	//
	// We wrap with BeginComponentRender/EndComponentRender so that hooks
	// (useState, useEffect, etc.) know which component they belong to.
	renderFn := func(state map[string]any, props map[string]any) *layout.VNode {
		comp := a.manager.Get(id)
		if comp == nil {
			return layout.NewVNode("box")
		}
		b.BeginComponentRender(comp)
		vn := b.WrapRenderFn(renderRef)(state, props)
		if err := b.EndComponentRender(); err != nil {
			// Hook count mismatch — log but don't crash.
			_ = err
		}
		return vn
	}

	a.RegisterComponent(id, name, rect, zIndex, renderFn)
	return 0
}

// luaRemoveComponent implements lumina.removeComponent(id).
func (a *App) luaRemoveComponent(L *lua.State) int {
	id := L.CheckString(1)
	a.UnregisterComponent(id)
	if a.bridge != nil {
		a.bridge.DestroyComponent(id)
	}
	return 0
}

// luaQuit implements lumina.quit().
func (a *App) luaQuit(_ *lua.State) int {
	a.Stop()
	return 0
}
