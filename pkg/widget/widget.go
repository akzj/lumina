package widget

import "github.com/akzj/lumina/pkg/render"

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
	OnEvent  func(props map[string]any, state any, event *render.WidgetEvent) bool
	NewState func() any
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
func (w *Widget) DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool {
	if w.OnEvent == nil {
		return false
	}
	return w.OnEvent(props, state, event)
}
