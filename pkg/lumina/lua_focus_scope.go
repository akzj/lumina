package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)


// -----------------------------------------------------------------------
// Focus Scope Lua API
// -----------------------------------------------------------------------

// luaPushFocusScope pushes a new focus scope.
// lumina.pushFocusScope({ focusableIDs = {"input1", "btn-ok", "btn-cancel"} })
func luaPushFocusScope(L *lua.State) int {
	scope := &FocusScope{}

	if L.Type(1) == lua.TypeTable {
		L.GetField(1, "id")
		if !L.IsNoneOrNil(-1) {
			scope.ID, _ = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(1, "focusableIDs")
		if L.Type(-1) == lua.TypeTable {
			n := int(L.RawLen(-1))
			for i := 1; i <= n; i++ {
				L.RawGetI(-1, int64(i))
				if s, ok := L.ToString(-1); ok {
					scope.FocusableIDs = append(scope.FocusableIDs, s)
				}
				L.Pop(1)
			}
		}
		L.Pop(1)
	}

	PushFocusScope(scope)
	return 0
}

// luaPopFocusScope pops the top focus scope.
// lumina.popFocusScope()


// luaPopFocusScope pops the top focus scope.
// lumina.popFocusScope()
func luaPopFocusScope(L *lua.State) int {
	PopFocusScope()
	return 0
}

// -----------------------------------------------------------------------
// Router Lua API
// -----------------------------------------------------------------------

// luaCreateRouter creates a router with route definitions.
// lumina.createRouter({ routes = { {path="/"}, {path="/users/:id"} } })
// NOTE: Only called from Lua main thread — globalRouter assignment is safe.
