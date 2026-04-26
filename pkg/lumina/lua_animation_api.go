package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)


// -----------------------------------------------------------------------
// Animation Lua API
// -----------------------------------------------------------------------

// luaStartAnimation starts a named animation imperatively.
// lumina.startAnimation({ id="fade", from=0, to=1, duration=300, easing="easeInOut", loop=false })
func luaStartAnimation(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("startAnimation: argument must be a table")
		L.Error()
		return 0
	}

	L.GetField(1, "id")
	id, _ := L.ToString(-1)
	L.Pop(1)
	if id == "" {
		L.PushString("startAnimation: 'id' is required")
		L.Error()
		return 0
	}

	from := 0.0
	to := 1.0
	duration := int64(300)
	easingName := "linear"
	loop := false

	L.GetField(1, "from")
	if !L.IsNoneOrNil(-1) {
		from, _ = L.ToNumber(-1)
	}
	L.Pop(1)

	L.GetField(1, "to")
	if !L.IsNoneOrNil(-1) {
		to, _ = L.ToNumber(-1)
	}
	L.Pop(1)

	L.GetField(1, "duration")
	if !L.IsNoneOrNil(-1) {
		d, _ := L.ToNumber(-1)
		duration = int64(d)
	}
	L.Pop(1)

	L.GetField(1, "easing")
	if !L.IsNoneOrNil(-1) {
		easingName, _ = L.ToString(-1)
	}
	L.Pop(1)

	L.GetField(1, "loop")
	if !L.IsNoneOrNil(-1) {
		loop = L.ToBoolean(-1)
	}
	L.Pop(1)

	anim := &AnimationState{
		ID:        id,
		StartTime: timeNowMs(),
		Duration:  duration,
		From:      from,
		To:        to,
		Current:   from,
		Easing:    easingByName(easingName),
		Loop:      loop,
	}
	globalAnimationManager.Start(anim)
	return 0
}

// luaStopAnimation stops an animation by ID.
// lumina.stopAnimation("fade")


// luaStopAnimation stops an animation by ID.
// lumina.stopAnimation("fade")
func luaStopAnimation(L *lua.State) int {
	id := L.CheckString(1)
	globalAnimationManager.Stop(id)
	return 0
}

// registerAnimationPresets creates the lumina.animation sub-table with preset factories.
// Each preset returns a config table suitable for useAnimation.


// registerAnimationPresets creates the lumina.animation sub-table with preset factories.
// Each preset returns a config table suitable for useAnimation.
func registerAnimationPresets(L *lua.State) {
	// lumina is at top of stack (-1) during luaLoader
	L.PushString("animation")
	L.NewTable()

	// lumina.animation.fadeIn(duration) → { from=0, to=1, duration=N, easing="easeInOut" }
	L.SetFuncs(map[string]lua.Function{
		"fadeIn": func(L *lua.State) int {
			dur := int64(300)
			if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
				d, _ := L.ToNumber(1)
				dur = int64(d)
			}
			L.NewTable()
			L.PushNumber(0)
			L.SetField(-2, "from")
			L.PushNumber(1)
			L.SetField(-2, "to")
			L.PushNumber(float64(dur))
			L.SetField(-2, "duration")
			L.PushString("easeInOut")
			L.SetField(-2, "easing")
			return 1
		},
		"fadeOut": func(L *lua.State) int {
			dur := int64(300)
			if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
				d, _ := L.ToNumber(1)
				dur = int64(d)
			}
			L.NewTable()
			L.PushNumber(1)
			L.SetField(-2, "from")
			L.PushNumber(0)
			L.SetField(-2, "to")
			L.PushNumber(float64(dur))
			L.SetField(-2, "duration")
			L.PushString("easeInOut")
			L.SetField(-2, "easing")
			return 1
		},
		"pulse": func(L *lua.State) int {
			dur := int64(1000)
			if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
				d, _ := L.ToNumber(1)
				dur = int64(d)
			}
			L.NewTable()
			L.PushNumber(0)
			L.SetField(-2, "from")
			L.PushNumber(1)
			L.SetField(-2, "to")
			L.PushNumber(float64(dur))
			L.SetField(-2, "duration")
			L.PushString("easeInOut")
			L.SetField(-2, "easing")
			L.PushBoolean(true)
			L.SetField(-2, "loop")
			return 1
		},
		"spin": func(L *lua.State) int {
			dur := int64(1000)
			if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
				d, _ := L.ToNumber(1)
				dur = int64(d)
			}
			L.NewTable()
			L.PushNumber(0)
			L.SetField(-2, "from")
			L.PushNumber(360)
			L.SetField(-2, "to")
			L.PushNumber(float64(dur))
			L.SetField(-2, "duration")
			L.PushString("linear")
			L.SetField(-2, "easing")
			L.PushBoolean(true)
			L.SetField(-2, "loop")
			return 1
		},
	}, 0)

	L.SetTable(-3) // lumina.animation = table
}

// -----------------------------------------------------------------------
// Hot Reload Lua API
// -----------------------------------------------------------------------

// luaEnableHotReload enables hot reload with optional config.
// lumina.enableHotReload({ paths = {"app.lua"}, interval = 500 })
