package component

import (
	"strings"

	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// ExtractHandlers walks the VNode tree and collects event handlers and
// focusable VNode IDs into the Component's Handlers and Focusables fields.
func (c *Component) ExtractHandlers() {
	c.Handlers = make(map[string]event.HandlerMap)
	c.Focusables = nil
	walkExtract(c.VNodeTree, c)
}

// walkExtract recursively walks the VNode tree, extracting event handlers
// from VNode Props and registering focusable nodes.
func walkExtract(vn *layout.VNode, c *Component) {
	if vn == nil {
		return
	}
	if vn.ID != "" && vn.Props != nil {
		hm := make(event.HandlerMap)
		for _, evtName := range []string{
			"onClick", "onMouseEnter", "onMouseLeave",
			"onKeyDown", "onMouseDown", "onMouseUp", "onMouseMove",
		} {
			if fn, ok := vn.Props[evtName]; ok {
				if handler, ok := fn.(event.EventHandler); ok {
					hm[mapEventName(evtName)] = handler
				}
			}
		}
		if len(hm) > 0 {
			c.Handlers[vn.ID] = hm
		}
		// Check focusable.
		if _, ok := vn.Props["focusable"]; ok {
			c.Focusables = append(c.Focusables, vn.ID)
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

// shallowEqual compares two maps by key count and value identity (==).
func shallowEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok || va != vb {
			return false
		}
	}
	return true
}
