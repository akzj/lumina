package lumina_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

// ─── Task 1: lumina init — Project Scaffolding ───────────────────────────

func TestInitCreatesProjectStructure(t *testing.T) {
	// Create project in temp dir
	tmpDir := t.TempDir()
	projName := filepath.Join(tmpDir, "testapp")

	err := lumina.ScaffoldProject(projName)
	if err != nil {
		t.Fatalf("ScaffoldProject failed: %v", err)
	}

	// Verify all files exist
	expectedFiles := []string{
		"main.lua",
		"components/hello.lua",
		"lumina.json",
		"README.md",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(projName, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s not created", f)
		}
	}

	// Verify lumina.json content
	cfgData, err := os.ReadFile(filepath.Join(projName, "lumina.json"))
	if err != nil {
		t.Fatalf("read lumina.json: %v", err)
	}

	var cfg lumina.ProjectConfig
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		t.Fatalf("parse lumina.json: %v", err)
	}

	if cfg.Name != projName {
		t.Errorf("expected name %q, got %q", projName, cfg.Name)
	}
	if cfg.Version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", cfg.Version)
	}
	if cfg.Entry != "main.lua" {
		t.Errorf("expected entry 'main.lua', got %q", cfg.Entry)
	}
	if cfg.Theme != "catppuccin-mocha" {
		t.Errorf("expected theme 'catppuccin-mocha', got %q", cfg.Theme)
	}

	// Verify main.lua contains expected content
	mainData, err := os.ReadFile(filepath.Join(projName, "main.lua"))
	if err != nil {
		t.Fatalf("read main.lua: %v", err)
	}
	mainStr := string(mainData)

	if !strings.Contains(mainStr, "lumina.defineComponent") {
		t.Error("main.lua should contain 'lumina.defineComponent'")
	}
	if !strings.Contains(mainStr, "lumina.mount") {
		t.Error("main.lua should contain 'lumina.mount'")
	}
	if !strings.Contains(mainStr, "lumina.run") {
		t.Error("main.lua should contain 'lumina.run'")
	}

	// Verify README.md contains project name
	readmeData, err := os.ReadFile(filepath.Join(projName, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	if !strings.Contains(string(readmeData), "Lumina") {
		t.Error("README.md should mention Lumina")
	}

	// Verify hello.lua is a valid component
	helloData, err := os.ReadFile(filepath.Join(projName, "components", "hello.lua"))
	if err != nil {
		t.Fatalf("read hello.lua: %v", err)
	}
	if !strings.Contains(string(helloData), "defineComponent") {
		t.Error("hello.lua should contain 'defineComponent'")
	}
}

func TestInitMainLuaLoads(t *testing.T) {
	// Create project in temp dir
	tmpDir := t.TempDir()
	projName := filepath.Join(tmpDir, "loadtest")

	err := lumina.ScaffoldProject(projName)
	if err != nil {
		t.Fatalf("ScaffoldProject failed: %v", err)
	}

	// Load the generated main.lua headlessly
	app := lumina.NewApp()
	defer app.Close()

	tio := lumina.NewBufferTermIO(80, 24, nil)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	mainPath := filepath.Join(projName, "main.lua")
	err = app.LoadScript(mainPath, tio)
	if err != nil {
		t.Fatalf("generated main.lua failed to load: %v", err)
	}

	// Should produce output
	if len(tio.Output()) == 0 {
		t.Error("generated main.lua produced no output")
	}
}

func TestInitDuplicateProject(t *testing.T) {
	tmpDir := t.TempDir()
	projName := filepath.Join(tmpDir, "duptest")

	// First creation should succeed
	err := lumina.ScaffoldProject(projName)
	if err != nil {
		t.Fatalf("first ScaffoldProject failed: %v", err)
	}

	// Second creation should also succeed (overwrites)
	err = lumina.ScaffoldProject(projName)
	if err != nil {
		t.Fatalf("second ScaffoldProject failed: %v", err)
	}
}

// ─── Task 3: Error Reporting ─────────────────────────────────────────────

func TestFormatLuaError(t *testing.T) {
	err := fmt.Errorf("main.lua:42: attempt to call a nil value")
	result := lumina.FormatLuaErrorPlain(err, "Counter")

	if !strings.Contains(result, "Lumina Error") {
		t.Error("should contain 'Lumina Error' header")
	}
	if !strings.Contains(result, "Counter") {
		t.Error("should contain component name 'Counter'")
	}
	if !strings.Contains(result, "attempt to call a nil value") {
		t.Error("should contain error message")
	}
}

func TestFormatLuaErrorParsesFileAndLine(t *testing.T) {
	err := fmt.Errorf("runtime error: app.lua:99: undefined variable 'foo'")
	le := lumina.ParseLuaError(err, "MyApp")

	if le.File != "app.lua" {
		t.Errorf("expected file 'app.lua', got %q", le.File)
	}
	if le.Line != 99 {
		t.Errorf("expected line 99, got %d", le.Line)
	}
	if le.Component != "MyApp" {
		t.Errorf("expected component 'MyApp', got %q", le.Component)
	}
	if !strings.Contains(le.Message, "undefined variable") {
		t.Errorf("expected message about 'undefined variable', got %q", le.Message)
	}
}

func TestFormatLuaErrorNoFile(t *testing.T) {
	err := fmt.Errorf("some generic error")
	le := lumina.ParseLuaError(err, "")

	if le.File != "" {
		t.Errorf("expected empty file, got %q", le.File)
	}
	if le.Line != 0 {
		t.Errorf("expected line 0, got %d", le.Line)
	}

	// Should still format without crashing
	result := le.FormatPlain()
	if !strings.Contains(result, "some generic error") {
		t.Error("should contain the error message")
	}
}

func TestFormatLuaErrorNil(t *testing.T) {
	le := lumina.ParseLuaError(nil, "")
	if le.Message != "" {
		t.Errorf("nil error should produce empty message, got %q", le.Message)
	}
}

func TestFormatLuaErrorWithStack(t *testing.T) {
	le := lumina.LuaError{
		File:      "test.lua",
		Line:      10,
		Component: "TestComp",
		Message:   "stack overflow",
		Stack:     "test.lua:10: in function 'render'\ntest.lua:5: in main chunk",
	}

	result := le.FormatPlain()
	if !strings.Contains(result, "Stack") || !strings.Contains(result, "render") {
		t.Error("should contain stack trace")
	}
}

func TestFormatLuaErrorColored(t *testing.T) {
	err := fmt.Errorf("app.lua:1: syntax error")
	result := lumina.FormatLuaError(err, "App")

	// Should contain ANSI color codes
	if !strings.Contains(result, "\033[") {
		t.Error("colored format should contain ANSI escape codes")
	}
	if !strings.Contains(result, "Lumina Error") {
		t.Error("should contain 'Lumina Error' header")
	}
}

// ─── Task 4: Version Command ─────────────────────────────────────────────

func TestVersionString(t *testing.T) {
	// Verify the version constant is accessible
	app := lumina.NewApp()
	defer app.Close()

	err := app.L.DoString(`
		local lumina = require("lumina")
		local v = lumina.version()
		assert(v == "0.3.0", "expected version 0.3.0, got " .. tostring(v))
	`)
	if err != nil {
		t.Fatalf("version check failed: %v", err)
	}
}

// ─── Task 5: Dev Mode Flags ─────────────────────────────────────────────

func TestDevModeEnablesHotReload(t *testing.T) {
	app := lumina.NewApp()
	defer app.Close()

	// Simulate what dev mode does: enable hot reload via Lua
	err := app.L.DoString(`
		local lumina = require("lumina")
		lumina.enableHotReload({ interval = 500 })
	`)
	if err != nil {
		t.Fatalf("enableHotReload failed: %v", err)
	}

	// Disable should also work
	err = app.L.DoString(`
		local lumina = require("lumina")
		lumina.disableHotReload()
	`)
	if err != nil {
		t.Fatalf("disableHotReload failed: %v", err)
	}
}

// ─── Scaffold Edge Cases ─────────────────────────────────────────────────

func TestInitProjectNameInContent(t *testing.T) {
	tmpDir := t.TempDir()
	projName := filepath.Join(tmpDir, "myspecialapp")

	err := lumina.ScaffoldProject(projName)
	if err != nil {
		t.Fatalf("ScaffoldProject failed: %v", err)
	}

	// main.lua should contain the project name in the welcome message
	mainData, err := os.ReadFile(filepath.Join(projName, "main.lua"))
	if err != nil {
		t.Fatalf("read main.lua: %v", err)
	}

	if !strings.Contains(string(mainData), "myspecialapp") {
		t.Error("main.lua should contain the project name in welcome message")
	}
}
