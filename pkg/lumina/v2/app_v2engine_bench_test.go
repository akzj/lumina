package v2

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
)

func BenchmarkAppV2Engine_HoverCycle(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	ta := output.NewTestAdapter()
	app := NewAppWithEngine(L, 80, 24, ta)

	// Grid with hover tracking via useState
	err := app.RunString(`
		lumina.createComponent({
			id = "stress",
			name = "Stress",
			render = function(props)
				local hovered, setHovered = lumina.useState("h", "")
				local children = {}
				for y = 0, 22 do
					local row = {}
					for x = 0, 79 do
						local id = x..","..y
						local ch = "."
						if hovered == id then ch = "#" end
						row[#row+1] = lumina.createElement("box", {
							id = id,
							key = id,
							style = {width=1, height=1},
							onMouseEnter = function() setHovered(id) end,
						}, lumina.createElement("text", {}, ch))
					end
					children[#children+1] = lumina.createElement("hbox", {}, table.unpack(row))
				end
				return lumina.createElement("vbox", {
					style = {width=80, height=24},
				}, table.unpack(children))
			end,
		})
	`)
	if err != nil {
		b.Fatal(err)
	}

	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 80
		y := i % 23
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
		app.RenderDirty()
	}
}

func BenchmarkAppV2Engine_SimpleRenderDirty(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	ta := output.NewTestAdapter()
	app := NewAppWithEngine(L, 80, 24, ta)

	err := app.RunString(`
		lumina.createComponent({
			id = "counter",
			name = "Counter",
			render = function(props)
				local count, setCount = lumina.useState("c", 0)
				return lumina.createElement("box", {
					style = {width = 80, height = 24},
					onClick = function() setCount(count + 1) end,
				}, lumina.createElement("text", {}, tostring(count)))
			end,
		})
	`)
	if err != nil {
		b.Fatal(err)
	}

	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.HandleEvent(&event.Event{Type: "click", X: 10, Y: 10})
		app.RenderDirty()
	}
}
