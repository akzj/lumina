package lumina

import (
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestPatchAPIExists(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	Open(L)

	L.DoString(`
		local lumina = require("lumina")
		assert(type(lumina.patch) == "function", "patch should exist")
		assert(type(lumina.eval) == "function", "eval should exist")
	`)
}

func TestProfileAPIExists(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	Open(L)

	L.DoString(`
		local lumina = require("lumina")
		assert(type(lumina.profile) == "function", "profile should exist")
		assert(type(lumina.profileReset) == "function", "profileReset should exist")
	`)
}

func TestProfileCore(t *testing.T) {
	RecordFrameTiming(time.Millisecond * 5)
	RecordFrameTiming(time.Millisecond * 10)
	RecordFrameTiming(time.Millisecond * 15)

	result := ProfileFrames()
	if result.TotalFrames != 3 {
		t.Errorf("Expected 3 frames, got %d", result.TotalFrames)
	}

	ResetProfile()
	result2 := ProfileFrames()
	if result2.TotalFrames != 0 {
		t.Errorf("After reset, expected 0 frames")
	}
}

func TestDiffCore(t *testing.T) {
	before := NewFrame(5, 3)
	before.Cells[0][0] = Cell{Char: 'A'}
	after := NewFrame(5, 3)
	after.Cells[0][0] = Cell{Char: 'B'}
	result := DiffFrames(before, after)
	if len(result.Patches) < 1 {
		t.Error("Expected at least 1 patch")
	}
}

func TestConsoleCore(t *testing.T) {
	globalConsole.Clear()
	globalConsole.Log("log", "test", nil)
	globalConsole.Log("warn", "warning", nil)
	globalConsole.Log("error", "error", nil)
	entries := globalConsole.GetEntries()
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
	errors := globalConsole.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}
}
