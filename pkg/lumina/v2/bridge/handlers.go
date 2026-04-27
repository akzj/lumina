package bridge

import (
	"strings"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// walkExtract recursively walks the VNode tree, extracting event handlers
// from VNode Props and populating the result map.
func (b *Bridge) walkExtract(vn *layout.VNode, result map[string]event.HandlerMap) {
	if vn == nil {
		return
	}
	if vn.ID != "" && vn.Props != nil {
		hm := make(event.HandlerMap)
		for propName, propVal := range vn.Props {
			if !isEventProp(propName) {
				continue
			}
			ref, ok := toLuaRef(propVal)
			if !ok {
				continue
			}
			evtName := mapPropToEvent(propName)
			hm[evtName] = b.WrapLuaHandler(ref)
		}
		if len(hm) > 0 {
			result[vn.ID] = hm
		}
	}
	for _, child := range vn.Children {
		b.walkExtract(child, result)
	}
}

// WrapLuaHandler wraps a Lua registry ref as an event.EventHandler.
// When the handler is called, it pushes the Lua function and event data,
// then calls PCall(1, 0, 0). Exported so App can wrap Lua handler refs
// during handler synchronization.
func (b *Bridge) WrapLuaHandler(ref int) event.EventHandler {
	return func(e *event.Event) {
		L := b.L
		L.RawGetI(lua.RegistryIndex, int64(ref))
		if !L.IsFunction(-1) {
			L.Pop(1)
			return
		}
		pushEventToLua(L, e)
		if status := L.PCall(1, 0, 0); status != lua.OK {
			L.Pop(1) // pop error message
		}
	}
}

// pushEventToLua pushes an event.Event as a Lua table onto the stack.
func pushEventToLua(L *lua.State, e *event.Event) {
	L.NewTableFrom(map[string]any{
		"type":   e.Type,
		"x":      int64(e.X),
		"y":      int64(e.Y),
		"localX": int64(e.LocalX),
		"localY": int64(e.LocalY),
		"key":    e.Key,
		"target": e.Target,
	})
}

// isEventProp returns true if the key is an event handler prop name.
func isEventProp(key string) bool {
	switch key {
	case "onClick", "onChange", "onFocus", "onBlur",
		"onKeyDown", "onKeyUp", "onSubmit", "onScroll",
		"onMouseDown", "onMouseUp", "onMouseMove",
		"onMouseEnter", "onMouseLeave",
		"onDragOver", "onDrop", "onWheel",
		"onInput", "onResize", "onContextMenu":
		return true
	}
	return false
}

// mapPropToEvent converts a camelCase prop name to a lowercase event type.
//
//	"onClick"      → "click"
//	"onMouseEnter" → "mouseenter"
//	"onKeyDown"    → "keydown"
func mapPropToEvent(propName string) string {
	name := strings.TrimPrefix(propName, "on")
	return strings.ToLower(name)
}

// toLuaRef extracts a Lua registry reference from a prop value.
// Props store function refs as int (from Lua.Ref).
func toLuaRef(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		if v > 0 {
			return v, true
		}
	case int64:
		if v > 0 {
			return int(v), true
		}
	}
	return 0, false
}
