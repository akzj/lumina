package store

import (
	"sort"
	"sync"
	"testing"
)

// --- Store Tests ---

func TestNew_WithInitialState(t *testing.T) {
	s := New(map[string]any{"a": 1, "b": "hello"})
	v, ok := s.Get("a")
	if !ok || v != 1 {
		t.Fatalf("expected (1, true), got (%v, %v)", v, ok)
	}
	v, ok = s.Get("b")
	if !ok || v != "hello" {
		t.Fatalf("expected (hello, true), got (%v, %v)", v, ok)
	}
}

func TestNew_NilInitial(t *testing.T) {
	s := New(nil)
	v, ok := s.Get("anything")
	if ok || v != nil {
		t.Fatalf("expected (nil, false), got (%v, %v)", v, ok)
	}
}

func TestGetSet(t *testing.T) {
	s := New(nil)

	// Missing key
	v, ok := s.Get("x")
	if ok || v != nil {
		t.Fatalf("expected (nil, false), got (%v, %v)", v, ok)
	}

	// Set and get
	s.Set("x", 42)
	v, ok = s.Get("x")
	if !ok || v != 42 {
		t.Fatalf("expected (42, true), got (%v, %v)", v, ok)
	}

	// Overwrite
	s.Set("x", "new")
	v, ok = s.Get("x")
	if !ok || v != "new" {
		t.Fatalf("expected (new, true), got (%v, %v)", v, ok)
	}
}

func TestDelete(t *testing.T) {
	s := New(map[string]any{"k": 10})
	s.Delete("k")
	v, ok := s.Get("k")
	if ok || v != nil {
		t.Fatalf("expected (nil, false) after delete, got (%v, %v)", v, ok)
	}
}

func TestDelete_NonExistent(t *testing.T) {
	s := New(nil)
	// Should not panic or notify
	s.Delete("nope")
}

func TestGetAll_ReturnsCopy(t *testing.T) {
	s := New(map[string]any{"a": 1, "b": 2})
	all := s.GetAll()

	// Mutating the copy should not affect the store.
	all["a"] = 999
	all["c"] = 3

	v, _ := s.Get("a")
	if v != 1 {
		t.Fatalf("GetAll copy mutation affected store: a=%v", v)
	}
	_, ok := s.Get("c")
	if ok {
		t.Fatal("GetAll copy mutation added key to store")
	}
}

func TestSubscribe_CalledOnSet(t *testing.T) {
	s := New(nil)
	var calls []string
	s.Subscribe(func(key string, value any) {
		calls = append(calls, key)
	})

	s.Set("x", 1)
	s.Set("y", 2)

	if len(calls) != 2 || calls[0] != "x" || calls[1] != "y" {
		t.Fatalf("expected [x, y], got %v", calls)
	}
}

func TestSubscribe_CalledOnDelete(t *testing.T) {
	s := New(map[string]any{"k": 1})
	var gotKey string
	var gotValue any
	s.Subscribe(func(key string, value any) {
		gotKey = key
		gotValue = value
	})

	s.Delete("k")

	if gotKey != "k" {
		t.Fatalf("expected key=k, got %q", gotKey)
	}
	if gotValue != nil {
		t.Fatalf("expected value=nil on delete, got %v", gotValue)
	}
}

func TestSubscribeKey_OnlyMatchingKey(t *testing.T) {
	s := New(nil)
	var values []any
	s.SubscribeKey("target", func(value any) {
		values = append(values, value)
	})

	s.Set("other", 1)
	s.Set("target", 2)
	s.Set("other2", 3)
	s.Set("target", 4)

	if len(values) != 2 || values[0] != 2 || values[1] != 4 {
		t.Fatalf("expected [2, 4], got %v", values)
	}
}

func TestUnsubscribe(t *testing.T) {
	s := New(nil)
	callCount := 0
	unsub := s.Subscribe(func(key string, value any) {
		callCount++
	})

	s.Set("a", 1)
	if callCount != 1 {
		t.Fatalf("expected 1 call before unsub, got %d", callCount)
	}

	unsub()
	s.Set("b", 2)
	if callCount != 1 {
		t.Fatalf("expected 1 call after unsub, got %d", callCount)
	}
}

func TestBatch(t *testing.T) {
	s := New(nil)
	var keys []string
	s.Subscribe(func(key string, value any) {
		keys = append(keys, key)
	})

	s.Batch(map[string]any{"x": 1, "y": 2, "z": 3})

	// All keys should be in state.
	for _, k := range []string{"x", "y", "z"} {
		if _, ok := s.Get(k); !ok {
			t.Fatalf("key %q missing after Batch", k)
		}
	}

	// Subscriber should have been called for each key (order may vary due to map iteration).
	sort.Strings(keys)
	if len(keys) != 3 || keys[0] != "x" || keys[1] != "y" || keys[2] != "z" {
		t.Fatalf("expected [x, y, z], got %v", keys)
	}
}

func TestConcurrency(t *testing.T) {
	s := New(nil)
	const goroutines = 50
	const ops = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range ops {
				key := "key"
				s.Set(key, i)
				s.Get(key)
				s.GetAll()
				_ = id
			}
		}(g)
	}

	wg.Wait()

	// If we get here without -race detecting anything, the test passes.
	_, ok := s.Get("key")
	if !ok {
		t.Fatal("key should exist after concurrent writes")
	}
}

func TestConcurrency_SubscribeUnsubscribe(t *testing.T) {
	s := New(nil)
	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			unsub := s.Subscribe(func(key string, value any) {})
			s.Set("k", 1)
			unsub()
		}()
	}

	wg.Wait()
}

// --- Registry Tests ---

func TestRegistry_GetOrCreate(t *testing.T) {
	r := NewRegistry()

	s1 := r.GetOrCreate("app", map[string]any{"theme": "dark"})
	if s1 == nil {
		t.Fatal("expected non-nil store")
	}

	v, ok := s1.Get("theme")
	if !ok || v != "dark" {
		t.Fatalf("expected (dark, true), got (%v, %v)", v, ok)
	}

	// Second call returns same store.
	s2 := r.GetOrCreate("app", map[string]any{"theme": "light"})
	if s1 != s2 {
		t.Fatal("expected same store instance on second GetOrCreate")
	}

	// Initial state should NOT be overwritten.
	v, _ = s2.Get("theme")
	if v != "dark" {
		t.Fatalf("expected theme=dark (original), got %v", v)
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	if s := r.Get("missing"); s != nil {
		t.Fatal("expected nil for missing store")
	}

	r.GetOrCreate("x", nil)
	if s := r.Get("x"); s == nil {
		t.Fatal("expected non-nil for existing store")
	}
}

func TestRegistry_Delete(t *testing.T) {
	r := NewRegistry()
	r.GetOrCreate("temp", nil)
	r.Delete("temp")

	if s := r.Get("temp"); s != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	r.GetOrCreate("alpha", nil)
	r.GetOrCreate("beta", nil)
	r.GetOrCreate("gamma", nil)

	names := r.Names()
	sort.Strings(names)
	if len(names) != 3 || names[0] != "alpha" || names[1] != "beta" || names[2] != "gamma" {
		t.Fatalf("expected [alpha, beta, gamma], got %v", names)
	}
}
