package render

import "testing"

func TestDepsEqual_NonComparable(t *testing.T) {
	// Should not panic on maps (from Lua tables)
	a := []any{map[string]any{"x": 1}, []any{1, 2}}
	b := []any{map[string]any{"x": 1}, []any{1, 2}}
	c := []any{map[string]any{"x": 2}, []any{1, 2}}

	if !depsEqual(a, b) {
		t.Error("equal maps/slices should be equal")
	}
	if depsEqual(a, c) {
		t.Error("different maps should not be equal")
	}
}

func TestDepsEqual_Comparable(t *testing.T) {
	// Basic comparable types still work
	if !depsEqual([]any{1, "hello", true}, []any{1, "hello", true}) {
		t.Error("equal comparable values should be equal")
	}
	if depsEqual([]any{1, "hello"}, []any{1, "world"}) {
		t.Error("different comparable values should not be equal")
	}
	if depsEqual([]any{1}, []any{1, 2}) {
		t.Error("different length slices should not be equal")
	}
}

func TestDepsEqual_Nil(t *testing.T) {
	if !depsEqual(nil, nil) {
		t.Error("nil deps should be equal")
	}
	if !depsEqual([]any{nil}, []any{nil}) {
		t.Error("nil elements should be equal")
	}
	if depsEqual([]any{nil}, []any{1}) {
		t.Error("nil vs non-nil should not be equal")
	}
}
