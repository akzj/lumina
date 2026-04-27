package bridge

import (
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/animation"
)

// luaUseAnimation implements lumina.useAnimation(config).
// Config is a Lua table with fields: id, from, to, duration, easing, loop.
// Returns a table: { value = currentValue, start = fn(), stop = fn() }
func (b *Bridge) luaUseAnimation(L *lua.State) int {
	if b.animManager == nil {
		L.PushString("useAnimation: no animation manager configured")
		L.Error()
		return 0
	}

	L.CheckType(1, lua.TypeTable)
	cfg := readAnimConfig(L, 1)

	// Check if animation is already running.
	anim := b.animManager.Get(cfg.ID)
	var current float64
	if anim != nil {
		current = anim.Current()
	} else {
		current = cfg.From
	}

	// Build result table.
	L.NewTable()
	resultIdx := L.AbsIndex(-1)

	L.PushNumber(current)
	L.SetField(resultIdx, "value")

	// start function
	mgr := b.animManager
	animCfg := cfg
	L.PushFunction(func(L *lua.State) int {
		mgr.Start(animCfg, timeNowMs())
		return 0
	})
	L.SetField(resultIdx, "start")

	// stop function
	animID := cfg.ID
	L.PushFunction(func(L *lua.State) int {
		mgr.Stop(animID)
		return 0
	})
	L.SetField(resultIdx, "stop")

	return 1
}

// readAnimConfig reads an animation.Config from a Lua table at the given index.
func readAnimConfig(L *lua.State, idx int) animation.Config {
	absIdx := L.AbsIndex(idx)

	cfg := animation.Config{
		ID:       L.GetFieldString(absIdx, "id"),
		From:     L.GetFieldNumber(absIdx, "from"),
		To:       L.GetFieldNumber(absIdx, "to"),
		Duration: L.GetFieldInt(absIdx, "duration"),
		Easing:   L.GetFieldString(absIdx, "easing"),
		Loop:     L.GetFieldBool(absIdx, "loop"),
	}

	if cfg.ID == "" {
		cfg.ID = "default"
	}

	return cfg
}

// timeNowMs returns the current time in milliseconds.
func timeNowMs() int64 {
	return time.Now().UnixMilli()
}
