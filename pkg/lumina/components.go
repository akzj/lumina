// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"sync/atomic"

	"github.com/akzj/go-lua/pkg/lua"
)

var componentID int64

// SelectComponent creates a Select/Dropdown component factory
func SelectComponent(L *lua.State) int {
	// Get props table (argument 1)
	if L.IsNoneOrNil(1) {
		L.NewTable()
	}
	// Stack: [props]

	// Create component factory table
	L.NewTable()
	// Stack: [props, factory]

	// Set fields on factory (at -1)
	L.PushString("Select")
	L.SetField(-1, "name")
	L.PushString("Select")
	L.SetField(-1, "type")
	L.PushBoolean(true)
	L.SetField(-1, "isComponent")
	L.PushValue(1) // copy props
	L.SetField(-1, "props")
	id := fmt.Sprintf("select_%d", atomic.AddInt64(&componentID, 1))
	L.PushString(id)
	L.SetField(-1, "id")
	L.PushFunction(selectRenderFn)
	L.SetField(-1, "render")
	L.PushFunction(selectInitFn)
	L.SetField(-1, "init")

	// Remove original props from stack, return factory
	L.Remove(1)
	return 1
}

func selectInitFn(L *lua.State) int {
	L.NewTable()
	L.PushString("isOpen")
	L.PushBoolean(false)
	L.SetTable(-3)
	L.PushString("selectedIndex")
	L.PushInteger(0)
	L.SetTable(-3)
	return 1
}

func selectRenderFn(L *lua.State) int {
	L.GetField(1, "props")
	defer L.Pop(1)

	options := []string{}
	L.GetField(-1, "options")
	if !L.IsNoneOrNil(-1) {
		L.PushNil()
		for L.Next(-2) {
			if s, _ := L.ToString(-1); s != "" {
				options = append(options, s)
			}
			L.Pop(1)
		}
	}
	L.Pop(1)

	value := "Select..."
	L.GetField(-1, "value")
	if s, _ := L.ToString(-1); s != "" {
		value = s
	}
	L.Pop(1)

	display := value + " ▼"

	L.NewTable()
	L.PushString("type")
	L.PushString("text")
	L.SetTable(-3)
	L.PushString("content")
	L.PushString(display)
	L.SetTable(-3)

	return 1
}

// CheckboxComponent creates a Checkbox component factory
func CheckboxComponent(L *lua.State) int {
	if L.IsNoneOrNil(1) {
		L.NewTable()
	}

	L.NewTable()
	L.PushString("Checkbox")
	L.SetField(-1, "name")
	L.PushString("Checkbox")
	L.SetField(-1, "type")
	L.PushBoolean(true)
	L.SetField(-1, "isComponent")
	L.PushValue(1)
	L.SetField(-1, "props")
	id := fmt.Sprintf("checkbox_%d", atomic.AddInt64(&componentID, 1))
	L.PushString(id)
	L.SetField(-1, "id")
	L.PushFunction(checkboxRenderFn)
	L.SetField(-1, "render")
	L.PushFunction(checkboxInitFn)
	L.SetField(-1, "init")

	L.Remove(1)
	return 1
}

func checkboxInitFn(L *lua.State) int {
	L.NewTable()
	L.PushString("checked")
	L.PushBoolean(false)
	L.SetTable(-3)
	return 1
}

func checkboxRenderFn(L *lua.State) int {
	L.GetField(1, "props")
	defer L.Pop(1)

	checked := false
	L.GetField(-1, "checked")
	if L.ToBoolean(-1) {
		checked = true
	}
	L.Pop(1)

	label := "Checkbox"
	L.GetField(-1, "label")
	if s, _ := L.ToString(-1); s != "" {
		label = s
	}
	L.Pop(1)

	mark := "[ ]"
	if checked {
		mark = "[x]"
	}

	content := mark + " " + label

	L.NewTable()
	L.PushString("type")
	L.PushString("text")
	L.SetTable(-3)
	L.PushString("content")
	L.PushString(content)
	L.SetTable(-3)

	return 1
}

// MenuComponent creates a Menu component factory
func MenuComponent(L *lua.State) int {
	if L.IsNoneOrNil(1) {
		L.NewTable()
	}

	L.NewTable()
	L.PushString("Menu")
	L.SetField(-1, "name")
	L.PushString("Menu")
	L.SetField(-1, "type")
	L.PushBoolean(true)
	L.SetField(-1, "isComponent")
	L.PushValue(1)
	L.SetField(-1, "props")
	id := fmt.Sprintf("menu_%d", atomic.AddInt64(&componentID, 1))
	L.PushString(id)
	L.SetField(-1, "id")
	L.PushFunction(menuRenderFn)
	L.SetField(-1, "render")
	L.PushFunction(menuInitFn)
	L.SetField(-1, "init")

	L.Remove(1)
	return 1
}

func menuInitFn(L *lua.State) int {
	L.NewTable()
	L.PushString("selectedIndex")
	L.PushInteger(-1)
	L.SetTable(-3)
	L.PushString("isOpen")
	L.PushBoolean(false)
	L.SetTable(-3)
	return 1
}

func menuRenderFn(L *lua.State) int {
	L.GetField(1, "props")
	defer L.Pop(1)

	items := []string{}
	L.GetField(-1, "items")
	if !L.IsNoneOrNil(-1) {
		L.PushNil()
		for L.Next(-2) {
			if L.Type(-1) == lua.TypeString {
				if s, _ := L.ToString(-1); s != "" {
					items = append(items, s)
				}
			} else if L.Type(-1) == lua.TypeTable {
				L.GetField(-1, "label")
				if s, _ := L.ToString(-1); s != "" {
					items = append(items, s)
				}
				L.Pop(1)
			}
			L.Pop(1)
		}
	}
	L.Pop(1)

	content := ""
	for _, item := range items {
		content += "  " + item + "\n"
	}

	L.NewTable()
	L.PushString("type")
	L.PushString("vbox")
	L.SetTable(-3)
	L.PushString("content")
	L.PushString(content)
	L.SetTable(-3)

	return 1
}

// TextFieldComponent creates a TextField component factory
func TextFieldComponent(L *lua.State) int {
	if L.IsNoneOrNil(1) {
		L.NewTable()
	}

	L.NewTable()
	L.PushString("TextField")
	L.SetField(-1, "name")
	L.PushString("TextField")
	L.SetField(-1, "type")
	L.PushBoolean(true)
	L.SetField(-1, "isComponent")
	L.PushValue(1)
	L.SetField(-1, "props")
	id := fmt.Sprintf("textfield_%d", atomic.AddInt64(&componentID, 1))
	L.PushString(id)
	L.SetField(-1, "id")
	L.PushFunction(textFieldRenderFn)
	L.SetField(-1, "render")
	L.PushFunction(textFieldInitFn)
	L.SetField(-1, "init")

	L.Remove(1)
	return 1
}

func textFieldInitFn(L *lua.State) int {
	L.NewTable()
	L.PushString("value")
	L.PushString("")
	L.SetTable(-3)
	L.PushString("focused")
	L.PushBoolean(false)
	L.SetTable(-3)
	return 1
}

func textFieldRenderFn(L *lua.State) int {
	L.GetField(1, "props")
	defer L.Pop(1)

	value := ""
	L.GetField(-1, "value")
	if s, _ := L.ToString(-1); s != "" {
		value = s
	}
	L.Pop(1)

	placeholder := "Enter text..."
	L.GetField(-1, "placeholder")
	if s, _ := L.ToString(-1); s != "" {
		placeholder = s
	}
	L.Pop(1)

	display := value
	if display == "" {
		display = placeholder
	}

	content := "[" + display + "]"

	L.NewTable()
	L.PushString("type")
	L.PushString("text")
	L.SetTable(-3)
	L.PushString("content")
	L.PushString(content)
	L.SetTable(-3)

	return 1
}
