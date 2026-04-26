package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)


// -----------------------------------------------------------------------
// SubPixel Canvas Lua API
// -----------------------------------------------------------------------

// luaCreateCanvas creates a SubPixelCanvas.
// lumina.createCanvas(cellW, cellH) → canvas userdata with methods
func luaCreateCanvas(L *lua.State) int {
	cellW := int(L.CheckInteger(1))
	cellH := int(L.CheckInteger(2))
	if cellW <= 0 {
		cellW = 1
	}
	if cellH <= 0 {
		cellH = 1
	}

	canvas := NewSubPixelCanvas(cellW, cellH)

	// Return as a table with methods
	L.NewTable()

	// Store canvas pointer as light userdata
	L.PushAny(canvas)
	L.SetField(-2, "_canvas")

	L.PushNumber(float64(canvas.CellW))
	L.SetField(-2, "width")
	L.PushNumber(float64(canvas.CellH))
	L.SetField(-2, "height")
	L.PushNumber(float64(canvas.PixW))
	L.SetField(-2, "pixelWidth")
	L.PushNumber(float64(canvas.PixH))
	L.SetField(-2, "pixelHeight")

	// setPixel(x, y, color)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		hex := L.CheckString(3)
		canvas.SetPixel(x, y, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "setPixel")

	// drawLine(x1, y1, x2, y2, color)
	L.PushFunction(func(L *lua.State) int {
		x1 := int(L.CheckInteger(1))
		y1 := int(L.CheckInteger(2))
		x2 := int(L.CheckInteger(3))
		y2 := int(L.CheckInteger(4))
		hex := L.CheckString(5)
		canvas.DrawLine(x1, y1, x2, y2, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "drawLine")

	// drawRect(x, y, w, h, color)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		w := int(L.CheckInteger(3))
		h := int(L.CheckInteger(4))
		hex := L.CheckString(5)
		canvas.DrawRect(x, y, w, h, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "drawRect")

	// fillRect(x, y, w, h, color)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		w := int(L.CheckInteger(3))
		h := int(L.CheckInteger(4))
		hex := L.CheckString(5)
		canvas.FillRect(x, y, w, h, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "fillRect")

	// drawCircle(cx, cy, r, color)
	L.PushFunction(func(L *lua.State) int {
		cx := int(L.CheckInteger(1))
		cy := int(L.CheckInteger(2))
		r := int(L.CheckInteger(3))
		hex := L.CheckString(4)
		canvas.DrawCircle(cx, cy, r, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "drawCircle")

	// drawRoundedRect(x, y, w, h, radius, color)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		w := int(L.CheckInteger(3))
		h := int(L.CheckInteger(4))
		radius := int(L.CheckInteger(5))
		hex := L.CheckString(6)
		canvas.DrawRoundedRect(x, y, w, h, radius, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "drawRoundedRect")

	// clear()
	L.PushFunction(func(L *lua.State) int {
		canvas.Clear()
		return 0
	})
	L.SetField(-2, "clear")

	return 1
}
