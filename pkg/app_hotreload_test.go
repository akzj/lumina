package v2

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/output"
)

// --- pathToModuleName tests ---

func TestPathToModuleName_SimpleModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	libPath := filepath.Join(dir, "mylib.lua")
	writeFile(t, libPath, `local M = {} function M.greet() return "hello" end return M`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		local mylib = require("mylib")
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = mylib.greet()})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	got := app.pathToModuleName(libPath)
	if got != "mylib" {
		t.Errorf("pathToModuleName(%q) = %q, want %q", libPath, got, "mylib")
	}
}

func TestPathToModuleName_NestedModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "lib"), 0755); err != nil {
		t.Fatal(err)
	}
	libPath := filepath.Join(dir, "lib", "utils.lua")
	writeFile(t, libPath, `local M = {} function M.name() return "utils" end return M`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		local utils = require("lib.utils")
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = utils.name()})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	got := app.pathToModuleName(libPath)
	if got != "lib.utils" {
		t.Errorf("pathToModuleName(%q) = %q, want %q", libPath, got, "lib.utils")
	}
}

func TestPathToModuleName_InitLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "mymod"), 0755); err != nil {
		t.Fatal(err)
	}
	initPath := filepath.Join(dir, "mymod", "init.lua")
	writeFile(t, initPath, `local M = {} function M.val() return 42 end return M`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		local mymod = require("mymod")
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = tostring(mymod.val())})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	got := app.pathToModuleName(initPath)
	if got != "mymod" {
		t.Errorf("pathToModuleName(%q) = %q, want %q", initPath, got, "mymod")
	}
}

func TestPathToModuleName_NonLuaFile(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	got := app.pathToModuleName("/some/path/file.txt")
	if got != "" {
		t.Errorf("pathToModuleName for non-.lua file = %q, want empty", got)
	}
}

// --- isModuleLoaded tests ---

func TestIsModuleLoaded_True(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	libPath := filepath.Join(dir, "testmod.lua")
	writeFile(t, libPath, `local M = {} function M.f() return 1 end return M`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `require("testmod")`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	if !app.isModuleLoaded("testmod") {
		t.Error("expected testmod to be loaded")
	}
}

func TestIsModuleLoaded_False(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	if app.isModuleLoaded("nonexistent_module_xyz") {
		t.Error("expected nonexistent module to not be loaded")
	}
}

// --- Module-level hot reload integration tests ---

func TestHotReload_ModuleReload_SwapsCode(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	libPath := filepath.Join(dir, "mylib.lua")
	writeFile(t, libPath, `
		local M = {}
		function M.greet() return "hello" end
		return M
	`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		local mylib = require("mylib")
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = mylib.greet()})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.RenderAll()

	// Verify initial render
	if !screenHasString(ta, "hello") {
		t.Fatal("expected 'hello' on screen after initial render")
	}

	// Modify the library
	writeFile(t, libPath, `
		local M = {}
		function M.greet() return "world" end
		return M
	`)

	// Trigger module-level reload
	ok := app.reloadModule(libPath)
	if !ok {
		t.Fatal("reloadModule returned false, expected true")
	}

	// Write output to test adapter
	screen := app.engine.ToBuffer()
	_ = ta.WriteFull(screen)

	// Verify new content
	if !screenHasString(ta, "world") {
		t.Fatal("expected 'world' on screen after module reload")
	}
	if screenHasString(ta, "hello") {
		t.Fatal("'hello' should no longer be on screen after module reload")
	}
}

func TestHotReload_ModuleReload_PreservesState(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	libPath := filepath.Join(dir, "display.lua")
	writeFile(t, libPath, `
		local M = {}
		function M.format(n) return "count:" .. tostring(n) end
		return M
	`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		local display = require("display")
		lumina.createComponent({
			id = "root",
			render = function(props)
				local count, setCount = lumina.useState("count", 5)
				return lumina.createElement("text", {content = display.format(count)})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.RenderAll()

	// Verify initial render shows count:5
	screen := app.engine.ToBuffer()
	_ = ta.WriteFull(screen)
	if !screenHasString(ta, "count:5") {
		t.Fatal("expected 'count:5' on screen")
	}

	// Modify display module to change format
	writeFile(t, libPath, `
		local M = {}
		function M.format(n) return "N=" .. tostring(n) end
		return M
	`)

	// Module reload
	ok := app.reloadModule(libPath)
	if !ok {
		t.Fatal("reloadModule returned false")
	}

	screen = app.engine.ToBuffer()
	_ = ta.WriteFull(screen)

	// State should be preserved (count=5) but format changed
	if !screenHasString(ta, "N=5") {
		t.Fatal("expected 'N=5' on screen after reload (state preserved, format changed)")
	}
}

func TestHotReload_FallbackToFullReload(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = "v1"})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.RenderAll()

	// The main script is NOT a module (not loaded via require),
	// so reloadModule should return false
	ok := app.reloadModule(mainPath)
	if ok {
		t.Error("reloadModule should return false for main script (not a module)")
	}

	// Now test that reloadScript falls back to full reload
	writeFile(t, mainPath, `
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = "v2"})
			end
		})
	`)

	app.reloadScript(mainPath)

	screen := app.engine.ToBuffer()
	_ = ta.WriteFull(screen)
	if !screenHasString(ta, "v2") {
		t.Fatal("expected 'v2' on screen after full reload fallback")
	}
}

func TestHotReload_ReloadScript_TriesModuleFirst(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	libPath := filepath.Join(dir, "greeter.lua")
	writeFile(t, libPath, `
		local M = {}
		function M.say() return "hi" end
		return M
	`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		local greeter = require("greeter")
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = greeter.say()})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.RenderAll()

	screen := app.engine.ToBuffer()
	_ = ta.WriteFull(screen)
	if !screenHasString(ta, "hi") {
		t.Fatal("expected 'hi' on screen")
	}

	// Modify module and call reloadScript (which tries module reload first)
	writeFile(t, libPath, `
		local M = {}
		function M.say() return "bye" end
		return M
	`)

	app.reloadScript(libPath)

	screen = app.engine.ToBuffer()
	_ = ta.WriteFull(screen)
	if !screenHasString(ta, "bye") {
		t.Fatal("expected 'bye' on screen after reloadScript with module")
	}
}

func TestHotReload_ModuleReload_CompileError_DoesNotCrash(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	libPath := filepath.Join(dir, "buglib.lua")
	writeFile(t, libPath, `
		local M = {}
		function M.val() return "ok" end
		return M
	`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		local buglib = require("buglib")
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = buglib.val()})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.RenderAll()

	// Write broken code
	writeFile(t, libPath, `
		this is not valid lua!!!
	`)

	// Should not panic — returns false
	ok := app.reloadModule(libPath)
	if ok {
		t.Error("reloadModule should return false for broken module")
	}

	// Original content should still work
	screen := app.engine.ToBuffer()
	_ = ta.WriteFull(screen)
	if !screenHasString(ta, "ok") {
		t.Fatal("expected 'ok' still on screen after failed reload")
	}
}

func TestHotReload_LuaReloadAPI(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	libPath := filepath.Join(dir, "apilib.lua")
	writeFile(t, libPath, `
		local M = {}
		function M.msg() return "before" end
		return M
	`)

	mainPath := filepath.Join(dir, "main.lua")
	writeFile(t, mainPath, `
		local apilib = require("apilib")
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = apilib.msg()})
			end
		})
	`)

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.RenderAll()

	screen := app.engine.ToBuffer()
	_ = ta.WriteFull(screen)
	if !screenHasString(ta, "before") {
		t.Fatal("expected 'before' on screen")
	}

	// Modify the library
	writeFile(t, libPath, `
		local M = {}
		function M.msg() return "after" end
		return M
	`)

	// Use the Lua API to reload
	if err := app.RunString(`lumina.reload("apilib")`); err != nil {
		t.Fatalf("lumina.reload failed: %v", err)
	}
	app.RenderDirty()

	screen = app.engine.ToBuffer()
	_ = ta.WriteFull(screen)
	if !screenHasString(ta, "after") {
		t.Fatal("expected 'after' on screen after lumina.reload()")
	}
}

func TestHotReload_LuaReloadAPI_Error(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	_ = NewApp(L, 40, 10, ta)

	// Reload a module that doesn't exist — should return nil, error
	err := L.DoString(`
		local result, err = lumina.reload("nonexistent_xyz")
		if result ~= nil then error("expected nil result") end
		if err == nil then error("expected error message") end
	`)
	if err != nil {
		t.Fatalf("Lua error: %v", err)
	}
}

// --- collectLuaFiles tests ---

func TestCollectLuaFiles(t *testing.T) {
	dir := t.TempDir()

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(dir, "lib"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "main.lua"), "-- main")
	writeFile(t, filepath.Join(dir, "lib", "utils.lua"), "-- utils")
	writeFile(t, filepath.Join(dir, "lib", "helpers.lua"), "-- helpers")
	writeFile(t, filepath.Join(dir, "readme.txt"), "not lua")

	files := collectLuaFiles(dir)
	if len(files) != 3 {
		t.Errorf("expected 3 .lua files, got %d: %v", len(files), files)
	}

	// Verify all are .lua files
	for _, f := range files {
		if filepath.Ext(f) != ".lua" {
			t.Errorf("non-lua file in results: %s", f)
		}
	}
}

// --- MarkAllComponentsDirty tests ---

func TestMarkAllComponentsDirty(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	if err := app.RunString(`
		lumina.createComponent({
			id = "root",
			render = function(props)
				return lumina.createElement("text", {content = "test"})
			end
		})
	`); err != nil {
		t.Fatalf("RunString: %v", err)
	}
	app.RenderAll()

	// After RenderAll, components should not be dirty
	comp := app.engine.GetComponent("root")
	if comp == nil {
		t.Fatal("root component not found")
	}
	if comp.Dirty {
		t.Error("component should not be dirty after RenderAll")
	}

	// Mark all dirty
	app.engine.MarkAllComponentsDirty()

	if !comp.Dirty {
		t.Error("component should be dirty after MarkAllComponentsDirty")
	}
	if !app.engine.NeedsRender() {
		t.Error("engine should need render after MarkAllComponentsDirty")
	}
}
