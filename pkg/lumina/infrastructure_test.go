package lumina

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// -----------------------------------------------------------------------
// Hot Reload tests
// -----------------------------------------------------------------------

func TestFileWatcher_DetectsChange(t *testing.T) {
	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lua")
	if err := os.WriteFile(path, []byte("-- v1"), 0644); err != nil {
		t.Fatal(err)
	}

	changed := make(chan string, 1)
	fw := NewFileWatcher([]string{path}, 50*time.Millisecond)
	fw.SetOnChange(func(p string) {
		changed <- p
	})
	fw.Start()
	defer fw.Stop()

	// Wait a bit, then modify the file
	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(path, []byte("-- v2"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case p := <-changed:
		if p != path {
			t.Fatalf("expected path %q, got %q", path, p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for file change notification")
	}
}

func TestSnapshotAllComponents_SavesAllStates(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	// Create test components
	comp1 := &Component{
		ID:    "c1",
		Type:  "Counter",
		State: map[string]any{"count": 5},
	}
	comp2 := &Component{
		ID:    "c2",
		Type:  "Timer",
		State: map[string]any{"elapsed": 42},
	}

	globalRegistry.components["c1"] = comp1
	globalRegistry.components["c2"] = comp2
	defer func() {
		delete(globalRegistry.components, "c1")
		delete(globalRegistry.components, "c2")
	}()

	hr.SnapshotAllComponents()

	snap1 := hr.GetSnapshot("c1")
	if snap1 == nil || snap1.State["count"] != 5 {
		t.Fatal("expected snapshot for c1 with count=5")
	}
	if snap1.ComponentType != "Counter" {
		t.Fatalf("expected ComponentType=Counter, got %q", snap1.ComponentType)
	}

	snap2 := hr.GetSnapshot("c2")
	if snap2 == nil || snap2.State["elapsed"] != 42 {
		t.Fatal("expected snapshot for c2 with elapsed=42")
	}
}

func TestRestoreByType_RestoresState(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	// Snapshot a component
	comp := &Component{
		ID:    "old-id",
		Type:  "Counter",
		State: map[string]any{"count": 10},
	}
	hr.SnapshotState(comp)

	// Create a new component with same Type but different ID (simulating reload)
	newComp := &Component{
		ID:    "new-id",
		Type:  "Counter",
		State: make(map[string]any),
	}

	restored := hr.RestoreByType(newComp)
	if !restored {
		t.Fatal("expected restore to succeed")
	}
	if newComp.State["count"] != 10 {
		t.Fatalf("expected count=10 after restore, got %v", newComp.State["count"])
	}
}

func TestLua_EnableHotReload(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	globalHotReloader.Enable(false)

	err := L.DoString(`
		lumina.enableHotReload({ paths = {"app.lua", "lib.lua"}, interval = 250 })
	`)
	if err != nil {
		t.Fatalf("enableHotReload error: %v", err)
	}

	if !globalHotReloader.IsEnabled() {
		t.Fatal("expected hot reload to be enabled")
	}
	if len(globalHotReloader.config.WatchPaths) != 2 {
		t.Fatalf("expected 2 watch paths, got %d", len(globalHotReloader.config.WatchPaths))
	}
	if globalHotReloader.config.Interval != 250*time.Millisecond {
		t.Fatalf("expected interval=250ms, got %v", globalHotReloader.config.Interval)
	}

	err = L.DoString(`lumina.disableHotReload()`)
	if err != nil {
		t.Fatalf("disableHotReload error: %v", err)
	}
	if globalHotReloader.IsEnabled() {
		t.Fatal("expected hot reload to be disabled")
	}
}

func TestHotReload_PreservesStateAcrossReload(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	// Simulate: component exists with state
	comp := &Component{
		ID:    "comp-1",
		Type:  "MyWidget",
		State: map[string]any{"value": "hello", "count": int64(7)},
	}
	hr.SnapshotState(comp)

	// After reload: new component instance with same Type
	newComp := &Component{
		ID:    "comp-99",
		Type:  "MyWidget",
		State: make(map[string]any),
	}

	hr.RestoreByType(newComp)

	if newComp.State["value"] != "hello" {
		t.Fatalf("expected value=hello, got %v", newComp.State["value"])
	}
	if newComp.State["count"] != int64(7) {
		t.Fatalf("expected count=7, got %v", newComp.State["count"])
	}
}

// -----------------------------------------------------------------------
// Focus Trap tests
// -----------------------------------------------------------------------

func TestFocusScope_LimitsFocusCycling(t *testing.T) {
	ClearFocusScopes()
	eb := NewEventBus()

	// Register global focusables
	eb.RegisterFocusable("input1")
	eb.RegisterFocusable("input2")
	eb.RegisterFocusable("btn-ok")
	eb.RegisterFocusable("btn-cancel")
	eb.RegisterFocusable("input3")

	// Push scope limiting to btn-ok and btn-cancel
	PushFocusScope(&FocusScope{
		ID:           "dialog",
		FocusableIDs: []string{"btn-ok", "btn-cancel"},
	})

	// FocusNext should cycle within scope
	eb.FocusNext()
	if eb.GetFocused() != "btn-ok" {
		t.Fatalf("expected btn-ok, got %q", eb.GetFocused())
	}

	eb.FocusNext()
	if eb.GetFocused() != "btn-cancel" {
		t.Fatalf("expected btn-cancel, got %q", eb.GetFocused())
	}

	// Wrap around
	eb.FocusNext()
	if eb.GetFocused() != "btn-ok" {
		t.Fatalf("expected btn-ok (wrap), got %q", eb.GetFocused())
	}

	ClearFocusScopes()
}

func TestFocusScope_PopRestoresGlobalFocus(t *testing.T) {
	ClearFocusScopes()
	eb := NewEventBus()

	eb.RegisterFocusable("a")
	eb.RegisterFocusable("b")
	eb.RegisterFocusable("c")

	// Focus first
	eb.FocusNext()
	if eb.GetFocused() != "a" {
		t.Fatalf("expected a, got %q", eb.GetFocused())
	}

	// Push scope
	PushFocusScope(&FocusScope{
		FocusableIDs: []string{"b"},
	})

	eb.FocusNext()
	if eb.GetFocused() != "b" {
		t.Fatalf("expected b in scope, got %q", eb.GetFocused())
	}

	// Pop scope — global focus restored
	PopFocusScope()

	eb.FocusNext()
	// Now should cycle through global list, after "b" comes "c"
	if eb.GetFocused() != "c" {
		t.Fatalf("expected c after pop, got %q", eb.GetFocused())
	}

	ClearFocusScopes()
}

func TestFocusScope_NestedInnerTakesPriority(t *testing.T) {
	ClearFocusScopes()
	eb := NewEventBus()

	eb.RegisterFocusable("x")
	eb.RegisterFocusable("y")
	eb.RegisterFocusable("z")

	PushFocusScope(&FocusScope{
		ID:           "outer",
		FocusableIDs: []string{"x", "y"},
	})
	PushFocusScope(&FocusScope{
		ID:           "inner",
		FocusableIDs: []string{"z"},
	})

	// Inner scope should be active
	scope := GetActiveFocusScope()
	if scope == nil || scope.ID != "inner" {
		t.Fatal("expected inner scope to be active")
	}

	eb.FocusNext()
	if eb.GetFocused() != "z" {
		t.Fatalf("expected z (inner scope), got %q", eb.GetFocused())
	}

	// Pop inner → outer takes over
	PopFocusScope()
	scope = GetActiveFocusScope()
	if scope == nil || scope.ID != "outer" {
		t.Fatal("expected outer scope after pop")
	}

	eb.FocusNext()
	// After z, cycling in outer scope should give x
	if f := eb.GetFocused(); f != "x" {
		t.Fatalf("expected x (outer scope), got %q", f)
	}

	ClearFocusScopes()
}

func TestFocusPrev_WithinScope(t *testing.T) {
	ClearFocusScopes()
	eb := NewEventBus()

	eb.RegisterFocusable("a")
	eb.RegisterFocusable("b")
	eb.RegisterFocusable("c")

	PushFocusScope(&FocusScope{
		FocusableIDs: []string{"a", "b", "c"},
	})

	// Start at first
	eb.FocusNext()
	if eb.GetFocused() != "a" {
		t.Fatalf("expected a, got %q", eb.GetFocused())
	}

	// FocusPrev should wrap to last
	eb.FocusPrev()
	if eb.GetFocused() != "c" {
		t.Fatalf("expected c (wrap back), got %q", eb.GetFocused())
	}

	eb.FocusPrev()
	if eb.GetFocused() != "b" {
		t.Fatalf("expected b, got %q", eb.GetFocused())
	}

	ClearFocusScopes()
}

// -----------------------------------------------------------------------
// Router tests
// -----------------------------------------------------------------------

func TestRouter_NavigateChangesPath(t *testing.T) {
	r := NewRouter()
	r.AddRoute("/")
	r.AddRoute("/about")

	r.Navigate("/about")
	if r.GetCurrentPath() != "/about" {
		t.Fatalf("expected /about, got %q", r.GetCurrentPath())
	}
}

func TestRouter_RouteMatchingWithParams(t *testing.T) {
	r := NewRouter()
	r.AddRoute("/users/:id")
	r.AddRoute("/posts/:postId/comments/:commentId")

	route, params := r.Match("/users/42")
	if route == nil {
		t.Fatal("expected route match")
	}
	if params["id"] != "42" {
		t.Fatalf("expected id=42, got %q", params["id"])
	}

	route2, params2 := r.Match("/posts/10/comments/5")
	if route2 == nil {
		t.Fatal("expected route match for nested params")
	}
	if params2["postId"] != "10" || params2["commentId"] != "5" {
		t.Fatalf("expected postId=10, commentId=5, got %v", params2)
	}

	// No match
	route3, _ := r.Match("/nonexistent")
	if route3 != nil {
		t.Fatal("expected no match for /nonexistent")
	}
}

func TestRouter_BackNavigation(t *testing.T) {
	r := NewRouter()
	r.AddRoute("/")
	r.AddRoute("/about")
	r.AddRoute("/contact")

	r.Navigate("/about")
	r.Navigate("/contact")

	if r.GetCurrentPath() != "/contact" {
		t.Fatalf("expected /contact, got %q", r.GetCurrentPath())
	}

	ok := r.Back()
	if !ok {
		t.Fatal("expected Back to succeed")
	}
	if r.GetCurrentPath() != "/about" {
		t.Fatalf("expected /about after back, got %q", r.GetCurrentPath())
	}

	ok = r.Back()
	if !ok {
		t.Fatal("expected Back to succeed again")
	}
	if r.GetCurrentPath() != "/" {
		t.Fatalf("expected / after second back, got %q", r.GetCurrentPath())
	}

	// No more history
	ok = r.Back()
	if ok {
		t.Fatal("expected Back to fail with empty history")
	}
}

func TestRouter_NavigateExtractsParams(t *testing.T) {
	r := NewRouter()
	r.AddRoute("/users/:id")

	r.Navigate("/users/123")
	params := r.GetParams()
	if params["id"] != "123" {
		t.Fatalf("expected id=123, got %q", params["id"])
	}

	// Navigate to non-parameterized route
	r.AddRoute("/about")
	r.Navigate("/about")
	params = r.GetParams()
	if len(params) != 0 {
		t.Fatalf("expected no params for /about, got %v", params)
	}
}

func TestLua_RouterAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalRouter = NewRouter()

	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local router = lumina.createRouter({
			routes = {
				{ path = "/" },
				{ path = "/about" },
				{ path = "/users/:id" },
			}
		})
		_routeCount = router.routeCount

		lumina.navigate("/users/42")
		_path = lumina.getCurrentPath()

		local route = lumina.useRoute()
		_routePath = route.path
		_userId = route.params.id
	`)
	if err != nil {
		t.Fatalf("router API error: %v", err)
	}

	L.GetGlobal("_routeCount")
	rc, _ := L.ToNumber(-1)
	L.Pop(1)
	if rc != 3 {
		t.Fatalf("expected routeCount=3, got %v", rc)
	}

	L.GetGlobal("_path")
	path, _ := L.ToString(-1)
	L.Pop(1)
	if path != "/users/42" {
		t.Fatalf("expected /users/42, got %q", path)
	}

	L.GetGlobal("_routePath")
	rp, _ := L.ToString(-1)
	L.Pop(1)
	if rp != "/users/42" {
		t.Fatalf("expected route.path=/users/42, got %q", rp)
	}

	L.GetGlobal("_userId")
	uid, _ := L.ToString(-1)
	L.Pop(1)
	if uid != "42" {
		t.Fatalf("expected userId=42, got %q", uid)
	}

	// Test back
	err = L.DoString(`
		lumina.navigate("/about")
		_ok = lumina.back()
		_afterBack = lumina.getCurrentPath()
	`)
	if err != nil {
		t.Fatalf("back error: %v", err)
	}

	L.GetGlobal("_ok")
	if !L.ToBoolean(-1) {
		t.Fatal("expected back() to return true")
	}
	L.Pop(1)

	L.GetGlobal("_afterBack")
	ab, _ := L.ToString(-1)
	L.Pop(1)
	if ab != "/users/42" {
		t.Fatalf("expected /users/42 after back, got %q", ab)
	}

	globalRouter = NewRouter()
}

func TestRouter_OnChangeCallback(t *testing.T) {
	r := NewRouter()
	r.AddRoute("/")
	r.AddRoute("/about")

	var notified []string
	r.OnChange(func(path string) {
		notified = append(notified, path)
	})

	r.Navigate("/about")
	r.Navigate("/")

	if len(notified) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(notified))
	}
	if notified[0] != "/about" || notified[1] != "/" {
		t.Fatalf("expected [/about, /], got %v", notified)
	}
}

func TestRouter_History(t *testing.T) {
	r := NewRouter()
	r.Navigate("/a")
	r.Navigate("/b")
	r.Navigate("/c")

	hist := r.GetHistory()
	// History should contain previous paths: /, /a, /b (current is /c)
	if len(hist) != 3 {
		t.Fatalf("expected 3 history entries, got %d: %v", len(hist), hist)
	}
}

func TestSplitPath(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"/", nil},
		{"/about", []string{"about"}},
		{"/users/42", []string{"users", "42"}},
		{"/a/b/c", []string{"a", "b", "c"}},
		{"", nil},
	}
	for _, c := range cases {
		got := splitPath(c.input)
		if len(got) != len(c.want) {
			t.Errorf("splitPath(%q) = %v, want %v", c.input, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitPath(%q)[%d] = %q, want %q", c.input, i, got[i], c.want[i])
			}
		}
	}
}

func TestMatchRoute(t *testing.T) {
	pattern := []string{"users", ":id"}

	params, ok := matchRoute(pattern, []string{"users", "42"})
	if !ok {
		t.Fatal("expected match")
	}
	if params["id"] != "42" {
		t.Fatalf("expected id=42, got %q", params["id"])
	}

	_, ok = matchRoute(pattern, []string{"posts", "42"})
	if ok {
		t.Fatal("expected no match for posts/42")
	}

	_, ok = matchRoute(pattern, []string{"users"})
	if ok {
		t.Fatal("expected no match for users (too few segments)")
	}
}
