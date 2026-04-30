package v2

import (
	"path/filepath"

	"github.com/akzj/go-lua/pkg/lua"
)

// installRequireHook wraps Lua's global require() to track loaded file paths.
// After each successful require(), the hook resolves the module name to a file
// path using package.searchpath and notifies the addPath callback so the hot
// reload watcher can monitor the file for changes.
//
// This must be called BEFORE the main script is executed so that all require()
// calls during script loading are captured.
//
// The hook is transparent: it does not change require()'s return value or error
// behavior. If addPath is nil, the hook is not installed.
func installRequireHook(L *lua.State, addPath func(string)) {
	if addPath == nil {
		return
	}

	// Register the Go callback as a global that the Lua wrapper will call.
	L.PushFunction(func(L *lua.State) int {
		path := L.CheckString(1)
		if path != "" {
			absPath, err := filepath.Abs(path)
			if err == nil && absPath != "" {
				addPath(absPath)
			}
		}
		return 0
	})
	L.SetGlobal("__lumina_require_hook")

	// Wrap the global require() function. The wrapper:
	// 1. Calls the original require(modname)
	// 2. On success, resolves modname → file path via package.searchpath
	// 3. Calls __lumina_require_hook(path) to notify the Go watcher
	//
	// Uses __lumina_original_require to store the REAL original require on
	// first install, preventing double-wrapping when called again after reload.
	hookCode := `
if not __lumina_original_require then
    __lumina_original_require = require
end
local _original_require = __lumina_original_require
local _searchpath = package.searchpath

require = function(modname)
    local result = _original_require(modname)
    -- After successful load, try to resolve file path and notify Go
    if _searchpath and __lumina_require_hook then
        local path = _searchpath(modname, package.path)
        if path then
            __lumina_require_hook(path)
        end
    end
    return result
end
`
	_ = L.DoString(hookCode)
}
