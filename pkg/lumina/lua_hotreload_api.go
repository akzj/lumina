package lumina

import (
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)


// -----------------------------------------------------------------------
// Hot Reload Lua API
// -----------------------------------------------------------------------

// luaEnableHotReload enables hot reload with optional config.
// lumina.enableHotReload({ paths = {"app.lua"}, interval = 500 })
func luaEnableHotReload(L *lua.State) int {
	globalHotReloader.Enable(true)

	if L.Type(1) == lua.TypeTable {
		var interval time.Duration
		var paths []string

		L.GetField(1, "interval")
		if !L.IsNoneOrNil(-1) {
			ms, _ := L.ToNumber(-1)
			if ms > 0 {
				interval = time.Duration(ms) * time.Millisecond
			}
		}
		L.Pop(1)

		L.GetField(1, "paths")
		if L.Type(-1) == lua.TypeTable {
			n := int(L.RawLen(-1))
			for i := 1; i <= n; i++ {
				L.RawGetI(-1, int64(i))
				if s, ok := L.ToString(-1); ok {
					paths = append(paths, s)
				}
				L.Pop(1)
			}
		}
		L.Pop(1)

		// Apply config under lock
		globalHotReloader.SetConfig(interval, paths)
	}

	return 0
}

// luaDisableHotReload disables hot reload.
// lumina.disableHotReload()


// luaDisableHotReload disables hot reload.
// lumina.disableHotReload()
func luaDisableHotReload(L *lua.State) int {
	globalHotReloader.Enable(false)
	return 0
}

// -----------------------------------------------------------------------
// Focus Scope Lua API
// -----------------------------------------------------------------------

// luaPushFocusScope pushes a new focus scope.
// lumina.pushFocusScope({ focusableIDs = {"input1", "btn-ok", "btn-cancel"} })
