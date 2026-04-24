package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// luaAnnounce implements lumina.announce(message, priority)
func luaAnnounce(L *lua.State) int {
	message := L.CheckString(1)
	priority := "polite"
	if s, ok := L.ToString(2); ok && s != "" {
		priority = s
	}
	GetAnnouncer().Announce(message, priority)
	return 0
}

// luaCreateTestRenderer implements lumina.createTestRenderer() -> table
func luaCreateTestRenderer(L *lua.State) int {
	tr := NewTestRenderer()

	L.NewTable()

	// render(vnodeTable)
	L.PushFunction(func(L *lua.State) int {
		if L.Type(1) != lua.TypeTable {
			L.PushString("render: expected table argument")
			L.Error()
			return 0
		}
		m := luaTableToMap(L, 1)
		tr.Render(m)
		return 0
	})
	L.SetField(-2, "render")

	// getByText(text) -> table or nil
	L.PushFunction(func(L *lua.State) int {
		text := L.CheckString(1)
		node := tr.GetByText(text)
		if node == nil {
			L.PushNil()
			return 1
		}
		pushTestVNode(L, node)
		return 1
	})
	L.SetField(-2, "getByText")

	// getByRole(role) -> table or nil
	L.PushFunction(func(L *lua.State) int {
		role := L.CheckString(1)
		node := tr.GetByRole(role)
		if node == nil {
			L.PushNil()
			return 1
		}
		pushTestVNode(L, node)
		return 1
	})
	L.SetField(-2, "getByRole")

	// getByType(type) -> table or nil
	L.PushFunction(func(L *lua.State) int {
		nodeType := L.CheckString(1)
		node := tr.GetByType(nodeType)
		if node == nil {
			L.PushNil()
			return 1
		}
		pushTestVNode(L, node)
		return 1
	})
	L.SetField(-2, "getByType")

	// fireEvent(target, eventType)
	L.PushFunction(func(L *lua.State) int {
		target := L.CheckString(1)
		eventType := L.CheckString(2)
		tr.FireEvent(target, eventType, nil)
		return 0
	})
	L.SetField(-2, "fireEvent")

	// tostring() -> string
	L.PushFunction(func(L *lua.State) int {
		text := RenderToString(tr.Root())
		L.PushString(text)
		return 1
	})
	L.SetField(-2, "tostring")

	// reset()
	L.PushFunction(func(L *lua.State) int {
		tr.Reset()
		return 0
	})
	L.SetField(-2, "reset")

	return 1
}

func luaTableToMap(L *lua.State, idx int) map[string]any {
	m := make(map[string]any)
	L.PushNil()
	for L.Next(idx) {
		key, ok := L.ToString(-2)
		if ok {
			m[key] = L.ToAny(-1)
		}
		L.Pop(1)
	}
	return m
}

func pushTestVNode(L *lua.State, node *TestVNode) {
	L.NewTable()
	L.PushString(node.Type)
	L.SetField(-2, "type")
	L.PushString(node.Content)
	L.SetField(-2, "content")
	if node.Aria.Role != "" {
		L.PushString(node.Aria.Role)
		L.SetField(-2, "role")
	}
	if node.Aria.Label != "" {
		L.PushString(node.Aria.Label)
		L.SetField(-2, "label")
	}
}
