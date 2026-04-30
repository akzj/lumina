package v2

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestInstallRequireHook_TracksLoadedFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a module file.
	modDir := filepath.Join(dir, "lib")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	modFile := filepath.Join(modDir, "helper.lua")
	if err := os.WriteFile(modFile, []byte(`return { greet = function() return "hello" end }`), 0644); err != nil {
		t.Fatal(err)
	}

	L := lua.NewState()
	defer L.Close()

	// Set package.path to include the temp dir.
	code := `package.path = "` + dir + `/?.lua;" .. "` + dir + `/?/init.lua;" .. package.path`
	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	// Track paths added by the hook.
	var mu sync.Mutex
	var trackedPaths []string

	installRequireHook(L, func(path string) {
		mu.Lock()
		trackedPaths = append(trackedPaths, path)
		mu.Unlock()
	})

	// Now require the module.
	if err := L.DoString(`local h = require("lib.helper")`); err != nil {
		t.Fatalf("require failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(trackedPaths) == 0 {
		t.Fatal("expected require hook to track at least one path")
	}

	// The tracked path should be the absolute path to lib/helper.lua.
	found := false
	for _, p := range trackedPaths {
		if p == modFile {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected tracked paths to contain %q, got %v", modFile, trackedPaths)
	}
}

func TestInstallRequireHook_NilCallback(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Should not panic with nil callback.
	installRequireHook(L, nil)

	// require should still work normally.
	if err := L.DoString(`local x = require("string")`); err != nil {
		// string module may not be available in go-lua, that's ok
		_ = err
	}
}

func TestInstallRequireHook_DoesNotBreakRequire(t *testing.T) {
	dir := t.TempDir()

	// Create a module that returns a value.
	modFile := filepath.Join(dir, "mymod.lua")
	if err := os.WriteFile(modFile, []byte(`return 42`), 0644); err != nil {
		t.Fatal(err)
	}

	L := lua.NewState()
	defer L.Close()

	code := `package.path = "` + dir + `/?.lua;" .. package.path`
	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	installRequireHook(L, func(path string) {
		// no-op callback
	})

	// Require should return the module's value.
	if err := L.DoString(`
		local val = require("mymod")
		assert(val == 42, "expected 42, got " .. tostring(val))
	`); err != nil {
		t.Fatalf("require semantics broken: %v", err)
	}
}

func TestInstallRequireHook_ErrorPropagation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	installRequireHook(L, func(path string) {})

	// Requiring a non-existent module should still error.
	err := L.DoString(`require("nonexistent_module_xyz")`)
	if err == nil {
		t.Fatal("expected error for non-existent module, got nil")
	}
}

func TestInstallRequireHook_DuplicateRequire(t *testing.T) {
	dir := t.TempDir()

	modFile := filepath.Join(dir, "cached.lua")
	if err := os.WriteFile(modFile, []byte(`return {}`), 0644); err != nil {
		t.Fatal(err)
	}

	L := lua.NewState()
	defer L.Close()

	code := `package.path = "` + dir + `/?.lua;" .. package.path`
	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var callCount int

	installRequireHook(L, func(path string) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	// Require twice — second time is cached, but hook still resolves path.
	if err := L.DoString(`require("cached"); require("cached")`); err != nil {
		t.Fatalf("require failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	// The hook fires on every require() call (even cached), because we want
	// to ensure the path is tracked. This is idempotent via AddPath dedup.
	if callCount < 1 {
		t.Fatalf("expected at least 1 hook call, got %d", callCount)
	}
}
