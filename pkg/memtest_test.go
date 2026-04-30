package v2

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/output"
)

// TestMemoryLeak_TimerCancelCycle exercises the timer create/cancel pattern
// in a tight loop and verifies no memory growth (regression for timer ref leak).
func TestMemoryLeak_TimerCancelCycle(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 60, 24, ta)
	defer app.Stop()

	// Setup a simple component
	err := app.RunString(`
		lumina.createComponent({
			id = "memtest-timer",
			name = "MemTestTimer",
			render = function(props)
				return lumina.createElement("text", {}, "timer test")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	// Define the timer test function once (avoids DoString compilation noise)
	err = app.RunString(`
		function _timerTest()
			local id = lumina.setInterval(function() end, 1000)
			lumina.clearInterval(id)
			local id2 = lumina.setTimeout(function() end, 500)
			lumina.clearTimeout(id2)
		end
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Warmup: run 50 iterations to stabilize
	for i := 0; i < 50; i++ {
		err = app.RunString(`_timerTest()`)
		if err != nil {
			t.Fatal(err)
		}
		// fireTimers processes canceled entries and Unrefs their callbacks
		app.fireTimers()
	}

	// Force GC and measure baseline
	L.GCCollect()
	runtime.GC()
	runtime.GC() // double GC for finalizers
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)
	luaBaseline := L.GCTotalBytes()

	// Run 500 iterations
	for i := 0; i < 500; i++ {
		err = app.RunString(`_timerTest()`)
		if err != nil {
			t.Fatal(err)
		}
		// fireTimers processes canceled entries and Unrefs their callbacks
		app.fireTimers()
	}

	// Force GC and measure final
	L.GCCollect()
	runtime.GC()
	runtime.GC()
	var final runtime.MemStats
	runtime.ReadMemStats(&final)
	luaFinal := L.GCTotalBytes()

	// Analyze
	goHeapDelta := int64(final.HeapAlloc) - int64(baseline.HeapAlloc)
	goGrowthPct := float64(goHeapDelta) / float64(baseline.HeapAlloc) * 100

	luaDelta := luaFinal - luaBaseline
	luaGrowthPct := float64(0)
	if luaBaseline > 0 {
		luaGrowthPct = float64(luaDelta) / float64(luaBaseline) * 100
	}

	t.Logf("Go Heap: baseline=%d final=%d delta=%+d (%.1f%%)",
		baseline.HeapAlloc, final.HeapAlloc, goHeapDelta, goGrowthPct)
	t.Logf("Lua Memory: baseline=%d final=%d delta=%+d (%.1f%%)",
		luaBaseline, luaFinal, luaDelta, luaGrowthPct)

	if goGrowthPct > 50 {
		t.Errorf("Possible Go heap leak: grew %.1f%% after 500 timer create/cancel cycles", goGrowthPct)
	}
	if luaGrowthPct > 50 {
		t.Errorf("Possible Lua memory leak: grew %.1f%% after 500 timer create/cancel cycles", luaGrowthPct)
	}
}

// TestMemoryLeak_ComponentMountUnmount exercises component mount/unmount cycles
// and verifies no memory growth (regression for component tree cleanup).
func TestMemoryLeak_ComponentMountUnmount(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 60, 24, ta)
	defer app.Stop()

	// Setup a component with conditional rendering
	err := app.RunString(`
		lumina.createComponent({
			id = "memtest-mount",
			name = "MemTestMount",
			store = { show = true, counter = 0 },
			render = function(props)
				local show = lumina.useStore("show")
				local counter = lumina.useStore("counter")
				local children = {}
				if show then
					children[#children + 1] = lumina.createElement("box", {
						key = "child",
						style = { width = 40, height = 3 },
					},
						lumina.createElement("text", {key="t1"}, "Mounted " .. tostring(counter)),
						lumina.createElement("text", {key="t2"}, "Line 2"),
						lumina.createElement("text", {key="t3"}, "Line 3")
					)
				end
				return lumina.createElement("vbox", {
					id = "root",
					style = { width = 60, height = 24 },
				}, table.unpack(children))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	// Warmup
	for i := 0; i < 50; i++ {
		err = app.RunString(fmt.Sprintf(`
			lumina.store.set("show", %v)
			lumina.store.set("counter", %d)
		`, i%2 == 0, i))
		if err != nil {
			t.Fatal(err)
		}
		app.RenderDirty()
	}

	// Force GC and measure baseline
	L.GCCollect()
	runtime.GC()
	runtime.GC()
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)
	luaBaseline := L.GCTotalBytes()

	// Run 300 mount/unmount cycles
	for i := 0; i < 300; i++ {
		err = app.RunString(fmt.Sprintf(`
			lumina.store.set("show", %v)
			lumina.store.set("counter", %d)
		`, i%2 == 0, i+50))
		if err != nil {
			t.Fatal(err)
		}
		app.RenderDirty()
	}

	// Force GC and measure final
	L.GCCollect()
	runtime.GC()
	runtime.GC()
	var final runtime.MemStats
	runtime.ReadMemStats(&final)
	luaFinal := L.GCTotalBytes()

	// Analyze
	goHeapDelta := int64(final.HeapAlloc) - int64(baseline.HeapAlloc)
	goGrowthPct := float64(goHeapDelta) / float64(baseline.HeapAlloc) * 100

	luaDelta := luaFinal - luaBaseline
	luaGrowthPct := float64(0)
	if luaBaseline > 0 {
		luaGrowthPct = float64(luaDelta) / float64(luaBaseline) * 100
	}

	t.Logf("Go Heap: baseline=%d final=%d delta=%+d (%.1f%%)",
		baseline.HeapAlloc, final.HeapAlloc, goHeapDelta, goGrowthPct)
	t.Logf("Lua Memory: baseline=%d final=%d delta=%+d (%.1f%%)",
		luaBaseline, luaFinal, luaDelta, luaGrowthPct)

	if goGrowthPct > 50 {
		t.Errorf("Possible Go heap leak: grew %.1f%% after 300 mount/unmount cycles", goGrowthPct)
	}
	if luaGrowthPct > 50 {
		t.Errorf("Possible Lua memory leak: grew %.1f%% after 300 mount/unmount cycles", luaGrowthPct)
	}
}

// TestMemoryLeak_PropFuncRefChurn exercises function prop re-renders
// and verifies no memory growth (regression for propFuncRef leak).
func TestMemoryLeak_PropFuncRefChurn(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 60, 24, ta)
	defer app.Stop()

	// Setup a component with function props that change every render
	err := app.RunString(`
		lumina.createComponent({
			id = "memtest-funcref",
			name = "MemTestFuncRef",
			store = { counter = 0 },
			render = function(props)
				local counter = lumina.useStore("counter")
				return lumina.createElement("box", {
					id = "root",
					style = { width = 60, height = 24 },
					onClick = function() lumina.store.set("counter", counter + 1) end,
					onMouseEnter = function() end,
					onMouseLeave = function() end,
				},
					lumina.createElement("text", {key="t"}, "Count: " .. tostring(counter))
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	// Warmup
	for i := 0; i < 50; i++ {
		err = app.RunString(fmt.Sprintf(`lumina.store.set("counter", %d)`, i))
		if err != nil {
			t.Fatal(err)
		}
		app.RenderDirty()
	}

	// Force GC and measure baseline
	L.GCCollect()
	runtime.GC()
	runtime.GC()
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)
	luaBaseline := L.GCTotalBytes()

	// Run 500 re-renders with function prop churn
	for i := 0; i < 500; i++ {
		err = app.RunString(fmt.Sprintf(`lumina.store.set("counter", %d)`, i+50))
		if err != nil {
			t.Fatal(err)
		}
		app.RenderDirty()
	}

	// Force GC and measure final
	L.GCCollect()
	runtime.GC()
	runtime.GC()
	var final runtime.MemStats
	runtime.ReadMemStats(&final)
	luaFinal := L.GCTotalBytes()

	// Analyze
	goHeapDelta := int64(final.HeapAlloc) - int64(baseline.HeapAlloc)
	goGrowthPct := float64(goHeapDelta) / float64(baseline.HeapAlloc) * 100

	luaDelta := luaFinal - luaBaseline
	luaGrowthPct := float64(0)
	if luaBaseline > 0 {
		luaGrowthPct = float64(luaDelta) / float64(luaBaseline) * 100
	}

	t.Logf("Go Heap: baseline=%d final=%d delta=%+d (%.1f%%)",
		baseline.HeapAlloc, final.HeapAlloc, goHeapDelta, goGrowthPct)
	t.Logf("Lua Memory: baseline=%d final=%d delta=%+d (%.1f%%)",
		luaBaseline, luaFinal, luaDelta, luaGrowthPct)

	if goGrowthPct > 50 {
		t.Errorf("Possible Go heap leak: grew %.1f%% after 500 func prop churn cycles", goGrowthPct)
	}
	if luaGrowthPct > 50 {
		t.Errorf("Possible Lua memory leak: grew %.1f%% after 500 func prop churn cycles", luaGrowthPct)
	}
}

// TestMemoryLeak_LuaAPIs verifies that lumina.memStats() and lumina.gc() work correctly.
func TestMemoryLeak_LuaAPIs(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 60, 24, ta)
	defer app.Stop()

	// Setup minimal component
	err := app.RunString(`
		lumina.createComponent({
			id = "memtest-api",
			name = "MemTestAPI",
			render = function(props)
				return lumina.createElement("text", {}, "api test")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	// Test lumina.memStats()
	err = app.RunString(`
		local stats = lumina.memStats()
		assert(type(stats) == "table", "memStats should return a table")
		assert(type(stats.goHeap) == "number", "goHeap should be a number")
		assert(type(stats.goObjects) == "number", "goObjects should be a number")
		assert(type(stats.goGCCycles) == "number", "goGCCycles should be a number")
		assert(type(stats.goTotalAlloc) == "number", "goTotalAlloc should be a number")
		assert(type(stats.goSys) == "number", "goSys should be a number")
		assert(type(stats.luaBytes) == "number", "luaBytes should be a number")
		assert(stats.goHeap > 0, "goHeap should be > 0")
		assert(stats.luaBytes > 0, "luaBytes should be > 0")
	`)
	if err != nil {
		t.Fatalf("lumina.memStats() failed: %v", err)
	}

	// Test lumina.gc()
	err = app.RunString(`
		local stats = lumina.gc()
		assert(type(stats) == "table", "gc should return a table")
		assert(type(stats.goHeap) == "number", "goHeap should be a number")
		assert(type(stats.goObjects) == "number", "goObjects should be a number")
		assert(type(stats.goGCCycles) == "number", "goGCCycles should be a number")
		assert(type(stats.luaBytes) == "number", "luaBytes should be a number")
		assert(stats.goHeap > 0, "goHeap should be > 0")
	`)
	if err != nil {
		t.Fatalf("lumina.gc() failed: %v", err)
	}

	// Test that gc() actually collects: allocate garbage, then gc, compare
	err = app.RunString(`
		-- Create some garbage
		local garbage = {}
		for i = 1, 10000 do
			garbage[i] = { x = i, y = tostring(i) }
		end
		local before = lumina.memStats()
		garbage = nil  -- make it eligible for collection
		local after = lumina.gc()
		-- Lua bytes should decrease (or at least not grow significantly)
		-- We can't guarantee exact behavior, but gc() should not error
		assert(after.luaBytes < before.luaBytes, 
			"expected Lua bytes to decrease after gc: before=" .. tostring(before.luaBytes) .. " after=" .. tostring(after.luaBytes))
	`)
	if err != nil {
		t.Fatalf("lumina.gc() collection verification failed: %v", err)
	}
}
