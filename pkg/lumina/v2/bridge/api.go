// Package bridge connects Lua scripts to the Go component system for Lumina v2.
// It converts Lua tables to VNode trees, wraps Lua render functions as Go
// RenderFunc, extracts event handlers, and provides Lua-callable hooks
// (useState, useEffect, useMemo, createElement).
package bridge

import (
	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
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
}

// NewBridge creates a new Bridge for the given Lua state.
func NewBridge(L *lua.State) *Bridge {
	return &Bridge{
		L: L,
	}
}

// SetManager sets the component manager used by hooks (e.g. useState setter).
func (b *Bridge) SetManager(m *component.Manager) {
	b.manager = m
}

// CurrentComponent returns the component currently being rendered (for hooks).
func (b *Bridge) CurrentComponent() *component.Component {
	return b.currentComp
}

// SetCurrentComponent sets the component being rendered (called by the render loop).
func (b *Bridge) SetCurrentComponent(c *component.Component) {
	b.currentComp = c
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
