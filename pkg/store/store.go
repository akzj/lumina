// Package store provides a simple global state container with subscription support.
// It is a pure Go package with zero v2 dependencies.
package store

import "sync"

// subscriber holds a callback and a unique ID for stable unsubscribe.
type subscriber struct {
	id  int                         // unique ID (never reused)
	key string                      // empty means subscribe to all keys
	fn  func(key string, value any) // always called with (key, value)
}

// Store is a global state container with subscription support.
type Store struct {
	mu          sync.RWMutex
	state       map[string]any
	subscribers []*subscriber
	nextID      int          // monotonically increasing subscriber ID
	removed     map[int]bool // tracks unsubscribed IDs for lazy cleanup
}

// New creates a new Store with optional initial state.
// If initial is nil, an empty store is created.
func New(initial map[string]any) *Store {
	state := make(map[string]any)
	for k, v := range initial {
		state[k] = v
	}
	return &Store{
		state:   state,
		removed: make(map[int]bool),
	}
}

// Get returns a value by key. The second return value indicates whether the key exists.
func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.state[key]
	return v, ok
}

// Set sets a value and notifies subscribers.
func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	s.state[key] = value
	subs := s.snapshotSubscribers()
	s.mu.Unlock()

	notify(subs, key, value)

	s.mu.Lock()
	s.compactIfNeeded()
	s.mu.Unlock()
}

// Delete removes a key and notifies subscribers with value=nil.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	_, existed := s.state[key]
	if !existed {
		s.mu.Unlock()
		return
	}
	delete(s.state, key)
	subs := s.snapshotSubscribers()
	s.mu.Unlock()

	notify(subs, key, nil)

	s.mu.Lock()
	s.compactIfNeeded()
	s.mu.Unlock()
}

// GetAll returns a shallow copy of all state.
func (s *Store) GetAll() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]any, len(s.state))
	for k, v := range s.state {
		cp[k] = v
	}
	return cp
}

// Subscribe registers a callback for all state changes.
// Returns an unsubscribe function.
func (s *Store) Subscribe(fn func(key string, value any)) func() {
	s.mu.Lock()
	s.nextID++
	id := s.nextID
	s.subscribers = append(s.subscribers, &subscriber{id: id, fn: fn})
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.removed[id] = true
	}
}

// SubscribeKey registers a callback for changes to a specific key.
// The callback receives only the value (not the key). Returns an unsubscribe function.
func (s *Store) SubscribeKey(key string, fn func(value any)) func() {
	return s.Subscribe(func(k string, v any) {
		if k == key {
			fn(v)
		}
	})
}

// Batch applies multiple updates atomically, notifying subscribers once per changed key.
func (s *Store) Batch(updates map[string]any) {
	s.mu.Lock()
	for k, v := range updates {
		s.state[k] = v
	}
	subs := s.snapshotSubscribers()
	s.mu.Unlock()

	// Notify per key, outside the lock.
	for k, v := range updates {
		notify(subs, k, v)
	}

	s.mu.Lock()
	s.compactIfNeeded()
	s.mu.Unlock()
}

// snapshotSubscribers returns a filtered copy of active subscribers.
// Must be called with s.mu held (read or write).
func (s *Store) snapshotSubscribers() []*subscriber {
	result := make([]*subscriber, 0, len(s.subscribers))
	for _, sub := range s.subscribers {
		if !s.removed[sub.id] {
			result = append(result, sub)
		}
	}
	return result
}

// compactIfNeeded removes unsubscribed entries from the subscribers slice
// to prevent unbounded growth in long-running apps. Must be called with s.mu held.
func (s *Store) compactIfNeeded() {
	if len(s.removed) == 0 {
		return
	}
	var compacted []*subscriber
	for _, sub := range s.subscribers {
		if !s.removed[sub.id] {
			compacted = append(compacted, sub)
		}
	}
	s.subscribers = compacted
	s.removed = make(map[int]bool)
}

// notify calls each subscriber with the given key and value.
func notify(subs []*subscriber, key string, value any) {
	for _, sub := range subs {
		sub.fn(key, value)
	}
}
