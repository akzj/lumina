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
	all := a.manager.GetAll()
	focused := a.FocusedID()
	result := make([]mcp.ComponentInfo, 0, len(all))
	for _, c := range all {
		if c.ID() == "__devtools" {
			continue
		}
		r := c.Rect()
		result = append(result, mcp.ComponentInfo{
			ID:      c.ID(),
			Name:    c.Name(),
			Focused: c.ID() == focused,
			Rect:    [4]int{r.X, r.Y, r.W, r.H},
		})
	}
	return result
}

// MCPInspectComponent returns detailed info for a single component.
func (a *App) MCPInspectComponent(id string) (*mcp.ComponentDetail, error) {
	c := a.manager.Get(id)
	if c == nil {
		return nil, fmt.Errorf("component not found: %s", id)
	}
	r := c.Rect()
	return &mcp.ComponentDetail{
		ID:      c.ID(),
		Name:    c.Name(),
		State:   c.State(),
		Focused: c.ID() == a.FocusedID(),
		Dirty:   c.IsDirtyPaint(),
		Rect:    [4]int{r.X, r.Y, r.W, r.H},
		ZIndex:  c.ZIndex(),
	}, nil
}

// MCPGetState returns component state. If key is empty, returns the full state map.
// If key is non-empty, returns the value for that key.
func (a *App) MCPGetState(compID, key string) (any, error) {
	c := a.manager.Get(compID)
	if c == nil {
		return nil, fmt.Errorf("component not found: %s", compID)
	}
	if key == "" {
		return c.State(), nil
	}
	val, ok := c.State()[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

// MCPSetState sets a state key on a component and marks it dirty.
func (a *App) MCPSetState(compID, key string, value any) error {
	c := a.manager.Get(compID)
	if c == nil {
		return fmt.Errorf("component not found: %s", compID)
	}
	a.SetState(compID, key, value)
	return nil
}

// MCPSimulateClick dispatches a click event targeting the given VNode ID.
func (a *App) MCPSimulateClick(id string) error {
	e := &event.Event{Type: "click", Target: id, Bubbles: true}
	a.dispatcher.Dispatch(e)
	return nil
}

// MCPSimulateKey dispatches a keydown event with the given key name.
func (a *App) MCPSimulateKey(key string) error {
	e := &event.Event{Type: "keydown", Key: key}
	a.HandleEvent(e)
	return nil
}

// MCPEval executes Lua code in the app's Lua state.
// Returns {"ok": true} on success, or an error if the code fails.
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
	a.dispatcher.FocusNext()
	return a.dispatcher.FocusedID()
}

// MCPFocusPrev moves focus to the previous focusable VNode and returns the new focused ID.
func (a *App) MCPFocusPrev() string {
	a.dispatcher.FocusPrev()
	return a.dispatcher.FocusedID()
}

// MCPSetFocus sets focus to a specific VNode by ID.
func (a *App) MCPSetFocus(id string) {
	a.dispatcher.SetFocus(id)
}

// MCPGetFocusableIDs returns the ordered list of focusable VNode IDs.
func (a *App) MCPGetFocusableIDs() []string {
	return a.dispatcher.GetFocusableIDs()
}

// MCPGetFocusedID returns the currently focused VNode ID.
func (a *App) MCPGetFocusedID() string {
	return a.FocusedID()
}

// MCPToggleDevTools toggles the devtools panel and returns the new visibility.
func (a *App) MCPToggleDevTools() bool {
	a.toggleDevTools()
	return a.devtools.Visible
}

// MCPGetScreenText reads the screen buffer and returns it as a text string.
// Each row is one line, terminated by '\n'. Zero-char cells become spaces.
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
