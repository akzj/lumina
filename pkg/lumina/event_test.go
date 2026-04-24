package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/stretchr/testify/assert"
)

func TestEventBus(t *testing.T) {
	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
	}

	called := false
	eb.On("click", "test", func(e *Event) {
		called = true
		assert.Equal(t, "click", e.Type)
	})

	eb.Emit(&Event{Type: "click", Target: "test"})
	assert.True(t, called, "handler should be called")
}

func TestEventBusMultipleHandlers(t *testing.T) {
	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
	}

	count1, count2 := 0, 0
	eb.On("keydown", "comp1", func(e *Event) { count1++ })
	eb.On("keydown", "comp2", func(e *Event) { count2++ })

	eb.Emit(&Event{Type: "keydown", Key: "a"})
	assert.Equal(t, 1, count1)
	assert.Equal(t, 1, count2)
}

func TestEventBusOff(t *testing.T) {
	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
	}

	count := 0
	handler := func(e *Event) { count++ }
	eb.On("click", "test", handler)

	eb.Emit(&Event{Type: "click"})
	assert.Equal(t, 1, count)

	eb.Off("click", "test")
	eb.Emit(&Event{Type: "click"})
	assert.Equal(t, 1, count, "handler should not be called after off()")
}

func TestShortcutNormalization(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"Ctrl+C", "ctrl+c"},
		{"SHIFT+A", "shift+a"},
		{"Escape", "escape"},
		{"ctrl-s", "ctrl-s"},
		{"Alt+F4", "alt+f4"},
	}

	for _, tc := range tests {
		result := normalizeShortcutKey(tc.input)
		assert.Equal(t, tc.expected, result, "input: %s", tc.input)
	}
}

func TestFocusManagement(t *testing.T) {
	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
	}

	focusEvents := []string{}
	eb.On("focus", "", func(e *Event) {
		focusEvents = append(focusEvents, e.Type+":"+e.Target)
	})
	eb.On("blur", "", func(e *Event) {
		focusEvents = append(focusEvents, e.Type+":"+e.Target)
	})

	eb.SetFocus("comp1")
	assert.Equal(t, "comp1", eb.GetFocused())
	assert.Contains(t, focusEvents, "focus:comp1")

	eb.SetFocus("comp2")
	assert.Equal(t, "comp2", eb.GetFocused())
	assert.Contains(t, focusEvents, "blur:comp1")
	assert.Contains(t, focusEvents, "focus:comp2")
}

func TestBlur(t *testing.T) {
	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
	}

	blurCalled := false
	eb.On("blur", "", func(e *Event) {
		blurCalled = true
	})

	eb.SetFocus("comp1")
	eb.Blur()
	assert.Equal(t, "", eb.GetFocused(), "should have no focus after blur")
	assert.True(t, blurCalled, "blur event should be called")
}

func TestFocusTrap(t *testing.T) {
	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
	}

	eb.PushFocusTrap("dialog1")
	eb.PushFocusTrap("dialog2")
	eb.PopFocusTrap()

	assert.Equal(t, 1, len(eb.focusStack))
}

func TestEventPropagation(t *testing.T) {
	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
	}

	calls := 0
	eb.On("click", "", func(e *Event) {
		calls++
		if calls == 1 {
			e.StopPropagation()
		}
	})
	eb.On("click", "", func(e *Event) {
		calls++
	})

	eb.Emit(&Event{Type: "click"})
	assert.Equal(t, 1, calls, "propagation should be stopped")
}

func TestOnLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local called = false
		lumina.on("click", "test", function(event)
			called = true
			assert(event.type == "click", "event type should be click")
		end)
		lumina.emit("test", "click", { x = 10, y = 20 })
		assert(called, "handler should be called")
	`)
	assert.NoError(t, err)
}

func TestRegisterShortcutLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local called = false
		lumina.registerShortcut({ key = "ctrl+c" }, function(event)
			called = true
		end)
		lumina.emitKeyEvent("c", { ctrl = true })
		assert(called, "shortcut handler should be called")
	`)
	assert.NoError(t, err)
}

func TestSetFocusLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		lumina.setFocus("input1")
		local focused = lumina.getFocused()
		assert(focused == "input1", "should be focused: input1, got: " .. tostring(focused))
	`)
	assert.NoError(t, err)
}

func TestCreateEvent(t *testing.T) {
	e := CreateEvent("keydown")
	assert.Equal(t, "keydown", e.Type)
	assert.True(t, e.Timestamp > 0)
}

func TestEventStopPropagation(t *testing.T) {
	e := CreateEvent("test")
	assert.False(t, e.IsStopped())
	e.StopPropagation()
	assert.True(t, e.IsStopped())
}
