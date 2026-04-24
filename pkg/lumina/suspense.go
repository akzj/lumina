package lumina

import (
	"fmt"
	"sync"
)

// Promise represents an asynchronous value that can be pending, resolved, or rejected.
type Promise struct {
	mu        sync.Mutex
	id        string
	status    string // "pending" | "resolved" | "rejected"
	value     any
	err       error
	onSettle  []func()
}

// NewPromise creates a new pending Promise.
func NewPromise(id string) *Promise {
	return &Promise{
		id:     id,
		status: "pending",
	}
}

// ID returns the promise ID.
func (p *Promise) ID() string { return p.id }

// Status returns the current status.
func (p *Promise) Status() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}

// IsPending returns true if the promise is still pending.
func (p *Promise) IsPending() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status == "pending"
}

// IsResolved returns true if the promise resolved successfully.
func (p *Promise) IsResolved() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status == "resolved"
}

// IsRejected returns true if the promise was rejected.
func (p *Promise) IsRejected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status == "rejected"
}

// Value returns the resolved value (nil if not resolved).
func (p *Promise) Value() any {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.value
}

// Error returns the rejection error (nil if not rejected).
func (p *Promise) Error() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

// Resolve sets the promise to resolved with a value.
func (p *Promise) Resolve(value any) {
	p.mu.Lock()
	if p.status != "pending" {
		p.mu.Unlock()
		return
	}
	p.status = "resolved"
	p.value = value
	callbacks := make([]func(), len(p.onSettle))
	copy(callbacks, p.onSettle)
	p.onSettle = nil
	p.mu.Unlock()

	for _, cb := range callbacks {
		cb()
	}
}

// Reject sets the promise to rejected with an error.
func (p *Promise) Reject(err error) {
	p.mu.Lock()
	if p.status != "pending" {
		p.mu.Unlock()
		return
	}
	p.status = "rejected"
	p.err = err
	callbacks := make([]func(), len(p.onSettle))
	copy(callbacks, p.onSettle)
	p.onSettle = nil
	p.mu.Unlock()

	for _, cb := range callbacks {
		cb()
	}
}

// OnSettle registers a callback to be called when the promise settles.
func (p *Promise) OnSettle(fn func()) {
	p.mu.Lock()
	if p.status != "pending" {
		p.mu.Unlock()
		fn()
		return
	}
	p.onSettle = append(p.onSettle, fn)
	p.mu.Unlock()
}

// SuspenseState tracks the state of a Suspense boundary.
type SuspenseState struct {
	Pending  bool
	Promises []*Promise
	Fallback any // VNode-like fallback
}

// NewSuspenseState creates a new SuspenseState.
func NewSuspenseState() *SuspenseState {
	return &SuspenseState{}
}

// AddPromise adds a promise to the suspense boundary.
func (ss *SuspenseState) AddPromise(p *Promise) {
	ss.Promises = append(ss.Promises, p)
	if p.IsPending() {
		ss.Pending = true
	}
}

// CheckPending re-evaluates whether any promises are still pending.
func (ss *SuspenseState) CheckPending() bool {
	ss.Pending = false
	for _, p := range ss.Promises {
		if p.IsPending() {
			ss.Pending = true
			return true
		}
	}
	return false
}

// LazyComponent represents a component that loads on demand.
type LazyComponent struct {
	mu      sync.Mutex
	loader  func() (any, error)
	loaded  any
	loading bool
	err     error
	status  string // "pending" | "resolved" | "rejected"
}

// NewLazyComponent creates a lazy component with the given loader.
func NewLazyComponent(loader func() (any, error)) *LazyComponent {
	return &LazyComponent{
		loader: loader,
		status: "pending",
	}
}

// Load triggers the lazy load if not already loaded.
func (lc *LazyComponent) Load() (any, error) {
	lc.mu.Lock()
	if lc.status == "resolved" {
		val := lc.loaded
		lc.mu.Unlock()
		return val, nil
	}
	if lc.status == "rejected" {
		err := lc.err
		lc.mu.Unlock()
		return nil, err
	}
	if lc.loading {
		lc.mu.Unlock()
		return nil, fmt.Errorf("lazy component is still loading")
	}
	lc.loading = true
	lc.mu.Unlock()

	val, err := lc.loader()

	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.loading = false
	if err != nil {
		lc.status = "rejected"
		lc.err = err
		return nil, err
	}
	lc.status = "resolved"
	lc.loaded = val
	return val, nil
}

// Status returns the lazy component status.
func (lc *LazyComponent) Status() string {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.status
}

// IsLoaded returns true if the component has been loaded.
func (lc *LazyComponent) IsLoaded() bool {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.status == "resolved"
}

// Component returns the loaded component (nil if not loaded).
func (lc *LazyComponent) Component() any {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.loaded
}
