package lumina

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestQueryCacheSetGet(t *testing.T) {
	qc := &QueryCache{entries: make(map[string]*QueryEntry)}

	qc.Set("users", "data123", nil, 60*time.Second)
	entry := qc.Get("users")
	if entry == nil {
		t.Fatal("expected entry")
	}
	if entry.Data != "data123" {
		t.Fatalf("expected data123, got %v", entry.Data)
	}
	if entry.IsStale() {
		t.Fatal("entry should not be stale yet")
	}
}

func TestQueryCacheStale(t *testing.T) {
	qc := &QueryCache{entries: make(map[string]*QueryEntry)}

	qc.Set("key", "val", nil, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	entry := qc.Get("key")
	if entry == nil {
		t.Fatal("expected entry")
	}
	if !entry.IsStale() {
		t.Fatal("entry should be stale")
	}
}

func TestQueryCacheInvalidate(t *testing.T) {
	qc := &QueryCache{entries: make(map[string]*QueryEntry)}

	qc.Set("a", "1", nil, time.Hour)
	qc.Set("b", "2", nil, time.Hour)
	if qc.Size() != 2 {
		t.Fatalf("expected 2 entries, got %d", qc.Size())
	}

	qc.Invalidate("a")
	if qc.Size() != 1 {
		t.Fatalf("expected 1 entry, got %d", qc.Size())
	}
	if qc.Get("a") != nil {
		t.Fatal("expected nil after invalidate")
	}

	qc.InvalidateAll()
	if qc.Size() != 0 {
		t.Fatalf("expected 0 entries, got %d", qc.Size())
	}
}

func TestQueryCacheError(t *testing.T) {
	qc := &QueryCache{entries: make(map[string]*QueryEntry)}

	qc.Set("err-key", nil, fmt.Errorf("network error"), time.Hour)
	entry := qc.Get("err-key")
	if entry == nil {
		t.Fatal("expected entry")
	}
	if entry.Error == nil {
		t.Fatal("expected error")
	}
	if entry.Error.Error() != "network error" {
		t.Fatalf("expected 'network error', got '%s'", entry.Error.Error())
	}
}

func TestFetchHTTP(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	data, err := Fetch(server.URL)
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}
	if data != `{"status":"ok"}` {
		t.Fatalf("expected JSON, got '%s'", data)
	}
}

func TestFetchHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	_, err := Fetch(server.URL)
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestLuaFetchAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello from server"))
	}))
	defer server.Close()

	ClearComponents()
	ClearContextValues()
	globalQueryCache.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(fmt.Sprintf(`
		local data, err = lumina.fetch("%s")
		assert(data == "hello from server", "expected 'hello from server', got '" .. tostring(data) .. "'")
		assert(err == nil, "expected no error")
	`, server.URL))
	if err != nil {
		t.Fatalf("Lua fetch: %v", err)
	}
}

func TestLuaUseQueryAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalQueryCache.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local result = lumina.useQuery("test-key", function()
			return { items = {"a", "b", "c"} }
		end, { staleTime = 60 })
		assert(result ~= nil, "expected result")
		assert(result.loading == false, "expected not loading")
		assert(result.error == nil, "expected no error")
		assert(result.data ~= nil, "expected data")
	`)
	if err != nil {
		t.Fatalf("Lua useQuery: %v", err)
	}
}

func TestLuaUseQueryCaching(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalQueryCache.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local callCount = 0
		local function fetcher()
			callCount = callCount + 1
			return "data-" .. callCount
		end

		-- First call: fetcher runs
		local r1 = lumina.useQuery("cached-key", fetcher, { staleTime = 60 })
		assert(r1.data == "data-1", "first call should return data-1, got " .. tostring(r1.data))

		-- Second call: should return cached result
		local r2 = lumina.useQuery("cached-key", fetcher, { staleTime = 60 })
		assert(r2.data == "data-1", "second call should return cached data-1, got " .. tostring(r2.data))

		-- Invalidate and re-fetch
		lumina.invalidateQuery("cached-key")
		local r3 = lumina.useQuery("cached-key", fetcher, { staleTime = 60 })
		assert(r3.data == "data-2", "after invalidate should return data-2, got " .. tostring(r3.data))
	`)
	if err != nil {
		t.Fatalf("Lua useQuery caching: %v", err)
	}
}

func TestLuaInvalidateAllQueries(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalQueryCache.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.useQuery("key1", function() return "a" end)
		lumina.useQuery("key2", function() return "b" end)
		lumina.invalidateAllQueries()
	`)
	if err != nil {
		t.Fatalf("invalidateAllQueries: %v", err)
	}
	if globalQueryCache.Size() != 0 {
		t.Fatalf("expected 0 entries after invalidateAll, got %d", globalQueryCache.Size())
	}
}
