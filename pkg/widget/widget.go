package widget

// Widget defines a built-in control with Go-side interaction logic.
// Widgets integrate into the existing component system: they are registered
// as factories and rendered by the engine like Lua components, but their
// Render function returns a node tree directly (no Lua call).
//
// The Node/Style types are referenced via the WidgetDef interface in the
// render package to avoid import cycles. This package provides concrete
// implementations.
type Widget struct {
	Name     string
	Render   func(props map[string]any, state any) any // returns *render.Node (as any to avoid import cycle)
	OnEvent  func(props map[string]any, state any, event *Event) bool
	NewState func() any
}

// Event is a widget-level event dispatched by the engine.
type Event struct {
	Type string // "click", "mousedown", "mouseup", "mouseenter", "mouseleave", "keydown", "focus", "blur"
	Key  string
	X, Y int
}

// GetName returns the widget name.
func (w *Widget) GetName() string { return w.Name }

// GetNewState returns a new state instance.
func (w *Widget) GetNewState() any { return w.NewState() }

// DoRender calls the widget's render function.
func (w *Widget) DoRender(props map[string]any, state any) any {
	return w.Render(props, state)
}

// DoOnEvent calls the widget's event handler.
func (w *Widget) DoOnEvent(props map[string]any, state any, eventType, key string, x, y int) bool {
	if w.OnEvent == nil {
		return false
	}
	return w.OnEvent(props, state, &Event{Type: eventType, Key: key, X: x, Y: y})
}
