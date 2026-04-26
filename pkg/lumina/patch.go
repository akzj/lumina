package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// patchComponent replaces a component's render function with new code.
// lumina.patch(id, code) → success, error?
//
// TODO: Currently this only validates that the new code compiles and executes
// without error. It does NOT actually replace the component's render function.
// A full implementation would need to:
//   1. Parse the code to extract the new render function
//   2. Update comp.RenderRef in the Lua registry
//   3. Mark the component dirty to trigger re-render
// For now, this is useful for hot-reload validation (syntax check) only.
func patchComponent(L *lua.State) int {
	id := L.CheckString(1)
	code := L.CheckString(2)


	comp, ok := globalRegistry.components[id]
	if !ok {
		// Try to find by name/type
		for _, c := range globalRegistry.components {
			if c.Name == id || c.Type == id {
				comp = c
				break
			}
		}
	}

	if comp == nil {
		L.PushBoolean(false)
		L.PushString("patchComponent: component not found: " + id)
		return 2
	}

	// Execute the new code - in real implementation this would
	// compile and update the component's render function
	// For now, we just validate the code compiles
	err := L.DoString(code)
	if err != nil {
		L.Pop(1) // pop error
		L.PushBoolean(false)
		L.PushString("patchComponent: syntax error: " + err.Error())
		return 2
	}

	L.PushBoolean(true)
	return 1
}

// eval executes arbitrary Lua code and returns result.
// lumina.eval(code) → success, result?, error?
func eval(L *lua.State) int {
	code := L.CheckString(1)

	err := L.DoString(code)
	if err != nil {
		L.PushBoolean(false)
		L.PushString(err.Error())
		return 2
	}

	// If there's a return value on stack
	if L.GetTop() > 0 {
		L.PushBoolean(true)
		return 2
	}

	L.PushBoolean(true)
	return 1
}
