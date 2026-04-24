package lumina

import (
	"fmt"
	"runtime/debug"
)

// RecoverFunc wraps a function with panic recovery, logging the error
// and returning it instead of crashing.
func RecoverFunc(name string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err = fmt.Errorf("[%s] panic recovered: %v\n%s", name, r, stack)
		}
	}()
	return fn()
}

// RecoverValue wraps a function that returns (T, error) with panic recovery.
func RecoverValue[T any](name string, fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err = fmt.Errorf("[%s] panic recovered: %v\n%s", name, r, stack)
		}
	}()
	return fn()
}

// SafeRenderComponent wraps component rendering with panic recovery.
func SafeRenderComponent(name string, renderFn func() (VNode, error)) (VNode, error) {
	return RecoverValue[VNode]("renderComponent:"+name, renderFn)
}

// SafeHandleEvent wraps event handling with panic recovery.
func SafeHandleEvent(eventType string, handler func() error) error {
	return RecoverFunc("handleEvent:"+eventType, handler)
}

// SafeWebSocket wraps WebSocket operations with panic recovery.
func SafeWebSocket(op string, fn func() error) error {
	return RecoverFunc("websocket:"+op, fn)
}

// ErrorNode returns a VNode that displays an error message.
// Used as fallback when component rendering fails.
func ErrorNode(componentName string, err error) VNode {
	return VNode{
		Type:    "text",
		Content: fmt.Sprintf("⚠ Error in %s: %v", componentName, err),
	}
}
