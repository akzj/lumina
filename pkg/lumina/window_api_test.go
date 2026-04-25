package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func windowTestState(t *testing.T) *lua.State {
	t.Helper()
	ClearComponents()
	ClearContextValues()
	globalWindowManager.Clear()
	L := lua.NewState()
	Open(L)
	return L
}

func TestWindowAPI_CreateAndClose(t *testing.T) {
	L := windowTestState(t)
	defer L.Close()

	err := L.DoString(`
		local lumina = require("lumina")
		local id = lumina.createWindow({ id = "w1", title = "Test Window", x = 10, y = 5, w = 40, h = 20 })
		assert(id == "w1", "expected id w1, got " .. tostring(id))
	`)
	if err != nil {
		t.Fatalf("createWindow: %v", err)
	}

	// Verify window exists
	win := globalWindowManager.GetWindow("w1")
	if win == nil {
		t.Fatal("window w1 not found after createWindow")
	}
	if win.Title != "Test Window" {
		t.Errorf("title = %q, want 'Test Window'", win.Title)
	}
	if win.X != 10 || win.Y != 5 {
		t.Errorf("position = (%d,%d), want (10,5)", win.X, win.Y)
	}

	// Close via Lua
	err = L.DoString(`
		local lumina = require("lumina")
		lumina.closeWindow("w1")
	`)
	if err != nil {
		t.Fatalf("closeWindow: %v", err)
	}
	if globalWindowManager.Count() != 0 {
		t.Error("window should be removed after closeWindow")
	}
}

func TestWindowAPI_MoveAndResize(t *testing.T) {
	L := windowTestState(t)
	defer L.Close()

	err := L.DoString(`
		local lumina = require("lumina")
		lumina.createWindow({ id = "w1", title = "Movable", x = 0, y = 0, w = 30, h = 15 })
		lumina.moveWindow("w1", 20, 10)
		lumina.resizeWindow("w1", 50, 25)
	`)
	if err != nil {
		t.Fatalf("Lua error: %v", err)
	}

	win := globalWindowManager.GetWindow("w1")
	if win.X != 20 || win.Y != 10 {
		t.Errorf("position = (%d,%d), want (20,10)", win.X, win.Y)
	}
	if win.W != 50 || win.H != 25 {
		t.Errorf("size = (%d,%d), want (50,25)", win.W, win.H)
	}
}

func TestWindowAPI_MaximizeRestore(t *testing.T) {
	L := windowTestState(t)
	defer L.Close()

	err := L.DoString(`
		local lumina = require("lumina")
		lumina.createWindow({ id = "w1", title = "Max", x = 5, y = 3, w = 30, h = 15 })
		lumina.maximizeWindow("w1")
	`)
	if err != nil {
		t.Fatalf("Lua error: %v", err)
	}

	win := globalWindowManager.GetWindow("w1")
	if !win.Maximized {
		t.Error("window should be maximized")
	}
	if win.X != 0 || win.Y != 0 {
		t.Errorf("maximized position = (%d,%d), want (0,0)", win.X, win.Y)
	}

	err = L.DoString(`
		local lumina = require("lumina")
		lumina.restoreWindow("w1")
	`)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	win = globalWindowManager.GetWindow("w1")
	if win.Maximized {
		t.Error("should not be maximized after restore")
	}
	if win.X != 5 || win.Y != 3 {
		t.Errorf("restored position = (%d,%d), want (5,3)", win.X, win.Y)
	}
}

func TestWindowAPI_TileWindows(t *testing.T) {
	L := windowTestState(t)
	defer L.Close()

	err := L.DoString(`
		local lumina = require("lumina")
		lumina.createWindow({ id = "w1", title = "A", x = 0, y = 0, w = 20, h = 10 })
		lumina.createWindow({ id = "w2", title = "B", x = 0, y = 0, w = 20, h = 10 })
		lumina.tileWindows("horizontal")
	`)
	if err != nil {
		t.Fatalf("Lua error: %v", err)
	}

	w1 := globalWindowManager.GetWindow("w1")
	w2 := globalWindowManager.GetWindow("w2")
	// Tiled horizontally: w1 left half, w2 right half
	if w1.X != 0 {
		t.Errorf("w1.X = %d, want 0", w1.X)
	}
	if w2.X == 0 {
		t.Error("w2.X should not be 0 after horizontal tile")
	}
}

func TestWindowAPI_FocusWindow(t *testing.T) {
	L := windowTestState(t)
	defer L.Close()

	err := L.DoString(`
		local lumina = require("lumina")
		lumina.createWindow({ id = "w1", title = "A", x = 0, y = 0, w = 20, h = 10 })
		lumina.createWindow({ id = "w2", title = "B", x = 10, y = 5, w = 20, h = 10 })
		lumina.focusWindow("w1")
	`)
	if err != nil {
		t.Fatalf("Lua error: %v", err)
	}

	focused := globalWindowManager.GetFocused()
	if focused == nil || focused.ID != "w1" {
		t.Errorf("focused = %v, want w1", focused)
	}
}
