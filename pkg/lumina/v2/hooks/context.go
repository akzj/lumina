package hooks

import "sync/atomic"

// ---------- Context System ----------
// Implements React-style context: define a context, provide values,
// and consume them via UseContext.

var contextIDCounter int64

func nextContextID() int64 {
	return atomic.AddInt64(&contextIDCounter, 1)
}

// ContextDef defines a context with a default value.
type ContextDef struct {
	id           int64
	defaultValue any
}

// NewContext creates a new context definition with a default value.
func NewContext(defaultValue any) *ContextDef {
	return &ContextDef{
		id:           nextContextID(),
		defaultValue: defaultValue,
	}
}

// DefaultValue returns the context's default value.
func (c *ContextDef) DefaultValue() any { return c.defaultValue }

// ContextProvider maps context IDs to values for a scope.
// Providers can be nested; lookups walk up the chain.
type ContextProvider struct {
	values map[int64]any
	parent *ContextProvider
}

// NewContextProvider creates a provider with an optional parent.
// Pass nil for a root provider.
func NewContextProvider(parent *ContextProvider) *ContextProvider {
	return &ContextProvider{
		values: make(map[int64]any),
		parent: parent,
	}
}

// Set stores a value for a context definition in this provider.
func (p *ContextProvider) Set(ctx *ContextDef, value any) {
	p.values[ctx.id] = value
}

// Get looks up a context value, walking up the provider chain.
// Returns the value and true if found, or the default value and false.
func (p *ContextProvider) Get(ctx *ContextDef) (any, bool) {
	for cur := p; cur != nil; cur = cur.parent {
		if val, ok := cur.values[ctx.id]; ok {
			return val, true
		}
	}
	return ctx.defaultValue, false
}

// UseContext resolves a context value from the provider chain.
// If provider is nil or the context is not found, returns the default value.
func (h *HookContext) UseContext(ctx *ContextDef, provider *ContextProvider) any {
	// UseContext doesn't consume a hook slot in React, but we track it
	// for hook-count consistency.
	h.callIdx++

	if provider == nil {
		return ctx.defaultValue
	}
	val, _ := provider.Get(ctx)
	return val
}
