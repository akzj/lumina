package lumina

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TransitionState represents the state of a useTransition hook.
type TransitionState struct {
	IsPending bool
	mu        sync.Mutex
}

// NewTransitionState creates a new TransitionState.
func NewTransitionState() *TransitionState {
	return &TransitionState{}
}

// Start marks the transition as pending, runs the callback, then marks it as done.
func (ts *TransitionState) Start(callback func()) {
	ts.mu.Lock()
	ts.IsPending = true
	ts.mu.Unlock()

	callback()

	ts.mu.Lock()
	ts.IsPending = false
	ts.mu.Unlock()
}

// Pending returns whether the transition is pending.
func (ts *TransitionState) Pending() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.IsPending
}

// DeferredValue holds a value that can be deferred for expensive re-renders.
type DeferredValue struct {
	mu       sync.Mutex
	current  any
	deferred any
	timeout  time.Duration
	timer    *time.Timer
}

// NewDeferredValue creates a DeferredValue with the given initial value and timeout.
func NewDeferredValue(initial any, timeout time.Duration) *DeferredValue {
	return &DeferredValue{
		current:  initial,
		deferred: initial,
		timeout:  timeout,
	}
}

// Update sets a new current value. The deferred value will update after the timeout.
func (dv *DeferredValue) Update(value any) {
	dv.mu.Lock()
	defer dv.mu.Unlock()
	dv.current = value

	if dv.timer != nil {
		dv.timer.Stop()
	}

	dv.timer = time.AfterFunc(dv.timeout, func() {
		dv.mu.Lock()
		defer dv.mu.Unlock()
		dv.deferred = dv.current
	})
}

// SetImmediate updates both current and deferred immediately (for sync use).
func (dv *DeferredValue) SetImmediate(value any) {
	dv.mu.Lock()
	defer dv.mu.Unlock()
	dv.current = value
	dv.deferred = value
}

// Current returns the current (latest) value.
func (dv *DeferredValue) Current() any {
	dv.mu.Lock()
	defer dv.mu.Unlock()
	return dv.current
}

// Deferred returns the deferred value (may lag behind current).
func (dv *DeferredValue) Deferred() any {
	dv.mu.Lock()
	defer dv.mu.Unlock()
	return dv.deferred
}

// IDGenerator generates unique IDs for useId.
var globalIDCounter atomic.Int64

// GenerateID returns a unique ID string.
func GenerateID() string {
	id := globalIDCounter.Add(1)
	return fmt.Sprintf("lumina-%d", id)
}

// ResetIDCounter resets the ID counter (for testing).
func ResetIDCounter() {
	globalIDCounter.Store(0)
}
