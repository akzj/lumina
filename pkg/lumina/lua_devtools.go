package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// registerDevToolsModule registers the lumina.devtools subtable.
func registerDevToolsModule(L *lua.State) {
	L.NewTable()

	L.PushFunction(luaDevToolsEnable)
	L.SetField(-2, "enable")

	L.PushFunction(luaDevToolsDisable)
	L.SetField(-2, "disable")

	L.PushFunction(luaDevToolsToggle)
	L.SetField(-2, "toggle")

	L.PushFunction(luaDevToolsIsEnabled)
	L.SetField(-2, "isEnabled")

	L.PushFunction(luaDevToolsIsVisible)
	L.SetField(-2, "isVisible")

	L.PushFunction(luaDevToolsGetTree)
	L.SetField(-2, "getTree")

	L.PushFunction(luaDevToolsGetInspector)
	L.SetField(-2, "getInspector")

	L.PushFunction(luaDevToolsSelect)
	L.SetField(-2, "select")

	L.PushFunction(luaDevToolsSummary)
	L.SetField(-2, "summary")

	// Element Inspector API (uses globalInspector)
	L.PushFunction(luaInspectorToggle)
	L.SetField(-2, "toggleInspector")

	L.PushFunction(luaInspectorIsEnabled)
	L.SetField(-2, "isInspectorEnabled")

	L.PushFunction(luaInspectorSelect)
	L.SetField(-2, "selectElement")

	// Lua DevTools rewrite API
	L.PushFunction(luaDevToolsGetElementTree)
	L.SetField(-2, "getElementTree")

	L.PushFunction(luaDevToolsGetSelectedStyles)
	L.SetField(-2, "getSelectedStyles")

	L.PushFunction(luaDevToolsGetHighlightID)
	L.SetField(-2, "getHighlightID")

	L.PushFunction(luaDevToolsGetSelectedID)
	L.SetField(-2, "getSelectedID")

	L.PushFunction(luaDevToolsGetScrollY)
	L.SetField(-2, "getScrollY")

	L.PushFunction(luaDevToolsSetScrollY)
	L.SetField(-2, "setScrollY")

	L.PushFunction(luaDevToolsGetPanelWidth)
	L.SetField(-2, "getPanelWidth")

	L.SetField(-2, "devtools")
}

func luaDevToolsEnable(L *lua.State) int {
	globalDevTools.Enable()
	return 0
}

func luaDevToolsDisable(L *lua.State) int {
	globalDevTools.Disable()
	return 0
}

func luaDevToolsToggle(L *lua.State) int {
	globalDevTools.Toggle()
	return 0
}

func luaDevToolsIsEnabled(L *lua.State) int {
	L.PushBoolean(globalDevTools.IsEnabled())
	return 1
}

func luaDevToolsIsVisible(L *lua.State) int {
	L.PushBoolean(globalDevTools.IsVisible())
	return 1
}

func luaDevToolsGetTree(L *lua.State) int {
	tree := globalDevTools.RenderTree()
	L.PushString(tree)
	return 1
}

func luaDevToolsGetInspector(L *lua.State) int {
	inspector := globalDevTools.RenderInspector()
	L.PushString(inspector)
	return 1
}

func luaDevToolsSelect(L *lua.State) int {
	id := L.CheckString(1)
	globalDevTools.SetSelected(id)
	return 0
}

func luaDevToolsSummary(L *lua.State) int {
	summary := globalDevTools.Summary()
	L.PushAny(summary)
	return 1
}

// --- Element Inspector API ---

func luaInspectorToggle(L *lua.State) int {
	ToggleInspector()
	return 0
}

func luaInspectorIsEnabled(L *lua.State) int {
	L.PushBoolean(IsInspectorVisible())
	return 1
}

func luaInspectorSelect(L *lua.State) int {
	id := L.CheckString(1)
	SetInspectorSelected(id)
	return 0
}

// --- Lua DevTools rewrite API ---

// luaDevToolsGetElementTree returns the VNode tree as a structured Lua table.
// Each node: { type="vbox", id="my-id", x=0, y=0, w=80, h=24, children={...} }
func luaDevToolsGetElementTree(L *lua.State) int {
	app := GetApp(L)
	if app == nil {
		L.NewTable()
		return 1
	}
	rootVNode := app.findRootVNode()
	if rootVNode == nil {
		L.NewTable()
		return 1
	}
	tree := vnodeToMap(rootVNode, 0)
	L.PushAny(tree)
	return 1
}

// vnodeToMap converts a VNode tree to a map structure for Lua.
// maxDepth prevents infinite recursion.
func vnodeToMap(vnode *VNode, depth int) map[string]any {
	if vnode == nil || depth > 50 {
		return nil
	}

	id := ""
	if idVal, ok := vnode.Props["id"].(string); ok {
		id = idVal
	}

	node := map[string]any{
		"type": vnode.Type,
		"id":   id,
		"x":    vnode.X,
		"y":    vnode.Y,
		"w":    vnode.W,
		"h":    vnode.H,
	}

	if vnode.Content != "" {
		node["content"] = vnode.Content
	}

	if len(vnode.Children) > 0 {
		children := make([]any, 0, len(vnode.Children))
		for _, child := range vnode.Children {
			if m := vnodeToMap(child, depth+1); m != nil {
				children = append(children, m)
			}
		}
		node["children"] = children
	}

	return node
}

// luaDevToolsGetSelectedStyles returns style info for the selected/highlighted element.
func luaDevToolsGetSelectedStyles(L *lua.State) int {
	selectedID := globalInspector.selectedID
	if selectedID == "" {
		selectedID = globalInspector.highlightID
	}
	if selectedID == "" {
		L.PushAny(map[string]any{"selected": false})
		return 1
	}

	app := GetApp(L)
	if app == nil {
		L.PushAny(map[string]any{"selected": false})
		return 1
	}

	rootVNode := app.findRootVNode()
	if rootVNode == nil {
		L.PushAny(map[string]any{"selected": false})
		return 1
	}

	node := findVNodeByID(rootVNode, selectedID)
	if node == nil {
		L.PushAny(map[string]any{"selected": false, "id": selectedID, "error": "not found"})
		return 1
	}

	result := map[string]any{
		"selected": true,
		"id":       selectedID,
		"type":     node.Type,
		"x":        node.X,
		"y":        node.Y,
		"w":        node.W,
		"h":        node.H,
	}

	if node.Content != "" {
		content := node.Content
		if len(content) > 50 {
			content = content[:47] + "..."
		}
		result["content"] = content
	}

	// Style info
	s := node.Style
	style := map[string]any{}
	if s.Position != "" {
		style["position"] = s.Position
	}
	if s.Flex > 0 {
		style["flex"] = s.Flex
	}
	if s.Width > 0 {
		style["width"] = s.Width
	}
	if s.Height > 0 {
		style["height"] = s.Height
	}
	if s.Padding > 0 {
		style["padding"] = s.Padding
	}
	if s.Margin > 0 {
		style["margin"] = s.Margin
	}
	if s.Border != "" {
		style["border"] = s.Border
	}
	if s.Foreground != "" {
		style["fg"] = s.Foreground
	}
	if s.Background != "" {
		style["bg"] = s.Background
	}
	if s.Bold {
		style["bold"] = true
	}
	if s.Overflow != "" {
		style["overflow"] = s.Overflow
	}
	if s.ZIndex != 0 {
		style["zIndex"] = s.ZIndex
	}
	if s.Justify != "" {
		style["justify"] = s.Justify
	}
	if s.Align != "" {
		style["align"] = s.Align
	}
	result["style"] = style

	L.PushAny(result)
	return 1
}

func luaDevToolsGetHighlightID(L *lua.State) int {
	L.PushString(globalInspector.highlightID)
	return 1
}

func luaDevToolsGetSelectedID(L *lua.State) int {
	L.PushString(globalInspector.selectedID)
	return 1
}

func luaDevToolsGetScrollY(L *lua.State) int {
	L.PushInteger(int64(globalInspector.scrollY))
	return 1
}

func luaDevToolsSetScrollY(L *lua.State) int {
	y := int(L.CheckInteger(1))
	globalInspector.scrollY = y
	return 0
}

func luaDevToolsGetPanelWidth(L *lua.State) int {
	L.PushInteger(int64(globalInspector.panelWidth))
	return 1
}
