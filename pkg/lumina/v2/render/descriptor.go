package render

// Descriptor is a lightweight description of a UI element.
// In production, this is read directly from the Lua stack without allocation.
// This Go struct is used for testing the reconciler.
type Descriptor struct {
	Type     string       // "box", "vbox", "hbox", "text", "component"
	ID       string
	Key      string
	Content  string
	Style    Style
	Children []Descriptor

	// Event handler refs (Lua registry refs)
	OnClick      LuaRef
	OnMouseEnter LuaRef
	OnMouseLeave LuaRef
	OnKeyDown    LuaRef
	OnChange     LuaRef
	OnScroll     LuaRef

	// Component-specific (when Type == "component")
	ComponentType  string // factory name
	ComponentProps map[string]any
}
