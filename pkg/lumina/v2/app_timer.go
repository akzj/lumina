package v2

import (
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// timerEntry represents a single timer (interval or timeout).
type timerEntry struct {
	id       int
	ref      int   // Lua registry ref to callback function
	interval int64 // milliseconds
	nextFire int64 // unix milliseconds of next fire
	repeat   bool  // true = setInterval, false = setTimeout
	canceled bool
}

// timerManager manages setInterval/setTimeout timers.
// All methods are called from the main goroutine (event loop) — no locking needed.
type timerManager struct {
	timers map[int]*timerEntry
	nextID int
}

func newTimerManager() *timerManager {
	return &timerManager{
		timers: make(map[int]*timerEntry),
	}
}

// add registers a new timer and returns its ID.
func (tm *timerManager) add(ref int, ms int64, repeat bool) int {
	tm.nextID++
	id := tm.nextID
	tm.timers[id] = &timerEntry{
		id:       id,
		ref:      ref,
		interval: ms,
		nextFire: time.Now().UnixMilli() + ms,
		repeat:   repeat,
	}
	return id
}

// cancel marks a timer as canceled. It will be removed on the next fireDue call.
func (tm *timerManager) cancel(id int) {
	if t, ok := tm.timers[id]; ok {
		t.canceled = true
	}
}

// fireDue returns Lua registry refs of timers that are due to fire.
// One-shot timers are removed; repeating timers are rescheduled.
// Canceled timers are cleaned up. Returns refs of one-shot timers
// separately so the caller can Unref them after calling.
func (tm *timerManager) fireDue(now int64) (refs []int, oneshotRefs []int) {
	var toRemove []int
	for id, t := range tm.timers {
		if t.canceled {
			toRemove = append(toRemove, id)
			continue
		}
		if now >= t.nextFire {
			refs = append(refs, t.ref)
			if t.repeat {
				t.nextFire = now + t.interval
			} else {
				oneshotRefs = append(oneshotRefs, t.ref)
				toRemove = append(toRemove, id)
			}
		}
	}
	for _, id := range toRemove {
		delete(tm.timers, id)
	}
	return refs, oneshotRefs
}

// releaseAll releases all Lua refs and clears timers.
// Called during hot reload to prevent stale refs.
func (tm *timerManager) releaseAll(L *lua.State) {
	for _, t := range tm.timers {
		L.Unref(lua.RegistryIndex, t.ref)
	}
	tm.timers = make(map[int]*timerEntry)
	tm.nextID = 0
}

// count returns the number of active (non-canceled) timers.
func (tm *timerManager) count() int {
	n := 0
	for _, t := range tm.timers {
		if !t.canceled {
			n++
		}
	}
	return n
}

// fireTimers checks all timers and fires any that are due.
// Timer callbacks are invoked via Lua pcall on the main goroutine.
// One-shot timer refs are released after firing.
func (a *App) fireTimers() {
	if a.timerMgr == nil {
		return
	}
	now := time.Now().UnixMilli()
	refs, oneshotRefs := a.timerMgr.fireDue(now)
	L := a.luaState
	for _, ref := range refs {
		L.RawGetI(lua.RegistryIndex, int64(ref))
		if L.IsFunction(-1) {
			if status := L.PCall(0, 0, 0); status != lua.OK {
				L.Pop(1) // pop error message
			}
		} else {
			L.Pop(1)
		}
	}
	// Release one-shot timer refs to avoid memory leaks.
	for _, ref := range oneshotRefs {
		L.Unref(lua.RegistryIndex, ref)
	}
}

// --- Lua API implementations ---

// luaSetInterval implements lumina.setInterval(fn, ms).
// Returns the timer ID.
func (a *App) luaSetInterval(L *lua.State) int {
	L.CheckType(1, lua.TypeFunction)
	ms := L.CheckInteger(2)
	L.PushValue(1)
	ref := L.Ref(lua.RegistryIndex)
	id := a.timerMgr.add(ref, int64(ms), true)
	L.PushInteger(int64(id))
	return 1
}

// luaSetTimeout implements lumina.setTimeout(fn, ms).
// Returns the timer ID.
func (a *App) luaSetTimeout(L *lua.State) int {
	L.CheckType(1, lua.TypeFunction)
	ms := L.CheckInteger(2)
	L.PushValue(1)
	ref := L.Ref(lua.RegistryIndex)
	id := a.timerMgr.add(ref, int64(ms), false)
	L.PushInteger(int64(id))
	return 1
}

// luaClearTimer implements lumina.clearInterval(id) and lumina.clearTimeout(id).
// Both are the same function — cancels a timer by ID.
func (a *App) luaClearTimer(L *lua.State) int {
	id := int(L.CheckInteger(1))
	a.timerMgr.cancel(id)
	return 0
}
