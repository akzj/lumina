// Package bridge connects Lua scripts to the Go component system for Lumina v2.
// It converts Lua tables to VNode trees, wraps Lua render functions as Go
// RenderFunc, extracts event handlers, and provides Lua-callable hooks
// (useState, useEffect, useMemo, createElement, useCallback, useRef,
// useReducer, useId, useLayoutEffect) plus animation and router bindings.
package bridge

import (
	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/animation"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/hooks"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/router"
)

// Bridge connects Lua scripts to the Go component system.
type Bridge struct {
	L    *lua.State
	refs []int // Lua registry refs to release after render cycle

	// currentComp is set during a render call so that hooks (useState, etc.)
	// know which component they belong to.
	currentComp *component.Component

	// manager provides SetState for hook-driven re-renders.
	manager *component.Manager

	// hookContexts stores a HookContext per component ID for proper hook
	// lifecycle management (call-index validation, cleanup, etc.).
	hookContexts map[string]*hooks.HookContext

	// animManager manages concurrent animations accessible from Lua.
	animManager *animation.Manager

	// router provides SPA-style routing accessible from Lua.
	router *router.Router
}

// NewBridge creates a new Bridge for the given Lua state.
func NewBridge(L *lua.State) *Bridge {
	return &Bridge{
		L:            L,
		hookContexts: make(map[string]*hooks.HookContext),
	}
}

// SetManager sets the component manager used by hooks (e.g. useState setter).
func (b *Bridge) SetManager(m *component.Manager) {
	b.manager = m
}

// SetAnimationManager sets the animation manager for Lua animation hooks.
func (b *Bridge) SetAnimationManager(mgr *animation.Manager) {
	b.animManager = mgr
}

// SetRouter sets the router for Lua navigation hooks.
func (b *Bridge) SetRouter(r *router.Router) {
	b.router = r
}

// CurrentComponent returns the component currently being rendered (for hooks).
func (b *Bridge) CurrentComponent() *component.Component {
	return b.currentComp
}

// SetCurrentComponent sets the component being rendered (called by the render loop).
func (b *Bridge) SetCurrentComponent(c *component.Component) {
	b.currentComp = c
}

// GetHookContext returns or creates a HookContext for the given component.
// The onDirty callback marks the component dirty for re-rendering.
func (b *Bridge) GetHookContext(comp *component.Component) *hooks.HookContext {
	id := comp.ID()
	if hc, ok := b.hookContexts[id]; ok {
		return hc
	}
	onDirty := func() {
		comp.MarkDirty()
	}
	hc := hooks.NewHookContext(id, onDirty)
	b.hookContexts[id] = hc
	return hc
}

// BeginComponentRender should be called before rendering a component.
// It sets the current component and begins the hook render cycle.
func (b *Bridge) BeginComponentRender(comp *component.Component) {
	b.currentComp = comp
	hc := b.GetHookContext(comp)
	hc.BeginRender()
}

// EndComponentRender should be called after rendering a component.
// It finalizes the hook render cycle and clears the current component.
func (b *Bridge) EndComponentRender() error {
	if b.currentComp != nil {
		hc := b.GetHookContext(b.currentComp)
		b.currentComp = nil
		return hc.EndRender()
	}
	return nil
}

// DestroyComponent cleans up hook context when a component is unmounted.
func (b *Bridge) DestroyComponent(compID string) {
	if hc, ok := b.hookContexts[compID]; ok {
		hc.Destroy()
		delete(b.hookContexts, compID)
	}
}

// Reset clears all hook contexts and tracked refs. Used during hot reload
// to start fresh while allowing state to be restored externally.
func (b *Bridge) Reset() {
	// Destroy all hook contexts (runs effect cleanups).
	for id, hc := range b.hookContexts {
		hc.Destroy()
		delete(b.hookContexts, id)
	}
	// Release tracked Lua registry refs.
	b.ReleaseRefs()
	// Clear current component.
	b.currentComp = nil
}

// WrapRenderFn wraps a Lua render function (stored as registry ref) as a Go
// RenderFunc. When the returned function is called, it:
//  1. Pushes the Lua function from the registry
//  2. Pushes state and props as Lua tables
//  3. Calls the function with PCall(2, 1, 0)
//  4. Converts the returned Lua table to a VNode tree
func (b *Bridge) WrapRenderFn(luaFuncRef int) component.RenderFunc {
	return func(state map[string]any, props map[string]any) *layout.VNode {
		L := b.L

		// Push function from registry.
		L.RawGetI(lua.RegistryIndex, int64(luaFuncRef))
		if !L.IsFunction(-1) {
			L.Pop(1)
			return layout.NewVNode("box")
		}

		// Push state table.
		pushMapAsTable(L, state)

		// Push props table.
		pushMapAsTable(L, props)

		// PCall(2 args, 1 result, 0 error handler).
		if status := L.PCall(2, 1, 0); status != lua.OK {
			L.Pop(1) // pop error
			return layout.NewVNode("box")
		}

		// Convert returned table to VNode.
		if !L.IsTable(-1) {
			L.Pop(1)
			return layout.NewVNode("box")
		}
		vn := b.LuaTableToVNode(-1)
		L.Pop(1)
		return vn
	}
}

// LuaTableToVNode converts a Lua table at stack index idx to a VNode tree.
// See convert.go for implementation.
// (declared here for documentation; implemented in convert.go)

// ExtractHandlers walks a VNode tree and extracts event handlers from Props.
// Returns vnodeID → HandlerMap.
// See handlers.go for implementation.
func (b *Bridge) ExtractHandlers(root *layout.VNode) map[string]event.HandlerMap {
	result := make(map[string]event.HandlerMap)
	b.walkExtract(root, result)
	return result
}

// RegisterHooks registers Lua-callable hooks on the global "lumina" table.
// See hooks.go for implementation.
// (declared here for documentation; implemented in hooks.go)

// TrackRef records a Lua registry ref for release at end of render cycle.
func (b *Bridge) TrackRef(ref int) {
	b.refs = append(b.refs, ref)
}

// ReleaseRefs releases all Lua registry refs accumulated during the render cycle.
func (b *Bridge) ReleaseRefs() {
	for _, ref := range b.refs {
		b.L.Unref(lua.RegistryIndex, ref)
	}
	b.refs = b.refs[:0]
}

// pushMapAsTable pushes a Go map[string]any as a Lua table onto the stack.
func pushMapAsTable(L *lua.State, m map[string]any) {
	if m == nil {
		L.NewTable()
		return
	}
	L.NewTableFrom(m)
}
