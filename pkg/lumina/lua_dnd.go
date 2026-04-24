package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// luaUseDrag implements lumina.useDrag(opts) -> drag table
func luaUseDrag(L *lua.State) int {
	dragType := "default"
	var dragData any

	if L.Type(1) == lua.TypeTable {
		L.GetField(1, "type")
		if s, ok := L.ToString(-1); ok && s != "" {
			dragType = s
		}
		L.Pop(1)

		L.GetField(1, "data")
		dragData = L.ToAny(-1)
		L.Pop(1)
	}

	ds := GetDragState()

	L.NewTable()

	// drag.start(sourceID)
	L.PushFunction(func(L *lua.State) int {
		sourceID := L.CheckString(1)
		ds.StartDrag(sourceID, dragType, dragData)
		return 0
	})
	L.SetField(-2, "start")

	// drag.end() -> sourceID, targetID, data
	L.PushFunction(func(L *lua.State) int {
		src, tgt, data := ds.EndDrag()
		L.PushString(src)
		L.PushString(tgt)
		L.PushAny(data)
		return 3
	})
	L.SetField(-2, "stop")

	// drag.isDragging() -> bool
	L.PushFunction(func(L *lua.State) int {
		L.PushBoolean(ds.Dragging())
		return 1
	})
	L.SetField(-2, "isDragging")

	// drag.updatePosition(x, y)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		ds.UpdatePosition(x, y)
		return 0
	})
	L.SetField(-2, "updatePosition")

	return 1
}

// luaUseDrop implements lumina.useDrop(opts) -> drop table
func luaUseDrop(L *lua.State) int {
	var accept []string
	var onDropRef int
	hasOnDrop := false

	if L.Type(1) == lua.TypeTable {
		L.GetField(1, "accept")
		if L.Type(-1) == lua.TypeTable {
			n := int(L.RawLen(-1))
			for i := 1; i <= n; i++ {
				L.RawGetI(-1, int64(i))
				if s, ok := L.ToString(-1); ok {
					accept = append(accept, s)
				}
				L.Pop(1)
			}
		}
		L.Pop(1)

		L.GetField(1, "onDrop")
		if L.Type(-1) == lua.TypeFunction {
			onDropRef = L.Ref(lua.RegistryIndex)
			hasOnDrop = true
		} else {
			L.Pop(1)
		}
	}

	dz := &DropZone{Accept: accept}
	ds := GetDragState()

	L.NewTable()

	// drop.canDrop() -> bool
	L.PushFunction(func(L *lua.State) int {
		if !ds.Dragging() {
			L.PushBoolean(false)
			return 1
		}
		L.PushBoolean(dz.CanDrop(ds.GetDragType()))
		return 1
	})
	L.SetField(-2, "canDrop")

	// drop.setTarget(targetID)
	L.PushFunction(func(L *lua.State) int {
		targetID := L.CheckString(1)
		ds.SetDropTarget(targetID)
		return 0
	})
	L.SetField(-2, "setTarget")

	// drop.drop() -> bool (triggers onDrop if valid)
	L.PushFunction(func(L *lua.State) int {
		if !ds.Dragging() || !dz.CanDrop(ds.GetDragType()) {
			L.PushBoolean(false)
			return 1
		}
		data := ds.GetDragData()
		ds.EndDrag()

		if hasOnDrop {
			L.RawGetI(lua.RegistryIndex, int64(onDropRef))
			L.PushAny(data)
			L.PCall(1, 0, 0)
		}
		L.PushBoolean(true)
		return 1
	})
	L.SetField(-2, "drop")

	// drop.isOver() -> bool
	L.PushFunction(func(L *lua.State) int {
		L.PushBoolean(ds.Dragging())
		return 1
	})
	L.SetField(-2, "isOver")

	return 1
}
