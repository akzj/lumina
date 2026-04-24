package lumina

import (
	"encoding/json"

	"github.com/akzj/go-lua/pkg/lua"
)

// inspectDispatch handles lumina.inspect("tree") / lumina.inspect("styles", id)
func inspectDispatch(L *lua.State) int {
	action := L.CheckString(1)

	switch action {
	case "tree":
		return inspectTree(L)
	case "component":
		if L.GetTop() < 2 {
			L.PushString(`{"error": "inspect: component id required"}`)
			return 1
		}
		return inspectComponent(L)
	case "styles":
		if L.GetTop() < 2 {
			L.PushString(`{"error": "inspect: styles id required"}`)
			return 1
		}
		return inspectStyles(L)
	case "frames":
		count := 10
		if L.GetTop() >= 2 {
			if c, ok := L.ToInteger(2); ok {
				count = int(c)
			}
		}
		L.PushInteger(int64(count))
		return inspectFrames(L)
	default:
		L.PushString(`{"error": "unknown inspect action: "` + action + `"`)
		return 1
	}
}

// inspectTree returns JSON string of component tree.
func inspectTree(L *lua.State) int {
	tree := InspectTree()
	jsonBytes, err := json.Marshal(tree)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}

// inspectComponent returns JSON string of component details.
func inspectComponent(L *lua.State) int {
	id := L.CheckString(1)
	comp := InspectComponent(id)
	if comp == nil {
		L.PushString(`{"error": "component not found"}`)
		return 1
	}
	jsonBytes, err := json.Marshal(comp)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}

// inspectStyles returns JSON string of computed styles.
func inspectStyles(L *lua.State) int {
	id := L.CheckString(1)
	styles := InspectStyles(id)
	if styles == nil {
		L.PushString(`{"error": "component not found"}`)
		return 1
	}
	jsonBytes, err := json.Marshal(styles)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}

// inspectFrames returns JSON array of recent frames.
func inspectFrames(L *lua.State) int {
	count := 10
	if L.GetTop() >= 1 {
		if c, ok := L.ToInteger(1); ok {
			count = int(c)
		}
		if count < 1 {
			count = 10
		}
		if count > 100 {
			count = 100
		}
	}

	history := GetFrameHistory()
	start := 0
	if len(history) > count {
		start = len(history) - count
	}

	frames := history[start:]

	// Convert frames to serializable format
	type SerialFrame struct {
		Timestamp int64           `json:"timestamp"`
		Width     int              `json:"width"`
		Height    int              `json:"height"`
		Dirty     int              `json:"dirty_rects"`
	}

	serialFrames := make([]SerialFrame, len(frames))
	for i, f := range frames {
		serialFrames[i] = SerialFrame{
			Timestamp: f.Timestamp,
			Width:     f.Width,
			Height:    f.Height,
			Dirty:     len(f.DirtyRects),
		}
	}

	jsonBytes, err := json.Marshal(serialFrames)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}

// getState returns component state as JSON.
func getState(L *lua.State) int {
	id := L.CheckString(1)
	state, ok := GetState(id)
	if !ok {
		L.PushString(`{"error": "component not found"}`)
		return 1
	}
	jsonBytes, err := json.Marshal(state)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}

// getAllComponents returns all component IDs as JSON.
func getAllComponents(L *lua.State) int {
	ids := GetAllComponentIDs()
	jsonBytes, err := json.Marshal(ids)
	if err != nil {
		L.PushString(`{"error": "` + err.Error() + `"}`)
		return 1
	}
	L.PushString(string(jsonBytes))
	return 1
}
