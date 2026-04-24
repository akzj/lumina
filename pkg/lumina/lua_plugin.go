package lumina

import (
	"fmt"

	"github.com/akzj/go-lua/pkg/lua"
)

// luaRegisterPlugin implements lumina.registerPlugin(opts)
func luaRegisterPlugin(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("registerPlugin: expected table argument")
		L.Error()
		return 0
	}

	plugin := &Plugin{
		Hooks:      make(map[string]any),
		Components: make(map[string]string),
		Metadata:   make(map[string]string),
	}

	// Read name
	L.GetField(1, "name")
	if s, ok := L.ToString(-1); ok {
		plugin.Name = s
	}
	L.Pop(1)

	// Read version
	L.GetField(1, "version")
	if s, ok := L.ToString(-1); ok {
		plugin.Version = s
	}
	L.Pop(1)

	// Read init function
	L.GetField(1, "init")
	hasInit := L.Type(-1) == lua.TypeFunction
	initRef := 0
	if hasInit {
		initRef = L.Ref(lua.RegistryIndex)
	} else {
		L.Pop(1)
	}

	// Read hooks table
	L.GetField(1, "hooks")
	if L.Type(-1) == lua.TypeTable {
		L.PushNil()
		for L.Next(-2) {
			hookName, ok := L.ToString(-2)
			if ok && L.Type(-1) == lua.TypeFunction {
				ref := L.Ref(lua.RegistryIndex)
				plugin.Hooks[hookName] = ref
			} else {
				L.Pop(1)
			}
		}
	}
	L.Pop(1)

	// Set init function that calls the Lua init
	if hasInit {
		capturedL := L
		capturedRef := initRef
		plugin.InitFn = func() error {
			capturedL.RawGetI(lua.RegistryIndex, int64(capturedRef))
			status := capturedL.PCall(0, 0, 0)
			if status != 0 {
				msg, _ := capturedL.ToString(-1)
				capturedL.Pop(1)
				return fmt.Errorf("plugin init error: %s", msg)
			}
			return nil
		}
	}

	// Register
	registry := GetPluginRegistry()
	if err := registry.Register(plugin); err != nil {
		L.PushString(err.Error())
		L.Error()
		return 0
	}

	return 0
}

// luaUsePlugin implements lumina.usePlugin(name)
func luaUsePlugin(L *lua.State) int {
	name := L.CheckString(1)

	registry := GetPluginRegistry()
	if err := registry.InitPlugin(name); err != nil {
		L.PushString(err.Error())
		L.Error()
		return 0
	}

	// Make plugin hooks available
	plugin, ok := registry.Get(name)
	if !ok {
		return 0
	}

	// Register each hook as lumina.<hookName>
	for hookName, ref := range plugin.Hooks {
		refInt, ok := ref.(int)
		if !ok {
			continue
		}
		// Get the lumina table
		L.GetGlobal("lumina")
		if L.Type(-1) != lua.TypeTable {
			L.Pop(1)
			continue
		}
		// Push the hook function from registry
		L.RawGetI(lua.RegistryIndex, int64(refInt))
		L.SetField(-2, hookName)
		L.Pop(1) // pop lumina table
	}

	return 0
}

// luaGetPlugins implements lumina.getPlugins() -> table of names
func luaGetPlugins(L *lua.State) int {
	registry := GetPluginRegistry()
	names := registry.List()
	L.NewTable()
	for i, name := range names {
		L.PushString(name)
		L.RawSetI(-2, int64(i+1))
	}
	return 1
}
