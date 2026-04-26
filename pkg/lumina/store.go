// Package lumina — Zustand-like global state management.
package lumina

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/akzj/go-lua/pkg/lua"
)

// Store is a global state container (like Zustand/Redux).
type Store struct {
	ID         string
	state      map[string]any
	actionRefs map[string]int // Lua registry refs for action functions
	listeners  []func()
	nilCount   int // track nil'd listeners for periodic compaction
	mu         sync.RWMutex
}

var storeCounter int64

// NewStore creates a new store with initial state.
func NewStore(initialState map[string]any) *Store {
	id := fmt.Sprintf("store_%d", atomic.AddInt64(&storeCounter, 1))
	if initialState == nil {
		initialState = make(map[string]any)
	}
	return &Store{
		ID:         id,
		state:      initialState,
		actionRefs: make(map[string]int),
	}
}

// GetState returns a shallow copy of the current state.
func (s *Store) GetState() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]any, len(s.state))
	for k, v := range s.state {
		cp[k] = v
	}
	return cp
}

// GetValue returns a single value from the store.
func (s *Store) GetValue(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state[key]
}

// SetState updates a single key and notifies listeners.
func (s *Store) SetState(key string, value any) {
	s.mu.Lock()
	s.state[key] = value
	listeners := make([]func(), len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()
	for _, fn := range listeners {
		if fn != nil {
			fn()
		}
	}
}

// SetBatch updates multiple keys and notifies listeners once.
func (s *Store) SetBatch(updates map[string]any) {
	s.mu.Lock()
	for k, v := range updates {
		s.state[k] = v
	}
	listeners := make([]func(), len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()
	for _, fn := range listeners {
		if fn != nil {
			fn()
		}
	}
}

// Subscribe adds a listener and returns an unsubscribe function.
func (s *Store) Subscribe(listener func()) func() {
	s.mu.Lock()
	s.listeners = append(s.listeners, listener)
	idx := len(s.listeners) - 1
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if idx < len(s.listeners) {
			s.listeners[idx] = nil
			s.nilCount++
			// Compact when more than half are nil
			if s.nilCount > len(s.listeners)/2 && s.nilCount > 4 {
				compacted := make([]func(), 0, len(s.listeners)-s.nilCount)
				for _, fn := range s.listeners {
					if fn != nil {
						compacted = append(compacted, fn)
					}
				}
				s.listeners = compacted
				s.nilCount = 0
			}
		}
	}
}

// -----------------------------------------------------------------------
// Lua API
// -----------------------------------------------------------------------

// luaCreateStore creates a store from Lua.
// Args: { state = {...}, actions = { name = function(state, payload) ... } }
// Returns: store table with getState, setState, subscribe, dispatch methods
func luaCreateStore(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("createStore: expected table argument")
		L.Error()
		return 0
	}

	// Read initial state
	L.GetField(1, "state")
	var initialState map[string]any
	if L.Type(-1) == lua.TypeTable {
		if m, ok := L.ToMap(-1); ok {
			initialState = m
		}
	}
	L.Pop(1)

	store := NewStore(initialState)

	// Read actions and store as registry refs
	L.GetField(1, "actions")
	if L.Type(-1) == lua.TypeTable {
		L.PushNil()
		for L.Next(-2) {
			name, ok := L.ToString(-2)
			if ok && L.Type(-1) == lua.TypeFunction {
				ref := L.Ref(lua.RegistryIndex) // pops value
				store.actionRefs[name] = ref
			} else {
				L.Pop(1) // pop value, keep key
			}
		}
	}
	L.Pop(1) // pop actions table

	// Build the store table returned to Lua
	L.NewTable()

	// Store the Go store as userdata field
	L.PushUserdata(store)
	L.SetField(-2, "_store")

	// store.getState() -> table
	L.PushFunction(storeGetStateFn(store))
	L.SetField(-2, "getState")

	// store.setState(key, value)
	L.PushFunction(storeSetStateFn(store))
	L.SetField(-2, "setState")

	// store.subscribe(callback) -> unsubscribe
	L.PushFunction(storeSubscribeFn(store))
	L.SetField(-2, "subscribe")

	// store.dispatch(actionName, payload?)
	L.PushFunction(storeDispatchFn(store))
	L.SetField(-2, "dispatch")

	return 1
}

func storeGetStateFn(store *Store) lua.Function {
	return func(L *lua.State) int {
		L.PushAny(store.GetState())
		return 1
	}
}

func storeSetStateFn(store *Store) lua.Function {
	return func(L *lua.State) int {
		key, ok := L.ToString(1)
		if !ok {
			L.PushString("store.setState: expected string key")
			L.Error()
			return 0
		}
		val := luaToAny(L, 2)
		store.SetState(key, val)
		return 0
	}
}

func storeSubscribeFn(store *Store) lua.Function {
	return func(L *lua.State) int {
		if L.Type(1) != lua.TypeFunction {
			L.PushString("store.subscribe: expected function")
			L.Error()
			return 0
		}
		L.PushValue(1)
		ref := L.Ref(lua.RegistryIndex)

		// Get the app so we can post events safely from any goroutine.
		// If no app exists (direct Lua usage), call the function synchronously
		// which is safe since we're on the main thread.
		app := GetApp(L)

		unsub := store.Subscribe(func() {
			if app != nil {
				// Post a lua_callback event — safe from any goroutine
				app.PostEvent(AppEvent{
					Type:    "lua_callback",
					Payload: LuaCallbackEvent{RefID: ref},
				})
			} else {
				// No app — direct call (synchronous, main thread only)
				L.RawGetI(lua.RegistryIndex, int64(ref))
				if L.Type(-1) == lua.TypeFunction {
					_ = L.PCall(0, 0, 0)
				} else {
					L.Pop(1)
				}
			}
		})

		L.PushFunction(func(L *lua.State) int {
			unsub()
			L.Unref(lua.RegistryIndex, ref)
			return 0
		})
		return 1
	}
}

func storeDispatchFn(store *Store) lua.Function {
	return func(L *lua.State) int {
		actionName, ok := L.ToString(1)
		if !ok {
			L.PushString("store.dispatch: expected action name string")
			L.Error()
			return 0
		}
		// Built-in "setState" action: merges payload into state directly
		if actionName == "setState" {
			if L.Type(2) == lua.TypeTable {
				if payload, ok := L.ToMap(2); ok {
					store.mu.Lock()
					for k, v := range payload {
						store.state[k] = v
					}
					store.mu.Unlock()
				}
			}
			// Notify listeners
			store.mu.RLock()
			listeners := make([]func(), len(store.listeners))
			copy(listeners, store.listeners)
			store.mu.RUnlock()
			for _, fn := range listeners {
				if fn != nil {
					fn()
				}
			}
			return 0
		}

		store.mu.RLock()
		ref, exists := store.actionRefs[actionName]
		store.mu.RUnlock()
		if !exists {
			L.PushString(fmt.Sprintf("store.dispatch: unknown action %q", actionName))
			L.Error()
			return 0
		}

		// Push action function from registry
		L.RawGetI(lua.RegistryIndex, int64(ref))

		// Push current state as Lua table — action modifies it in-place
		store.mu.RLock()
		L.PushAny(store.state)
		store.mu.RUnlock()

		// Keep a ref to the state table so we can read it back after PCall
		L.PushValue(-1) // duplicate state table
		stateRef := L.Ref(lua.RegistryIndex)

		// Push payload if provided
		nargs := 1
		if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
			L.PushValue(2)
			nargs = 2
		}

		// Call action(state, payload?)
		if status := L.PCall(nargs, 0, 0); status != 0 {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.Unref(lua.RegistryIndex, stateRef)
			L.PushString(fmt.Sprintf("store.dispatch: %s", msg))
			L.Error()
			return 0
		}

		// Read back the state table that the action modified in-place
		L.RawGetI(lua.RegistryIndex, int64(stateRef))
		if m, ok := L.ToMap(-1); ok {
			store.mu.Lock()
			for k, v := range m {
				store.state[k] = v
			}
			store.mu.Unlock()
		}
		L.Pop(1)
		L.Unref(lua.RegistryIndex, stateRef)

		// Notify listeners
		store.mu.RLock()
		listeners := make([]func(), len(store.listeners))
		copy(listeners, store.listeners)
		store.mu.RUnlock()
		for _, fn := range listeners {
			if fn != nil {
				fn()
			}
		}

		return 0
	}
}

// luaUseStore is a hook that subscribes a component to a store.
// Args: store table
// Returns: current state table
func luaUseStore(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useStore: no current component")
		L.Error()
		return 0
	}

	if L.Type(1) != lua.TypeTable {
		L.PushString("useStore: expected store table")
		L.Error()
		return 0
	}

	L.GetField(1, "_store")
	storeUD := L.UserdataValue(-1)
	ok := storeUD != nil
	L.Pop(1)
	if !ok {
		L.PushString("useStore: invalid store")
		L.Error()
		return 0
	}

	store, ok := storeUD.(*Store)
	if !ok {
		L.PushString("useStore: invalid store type")
		L.Error()
		return 0
	}

	// Track subscription via hook index
	comp.mu.Lock()
	idx := comp.generalHookIndex
	comp.generalHookIndex++
	for idx >= len(comp.externalStoreHooks) {
		comp.externalStoreHooks = append(comp.externalStoreHooks, &ExternalStoreHook{})
	}
	hook := comp.externalStoreHooks[idx]
	comp.mu.Unlock()

	// Subscribe on first render
	if !hook.Subscribed {
		hook.Subscribed = true
		store.Subscribe(func() {
			comp.Dirty.Store(true)
		})
	}

	// Return current state snapshot
	L.PushAny(store.GetState())
	return 1
}
