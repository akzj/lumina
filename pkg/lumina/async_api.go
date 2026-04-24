// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// useAsync(taskFn) — spawn a Lua function as an async coroutine.
//
// The task function runs as a coroutine managed by the App's Scheduler.
// Inside the coroutine, the function can call async.await(future) to yield
// until a Future resolves.
//
// For HTTP requests, use go-lua's built-in modules:
//
//	local http = require("http")
//	local json = require("json")
//	local async = require("async")
//
// Lua usage:
//
//	lumina.useAsync(function()
//	    local http = require("http")
//	    local resp = http.get("https://example.com/api")
//	    print(resp.body)
//	end)
func useAsync(L *lua.State) int {
	// Arg 1: task function (required)
	if L.Type(1) != lua.TypeFunction {
		L.ArgError(1, "function expected")
		return 0
	}

	// Get the App's scheduler
	app := GetApp(L)
	if app != nil && app.sched != nil {
		// Push the task function for Spawn
		L.PushValue(1)
		if err := app.sched.Spawn(L); err != nil {
			L.PushString(fmt.Sprintf("useAsync: %v", err))
			L.Error()
			return 0
		}
		return 0
	}

	// Fallback for tests without App: use a standalone scheduler stored on
	// the State's user values, or create one.
	sched := getOrCreateScheduler(L)
	L.PushValue(1)
	if err := sched.Spawn(L); err != nil {
		L.PushString(fmt.Sprintf("useAsync: %v", err))
		L.Error()
		return 0
	}
	return 0
}

// getOrCreateScheduler retrieves or creates a per-State Scheduler for
// environments without an App (e.g., unit tests).
func getOrCreateScheduler(L *lua.State) *lua.Scheduler {
	if v := L.UserValue("lumina_scheduler"); v != nil {
		if s, ok := v.(*lua.Scheduler); ok {
			return s
		}
	}
	s := lua.NewScheduler(L)
	L.SetUserValue("lumina_scheduler", s)
	return s
}

// GetScheduler returns the Scheduler for the given Lua State.
// Prefers the App's scheduler; falls back to a per-State scheduler.
func GetScheduler(L *lua.State) *lua.Scheduler {
	if app := GetApp(L); app != nil && app.sched != nil {
		return app.sched
	}
	return getOrCreateScheduler(L)
}

// luminaDelay(seconds) — returns a Future that resolves after a delay.
//
// Useful for testing async flows or implementing debounce/throttle.
//
// Lua usage:
//
//	local async = require("async")
//	lumina.useAsync(function()
//	    async.await(lumina.delay(0.5))  -- yields coroutine for 500ms
//	end)
func luminaDelay(L *lua.State) int {
	seconds, ok := L.ToNumber(1)
	if !ok || seconds < 0 {
		L.ArgError(1, "positive number expected")
		return 0
	}

	future := lua.NewFuture()
	L.PushUserdata(future)

	go func() {
		time.Sleep(time.Duration(seconds * float64(time.Second)))
		future.Resolve(true)
	}()

	return 1
}
