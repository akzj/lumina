package lumina

import (
	"fmt"
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestPromiseResolve(t *testing.T) {
	p := NewPromise("test-1")
	if p.Status() != "pending" {
		t.Fatalf("expected pending, got %s", p.Status())
	}
	if !p.IsPending() {
		t.Fatal("expected IsPending")
	}

	settled := false
	p.OnSettle(func() { settled = true })

	p.Resolve("hello")
	if p.Status() != "resolved" {
		t.Fatalf("expected resolved, got %s", p.Status())
	}
	if !p.IsResolved() {
		t.Fatal("expected IsResolved")
	}
	if p.Value() != "hello" {
		t.Fatalf("expected 'hello', got %v", p.Value())
	}
	if !settled {
		t.Fatal("expected OnSettle callback to fire")
	}
}

func TestPromiseReject(t *testing.T) {
	p := NewPromise("test-2")
	p.Reject(fmt.Errorf("failed"))
	if p.Status() != "rejected" {
		t.Fatalf("expected rejected, got %s", p.Status())
	}
	if !p.IsRejected() {
		t.Fatal("expected IsRejected")
	}
	if p.Error() == nil || p.Error().Error() != "failed" {
		t.Fatalf("expected 'failed' error, got %v", p.Error())
	}
}

func TestPromiseDoubleResolve(t *testing.T) {
	p := NewPromise("test-3")
	p.Resolve("first")
	p.Resolve("second") // should be ignored
	if p.Value() != "first" {
		t.Fatalf("expected 'first', got %v", p.Value())
	}
}

func TestPromiseOnSettleAlreadyResolved(t *testing.T) {
	p := NewPromise("test-4")
	p.Resolve("done")

	called := false
	p.OnSettle(func() { called = true })
	if !called {
		t.Fatal("OnSettle should fire immediately for already-resolved promise")
	}
}

func TestSuspenseStatePending(t *testing.T) {
	ss := NewSuspenseState()
	p1 := NewPromise("p1")
	p2 := NewPromise("p2")

	ss.AddPromise(p1)
	ss.AddPromise(p2)
	if !ss.Pending {
		t.Fatal("expected pending with unresolved promises")
	}

	p1.Resolve("ok")
	if !ss.CheckPending() {
		t.Fatal("expected still pending (p2 unresolved)")
	}

	p2.Resolve("ok")
	if ss.CheckPending() {
		t.Fatal("expected not pending (all resolved)")
	}
}

func TestLazyComponentLoad(t *testing.T) {
	callCount := 0
	lc := NewLazyComponent(func() (any, error) {
		callCount++
		return "MyComponent", nil
	})

	if lc.Status() != "pending" {
		t.Fatalf("expected pending, got %s", lc.Status())
	}

	val, err := lc.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "MyComponent" {
		t.Fatalf("expected MyComponent, got %v", val)
	}
	if lc.Status() != "resolved" {
		t.Fatalf("expected resolved, got %s", lc.Status())
	}

	// Second call should return cached
	val2, err := lc.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != "MyComponent" {
		t.Fatalf("expected cached MyComponent, got %v", val2)
	}
	if callCount != 1 {
		t.Fatalf("expected loader called once, got %d", callCount)
	}
}

func TestLazyComponentError(t *testing.T) {
	lc := NewLazyComponent(func() (any, error) {
		return nil, fmt.Errorf("load failed")
	})

	_, err := lc.Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if lc.Status() != "rejected" {
		t.Fatalf("expected rejected, got %s", lc.Status())
	}
}

func TestPromiseConcurrentResolve(t *testing.T) {
	p := NewPromise("concurrent")
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(val int) {
			p.Resolve(val)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	if !p.IsResolved() {
		t.Fatal("expected resolved")
	}
}

func TestLuaSuspenseFallbackPending(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		-- Register parent component
		lumina.defineComponent({
			name = "SuspenseTest",
			render = function(self)
				return lumina.createElement(lumina.Suspense, {
					fallback = { type = "text", content = "Loading..." },
					children = {
						{ type = "box", _lazy_status = "pending" },
					}
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("Lua Suspense fallback: %v", err)
	}
}

func TestLuaLazyResolves(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local Heavy = lumina.lazy(function()
			return lumina.defineComponent({
				name = "HeavyComponent",
				render = function(self)
					return { type = "text", content = "Heavy loaded!" }
				end
			})
		end)
		assert(Heavy ~= nil, "lazy should return a factory")
		assert(Heavy.name == "Lazy", "lazy factory should have name 'Lazy'")
	`)
	if err != nil {
		t.Fatalf("Lua lazy: %v", err)
	}
}
