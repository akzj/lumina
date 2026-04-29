package v2

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAsync_SpawnAndSleep(t *testing.T) {
	app, _, _ := newEngineApp(t, 40, 10)
	err := app.RunString(`
		lumina.app {
			id = "async-test",
			store = { done = false },
			render = function()
				return lumina.createElement("text", {}, "hi")
			end,
		}
		lumina.spawn(function()
			local async = require("async")
			local f = lumina.sleep(10)
			async.await(f)
			lumina.store.set("done", true)
		end)
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	// Tick scheduler until done
	deadline := time.Now().Add(2 * time.Second)
	for {
		app.Engine().TickScheduler()
		v, _ := app.Store().Get("done")
		if v == true {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for spawn+sleep")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestAsync_Exec(t *testing.T) {
	app, _, _ := newEngineApp(t, 40, 10)
	err := app.RunString(`
		lumina.app {
			id = "exec-test",
			store = { result = "" },
			render = function()
				return lumina.createElement("text", {}, "hi")
			end,
		}
		lumina.spawn(function()
			local async = require("async")
			local f = lumina.exec("echo hello")
			local r = async.await(f)
			lumina.store.set("result", r.output)
		end)
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	deadline := time.Now().Add(5 * time.Second)
	for {
		app.Engine().TickScheduler()
		v, _ := app.Store().Get("result")
		if s, ok := v.(string); ok && s != "" {
			if s != "hello\n" {
				t.Fatalf("expected 'hello\\n', got %q", s)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestAsync_ReadFile(t *testing.T) {
	tmpFile := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tmpFile, []byte("lumina"), 0644); err != nil {
		t.Fatal(err)
	}

	app, _, _ := newEngineApp(t, 40, 10)
	err := app.RunString(`
		lumina.app {
			id = "readfile-test",
			store = { content = "" },
			render = function()
				return lumina.createElement("text", {}, "hi")
			end,
		}
		lumina.spawn(function()
			local async = require("async")
			local f = lumina.readFile("` + tmpFile + `")
			local data = async.await(f)
			lumina.store.set("content", data)
		end)
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	deadline := time.Now().Add(2 * time.Second)
	for {
		app.Engine().TickScheduler()
		v, _ := app.Store().Get("content")
		if s, ok := v.(string); ok && s == "lumina" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestAsync_Fetch(t *testing.T) {
	// Create a test HTTP server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"items":[1,2,3]}`))
	}))
	defer srv.Close()

	app, _, _ := newEngineApp(t, 40, 10)
	err := app.RunString(`
		lumina.app {
			id = "fetch-test",
			store = { body = "", status = 0 },
			render = function()
				return lumina.createElement("text", {}, "hi")
			end,
		}
		lumina.spawn(function()
			local async = require("async")
			local f = lumina.fetch("` + srv.URL + `")
			local resp = async.await(f)
			lumina.store.set("body", resp.body)
			lumina.store.set("status", resp.status)
		end)
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	deadline := time.Now().Add(5 * time.Second)
	for {
		app.Engine().TickScheduler()
		v, _ := app.Store().Get("body")
		if s, ok := v.(string); ok && s != "" {
			if s != `{"items":[1,2,3]}` {
				t.Fatalf("unexpected body: %q", s)
			}
			status, _ := app.Store().Get("status")
			switch st := status.(type) {
			case int:
				if st != 200 {
					t.Fatalf("unexpected status: %d", st)
				}
			case int64:
				if st != 200 {
					t.Fatalf("unexpected status: %d", st)
				}
			case float64:
				if st != 200 {
					t.Fatalf("unexpected status: %f", st)
				}
			default:
				t.Fatalf("unexpected status type: %T value %v", status, status)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestAsync_FetchPost(t *testing.T) {
	// Test POST with body and custom headers
	var receivedBody string
	var receivedHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		receivedHeader = r.Header.Get("X-Custom")
		w.WriteHeader(201)
		w.Write([]byte("created"))
	}))
	defer srv.Close()

	app, _, _ := newEngineApp(t, 40, 10)
	err := app.RunString(`
		lumina.app {
			id = "fetch-post-test",
			store = { status = 0 },
			render = function()
				return lumina.createElement("text", {}, "hi")
			end,
		}
		lumina.spawn(function()
			local async = require("async")
			local f = lumina.fetch("` + srv.URL + `", {
				method = "POST",
				body = "hello world",
				headers = { ["X-Custom"] = "test-value" },
			})
			local resp = async.await(f)
			lumina.store.set("status", resp.status)
		end)
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	deadline := time.Now().Add(5 * time.Second)
	for {
		app.Engine().TickScheduler()
		v, _ := app.Store().Get("status")
		switch st := v.(type) {
		case int64:
			if st == 201 {
				goto done
			}
		case int:
			if st == 201 {
				goto done
			}
		case float64:
			if st == 201 {
				goto done
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
done:
	if receivedBody != "hello world" {
		t.Fatalf("expected body 'hello world', got %q", receivedBody)
	}
	if receivedHeader != "test-value" {
		t.Fatalf("expected header 'test-value', got %q", receivedHeader)
	}
}

func TestAsync_Cancel(t *testing.T) {
	app, _, _ := newEngineApp(t, 40, 10)
	err := app.RunString(`
		lumina.app {
			id = "cancel-test",
			store = { ran = false },
			render = function()
				return lumina.createElement("text", {}, "hi")
			end,
		}
		local h = lumina.spawn(function()
			local async = require("async")
			async.await(lumina.sleep(500))
			lumina.store.set("ran", true)
		end)
		lumina.cancel(h)
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	// Tick for a bit to confirm it never runs
	for i := 0; i < 100; i++ {
		app.Engine().TickScheduler()
		time.Sleep(5 * time.Millisecond)
	}

	v, _ := app.Store().Get("ran")
	if v == true {
		t.Fatal("cancelled coroutine should not have run")
	}
}
