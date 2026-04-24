// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"sync/atomic"

	"github.com/akzj/go-lua/pkg/lua"
)

var componentID int64

// SelectComponent creates a Select/Dropdown component factory.
func SelectComponent(L *lua.State) int {
	if L.IsNoneOrNil(1) {
		L.NewTable()
	}
	id := fmt.Sprintf("select_%d", atomic.AddInt64(&componentID, 1))
	L.NewTableFrom(map[string]any{
		"name":        "Select",
		"type":        "Select",
		"isComponent": true,
		"id":          id,
		"render":      lua.Function(selectRenderFn),
		"init":        lua.Function(selectInitFn),
		"props":       lua.StackRef{Index: 1},
	})
	L.Remove(1)
	return 1
}

func selectInitFn(L *lua.State) int {
	L.NewTableFrom(map[string]any{
		"isOpen":        false,
		"selectedIndex": int64(0),
	})
	return 1
}

func selectRenderFn(L *lua.State) int {
	L.GetField(1, "props")
	props, _ := L.ToMap(-1)
	L.Pop(1)

	value := "Select..."
	if v, ok := props["value"].(string); ok && v != "" {
		value = v
	}

	L.NewTableFrom(map[string]any{
		"type":    "text",
		"content": value + " ▼",
	})
	return 1
}

// CheckboxComponent creates a Checkbox component factory.
func CheckboxComponent(L *lua.State) int {
	if L.IsNoneOrNil(1) {
		L.NewTable()
	}
	id := fmt.Sprintf("checkbox_%d", atomic.AddInt64(&componentID, 1))
	L.NewTableFrom(map[string]any{
		"name":        "Checkbox",
		"type":        "Checkbox",
		"isComponent": true,
		"id":          id,
		"render":      lua.Function(checkboxRenderFn),
		"init":        lua.Function(checkboxInitFn),
		"props":       lua.StackRef{Index: 1},
	})
	L.Remove(1)
	return 1
}

func checkboxInitFn(L *lua.State) int {
	L.NewTableFrom(map[string]any{
		"checked": false,
	})
	return 1
}

func checkboxRenderFn(L *lua.State) int {
	L.GetField(1, "props")
	props, _ := L.ToMap(-1)
	L.Pop(1)

	checked := false
	if v, ok := props["checked"].(bool); ok {
		checked = v
	}

	label := "Checkbox"
	if v, ok := props["label"].(string); ok && v != "" {
		label = v
	}

	mark := "[ ]"
	if checked {
		mark = "[x]"
	}

	L.NewTableFrom(map[string]any{
		"type":    "text",
		"content": mark + " " + label,
	})
	return 1
}

// MenuComponent creates a Menu component factory.
func MenuComponent(L *lua.State) int {
	if L.IsNoneOrNil(1) {
		L.NewTable()
	}
	id := fmt.Sprintf("menu_%d", atomic.AddInt64(&componentID, 1))
	L.NewTableFrom(map[string]any{
		"name":        "Menu",
		"type":        "Menu",
		"isComponent": true,
		"id":          id,
		"render":      lua.Function(menuRenderFn),
		"init":        lua.Function(menuInitFn),
		"props":       lua.StackRef{Index: 1},
	})
	L.Remove(1)
	return 1
}

func menuInitFn(L *lua.State) int {
	L.NewTableFrom(map[string]any{
		"selectedIndex": int64(-1),
		"isOpen":        false,
	})
	return 1
}

func menuRenderFn(L *lua.State) int {
	L.GetField(1, "props")
	props, _ := L.ToMap(-1)
	L.Pop(1)

	var items []string
	if raw, ok := props["items"]; ok {
		switch v := raw.(type) {
		case []any:
			for _, item := range v {
				switch it := item.(type) {
				case string:
					if it != "" {
						items = append(items, it)
					}
				case map[string]any:
					if label, ok := it["label"].(string); ok && label != "" {
						items = append(items, label)
					}
				}
			}
		}
	}

	content := ""
	for _, item := range items {
		content += "  " + item + "\n"
	}

	L.NewTableFrom(map[string]any{
		"type":    "vbox",
		"content": content,
	})
	return 1
}

// TextFieldComponent creates a TextField component factory.
func TextFieldComponent(L *lua.State) int {
	if L.IsNoneOrNil(1) {
		L.NewTable()
	}
	id := fmt.Sprintf("textfield_%d", atomic.AddInt64(&componentID, 1))
	L.NewTableFrom(map[string]any{
		"name":        "TextField",
		"type":        "TextField",
		"isComponent": true,
		"id":          id,
		"render":      lua.Function(textFieldRenderFn),
		"init":        lua.Function(textFieldInitFn),
		"props":       lua.StackRef{Index: 1},
	})
	L.Remove(1)
	return 1
}

func textFieldInitFn(L *lua.State) int {
	L.NewTableFrom(map[string]any{
		"value":   "",
		"focused": false,
	})
	return 1
}

func textFieldRenderFn(L *lua.State) int {
	L.GetField(1, "props")
	props, _ := L.ToMap(-1)
	L.Pop(1)

	value := ""
	if v, ok := props["value"].(string); ok && v != "" {
		value = v
	}

	placeholder := "Enter text..."
	if v, ok := props["placeholder"].(string); ok && v != "" {
		placeholder = v
	}

	display := value
	if display == "" {
		display = placeholder
	}

	L.NewTableFrom(map[string]any{
		"type":    "text",
		"content": "[" + display + "]",
	})
	return 1
}
