package lumina_test

import (
	"fmt"
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

// ─── Render Benchmarks ───────────────────────────────────────────────

func BenchmarkRenderSimpleComponent(b *testing.B) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "BenchSimple",
			render = function(self)
				return { type = "text", content = "Hello Benchmark" }
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tio.Reset()
		app.RenderOnce()
	}
}

func BenchmarkRenderNestedComponents(b *testing.B) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "BenchNested",
			render = function(self)
				return {
					type = "vbox",
					children = {
						{
							type = "hbox",
							children = {
								{ type = "text", content = "Left" },
								{ type = "text", content = "Right" },
							}
						},
						{
							type = "vbox",
							children = {
								{ type = "text", content = "Row 1" },
								{ type = "text", content = "Row 2" },
								{ type = "text", content = "Row 3" },
							}
						},
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tio.Reset()
		app.RenderOnce()
	}
}

func BenchmarkRenderLargeList(b *testing.B) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(120, 50, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "BenchList",
			render = function(self)
				local items = {}
				for i = 1, 50 do
					table.insert(items, { type = "text", content = "Item " .. tostring(i) })
				end
				return {
					type = "vbox",
					children = items,
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tio.Reset()
		app.RenderOnce()
	}
}

func BenchmarkShadcnButton(b *testing.B) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local shadcn = require("shadcn")
		local App = lumina.defineComponent({
			name = "BenchButton",
			render = function(self)
				return {
					type = "vbox",
					children = {
						lumina.createElement(shadcn.Button, { label = "Click me", variant = "default" }),
						lumina.createElement(shadcn.Button, { label = "Secondary", variant = "secondary" }),
						lumina.createElement(shadcn.Button, { label = "Outline", variant = "outline" }),
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tio.Reset()
		app.RenderOnce()
	}
}

func BenchmarkShadcnCard(b *testing.B) {
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local shadcn = require("shadcn")
		local App = lumina.defineComponent({
			name = "BenchCard",
			render = function(self)
				return lumina.createElement(shadcn.Card, {
					children = {
						{ type = "text", content = "Card Title", style = { bold = true } },
						{ type = "text", content = "Card content goes here" },
						lumina.createElement(shadcn.Button, { label = "Action" }),
					}
				})
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tio.Reset()
		app.RenderOnce()
	}
}

// ─── Lua VM Benchmarks ───────────────────────────────────────────────

func BenchmarkLuaDefineComponent(b *testing.B) {
	app := lumina.NewApp()
	defer app.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script := fmt.Sprintf(`
			local lumina = require("lumina")
			lumina.defineComponent({
				name = "Comp%d",
				render = function(self)
					return { type = "text", content = "test" }
				end
			})
		`, i)
		app.L.DoString(script)
	}
}

func BenchmarkLuaCreateElement(b *testing.B) {
	app := lumina.NewApp()
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		local shadcn = require("shadcn")
		_G._lumina = lumina
		_G._shadcn = shadcn
	`)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.L.DoString(`
			_G._lumina.createElement(_G._shadcn.Button, { label = "test" })
		`)
	}
}

func BenchmarkLuaStoreDispatch(b *testing.B) {
	app := lumina.NewApp()
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		_G._store = lumina.createStore({
			state = { count = 0 },
			actions = {
				increment = function(state)
					state.count = state.count + 1
				end,
			},
		})
	`)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.L.DoString(`_G._store.dispatch("increment")`)
	}
}
