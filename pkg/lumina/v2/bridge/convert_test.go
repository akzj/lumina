package bridge

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestLuaTableToVNode_TopLevelStyleProps(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	b := NewBridge(L)

	// Simulate createElement("text", {foreground="#89B4FA", bold=true}, "Hello")
	// which produces: {type="text", foreground="#89B4FA", bold=true, content="Hello"}
	L.CreateTable(0, 4)
	idx := L.AbsIndex(-1)
	L.PushString("text")
	L.SetField(idx, "type")
	L.PushString("#89B4FA")
	L.SetField(idx, "foreground")
	L.PushBoolean(true)
	L.SetField(idx, "bold")
	L.PushString("Hello")
	L.SetField(idx, "content")

	vn := b.LuaTableToVNode(-1)

	if vn.Style.Foreground != "#89B4FA" {
		t.Errorf("Style.Foreground = %q, want %q", vn.Style.Foreground, "#89B4FA")
	}
	if !vn.Style.Bold {
		t.Error("Style.Bold = false, want true")
	}
	// Should NOT be in Props
	if _, ok := vn.Props["foreground"]; ok {
		t.Error("foreground should not be in Props")
	}
	if _, ok := vn.Props["bold"]; ok {
		t.Error("bold should not be in Props")
	}
}

func TestLuaTableToVNode_StyleSubTableOverridesTopLevel(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	b := NewBridge(L)

	// {type="box", background="#111", style={background="#222"}}
	// style sub-table should win
	L.CreateTable(0, 3)
	idx := L.AbsIndex(-1)
	L.PushString("box")
	L.SetField(idx, "type")
	L.PushString("#111")
	L.SetField(idx, "background")

	L.CreateTable(0, 1)
	L.PushString("#222")
	L.SetField(-2, "background")
	L.SetField(idx, "style")

	vn := b.LuaTableToVNode(-1)

	if vn.Style.Background != "#222" {
		t.Errorf("Style.Background = %q, want %q (style sub-table should override)", vn.Style.Background, "#222")
	}
}

func TestLuaTableToVNode_TopLevelNumericStyleProps(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	b := NewBridge(L)

	// {type="box", width=20, height=10, padding=2}
	L.CreateTable(0, 4)
	idx := L.AbsIndex(-1)
	L.PushString("box")
	L.SetField(idx, "type")
	L.PushInteger(20)
	L.SetField(idx, "width")
	L.PushInteger(10)
	L.SetField(idx, "height")
	L.PushInteger(2)
	L.SetField(idx, "padding")

	vn := b.LuaTableToVNode(-1)

	if vn.Style.Width != 20 {
		t.Errorf("Style.Width = %d, want 20", vn.Style.Width)
	}
	if vn.Style.Height != 10 {
		t.Errorf("Style.Height = %d, want 10", vn.Style.Height)
	}
	if vn.Style.Padding != 2 {
		t.Errorf("Style.Padding = %d, want 2", vn.Style.Padding)
	}
	// Should NOT be in Props
	if _, ok := vn.Props["width"]; ok {
		t.Error("width should not be in Props")
	}
	if _, ok := vn.Props["height"]; ok {
		t.Error("height should not be in Props")
	}
}

func TestLuaTableToVNode_NonStylePropsStillWork(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	b := NewBridge(L)

	// {type="box", customProp="hello", foreground="#FFF"}
	// customProp should go to Props, foreground should go to Style
	L.CreateTable(0, 3)
	idx := L.AbsIndex(-1)
	L.PushString("box")
	L.SetField(idx, "type")
	L.PushString("hello")
	L.SetField(idx, "customProp")
	L.PushString("#FFF")
	L.SetField(idx, "foreground")

	vn := b.LuaTableToVNode(-1)

	if vn.Style.Foreground != "#FFF" {
		t.Errorf("Style.Foreground = %q, want %q", vn.Style.Foreground, "#FFF")
	}
	if v, ok := vn.Props["customProp"]; !ok || v != "hello" {
		t.Errorf("Props[customProp] = %v (ok=%v), want 'hello'", v, ok)
	}
	if _, ok := vn.Props["foreground"]; ok {
		t.Error("foreground should not be in Props")
	}
}
