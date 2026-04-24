package lumina_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina"
)

// TestUseAsyncBasic verifies that useAsync spawns a coroutine that runs
// and completes immediately when no await is used.
func TestUseAsyncBasic(t *testing.T) {
	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	err := app.L.DoString(`
		local lumina = require("lumina")
		_G.result = nil
		lumina.useAsync(function()
			_G.result = 42
		end)
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	app.L.GetGlobal("result")
	val, ok := app.L.ToInteger(-1)
	app.L.Pop(1)
	if !ok || val != 42 {
		t.Errorf("expected result=42, got %v (ok=%v)", val, ok)
	}
}

// TestUseAsyncWithPreResolvedFuture verifies that async.await on a
// pre-resolved future returns immediately without yielding.
func TestUseAsyncWithPreResolvedFuture(t *testing.T) {
	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	err := app.L.DoString(`
		local lumina = require("lumina")
		local async = require("async")
		_G.result = nil
		lumina.useAsync(function()
			local f = async.resolve("hello from async")
			local val = async.await(f)
			_G.result = val
		end)
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	app.L.GetGlobal("result")
	result, ok := app.L.ToString(-1)
	app.L.Pop(1)
	if !ok || result != "hello from async" {
		t.Errorf("expected 'hello from async', got %q (ok=%v)", result, ok)
	}
	if app.Scheduler().Pending() != 0 {
		t.Errorf("expected 0 pending (pre-resolved completes immediately), got %d", app.Scheduler().Pending())
	}
}

// TestUseAsyncWithDelay verifies that lumina.delay creates a Future that
// causes the coroutine to yield, and Scheduler.Tick resumes it after the delay.
func TestUseAsyncWithDelay(t *testing.T) {
	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	err := app.L.DoString(`
		local lumina = require("lumina")
		local async = require("async")
		_G.step = 0
		lumina.useAsync(function()
			_G.step = 1
			async.await(lumina.delay(0.05))
			_G.step = 2
		end)
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	app.L.GetGlobal("step")
	step, _ := app.L.ToInteger(-1)
	app.L.Pop(1)
	if step != 1 {
		t.Fatalf("expected step=1 after spawn, got %d", step)
	}

	sched := app.Scheduler()
	if sched.Pending() != 1 {
		t.Fatalf("expected 1 pending coroutine (waiting on delay), got %d", sched.Pending())
	}

	// Wait for the delay to resolve
	time.Sleep(80 * time.Millisecond)
	sched.Tick()

	app.L.GetGlobal("step")
	step, _ = app.L.ToInteger(-1)
	app.L.Pop(1)
	if step != 2 {
		t.Errorf("expected step=2 after tick, got %d", step)
	}
	if sched.Pending() != 0 {
		t.Errorf("expected 0 pending after completion, got %d", sched.Pending())
	}
}

// TestDelayFuture verifies that lumina.delay returns a Future userdata
// that resolves after the specified time.
func TestDelayFuture(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})
	lumina.Open(L)

	start := time.Now()
	err := L.DoString(`
		local lumina = require("lumina")
		_G.future = lumina.delay(0.05)
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	L.GetGlobal("future")
	ud := L.UserdataValue(-1)
	L.Pop(1)

	future, ok := ud.(*lua.Future)
	if !ok {
		t.Fatalf("expected *lua.Future, got %T", ud)
	}

	<-future.Wait()
	elapsed := time.Since(start)
	if elapsed < 40*time.Millisecond {
		t.Errorf("delay resolved too fast: %v", elapsed)
	}

	val, err2 := future.Result()
	if err2 != nil {
		t.Errorf("expected no error, got %v", err2)
	}
	if val != true {
		t.Errorf("expected true, got %v", val)
	}
}

// TestMultipleAsyncTasks verifies that multiple useAsync calls all complete.
func TestMultipleAsyncTasks(t *testing.T) {
	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	err := app.L.DoString(`
		local lumina = require("lumina")
		local async = require("async")
		_G.count = 0
		for i = 1, 5 do
			lumina.useAsync(function()
				local f = async.resolve(i)
				local val = async.await(f)
				_G.count = _G.count + 1
			end)
		end
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	app.L.GetGlobal("count")
	count, ok := app.L.ToInteger(-1)
	app.L.Pop(1)
	if !ok || count != 5 {
		t.Errorf("expected count=5, got %v", count)
	}
}

// TestAsyncNoBlock verifies that Tick() is non-blocking even when coroutines
// are pending (waiting on unresolved futures).
func TestAsyncNoBlock(t *testing.T) {
	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	err := app.L.DoString(`
		local lumina = require("lumina")
		local async = require("async")
		_G.done = false
		lumina.useAsync(function()
			async.await(lumina.delay(0.2))
			_G.done = true
		end)
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	sched := app.Scheduler()
	if sched.Pending() != 1 {
		t.Fatalf("expected 1 pending, got %d", sched.Pending())
	}

	// Tick should return immediately (non-blocking)
	start := time.Now()
	sched.Tick()
	if time.Since(start) > 50*time.Millisecond {
		t.Errorf("Tick() took too long (should be non-blocking)")
	}

	app.L.GetGlobal("done")
	if app.L.ToBoolean(-1) {
		t.Error("expected done=false while delay is pending")
	}
	app.L.Pop(1)

	// Wait for delay and tick again
	time.Sleep(250 * time.Millisecond)
	sched.Tick()

	app.L.GetGlobal("done")
	if !app.L.ToBoolean(-1) {
		t.Error("expected done=true after delay and tick")
	}
	app.L.Pop(1)
}

// TestAsyncErrorHandling verifies that async.await propagates errors from
// rejected futures.
func TestAsyncErrorHandling(t *testing.T) {
	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	err := app.L.DoString(`
		local lumina = require("lumina")
		local async = require("async")
		_G.got_error = false
		_G.error_msg = nil
		lumina.useAsync(function()
			local f = async.reject("something went wrong")
			local val, err = async.await(f)
			if err then
				_G.got_error = true
				_G.error_msg = err
			end
		end)
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	app.L.GetGlobal("got_error")
	if !app.L.ToBoolean(-1) {
		t.Error("expected got_error=true for rejected future")
	}
	app.L.Pop(1)

	app.L.GetGlobal("error_msg")
	errMsg, _ := app.L.ToString(-1)
	app.L.Pop(1)
	if errMsg != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", errMsg)
	}
}

// TestAsyncHTTPIntegration tests the full async flow using go-lua's built-in
// http module: useAsync → http.get → json.decode → update global.
func TestAsyncHTTPIntegration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"lumina","version":3}`)
	}))
	defer srv.Close()

	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	err := app.L.DoString(fmt.Sprintf(`
		local lumina = require("lumina")
		local http = require("http")
		local json = require("json")
		_G.name = nil
		_G.version = nil
		lumina.useAsync(function()
			local resp = http.get(%q)
			local data = json.decode(resp.body)
			_G.name = data.name
			_G.version = data.version
		end)
	`, srv.URL))
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	// http.get is synchronous within the coroutine, so it completes in Spawn
	app.L.GetGlobal("name")
	name, ok := app.L.ToString(-1)
	app.L.Pop(1)
	if !ok || name != "lumina" {
		t.Errorf("expected name='lumina', got %q", name)
	}

	app.L.GetGlobal("version")
	version, ok := app.L.ToInteger(-1)
	app.L.Pop(1)
	if !ok || version != 3 {
		t.Errorf("expected version=3, got %v", version)
	}
}

// TestUseAsyncWithoutApp verifies that useAsync falls back to a per-State
// scheduler when no App is present.
func TestUseAsyncWithoutApp(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})
	lumina.Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local async = require("async")
		_G.result = nil
		lumina.useAsync(function()
			local f = async.resolve(99)
			local val = async.await(f)
			_G.result = val
		end)
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	L.GetGlobal("result")
	val, ok := L.ToInteger(-1)
	L.Pop(1)
	if !ok || val != 99 {
		t.Errorf("expected result=99, got %v (ok=%v)", val, ok)
	}
}

// TestGetScheduler verifies the GetScheduler helper returns the right scheduler.
func TestGetScheduler(t *testing.T) {
	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	sched := lumina.GetScheduler(app.L)
	if sched == nil {
		t.Fatal("expected non-nil scheduler from App")
	}
	if sched != app.Scheduler() {
		t.Error("expected GetScheduler to return App's scheduler")
	}

	// Without App
	L := lua.NewState()
	defer L.Close()
	lumina.Open(L)

	sched2 := lumina.GetScheduler(L)
	if sched2 == nil {
		t.Fatal("expected non-nil fallback scheduler")
	}
	sched3 := lumina.GetScheduler(L)
	if sched2 != sched3 {
		t.Error("expected same fallback scheduler on repeated calls")
	}
}

// TestAsyncMultipleDelays verifies multiple concurrent delayed coroutines.
func TestAsyncMultipleDelays(t *testing.T) {
	app := lumina.NewAppWithSize(80, 24)
	defer app.Close()
	lumina.SetOutputAdapter(&lumina.NopAdapter{})

	err := app.L.DoString(`
		local lumina = require("lumina")
		local async = require("async")
		_G.results = {}

		for i = 1, 3 do
			lumina.useAsync(function()
				async.await(lumina.delay(0.05))
				table.insert(_G.results, i)
			end)
		end
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	sched := app.Scheduler()
	if sched.Pending() != 3 {
		t.Fatalf("expected 3 pending, got %d", sched.Pending())
	}

	// Wait for all delays
	err = sched.WaitAll(5 * time.Second)
	if err != nil {
		t.Fatalf("WaitAll failed: %v", err)
	}

	// All 3 should have completed
	app.L.GetGlobal("results")
	length := app.L.RawLen(-1)
	app.L.Pop(1)
	if length != 3 {
		t.Errorf("expected 3 results, got %d", length)
	}
}
