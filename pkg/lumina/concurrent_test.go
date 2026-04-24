package lumina

import (
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestTransitionState(t *testing.T) {
	ts := NewTransitionState()
	if ts.Pending() {
		t.Fatal("expected not pending initially")
	}

	var insidePending bool
	ts.Start(func() {
		insidePending = ts.Pending()
	})

	if !insidePending {
		t.Fatal("expected pending during callback")
	}
	if ts.Pending() {
		t.Fatal("expected not pending after callback")
	}
}

func TestDeferredValueImmediate(t *testing.T) {
	dv := NewDeferredValue("initial", 100*time.Millisecond)
	if dv.Current() != "initial" {
		t.Fatalf("expected 'initial', got %v", dv.Current())
	}
	if dv.Deferred() != "initial" {
		t.Fatalf("expected deferred 'initial', got %v", dv.Deferred())
	}

	dv.SetImmediate("updated")
	if dv.Current() != "updated" {
		t.Fatalf("expected 'updated', got %v", dv.Current())
	}
	if dv.Deferred() != "updated" {
		t.Fatalf("expected deferred 'updated', got %v", dv.Deferred())
	}
}

func TestDeferredValueDelay(t *testing.T) {
	dv := NewDeferredValue("v1", 20*time.Millisecond)
	dv.Update("v2")

	// Current should update immediately
	if dv.Current() != "v2" {
		t.Fatalf("expected current 'v2', got %v", dv.Current())
	}
	// Deferred should still be old
	if dv.Deferred() != "v1" {
		t.Fatalf("expected deferred 'v1', got %v", dv.Deferred())
	}

	// Wait for timeout
	time.Sleep(50 * time.Millisecond)
	if dv.Deferred() != "v2" {
		t.Fatalf("expected deferred 'v2' after timeout, got %v", dv.Deferred())
	}
}

func TestGenerateID(t *testing.T) {
	ResetIDCounter()
	id1 := GenerateID()
	id2 := GenerateID()
	id3 := GenerateID()

	if id1 != "lumina-1" {
		t.Fatalf("expected lumina-1, got %s", id1)
	}
	if id2 != "lumina-2" {
		t.Fatalf("expected lumina-2, got %s", id2)
	}
	if id3 != "lumina-3" {
		t.Fatalf("expected lumina-3, got %s", id3)
	}
}

func TestGenerateIDUnique(t *testing.T) {
	ResetIDCounter()
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateID()
		if ids[id] {
			t.Fatalf("duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestLuaUseTransition(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	// useTransition requires a component context, test the no-component fallback
	err := L.DoString(`
		local isPending, startTransition = lumina.useTransition()
		assert(type(isPending) == "boolean", "isPending should be boolean")
		assert(type(startTransition) == "function", "startTransition should be function")
		assert(isPending == false, "should not be pending initially")
		-- Without a component, startTransition is a no-op but should not error
		startTransition(function() end)
	`)
	if err != nil {
		t.Fatalf("Lua useTransition: %v", err)
	}
}

func TestLuaUseDeferredValue(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local deferred = lumina.useDeferredValue("hello", { timeoutMs = 300 })
		assert(deferred == "hello", "deferred should return the value, got " .. tostring(deferred))
	`)
	if err != nil {
		t.Fatalf("Lua useDeferredValue: %v", err)
	}
}

func TestLuaUseId(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	ResetIDCounter()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	// useId without component context returns "" — test that it doesn't crash
	// Also test our Go-side GenerateID which is always unique
	err := L.DoString(`
		local id1 = lumina.useId()
		assert(type(id1) == "string", "id should be string")
		-- Without a component context, useId returns empty string
		-- This is expected behavior
	`)
	if err != nil {
		t.Fatalf("Lua useId: %v", err)
	}

	// Verify Go-side GenerateID produces unique IDs
	ResetIDCounter()
	id1 := GenerateID()
	id2 := GenerateID()
	if id1 == id2 {
		t.Fatal("GenerateID should produce unique IDs")
	}
}
