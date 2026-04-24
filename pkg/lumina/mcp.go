package lumina

import (
	"encoding/json"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/proto"
)

// setOutputMode(mode) — "ansi" or "json"
func setOutputMode(L *lua.State) int {
	mode := L.CheckString(1)
	switch mode {
	case "json":
		SetOutputMode(ModeJSON)
	case "ansi":
		SetOutputMode(ModeANSI)
	default:
		L.PushString("setOutputMode: unknown mode '" + mode + "'")
		L.Error()
	}
	return 0
}

// getOutputMode() → mode string
func getOutputMode(L *lua.State) int {
	L.PushString(GetOutputMode().String())
	return 1
}

// getMCPFrame() → JSON frame string (for testing)
func getMCPFrame(L *lua.State) int {
	// Create a sample proto frame for testing
	frame := &proto.Frame{
		Timestamp: time.Now().UnixMilli(),
		Cells: [][]proto.Cell{
			{{Char: "H", Foreground: "cyan"}, {Char: "i", Foreground: "white"}},
		},
	}

	data, err := json.Marshal(frame)
	if err != nil {
		L.PushNil()
		return 1
	}

	L.PushString(string(data))
	return 1
}

// createComponentRequest(type, props) → JSON string
func createComponentRequest(L *lua.State) int {
	compType := L.CheckString(1)
	props := make(map[string]string)

	if L.GetTop() >= 2 && L.Type(2) == lua.TypeTable {
		L.PushNil()
		for L.Next(2) {
			if k, _ := L.ToString(-2); k != "" {
				if v, _ := L.ToString(-1); v != "" {
					props[k] = v
				}
			}
			L.Pop(1)
		}
	}

	req := &proto.ComponentRequest{
		Type:  compType,
		Props: props,
	}

	data, err := json.Marshal(req)
	if err != nil {
		L.PushNil()
		return 1
	}

	L.PushString(string(data))
	return 1
}

// createEventNotification(componentID, eventType, eventData?) → JSON string
func createEventNotification(L *lua.State) int {
	compID := L.CheckString(1)
	eventType := L.CheckString(2)

	eventData := make(map[string]string)
	if L.GetTop() >= 3 && L.Type(3) == lua.TypeTable {
		L.PushNil()
		for L.Next(3) {
			if k, _ := L.ToString(-2); k != "" {
				if v, _ := L.ToString(-1); v != "" {
					eventData[k] = v
				}
			}
			L.Pop(1)
		}
	}

	notif := proto.NewEventNotification(compID, eventType, eventData)

	data, err := json.Marshal(notif)
	if err != nil {
		L.PushNil()
		return 1
	}

	L.PushString(string(data))
	return 1
}
