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

	// Event handlers (stored as propFuncRef by readPropValueFromStack)
	if ref, ok := m["onClick"].(propFuncRef); ok {
		desc.OnClick = LuaRef(ref)
	}
	if ref, ok := m["onMouseEnter"].(propFuncRef); ok {
		desc.OnMouseEnter = LuaRef(ref)
	}
	if ref, ok := m["onMouseLeave"].(propFuncRef); ok {
		desc.OnMouseLeave = LuaRef(ref)
	}
	if ref, ok := m["onKeyDown"].(propFuncRef); ok {
		desc.OnKeyDown = LuaRef(ref)
	}
	if ref, ok := m["onChange"].(propFuncRef); ok {
		desc.OnChange = LuaRef(ref)
	}
	if ref, ok := m["onScroll"].(propFuncRef); ok {
		desc.OnScroll = LuaRef(ref)
	}
	if ref, ok := m["onMouseDown"].(propFuncRef); ok {
		desc.OnMouseDown = LuaRef(ref)
	}
	if ref, ok := m["onMouseUp"].(propFuncRef); ok {
		desc.OnMouseUp = LuaRef(ref)
	}
	if ref, ok := m["onFocus"].(propFuncRef); ok {
		desc.OnFocus = LuaRef(ref)
	}
	if ref, ok := m["onBlur"].(propFuncRef); ok {
		desc.OnBlur = LuaRef(ref)
	}
	if ref, ok := m["onSubmit"].(propFuncRef); ok {
		desc.OnSubmit = LuaRef(ref)
	}
	if ref, ok := m["onOutsideClick"].(propFuncRef); ok {
		desc.OnOutsideClick = LuaRef(ref)
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

	// Also read top-level style fields (they override the style sub-table)
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

// readStyleFields reads style fields from the top-level element table.
// Top-level fields override the style sub-table (more specific wins).
// Only writes when the field is explicitly present in the Lua table.
func (e *Engine) readStyleFields(L *lua.State, idx int, s *Style) {
	absIdx := L.AbsIndex(idx)

	// Width/Height: check for int first, then string (percentage/viewport)
	if n, ok := getIntFieldIfPresent(L, absIdx, "width"); ok {
		if n != 0 {
			s.Width = int(n)
			s.WidthPercent = 0
			s.WidthVW = 0
		} else {
			// Explicit width=0 could be a string value
			if str, sok := getStringFieldIfPresent(L, absIdx, "width"); sok && str != "" {
				if pct, pok := parsePercent(str); pok {
					s.WidthPercent = pct
					s.Width = 0
				} else if v, unit, vok := parseViewport(str); vok && unit == "vw" {
					s.WidthVW = v
					s.Width = 0
				}
			}
		}
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "height"); ok {
		if n != 0 {
			s.Height = int(n)
			s.HeightPercent = 0
			s.HeightVH = 0
		} else {
			if str, sok := getStringFieldIfPresent(L, absIdx, "height"); sok && str != "" {
				if pct, pok := parsePercent(str); pok {
					s.HeightPercent = pct
					s.Height = 0
				} else if v, unit, vok := parseViewport(str); vok && unit == "vh" {
					s.HeightVH = v
					s.Height = 0
				}
			}
		}
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "minWidth"); ok {
		s.MinWidth = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "minHeight"); ok {
		s.MinHeight = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "maxWidth"); ok {
		s.MaxWidth = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "maxHeight"); ok {
		s.MaxHeight = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "flex"); ok {
		s.Flex = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "flexShrink"); ok {
		s.FlexShrink = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "flexBasis"); ok {
		s.FlexBasis = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "gap"); ok {
		s.Gap = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "padding"); ok {
		s.Padding = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "paddingTop"); ok {
		s.PaddingTop = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "paddingRight"); ok {
		s.PaddingRight = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "paddingBottom"); ok {
		s.PaddingBottom = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "paddingLeft"); ok {
		s.PaddingLeft = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "margin"); ok {
		s.Margin = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "marginTop"); ok {
		s.MarginTop = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "marginRight"); ok {
		s.MarginRight = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "marginBottom"); ok {
		s.MarginBottom = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "marginLeft"); ok {
		s.MarginLeft = int(n)
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "foreground"); ok {
		s.Foreground = str
	}
	if s.Foreground == "" {
		if str, ok := getStringFieldIfPresent(L, absIdx, "fg"); ok {
			s.Foreground = str
		}
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "background"); ok {
		s.Background = str
	}
	if s.Background == "" {
		if str, ok := getStringFieldIfPresent(L, absIdx, "bg"); ok {
			s.Background = str
		}
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "border"); ok {
		s.Border = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "justify"); ok {
		s.Justify = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "align"); ok {
		s.Align = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "alignSelf"); ok {
		s.AlignSelf = str
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "order"); ok {
		s.Order = int(n)
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "overflow"); ok {
		s.Overflow = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "position"); ok {
		s.Position = str
	}
	if b, ok := getBoolFieldIfPresent(L, absIdx, "bold"); ok {
		s.Bold = b
	}
	if b, ok := getBoolFieldIfPresent(L, absIdx, "dim"); ok {
		s.Dim = b
	}
	if b, ok := getBoolFieldIfPresent(L, absIdx, "underline"); ok {
		s.Underline = b
	}
	if b, ok := getBoolFieldIfPresent(L, absIdx, "italic"); ok {
		s.Italic = b
	}
	if b, ok := getBoolFieldIfPresent(L, absIdx, "strikethrough"); ok {
		s.Strikethrough = b
	}
	if b, ok := getBoolFieldIfPresent(L, absIdx, "inverse"); ok {
		s.Inverse = b
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "top"); ok {
		s.Top = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "left"); ok {
		s.Left = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "right"); ok {
		s.Right = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "bottom"); ok {
		s.Bottom = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "zIndex"); ok {
		s.ZIndex = int(n)
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "textAlign"); ok {
		s.TextAlign = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "textOverflow"); ok {
		s.TextOverflow = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "whiteSpace"); ok {
		s.WhiteSpace = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "display"); ok {
		s.Display = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "visibility"); ok {
		s.Visibility = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "borderColor"); ok {
		s.BorderColor = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "flexWrap"); ok {
		s.FlexWrap = str
	}
	// Grid container properties
	if str, ok := getStringFieldIfPresent(L, absIdx, "gridTemplateColumns"); ok {
		s.GridTemplateColumns = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "gridTemplateRows"); ok {
		s.GridTemplateRows = str
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "gridColumnGap"); ok {
		s.GridColumnGap = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "gridRowGap"); ok {
		s.GridRowGap = int(n)
	}
	// Grid item properties
	if str, ok := getStringFieldIfPresent(L, absIdx, "gridColumn"); ok {
		s.GridColumn = str
	}
	if str, ok := getStringFieldIfPresent(L, absIdx, "gridRow"); ok {
		s.GridRow = str
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "gridColumnStart"); ok {
		s.GridColumnStart = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "gridColumnEnd"); ok {
		s.GridColumnEnd = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "gridRowStart"); ok {
		s.GridRowStart = int(n)
	}
	if n, ok := getIntFieldIfPresent(L, absIdx, "gridRowEnd"); ok {
		s.GridRowEnd = int(n)
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

// getIntFieldIfPresent reads an int field only if it exists (not nil) in the table.
// Returns the value and true if present, or 0 and false if absent.
func getIntFieldIfPresent(L *lua.State, idx int, field string) (int64, bool) {
	L.GetField(idx, field)
	if L.IsNoneOrNil(-1) {
		L.Pop(1)
		return 0, false
	}
	n, _ := L.ToInteger(-1)
	L.Pop(1)
	return n, true
}

// getStringFieldIfPresent reads a string field only if it exists (not nil) in the table.
// Returns the value and true if present, or "" and false if absent.
func getStringFieldIfPresent(L *lua.State, idx int, field string) (string, bool) {
	L.GetField(idx, field)
	if L.IsNoneOrNil(-1) {
		L.Pop(1)
		return "", false
	}
	s, _ := L.ToString(-1)
	L.Pop(1)
	return s, true
}

// getBoolFieldIfPresent reads a bool field only if it exists (not nil) in the table.
// Returns the value and true if present, or false and false if absent.
func getBoolFieldIfPresent(L *lua.State, idx int, field string) (bool, bool) {
	L.GetField(idx, field)
	if L.IsNoneOrNil(-1) {
		L.Pop(1)
		return false, false
	}
	b := L.ToBoolean(-1)
	L.Pop(1)
	return b, true
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
// maxPropDepth limits recursion depth when reading nested Lua tables.
// Prevents stack overflow from self-referencing tables.
const maxPropDepth = 20

func readPropValueFromStack(L *lua.State) any {
	return readPropValueFromStackDepth(L, 0)
}

func readPropValueFromStackDepth(L *lua.State, depth int) any {
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
		out := readPropTableDepth(L, tIdx, depth+1)
		L.Pop(1)
		return out
	default:
		return L.ToAny(-1)
	}
}

func readPropTable(L *lua.State, idx int) any {
	return readPropTableDepth(L, idx, 0)
}

func readPropTableDepth(L *lua.State, idx int, depth int) any {
	if depth > maxPropDepth {
		return nil // prevent infinite recursion from circular tables
	}
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
				arr[i-1] = readPropValueFromStackDepth(L, depth)
				L.Pop(1)
			}
			return arr
		}
	}
	m := make(map[string]any)
	L.PushNil()
	for L.Next(idx) {
		key := luaPropTableKeyString(L, -2)
		m[key] = readPropValueFromStackDepth(L, depth)
		L.Pop(1)
	}
	return m
}

