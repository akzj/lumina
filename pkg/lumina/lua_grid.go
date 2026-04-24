package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// luaCreateVirtualList implements lumina.createVirtualList(opts) -> table
func luaCreateVirtualList(L *lua.State) int {
	totalItems := 0
	itemHeight := 1
	buffer := 3

	if L.Type(1) == lua.TypeTable {
		L.GetField(1, "totalItems")
		if n, ok := L.ToInteger(-1); ok {
			totalItems = int(n)
		}
		L.Pop(1)

		L.GetField(1, "itemHeight")
		if n, ok := L.ToInteger(-1); ok && n > 0 {
			itemHeight = int(n)
		}
		L.Pop(1)

		L.GetField(1, "buffer")
		if n, ok := L.ToInteger(-1); ok && n >= 0 {
			buffer = int(n)
		}
		L.Pop(1)
	}

	vl := NewVirtualList(totalItems, itemHeight)
	vl.SetBuffer(buffer)

	// Return a table with methods
	L.NewTable()

	// visibleRange(viewportHeight) -> start, end
	L.PushFunction(func(L *lua.State) int {
		vpHeight := int(L.CheckInteger(1))
		start, end := vl.VisibleRange(vpHeight)
		L.PushInteger(int64(start))
		L.PushInteger(int64(end))
		return 2
	})
	L.SetField(-2, "visibleRange")

	// scrollTo(index)
	L.PushFunction(func(L *lua.State) int {
		idx := int(L.CheckInteger(1))
		vl.ScrollTo(idx)
		return 0
	})
	L.SetField(-2, "scrollTo")

	// scrollBy(delta)
	L.PushFunction(func(L *lua.State) int {
		delta := int(L.CheckInteger(1))
		vl.ScrollBy(delta)
		return 0
	})
	L.SetField(-2, "scrollBy")

	// totalHeight() -> int
	L.PushFunction(func(L *lua.State) int {
		L.PushInteger(int64(vl.TotalHeight()))
		return 1
	})
	L.SetField(-2, "totalHeight")

	// itemOffset(index) -> int
	L.PushFunction(func(L *lua.State) int {
		idx := int(L.CheckInteger(1))
		L.PushInteger(int64(vl.ItemOffset(idx)))
		return 1
	})
	L.SetField(-2, "itemOffset")

	// setTotalItems(n)
	L.PushFunction(func(L *lua.State) int {
		n := int(L.CheckInteger(1))
		vl.TotalItems = n
		return 0
	})
	L.SetField(-2, "setTotalItems")

	return 1
}
