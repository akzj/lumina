package render

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
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

// --- Lua API: lumina.fetch(url [, options]) ---

// luaFetch implements lumina.fetch(url [, options]) — returns a Future that resolves
// with {status=int, body=string, headers={...}} or rejects on error.
// options: { method="GET", body="", headers={}, timeout=30 }
func (e *Engine) luaFetch(L *lua.State) int {
	url := L.CheckString(1)

	method := "GET"
	var body string
	var headers map[string]string
	timeout := 30 * time.Second

	if L.GetTop() >= 2 && L.IsTable(2) {
		// Extract method
		L.GetField(2, "method")
		if L.IsString(-1) {
			method, _ = L.ToString(-1)
		}
		L.Pop(1)

		// Extract body
		L.GetField(2, "body")
		if L.IsString(-1) {
			body, _ = L.ToString(-1)
		}
		L.Pop(1)

		// Extract timeout
		L.GetField(2, "timeout")
		if L.IsNumber(-1) {
			t, _ := L.ToNumber(-1)
			if t > 0 {
				timeout = time.Duration(t * float64(time.Second))
			}
		}
		L.Pop(1)

		// Extract headers
		L.GetField(2, "headers")
		if L.IsTable(-1) {
			headers = make(map[string]string)
			L.PushNil()
			for L.Next(-2) {
				k, _ := L.ToString(-2)
				v, _ := L.ToString(-1)
				headers[k] = v
				L.Pop(1)
			}
		}
		L.Pop(1)
	}

	future := lua.NewFuture()
	go func() {
		var bodyReader io.Reader
		if body != "" {
			bodyReader = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			future.Reject(err)
			return
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		client := &http.Client{Timeout: timeout}
		resp, err := client.Do(req)
		if err != nil {
			future.Reject(err)
			return
		}
		defer resp.Body.Close()
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
		if err != nil {
			future.Reject(err)
			return
		}

		// Build result as map[string]any for PushAny
		respHeaders := make(map[string]any)
		for k, v := range resp.Header {
			respHeaders[strings.ToLower(k)] = strings.Join(v, ", ")
		}
		result := map[string]any{
			"status":  resp.StatusCode,
			"body":    string(respBody),
			"headers": respHeaders,
		}
		future.Resolve(result)
	}()

	L.PushUserdata(future)
	return 1
}
