package v2

import (
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/output"
)

// --- timerManager unit tests ---

func TestTimerManager_AddAndFire(t *testing.T) {
	tm := newTimerManager()
	id := tm.add(42, 100, false) // one-shot, 100ms
	if id != 1 {
		t.Fatalf("expected id=1, got %d", id)
	}
	if tm.count() != 1 {
		t.Fatalf("expected 1 timer, got %d", tm.count())
	}

	// Not yet due.
	now := time.Now().UnixMilli()
	refs, oneshot := tm.fireDue(now)
	if len(refs) != 0 {
		t.Fatalf("expected 0 refs, got %d", len(refs))
	}
	if len(oneshot) != 0 {
		t.Fatalf("expected 0 oneshot, got %d", len(oneshot))
	}

	// Advance past fire time.
	refs, oneshot = tm.fireDue(now + 200)
	if len(refs) != 1 || refs[0] != 42 {
		t.Fatalf("expected [42], got %v", refs)
	}
	if len(oneshot) != 1 || oneshot[0] != 42 {
		t.Fatalf("expected oneshot [42], got %v", oneshot)
	}
	// One-shot should be removed.
	if tm.count() != 0 {
		t.Fatalf("expected 0 timers after one-shot, got %d", tm.count())
	}
}

func TestTimerManager_Repeat(t *testing.T) {
	tm := newTimerManager()
	tm.add(99, 50, true) // repeating, 50ms

	base := time.Now().UnixMilli()

	// First fire.
	refs, oneshot := tm.fireDue(base + 60)
	if len(refs) != 1 || refs[0] != 99 {
		t.Fatalf("first fire: expected [99], got %v", refs)
	}
	if len(oneshot) != 0 {
		t.Fatalf("repeat should have no oneshot refs, got %v", oneshot)
	}
	// Still active.
	if tm.count() != 1 {
		t.Fatalf("expected 1 timer after repeat fire, got %d", tm.count())
	}

	// Not yet due for second fire (only 10ms after first fire).
	refs, _ = tm.fireDue(base + 70)
	if len(refs) != 0 {
		t.Fatalf("too early: expected 0 refs, got %d", len(refs))
	}

	// Second fire (50ms after first fire time).
	refs, _ = tm.fireDue(base + 120)
	if len(refs) != 1 || refs[0] != 99 {
		t.Fatalf("second fire: expected [99], got %v", refs)
	}
}

func TestTimerManager_OneShot(t *testing.T) {
	tm := newTimerManager()
	tm.add(10, 30, false) // one-shot, 30ms

	base := time.Now().UnixMilli()

	// Fire.
	refs, oneshot := tm.fireDue(base + 50)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if len(oneshot) != 1 {
		t.Fatalf("expected 1 oneshot, got %d", len(oneshot))
	}

	// Should not fire again.
	refs, _ = tm.fireDue(base + 100)
	if len(refs) != 0 {
		t.Fatalf("one-shot fired again: got %v", refs)
	}
}

func TestTimerManager_Cancel(t *testing.T) {
	tm := newTimerManager()
	id := tm.add(77, 100, true) // repeating

	base := time.Now().UnixMilli()

	// Cancel before fire.
	tm.cancel(id)

	refs, _ := tm.fireDue(base + 200)
	if len(refs) != 0 {
		t.Fatalf("canceled timer fired: got %v", refs)
	}
	// Canceled timer should be cleaned up.
	if tm.count() != 0 {
		t.Fatalf("expected 0 timers after cancel cleanup, got %d", tm.count())
	}
}

func TestTimerManager_ReleaseAll(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tm := newTimerManager()
	// Add a few timers with real Lua refs.
	L.PushFunction(func(L *lua.State) int { return 0 })
	ref1 := L.Ref(lua.RegistryIndex)
	L.PushFunction(func(L *lua.State) int { return 0 })
	ref2 := L.Ref(lua.RegistryIndex)

	tm.add(ref1, 100, true)
	tm.add(ref2, 200, false)

	if tm.count() != 2 {
		t.Fatalf("expected 2 timers, got %d", tm.count())
	}

	tm.releaseAll(L)

	if tm.count() != 0 {
		t.Fatalf("expected 0 timers after releaseAll, got %d", tm.count())
	}
	if tm.nextID != 0 {
		t.Fatalf("expected nextID=0 after releaseAll, got %d", tm.nextID)
	}
}

// --- E2E Lua tests ---

func TestLuaE2E_SetTimeout(t *testing.T) {
	app, _, L := newLuaApp(t, 40, 10)

	// Use a Lua global to track if callback fired.
	err := app.RunString(`
		_timeout_fired = false
		lumina.createComponent({
			id = "timer-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {width = 40, height = 10})
			end
		})
		lumina.setTimeout(function()
			_timeout_fired = true
		end, 10)
	`)
	if err != nil {
		t.Fatalf("RunString: %v", err)
	}
	app.RenderAll()

	// Callback should not have fired yet.
	L.GetGlobal("_timeout_fired")
	if L.ToBoolean(-1) {
		t.Fatal("timeout fired too early")
	}
	L.Pop(1)

	// Simulate time passing and fire timers.
	time.Sleep(20 * time.Millisecond)
	app.fireTimers()

	L.GetGlobal("_timeout_fired")
	if !L.ToBoolean(-1) {
		t.Fatal("timeout did not fire after delay")
	}
	L.Pop(1)

	// Fire again — should NOT fire a second time (one-shot).
	L.DoString(`_timeout_fired = false`)
	time.Sleep(20 * time.Millisecond)
	app.fireTimers()

	L.GetGlobal("_timeout_fired")
	if L.ToBoolean(-1) {
		t.Fatal("one-shot timeout fired again")
	}
	L.Pop(1)
}

func TestLuaE2E_SetInterval(t *testing.T) {
	app, _, L := newLuaApp(t, 40, 10)

	err := app.RunString(`
		_interval_count = 0
		lumina.createComponent({
			id = "interval-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {width = 40, height = 10})
			end
		})
		lumina.setInterval(function()
			_interval_count = _interval_count + 1
		end, 10)
	`)
	if err != nil {
		t.Fatalf("RunString: %v", err)
	}
	app.RenderAll()

	// Fire multiple times.
	for i := 0; i < 3; i++ {
		time.Sleep(15 * time.Millisecond)
		app.fireTimers()
	}

	L.GetGlobal("_interval_count")
	count, _ := L.ToInteger(-1)
	L.Pop(1)
	if count < 2 {
		t.Fatalf("expected interval to fire at least 2 times, got %d", count)
	}
}

func TestLuaE2E_ClearInterval(t *testing.T) {
	app, _, L := newLuaApp(t, 40, 10)

	err := app.RunString(`
		_clear_count = 0
		lumina.createComponent({
			id = "clear-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {width = 40, height = 10})
			end
		})
		_timer_id = lumina.setInterval(function()
			_clear_count = _clear_count + 1
		end, 10)
	`)
	if err != nil {
		t.Fatalf("RunString: %v", err)
	}
	app.RenderAll()

	// Let it fire once.
	time.Sleep(15 * time.Millisecond)
	app.fireTimers()

	L.GetGlobal("_clear_count")
	count1, _ := L.ToInteger(-1)
	L.Pop(1)
	if count1 < 1 {
		t.Fatalf("expected at least 1 fire before clear, got %d", count1)
	}

	// Clear the interval.
	err = L.DoString(`lumina.clearInterval(_timer_id)`)
	if err != nil {
		t.Fatalf("clearInterval: %v", err)
	}

	// Wait and fire again — should not increment.
	time.Sleep(20 * time.Millisecond)
	app.fireTimers()

	L.GetGlobal("_clear_count")
	count2, _ := L.ToInteger(-1)
	L.Pop(1)
	if count2 != count1 {
		t.Fatalf("interval fired after clear: before=%d after=%d", count1, count2)
	}
}

func TestLuaE2E_ClearTimeout(t *testing.T) {
	app, _, L := newLuaApp(t, 40, 10)

	err := app.RunString(`
		_ct_fired = false
		lumina.createComponent({
			id = "ct-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {width = 40, height = 10})
			end
		})
		_ct_id = lumina.setTimeout(function()
			_ct_fired = true
		end, 50)
		lumina.clearTimeout(_ct_id)
	`)
	if err != nil {
		t.Fatalf("RunString: %v", err)
	}
	app.RenderAll()

	// Wait past the timeout and fire.
	time.Sleep(60 * time.Millisecond)
	app.fireTimers()

	L.GetGlobal("_ct_fired")
	if L.ToBoolean(-1) {
		t.Fatal("cleared timeout still fired")
	}
	L.Pop(1)
}

// TestLuaE2E_TimerCallbackError verifies that a timer callback error
// doesn't crash the app — it's silently handled.
func TestLuaE2E_TimerCallbackError(t *testing.T) {
	app, _ := NewTestApp(40, 10)
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app = NewApp(L, 40, 10, ta)

	err := app.RunString(`
		_error_after_count = 0
		lumina.createComponent({
			id = "err-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {width = 40, height = 10})
			end
		})
		lumina.setTimeout(function()
			error("intentional error")
		end, 5)
		lumina.setTimeout(function()
			_error_after_count = 1
		end, 5)
	`)
	if err != nil {
		t.Fatalf("RunString: %v", err)
	}
	app.RenderAll()

	// Both timers should be due. The first errors, the second should still run.
	time.Sleep(10 * time.Millisecond)
	app.fireTimers() // should not panic

	L.GetGlobal("_error_after_count")
	count, _ := L.ToInteger(-1)
	L.Pop(1)
	if count != 1 {
		t.Fatalf("expected second timer to fire despite first error, got count=%d", count)
	}
}
