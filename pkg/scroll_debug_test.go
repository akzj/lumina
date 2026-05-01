package v2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
	"github.com/akzj/lumina/pkg/render"
)

func TestScrollOverlapDebug(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

	err := app.RunString(`
		lumina.createComponent({
			id = "test", name = "Test",
			render = function()
				local lines = {}
				for i = 1, 30 do
					lines[i] = lumina.createElement("text", {key = "l"..i}, string.format("%02d | Line content here", i))
				end
				return lumina.createElement("box", {
					style = {width = 80, height = 24, background = "#000000"}},
					-- Back box: scrollable editor-like content
					lumina.createElement("vbox", {
						id = "editor",
						style = {
							position = "absolute",
							left = 2, top = 1,
							width = 35, height = 12,
							overflow = "scroll",
							background = "#222222",
						},
					}, table.unpack(lines)),
					-- Front box: overlapping "palette" on top
					lumina.createElement("vbox", {
						id = "palette",
						style = {
							position = "absolute",
							left = 10, top = 3,
							width = 30, height = 10,
							background = "#444444",
						},
					},
						lumina.createElement("text", {}, "Palette"),
						lumina.createElement("text", {}, "Color palette content")
					)
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	// Dump the node tree to understand structure
	root := app.Engine().Root()
	if root != nil && root.RootNode != nil {
		fmt.Println("=== Node tree ===")
		dumpNodeTree(root.RootNode, 0)
	}

	// Dump screen before scroll
	fmt.Println("\n=== Screen BEFORE scroll ===")
	dumpFullScreen(ta, 80, 14)

	// Scroll Editor at x=5, y=5
	for i := 0; i < 5; i++ {
		app.HandleEvent(&event.Event{Type: "scroll", X: 5, Y: 5, Key: "down"})
		app.RenderDirty()
	}

	fmt.Println("\n=== Screen AFTER scroll ===")
	dumpFullScreen(ta, 80, 14)

	// Check if Palette is still visible
	if !screenHasString(ta, "Palette") {
		t.Error("BUG CONFIRMED: Palette title overwritten by Editor scroll")
	}
}

func dumpNodeTree(node *render.Node, depth int) {
	indent := strings.Repeat("  ", depth)
	extra := ""
	if node.Style.Position == "absolute" {
		extra += " [absolute]"
	}
	if node.Style.Overflow == "scroll" {
		extra += fmt.Sprintf(" [scroll scrollY=%d]", node.ScrollY)
	}
	if node.Component != nil {
		extra += fmt.Sprintf(" [comp:%s id:%s]", node.Component.Type, node.Component.ID)
	}
	fmt.Printf("%s%s x=%d y=%d w=%d h=%d%s\n", indent, node.Type, node.X, node.Y, node.W, node.H, extra)
	if depth < 6 {
		for _, ch := range node.Children {
			dumpNodeTree(ch, depth+1)
		}
	}
}

func dumpFullScreen(ta *output.TestAdapter, w, maxRow int) {
	for y := 0; y < maxRow; y++ {
		var sb strings.Builder
		for x := 0; x < w; x++ {
			c := ta.LastScreen.Get(x, y)
			if c.Char == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(c.Char)
			}
		}
		fmt.Printf("  row %2d: %s\n", y, sb.String())
	}
}
