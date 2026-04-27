// Package focus provides focus scope management, Tab/Shift-Tab navigation,
// and focus ring tracking. Pure Go, zero v2 dependencies.
package focus

import "sort"

// Manager manages focus state and navigation.
type Manager struct {
	focusedID  string
	focusables []entry
	scopes     []string // focus scope stack
	onChange   []func(oldID, newID string)
}

type entry struct {
	id       string
	tabIndex int
	scope    string // "" = global
}

// New creates a new focus Manager.
func New() *Manager {
	return &Manager{}
}

// --- Registration ---

// Register adds a focusable element with a tab index.
func (m *Manager) Register(id string, tabIndex int) {
	m.RegisterInScope(id, tabIndex, "")
}

// RegisterInScope adds a focusable element within a named scope.
func (m *Manager) RegisterInScope(id string, tabIndex int, scope string) {
	// Avoid duplicates.
	for i, e := range m.focusables {
		if e.id == id {
			m.focusables[i].tabIndex = tabIndex
			m.focusables[i].scope = scope
			return
		}
	}
	m.focusables = append(m.focusables, entry{
		id:       id,
		tabIndex: tabIndex,
		scope:    scope,
	})
}

// Unregister removes a focusable element.
func (m *Manager) Unregister(id string) {
	for i, e := range m.focusables {
		if e.id == id {
			m.focusables = append(m.focusables[:i], m.focusables[i+1:]...)
			if m.focusedID == id {
				m.setFocused("")
			}
			return
		}
	}
}

// Clear removes all focusable elements and blurs.
func (m *Manager) Clear() {
	m.focusables = m.focusables[:0]
	if m.focusedID != "" {
		m.setFocused("")
	}
}

// --- Focus Control ---

// Focus sets focus to the given ID. Triggers onChange callbacks.
func (m *Manager) Focus(id string) {
	if m.focusedID == id {
		return
	}
	m.setFocused(id)
}

// Blur removes focus. Triggers onChange.
func (m *Manager) Blur() {
	if m.focusedID == "" {
		return
	}
	m.setFocused("")
}

// FocusedID returns the currently focused element ID.
func (m *Manager) FocusedID() string {
	return m.focusedID
}

// IsFocused returns true if the given ID is focused.
func (m *Manager) IsFocused(id string) bool {
	return m.focusedID == id
}

// --- Navigation ---

// FocusNext moves focus to the next focusable element (Tab).
// Wraps around. Respects active scope.
func (m *Manager) FocusNext() {
	candidates := m.activeFocusables()
	if len(candidates) == 0 {
		return
	}

	currentIdx := -1
	for i, e := range candidates {
		if e.id == m.focusedID {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(candidates)
	m.setFocused(candidates[nextIdx].id)
}

// FocusPrev moves focus to the previous focusable element (Shift-Tab).
// Wraps around. Respects active scope.
func (m *Manager) FocusPrev() {
	candidates := m.activeFocusables()
	if len(candidates) == 0 {
		return
	}

	currentIdx := -1
	for i, e := range candidates {
		if e.id == m.focusedID {
			currentIdx = i
			break
		}
	}

	var prevIdx int
	if currentIdx <= 0 {
		prevIdx = len(candidates) - 1
	} else {
		prevIdx = currentIdx - 1
	}
	m.setFocused(candidates[prevIdx].id)
}

// FocusFirst focuses the first focusable element in active scope.
func (m *Manager) FocusFirst() {
	candidates := m.activeFocusables()
	if len(candidates) == 0 {
		return
	}
	m.setFocused(candidates[0].id)
}

// FocusLast focuses the last focusable element in active scope.
func (m *Manager) FocusLast() {
	candidates := m.activeFocusables()
	if len(candidates) == 0 {
		return
	}
	m.setFocused(candidates[len(candidates)-1].id)
}

// --- Scope Management ---

// PushScope pushes a new focus scope. Only elements in this scope
// will be navigable until the scope is popped.
func (m *Manager) PushScope(scope string) {
	m.scopes = append(m.scopes, scope)
}

// PopScope pops the current focus scope. Returns the popped scope name.
// Returns "" if no scope is active.
func (m *Manager) PopScope() string {
	if len(m.scopes) == 0 {
		return ""
	}
	popped := m.scopes[len(m.scopes)-1]
	m.scopes = m.scopes[:len(m.scopes)-1]
	return popped
}

// ActiveScope returns the current active scope ("" = global).
func (m *Manager) ActiveScope() string {
	if len(m.scopes) == 0 {
		return ""
	}
	return m.scopes[len(m.scopes)-1]
}

// --- Callbacks ---

// OnChange registers a callback for focus changes.
func (m *Manager) OnChange(fn func(oldID, newID string)) {
	m.onChange = append(m.onChange, fn)
}

// --- Query ---

// Focusables returns all focusable IDs in the active scope, sorted by tabIndex.
func (m *Manager) Focusables() []string {
	candidates := m.activeFocusables()
	ids := make([]string, len(candidates))
	for i, e := range candidates {
		ids[i] = e.id
	}
	return ids
}

// Count returns the number of focusable elements in active scope.
func (m *Manager) Count() int {
	return len(m.activeFocusables())
}

// --- Internal ---

func (m *Manager) setFocused(newID string) {
	oldID := m.focusedID
	m.focusedID = newID
	for _, fn := range m.onChange {
		fn(oldID, newID)
	}
}

func (m *Manager) activeFocusables() []entry {
	scope := m.ActiveScope()
	var result []entry
	for _, e := range m.focusables {
		if scope == "" || e.scope == scope {
			result = append(result, e)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].tabIndex != result[j].tabIndex {
			return result[i].tabIndex < result[j].tabIndex
		}
		return result[i].id < result[j].id
	})
	return result
}
