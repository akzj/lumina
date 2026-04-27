package store

import "sync"

// Registry manages named stores.
type Registry struct {
	mu     sync.RWMutex
	stores map[string]*Store
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		stores: make(map[string]*Store),
	}
}

// GetOrCreate returns an existing store by name, or creates a new one
// with the given initial state.
func (r *Registry) GetOrCreate(name string, initial map[string]any) *Store {
	// Fast path: read lock.
	r.mu.RLock()
	if s, ok := r.stores[name]; ok {
		r.mu.RUnlock()
		return s
	}
	r.mu.RUnlock()

	// Slow path: write lock, double-check.
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.stores[name]; ok {
		return s
	}
	s := New(initial)
	r.stores[name] = s
	return s
}

// Get returns a store by name, or nil if it doesn't exist.
func (r *Registry) Get(name string) *Store {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stores[name]
}

// Delete removes a store from the registry.
func (r *Registry) Delete(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.stores, name)
}

// Names returns all store names in the registry.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.stores))
	for name := range r.stores {
		names = append(names, name)
	}
	return names
}
