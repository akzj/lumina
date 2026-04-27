package animation

// Animation tracks a single running animation.
type Animation struct {
	id        string
	from      float64
	to        float64
	current   float64
	duration  int64      // milliseconds
	startTime int64      // milliseconds
	easing    EasingFunc
	done      bool
	loop      bool
	onUpdate  func(value float64) // called on each tick
	onDone    func()              // called when complete
}

// Config for creating an animation.
type Config struct {
	ID       string
	From     float64
	To       float64
	Duration int64  // milliseconds
	Easing   string // easing name (or "linear" default)
	Loop     bool
	OnUpdate func(value float64)
	OnDone   func()
}

// New creates a new Animation from config.
// The animation starts at nowMs — call Tick to advance it.
func New(cfg Config, nowMs int64) *Animation {
	easing := EasingByName(cfg.Easing)
	return &Animation{
		id:        cfg.ID,
		from:      cfg.From,
		to:        cfg.To,
		current:   cfg.From,
		duration:  cfg.Duration,
		startTime: nowMs,
		easing:    easing,
		done:      false,
		loop:      cfg.Loop,
		onUpdate:  cfg.OnUpdate,
		onDone:    cfg.OnDone,
	}
}

// ID returns the animation's identifier.
func (a *Animation) ID() string { return a.id }

// Current returns the current interpolated value.
func (a *Animation) Current() float64 { return a.current }

// IsDone returns true if the animation has completed.
func (a *Animation) IsDone() bool { return a.done }

// Tick advances the animation to the given time (milliseconds).
// Returns the interpolated value.
func (a *Animation) Tick(nowMs int64) float64 {
	if a.done {
		return a.current
	}

	elapsed := nowMs - a.startTime
	if elapsed < 0 {
		elapsed = 0
	}

	if a.duration <= 0 {
		// Zero or negative duration: snap to end immediately.
		a.current = a.to
		a.done = true
		if a.onUpdate != nil {
			a.onUpdate(a.to)
		}
		if a.onDone != nil {
			a.onDone()
		}
		return a.current
	}

	if elapsed >= a.duration {
		if a.loop {
			a.startTime = nowMs
			elapsed = 0
		} else {
			a.current = a.to
			a.done = true
			if a.onUpdate != nil {
				a.onUpdate(a.to)
			}
			if a.onDone != nil {
				a.onDone()
			}
			return a.current
		}
	}

	t := float64(elapsed) / float64(a.duration)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	eased := a.easing(t)
	a.current = a.from + (a.to-a.from)*eased

	if a.onUpdate != nil {
		a.onUpdate(a.current)
	}
	return a.current
}

// Reset restarts the animation from the beginning at the given time.
func (a *Animation) Reset(nowMs int64) {
	a.startTime = nowMs
	a.current = a.from
	a.done = false
}
