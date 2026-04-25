package lumina

import (
	"math"
	"sort"
	"sync"
)

// EasingFunc takes a normalized time t ∈ [0,1] and returns the eased value.
type EasingFunc func(t float64) float64

// Built-in easing functions.

// EaseLinear returns t unchanged.
func EaseLinear(t float64) float64 { return t }

// EaseIn starts slow and accelerates (quadratic).
func EaseIn(t float64) float64 { return t * t }

// EaseOut starts fast and decelerates (quadratic).
func EaseOut(t float64) float64 { return t * (2 - t) }

// EaseInOut is a smoothstep curve — slow start and end, fast middle.
func EaseInOut(t float64) float64 { return t * t * (3 - 2*t) }

// EaseBounce simulates a bouncing effect at the end.
func EaseBounce(t float64) float64 {
	const n1 = 7.5625
	const d1 = 2.75

	if t < 1/d1 {
		return n1 * t * t
	} else if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	} else if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	}
	t -= 2.625 / d1
	return n1*t*t + 0.984375
}

// EaseElastic produces an elastic/spring effect.
func EaseElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return -math.Pow(2, 10*(t-1)) * math.Sin((t-1.1)*5*math.Pi)
}

// easingByName returns an easing function by name, defaulting to linear.
func easingByName(name string) EasingFunc {
	switch name {
	case "easeIn":
		return EaseIn
	case "easeOut":
		return EaseOut
	case "easeInOut":
		return EaseInOut
	case "bounce":
		return EaseBounce
	case "elastic":
		return EaseElastic
	default:
		return EaseLinear
	}
}

// AnimationState tracks a single running animation.
type AnimationState struct {
	ID        string     // unique identifier
	StartTime int64      // start time in milliseconds
	Duration  int64      // duration in milliseconds
	From      float64    // start value
	To        float64    // end value
	Current   float64    // current interpolated value
	Easing    EasingFunc // easing function
	Done      bool       // true when animation is complete
	Loop      bool       // if true, repeats indefinitely
	CompID    string     // owning component ID (for dirty marking)
}

// AnimationManager manages all running animations.
type AnimationManager struct {
	animations map[string]*AnimationState
	mu         sync.RWMutex
}

// NewAnimationManager creates a new AnimationManager.
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		animations: make(map[string]*AnimationState),
	}
}

// Start registers and starts an animation. If an animation with the same ID
// already exists, it is replaced.
func (am *AnimationManager) Start(anim *AnimationState) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if anim.Easing == nil {
		anim.Easing = EaseLinear
	}
	anim.Current = anim.From
	anim.Done = false
	am.animations[anim.ID] = anim
}

// Stop removes an animation by ID.
func (am *AnimationManager) Stop(id string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.animations, id)
}

// Get returns an animation by ID, or nil if not found.
func (am *AnimationManager) Get(id string) *AnimationState {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.animations[id]
}

// GetState returns a snapshot of an animation's current value and done status.
// Returns (current, done, found). Safe for concurrent use — values are copied under lock.
func (am *AnimationManager) GetState(id string) (current float64, done bool, found bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	anim, ok := am.animations[id]
	if !ok {
		return 0, false, false
	}
	return anim.Current, anim.Done, true
}

// Tick updates all running animations based on the current time.
// Returns the set of component IDs that need re-rendering.
func (am *AnimationManager) Tick(nowMs int64) []string {
	am.mu.Lock()
	defer am.mu.Unlock()

	var dirtyComps []string
	seen := make(map[string]bool)

	for id, anim := range am.animations {
		if anim.Done && !anim.Loop {
			continue
		}

		elapsed := nowMs - anim.StartTime
		if elapsed < 0 {
			elapsed = 0
		}

		var t float64
		if anim.Duration <= 0 {
			t = 1.0
		} else {
			t = float64(elapsed) / float64(anim.Duration)
		}

		if anim.Loop {
			// For looping, wrap t into [0, 1)
			if t >= 1.0 {
				t = t - math.Floor(t)
			}
		} else if t >= 1.0 {
			t = 1.0
			anim.Done = true
		}

		eased := anim.Easing(t)
		anim.Current = anim.From + (anim.To-anim.From)*eased

		if anim.CompID != "" && !seen[anim.CompID] {
			dirtyComps = append(dirtyComps, anim.CompID)
			seen[anim.CompID] = true
		}

		// Remove completed non-looping animations
		if anim.Done && !anim.Loop {
			// Keep it around so Get() can read the final value,
			// but it won't be ticked again (checked at top of loop)
			_ = id
		}
	}

	return dirtyComps
}

// Count returns the number of animations (including completed ones).
func (am *AnimationManager) Count() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.animations)
}

// Active returns the number of animations that are still running.
func (am *AnimationManager) Active() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	count := 0
	for _, a := range am.animations {
		if !a.Done || a.Loop {
			count++
		}
	}
	return count
}

// Clear removes all animations.
func (am *AnimationManager) Clear() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.animations = make(map[string]*AnimationState)
}

// GetAll returns all animations sorted by ID for deterministic iteration.
func (am *AnimationManager) GetAll() []*AnimationState {
	am.mu.RLock()
	defer am.mu.RUnlock()
	result := make([]*AnimationState, 0, len(am.animations))
	for _, a := range am.animations {
		result = append(result, a)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// globalAnimationManager is the singleton animation manager.
var globalAnimationManager = NewAnimationManager()
