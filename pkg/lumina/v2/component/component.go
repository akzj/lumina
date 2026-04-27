package component

import (
	"strings"

	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// ExtractHandlers walks the VNode tree and collects event handlers and
// focusable VNode IDs into the Component's handlers and focusables fields.
func (c *Component) ExtractHandlers() {
	c.handlers = make(map[string]HandlerMap)
	c.focusables = nil
	walkExtract(c.vnodeTree, c)
}

// walkExtract recursively walks the VNode tree, extracting event handlers
// from VNode Props and registering focusable nodes.
func walkExtract(vn *layout.VNode, c *Component) {
	if vn == nil {
		return
	}
	if vn.ID != "" && vn.Props != nil {
		hm := make(HandlerMap)
		for _, evtName := range []string{
			"onClick", "onMouseEnter", "onMouseLeave",
			"onKeyDown", "onMouseDown", "onMouseUp", "onMouseMove",
			"onScroll",
		} {
			if fn, ok := vn.Props[evtName]; ok {
				// Store as HandlerFunc (any) — the event package will
				// type-assert to event.EventHandler when dispatching.
				hm[mapEventName(evtName)] = fn
			}
		}
		if len(hm) > 0 {
			c.handlers[vn.ID] = hm
		}
		// Check focusable. Input and textarea are auto-focusable.
		if _, ok := vn.Props["focusable"]; ok {
			c.focusables = append(c.focusables, vn.ID)
		} else if vn.ID != "" && (vn.Type == "input" || vn.Type == "textarea") {
			c.focusables = append(c.focusables, vn.ID)
		}
	}
	for _, child := range vn.Children {
		walkExtract(child, c)
	}
}

// mapEventName converts a camelCase prop name like "onClick" to the
// corresponding event type "click".
//
//	"onClick"      → "click"
//	"onMouseEnter" → "mouseenter"
//	"onKeyDown"    → "keydown"
func mapEventName(propName string) string {
	// Strip the "on" prefix.
	name := strings.TrimPrefix(propName, "on")
	// Lowercase the whole thing.
	return strings.ToLower(name)
}

// stateEqual compares two any values for equality. For comparable types
// (string, int, float64, bool, nil) it uses ==. For uncomparable types
// (slices, maps) it returns false — a new value always triggers re-render.
func stateEqual(a, b any) bool {
	// Fast path: both nil.
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Use a recover guard: == panics on uncomparable types like []interface{}.
	defer func() { recover() }()
	return a == b
}

// shallowEqual compares two maps by key count and value identity (==).
// Uses stateEqual to safely handle uncomparable value types.
func shallowEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok || !stateEqual(va, vb) {
			return false
		}
	}
	return true
}
