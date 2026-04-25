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

	// Check for content VNode
	L.GetField(1, "content")
	if !L.IsNil(-1) {
		win.VNode = LuaVNodeToVNode(L, -1)
	}
	L.Pop(1)

	L.PushString(win.ID)
	return 1
}

// luaCloseWindow closes a managed window.
// lumina.closeWindow("win1")
func luaCloseWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.CloseWindow(id)
	return 0
}

// luaFocusWindow brings a window to the front.
// lumina.focusWindow("win1")
func luaFocusWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.FocusWindow(id)
	return 0
}

// luaMoveWindow moves a window to a new position.
// lumina.moveWindow("win1", 20, 10)
func luaMoveWindow(L *lua.State) int {
	id := L.CheckString(1)
	x := int(L.CheckInteger(2))
	y := int(L.CheckInteger(3))
	globalWindowManager.MoveWindow(id, x, y)
	return 0
}

// luaResizeWindow resizes a window.
// lumina.resizeWindow("win1", 50, 30)
func luaResizeWindow(L *lua.State) int {
	id := L.CheckString(1)
	w := int(L.CheckInteger(2))
	h := int(L.CheckInteger(3))
	globalWindowManager.ResizeWindow(id, w, h)
	return 0
}

// luaMinimizeWindow minimizes a window.
// lumina.minimizeWindow("win1")
func luaMinimizeWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.MinimizeWindow(id)
	return 0
}

// luaMaximizeWindow maximizes a window to fill the screen.
// lumina.maximizeWindow("win1")
func luaMaximizeWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.MaximizeWindow(id)
	return 0
}

// luaRestoreWindow restores a minimized or maximized window.
// lumina.restoreWindow("win1")
func luaRestoreWindow(L *lua.State) int {
	id := L.CheckString(1)
	globalWindowManager.RestoreWindow(id)
	return 0
}

// luaTileWindows arranges windows in a tiling layout.
// lumina.tileWindows("horizontal") | "vertical" | "grid"
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
