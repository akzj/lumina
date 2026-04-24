package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// Global style registry
var globalStyles = make(map[string]map[string]any)

// defineStyle(name, styleTable or styleFn) — define a named style
func defineStyle(L *lua.State) int {
	name := L.CheckString(1)
	// Style can be a table or a function (theme-aware)
	if L.Type(2) == lua.TypeFunction {
		// Store as function reference
		refID := L.Ref(lua.RegistryIndex)
		globalStyles[name] = map[string]any{"_fn": refID, "_isFn": true}
	} else if L.Type(2) == lua.TypeTable {
		if m, ok := L.ToMap(2); ok {
			globalStyles[name] = m
		}
	} else {
		L.PushString("defineStyle: expected table or function")
		L.Error()
	}
	return 0
}

// defineGlobalStyles(stylesTable) — define multiple styles at once
func defineGlobalStyles(L *lua.State) int {
	if styles, ok := L.ToMap(1); ok {
		for k, v := range styles {
			if m, ok := v.(map[string]any); ok {
				globalStyles[k] = m
			}
		}
	}
	return 0
}

// getStyle(name) → styleTable — get a named style
func getStyle(L *lua.State) int {
	name := L.CheckString(1)
	if style, ok := globalStyles[name]; ok {
		// Convert back to Lua table using PushAny
		L.PushAny(style)
	} else {
		L.PushNil()
	}
	return 1
}

// defineTheme(name, themeTable) → theme
func defineTheme(L *lua.State) int {
	name := L.CheckString(1)
	if L.Type(2) != lua.TypeTable {
		L.PushString("defineTheme: expected table")
		L.Error()
	}

	theme := &Theme{Name: name}
	theme.Colors = make(map[string]string)
	theme.Spacing = make(map[string]int)
	theme.Borders = make(map[string]string)

	// Extract colors
	L.GetField(2, "colors")
	if L.Type(-1) == lua.TypeTable {
		L.ForEach(-1, func(L *lua.State) bool {
			if k, _ := L.ToString(-2); k != "" {
				if v, _ := L.ToString(-1); v != "" {
					theme.Colors[k] = v
				}
			}
			return true
		})
	}
	L.Pop(1)

	// Extract spacing
	L.GetField(2, "spacing")
	if L.Type(-1) == lua.TypeTable {
		L.ForEach(-1, func(L *lua.State) bool {
			if k, _ := L.ToString(-2); k != "" {
				if v, ok := L.ToInteger(-1); ok {
					theme.Spacing[k] = int(v)
				}
			}
			return true
		})
	}
	L.Pop(1)

	// Extract borders
	L.GetField(2, "borders")
	if L.Type(-1) == lua.TypeTable {
		L.ForEach(-1, func(L *lua.State) bool {
			if k, _ := L.ToString(-2); k != "" {
				if v, _ := L.ToString(-1); v != "" {
					theme.Borders[k] = v
				}
			}
			return true
		})
	}
	L.Pop(1)

	// Register as named theme
	RegisterTheme(name, theme)

	// Push theme object as Lua table
	L.NewTableFrom(map[string]any{
		"name":   name,
		"colors": theme.Colors,
	})

	return 1
}

// setTheme(name or themeTable) — set current theme
func setTheme(L *lua.State) int {
	if L.Type(1) == lua.TypeString {
		name := L.CheckString(1)
		t := GetThemeByName(name)
		if t == nil {
			L.PushString("setTheme: unknown theme '" + name + "'")
			L.Error()
			return 0
		}
		SetTheme(t)
	} else if L.Type(1) == lua.TypeTable {
		// Build custom theme from table
		theme := &Theme{
			Colors:  make(map[string]string),
			Space:   map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
			Spacing: map[string]int{"xs": 0, "sm": 1, "md": 2, "lg": 3, "xl": 4},
			Borders: map[string]string{"none": "", "rounded": "rounded", "double": "double", "single": "single"},
		}
		// Read name
		L.GetField(1, "name")
		if s, ok := L.ToString(-1); ok {
			theme.Name = s
		} else {
			theme.Name = "custom"
		}
		L.Pop(1)
		// Read colors
		L.GetField(1, "colors")
		if L.Type(-1) == lua.TypeTable {
			L.PushNil()
			for L.Next(-2) {
				if k, ok := L.ToString(-2); ok {
					if v, ok2 := L.ToString(-1); ok2 {
						theme.Colors[k] = v
					}
				}
				L.Pop(1)
			}
		}
		L.Pop(1)
		// Read spacing
		L.GetField(1, "spacing")
		if L.Type(-1) == lua.TypeTable {
			L.PushNil()
			for L.Next(-2) {
				if k, ok := L.ToString(-2); ok {
					if v, ok2 := L.ToInteger(-1); ok2 {
						theme.Spacing[k] = int(v)
						theme.Space[k] = int(v)
					}
				}
				L.Pop(1)
			}
		}
		L.Pop(1)
		// Read borders
		L.GetField(1, "borders")
		if L.Type(-1) == lua.TypeTable {
			L.PushNil()
			for L.Next(-2) {
				if k, ok := L.ToString(-2); ok {
					if v, ok2 := L.ToString(-1); ok2 {
						theme.Borders[k] = v
					}
				}
				L.Pop(1)
			}
		}
		L.Pop(1)
		SetTheme(theme)
	}
	return 0
}
