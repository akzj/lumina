package lumina

import (
	"math"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// -----------------------------------------------------------------------
// Easing function tests
// -----------------------------------------------------------------------

func TestEaseLinear(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{0, 0}, {0.25, 0.25}, {0.5, 0.5}, {0.75, 0.75}, {1, 1},
	}
	for _, c := range cases {
		got := EaseLinear(c.in)
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("EaseLinear(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestEaseInOut(t *testing.T) {
	// smoothstep: t*t*(3-2t)
	// t=0 → 0, t=0.5 → 0.5, t=1 → 1
	if v := EaseInOut(0); v != 0 {
		t.Errorf("EaseInOut(0) = %v, want 0", v)
	}
	if v := EaseInOut(1); v != 1 {
		t.Errorf("EaseInOut(1) = %v, want 1", v)
	}
	if v := EaseInOut(0.5); math.Abs(v-0.5) > 1e-9 {
		t.Errorf("EaseInOut(0.5) = %v, want 0.5", v)
	}
	// t=0.25: 0.0625*(3-0.5) = 0.0625*2.5 = 0.15625
	if v := EaseInOut(0.25); math.Abs(v-0.15625) > 1e-9 {
		t.Errorf("EaseInOut(0.25) = %v, want 0.15625", v)
	}
}

func TestEaseIn(t *testing.T) {
	if v := EaseIn(0); v != 0 {
		t.Errorf("EaseIn(0) = %v", v)
	}
	if v := EaseIn(1); v != 1 {
		t.Errorf("EaseIn(1) = %v", v)
	}
	if v := EaseIn(0.5); math.Abs(v-0.25) > 1e-9 {
		t.Errorf("EaseIn(0.5) = %v, want 0.25", v)
	}
}

func TestEaseOut(t *testing.T) {
	if v := EaseOut(0); v != 0 {
		t.Errorf("EaseOut(0) = %v", v)
	}
	if v := EaseOut(1); v != 1 {
		t.Errorf("EaseOut(1) = %v", v)
	}
	// t*(2-t) at t=0.5 → 0.5*1.5 = 0.75
	if v := EaseOut(0.5); math.Abs(v-0.75) > 1e-9 {
		t.Errorf("EaseOut(0.5) = %v, want 0.75", v)
	}
}

func TestEaseBounce(t *testing.T) {
	if v := EaseBounce(0); v != 0 {
		t.Errorf("EaseBounce(0) = %v", v)
	}
	if v := EaseBounce(1); math.Abs(v-1) > 1e-6 {
		t.Errorf("EaseBounce(1) = %v, want ~1", v)
	}
	// Bounce should be in [0, 1] for all t in [0, 1]
	for i := 0; i <= 100; i++ {
		tt := float64(i) / 100.0
		v := EaseBounce(tt)
		if v < -0.01 || v > 1.01 {
			t.Errorf("EaseBounce(%v) = %v, out of range", tt, v)
		}
	}
}

func TestEaseElastic(t *testing.T) {
	if v := EaseElastic(0); v != 0 {
		t.Errorf("EaseElastic(0) = %v", v)
	}
	if v := EaseElastic(1); v != 1 {
		t.Errorf("EaseElastic(1) = %v", v)
	}
}

func TestEasingByName(t *testing.T) {
	// Known names
	if easingByName("linear")(0.5) != EaseLinear(0.5) {
		t.Error("linear mismatch")
	}
	if easingByName("easeIn")(0.5) != EaseIn(0.5) {
		t.Error("easeIn mismatch")
	}
	if easingByName("easeOut")(0.5) != EaseOut(0.5) {
		t.Error("easeOut mismatch")
	}
	if easingByName("easeInOut")(0.5) != EaseInOut(0.5) {
		t.Error("easeInOut mismatch")
	}
	if easingByName("bounce")(0.5) != EaseBounce(0.5) {
		t.Error("bounce mismatch")
	}
	// Unknown defaults to linear
	if easingByName("unknown")(0.5) != EaseLinear(0.5) {
		t.Error("unknown should default to linear")
	}
}

// -----------------------------------------------------------------------
// AnimationManager tests
// -----------------------------------------------------------------------

func TestAnimationManager_StartStop(t *testing.T) {
	am := NewAnimationManager()

	anim := &AnimationState{
		ID:       "test1",
		Duration: 1000,
		From:     0,
		To:       100,
	}
	am.Start(anim)

	if am.Count() != 1 {
		t.Fatalf("expected 1 animation, got %d", am.Count())
	}
	if am.Active() != 1 {
		t.Fatalf("expected 1 active, got %d", am.Active())
	}

	got := am.Get("test1")
	if got == nil {
		t.Fatal("expected to find test1")
	}
	if got.Current != 0 {
		t.Fatalf("expected Current=0 at start, got %v", got.Current)
	}

	am.Stop("test1")
	if am.Count() != 0 {
		t.Fatalf("expected 0 after Stop, got %d", am.Count())
	}
}

func TestAnimationManager_TickUpdatesValue(t *testing.T) {
	am := NewAnimationManager()

	anim := &AnimationState{
		ID:        "fade",
		StartTime: 1000,
		Duration:  1000,
		From:      0,
		To:        100,
		Easing:    EaseLinear,
	}
	am.Start(anim)

	// At t=500ms (halfway)
	am.Tick(1500)
	got := am.Get("fade")
	if got == nil {
		t.Fatal("animation not found")
	}
	if math.Abs(got.Current-50) > 1 {
		t.Fatalf("expected Current≈50 at halfway, got %v", got.Current)
	}
	if got.Done {
		t.Fatal("should not be done at halfway")
	}

	// At t=1000ms (complete)
	am.Tick(2000)
	got = am.Get("fade")
	if math.Abs(got.Current-100) > 0.01 {
		t.Fatalf("expected Current=100 at end, got %v", got.Current)
	}
	if !got.Done {
		t.Fatal("should be done at end")
	}
}

func TestAnimationManager_LoopAnimation(t *testing.T) {
	am := NewAnimationManager()

	anim := &AnimationState{
		ID:        "spin",
		StartTime: 0,
		Duration:  1000,
		From:      0,
		To:        360,
		Easing:    EaseLinear,
		Loop:      true,
	}
	am.Start(anim)

	// At 500ms — halfway through first cycle
	am.Tick(500)
	got := am.Get("spin")
	if math.Abs(got.Current-180) > 1 {
		t.Fatalf("expected ~180 at 500ms, got %v", got.Current)
	}
	if got.Done {
		t.Fatal("loop animation should never be done")
	}

	// At 1500ms — halfway through second cycle
	am.Tick(1500)
	got = am.Get("spin")
	if math.Abs(got.Current-180) > 1 {
		t.Fatalf("expected ~180 at 1500ms (loop), got %v", got.Current)
	}
	if got.Done {
		t.Fatal("loop animation should never be done")
	}

	// Active should still be 1
	if am.Active() != 1 {
		t.Fatalf("expected 1 active for loop, got %d", am.Active())
	}
}

func TestAnimationManager_CompletedNotActive(t *testing.T) {
	am := NewAnimationManager()

	anim := &AnimationState{
		ID:        "once",
		StartTime: 0,
		Duration:  100,
		From:      0,
		To:        1,
		Easing:    EaseLinear,
	}
	am.Start(anim)

	// Complete it
	am.Tick(200)

	// Count still 1 (animation kept for Get), Active is 0
	if am.Count() != 1 {
		t.Fatalf("expected count=1, got %d", am.Count())
	}
	if am.Active() != 0 {
		t.Fatalf("expected active=0 after completion, got %d", am.Active())
	}
}

func TestAnimationManager_TickReturnsDirtyComps(t *testing.T) {
	am := NewAnimationManager()

	anim := &AnimationState{
		ID:        "a1",
		StartTime: 0,
		Duration:  1000,
		From:      0,
		To:        1,
		Easing:    EaseLinear,
		CompID:    "comp-1",
	}
	am.Start(anim)

	dirty := am.Tick(500)
	if len(dirty) != 1 || dirty[0] != "comp-1" {
		t.Fatalf("expected dirty=[comp-1], got %v", dirty)
	}

	// After completion, no more dirty
	am.Tick(2000)
	dirty2 := am.Tick(3000)
	if len(dirty2) != 0 {
		t.Fatalf("expected no dirty after completion, got %v", dirty2)
	}
}

func TestAnimationManager_MultipleAnimations(t *testing.T) {
	am := NewAnimationManager()

	am.Start(&AnimationState{
		ID: "a", StartTime: 0, Duration: 1000, From: 0, To: 100, Easing: EaseLinear,
	})
	am.Start(&AnimationState{
		ID: "b", StartTime: 0, Duration: 500, From: 10, To: 20, Easing: EaseLinear,
	})

	if am.Count() != 2 {
		t.Fatalf("expected 2, got %d", am.Count())
	}

	am.Tick(250)

	a := am.Get("a")
	b := am.Get("b")
	if math.Abs(a.Current-25) > 1 {
		t.Fatalf("a.Current expected ~25, got %v", a.Current)
	}
	if math.Abs(b.Current-15) > 1 {
		t.Fatalf("b.Current expected ~15, got %v", b.Current)
	}

	// b should complete at 500ms
	am.Tick(500)
	b = am.Get("b")
	if !b.Done {
		t.Fatal("b should be done at 500ms")
	}
	a = am.Get("a")
	if a.Done {
		t.Fatal("a should not be done at 500ms")
	}
}

func TestAnimationManager_Clear(t *testing.T) {
	am := NewAnimationManager()
	am.Start(&AnimationState{ID: "x", Duration: 100})
	am.Start(&AnimationState{ID: "y", Duration: 100})
	am.Clear()
	if am.Count() != 0 {
		t.Fatalf("expected 0 after Clear, got %d", am.Count())
	}
}

func TestAnimationManager_GetAll(t *testing.T) {
	am := NewAnimationManager()
	am.Start(&AnimationState{ID: "c", Duration: 100})
	am.Start(&AnimationState{ID: "a", Duration: 100})
	am.Start(&AnimationState{ID: "b", Duration: 100})

	all := am.GetAll()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	// Should be sorted by ID
	if all[0].ID != "a" || all[1].ID != "b" || all[2].ID != "c" {
		t.Fatalf("expected sorted [a,b,c], got [%s,%s,%s]", all[0].ID, all[1].ID, all[2].ID)
	}
}

func TestAnimationManager_ZeroDuration(t *testing.T) {
	am := NewAnimationManager()
	am.Start(&AnimationState{
		ID: "instant", StartTime: 0, Duration: 0, From: 0, To: 42, Easing: EaseLinear,
	})

	am.Tick(0)
	got := am.Get("instant")
	if !got.Done {
		t.Fatal("zero-duration animation should be immediately done")
	}
	if math.Abs(got.Current-42) > 0.01 {
		t.Fatalf("expected Current=42, got %v", got.Current)
	}
}

func TestAnimationManager_ReplaceExisting(t *testing.T) {
	am := NewAnimationManager()

	am.Start(&AnimationState{
		ID: "x", StartTime: 0, Duration: 1000, From: 0, To: 100, Easing: EaseLinear,
	})
	am.Tick(500)
	got := am.Get("x")
	if math.Abs(got.Current-50) > 1 {
		t.Fatalf("expected ~50, got %v", got.Current)
	}

	// Replace with new animation
	am.Start(&AnimationState{
		ID: "x", StartTime: 500, Duration: 500, From: 200, To: 300, Easing: EaseLinear,
	})
	got = am.Get("x")
	if got.Current != 200 {
		t.Fatalf("expected Current=200 after replace, got %v", got.Current)
	}
	if got.Done {
		t.Fatal("replaced animation should not be done")
	}
}

// -----------------------------------------------------------------------
// Lua API tests
// -----------------------------------------------------------------------

func TestLua_StartStopAnimation(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalAnimationManager.Clear()

	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.startAnimation({
			id = "test-anim",
			from = 10,
			to = 90,
			duration = 500,
			easing = "easeOut",
		})
	`)
	if err != nil {
		t.Fatalf("startAnimation error: %v", err)
	}

	anim := globalAnimationManager.Get("test-anim")
	if anim == nil {
		t.Fatal("expected animation to exist")
	}
	if anim.From != 10 || anim.To != 90 {
		t.Fatalf("expected from=10, to=90, got from=%v, to=%v", anim.From, anim.To)
	}
	if anim.Duration != 500 {
		t.Fatalf("expected duration=500, got %d", anim.Duration)
	}

	err = L.DoString(`lumina.stopAnimation("test-anim")`)
	if err != nil {
		t.Fatalf("stopAnimation error: %v", err)
	}
	if globalAnimationManager.Get("test-anim") != nil {
		t.Fatal("expected animation to be removed after stop")
	}
	globalAnimationManager.Clear()
}

func TestLua_AnimationPresets(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalAnimationManager.Clear()

	L := lua.NewState()
	Open(L)
	defer L.Close()

	// Test fadeIn preset
	err := L.DoString(`
		local cfg = lumina.animation.fadeIn(200)
		_from = cfg.from
		_to = cfg.to
		_dur = cfg.duration
		_easing = cfg.easing
	`)
	if err != nil {
		t.Fatalf("fadeIn error: %v", err)
	}

	L.GetGlobal("_from")
	from, _ := L.ToNumber(-1)
	L.Pop(1)
	L.GetGlobal("_to")
	to, _ := L.ToNumber(-1)
	L.Pop(1)
	L.GetGlobal("_dur")
	dur, _ := L.ToNumber(-1)
	L.Pop(1)
	L.GetGlobal("_easing")
	easing, _ := L.ToString(-1)
	L.Pop(1)

	if from != 0 || to != 1 {
		t.Fatalf("fadeIn: expected from=0, to=1, got from=%v, to=%v", from, to)
	}
	if dur != 200 {
		t.Fatalf("fadeIn: expected duration=200, got %v", dur)
	}
	if easing != "easeInOut" {
		t.Fatalf("fadeIn: expected easing=easeInOut, got %q", easing)
	}

	// Test pulse preset (loop)
	err = L.DoString(`
		local cfg = lumina.animation.pulse(500)
		_loop = cfg.loop
	`)
	if err != nil {
		t.Fatalf("pulse error: %v", err)
	}
	L.GetGlobal("_loop")
	if !L.ToBoolean(-1) {
		t.Fatal("pulse should have loop=true")
	}
	L.Pop(1)

	// Test spin preset
	err = L.DoString(`
		local cfg = lumina.animation.spin(800)
		_spinTo = cfg.to
		_spinLoop = cfg.loop
	`)
	if err != nil {
		t.Fatalf("spin error: %v", err)
	}
	L.GetGlobal("_spinTo")
	spinTo, _ := L.ToNumber(-1)
	L.Pop(1)
	if spinTo != 360 {
		t.Fatalf("spin: expected to=360, got %v", spinTo)
	}
	L.GetGlobal("_spinLoop")
	if !L.ToBoolean(-1) {
		t.Fatal("spin should have loop=true")
	}
	L.Pop(1)

	// Test fadeOut preset
	err = L.DoString(`
		local cfg = lumina.animation.fadeOut(150)
		_foFrom = cfg.from
		_foTo = cfg.to
	`)
	if err != nil {
		t.Fatalf("fadeOut error: %v", err)
	}
	L.GetGlobal("_foFrom")
	foFrom, _ := L.ToNumber(-1)
	L.Pop(1)
	L.GetGlobal("_foTo")
	foTo, _ := L.ToNumber(-1)
	L.Pop(1)
	if foFrom != 1 || foTo != 0 {
		t.Fatalf("fadeOut: expected from=1, to=0, got from=%v, to=%v", foFrom, foTo)
	}

	globalAnimationManager.Clear()
}

func TestLua_UseAnimationHook(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalAnimationManager.Clear()

	// Override timeNowMs for deterministic testing
	origTime := timeNowMs
	defer func() { timeNowMs = origTime }()
	timeNowMs = func() int64 { return 0 }

	L := lua.NewState()
	Open(L)
	defer L.Close()

	// Create a component that uses useAnimation
	comp := &Component{
		ID:    "anim-comp",
		Type:  "AnimTest",
		Name:  "AnimTest",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[comp.ID] = comp
	SetCurrentComponent(comp)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, comp.ID)
	}()

	err := L.DoString(`
		local anim = lumina.useAnimation({
			from = 0,
			to = 100,
			duration = 1000,
			easing = "linear",
		})
		_value = anim.value
		_done = anim.done
	`)
	if err != nil {
		t.Fatalf("useAnimation error: %v", err)
	}

	L.GetGlobal("_value")
	val, _ := L.ToNumber(-1)
	L.Pop(1)
	L.GetGlobal("_done")
	done := L.ToBoolean(-1)
	L.Pop(1)

	// At start, value should be 0 (from)
	if val != 0 {
		t.Fatalf("expected initial value=0, got %v", val)
	}
	if done {
		t.Fatal("should not be done at start")
	}

	// Tick to halfway
	globalAnimationManager.Tick(500)

	// Re-render (simulate)
	comp.ResetHookIndex()
	err = L.DoString(`
		local anim = lumina.useAnimation({
			from = 0,
			to = 100,
			duration = 1000,
			easing = "linear",
		})
		_value2 = anim.value
		_done2 = anim.done
	`)
	if err != nil {
		t.Fatalf("useAnimation re-render error: %v", err)
	}

	L.GetGlobal("_value2")
	val2, _ := L.ToNumber(-1)
	L.Pop(1)
	if math.Abs(val2-50) > 1 {
		t.Fatalf("expected value≈50 at halfway, got %v", val2)
	}

	L.GetGlobal("_done2")
	if L.ToBoolean(-1) {
		t.Fatal("should not be done at halfway")
	}
	L.Pop(1)

	globalAnimationManager.Clear()
}

func TestLua_UseAnimationNoComponent(t *testing.T) {
	ClearComponents()
	globalAnimationManager.Clear()

	L := lua.NewState()
	Open(L)
	defer L.Close()

	// Call useAnimation outside a component context — should return static values
	SetCurrentComponent(nil)

	err := L.DoString(`
		local anim = lumina.useAnimation({ from = 0, to = 1, duration = 300 })
		_val = anim.value
		_dn = anim.done
	`)
	if err != nil {
		t.Fatalf("useAnimation outside component error: %v", err)
	}

	L.GetGlobal("_val")
	v, _ := L.ToNumber(-1)
	L.Pop(1)
	if v != 0 {
		t.Fatalf("expected value=0 outside component, got %v", v)
	}
	L.GetGlobal("_dn")
	if !L.ToBoolean(-1) {
		t.Fatal("expected done=true outside component")
	}
	L.Pop(1)

	globalAnimationManager.Clear()
}

func TestAnimationManager_EasingApplied(t *testing.T) {
	am := NewAnimationManager()

	am.Start(&AnimationState{
		ID: "ease", StartTime: 0, Duration: 1000, From: 0, To: 100, Easing: EaseIn,
	})

	// At t=0.5, EaseIn(0.5) = 0.25, so value should be 25
	am.Tick(500)
	got := am.Get("ease")
	if math.Abs(got.Current-25) > 1 {
		t.Fatalf("expected ~25 with EaseIn at halfway, got %v", got.Current)
	}
}
