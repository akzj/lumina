package animation

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// ─── Easing function tests ───

func TestEasingBoundaries(t *testing.T) {
	easings := []struct {
		name string
		fn   EasingFunc
	}{
		{"Linear", Linear},
		{"EaseIn", EaseIn},
		{"EaseOut", EaseOut},
		{"EaseInOut", EaseInOut},
		{"Bounce", Bounce},
		{"Elastic", Elastic},
	}

	for _, e := range easings {
		t.Run(e.name, func(t *testing.T) {
			v0 := e.fn(0)
			if !approxEqual(v0, 0) {
				t.Errorf("%s(0) = %v, want 0", e.name, v0)
			}
			v1 := e.fn(1)
			if !approxEqual(v1, 1) {
				t.Errorf("%s(1) = %v, want 1", e.name, v1)
			}
		})
	}
}

func TestEasingMonotonicity(t *testing.T) {
	// Linear, EaseIn, EaseOut, EaseInOut should be monotonically non-decreasing
	monotonic := []struct {
		name string
		fn   EasingFunc
	}{
		{"Linear", Linear},
		{"EaseIn", EaseIn},
		{"EaseOut", EaseOut},
		{"EaseInOut", EaseInOut},
	}

	for _, e := range monotonic {
		t.Run(e.name, func(t *testing.T) {
			prev := e.fn(0)
			for i := 1; i <= 100; i++ {
				tt := float64(i) / 100.0
				cur := e.fn(tt)
				if cur < prev-epsilon {
					t.Errorf("%s not monotonic: f(%v)=%v < f(%v)=%v",
						e.name, tt, cur, float64(i-1)/100.0, prev)
				}
				prev = cur
			}
		})
	}
}

func TestEasingByNameKnown(t *testing.T) {
	cases := []struct {
		name string
		want EasingFunc
	}{
		{"linear", Linear},
		{"easeIn", EaseIn},
		{"easeOut", EaseOut},
		{"easeInOut", EaseInOut},
		{"bounce", Bounce},
		{"elastic", Elastic},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fn := EasingByName(c.name)
			// Verify at a few points
			for _, tt := range []float64{0, 0.25, 0.5, 0.75, 1.0} {
				got := fn(tt)
				want := c.want(tt)
				if !approxEqual(got, want) {
					t.Errorf("EasingByName(%q)(%.2f) = %v, want %v", c.name, tt, got, want)
				}
			}
		})
	}
}

func TestEasingByNameUnknown(t *testing.T) {
	fn := EasingByName("nonexistent")
	// Should return Linear
	if v := fn(0.5); !approxEqual(v, 0.5) {
		t.Errorf("EasingByName(unknown)(0.5) = %v, want 0.5 (Linear)", v)
	}
}

// ─── Animation tests ───

func TestAnimationBasicTick(t *testing.T) {
	anim := New(Config{
		ID:       "test",
		From:     0,
		To:       100,
		Duration: 1000,
		Easing:   "linear",
	}, 0)

	// At t=0
	v := anim.Tick(0)
	if !approxEqual(v, 0) {
		t.Errorf("Tick(0) = %v, want 0", v)
	}

	// At t=500 (halfway)
	v = anim.Tick(500)
	if !approxEqual(v, 50) {
		t.Errorf("Tick(500) = %v, want 50", v)
	}

	// At t=1000 (done)
	v = anim.Tick(1000)
	if !approxEqual(v, 100) {
		t.Errorf("Tick(1000) = %v, want 100", v)
	}
	if !anim.IsDone() {
		t.Error("animation should be done at t=duration")
	}

	// After done, still returns final value
	v = anim.Tick(2000)
	if !approxEqual(v, 100) {
		t.Errorf("Tick(2000) after done = %v, want 100", v)
	}
}

func TestAnimationAccessors(t *testing.T) {
	anim := New(Config{
		ID:       "acc",
		From:     10,
		To:       20,
		Duration: 100,
	}, 0)

	if anim.ID() != "acc" {
		t.Errorf("ID() = %q, want %q", anim.ID(), "acc")
	}
	if !approxEqual(anim.Current(), 10) {
		t.Errorf("Current() = %v, want 10", anim.Current())
	}
	if anim.IsDone() {
		t.Error("IsDone() should be false initially")
	}
}

func TestAnimationWithEasing(t *testing.T) {
	anim := New(Config{
		ID:       "eased",
		From:     0,
		To:       100,
		Duration: 1000,
		Easing:   "easeIn",
	}, 0)

	// At halfway, easeIn(0.5) = 0.25, so value should be 25
	v := anim.Tick(500)
	expected := 100 * EaseIn(0.5) // 25
	if !approxEqual(v, expected) {
		t.Errorf("Tick(500) with easeIn = %v, want %v", v, expected)
	}
}

func TestAnimationLoop(t *testing.T) {
	anim := New(Config{
		ID:       "loop",
		From:     0,
		To:       100,
		Duration: 1000,
		Easing:   "linear",
		Loop:     true,
	}, 0)

	// Run to completion
	anim.Tick(1000)
	if anim.IsDone() {
		t.Error("looping animation should not be done")
	}

	// After loop reset, ticking at 0 elapsed should give from value
	v := anim.Tick(1000)
	if !approxEqual(v, 0) {
		t.Errorf("after loop reset, Tick = %v, want 0", v)
	}

	// Advance within new cycle
	v = anim.Tick(1500)
	if !approxEqual(v, 50) {
		t.Errorf("in second loop cycle, Tick(1500) = %v, want 50", v)
	}
}

func TestAnimationCallbacks(t *testing.T) {
	var updates []float64
	doneCalled := 0

	anim := New(Config{
		ID:       "cb",
		From:     0,
		To:       10,
		Duration: 100,
		Easing:   "linear",
		OnUpdate: func(v float64) { updates = append(updates, v) },
		OnDone:   func() { doneCalled++ },
	}, 0)

	anim.Tick(50) // mid
	anim.Tick(100) // end

	if len(updates) != 2 {
		t.Errorf("onUpdate called %d times, want 2", len(updates))
	}
	if doneCalled != 1 {
		t.Errorf("onDone called %d times, want 1", doneCalled)
	}

	// After done, no more callbacks
	anim.Tick(200)
	if len(updates) != 2 {
		t.Errorf("onUpdate called after done: %d times", len(updates))
	}
	if doneCalled != 1 {
		t.Errorf("onDone called after done: %d times", doneCalled)
	}
}

func TestAnimationZeroDuration(t *testing.T) {
	doneCalled := 0
	anim := New(Config{
		ID:       "zero",
		From:     0,
		To:       100,
		Duration: 0,
		OnDone:   func() { doneCalled++ },
	}, 0)

	v := anim.Tick(0)
	if !approxEqual(v, 100) {
		t.Errorf("zero duration Tick = %v, want 100", v)
	}
	if !anim.IsDone() {
		t.Error("zero duration should be immediately done")
	}
	if doneCalled != 1 {
		t.Errorf("onDone called %d times, want 1", doneCalled)
	}
}

func TestAnimationNegativeElapsed(t *testing.T) {
	anim := New(Config{
		ID:       "neg",
		From:     0,
		To:       100,
		Duration: 1000,
		Easing:   "linear",
	}, 100)

	// Tick before start time
	v := anim.Tick(50)
	if !approxEqual(v, 0) {
		t.Errorf("Tick before start = %v, want 0 (from)", v)
	}
}

func TestAnimationFromEqualsTo(t *testing.T) {
	anim := New(Config{
		ID:       "same",
		From:     42,
		To:       42,
		Duration: 1000,
		Easing:   "linear",
	}, 0)

	v := anim.Tick(500)
	if !approxEqual(v, 42) {
		t.Errorf("from==to Tick = %v, want 42", v)
	}
}

func TestAnimationReset(t *testing.T) {
	anim := New(Config{
		ID:       "reset",
		From:     0,
		To:       100,
		Duration: 1000,
		Easing:   "linear",
	}, 0)

	anim.Tick(1000) // complete
	if !anim.IsDone() {
		t.Error("should be done")
	}

	anim.Reset(2000)
	if anim.IsDone() {
		t.Error("should not be done after reset")
	}
	if !approxEqual(anim.Current(), 0) {
		t.Errorf("Current after reset = %v, want 0", anim.Current())
	}

	v := anim.Tick(2500)
	if !approxEqual(v, 50) {
		t.Errorf("Tick after reset = %v, want 50", v)
	}
}

// ─── Manager tests ───

func TestManagerStartStop(t *testing.T) {
	m := NewManager()

	m.Start(Config{ID: "a", From: 0, To: 10, Duration: 100}, 0)
	m.Start(Config{ID: "b", From: 0, To: 20, Duration: 200}, 0)

	if m.Count() != 2 {
		t.Errorf("Count = %d, want 2", m.Count())
	}

	m.Stop("a")
	if m.Count() != 1 {
		t.Errorf("Count after stop = %d, want 1", m.Count())
	}
	if m.Get("a") != nil {
		t.Error("Get('a') should be nil after stop")
	}
	if m.Get("b") == nil {
		t.Error("Get('b') should not be nil")
	}
}

func TestManagerReplace(t *testing.T) {
	m := NewManager()

	m.Start(Config{ID: "x", From: 0, To: 10, Duration: 100}, 0)
	m.Start(Config{ID: "x", From: 50, To: 100, Duration: 200}, 0)

	if m.Count() != 1 {
		t.Errorf("Count after replace = %d, want 1", m.Count())
	}

	anim := m.Get("x")
	if anim == nil {
		t.Fatal("Get('x') = nil")
	}
	if !approxEqual(anim.Current(), 50) {
		t.Errorf("replaced animation From = %v, want 50", anim.Current())
	}
}

func TestManagerTick(t *testing.T) {
	m := NewManager()

	m.Start(Config{ID: "short", From: 0, To: 10, Duration: 100, Easing: "linear"}, 0)
	m.Start(Config{ID: "long", From: 0, To: 20, Duration: 200, Easing: "linear"}, 0)

	// Tick at 100ms: "short" completes, "long" is at halfway
	completed := m.Tick(100)
	if len(completed) != 1 || completed[0] != "short" {
		t.Errorf("completed = %v, want [short]", completed)
	}

	// "short" should be removed
	if m.Get("short") != nil {
		t.Error("completed animation should be removed")
	}
	if m.Count() != 1 {
		t.Errorf("Count after completion = %d, want 1", m.Count())
	}

	// "long" should be at halfway
	longAnim := m.Get("long")
	if longAnim == nil {
		t.Fatal("'long' should still exist")
	}
	if !approxEqual(longAnim.Current(), 10) {
		t.Errorf("long.Current() = %v, want 10", longAnim.Current())
	}
}

func TestManagerTickMultipleComplete(t *testing.T) {
	m := NewManager()

	m.Start(Config{ID: "a", From: 0, To: 10, Duration: 100}, 0)
	m.Start(Config{ID: "b", From: 0, To: 20, Duration: 100}, 0)

	completed := m.Tick(100)
	if len(completed) != 2 {
		t.Errorf("completed count = %d, want 2", len(completed))
	}
	// Should be sorted
	if len(completed) == 2 && (completed[0] != "a" || completed[1] != "b") {
		t.Errorf("completed = %v, want [a, b] (sorted)", completed)
	}
	if m.Count() != 0 {
		t.Errorf("Count after all complete = %d, want 0", m.Count())
	}
}

func TestManagerStopAll(t *testing.T) {
	m := NewManager()

	m.Start(Config{ID: "a", From: 0, To: 10, Duration: 100}, 0)
	m.Start(Config{ID: "b", From: 0, To: 20, Duration: 200}, 0)
	m.Start(Config{ID: "c", From: 0, To: 30, Duration: 300}, 0)

	m.StopAll()
	if m.Count() != 0 {
		t.Errorf("Count after StopAll = %d, want 0", m.Count())
	}
	if m.IsRunning() {
		t.Error("IsRunning should be false after StopAll")
	}
}

func TestManagerIsRunning(t *testing.T) {
	m := NewManager()

	if m.IsRunning() {
		t.Error("IsRunning should be false when empty")
	}

	m.Start(Config{ID: "a", From: 0, To: 10, Duration: 100}, 0)
	if !m.IsRunning() {
		t.Error("IsRunning should be true with active animation")
	}

	m.Tick(100) // complete
	if m.IsRunning() {
		t.Error("IsRunning should be false after all complete")
	}
}

func TestManagerLoopNotRemoved(t *testing.T) {
	m := NewManager()

	m.Start(Config{ID: "loop", From: 0, To: 10, Duration: 100, Loop: true}, 0)

	completed := m.Tick(100)
	if len(completed) != 0 {
		t.Errorf("looping animation should not appear in completed: %v", completed)
	}
	if m.Count() != 1 {
		t.Errorf("looping animation should still be active, Count = %d", m.Count())
	}
}
