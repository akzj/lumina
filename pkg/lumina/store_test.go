package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestCreateStore_InitialState(t *testing.T) {
	store := NewStore(map[string]any{"count": int64(0), "name": "test"})
	state := store.GetState()
	if state["count"] != int64(0) {
		t.Fatalf("expected count=0, got %v", state["count"])
	}
	if state["name"] != "test" {
		t.Fatalf("expected name=test, got %v", state["name"])
	}
}

func TestStore_GetState_ReturnsCopy(t *testing.T) {
	store := NewStore(map[string]any{"x": int64(1)})
	s1 := store.GetState()
	s1["x"] = int64(999) // mutate the copy
	s2 := store.GetState()
	if s2["x"] != int64(1) {
		t.Fatalf("GetState should return a copy, but original was mutated: got %v", s2["x"])
	}
}

func TestStore_SetState_UpdatesAndNotifies(t *testing.T) {
	store := NewStore(map[string]any{"count": int64(0)})
	notified := false
	store.Subscribe(func() { notified = true })

	store.SetState("count", int64(42))

	state := store.GetState()
	if state["count"] != int64(42) {
		t.Fatalf("expected count=42, got %v", state["count"])
	}
	if !notified {
		t.Fatal("listener was not notified")
	}
}

func TestStore_Subscribe_Unsubscribe(t *testing.T) {
	store := NewStore(map[string]any{})
	callCount := 0
	unsub := store.Subscribe(func() { callCount++ })

	store.SetState("a", int64(1))
	if callCount != 1 {
		t.Fatalf("expected 1 notification, got %d", callCount)
	}

	unsub()
	store.SetState("a", int64(2))
	if callCount != 1 {
		t.Fatalf("expected still 1 notification after unsub, got %d", callCount)
	}
}

func TestStore_SetBatch(t *testing.T) {
	store := NewStore(map[string]any{"a": int64(0), "b": int64(0)})
	notifyCount := 0
	store.Subscribe(func() { notifyCount++ })

	store.SetBatch(map[string]any{"a": int64(10), "b": int64(20)})

	state := store.GetState()
	if state["a"] != int64(10) || state["b"] != int64(20) {
		t.Fatalf("expected a=10 b=20, got a=%v b=%v", state["a"], state["b"])
	}
	if notifyCount != 1 {
		t.Fatalf("expected 1 batch notification, got %d", notifyCount)
	}
}

func TestStore_MultipleStores_Independent(t *testing.T) {
	s1 := NewStore(map[string]any{"x": int64(1)})
	s2 := NewStore(map[string]any{"x": int64(100)})

	s1.SetState("x", int64(2))
	if s2.GetState()["x"] != int64(100) {
		t.Fatal("stores should be independent")
	}
}

func TestLua_CreateStore(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		_store = lumina.createStore({
			state = { count = 0, name = "hello" },
			actions = {
				increment = function(state)
					state.count = state.count + 1
				end,
			}
		})
		-- getState
		local s = _store.getState()
		_count = s.count
		_name = s.name
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	L.GetGlobal("_count")
	count, ok := L.ToNumber(-1)
	L.Pop(1)
	if !ok || count != 0 {
		t.Fatalf("expected count=0, got %v (ok=%v)", count, ok)
	}

	L.GetGlobal("_name")
	name, ok := L.ToString(-1)
	L.Pop(1)
	if !ok || name != "hello" {
		t.Fatalf("expected name=hello, got %v", name)
	}
}

func TestLua_Store_SetState(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		_store = lumina.createStore({
			state = { count = 0 },
			actions = {}
		})
		_store.setState("count", 42)
		local s = _store.getState()
		_count = s.count
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	L.GetGlobal("_count")
	count, ok := L.ToNumber(-1)
	L.Pop(1)
	if !ok || count != 42 {
		t.Fatalf("expected count=42, got %v", count)
	}
}

func TestLua_Store_Dispatch(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		_store = lumina.createStore({
			state = { count = 0 },
			actions = {
				increment = function(state)
					state.count = state.count + 1
				end,
				add = function(state, n)
					state.count = state.count + n
				end,
			}
		})
		_store.dispatch("increment")
		local s1 = _store.getState()
		_count1 = s1.count

		_store.dispatch("add", 10)
		local s2 = _store.getState()
		_count2 = s2.count
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	L.GetGlobal("_count1")
	c1, ok := L.ToNumber(-1)
	L.Pop(1)
	if !ok || c1 != 1 {
		t.Fatalf("expected count1=1, got %v", c1)
	}

	L.GetGlobal("_count2")
	c2, ok := L.ToNumber(-1)
	L.Pop(1)
	if !ok || c2 != 11 {
		t.Fatalf("expected count2=11, got %v", c2)
	}
}

func TestLua_Store_Subscribe(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		_notify_count = 0
		_store = lumina.createStore({
			state = { x = 0 },
			actions = {}
		})
		local unsub = _store.subscribe(function()
			_notify_count = _notify_count + 1
		end)
		_store.setState("x", 1)
		_store.setState("x", 2)
		_count_before_unsub = _notify_count
		unsub()
		_store.setState("x", 3)
		_count_after_unsub = _notify_count
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	L.GetGlobal("_count_before_unsub")
	before, _ := L.ToNumber(-1)
	L.Pop(1)
	if before != 2 {
		t.Fatalf("expected 2 notifications before unsub, got %v", before)
	}

	L.GetGlobal("_count_after_unsub")
	after, _ := L.ToNumber(-1)
	L.Pop(1)
	if after != 2 {
		t.Fatalf("expected still 2 after unsub, got %v", after)
	}
}

func TestLua_UseStore(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	// Set up a component as current
	comp := &Component{
		ID: "store_test_comp", Type: "StoreTestComp", Name: "StoreTestComp",
		Props: make(map[string]any), State: make(map[string]any),
	}
	globalRegistry.mu.Lock()
	globalRegistry.components[comp.ID] = comp
	globalRegistry.mu.Unlock()
	SetCurrentComponent(comp)
	defer func() {
		SetCurrentComponent(nil)
		globalRegistry.mu.Lock()
		delete(globalRegistry.components, comp.ID)
		globalRegistry.mu.Unlock()
	}()

	err := L.DoString(`
		_store = lumina.createStore({
			state = { count = 5, label = "hello" },
			actions = {}
		})
		local state = lumina.useStore(_store)
		_count = state.count
		_label = state.label
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	L.GetGlobal("_count")
	count, ok := L.ToNumber(-1)
	L.Pop(1)
	if !ok || count != 5 {
		t.Fatalf("expected count=5, got %v", count)
	}

	L.GetGlobal("_label")
	label, ok := L.ToString(-1)
	L.Pop(1)
	if !ok || label != "hello" {
		t.Fatalf("expected label=hello, got %v", label)
	}
}

