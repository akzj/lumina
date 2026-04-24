package lumina

import (
	"sync"
)

// AriaAttributes represents ARIA accessibility attributes on a UI element.
type AriaAttributes struct {
	Role        string  // e.g. "button", "dialog", "navigation"
	Label       string  // aria-label
	Description string  // aria-description
	Expanded    *bool   // aria-expanded
	Selected    *bool   // aria-selected
	Disabled    *bool   // aria-disabled
	Hidden      *bool   // aria-hidden
	Live        string  // aria-live: "polite" | "assertive" | "off"
	Controls    string  // aria-controls (ID of controlled element)
	LabelledBy  string  // aria-labelledby (ID of labelling element)
	DescribedBy string  // aria-describedby
	HasPopup    string  // aria-haspopup
	Pressed     *bool   // aria-pressed (for toggle buttons)
	Checked     *bool   // aria-checked
	Level       int     // aria-level (heading level)
	ValueNow    float64 // aria-valuenow (progress)
	ValueMin    float64 // aria-valuemin
	ValueMax    float64 // aria-valuemax
}

// BoolPtr is a helper to create a *bool.
func BoolPtr(v bool) *bool {
	return &v
}

// Announcer manages screen reader announcements.
type Announcer struct {
	mu    sync.Mutex
	queue []Announcement
}

// Announcement is a single screen reader announcement.
type Announcement struct {
	Message  string
	Priority string // "polite" | "assertive"
}

var globalAnnouncer = &Announcer{}

// GetAnnouncer returns the global announcer instance.
func GetAnnouncer() *Announcer {
	return globalAnnouncer
}

// Announce adds a message to the announcement queue.
func (a *Announcer) Announce(message, priority string) {
	if priority == "" {
		priority = "polite"
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.queue = append(a.queue, Announcement{
		Message:  message,
		Priority: priority,
	})
}

// Drain returns all pending announcements and clears the queue.
func (a *Announcer) Drain() []Announcement {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]Announcement, len(a.queue))
	copy(result, a.queue)
	a.queue = a.queue[:0]
	return result
}

// Pending returns the number of pending announcements.
func (a *Announcer) Pending() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.queue)
}

// Reset clears all announcements (for testing).
func (a *Announcer) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.queue = nil
}

// ParseAriaFromMap extracts ARIA attributes from a map (e.g. from Lua table).
func ParseAriaFromMap(m map[string]any) AriaAttributes {
	attrs := AriaAttributes{}
	if v, ok := m["role"].(string); ok {
		attrs.Role = v
	}
	if v, ok := m["label"].(string); ok {
		attrs.Label = v
	}
	if v, ok := m["description"].(string); ok {
		attrs.Description = v
	}
	if v, ok := m["expanded"].(bool); ok {
		attrs.Expanded = &v
	}
	if v, ok := m["selected"].(bool); ok {
		attrs.Selected = &v
	}
	if v, ok := m["disabled"].(bool); ok {
		attrs.Disabled = &v
	}
	if v, ok := m["hidden"].(bool); ok {
		attrs.Hidden = &v
	}
	if v, ok := m["live"].(string); ok {
		attrs.Live = v
	}
	if v, ok := m["controls"].(string); ok {
		attrs.Controls = v
	}
	if v, ok := m["labelledBy"].(string); ok {
		attrs.LabelledBy = v
	}
	if v, ok := m["pressed"].(bool); ok {
		attrs.Pressed = &v
	}
	if v, ok := m["checked"].(bool); ok {
		attrs.Checked = &v
	}
	return attrs
}

// AriaToMap converts AriaAttributes to a map for Lua consumption.
func AriaToMap(attrs AriaAttributes) map[string]any {
	m := make(map[string]any)
	if attrs.Role != "" {
		m["role"] = attrs.Role
	}
	if attrs.Label != "" {
		m["label"] = attrs.Label
	}
	if attrs.Description != "" {
		m["description"] = attrs.Description
	}
	if attrs.Expanded != nil {
		m["expanded"] = *attrs.Expanded
	}
	if attrs.Selected != nil {
		m["selected"] = *attrs.Selected
	}
	if attrs.Disabled != nil {
		m["disabled"] = *attrs.Disabled
	}
	if attrs.Hidden != nil {
		m["hidden"] = *attrs.Hidden
	}
	if attrs.Live != "" {
		m["live"] = attrs.Live
	}
	if attrs.Controls != "" {
		m["controls"] = attrs.Controls
	}
	if attrs.LabelledBy != "" {
		m["labelledBy"] = attrs.LabelledBy
	}
	if attrs.Pressed != nil {
		m["pressed"] = *attrs.Pressed
	}
	if attrs.Checked != nil {
		m["checked"] = *attrs.Checked
	}
	return m
}
