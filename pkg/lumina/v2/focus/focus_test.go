package focus

import "testing"

// --- Registration ---

func TestRegisterAndCount(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Register("b", 2)
	m.Register("c", 3)

	if got := m.Count(); got != 3 {
		t.Fatalf("Count() = %d, want 3", got)
	}
}

func TestRegisterDuplicateUpdates(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Register("a", 5) // update tabIndex

	if got := m.Count(); got != 1 {
		t.Fatalf("Count() = %d after duplicate register, want 1", got)
	}

	// Should still be navigable.
	m.FocusFirst()
	if got := m.FocusedID(); got != "a" {
		t.Fatalf("FocusedID() = %q, want %q", got, "a")
	}
}

func TestUnregister(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Register("b", 2)
	m.Unregister("a")

	if got := m.Count(); got != 1 {
		t.Fatalf("Count() = %d, want 1", got)
	}

	ids := m.Focusables()
	if len(ids) != 1 || ids[0] != "b" {
		t.Fatalf("Focusables() = %v, want [b]", ids)
	}
}

func TestUnregisterFocusedBlurs(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Focus("a")
	m.Unregister("a")

	if got := m.FocusedID(); got != "" {
		t.Fatalf("FocusedID() = %q after unregistering focused, want empty", got)
	}
}

func TestUnregisterNonExistent(t *testing.T) {
	m := New()
	m.Unregister("nonexistent") // should not panic
}

func TestClear(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Register("b", 2)
	m.Focus("a")
	m.Clear()

	if got := m.Count(); got != 0 {
		t.Fatalf("Count() = %d after Clear(), want 0", got)
	}
	if got := m.FocusedID(); got != "" {
		t.Fatalf("FocusedID() = %q after Clear(), want empty", got)
	}
}

// --- Focus / Blur ---

func TestFocusAndBlur(t *testing.T) {
	m := New()
	m.Register("a", 1)

	m.Focus("a")
	if !m.IsFocused("a") {
		t.Fatal("IsFocused(a) = false, want true")
	}

	m.Blur()
	if m.FocusedID() != "" {
		t.Fatalf("FocusedID() = %q after Blur(), want empty", m.FocusedID())
	}
}

func TestFocusSameIDNoOp(t *testing.T) {
	m := New()
	m.Register("a", 1)

	calls := 0
	m.OnChange(func(_, _ string) { calls++ })

	m.Focus("a")
	m.Focus("a") // same ID, should not trigger onChange again

	if calls != 1 {
		t.Fatalf("onChange called %d times, want 1", calls)
	}
}

func TestBlurWhenAlreadyBlurred(t *testing.T) {
	m := New()
	calls := 0
	m.OnChange(func(_, _ string) { calls++ })

	m.Blur() // already blurred, should not trigger onChange

	if calls != 0 {
		t.Fatalf("onChange called %d times on no-op Blur, want 0", calls)
	}
}

// --- OnChange ---

func TestOnChange(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Register("b", 2)

	var oldGot, newGot string
	m.OnChange(func(old, new string) {
		oldGot = old
		newGot = new
	})

	m.Focus("a")
	if oldGot != "" || newGot != "a" {
		t.Fatalf("onChange got (%q, %q), want (\"\", \"a\")", oldGot, newGot)
	}

	m.Focus("b")
	if oldGot != "a" || newGot != "b" {
		t.Fatalf("onChange got (%q, %q), want (\"a\", \"b\")", oldGot, newGot)
	}
}

func TestMultipleOnChange(t *testing.T) {
	m := New()
	m.Register("a", 1)

	calls := 0
	m.OnChange(func(_, _ string) { calls++ })
	m.OnChange(func(_, _ string) { calls++ })

	m.Focus("a")
	if calls != 2 {
		t.Fatalf("onChange called %d times, want 2", calls)
	}
}

// --- FocusNext / FocusPrev ---

func TestFocusNextCycle(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Register("b", 2)
	m.Register("c", 3)

	// Starting from nothing: should go to first.
	m.FocusNext()
	if got := m.FocusedID(); got != "a" {
		t.Fatalf("FocusNext from none = %q, want a", got)
	}

	m.FocusNext()
	if got := m.FocusedID(); got != "b" {
		t.Fatalf("FocusNext = %q, want b", got)
	}

	m.FocusNext()
	if got := m.FocusedID(); got != "c" {
		t.Fatalf("FocusNext = %q, want c", got)
	}

	// Wrap around.
	m.FocusNext()
	if got := m.FocusedID(); got != "a" {
		t.Fatalf("FocusNext wrap = %q, want a", got)
	}
}

func TestFocusPrevCycle(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Register("b", 2)
	m.Register("c", 3)

	// Starting from nothing: should go to last.
	m.FocusPrev()
	if got := m.FocusedID(); got != "c" {
		t.Fatalf("FocusPrev from none = %q, want c", got)
	}

	m.FocusPrev()
	if got := m.FocusedID(); got != "b" {
		t.Fatalf("FocusPrev = %q, want b", got)
	}

	m.FocusPrev()
	if got := m.FocusedID(); got != "a" {
		t.Fatalf("FocusPrev = %q, want a", got)
	}

	// Wrap around.
	m.FocusPrev()
	if got := m.FocusedID(); got != "c" {
		t.Fatalf("FocusPrev wrap = %q, want c", got)
	}
}

func TestFocusNextEmpty(t *testing.T) {
	m := New()
	m.FocusNext() // should not panic
	if got := m.FocusedID(); got != "" {
		t.Fatalf("FocusNext empty = %q, want empty", got)
	}
}

func TestFocusPrevEmpty(t *testing.T) {
	m := New()
	m.FocusPrev() // should not panic
	if got := m.FocusedID(); got != "" {
		t.Fatalf("FocusPrev empty = %q, want empty", got)
	}
}

func TestFocusNextSingle(t *testing.T) {
	m := New()
	m.Register("only", 1)

	m.FocusNext()
	if got := m.FocusedID(); got != "only" {
		t.Fatalf("FocusNext single = %q, want only", got)
	}

	m.FocusNext() // wrap to itself
	if got := m.FocusedID(); got != "only" {
		t.Fatalf("FocusNext single wrap = %q, want only", got)
	}
}

// --- FocusFirst / FocusLast ---

func TestFocusFirst(t *testing.T) {
	m := New()
	m.Register("c", 3)
	m.Register("a", 1)
	m.Register("b", 2)

	m.FocusFirst()
	if got := m.FocusedID(); got != "a" {
		t.Fatalf("FocusFirst = %q, want a (lowest tabIndex)", got)
	}
}

func TestFocusLast(t *testing.T) {
	m := New()
	m.Register("c", 3)
	m.Register("a", 1)
	m.Register("b", 2)

	m.FocusLast()
	if got := m.FocusedID(); got != "c" {
		t.Fatalf("FocusLast = %q, want c (highest tabIndex)", got)
	}
}

func TestFocusFirstEmpty(t *testing.T) {
	m := New()
	m.FocusFirst() // should not panic
}

func TestFocusLastEmpty(t *testing.T) {
	m := New()
	m.FocusLast() // should not panic
}

// --- TabIndex ordering ---

func TestTabIndexOrdering(t *testing.T) {
	m := New()
	m.Register("z", 1)
	m.Register("a", 3)
	m.Register("m", 2)

	ids := m.Focusables()
	want := []string{"z", "m", "a"}
	if len(ids) != len(want) {
		t.Fatalf("Focusables() = %v, want %v", ids, want)
	}
	for i, id := range ids {
		if id != want[i] {
			t.Fatalf("Focusables()[%d] = %q, want %q", i, id, want[i])
		}
	}
}

func TestSameTabIndexSortedByID(t *testing.T) {
	m := New()
	m.Register("c", 1)
	m.Register("a", 1)
	m.Register("b", 1)

	ids := m.Focusables()
	want := []string{"a", "b", "c"}
	for i, id := range ids {
		if id != want[i] {
			t.Fatalf("Focusables()[%d] = %q, want %q (same tabIndex, sorted by ID)", i, id, want[i])
		}
	}
}

// --- Scopes ---

func TestPushScopeRestrictsNavigation(t *testing.T) {
	m := New()
	m.Register("global1", 1)
	m.Register("global2", 2)
	m.RegisterInScope("modal1", 1, "dialog")
	m.RegisterInScope("modal2", 2, "dialog")

	// Global scope: 4 elements (global + scoped visible).
	if got := m.Count(); got != 4 {
		t.Fatalf("global Count() = %d, want 4", got)
	}

	m.PushScope("dialog")
	if got := m.ActiveScope(); got != "dialog" {
		t.Fatalf("ActiveScope() = %q, want dialog", got)
	}

	// Scoped: only dialog elements.
	if got := m.Count(); got != 2 {
		t.Fatalf("scoped Count() = %d, want 2", got)
	}

	ids := m.Focusables()
	if len(ids) != 2 || ids[0] != "modal1" || ids[1] != "modal2" {
		t.Fatalf("scoped Focusables() = %v, want [modal1, modal2]", ids)
	}
}

func TestPopScopeRestoresGlobal(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.RegisterInScope("b", 1, "s1")

	m.PushScope("s1")
	if got := m.Count(); got != 1 {
		t.Fatalf("scoped Count() = %d, want 1", got)
	}

	popped := m.PopScope()
	if popped != "s1" {
		t.Fatalf("PopScope() = %q, want s1", popped)
	}

	if got := m.Count(); got != 2 {
		t.Fatalf("global Count() = %d, want 2", got)
	}
}

func TestNestedScopes(t *testing.T) {
	m := New()
	m.RegisterInScope("s1-a", 1, "scope1")
	m.RegisterInScope("s2-a", 1, "scope2")

	m.PushScope("scope1")
	m.PushScope("scope2")

	if got := m.ActiveScope(); got != "scope2" {
		t.Fatalf("ActiveScope() = %q, want scope2", got)
	}
	if got := m.Count(); got != 1 {
		t.Fatalf("scope2 Count() = %d, want 1", got)
	}

	m.PopScope()
	if got := m.ActiveScope(); got != "scope1" {
		t.Fatalf("ActiveScope() after pop = %q, want scope1", got)
	}
	if got := m.Count(); got != 1 {
		t.Fatalf("scope1 Count() = %d, want 1", got)
	}
}

func TestPopScopeEmpty(t *testing.T) {
	m := New()
	got := m.PopScope()
	if got != "" {
		t.Fatalf("PopScope() on empty = %q, want empty", got)
	}
}

func TestFocusNextWithinScope(t *testing.T) {
	m := New()
	m.Register("global1", 1)
	m.RegisterInScope("modal1", 1, "dialog")
	m.RegisterInScope("modal2", 2, "dialog")
	m.Register("global2", 2)

	m.PushScope("dialog")
	m.FocusNext()
	if got := m.FocusedID(); got != "modal1" {
		t.Fatalf("scoped FocusNext = %q, want modal1", got)
	}

	m.FocusNext()
	if got := m.FocusedID(); got != "modal2" {
		t.Fatalf("scoped FocusNext = %q, want modal2", got)
	}

	// Wrap within scope.
	m.FocusNext()
	if got := m.FocusedID(); got != "modal1" {
		t.Fatalf("scoped FocusNext wrap = %q, want modal1", got)
	}
}

func TestFocusPrevWithinScope(t *testing.T) {
	m := New()
	m.RegisterInScope("m1", 1, "s")
	m.RegisterInScope("m2", 2, "s")
	m.RegisterInScope("m3", 3, "s")

	m.PushScope("s")
	m.FocusPrev()
	if got := m.FocusedID(); got != "m3" {
		t.Fatalf("scoped FocusPrev from none = %q, want m3", got)
	}

	m.FocusPrev()
	if got := m.FocusedID(); got != "m2" {
		t.Fatalf("scoped FocusPrev = %q, want m2", got)
	}
}

// --- Focusables query ---

func TestFocusablesReturnsCorrectOrder(t *testing.T) {
	m := New()
	m.Register("third", 30)
	m.Register("first", 10)
	m.Register("second", 20)

	ids := m.Focusables()
	if len(ids) != 3 {
		t.Fatalf("Focusables() len = %d, want 3", len(ids))
	}
	if ids[0] != "first" || ids[1] != "second" || ids[2] != "third" {
		t.Fatalf("Focusables() = %v, want [first, second, third]", ids)
	}
}

// --- OnChange with Clear ---

func TestClearTriggersOnChange(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Focus("a")

	var oldGot, newGot string
	m.OnChange(func(old, new string) {
		oldGot = old
		newGot = new
	})

	m.Clear()
	if oldGot != "a" || newGot != "" {
		t.Fatalf("onChange on Clear got (%q, %q), want (\"a\", \"\")", oldGot, newGot)
	}
}

// --- OnChange with Unregister ---

func TestUnregisterTriggersOnChange(t *testing.T) {
	m := New()
	m.Register("a", 1)
	m.Focus("a")

	var oldGot, newGot string
	m.OnChange(func(old, new string) {
		oldGot = old
		newGot = new
	})

	m.Unregister("a")
	if oldGot != "a" || newGot != "" {
		t.Fatalf("onChange on Unregister got (%q, %q), want (\"a\", \"\")", oldGot, newGot)
	}
}
