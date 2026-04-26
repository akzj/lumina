package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)


// luaOnKey registers a key binding: lumina.onKey("q", function() ... end)
func luaOnKey(L *lua.State) int {
	key := L.CheckString(1)
	if L.Type(2) != lua.TypeFunction {
		L.PushString("onKey: second argument must be a function")
		L.Error()
		return 0
	}

	// Store the callback in the Lua registry
	L.PushValue(2)
	ref := L.Ref(lua.RegistryIndex)

	keyBindingsMu.Lock()
	keyBindings[key] = ref
	keyBindingsMu.Unlock()

	return 0
}

// ClearKeyBindings removes all key bindings (for testing).


// ClearKeyBindings removes all key bindings (for testing).
func ClearKeyBindings() {
	keyBindingsMu.Lock()
	defer keyBindingsMu.Unlock()
	keyBindings = make(map[string]int)
}

// -----------------------------------------------------------------------
// Suspense + lazy
// -----------------------------------------------------------------------

// registerSuspenseFactory creates the Suspense component factory table
// and sets it as lumina.Suspense on the module table at stack top.
