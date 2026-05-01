package animation

import "sort"

// Manager manages multiple concurrent animations.
// Not safe for concurrent use — callers must synchronize externally.
type Manager struct {
	animations map[string]*Animation
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{
		animations: make(map[string]*Animation),
	}
}

// Start begins a new animation. If an animation with the same ID exists,
// it is replaced. Returns the created Animation.
func (m *Manager) Start(cfg Config, nowMs int64) *Animation {
	anim := New(cfg, nowMs)
	m.animations[anim.id] = anim
	return anim
}

// Stop stops and removes an animation by ID.
func (m *Manager) Stop(id string) {
	delete(m.animations, id)
}

// StopAll stops and removes all animations.
func (m *Manager) StopAll() {
	m.animations = make(map[string]*Animation)
}

// Tick advances all animations to the given time.
// Returns IDs of animations that completed this tick (sorted for determinism).
// Completed non-looping animations are removed after ticking.
func (m *Manager) Tick(nowMs int64) []string {
	var completed []string

	for id, anim := range m.animations {
		wasDone := anim.done
		anim.Tick(nowMs)
		if anim.done && !wasDone {
			completed = append(completed, id)
		}
	}

	// Remove completed non-looping animations.
	for _, id := range completed {
		delete(m.animations, id)
	}

	sort.Strings(completed)
	return completed
}

// Get returns an animation by ID, or nil if not found.
func (m *Manager) Get(id string) *Animation {
	return m.animations[id]
}

// Count returns the number of active (running) animations.
func (m *Manager) Count() int {
	return len(m.animations)
}

// IsRunning returns true if any animations are active.
func (m *Manager) IsRunning() bool {
	return len(m.animations) > 0
}

// StartAnim implements render.AnimationManager interface.
// Starts a new animation and returns its initial value.
func (m *Manager) StartAnim(id string, from, to float64, duration int64, easing string, loop bool, onUpdate func(float64), onDone func(), nowMs int64) float64 {
	anim := m.Start(Config{
		ID:       id,
		From:     from,
		To:       to,
		Duration: duration,
		Easing:   easing,
		Loop:     loop,
		OnUpdate: onUpdate,
		OnDone:   onDone,
	}, nowMs)
	return anim.Current()
}

// GetAnimValue implements render.AnimationManager interface.
// Returns the current value of an animation, or (0, false) if not found.
func (m *Manager) GetAnimValue(id string) (float64, bool) {
	anim := m.Get(id)
	if anim == nil {
		return 0, false
	}
	return anim.Current(), true
}

// StopAnim implements render.AnimationManager interface.
func (m *Manager) StopAnim(id string) {
	m.Stop(id)
}
