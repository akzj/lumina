package bridge

import (
	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// LuaTableToVNode converts a Lua table at stack index idx to a VNode tree.
// Expected Lua table structure:
//
//	{
//	  type = "box",          -- node type
//	  id = "my-node",        -- optional unique ID
//	  content = "hello",     -- text content (for "text" nodes)
//	  style = { ... },       -- layout/visual style table
//	  children = { ... },    -- array of child VNode tables
//	  -- remaining keys become Props (including event handler refs)
//	}
func (b *Bridge) LuaTableToVNode(idx int) *layout.VNode {
	L := b.L
	absIdx := L.AbsIndex(idx)

	// Read "type" field.
	nodeType := L.GetFieldString(absIdx, "type")
	if nodeType == "" {
		nodeType = "box"
	}
	vn := layout.NewVNode(nodeType)

	// Read "id" field.
	vn.ID = L.GetFieldString(absIdx, "id")

	// Read "content" field (for text nodes).
	vn.Content = L.GetFieldString(absIdx, "content")

	// Read "style" table → populate vn.Style.
	L.GetField(absIdx, "style")
	if L.IsTable(-1) {
		extractStyle(L, -1, &vn.Style)
	}
	L.Pop(1)

	// Read remaining keys as Props (skip known fields).
	// This captures event handlers (onClick, etc.) and arbitrary props.
	L.ForEach(absIdx, func(L *lua.State) bool {
		if L.Type(-2) != lua.TypeString {
			return true // skip non-string keys
		}
		key, _ := L.ToString(-2)
		if key == "type" || key == "id" || key == "content" || key == "style" || key == "children" {
			return true // skip known structural fields
		}
		if L.Type(-1) == lua.TypeFunction {
			// Store function as registry ref so it survives stack cleanup.
			L.PushValue(-1)
			ref := L.Ref(lua.RegistryIndex)
			b.TrackRef(ref)
			vn.Props[key] = ref
		} else {
			vn.Props[key] = L.ToAny(-1)
		}
		return true
	})

	// Read "children" array → recurse.
	L.GetField(absIdx, "children")
	if L.IsTable(-1) {
		n := int(L.RawLen(-1))
		for i := 1; i <= n; i++ {
			L.RawGetI(-1, int64(i))
			if L.IsTable(-1) {
				child := b.LuaTableToVNode(-1)
				vn.AddChild(child)
			} else if L.IsString(-1) {
				// String child → create a text VNode (defense-in-depth).
				s, _ := L.ToString(-1)
				textVN := layout.NewVNode("text")
				textVN.Content = s
				vn.AddChild(textVN)
			}
			L.Pop(1)
		}
	}
	L.Pop(1) // pop children

	return vn
}

// extractStyle reads a Lua style table into a layout.Style struct.
func extractStyle(L *lua.State, idx int, s *layout.Style) {
	absIdx := L.AbsIndex(idx)

	// Sizing
	s.Width = int(L.GetFieldInt(absIdx, "width"))
	s.Height = int(L.GetFieldInt(absIdx, "height"))
	s.MinWidth = int(L.GetFieldInt(absIdx, "minWidth"))
	s.MaxWidth = int(L.GetFieldInt(absIdx, "maxWidth"))
	s.MinHeight = int(L.GetFieldInt(absIdx, "minHeight"))
	s.MaxHeight = int(L.GetFieldInt(absIdx, "maxHeight"))
	s.Flex = int(L.GetFieldInt(absIdx, "flex"))

	// Spacing
	s.Padding = int(L.GetFieldInt(absIdx, "padding"))
	s.PaddingTop = int(L.GetFieldInt(absIdx, "paddingTop"))
	s.PaddingBottom = int(L.GetFieldInt(absIdx, "paddingBottom"))
	s.PaddingLeft = int(L.GetFieldInt(absIdx, "paddingLeft"))
	s.PaddingRight = int(L.GetFieldInt(absIdx, "paddingRight"))
	s.Margin = int(L.GetFieldInt(absIdx, "margin"))
	s.MarginTop = int(L.GetFieldInt(absIdx, "marginTop"))
	s.MarginBottom = int(L.GetFieldInt(absIdx, "marginBottom"))
	s.MarginLeft = int(L.GetFieldInt(absIdx, "marginLeft"))
	s.MarginRight = int(L.GetFieldInt(absIdx, "marginRight"))
	s.Gap = int(L.GetFieldInt(absIdx, "gap"))

	// Alignment
	if v := L.GetFieldString(absIdx, "justify"); v != "" {
		s.Justify = v
	}
	if v := L.GetFieldString(absIdx, "align"); v != "" {
		s.Align = v
	}

	// Visual
	s.Border = L.GetFieldString(absIdx, "border")
	s.Foreground = L.GetFieldString(absIdx, "foreground")
	if v := L.GetFieldString(absIdx, "fg"); v != "" && s.Foreground == "" {
		s.Foreground = v
	}
	s.Background = L.GetFieldString(absIdx, "background")
	if v := L.GetFieldString(absIdx, "bg"); v != "" && s.Background == "" {
		s.Background = v
	}
	s.Bold = L.GetFieldBool(absIdx, "bold")
	s.Dim = L.GetFieldBool(absIdx, "dim")
	s.Underline = L.GetFieldBool(absIdx, "underline")

	// Overflow
	s.Overflow = L.GetFieldString(absIdx, "overflow")

	// Positioning
	s.Position = L.GetFieldString(absIdx, "position")
	s.Top = int(L.GetFieldInt(absIdx, "top"))
	s.Left = int(L.GetFieldInt(absIdx, "left"))

	// Right and Bottom: -1 = unset (default from NewVNode).
	// Only override if explicitly set in Lua.
	if v := L.GetFieldInt(absIdx, "right"); v != 0 {
		s.Right = int(v)
	}
	if v := L.GetFieldInt(absIdx, "bottom"); v != 0 {
		s.Bottom = int(v)
	}

	s.ZIndex = int(L.GetFieldInt(absIdx, "zIndex"))
}
