// Package lumina — Lua API for window management.
package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// luaCreateWindow creates a new managed window.
// lumina.createWindow({ id="win1", title="My Window", x=10, y=5, w=40, h=20 })
// Returns the window ID string.
func luaCreateWindow(L *lua.State) int {
	L.CheckType(1, lua.TypeTable)

	L.GetField(1, "id")
	id, _ := L.ToString(-1)
	L.Pop(1)

	L.GetField(1, "title")
	title, _ := L.ToString(-1)
	L.Pop(1)

	L.GetField(1, "x")
	x64, _ := L.ToInteger(-1)
	L.Pop(1)

	L.GetField(1, "y")
	y64, _ := L.ToInteger(-1)
	L.Pop(1)

	L.GetField(1, "w")
	w64, _ := L.ToInteger(-1)
	L.Pop(1)
	if w64 == 0 {
		w64 = 40
	}

	L.GetField(1, "h")
	h64, _ := L.ToInteger(-1)
	L.Pop(1)
	if h64 == 0 {
		h64 = 20
	}

	if id == "" {
		id = title
	}

	win := globalWindowManager.CreateWindow(id, title, int(x64), int(y64), int(w64), int(h64))

	L.GetField(1, "content")
	if !L.IsNil(-1) {
		win.VNode = LuaVNodeToVNode(L, -1)
	}
	L.Pop(1)

	L.PushString(win.ID)
	return 1
}

// luaCloseWindow closes a managed window.
func luaCloseWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.CloseWindow(id)
	return 0
}

// luaFocusWindow brings a window to the front.
func luaFocusWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.FocusWindow(id)
	return 0
}

// luaMoveWindow moves a window to a new position.
func luaMoveWindow(L *lua.State) int {
	id := L.CheckString(1)
	x := int(L.CheckInteger(2))
	y := int(L.CheckInteger(3))
	globalWindowManager.MoveWindow(id, x, y)
	return 0
}

// luaResizeWindow resizes a window.
func luaResizeWindow(L *lua.State) int {
	id := L.CheckString(1)
	w := int(L.CheckInteger(2))
	h := int(L.CheckInteger(3))
	globalWindowManager.ResizeWindow(id, w, h)
	return 0
}

// luaMinimizeWindow minimizes a window.
func luaMinimizeWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.MinimizeWindow(id)
	return 0
}

// luaMaximizeWindow maximizes a window to fill the screen.
func luaMaximizeWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.MaximizeWindow(id)
	return 0
}

// luaRestoreWindow restores a minimized or maximized window.
func luaRestoreWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.RestoreWindow(id)
	return 0
}

// luaTileWindows arranges windows in a tiling layout.
func luaTileWindows(L *lua.State) int {
	layout := L.OptString(1, "grid")
	switch layout {
	case "horizontal":
		globalWindowManager.TileHorizontal()
	case "vertical":
		globalWindowManager.TileVertical()
	default:
		globalWindowManager.TileGrid()
	}
	return 0
}

// luaGetFocusedWindow returns the ID of the focused window, or nil.
func luaGetFocusedWindow(L *lua.State) int {
	globalWindowManager.mu.RLock()
	defer globalWindowManager.mu.RUnlock()
	if w := globalWindowManager.GetFocused(); w != nil {
		L.PushString(w.ID)
		return 1
	}
	L.PushNil()
	return 1
}

// luaGetWindow returns window info as a table, or nil.
func luaGetWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.mu.RLock()
	defer globalWindowManager.mu.RUnlock()
	w := globalWindowManager.GetWindow(id)
	if w == nil {
		L.PushNil()
		return 1
	}
	L.NewTable()
	L.PushString(w.ID)
	L.SetField(-2, "id")
	L.PushString(w.Title)
	L.SetField(-2, "title")
	L.PushNumber(float64(w.X))
	L.SetField(-2, "x")
	L.PushNumber(float64(w.Y))
	L.SetField(-2, "y")
	L.PushNumber(float64(w.W))
	L.SetField(-2, "w")
	L.PushNumber(float64(w.H))
	L.SetField(-2, "h")
	L.PushBoolean(w.Visible)
	L.SetField(-2, "visible")
	L.PushBoolean(w.Focused)
	L.SetField(-2, "focused")
	L.PushBoolean(w.Minimized)
	L.SetField(-2, "minimized")
	L.PushBoolean(w.Maximized)
	L.SetField(-2, "maximized")
	return 1
}

// luaListWindows returns a list of all window IDs.
func luaListWindows(L *lua.State) int {
	globalWindowManager.mu.RLock()
	defer globalWindowManager.mu.RUnlock()
	wins := globalWindowManager.GetVisible()
	L.NewTable()
	for i, w := range wins {
		L.PushString(w.ID)
		L.RawSetI(-2, int64(i+1))
	}
	return 1
}
