package v2

import (
	"fmt"
	"strings"

	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/mcp"
)

// Verify at compile time that *App implements mcp.AppInspector.
var _ mcp.AppInspector = (*App)(nil)

// MCPInspectTree returns a summary of all registered components.
func (a *App) MCPInspectTree() []mcp.ComponentInfo {
	var result []mcp.ComponentInfo
	root := a.engine.Root()
	if root == nil {
		return result
	}
	// Walk the engine's component tree.
	for id, comp := range a.engine.AllComponents() {
		result = append(result, mcp.ComponentInfo{
			ID:   id,
			Name: comp.Name,
		})
	}
	return result
}

// MCPInspectComponent returns detailed info for a single component.
func (a *App) MCPInspectComponent(id string) (*mcp.ComponentDetail, error) {
	comp := a.engine.GetComponent(id)
	if comp == nil {
		return nil, fmt.Errorf("component not found: %s", id)
	}
	return &mcp.ComponentDetail{
		ID:   comp.ID,
		Name: comp.Name,
	}, nil
}

// MCPGetState returns component state.
func (a *App) MCPGetState(compID, key string) (any, error) {
	comp := a.engine.GetComponent(compID)
	if comp == nil {
		return nil, fmt.Errorf("component not found: %s", compID)
	}
	if key == "" {
		return comp.State, nil
	}
	val, ok := comp.State[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

// MCPSetState sets a state key on a component and marks it dirty.
func (a *App) MCPSetState(compID, key string, value any) error {
	a.engine.SetState(compID, key, value)
	return nil
}

// MCPSimulateClick dispatches a click event at the given coordinates.
func (a *App) MCPSimulateClick(id string) error {
	// The engine handles click by coordinates. For ID-based click,
	// we dispatch through HandleEvent.
	a.HandleEvent(&event.Event{Type: "click", Target: id, Bubbles: true})
	return nil
}

// MCPSimulateKey dispatches a keydown event with the given key name.
func (a *App) MCPSimulateKey(key string) error {
	a.HandleEvent(&event.Event{Type: "keydown", Key: key})
	return nil
}

// MCPEval executes Lua code in the app's Lua state.
func (a *App) MCPEval(code string) (any, error) {
	if a.luaState == nil {
		return nil, fmt.Errorf("no Lua state available")
	}
	if err := a.luaState.DoString(code); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

// MCPFocusNext moves focus to the next focusable VNode and returns the new focused ID.
func (a *App) MCPFocusNext() string {
	// TODO: Implement focus management in engine.
	return ""
}

// MCPFocusPrev moves focus to the previous focusable VNode and returns the new focused ID.
func (a *App) MCPFocusPrev() string {
	// TODO: Implement focus management in engine.
	return ""
}

// MCPSetFocus sets focus to a specific VNode by ID.
func (a *App) MCPSetFocus(id string) {
	// TODO: Implement focus management in engine.
}

// MCPGetFocusableIDs returns the ordered list of focusable VNode IDs.
func (a *App) MCPGetFocusableIDs() []string {
	// TODO: Implement focus management in engine.
	return nil
}

// MCPGetFocusedID returns the currently focused VNode ID.
func (a *App) MCPGetFocusedID() string {
	return a.FocusedID()
}

// MCPToggleDevTools toggles the devtools panel and returns the new visibility.
func (a *App) MCPToggleDevTools() bool {
	a.toggleDevToolsV2()
	return a.devtools.Visible
}

// MCPGetScreenText reads the screen buffer and returns it as a text string.
func (a *App) MCPGetScreenText() string {
	buf := a.Screen()
	if buf == nil {
		return ""
	}
	var sb strings.Builder
	sb.Grow(buf.Width()*buf.Height() + buf.Height())
	for y := 0; y < buf.Height(); y++ {
		for x := 0; x < buf.Width(); x++ {
			cell := buf.Get(x, y)
			if cell.Char == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Char)
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// MCPGetVersion returns the Lumina v2 version string.
func (a *App) MCPGetVersion() string {
	return "lumina-v2"
}
