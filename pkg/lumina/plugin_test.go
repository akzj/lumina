package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestPluginRegister(t *testing.T) {
	pr := NewPluginRegistry()
	err := pr.Register(&Plugin{Name: "test-plugin", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if pr.Count() != 1 {
		t.Fatalf("expected 1 plugin, got %d", pr.Count())
	}
}

func TestPluginInitCalled(t *testing.T) {
	pr := NewPluginRegistry()
	initCalled := false
	err := pr.Register(&Plugin{
		Name:    "init-test",
		Version: "1.0.0",
		InitFn: func() error {
			initCalled = true
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := pr.InitAll(); err != nil {
		t.Fatal(err)
	}
	if !initCalled {
		t.Fatal("init function should have been called")
	}
}

func TestPluginGet(t *testing.T) {
	pr := NewPluginRegistry()
	pr.Register(&Plugin{Name: "my-plugin", Version: "2.0.0"})
	p, ok := pr.Get("my-plugin")
	if !ok {
		t.Fatal("should find registered plugin")
	}
	if p.Version != "2.0.0" {
		t.Fatalf("expected version '2.0.0', got '%s'", p.Version)
	}
	_, ok2 := pr.Get("nonexistent")
	if ok2 {
		t.Fatal("should not find nonexistent plugin")
	}
}

func TestPluginMultiple(t *testing.T) {
	pr := NewPluginRegistry()
	pr.Register(&Plugin{Name: "a", Version: "1.0"})
	pr.Register(&Plugin{Name: "b", Version: "1.0"})
	pr.Register(&Plugin{Name: "c", Version: "1.0"})
	if pr.Count() != 3 {
		t.Fatalf("expected 3 plugins, got %d", pr.Count())
	}
	names := pr.List()
	if len(names) != 3 || names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Fatalf("unexpected plugin order: %v", names)
	}
}

func TestPluginVersionConflict(t *testing.T) {
	pr := NewPluginRegistry()
	pr.Register(&Plugin{Name: "conflict", Version: "1.0.0"})
	err := pr.Register(&Plugin{Name: "conflict", Version: "2.0.0"})
	if err == nil {
		t.Fatal("should fail on version conflict")
	}
}

func TestPluginIdempotent(t *testing.T) {
	pr := NewPluginRegistry()
	pr.Register(&Plugin{Name: "idem", Version: "1.0.0"})
	err := pr.Register(&Plugin{Name: "idem", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("same version re-register should be idempotent: %v", err)
	}
	if pr.Count() != 1 {
		t.Fatal("should still be 1 plugin")
	}
}

func TestPluginNameRequired(t *testing.T) {
	pr := NewPluginRegistry()
	err := pr.Register(&Plugin{Version: "1.0.0"})
	if err == nil {
		t.Fatal("should fail with empty name")
	}
}

func TestLuaRegisterPlugin(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalPluginRegistry.Clear()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local initCalled = false

		lumina.registerPlugin({
			name = "test-lua-plugin",
			version = "1.0.0",
			init = function()
				initCalled = true
			end,
			hooks = {
				useCustomHook = function()
					return 42
				end,
			},
		})

		-- Plugin registered but not yet initialized
		assert(initCalled == false, "init should not be called yet")

		-- Use the plugin (triggers init + registers hooks)
		lumina.usePlugin("test-lua-plugin")
		assert(initCalled == true, "init should be called after usePlugin")

		-- Verify hook is now available
		local result = lumina.useCustomHook()
		assert(result == 42, "custom hook should return 42, got " .. tostring(result))
	`)
	if err != nil {
		t.Fatalf("Lua plugin: %v", err)
	}
	globalPluginRegistry.Clear()
}

func TestLuaGetPlugins(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalPluginRegistry.Clear()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.registerPlugin({ name = "alpha", version = "1.0" })
		lumina.registerPlugin({ name = "beta", version = "1.0" })

		local plugins = lumina.getPlugins()
		assert(#plugins == 2, "expected 2 plugins, got " .. #plugins)
		assert(plugins[1] == "alpha", "first plugin should be 'alpha'")
		assert(plugins[2] == "beta", "second plugin should be 'beta'")
	`)
	if err != nil {
		t.Fatalf("Lua getPlugins: %v", err)
	}
	globalPluginRegistry.Clear()
}
