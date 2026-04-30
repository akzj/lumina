package render

import (
	"strconv"

	"github.com/akzj/go-lua/pkg/lua"
)

func convertChildDescriptors(props map[string]any) []*Node {
	raw, ok := props["children"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}

	nodes := make([]*Node, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		desc := descriptorFromMap(m)
		node := createNodeFromDesc(desc)
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// descriptorFromMap converts a raw map (from Lua ToAny) back into a Descriptor.
// This is the inverse of readDescriptor but works on Go maps instead of Lua stack.
func descriptorFromMap(m map[string]any) Descriptor {
	var desc Descriptor

	// Type
	if t, ok := m["type"].(string); ok {
		desc.Type = t
	}
	if desc.Type == "" {
		desc.Type = "box"
	}

	// Content
	if c, ok := m["content"].(string); ok {
		desc.Content = c
		desc.ContentSet = true
	}

	// ID
	if id, ok := m["id"].(string); ok {
		desc.ID = id
	}

	// Key
	if k, ok := m["key"].(string); ok {
		desc.Key = k
	}

	// Component type (for sub-components)
	if ft, ok := m["_factoryName"].(string); ok {
		desc.Type = "component"
		desc.ComponentType = ft
		if p, ok := m["_props"].(map[string]any); ok {
			desc.ComponentProps = p
		}
	}

	// Style
	if s, ok := m["style"].(map[string]any); ok {
		desc.Style = styleFromMap(s)
	} else {
		desc.Style.Right = -1
		desc.Style.Bottom = -1
	}

	// Foreground/Background at top level (Lua shorthand)
	if fg, ok := m["foreground"].(string); ok {
		desc.Style.Foreground = fg
	}
	if bg, ok := m["background"].(string); ok {
		desc.Style.Background = bg
	}

	// Bold, Dim, Underline at top level
	if b, ok := m["bold"].(bool); ok {
		desc.Style.Bold = b
	}
	if d, ok := m["dim"].(bool); ok {
		desc.Style.Dim = d
	}
	if u, ok := m["underline"].(bool); ok {
		desc.Style.Underline = u
	}

	// Focusable / Disabled
	if f, ok := m["focusable"].(bool); ok {
		desc.Focusable = f
	}
	if d, ok := m["disabled"].(bool); ok {
		desc.Disabled = d
	}

	// Placeholder
	if p, ok := m["placeholder"].(string); ok {
		desc.Placeholder = p
	}

	// AutoFocus
	if af, ok := m["autoFocus"].(bool); ok {
		desc.AutoFocus = af
	}

	// Children (recursive)
	if children, ok := m["children"].([]any); ok {
		for _, child := range children {
			if cm, ok := child.(map[string]any); ok {
				childDesc := descriptorFromMap(cm)
				desc.Children = append(desc.Children, childDesc)
			}
		}
	}

	return desc
}


// readDescriptor reads a Lua table at stack index and converts to Descriptor.
func (e *Engine) readDescriptor(L *lua.State, idx int) Descriptor {
	absIdx := L.AbsIndex(idx)

	var desc Descriptor
	desc.Type = getStringField(L, absIdx, "type")
	if desc.Type == "" {
		desc.Type = "box"
	}
	desc.ID = getStringField(L, absIdx, "id")
	desc.Key = getStringField(L, absIdx, "key")
	desc.Content = getStringField(L, absIdx, "content")
	if desc.Content != "" {
		desc.ContentSet = true
	}
	// For input/textarea, also check "value" field
	if desc.Content == "" {
		if v := getStringField(L, absIdx, "value"); v != "" {
			desc.Content = v
			desc.ContentSet = true
		}
	}
	desc.Placeholder = getStringField(L, absIdx, "placeholder")
	desc.AutoFocus = getBoolField(L, absIdx, "autoFocus")
	L.GetField(absIdx, "scrollY")
	if !L.IsNoneOrNil(-1) {
		n, _ := L.ToInteger(-1)
		desc.ScrollY = int(n)
		desc.ScrollYSet = true
	}
	L.Pop(1)

	// Read style — check for nested "style" table first, then top-level style fields
	L.GetField(absIdx, "style")
	if L.IsTable(-1) {
		desc.Style = e.readStyle(L, -1)
	}
	L.Pop(1)

	// Also read top-level style fields (they override if style sub-table didn't set them)
	e.readStyleFields(L, absIdx, &desc.Style)

	// Read event handlers (store as Lua refs)
	desc.OnClick = getRefField(L, absIdx, "onClick")
	desc.OnMouseEnter = getRefField(L, absIdx, "onMouseEnter")
	desc.OnMouseLeave = getRefField(L, absIdx, "onMouseLeave")
	desc.OnKeyDown = getRefField(L, absIdx, "onKeyDown")
	desc.OnChange = getRefField(L, absIdx, "onChange")
	desc.OnScroll = getRefField(L, absIdx, "onScroll")
	desc.OnMouseDown = getRefField(L, absIdx, "onMouseDown")
	desc.OnMouseUp = getRefField(L, absIdx, "onMouseUp")
	desc.OnFocus = getRefField(L, absIdx, "onFocus")
	desc.OnBlur = getRefField(L, absIdx, "onBlur")
	desc.OnSubmit = getRefField(L, absIdx, "onSubmit")
	desc.OnOutsideClick = getRefField(L, absIdx, "onOutsideClick")
	desc.Focusable = getBoolField(L, absIdx, "focusable")
	desc.Disabled = getBoolField(L, absIdx, "disabled")

	// Read children
	L.GetField(absIdx, "children")
	if L.IsTable(-1) {
		childrenIdx := L.AbsIndex(-1)
		n := int(L.RawLen(childrenIdx))
		desc.Children = make([]Descriptor, 0, n)
		for i := 1; i <= n; i++ {
			L.RawGetI(childrenIdx, int64(i))
			if L.IsTable(-1) {
				child := e.readDescriptor(L, -1)
				desc.Children = append(desc.Children, child)
			} else if L.IsString(-1) {
				// String child → text descriptor
				s, _ := L.ToString(-1)
				desc.Children = append(desc.Children, Descriptor{
					Type:    "text",
					Content: s,
				})
			}
			L.Pop(1)
		}
	}
	L.Pop(1)

	// Check if this is a component type
	factoryName := getStringField(L, absIdx, "_factoryName")
	if factoryName != "" {
		desc.Type = "component"
		desc.ComponentType = factoryName
		L.GetField(absIdx, "_props")
		if L.IsTable(-1) {
			desc.ComponentProps = readMapFromTable(L, -1)
		}
		L.Pop(1)
	}

	// Backward compat: input/textarea are always focusable
	if desc.Type == "input" || desc.Type == "textarea" {
		desc.Focusable = true
	}

	return desc
}

// readStyle reads a style table from the Lua stack.
func (e *Engine) readStyle(L *lua.State, idx int) Style {
	absIdx := L.AbsIndex(idx)
	var s Style
	s.Width = int(getIntField(L, absIdx, "width"))
	s.Height = int(getIntField(L, absIdx, "height"))
	// Parse percentage/viewport string values for width/height
	if s.Width == 0 {
		if str := getStringField(L, absIdx, "width"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.WidthPercent = pct
			} else if v, unit, ok := parseViewport(str); ok && unit == "vw" {
				s.WidthVW = v
			}
		}
	}
	if s.Height == 0 {
		if str := getStringField(L, absIdx, "height"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.HeightPercent = pct
			} else if v, unit, ok := parseViewport(str); ok && unit == "vh" {
				s.HeightVH = v
			}
		}
	}
	s.Flex = int(getIntField(L, absIdx, "flex"))
	s.FlexShrink = int(getIntField(L, absIdx, "flexShrink"))
	s.FlexBasis = int(getIntField(L, absIdx, "flexBasis"))
	s.FlexWrap = getStringField(L, absIdx, "flexWrap")
	s.Padding = int(getIntField(L, absIdx, "padding"))
	s.PaddingTop = int(getIntField(L, absIdx, "paddingTop"))
	s.PaddingBottom = int(getIntField(L, absIdx, "paddingBottom"))
	s.PaddingLeft = int(getIntField(L, absIdx, "paddingLeft"))
	s.PaddingRight = int(getIntField(L, absIdx, "paddingRight"))
	s.Margin = int(getIntField(L, absIdx, "margin"))
	s.MarginTop = int(getIntField(L, absIdx, "marginTop"))
	s.MarginBottom = int(getIntField(L, absIdx, "marginBottom"))
	s.MarginLeft = int(getIntField(L, absIdx, "marginLeft"))
	s.MarginRight = int(getIntField(L, absIdx, "marginRight"))
	s.Gap = int(getIntField(L, absIdx, "gap"))
	s.MinWidth = int(getIntField(L, absIdx, "minWidth"))
	s.MaxWidth = int(getIntField(L, absIdx, "maxWidth"))
	s.MinHeight = int(getIntField(L, absIdx, "minHeight"))
	s.MaxHeight = int(getIntField(L, absIdx, "maxHeight"))
	// Parse percentage string values for min/max
	if s.MinWidth == 0 {
		if str := getStringField(L, absIdx, "minWidth"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.MinWidthPercent = pct
			}
		}
	}
	if s.MaxWidth == 0 {
		if str := getStringField(L, absIdx, "maxWidth"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.MaxWidthPercent = pct
			}
		}
	}
	if s.MinHeight == 0 {
		if str := getStringField(L, absIdx, "minHeight"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.MinHeightPercent = pct
			}
		}
	}
	if s.MaxHeight == 0 {
		if str := getStringField(L, absIdx, "maxHeight"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.MaxHeightPercent = pct
			}
		}
	}
	s.Justify = getStringField(L, absIdx, "justify")
	s.Align = getStringField(L, absIdx, "align")
	s.AlignSelf = getStringField(L, absIdx, "alignSelf")
	s.Order = int(getIntField(L, absIdx, "order"))
	s.Border = getStringField(L, absIdx, "border")
	s.Foreground = getStringField(L, absIdx, "foreground")
	if fg := getStringField(L, absIdx, "fg"); fg != "" && s.Foreground == "" {
		s.Foreground = fg
	}
	s.Background = getStringField(L, absIdx, "background")
	if bg := getStringField(L, absIdx, "bg"); bg != "" && s.Background == "" {
		s.Background = bg
	}
	s.Bold = getBoolField(L, absIdx, "bold")
	s.Dim = getBoolField(L, absIdx, "dim")
	s.Underline = getBoolField(L, absIdx, "underline")
	s.Italic = getBoolField(L, absIdx, "italic")
	s.Strikethrough = getBoolField(L, absIdx, "strikethrough")
	s.Inverse = getBoolField(L, absIdx, "inverse")
	s.Overflow = getStringField(L, absIdx, "overflow")
	s.Position = getStringField(L, absIdx, "position")
	s.Top = int(getIntField(L, absIdx, "top"))
	s.Left = int(getIntField(L, absIdx, "left"))
	s.Right = int(getIntFieldDefault(L, absIdx, "right", -1))
	s.Bottom = int(getIntFieldDefault(L, absIdx, "bottom", -1))
	s.ZIndex = int(getIntField(L, absIdx, "zIndex"))
	s.TextAlign = getStringField(L, absIdx, "textAlign")
	s.TextOverflow = getStringField(L, absIdx, "textOverflow")
	s.WhiteSpace = getStringField(L, absIdx, "whiteSpace")
	s.Display = getStringField(L, absIdx, "display")
	s.Visibility = getStringField(L, absIdx, "visibility")
	s.BorderColor = getStringField(L, absIdx, "borderColor")
	// Grid container properties
	s.GridTemplateColumns = getStringField(L, absIdx, "gridTemplateColumns")
	s.GridTemplateRows = getStringField(L, absIdx, "gridTemplateRows")
	s.GridColumnGap = int(getIntField(L, absIdx, "gridColumnGap"))
	s.GridRowGap = int(getIntField(L, absIdx, "gridRowGap"))
	// Grid item properties
	s.GridColumn = getStringField(L, absIdx, "gridColumn")
	s.GridRow = getStringField(L, absIdx, "gridRow")
	s.GridColumnStart = int(getIntField(L, absIdx, "gridColumnStart"))
	s.GridColumnEnd = int(getIntField(L, absIdx, "gridColumnEnd"))
	s.GridRowStart = int(getIntField(L, absIdx, "gridRowStart"))
	s.GridRowEnd = int(getIntField(L, absIdx, "gridRowEnd"))
	return s
}

// readStyleFields reads style fields from the top-level table (not a nested "style" sub-table).
// Only sets fields that are still at their zero/default value.
func (e *Engine) readStyleFields(L *lua.State, idx int, s *Style) {
	absIdx := L.AbsIndex(idx)

	if s.Width == 0 {
		s.Width = int(getIntField(L, absIdx, "width"))
	}
	if s.Height == 0 {
		s.Height = int(getIntField(L, absIdx, "height"))
	}
	// Parse percentage/viewport strings for width/height if still unset
	if s.Width == 0 && s.WidthPercent == 0 && s.WidthVW == 0 {
		if str := getStringField(L, absIdx, "width"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.WidthPercent = pct
			} else if v, unit, ok := parseViewport(str); ok && unit == "vw" {
				s.WidthVW = v
			}
		}
	}
	if s.Height == 0 && s.HeightPercent == 0 && s.HeightVH == 0 {
		if str := getStringField(L, absIdx, "height"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.HeightPercent = pct
			} else if v, unit, ok := parseViewport(str); ok && unit == "vh" {
				s.HeightVH = v
			}
		}
	}
	if s.MinWidth == 0 {
		s.MinWidth = int(getIntField(L, absIdx, "minWidth"))
	}
	if s.MinHeight == 0 {
		s.MinHeight = int(getIntField(L, absIdx, "minHeight"))
	}
	if s.MaxWidth == 0 {
		s.MaxWidth = int(getIntField(L, absIdx, "maxWidth"))
	}
	if s.MaxHeight == 0 {
		s.MaxHeight = int(getIntField(L, absIdx, "maxHeight"))
	}
	// Parse percentage strings for min/max
	if s.MinWidth == 0 && s.MinWidthPercent == 0 {
		if str := getStringField(L, absIdx, "minWidth"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.MinWidthPercent = pct
			}
		}
	}
	if s.MaxWidth == 0 && s.MaxWidthPercent == 0 {
		if str := getStringField(L, absIdx, "maxWidth"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.MaxWidthPercent = pct
			}
		}
	}
	if s.MinHeight == 0 && s.MinHeightPercent == 0 {
		if str := getStringField(L, absIdx, "minHeight"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.MinHeightPercent = pct
			}
		}
	}
	if s.MaxHeight == 0 && s.MaxHeightPercent == 0 {
		if str := getStringField(L, absIdx, "maxHeight"); str != "" {
			if pct, ok := parsePercent(str); ok {
				s.MaxHeightPercent = pct
			}
		}
	}
	if s.Flex == 0 {
		s.Flex = int(getIntField(L, absIdx, "flex"))
	}
	if s.FlexShrink == 0 {
		s.FlexShrink = int(getIntField(L, absIdx, "flexShrink"))
	}
	if s.FlexBasis == 0 {
		s.FlexBasis = int(getIntField(L, absIdx, "flexBasis"))
	}
	if s.Gap == 0 {
		s.Gap = int(getIntField(L, absIdx, "gap"))
	}
	if s.Padding == 0 {
		s.Padding = int(getIntField(L, absIdx, "padding"))
	}
	if s.PaddingTop == 0 {
		s.PaddingTop = int(getIntField(L, absIdx, "paddingTop"))
	}
	if s.PaddingRight == 0 {
		s.PaddingRight = int(getIntField(L, absIdx, "paddingRight"))
	}
	if s.PaddingBottom == 0 {
		s.PaddingBottom = int(getIntField(L, absIdx, "paddingBottom"))
	}
	if s.PaddingLeft == 0 {
		s.PaddingLeft = int(getIntField(L, absIdx, "paddingLeft"))
	}
	if s.Margin == 0 {
		s.Margin = int(getIntField(L, absIdx, "margin"))
	}
	if s.MarginTop == 0 {
		s.MarginTop = int(getIntField(L, absIdx, "marginTop"))
	}
	if s.MarginRight == 0 {
		s.MarginRight = int(getIntField(L, absIdx, "marginRight"))
	}
	if s.MarginBottom == 0 {
		s.MarginBottom = int(getIntField(L, absIdx, "marginBottom"))
	}
	if s.MarginLeft == 0 {
		s.MarginLeft = int(getIntField(L, absIdx, "marginLeft"))
	}
	if s.Foreground == "" {
		s.Foreground = getStringField(L, absIdx, "foreground")
		if s.Foreground == "" {
			s.Foreground = getStringField(L, absIdx, "fg")
		}
	}
	if s.Background == "" {
		s.Background = getStringField(L, absIdx, "background")
		if s.Background == "" {
			s.Background = getStringField(L, absIdx, "bg")
		}
	}
	if s.Border == "" {
		s.Border = getStringField(L, absIdx, "border")
	}
	if s.Justify == "" {
		s.Justify = getStringField(L, absIdx, "justify")
	}
	if s.Align == "" {
		s.Align = getStringField(L, absIdx, "align")
	}
	if s.AlignSelf == "" {
		s.AlignSelf = getStringField(L, absIdx, "alignSelf")
	}
	if s.Order == 0 {
		s.Order = int(getIntField(L, absIdx, "order"))
	}
	if s.Overflow == "" {
		s.Overflow = getStringField(L, absIdx, "overflow")
	}
	if s.Position == "" {
		s.Position = getStringField(L, absIdx, "position")
	}
	if !s.Bold {
		s.Bold = getBoolField(L, absIdx, "bold")
	}
	if !s.Dim {
		s.Dim = getBoolField(L, absIdx, "dim")
	}
	if !s.Underline {
		s.Underline = getBoolField(L, absIdx, "underline")
	}
	if !s.Italic {
		s.Italic = getBoolField(L, absIdx, "italic")
	}
	if !s.Strikethrough {
		s.Strikethrough = getBoolField(L, absIdx, "strikethrough")
	}
	if !s.Inverse {
		s.Inverse = getBoolField(L, absIdx, "inverse")
	}
	if s.Top == 0 {
		s.Top = int(getIntField(L, absIdx, "top"))
	}
	if s.Left == 0 {
		s.Left = int(getIntField(L, absIdx, "left"))
	}
	if s.Right == -1 {
		s.Right = int(getIntFieldDefault(L, absIdx, "right", -1))
	}
	if s.Bottom == -1 {
		s.Bottom = int(getIntFieldDefault(L, absIdx, "bottom", -1))
	}
	if s.ZIndex == 0 {
		s.ZIndex = int(getIntField(L, absIdx, "zIndex"))
	}
	if s.TextAlign == "" {
		s.TextAlign = getStringField(L, absIdx, "textAlign")
	}
	if s.TextOverflow == "" {
		s.TextOverflow = getStringField(L, absIdx, "textOverflow")
	}
	if s.WhiteSpace == "" {
		s.WhiteSpace = getStringField(L, absIdx, "whiteSpace")
	}
	if s.Display == "" {
		s.Display = getStringField(L, absIdx, "display")
	}
	if s.Visibility == "" {
		s.Visibility = getStringField(L, absIdx, "visibility")
	}
	if s.BorderColor == "" {
		s.BorderColor = getStringField(L, absIdx, "borderColor")
	}
	if s.FlexWrap == "" {
		s.FlexWrap = getStringField(L, absIdx, "flexWrap")
	}
	// Grid container properties
	if s.GridTemplateColumns == "" {
		s.GridTemplateColumns = getStringField(L, absIdx, "gridTemplateColumns")
	}
	if s.GridTemplateRows == "" {
		s.GridTemplateRows = getStringField(L, absIdx, "gridTemplateRows")
	}
	if s.GridColumnGap == 0 {
		s.GridColumnGap = int(getIntField(L, absIdx, "gridColumnGap"))
	}
	if s.GridRowGap == 0 {
		s.GridRowGap = int(getIntField(L, absIdx, "gridRowGap"))
	}
	// Grid item properties
	if s.GridColumn == "" {
		s.GridColumn = getStringField(L, absIdx, "gridColumn")
	}
	if s.GridRow == "" {
		s.GridRow = getStringField(L, absIdx, "gridRow")
	}
	if s.GridColumnStart == 0 {
		s.GridColumnStart = int(getIntField(L, absIdx, "gridColumnStart"))
	}
	if s.GridColumnEnd == 0 {
		s.GridColumnEnd = int(getIntField(L, absIdx, "gridColumnEnd"))
	}
	if s.GridRowStart == 0 {
		s.GridRowStart = int(getIntField(L, absIdx, "gridRowStart"))
	}
	if s.GridRowEnd == 0 {
		s.GridRowEnd = int(getIntField(L, absIdx, "gridRowEnd"))
	}
}

// reconcileChildComponents walks the RenderNode tree looking for component-type
// nodes and reconciles child components.


// collectPropFuncRefsFromAny walks values produced by readMapFromTable / readPropTable
// and collects nested propFuncRef ids (registry indices).
func collectPropFuncRefsFromAny(v any, out *[]int64) {
	switch x := v.(type) {
	case propFuncRef:
		if x != 0 {
			*out = append(*out, int64(x))
		}
	case map[string]any:
		for _, vv := range x {
			collectPropFuncRefsFromAny(vv, out)
		}
	case []any:
		for _, vv := range x {
			collectPropFuncRefsFromAny(vv, out)
		}
	}
}

// unrefPropFuncRefsInProps releases Lua registry refs held in a props map. Must be
// called before dropping the map: each render builds a new readMapFromTable result
// with new Ref() ids for the same logical functions, so propsEqual is usually false
// every frame and child.Props is reassigned without otherwise freeing old refs.
//
// NOTE: propsEqual currently compares propFuncRef by value (ref ID). Because each
// render produces fresh Ref() IDs, props with callbacks are always "different",
// triggering re-render and old-ref cleanup. If we ever cache/reuse Lua refs across
// renders (e.g. for performance), propsEqual would return true and the old refs
// would NOT be freed — causing a leak. Any such optimization must update the
// cleanup logic here.
func unrefPropFuncRefsInProps(L *lua.State, m map[string]any) {
	if L == nil || m == nil {
		return
	}
	var refs []int64
	collectPropFuncRefsFromAny(m, &refs)
	for _, ref := range refs {
		L.Unref(lua.RegistryIndex, int(ref))
	}
}

func pushMap(L *lua.State, m map[string]any) {
	if m == nil {
		L.NewTable()
		return
	}
	L.CreateTable(0, len(m))
	for k, v := range m {
		L.PushString(k)
		switch vv := v.(type) {
		case propFuncRef:
			if vv != 0 {
				L.RawGetI(lua.RegistryIndex, int64(vv))
			} else {
				L.PushNil()
			}
		default:
			pushPropValue(L, v)
		}
		L.SetTable(-3)
	}
}

// pushPropValue pushes a value stored in ComponentProps, preserving nested
// propFuncRef (Lua functions) that plain PushAny cannot represent.
func pushPropValue(L *lua.State, v any) {
	if v == nil {
		L.PushNil()
		return
	}
	if pf, ok := v.(propFuncRef); ok {
		if pf != 0 {
			L.RawGetI(lua.RegistryIndex, int64(pf))
		} else {
			L.PushNil()
		}
		return
	}
	switch vv := v.(type) {
	case map[string]any:
		L.CreateTable(0, len(vv))
		for k, val := range vv {
			L.PushString(k)
			pushPropValue(L, val)
			L.RawSet(-3)
		}
	case []any:
		L.CreateTable(len(vv), 0)
		for i, val := range vv {
			pushPropValue(L, val)
			L.RawSetI(-2, int64(i+1))
		}
	default:
		L.PushAny(v)
	}
}

func getStringField(L *lua.State, idx int, field string) string {
	return L.GetFieldString(idx, field)
}

func getIntField(L *lua.State, idx int, field string) int64 {
	return L.GetFieldInt(idx, field)
}

func getIntFieldDefault(L *lua.State, idx int, field string, def int64) int64 {
	L.GetField(idx, field)
	if L.IsNoneOrNil(-1) {
		L.Pop(1)
		return def
	}
	n, _ := L.ToInteger(-1)
	L.Pop(1)
	return n
}

func getBoolField(L *lua.State, idx int, field string) bool {
	return L.GetFieldBool(idx, field)
}

func getRefField(L *lua.State, idx int, field string) int64 {
	L.GetField(idx, field)
	if L.IsFunction(-1) {
		ref := L.Ref(lua.RegistryIndex)
		return int64(ref)
	}
	L.Pop(1)
	return 0
}

func readMapFromTable(L *lua.State, idx int) map[string]any {
	m := make(map[string]any)
	absIdx := L.AbsIndex(idx)
	L.ForEach(absIdx, func(L *lua.State) bool {
		if L.Type(-2) == lua.TypeString {
			key, _ := L.ToString(-2)
			m[key] = readPropValueFromStack(L)
		}
		return true
	})
	return m
}

func luaPropTableKeyString(L *lua.State, keyIdx int) string {
	switch L.Type(keyIdx) {
	case lua.TypeString:
		s, _ := L.ToString(keyIdx)
		return s
	case lua.TypeNumber:
		if L.IsInteger(keyIdx) {
			v, _ := L.ToInteger(keyIdx)
			return strconv.FormatInt(v, 10)
		}
		v, _ := L.ToNumber(keyIdx)
		return strconv.FormatFloat(v, 'g', -1, 64)
	default:
		s, _ := L.ToString(keyIdx)
		return s
	}
}

// readPropValueFromStack reads the Lua value at stack index -1 (without popping it).
// Used for ComponentProps so nested descriptor tables keep onClick etc. as propFuncRef.
// (L.ToAny maps Lua functions to nil.)
func readPropValueFromStack(L *lua.State) any {
	switch L.Type(-1) {
	case lua.TypeNil:
		return nil
	case lua.TypeBoolean:
		return L.ToBoolean(-1)
	case lua.TypeNumber:
		if L.IsInteger(-1) {
			v, _ := L.ToInteger(-1)
			return v
		}
		v, _ := L.ToNumber(-1)
		return v
	case lua.TypeString:
		s, _ := L.ToString(-1)
		return s
	case lua.TypeFunction:
		L.PushValue(-1)
		ref := L.Ref(lua.RegistryIndex)
		return propFuncRef(ref)
	case lua.TypeTable:
		L.PushValue(-1)
		tIdx := L.AbsIndex(-1)
		out := readPropTable(L, tIdx)
		L.Pop(1)
		return out
	default:
		return L.ToAny(-1)
	}
}

func readPropTable(L *lua.State, idx int) any {
	idx = L.AbsIndex(idx)
	length := int(L.LenI(idx))
	if length > 0 {
		var count int64
		L.PushNil()
		for L.Next(idx) {
			count++
			L.Pop(1)
		}
		if count == int64(length) {
			arr := make([]any, length)
			for i := 1; i <= length; i++ {
				L.RawGetI(idx, int64(i))
				arr[i-1] = readPropValueFromStack(L)
				L.Pop(1)
			}
			return arr
		}
	}
	m := make(map[string]any)
	L.PushNil()
	for L.Next(idx) {
		key := luaPropTableKeyString(L, -2)
		m[key] = readPropValueFromStack(L)
		L.Pop(1)
	}
	return m
}

