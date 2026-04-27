package bridge

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// luaNavigate implements lumina.navigate(path).
// Navigates the router to the given path.
func (b *Bridge) luaNavigate(L *lua.State) int {
	if b.router == nil {
		L.PushString("navigate: no router configured")
		L.Error()
		return 0
	}
	path := L.CheckString(1)
	b.router.Navigate(path)
	return 0
}

// luaBack implements lumina.back().
// Pops the last path from history and navigates back.
// Returns true if navigation happened, false if history is empty.
func (b *Bridge) luaBack(L *lua.State) int {
	if b.router == nil {
		L.PushBoolean(false)
		return 1
	}
	ok := b.router.Back()
	L.PushBoolean(ok)
	return 1
}

// luaUseRoute implements lumina.useRoute().
// Returns a table: { path = currentPath, params = { key = value, ... } }
func (b *Bridge) luaUseRoute(L *lua.State) int {
	if b.router == nil {
		// Return a default route table.
		L.NewTable()
		tblIdx := L.AbsIndex(-1)
		L.PushString("/")
		L.SetField(tblIdx, "path")
		L.NewTable()
		L.SetField(tblIdx, "params")
		return 1
	}

	L.NewTable()
	tblIdx := L.AbsIndex(-1)

	L.PushString(b.router.CurrentPath())
	L.SetField(tblIdx, "path")

	// Push params as a Lua table.
	params := b.router.Params()
	L.NewTable()
	paramsIdx := L.AbsIndex(-1)
	for k, v := range params {
		L.PushString(v)
		L.SetField(paramsIdx, k)
	}
	L.SetField(tblIdx, "params")

	return 1
}
