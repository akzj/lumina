package v2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/render"
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

// RunDir executes all *_test.lua files in a directory.
func (r *TestRunner) RunDir(dir string) ([]TestResult, error) {
	var allResults []TestResult
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.lua") {
			continue
		}
		results, err := r.RunFile(filepath.Join(dir, entry.Name()))
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
