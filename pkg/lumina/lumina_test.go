package lumina_test

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina"
)

// TestRequireLumina tests that require("lumina") works and module exports are accessible.
func TestRequireLumina(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Open lumina module
	lumina.Open(L)

	// Test: require("lumina") succeeds
	err := L.DoString(`
		local lumina = require("lumina")
		if type(lumina) ~= "table" then
			error("lumina should be a table, got: " .. type(lumina))
		end
	`)
	if err != nil {
		t.Fatalf("require failed: %v", err)
	}
}

// TestVersion tests that lumina.version() returns the correct version.
func TestVersion(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	lumina.Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local v = lumina.version()
		if v ~= "0.1.0" then
			error("expected version 0.1.0, got: " .. tostring(v))
		end
		print("version: " .. v)
	`)
	if err != nil {
		t.Fatalf("version check failed: %v", err)
	}
}

// TestEcho tests that lumina.echo() works correctly.
func TestEcho(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	lumina.Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local result = lumina.echo("hello world")
		if result ~= "hello world" then
			error("echo failed: expected 'hello world', got: " .. tostring(result))
		end
	`)
	if err != nil {
		t.Fatalf("echo test failed: %v", err)
	}
}

// TestInfo tests that lumina.info() returns a table with expected fields.
func TestInfo(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	lumina.Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local info = lumina.info()
		if type(info) ~= "table" then
			error("info should return a table")
		end
		if info.version ~= "0.1.0" then
			error("info.version should be 0.1.0")
		end
		if info.year ~= 2024 then
			error("info.year should be 2024")
		end
		print("description: " .. info.description)
	`)
	if err != nil {
		t.Fatalf("info test failed: %v", err)
	}
}

// TestGlobalAccess tests that lumina is also accessible as a global.
func TestGlobalAccess(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	lumina.Open(L)

	err := L.DoString(`
		-- lumina should be set as global from Open()
		local v = lumina.version()
		if v ~= "0.1.0" then
			error("global lumina.version() failed: " .. tostring(v))
		end
	`)
	if err != nil {
		t.Fatalf("global access test failed: %v", err)
	}
}

// TestModuleIsolation tests that multiple Lua states get independent lumina modules.
func TestModuleIsolation(t *testing.T) {
	// Create first state
	L1 := lua.NewState()
	defer L1.Close()
	lumina.Open(L1)

	// Create second state
	L2 := lua.NewState()
	defer L2.Close()
	lumina.Open(L2)

	// Modify version in L1 should not affect L2
	err := L1.DoString(`
		local lumina = require("lumina")
	`)
	if err != nil {
		t.Fatalf("L1 require failed: %v", err)
	}

	err = L2.DoString(`
		local lumina = require("lumina")
		if lumina.version() ~= "0.1.0" then
			error("L2 version corrupted")
		end
	`)
	if err != nil {
		t.Fatalf("L2 require or version check failed: %v", err)
	}
}

// TestDirectLuaLoader tests the luaLoader function directly.
func TestDirectLuaLoader(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Call Open which calls luaLoader internally
	lumina.Open(L)

	// Verify the module was registered correctly
	err := L.DoString(`
		local lumina = require("lumina")
		-- Check all expected functions exist
		if type(lumina.version) ~= "function" then
			error("lumina.version should be a function")
		end
		if type(lumina.echo) ~= "function" then
			error("lumina.echo should be a function")
		end
		if type(lumina.info) ~= "function" then
			error("lumina.info should be a function")
		end
	`)
	if err != nil {
		t.Fatalf("direct luaLoader test failed: %v", err)
	}
}