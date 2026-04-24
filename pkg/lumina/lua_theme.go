package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// luaUseTheme implements lumina.useTheme() — returns current theme as a Lua table.
func luaUseTheme(L *lua.State) int {
	theme := globalThemeManager.GetTheme()

	// Create theme table
	L.NewTable()

	// Set name
	L.PushString(theme.Name)
	L.SetField(-2, "name")

	// Set colors subtable
	L.NewTable()
	for k, v := range theme.Colors {
		L.PushString(v)
		L.SetField(-2, k)
	}
	L.SetField(-2, "colors")

	// Set space subtable
	L.NewTable()
	for k, v := range theme.Space {
		L.PushInteger(int64(v))
		L.SetField(-2, k)
	}
	L.SetField(-2, "space")

	return 1
}

// luaDefineStyles implements lumina.defineStyles(table).
func luaDefineStyles(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("defineStyles: expected table")
		L.Error()
		return 0
	}

	styles := make(map[string]map[string]string)

	L.PushNil()
	for L.Next(1) {
		className, ok := L.ToString(-2)
		if !ok {
			L.Pop(1)
			continue
		}
		if L.Type(-1) != lua.TypeTable {
			L.Pop(1)
			continue
		}

		props := make(map[string]string)
		L.PushNil()
		for L.Next(-2) {
			if pk, ok := L.ToString(-2); ok {
				if pv, ok2 := L.ToString(-1); ok2 {
					props[pk] = pv
				}
			}
			L.Pop(1)
		}
		styles[className] = props
		L.Pop(1)
	}

	globalThemeManager.DefineStyles(styles)
	return 0
}

// luaGetThemeColor implements lumina.getThemeColor(token) — resolves a theme color token.
func luaGetThemeColor(L *lua.State) int {
	token := L.CheckString(1)
	color := globalThemeManager.GetColor(token)
	L.PushString(color)
	return 1
}
