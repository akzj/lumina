package lumina

import (
	"encoding/json"
	"fmt"

	"github.com/akzj/go-lua/pkg/lua"
)

// HandleMCPRequest handles an MCP request and returns a response.
func (app *App) HandleMCPRequest(req MCPRequest) MCPResponse {
	var result interface{}
	var errMsg string

	switch req.Method {
	case "inspectTree":
		result = app.mcpInspectTree()
	case "inspectComponent":
		var params struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpInspectComponent(params.ID)
		} else {
			errMsg = "invalid params for inspectComponent"
		}
	case "inspectStyles":
		var params struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpInspectStyles(params.ID)
		} else {
			errMsg = "invalid params for inspectStyles"
		}
	case "simulateClick":
		var params struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpSimulateClick(params.ID)
		} else {
			errMsg = "invalid params for simulateClick"
		}
	case "eval":
		var params struct {
			Code string `json:"code"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpEval(params.Code)
		} else {
			errMsg = "invalid params for eval"
		}
	case "getState":
		var params struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpGetState(params.ID, params.Key)
		} else {
			errMsg = "invalid params for getState"
		}
	case "setState":
		var params struct {
			ID    string      `json:"id"`
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpSetState(params.ID, params.Key, params.Value)
		} else {
			errMsg = "invalid params for setState"
		}
	case "focusNext":
		app.mcpFocusNext()
		result = map[string]string{"focused": globalEventBus.GetFocused()}
	case "focusPrev":
		app.mcpFocusPrev()
		result = map[string]string{"focused": globalEventBus.GetFocused()}
	case "getFocusableIDs":
		result = map[string]interface{}{"ids": globalEventBus.GetFocusableIDs()}
	case "getFrame":
		result = app.mcpGetFrame()
	case "debug.toggleInspector":
		ToggleInspector()
		// Force re-render so the panel appears/disappears immediately.
		for _, c := range globalRegistry.components {
			c.Dirty.Store(true)
		}
		result = map[string]any{"visible": IsInspectorVisible()}
	case "debug.checkInspectorBounds":
		result = app.mcpCheckInspectorBounds()
	case "getVersion":
		result = map[string]string{"version": ModuleName}
	default:
		errMsg = "unknown method: " + req.Method
	}

	if errMsg != "" {
		return MCPResponse{
			ID:    req.ID,
			Error: &MCPError{Code: -32601, Message: errMsg},
		}
	}

	return MCPResponse{
		ID:     req.ID,
		Result: result,
	}
}

// mcpInspectTree returns the component tree.
// mcpInspectTree returns the component tree.
func (app *App) mcpInspectTree() map[string]interface{} {
	tree := []map[string]interface{}{}

	for id, comp := range globalRegistry.components {
		tree = append(tree, map[string]interface{}{
			"id":      id,
			"type":    comp.Type,
			"name":    comp.Name,
			"focused": id == globalEventBus.GetFocused(),
		})
	}

	return map[string]interface{}{
		"tree":         tree,
		"focusedID":    globalEventBus.GetFocused(),
		"focusableIDs": globalEventBus.GetFocusableIDs(),
	}
}

// mcpInspectComponent returns details of a specific component.
// mcpInspectComponent returns details of a specific component.
func (app *App) mcpInspectComponent(id string) map[string]interface{} {
	comp, ok := GetComponentByID(id)
	if !ok {
		return map[string]interface{}{"error": "component not found"}
	}


	return map[string]interface{}{
		"id":      comp.ID,
		"type":    comp.Type,
		"name":    comp.Name,
		"state":   comp.State,
		"props":   comp.Props,
		"focused": id == globalEventBus.GetFocused(),
		"dirty":   comp.Dirty.Load(),
	}
}

// mcpInspectStyles returns computed styles for a component.
// mcpInspectStyles returns computed styles for a component.
func (app *App) mcpInspectStyles(id string) map[string]interface{} {
	comp, ok := GetComponentByID(id)
	if !ok {
		return map[string]interface{}{"error": "component not found"}
	}

	return map[string]interface{}{
		"id":     id,
		"styles": comp.Props,
	}
}

// mcpSimulateClick simulates a click on a component.
// mcpSimulateClick simulates a click on a component.
func (app *App) mcpSimulateClick(id string) map[string]interface{} {
	globalEventBus.Emit(&Event{
		Type:    "click",
		Target:  id,
		Bubbles: true,
	})
	return map[string]interface{}{"clicked": id}
}

// mcpEval evaluates Lua code.
// mcpEval evaluates Lua code.
func (app *App) mcpEval(code string) map[string]interface{} {
	app.L.GetGlobal("lumina")
	app.L.SetGlobal("lumina")

	if err := app.L.DoString(code); err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	n := app.L.GetTop()
	if n == 0 {
		return map[string]interface{}{"ok": true}
	}

	results := make([]interface{}, n)
	for i := 1; i <= n; i++ {
		results[i-1] = luaValueToInterface(app.L, i)
	}
	app.L.Pop(n)

	return map[string]interface{}{"results": results}
}

func luaValueToInterface(L *lua.State, index int) interface{} {
	switch L.Type(index) {
	case lua.TypeString:
		if v, ok := L.ToString(index); ok {
			return v
		}
		return nil
	case lua.TypeNumber:
		if L.IsInteger(index) {
			v, _ := L.ToInteger(index)
			return v
		}
		v, _ := L.ToNumber(index)
		return v
	case lua.TypeBoolean:
		return L.ToBoolean(index)
	case lua.TypeTable:
		return "table"
	case lua.TypeNil:
		return nil
	default:
		return fmt.Sprintf("unknown(%s)", L.TypeName(L.Type(index)))
	}
}

// mcpGetState returns component state.
// mcpGetState returns component state.
func (app *App) mcpGetState(id, key string) map[string]interface{} {
	comp, ok := GetComponentByID(id)
	if !ok {
		return map[string]interface{}{"error": "component not found"}
	}

	if key != "" {
		value, exists := comp.GetState(key)
		if !exists {
			return map[string]interface{}{"error": "key not found"}
		}
		return map[string]interface{}{"value": value}
	}

	return map[string]interface{}{"state": comp.State}
}

// mcpSetState sets component state.
// mcpSetState sets component state.
func (app *App) mcpSetState(id, key string, value interface{}) map[string]interface{} {
	comp, ok := GetComponentByID(id)
	if !ok {
		return map[string]interface{}{"error": "component not found"}
	}

	comp.SetState(key, value)
	return map[string]interface{}{"ok": true}
}

// mcpFocusNext moves focus to next component.
// mcpFocusNext moves focus to next component.
func (app *App) mcpFocusNext() {
	globalEventBus.FocusNext()
}

// mcpFocusPrev moves focus to previous component.
// mcpFocusPrev moves focus to previous component.
func (app *App) mcpFocusPrev() {
	globalEventBus.FocusPrev()
}

// mcpGetFrame returns the current frame.
// mcpGetFrame returns the current frame.
func (app *App) mcpGetFrame() map[string]interface{} {
	return map[string]interface{}{
		"focusedID":      globalEventBus.GetFocused(),
		"componentCount": len(globalRegistry.components),
	}
}

// mcpCheckInspectorBounds checks if inspector right border has unexpected chars.
// mcpCheckInspectorBounds checks if inspector right border has unexpected chars.
func (app *App) mcpCheckInspectorBounds() map[string]any {
	w, h := app.getWidth(), app.getHeight()
	if !IsInspectorVisible() {
		return map[string]any{"ok": false, "reason": "inspector not visible"}
	}
	if app.lastFrame == nil || len(app.lastFrame.Cells) == 0 {
		return map[string]any{"ok": false, "reason": "no lastFrame available yet"}
	}

	panelW := globalInspector.panelWidth
	if panelW > w/2 {
		panelW = w / 2
	}
	if panelW < 1 {
		return map[string]any{"ok": false, "reason": "panelW < 1", "w": w, "h": h}
	}
	rightX := w - 1
	if rightX >= app.lastFrame.Width {
		rightX = app.lastFrame.Width - 1
	}

	allowed := map[rune]bool{
		'│': true, '┐': true, '┘': true, '┤': true, ' ': true,
	}

	type borderCell struct {
		X    int    `json:"x"`
		Y    int    `json:"y"`
		Char string `json:"char"`
	}

	var violations []borderCell
	maxY := h
	if maxY > app.lastFrame.Height {
		maxY = app.lastFrame.Height
	}

	for y := 0; y < maxY; y++ {
		c := app.lastFrame.Cells[y][rightX]
		if !allowed[c.Char] && !c.Transparent {
			violations = append(violations, borderCell{X: rightX, Y: y, Char: string(c.Char)})
		}
	}

	return map[string]any{
		"ok":              len(violations) == 0,
		"w":               w,
		"h":               h,
		"panelW":          panelW,
		"rightBorderX":    rightX,
		"violationCount":  len(violations),
		"violations":      violations,
	}
}

// clearScreenWithBg returns an ANSI escape sequence that sets the theme
// background color before clearing the screen. This ensures the cleared
// area uses our theme color instead of the terminal's default background.
