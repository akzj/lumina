package render

import (
	"os"
	"os/exec"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// SetScheduler sets the async coroutine scheduler on the engine.
// Must be called before RegisterLuaAPI.
func (e *Engine) SetScheduler(sched *lua.Scheduler) {
	e.scheduler = sched
}

// Scheduler returns the engine's async scheduler (may be nil).
func (e *Engine) Scheduler() *lua.Scheduler {
	return e.scheduler
}

// TickScheduler ticks the async scheduler, resuming completed coroutines.
// Safe to call even if scheduler is nil.
func (e *Engine) TickScheduler() {
	if e.scheduler != nil {
		e.scheduler.Tick()
	}
}

// preloadAsyncModule ensures the "async" module is loaded into package.loaded
// so that coroutines (which share the same global table) can use require("async")
// without needing the global searcher (which isn't propagated to NewThread).
func (e *Engine) preloadAsyncModule() {
	L := e.L
	// Call require("async") in the main state — this loads it into package.loaded
	L.GetGlobal("require")
	if L.IsFunction(-1) {
		L.PushString("async")
		if status := L.PCall(1, 1, 0); status == lua.OK {
			L.Pop(1) // pop the module table; it's now in package.loaded
		} else {
			L.Pop(1) // pop error
		}
	} else {
		L.Pop(1)
	}
}

// --- Lua API: lumina.spawn(fn) ---

// luaSpawn implements lumina.spawn(fn) — starts an async coroutine.
// fn is a Lua function that may call async.await(future) to yield.
// Returns a coroutine handle (userdata) for optional cancellation.
func (e *Engine) luaSpawn(L *lua.State) int {
	if e.scheduler == nil {
		L.PushString("spawn: no scheduler configured")
		L.Error()
		return 0
	}

	L.CheckType(1, lua.TypeFunction)
	L.PushValue(1) // push function for Spawn (it pops it)

	handle, err := e.scheduler.Spawn(L)
	if err != nil {
		L.PushString("spawn error: " + err.Error())
		L.Error()
		return 0
	}

	// Push handle as userdata so Lua can cancel it later
	L.PushUserdata(handle)
	return 1
}

// --- Lua API: lumina.cancel(handle) ---

// luaCancel implements lumina.cancel(handle) — cancels a spawned coroutine.
func (e *Engine) luaCancel(L *lua.State) int {
	if e.scheduler == nil {
		return 0
	}

	ud := L.UserdataValue(1)
	if ud == nil {
		L.ArgError(1, "CoroutineHandle expected, got nil")
		return 0
	}
	handle, ok := ud.(*lua.CoroutineHandle)
	if !ok {
		L.ArgError(1, "CoroutineHandle expected")
		return 0
	}

	e.scheduler.Cancel(handle)
	return 0
}

// --- Lua API: lumina.sleep(ms) ---

// luaSleep implements lumina.sleep(ms) — returns a Future that resolves after ms milliseconds.
// Use with async.await inside a spawned coroutine.
func (e *Engine) luaSleep(L *lua.State) int {
	ms := L.CheckInteger(1)
	future := lua.NewFuture()
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		future.Resolve(true)
	}()
	L.PushUserdata(future)
	return 1
}

// --- Lua API: lumina.exec(cmd) ---

// luaExec implements lumina.exec(cmd) — returns a Future that resolves with
// {output=string, error=string|nil} after running the shell command.
func (e *Engine) luaExec(L *lua.State) int {
	cmd := L.CheckString(1)
	future := lua.NewFuture()
	go func() {
		out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		result := map[string]any{
			"output": string(out),
		}
		if err != nil {
			result["error"] = err.Error()
		}
		future.Resolve(result)
	}()
	L.PushUserdata(future)
	return 1
}

// --- Lua API: lumina.readFile(path) ---

// luaReadFile implements lumina.readFile(path) — returns a Future that resolves
// with the file content as a string, or rejects on error.
func (e *Engine) luaReadFile(L *lua.State) int {
	path := L.CheckString(1)
	future := lua.NewFuture()
	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			future.Reject(err)
		} else {
			future.Resolve(string(data))
		}
	}()
	L.PushUserdata(future)
	return 1
}
