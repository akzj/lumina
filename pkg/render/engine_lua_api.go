package render

import (
	"runtime"
	"strings"

	"github.com/akzj/go-lua/pkg/lua"
)

// RegisterLuaAPI registers lumina.createElement, lumina.useState,
// lumina.defineComponent, lumina.createComponent on the Lua global table.
func (e *Engine) RegisterLuaAPI() {
	L := e.L

	// Create or get the "lumina" global table
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		L.Pop(1)
		L.NewTable()
	}
	tblIdx := L.AbsIndex(-1)

	// lumina.createElement(type, props, children...)
	L.PushFunction(e.luaCreateElement)
	L.SetField(tblIdx, "createElement")

	// lumina.defineComponent(name, renderFn) → factory table
	L.PushFunction(e.luaDefineComponent)
	L.SetField(tblIdx, "defineComponent")

	// lumina.createComponent(config) — root component
	L.PushFunction(e.luaCreateComponent)
	L.SetField(tblIdx, "createComponent")

	// lumina.useState(key, initial) → value, setter
	L.PushFunction(e.luaUseState)
	L.SetField(tblIdx, "useState")

	// lumina.useEffect(callback, deps?)
	L.PushFunction(e.luaUseEffect)
	L.SetField(tblIdx, "useEffect")

	// lumina.useRef(initialValue?)
	L.PushFunction(e.luaUseRef)
	L.SetField(tblIdx, "useRef")

	// lumina.useMemo(factory, deps)
	L.PushFunction(e.luaUseMemo)
	L.SetField(tblIdx, "useMemo")

	// lumina.useCallback(fn, deps)
	L.PushFunction(e.luaUseCallback)
	L.SetField(tblIdx, "useCallback")

	// lumina.spawn(fn) — start async coroutine
	L.PushFunction(e.luaSpawn)
	L.SetField(tblIdx, "spawn")

	// lumina.cancel(handle) — cancel a spawned coroutine
	L.PushFunction(e.luaCancel)
	L.SetField(tblIdx, "cancel")

	// lumina.sleep(ms) — returns Future
	L.PushFunction(e.luaSleep)
	L.SetField(tblIdx, "sleep")

	// lumina.exec(cmd) — returns Future
	L.PushFunction(e.luaExec)
	L.SetField(tblIdx, "exec")

	// lumina.readFile(path) — returns Future
	L.PushFunction(e.luaReadFile)
	L.SetField(tblIdx, "readFile")

	// lumina.fetch(url [, options]) — returns Future
	L.PushFunction(e.luaFetch)
	L.SetField(tblIdx, "fetch")

	// Create shared callable metatable for factory tables (__call → createElement)
	L.NewTable()
	L.PushFunction(e.luaFactoryCall)
	L.SetField(-2, "__call")
	sharedMetaIdx := L.AbsIndex(-1)

	// Register Go widgets as Lua-accessible factories (e.g., lumina.Button)
	for name := range e.widgets {
		L.NewTable()
		factoryIdx := L.AbsIndex(-1)
		L.PushBoolean(true)
		L.SetField(factoryIdx, "_isFactory")
		L.PushString(name)
		L.SetField(factoryIdx, "_name")
		// Set callable metatable so lumina.Button { props } works
		L.PushValue(sharedMetaIdx)
		L.SetMetatable(factoryIdx)
		L.SetField(tblIdx, name) // lumina.Button = {_isFactory=true, _name="Button"}
	}

	// Store shared metatable as registry ref for reuse by defineComponent
	e.factoryMetaRef = int64(L.Ref(lua.RegistryIndex)) // pops metatable

	// lumina.focusById(id) → boolean: programmatically focus a node by ID
	L.PushFunction(func(L *lua.State) int {
		id := L.CheckString(1)
		ok := e.FocusByID(id)
		L.PushBoolean(ok)
		return 1
	})
	L.SetField(tblIdx, "focusById")

	// lumina.getTheme() → returns theme color table
	L.PushFunction(e.luaGetTheme)
	L.SetField(tblIdx, "getTheme")

	// lumina.setTheme(nameOrTable) → switch theme by name or set custom
	L.PushFunction(e.luaSetTheme)
	L.SetField(tblIdx, "setTheme")

	// lumina.memStats() → {goHeap, goObjects, goGCCycles, goTotalAlloc, goSys, luaBytes}
	L.PushFunction(func(L *lua.State) int {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)

		L.NewTable()
		L.PushInteger(int64(ms.HeapAlloc))
		L.SetField(-2, "goHeap")
		L.PushInteger(int64(ms.HeapObjects))
		L.SetField(-2, "goObjects")
		L.PushInteger(int64(ms.NumGC))
		L.SetField(-2, "goGCCycles")
		L.PushInteger(int64(ms.TotalAlloc))
		L.SetField(-2, "goTotalAlloc")
		L.PushInteger(int64(ms.Sys))
		L.SetField(-2, "goSys")
		L.PushInteger(L.GCTotalBytes())
		L.SetField(-2, "luaBytes")
		return 1
	})
	L.SetField(tblIdx, "memStats")

	// lumina.gc() → force Go GC + Lua GC, return post-GC stats
	L.PushFunction(func(L *lua.State) int {
		L.GCCollect()
		runtime.GC()
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)

		L.NewTable()
		L.PushInteger(int64(ms.HeapAlloc))
		L.SetField(-2, "goHeap")
		L.PushInteger(int64(ms.HeapObjects))
		L.SetField(-2, "goObjects")
		L.PushInteger(int64(ms.NumGC))
		L.SetField(-2, "goGCCycles")
		L.PushInteger(L.GCTotalBytes())
		L.SetField(-2, "luaBytes")
		return 1
	})
	L.SetField(tblIdx, "gc")

	L.SetGlobal("lumina")
}

// luaGetTheme implements lumina.getTheme() → returns theme color table.
// Uses Engine.ThemeGetter to avoid import cycles (render cannot import widget).
func (e *Engine) luaGetTheme(L *lua.State) int {
	var theme map[string]string
	if e.customTheme != nil {
		theme = e.customTheme
	} else if e.ThemeGetter != nil {
		theme = e.ThemeGetter()
	}
	if theme == nil {
		L.NewTable()
		return 1
	}
	L.NewTable()
	for k, v := range theme {
		L.PushString(v)
		L.SetField(-2, k)
	}
	return 1
}

// luaSetTheme implements lumina.setTheme(nameOrTable).
// If a string is passed, switches to a built-in theme via ThemeSetter.
// If a table is passed, stores it as a custom theme override.
// After switching, marks all components dirty for re-render.
func (e *Engine) luaSetTheme(L *lua.State) int {
	if L.IsString(1) {
		name, _ := L.ToString(1)
		e.customTheme = nil
		if e.ThemeSetter != nil {
			e.ThemeSetter(name)
		}
	} else if L.IsTable(1) {
		custom := make(map[string]string)
		L.PushNil()
		for L.Next(1) {
			if L.IsString(-2) && L.IsString(-1) {
				key, _ := L.ToString(-2)
				val, _ := L.ToString(-1)
				custom[key] = val
			}
			L.Pop(1)
		}
		e.customTheme = custom
	}
	// Mark all components dirty so they re-render with the new theme
	e.markAllDirty()
	return 0
}

// markAllDirty marks every component as dirty so the next render cycle
// re-renders everything with updated theme colors.
func (e *Engine) markAllDirty() {
	for _, comp := range e.components {
		comp.Dirty = true
	}
	e.needsRender = true
}

// luaDefineComponent implements lumina.defineComponent(name, renderFn)
// Returns a factory table: {_isFactory=true, _name=name}
func (e *Engine) luaDefineComponent(L *lua.State) int {
	name := L.CheckString(1)
	L.CheckType(2, lua.TypeFunction)

	// Store render function as registry ref
	L.PushValue(2)
	ref := L.Ref(lua.RegistryIndex)
	// Free old factory ref if redefining (prevent Lua registry leak)
	if old, exists := e.factories[name]; exists && old != goWidgetSentinel {
		L.Unref(lua.RegistryIndex, int(old))
	}
	e.factories[name] = int64(ref)

	// Return a factory table that createElement can detect
	L.NewTable()
	resultIdx := L.AbsIndex(-1)
	L.PushBoolean(true)
	L.SetField(resultIdx, "_isFactory")
	L.PushString(name)
	L.SetField(resultIdx, "_name")

	// Set callable metatable so Factory { props } works
	e.setFactoryMetatable(L, resultIdx)

	return 1
}

// luaFactoryCall implements the __call metamethod for factory tables.
// Supports two calling patterns:
//
// Pattern 1 (standard): Factory(props, child1, child2)
//
//	__call receives: self, props, child1, child2
//	→ delegate to createElement(self, props, child1, child2)
//
// Pattern 2 (mixed single table): Factory { prop1=v1, Child1{}, Child2{} }
//
//	__call receives: self, mixedTable
//	→ split mixedTable: string keys → props, integer keys → children
//	→ rebuild stack as createElement(self, extractedProps, child1, child2, ...)
func (e *Engine) luaFactoryCall(L *lua.State) int {
	nArgs := L.GetTop()

	// Pattern 1: no args, or multiple args → standard delegation
	if nArgs != 2 || !L.IsTable(2) {
		return e.luaCreateElement(L)
	}

	// nArgs == 2, arg2 is a table. Check for integer keys (children).
	mixedIdx := 2
	childCount := int(L.LenI(mixedIdx))

	if childCount == 0 {
		// No integer keys → pure props table → standard delegation
		return e.luaCreateElement(L)
	}

	// Check if integer values are tables (descriptors/children) vs strings (content).
	// If first integer value is NOT a table, treat as standard (e.g., Text { "hello" }).
	L.RawGetI(mixedIdx, 1)
	firstIsTable := L.IsTable(-1)
	L.Pop(1)

	if !firstIsTable {
		// Integer values are strings/numbers → not mixed children pattern
		return e.luaCreateElement(L)
	}

	// === Pattern 2: mixed table ===
	// Split mixed table into props (string keys) and children (integer keys).
	// Build new stack: [1]=self, [2]=cleanProps, [3..]=children
	// without using registry refs (avoids go-lua Ref slot reuse bug).

	// Stack: [1]=self, [2]=mixed

	// Step 1: Build clean props table (string keys only) on top of stack
	L.NewTable() // [3] = newProps
	newPropsIdx := L.AbsIndex(-1)
	L.PushNil()
	for L.Next(mixedIdx) {
		// key at -2, value at -1
		if L.Type(-2) == lua.TypeString {
			L.PushValue(-2) // push key copy
			L.PushValue(-2) // push value copy
			L.SetTable(newPropsIdx)
		}
		L.Pop(1) // pop value, keep key for Next
	}
	// Stack: [1]=self, [2]=mixed, [3]=newProps

	// Step 2: Push children from mixed table
	for i := 1; i <= childCount; i++ {
		L.RawGetI(mixedIdx, int64(i)) // push child[i]
	}
	// Stack: [1]=self, [2]=mixed, [3]=newProps, [4..3+childCount]=children

	// Step 3: Remove mixed table at [2], shifting everything down
	L.Remove(2)
	// Stack: [1]=self, [2]=newProps, [3..2+childCount]=children

	// Stack is now: [1]=factory, [2]=props, [3..N]=children
	return e.luaCreateElement(L)
}

// setFactoryMetatable sets the shared __call metatable on a factory table at the given index.
// This enables the syntax: Factory { props } or Factory(props, child1, child2)
func (e *Engine) setFactoryMetatable(L *lua.State, tableIdx int) {
	absIdx := L.AbsIndex(tableIdx)
	L.RawGetI(lua.RegistryIndex, e.factoryMetaRef)
	L.SetMetatable(absIdx)
}

// luaCreateElement implements lumina.createElement(type_or_factory, props, children...)
func (e *Engine) luaCreateElement(L *lua.State) int {
	nArgs := L.GetTop()

	// Check if first arg is a factory table (from defineComponent)
	if L.IsTable(1) {
		L.GetField(1, "_isFactory")
		isFactory := L.ToBoolean(-1)
		L.Pop(1)

		if isFactory {
			return e.luaCreateComponentElement(L, nArgs)
		}
	}

	// Normal element: type is a string
	nodeType := L.CheckString(1)

	// Registered Go widgets must use the component placeholder shape so graftWalk,
	// hit-testing, and focus traversal see the widget's rendered RootNode.
	if _, isWidget := e.widgets[nodeType]; isWidget {
		L.NewTable()
		resultIdx := L.AbsIndex(-1)
		propsArg := 0
		firstChildArg := 2
		if nArgs >= 2 && L.IsTable(2) {
			propsArg = 2
			firstChildArg = 3
		}
		e.finishComponentPlaceholderDescriptor(L, resultIdx, nodeType, propsArg, firstChildArg, nArgs)
		return 1
	}

	// Create result table
	L.NewTable()
	resultIdx := L.AbsIndex(-1)

	L.PushString(nodeType)
	L.SetField(resultIdx, "type")

	// Copy props
	if nArgs >= 2 && L.IsTable(2) {
		L.ForEach(2, func(L *lua.State) bool {
			if L.Type(-2) == lua.TypeString {
				key, _ := L.ToString(-2)
				L.PushValue(-1)
				L.SetField(resultIdx, key)
			}
			return true
		})
	}

	// Handle children (args 3+)
	if nArgs > 2 {
		hasTable := false
		for i := 3; i <= nArgs; i++ {
			if L.Type(i) == lua.TypeTable {
				hasTable = true
				break
			}
		}

		if !hasTable {
			// String children → content
			var parts []string
			for i := 3; i <= nArgs; i++ {
				if L.Type(i) == lua.TypeString {
					s, _ := L.ToString(i)
					parts = append(parts, s)
				}
			}
			if len(parts) > 0 {
				L.PushString(strings.Join(parts, ""))
				L.SetField(resultIdx, "content")
			}
		} else {
			// Table children → children array
			L.CreateTable(nArgs-2, 0)
			childrenIdx := L.AbsIndex(-1)
			for i := 3; i <= nArgs; i++ {
				L.PushValue(i)
				L.RawSetI(childrenIdx, int64(i-2))
			}
			L.SetField(resultIdx, "children")
		}
	}

	return 1
}

// finishComponentPlaceholderDescriptor fills resultIdx as a component placeholder for
// a Lua-defined factory or a registered Go widget (same shape readDescriptor expects).
// propsArg is the stack index of the props table, or 0 if none. firstChildArg is the
// first stack index of optional varargs children (typically propsArg+1).
// When there is no props table, _props is still set to an empty Lua table so readDescriptor
// always sees a table (and later code injecting children never reads a nil _props).
func (e *Engine) finishComponentPlaceholderDescriptor(L *lua.State, resultIdx int, factoryName string, propsArg, firstChildArg, nArgs int) {
	L.PushString("component")
	L.SetField(resultIdx, "type")

	L.PushString(factoryName)
	L.SetField(resultIdx, "_factoryName")

	if propsArg > 0 && nArgs >= propsArg && L.IsTable(propsArg) {
		L.PushValue(propsArg)
		L.SetField(resultIdx, "_props")

		L.GetField(propsArg, "key")
		if L.IsString(-1) {
			s, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString(s)
			L.SetField(resultIdx, "key")
		} else {
			L.Pop(1)
		}

		L.GetField(propsArg, "id")
		if L.IsString(-1) {
			s, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString(s)
			L.SetField(resultIdx, "id")
		} else {
			L.Pop(1)
		}

		eventKeys := []string{
			"onClick", "onMouseEnter", "onMouseLeave", "onKeyDown",
			"onChange", "onScroll", "onMouseDown", "onMouseUp",
			"onFocus", "onBlur", "onSubmit", "onOutsideClick",
		}
		for _, key := range eventKeys {
			L.GetField(propsArg, key)
			if L.IsFunction(-1) {
				L.SetField(resultIdx, key)
			} else {
				L.Pop(1)
			}
		}

		L.GetField(propsArg, "disabled")
		if L.IsBoolean(-1) {
			L.SetField(resultIdx, "disabled")
		} else {
			L.Pop(1)
		}
		L.GetField(propsArg, "focusable")
		if L.IsBoolean(-1) {
			L.SetField(resultIdx, "focusable")
		} else {
			L.Pop(1)
		}
	} else {
		// No props arg: empty _props keeps the descriptor shape consistent for readDescriptor
		// and for the optional vararg-children block below (L.GetField(resultIdx, "_props")).
		L.NewTable()
		L.SetField(resultIdx, "_props")
	}

	if nArgs >= firstChildArg {
		childCount := 0
		for i := firstChildArg; i <= nArgs; i++ {
			if !L.IsNoneOrNil(i) {
				childCount++
			}
		}
		if childCount > 0 {
			L.GetField(resultIdx, "_props")
			if L.IsNil(-1) {
				L.Pop(1)
				L.NewTable()
				L.PushValue(-1)
				L.SetField(resultIdx, "_props")
			}
			propsIdx := L.AbsIndex(-1)

			L.CreateTable(childCount, 0)
			childrenIdx := L.AbsIndex(-1)
			idx := int64(1)
			for i := firstChildArg; i <= nArgs; i++ {
				if !L.IsNoneOrNil(i) {
					L.PushValue(i)
					L.RawSetI(childrenIdx, idx)
					idx++
				}
			}
			L.SetField(propsIdx, "children")
			L.Pop(1)
		}
	}
}

// luaCreateComponentElement handles createElement(Factory, props)
func (e *Engine) luaCreateComponentElement(L *lua.State, nArgs int) int {
	// Get factory name
	L.GetField(1, "_name")
	factoryName, _ := L.ToString(-1)
	L.Pop(1)

	L.NewTable()
	resultIdx := L.AbsIndex(-1)
	e.finishComponentPlaceholderDescriptor(L, resultIdx, factoryName, 2, 3, nArgs)
	return 1
}

// luaCreateComponent implements lumina.createComponent(config) for root components
func (e *Engine) luaCreateComponent(L *lua.State) int {
	L.CheckType(1, lua.TypeTable)
	absIdx := L.AbsIndex(1)

	id := getStringField(L, absIdx, "id")
	if id == "" {
		L.PushString("createComponent: 'id' is required")
		L.Error()
		return 0
	}

	name := getStringField(L, absIdx, "name")
	if name == "" {
		name = id
	}

	// Get render function ref
	L.GetField(absIdx, "render")
	if !L.IsFunction(-1) {
		L.Pop(1)
		L.PushString("createComponent: 'render' function is required")
		L.Error()
		return 0
	}
	ref := L.Ref(lua.RegistryIndex)

	e.CreateRootComponent(id, name, int64(ref))
	return 0
}

// luaUseState implements lumina.useState(key, initial) → value, setter
func (e *Engine) luaUseState(L *lua.State) int {
	comp := e.currentComp
	if comp == nil {
		L.PushString("useState: no current component")
		L.Error()
		return 0
	}

	key := L.CheckString(1)

	// Initialize if not exists
	if _, exists := comp.State[key]; !exists {
		var initial any
		if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
			initial = L.ToAny(2)
		}
		comp.State[key] = initial
	}

	// Push current value
	L.PushAny(comp.State[key])

	// Push setter function
	compID := comp.ID
	L.PushFunction(func(L *lua.State) int {
		newValue := L.ToAny(1)
		e.SetState(compID, key, newValue)
		return 0
	})

	return 2
}

// --- Helper functions for reading Lua tables ---

// propFuncRef is a Lua registry reference for a function stored in ComponentProps.
// Plain int64 in props would round-trip as a Lua number via PushAny; propFuncRef
// is restored in pushMap via RawGetI(registry, ref).
type propFuncRef int64
