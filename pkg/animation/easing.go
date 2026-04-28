// Package animation provides easing functions and an animation state machine.
// Pure Go, zero v2 package dependencies.
package animation

import "math"

// EasingFunc takes a normalized time t ∈ [0,1] and returns the eased value.
type EasingFunc func(t float64) float64

// Linear returns t unchanged.
func Linear(t float64) float64 { return t }

// EaseIn starts slow and accelerates (quadratic).
func EaseIn(t float64) float64 { return t * t }

// EaseOut starts fast and decelerates (quadratic).
func EaseOut(t float64) float64 { return t * (2 - t) }

// EaseInOut is a smoothstep curve — slow start and end, fast middle.
func EaseInOut(t float64) float64 { return t * t * (3 - 2*t) }

// Bounce simulates a bouncing effect at the end.
func Bounce(t float64) float64 {
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

// Elastic produces an elastic/spring effect.
func Elastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return -math.Pow(2, 10*(t-1)) * math.Sin((t-1.1)*5*math.Pi)
}

// EasingByName returns an easing function by name.
// Returns Linear for unknown names.
func EasingByName(name string) EasingFunc {
	switch name {
	case "easeIn":
		return EaseIn
	case "easeOut":
		return EaseOut
	case "easeInOut":
		return EaseInOut
	case "bounce":
		return Bounce
	case "elastic":
		return Elastic
	case "linear":
		return Linear
	default:
		return Linear
	}
}
