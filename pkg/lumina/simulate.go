package lumina

import (
	"errors"

	"github.com/akzj/go-lua/pkg/lua"
)

// SimulatedEvent represents a simulated user interaction.
type SimulatedEvent struct {
	Type      string
	Component string
	X, Y      int    // coordinates within component
	Key       string // for keyboard events
	Value     string // for change/input events
	Modifiers EventModifiers
}

// Error definitions
var (
	ErrComponentNotFound = errors.New("component not found")
	ErrNoHandler         = errors.New("no handler for event type")
)

// Simulate executes a simulated event on a component.
func Simulate(event *SimulatedEvent) error {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	// Find target component
	var comp *Component
	if compI, ok := globalRegistry.components[event.Component]; ok {
		comp = compI
	} else {
		// Try to find by name or type
		for _, c := range globalRegistry.components {
			if c.Name == event.Component || c.Type == event.Component {
				comp = c
				break
			}
		}
	}

	if comp == nil {
		return ErrComponentNotFound
	}

	// Create internal Event
	e := &Event{
		Type:      event.Type,
		Target:    comp.ID,
		X:         event.X,
		Y:         event.Y,
		Key:       event.Key,
		Modifiers: event.Modifiers,
		Timestamp: CurrentTimestamp(),
	}

	// Dispatch via event bus
	globalEventBus.Emit(e)

	return nil
}

// SimulateClick simulates a click on a component.
func SimulateClick(target string) error {
	return Simulate(&SimulatedEvent{
		Type:      "click",
		Component: target,
	})
}

// SimulateKey simulates a keyboard event.
func SimulateKey(key string, modifiers EventModifiers) error {
	return Simulate(&SimulatedEvent{
		Type:      "keydown",
		Key:       key,
		Modifiers: modifiers,
	})
}

// SimulateChange simulates a change event (e.g., input value change).
func SimulateChange(target, value string) error {
	return Simulate(&SimulatedEvent{
		Type:      "change",
		Component: target,
		Value:     value,
	})
}

// SimulateFocus simulates focus on a component.
func SimulateFocus(target string) error {
	return Simulate(&SimulatedEvent{
		Type:      "focus",
		Component: target,
	})
}

// Lua API implementations

// simulate(eventType, target, options) → success, error?
func simulate(L *lua.State) int {
	eventType := L.CheckString(1)
	target := L.CheckString(2)

	event := &SimulatedEvent{
		Type:      eventType,
		Component: target,
	}

	// Parse options table (optional 3rd arg)
	if L.GetTop() >= 3 && L.Type(3) == lua.TypeTable {
		opts := tableToMap(L, 3)

		if x, ok := opts["x"].(int64); ok {
			event.X = int(x)
		}
		if y, ok := opts["y"].(int64); ok {
			event.Y = int(y)
		}
		if key, ok := opts["key"].(string); ok {
			event.Key = key
		}
		if value, ok := opts["value"].(string); ok {
			event.Value = value
		}

		if mods, ok := opts["modifiers"].(map[string]any); ok {
			if ctrl, ok := mods["ctrl"].(bool); ok {
				event.Modifiers.Ctrl = ctrl
			}
			if shift, ok := mods["shift"].(bool); ok {
				event.Modifiers.Shift = shift
			}
			if alt, ok := mods["alt"].(bool); ok {
				event.Modifiers.Alt = alt
			}
			if meta, ok := mods["meta"].(bool); ok {
				event.Modifiers.Meta = meta
			}
		}
	}

	err := Simulate(event)
	if err != nil {
		L.PushBoolean(false)
		L.PushString(err.Error())
		return 2
	}

	L.PushBoolean(true)
	return 1
}

// simulateClick(component) → success
func simulateClick(L *lua.State) int {
	target := L.CheckString(1)

	err := SimulateClick(target)
	if err != nil {
		L.PushBoolean(false)
		return 1
	}

	L.PushBoolean(true)
	return 1
}

// simulateKey(key, options) → success
func simulateKey(L *lua.State) int {
	key := L.CheckString(1)

	event := &SimulatedEvent{
		Type: "keydown",
		Key:  key,
	}

	if L.GetTop() >= 2 && L.Type(2) == lua.TypeTable {
		opts := tableToMap(L, 2)
		if mods, ok := opts["modifiers"].(map[string]any); ok {
			if ctrl, ok := mods["ctrl"].(bool); ok {
				event.Modifiers.Ctrl = ctrl
			}
			if shift, ok := mods["shift"].(bool); ok {
				event.Modifiers.Shift = shift
			}
			if alt, ok := mods["alt"].(bool); ok {
				event.Modifiers.Alt = alt
			}
		}
	}

	err := Simulate(event)
	if err != nil {
		L.PushBoolean(false)
		return 1
	}

	L.PushBoolean(true)
	return 1
}

// simulateChange(component, value) → success
func simulateChange(L *lua.State) int {
	target := L.CheckString(1)
	value := L.CheckString(2)

	err := SimulateChange(target, value)
	if err != nil {
		L.PushBoolean(false)
		return 1
	}

	L.PushBoolean(true)
	return 1
}
