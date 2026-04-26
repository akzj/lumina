package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)


// -----------------------------------------------------------------------
// Overlay Lua API
// -----------------------------------------------------------------------

// luaShowOverlay shows an overlay.
// lumina.showOverlay({ id="...", content={...}, x=N, y=N, width=N, height=N, zIndex=N, modal=bool })
func luaShowOverlay(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("showOverlay: argument must be a table")
		L.Error()
		return 0
	}

	// Extract fields
	L.GetField(1, "id")
	id, _ := L.ToString(-1)
	L.Pop(1)
	if id == "" {
		L.PushString("showOverlay: 'id' is required")
		L.Error()
		return 0
	}

	L.GetField(1, "x")
	x, _ := L.ToInteger(-1)
	L.Pop(1)

	L.GetField(1, "y")
	y, _ := L.ToInteger(-1)
	L.Pop(1)

	L.GetField(1, "width")
	w, _ := L.ToInteger(-1)
	L.Pop(1)
	if w <= 0 {
		w = 20 // default width
	}

	L.GetField(1, "height")
	h, _ := L.ToInteger(-1)
	L.Pop(1)
	if h <= 0 {
		h = 10 // default height
	}

	L.GetField(1, "zIndex")
	zIndex, _ := L.ToInteger(-1)
	L.Pop(1)

	L.GetField(1, "modal")
	modal := L.ToBoolean(-1)
	L.Pop(1)

	// Get content VNode
	L.GetField(1, "content")
	var vnode *VNode
	if L.Type(-1) == lua.TypeTable {
		vnode = LuaVNodeToVNode(L, -1)
	}
	L.Pop(1)

	overlay := &Overlay{
		ID:      id,
		VNode:   vnode,
		X:       int(x),
		Y:       int(y),
		W:       int(w),
		H:       int(h),
		ZIndex:  int(zIndex),
		Visible: true,
		Modal:   modal,
	}

	globalOverlayManager.Show(overlay)
	return 0
}

// luaHideOverlay hides an overlay by ID.
// lumina.hideOverlay("my-dialog")


// luaHideOverlay hides an overlay by ID.
// lumina.hideOverlay("my-dialog")
func luaHideOverlay(L *lua.State) int {
	id := L.CheckString(1)
	globalOverlayManager.Hide(id)
	return 0
}

// luaIsOverlayVisible checks if an overlay is visible.
// lumina.isOverlayVisible("my-dialog") → bool


// luaIsOverlayVisible checks if an overlay is visible.
// lumina.isOverlayVisible("my-dialog") → bool
func luaIsOverlayVisible(L *lua.State) int {
	id := L.CheckString(1)
	L.PushBoolean(globalOverlayManager.IsVisible(id))
	return 1
}

// luaToggleOverlay toggles an overlay's visibility.
// lumina.toggleOverlay("my-dialog") → bool (new state)


// luaToggleOverlay toggles an overlay's visibility.
// lumina.toggleOverlay("my-dialog") → bool (new state)
func luaToggleOverlay(L *lua.State) int {
	id := L.CheckString(1)
	newState := globalOverlayManager.Toggle(id)
	L.PushBoolean(newState)
	return 1
}

// -----------------------------------------------------------------------
// Animation Lua API
// -----------------------------------------------------------------------

// luaStartAnimation starts a named animation imperatively.
// lumina.startAnimation({ id="fade", from=0, to=1, duration=300, easing="easeInOut", loop=false })
