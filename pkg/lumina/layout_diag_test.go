package lumina

import (
	"fmt"
	"testing"
	"time"
)

func TestComponentLib_LayoutDiag(t *testing.T) {
	app := NewAppWithSize(120, 40)
	tio := NewMockTermIO(120, 40)
	SetOutputAdapter(NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/components/main.lua", tio)
	if err != nil {
		t.Fatalf("LoadScript: %v", err)
	}

	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame := app.lastFrame
	if frame == nil {
		t.Fatal("No frame rendered")
	}

	// Find the root VNode
	for _, comp := range globalRegistry.components {
		if comp.LastVNode != nil {
			dumpVNode(comp.LastVNode, 0)
		}
	}

	// Check viewport for content-scroll
	vp := GetViewport("content-scroll")
	fmt.Printf("\nViewport 'content-scroll': ContentH=%d, ViewH=%d, ScrollY=%d, NeedsScroll=%v\n",
		vp.ContentH, vp.ViewH, vp.ScrollY, vp.NeedsScroll())
}

func dumpVNode(v *VNode, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	id := ""
	if idVal, ok := v.Props["id"].(string); ok {
		id = " #" + idVal
	}
	content := ""
	if v.Content != "" {
		if len(v.Content) > 40 {
			content = " \"" + v.Content[:40] + "...\""
		} else {
			content = " \"" + v.Content + "\""
		}
	}
	fmt.Printf("%s<%s%s> pos=(%d,%d) size=%dx%d overflow=%s%s\n",
		indent, v.Type, id, v.X, v.Y, v.W, v.H, v.Style.Overflow, content)
	if depth < 5 { // limit depth to avoid huge output
		for _, child := range v.Children {
			dumpVNode(child, depth+1)
		}
	}
}
