package v2

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/output"
	"github.com/akzj/lumina/pkg/render"
)

// TestResult holds the outcome of a single Lua test case.
type TestResult struct {
	Suite    string
	Name     string
	Passed   bool
	Error    string
	Duration time.Duration
}

// TestRunner executes Lua test files and collects results.
type TestRunner struct {
	Width  int
	Height int
}

// NewTestRunner creates a test runner with default 80×24 screen.
func NewTestRunner() *TestRunner {
	return &TestRunner{Width: 80, Height: 24}
}

// RunFile executes a single Lua test file and returns results.
func (r *TestRunner) RunFile(path string) ([]TestResult, error) {
	L := lua.NewState()
	defer L.Close()

	// Register test framework APIs on this Lua state
	fw := newTestFramework(L, r.Width, r.Height)
	fw.register()

	// Execute the test file
	if err := L.DoFile(path); err != nil {
		return nil, fmt.Errorf("loading %s: %w", path, err)
	}

	// Run all collected test suites
	return fw.run(), nil
}

// RunDir executes all *_test.lua files under dir (recursively).
func (r *TestRunner) RunDir(dir string) ([]TestResult, error) {
	var paths []string
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.lua") {
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		return nil, err
	}

	// Keep execution order stable across filesystems.
	sort.Strings(paths)

	var allResults []TestResult
	for _, p := range paths {
		results, err := r.RunFile(p)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}
	return allResults, nil
}

// --- Internal test framework ---

// testSuite holds a describe() block with its test cases.
type testSuite struct {
	name          string
	tests         []testCase
	beforeEachRef int // Lua registry ref, 0 = none
	afterEachRef  int // Lua registry ref, 0 = none
}

// testCase holds a single it() test.
type testCase struct {
	name string
	fn   int // Lua registry ref to the test function
}

// testFramework manages the Lua test state.
type testFramework struct {
	L      *lua.State
	width  int
	height int
	suites []testSuite

	// Current suite being defined (set during describe() callback)
	currentSuite *testSuite
}

func newTestFramework(L *lua.State, w, h int) *testFramework {
	return &testFramework{
		L:      L,
		width:  w,
		height: h,
	}
}

// register installs the `test` global table with describe/it/beforeEach/afterEach/assert/createApp.
func (fw *testFramework) register() {
	L := fw.L

	L.NewTable()
	tblIdx := L.AbsIndex(-1)

	// test.describe(name, fn)
	L.PushFunction(fw.luaDescribe)
	L.SetField(tblIdx, "describe")

	// test.it(name, fn)
	L.PushFunction(fw.luaIt)
	L.SetField(tblIdx, "it")

	// test.beforeEach(fn)
	L.PushFunction(fw.luaBeforeEach)
	L.SetField(tblIdx, "beforeEach")

	// test.afterEach(fn)
	L.PushFunction(fw.luaAfterEach)
	L.SetField(tblIdx, "afterEach")

	// test.createApp(w, h) → app table
	L.PushFunction(fw.luaCreateApp)
	L.SetField(tblIdx, "createApp")

	// test.assert table
	fw.registerAssert(L, tblIdx)

	L.SetGlobal("test")
}

// registerAssert creates the test.assert sub-table.
func (fw *testFramework) registerAssert(L *lua.State, parentIdx int) {
	L.NewTable()
	assertIdx := L.AbsIndex(-1)

	// test.assert.eq(a, b)
	L.PushFunction(func(L *lua.State) int {
		a := L.ToAny(1)
		b := L.ToAny(2)
		if fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b) {
			L.PushString(fmt.Sprintf("assert.eq failed: %v ~= %v", a, b))
			L.Error()
		}
		return 0
	})
	L.SetField(assertIdx, "eq")

	// test.assert.neq(a, b)
	L.PushFunction(func(L *lua.State) int {
		a := L.ToAny(1)
		b := L.ToAny(2)
		if fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b) {
			L.PushString(fmt.Sprintf("assert.neq failed: both are %v", a))
			L.Error()
		}
		return 0
	})
	L.SetField(assertIdx, "neq")

	// test.assert.notNil(v)
	L.PushFunction(func(L *lua.State) int {
		if L.IsNoneOrNil(1) {
			L.PushString("assert.notNil failed: value is nil")
			L.Error()
		}
		return 0
	})
	L.SetField(assertIdx, "notNil")

	// test.assert.isNil(v)
	L.PushFunction(func(L *lua.State) int {
		if !L.IsNoneOrNil(1) {
			v := L.ToAny(1)
			L.PushString(fmt.Sprintf("assert.isNil failed: value is %v", v))
			L.Error()
		}
		return 0
	})
	L.SetField(assertIdx, "isNil")

	// test.assert.contains(str, sub)
	L.PushFunction(func(L *lua.State) int {
		str := L.CheckString(1)
		sub := L.CheckString(2)
		if !strings.Contains(str, sub) {
			L.PushString(fmt.Sprintf("assert.contains failed: %q does not contain %q", str, sub))
			L.Error()
		}
		return 0
	})
	L.SetField(assertIdx, "contains")

	// test.assert.gt(a, b)
	L.PushFunction(func(L *lua.State) int {
		a, _ := L.ToNumber(1)
		b, _ := L.ToNumber(2)
		if a <= b {
			L.PushString(fmt.Sprintf("assert.gt failed: %v <= %v", a, b))
			L.Error()
		}
		return 0
	})
	L.SetField(assertIdx, "gt")

	L.SetField(parentIdx, "assert")
}

// luaDescribe implements test.describe(name, fn)
func (fw *testFramework) luaDescribe(L *lua.State) int {
	name := L.CheckString(1)
	L.CheckType(2, lua.TypeFunction)

	suite := testSuite{name: name}
	fw.currentSuite = &suite

	// Call the describe function — it registers it() and beforeEach/afterEach
	L.PushValue(2)
	if status := L.PCall(0, 0, 0); status != lua.OK {
		errMsg, _ := L.ToString(-1)
		L.Pop(1)
		// Store a failing test for the describe block itself
		suite.tests = append(suite.tests, testCase{name: "describe setup: " + errMsg})
	}

	fw.suites = append(fw.suites, suite)
	fw.currentSuite = nil
	return 0
}

// luaIt implements test.it(name, fn)
func (fw *testFramework) luaIt(L *lua.State) int {
	name := L.CheckString(1)
	L.CheckType(2, lua.TypeFunction)

	if fw.currentSuite == nil {
		L.PushString("test.it() must be called inside test.describe()")
		L.Error()
		return 0
	}

	// Store function as registry ref
	L.PushValue(2)
	ref := L.Ref(lua.RegistryIndex)

	fw.currentSuite.tests = append(fw.currentSuite.tests, testCase{
		name: name,
		fn:   ref,
	})
	return 0
}

// luaBeforeEach implements test.beforeEach(fn)
func (fw *testFramework) luaBeforeEach(L *lua.State) int {
	L.CheckType(1, lua.TypeFunction)
	if fw.currentSuite == nil {
		L.PushString("test.beforeEach() must be called inside test.describe()")
		L.Error()
		return 0
	}
	L.PushValue(1)
	fw.currentSuite.beforeEachRef = L.Ref(lua.RegistryIndex)
	return 0
}

// luaAfterEach implements test.afterEach(fn)
func (fw *testFramework) luaAfterEach(L *lua.State) int {
	L.CheckType(1, lua.TypeFunction)
	if fw.currentSuite == nil {
		L.PushString("test.afterEach() must be called inside test.describe()")
		L.Error()
		return 0
	}
	L.PushValue(1)
	fw.currentSuite.afterEachRef = L.Ref(lua.RegistryIndex)
	return 0
}

// run executes all collected suites and returns results.
func (fw *testFramework) run() []TestResult {
	var results []TestResult
	L := fw.L

	for _, suite := range fw.suites {
		for _, tc := range suite.tests {
			start := time.Now()
			result := TestResult{
				Suite:  suite.name,
				Name:   tc.name,
				Passed: true,
			}

			if tc.fn == 0 {
				// Setup error — already recorded as name
				result.Passed = false
				result.Error = tc.name
				result.Duration = time.Since(start)
				results = append(results, result)
				continue
			}

			// Run beforeEach
			if suite.beforeEachRef != 0 {
				L.RawGetI(lua.RegistryIndex, int64(suite.beforeEachRef))
				if status := L.PCall(0, 0, 0); status != lua.OK {
					errMsg, _ := L.ToString(-1)
					L.Pop(1)
					result.Passed = false
					result.Error = "beforeEach: " + errMsg
					result.Duration = time.Since(start)
					results = append(results, result)
					continue
				}
			}

			// Run the test function
			L.RawGetI(lua.RegistryIndex, int64(tc.fn))
			if status := L.PCall(0, 0, 0); status != lua.OK {
				errMsg, _ := L.ToString(-1)
				L.Pop(1)
				result.Passed = false
				result.Error = errMsg
			}

			// Run afterEach
			if suite.afterEachRef != 0 {
				L.RawGetI(lua.RegistryIndex, int64(suite.afterEachRef))
				if status := L.PCall(0, 0, 0); status != lua.OK {
					L.Pop(1) // pop error, don't override test result
				}
			}

			result.Duration = time.Since(start)
			results = append(results, result)
		}
	}

	return results
}

// --- createApp ---

// testAppHandle holds the Go-side state for a test app created in Lua.
type testAppHandle struct {
	app    *App
	appL   *lua.State
	closed bool
}

// luaCreateApp implements test.createApp(w, h) → app table
func (fw *testFramework) luaCreateApp(L *lua.State) int {
	w := int(L.CheckInteger(1))
	h := int(L.CheckInteger(2))

	// Create a SEPARATE Lua state for the tested app
	appL := lua.NewState()
	ta := output.NewTestAdapter()
	app := NewApp(appL, w, h, ta)

	handle := &testAppHandle{
		app:  app,
		appL: appL,
	}

	// Create the app table with methods
	L.NewTable()
	appTblIdx := L.AbsIndex(-1)

	// All methods below use : syntax in Lua, so arg 1 = self (table), real args start at 2.

	// app:loadString(code)
	L.PushFunction(func(L *lua.State) int {
		code := L.CheckString(2)
		if err := handle.appL.DoString(code); err != nil {
			L.PushString(fmt.Sprintf("loadString error: %v", err))
			L.Error()
			return 0
		}
		handle.app.engine.RenderAll()
		return 0
	})
	L.SetField(appTblIdx, "loadString")

	// app:loadFile(path)
	L.PushFunction(func(L *lua.State) int {
		path := L.CheckString(2)
		if err := handle.app.RunScript(path); err != nil {
			L.PushString(fmt.Sprintf("loadFile error: %v", err))
			L.Error()
			return 0
		}
		handle.app.engine.RenderAll()
		return 0
	})
	L.SetField(appTblIdx, "loadFile")

	// app:click(x, y) or app:click("id")
	L.PushFunction(func(L *lua.State) int {
		if L.Type(2) == lua.TypeString {
			// app:click("id") — find node by ID, click center
			id := L.CheckString(2)
			vn := handle.app.engine.VNodeTree()
			found := findVNodeByID(vn, id)
			if found == nil {
				L.PushString(fmt.Sprintf("click: node with id %q not found", id))
				L.Error()
				return 0
			}
			handle.app.engine.HandleClick(found.X+found.W/2, found.Y+found.H/2)
		} else {
			x := int(L.CheckInteger(2))
			y := int(L.CheckInteger(3))
			handle.app.engine.HandleClick(x, y)
		}
		handle.app.engine.RenderDirty()
		return 0
	})
	L.SetField(appTblIdx, "click")

	// app:scroll(x, y, delta)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(2))
		y := int(L.CheckInteger(3))
		delta := int(L.CheckInteger(4))
		handle.app.engine.HandleScroll(x, y, delta)
		handle.app.engine.RenderDirty()
		return 0
	})
	L.SetField(appTblIdx, "scroll")

	// app:keyPress(key)
	L.PushFunction(func(L *lua.State) int {
		key := L.CheckString(2)
		// Check global keybindings first (mirrors handleInputEvent)
		if handle.app.handleGlobalKeys(key) {
			handle.app.engine.RenderDirty()
			return 0
		}
		handle.app.engine.HandleKeyDown(key)
		handle.app.engine.RenderDirty()
		return 0
	})
	L.SetField(appTblIdx, "keyPress")

	// app:resize(w, h)
	L.PushFunction(func(L *lua.State) int {
		newW := int(L.CheckInteger(2))
		newH := int(L.CheckInteger(3))
		handle.app.Resize(newW, newH)
		handle.app.engine.RenderAll()
		return 0
	})
	L.SetField(appTblIdx, "resize")

	// app:vnodeTree() → Lua table
	L.PushFunction(func(L *lua.State) int {
		vn := handle.app.engine.VNodeTree()
		pushVNodeToLua(L, vn)
		return 1
	})
	L.SetField(appTblIdx, "vnodeTree")

	// app:find(id) → Lua table or nil
	L.PushFunction(func(L *lua.State) int {
		id := L.CheckString(2)
		vn := handle.app.engine.VNodeTree()
		found := findVNodeByID(vn, id)
		pushVNodeToLua(L, found)
		return 1
	})
	L.SetField(appTblIdx, "find")

	// app:findAll(type) → Lua array of tables
	L.PushFunction(func(L *lua.State) int {
		nodeType := L.CheckString(2)
		vn := handle.app.engine.VNodeTree()
		var matches []*render.VNode
		collectVNodesByType(vn, nodeType, &matches)
		L.NewTable()
		arrIdx := L.AbsIndex(-1)
		for i, m := range matches {
			pushVNodeToLua(L, m)
			L.RawSetI(arrIdx, int64(i+1))
		}
		return 1
	})
	L.SetField(appTblIdx, "findAll")

	// app:render() — trigger a RenderDirty cycle (useful after effects set state)
	L.PushFunction(func(L *lua.State) int {
		handle.app.engine.RenderDirty()
		return 0
	})
	L.SetField(appTblIdx, "render")

	// app:tick() — tick the async scheduler (resume completed coroutines)
	L.PushFunction(func(L *lua.State) int {
		if handle.app.scheduler != nil {
			handle.app.scheduler.Tick()
		}
		return 0
	})
	L.SetField(appTblIdx, "tick")

	// app:waitAsync(timeoutMs) — wait for all async coroutines to complete
	// Polls scheduler.Tick() in a loop until all pending coroutines finish.
	// Returns true on success, false on timeout.
	L.PushFunction(func(L *lua.State) int {
		timeoutMs := int64(100) // default 100ms
		if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
			timeoutMs = L.CheckInteger(2)
		}
		if handle.app.scheduler == nil {
			L.PushBoolean(true)
			return 1
		}
		err := handle.app.scheduler.WaitAll(time.Duration(timeoutMs) * time.Millisecond)
		L.PushBoolean(err == nil)
		return 1
	})
	L.SetField(appTblIdx, "waitAsync")

	// app:screenText() → string
	L.PushFunction(func(L *lua.State) int {
		buf := handle.app.Screen()
		if buf == nil {
			L.PushString("")
			return 1
		}
		var sb strings.Builder
		sb.Grow(buf.Width()*buf.Height() + buf.Height())
		for y := 0; y < buf.Height(); y++ {
			for x := 0; x < buf.Width(); x++ {
				cell := buf.Get(x, y)
				if cell.Char == 0 {
					sb.WriteRune(' ')
				} else {
					sb.WriteRune(cell.Char)
				}
			}
			sb.WriteRune('\n')
		}
		L.PushString(sb.String())
		return 1
	})
	L.SetField(appTblIdx, "screenText")

	// app:screenContains(text) → bool
	L.PushFunction(func(L *lua.State) int {
		text := L.CheckString(2)
		screen := handle.app.MCPGetScreenText()
		L.PushBoolean(strings.Contains(screen, text))
		return 1
	})
	L.SetField(appTblIdx, "screenContains")

	// app:cellAt(x, y) → {char, fg, bg, bold, dim, underline}
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(2))
		y := int(L.CheckInteger(3))
		buf := handle.app.Screen()
		if buf == nil || x < 0 || y < 0 || x >= buf.Width() || y >= buf.Height() {
			L.PushNil()
			return 1
		}
		cell := buf.Get(x, y)
		L.NewTable()
		tbl := L.AbsIndex(-1)
		if cell.Char == 0 {
			L.PushString(" ")
		} else {
			L.PushString(string(cell.Char))
		}
		L.SetField(tbl, "char")
		L.PushString(cell.Foreground)
		L.SetField(tbl, "fg")
		L.PushString(cell.Background)
		L.SetField(tbl, "bg")
		L.PushBoolean(cell.Bold)
		L.SetField(tbl, "bold")
		L.PushBoolean(cell.Dim)
		L.SetField(tbl, "dim")
		L.PushBoolean(cell.Underline)
		L.SetField(tbl, "underline")
		return 1
	})
	L.SetField(appTblIdx, "cellAt")

	// app:getState(compID, key) → value
	L.PushFunction(func(L *lua.State) int {
		compID := L.CheckString(2)
		key := L.CheckString(3)
		val, err := handle.app.MCPGetState(compID, key)
		if err != nil {
			L.PushNil()
			return 1
		}
		pushAnyToLua(L, val)
		return 1
	})
	L.SetField(appTblIdx, "getState")

	// app:setState(compID, key, value)
	L.PushFunction(func(L *lua.State) int {
		compID := L.CheckString(2)
		key := L.CheckString(3)
		value := L.ToAny(4)
		handle.app.MCPSetState(compID, key, value)
		handle.app.engine.RenderDirty()
		return 0
	})
	L.SetField(appTblIdx, "setState")

	// app:focusedID() → string or nil
	L.PushFunction(func(L *lua.State) int {
		id := handle.app.MCPGetFocusedID()
		if id == "" {
			L.PushNil()
		} else {
			L.PushString(id)
		}
		return 1
	})
	L.SetField(appTblIdx, "focusedID")

	// app:type(text) — sends each character as a keyPress event
	L.PushFunction(func(L *lua.State) int {
		text := L.CheckString(2)
		for _, ch := range text {
			handle.app.engine.HandleKeyDown(string(ch))
		}
		handle.app.engine.RenderDirty()
		return 0
	})
	L.SetField(appTblIdx, "type")

	// app:snapshot() → {screen, components, focused, focusable}
	L.PushFunction(func(L *lua.State) int {
		L.NewTable()
		tbl := L.AbsIndex(-1)

		// screen text
		L.PushString(handle.app.MCPGetScreenText())
		L.SetField(tbl, "screen")

		// focused ID
		L.PushString(handle.app.MCPGetFocusedID())
		L.SetField(tbl, "focused")

		// focusable IDs
		ids := handle.app.MCPGetFocusableIDs()
		L.NewTable()
		idsIdx := L.AbsIndex(-1)
		for i, id := range ids {
			L.PushString(id)
			L.RawSetI(idsIdx, int64(i+1))
		}
		L.SetField(tbl, "focusable")

		// components (basic info)
		comps := handle.app.MCPInspectTree()
		L.NewTable()
		compsIdx := L.AbsIndex(-1)
		for i, c := range comps {
			L.NewTable()
			cIdx := L.AbsIndex(-1)
			L.PushString(c.ID)
			L.SetField(cIdx, "id")
			L.PushString(c.Name)
			L.SetField(cIdx, "name")
			L.PushBoolean(c.Focused)
			L.SetField(cIdx, "focused")
			L.RawSetI(compsIdx, int64(i+1))
		}
		L.SetField(tbl, "components")

		return 1
	})
	L.SetField(appTblIdx, "snapshot")

	// app:destroy()
	L.PushFunction(func(L *lua.State) int {
		if !handle.closed {
			handle.appL.Close()
			handle.closed = true
		}
		return 0
	})
	L.SetField(appTblIdx, "destroy")

	return 1
}

// --- VNode helpers ---

// findVNodeByID searches a VNode tree by ID (DFS).
func findVNodeByID(vn *render.VNode, id string) *render.VNode {
	if vn == nil {
		return nil
	}
	if vn.ID == id {
		return vn
	}
	for _, child := range vn.Children {
		if found := findVNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// collectVNodesByType collects all VNodes matching a type.
func collectVNodesByType(vn *render.VNode, nodeType string, out *[]*render.VNode) {
	if vn == nil {
		return
	}
	if vn.Type == nodeType {
		*out = append(*out, vn)
	}
	for _, child := range vn.Children {
		collectVNodesByType(child, nodeType, out)
	}
}

// pushVNodeToLua converts a VNode to a Lua table and pushes it.
func pushVNodeToLua(L *lua.State, vn *render.VNode) {
	if vn == nil {
		L.PushNil()
		return
	}

	L.NewTable()
	tbl := L.AbsIndex(-1)

	L.PushString(vn.Type)
	L.SetField(tbl, "type")

	if vn.ID != "" {
		L.PushString(vn.ID)
		L.SetField(tbl, "id")
	}
	if vn.Key != "" {
		L.PushString(vn.Key)
		L.SetField(tbl, "key")
	}
	if vn.Content != "" {
		L.PushString(vn.Content)
		L.SetField(tbl, "content")
	}

	// Layout
	L.PushInteger(int64(vn.X))
	L.SetField(tbl, "x")
	L.PushInteger(int64(vn.Y))
	L.SetField(tbl, "y")
	L.PushInteger(int64(vn.W))
	L.SetField(tbl, "w")
	L.PushInteger(int64(vn.H))
	L.SetField(tbl, "h")

	// Scroll
	if vn.ScrollY != 0 {
		L.PushInteger(int64(vn.ScrollY))
		L.SetField(tbl, "scrollY")
	}
	if vn.ScrollHeight != 0 {
		L.PushInteger(int64(vn.ScrollHeight))
		L.SetField(tbl, "scrollHeight")
	}

	// Style as sub-table
	if vn.Style != nil {
		L.NewTable()
		styleTbl := L.AbsIndex(-1)
		for k, v := range vn.Style {
			switch val := v.(type) {
			case int:
				L.PushInteger(int64(val))
				L.SetField(styleTbl, k)
			case string:
				L.PushString(val)
				L.SetField(styleTbl, k)
			case bool:
				L.PushBoolean(val)
				L.SetField(styleTbl, k)
			}
		}
		L.SetField(tbl, "style")
	}

	// Children
	if len(vn.Children) > 0 {
		L.NewTable()
		childrenTbl := L.AbsIndex(-1)
		for i, child := range vn.Children {
			pushVNodeToLua(L, child)
			L.RawSetI(childrenTbl, int64(i+1))
		}
		L.SetField(tbl, "children")
	}
}

// pushAnyToLua converts an arbitrary Go value to a Lua value on the stack.
func pushAnyToLua(L *lua.State, v any) {
	switch val := v.(type) {
	case nil:
		L.PushNil()
	case bool:
		L.PushBoolean(val)
	case int:
		L.PushInteger(int64(val))
	case int64:
		L.PushInteger(val)
	case float64:
		L.PushNumber(val)
	case string:
		L.PushString(val)
	case map[string]any:
		L.NewTable()
		tbl := L.AbsIndex(-1)
		for k, inner := range val {
			pushAnyToLua(L, inner)
			L.SetField(tbl, k)
		}
	case []any:
		L.NewTable()
		tbl := L.AbsIndex(-1)
		for i, inner := range val {
			pushAnyToLua(L, inner)
			L.RawSetI(tbl, int64(i+1))
		}
	default:
		L.PushString(fmt.Sprintf("%v", val))
	}
}
