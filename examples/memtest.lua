-- examples/memtest.lua — Memory leak detection script
-- Run: lumina examples/memtest.lua
-- Expected: memory should stabilize after warmup, not grow indefinitely
--
-- Tests exercised:
--   1. Timer create/cancel cycles (timer ref leak regression)
--   2. Component mount/unmount (component tree cleanup)
--   3. Function prop churn (propFuncRef lifecycle)
--   4. DataGrid virtual scroll (event handler churn)
--   5. Theme switching (markAllDirty + full re-render)
--
-- Usage: lumina examples/memtest.lua
-- Quit:  q or Ctrl+Q (also auto-quits after test completes)

local ITERATIONS = 100      -- number of test cycles
local WARMUP = 10           -- warmup iterations (ignore for trend)
local REPORT_EVERY = 10     -- report every N iterations

-- Format bytes to human-readable
local function formatBytes(bytes)
    if bytes < 1024 then return tostring(bytes) .. " B" end
    if bytes < 1024*1024 then return string.format("%.1f KB", bytes/1024) end
    return string.format("%.1f MB", bytes/(1024*1024))
end

-- Memory samples
local samples = {}

-- Collect a memory sample (forces GC first)
local function sample(label)
    local stats = lumina.gc()
    samples[#samples + 1] = {
        label = label,
        goHeap = stats.goHeap,
        goObjects = stats.goObjects,
        luaBytes = stats.luaBytes,
    }
    return stats
end

-- Print memory report line
local function report(iter, stats)
    print(string.format("[iter %3d] Go Heap: %s | Go Objects: %d | Lua: %s",
        iter, formatBytes(stats.goHeap), stats.goObjects, formatBytes(stats.luaBytes)))
end

-- === TEST SCENARIOS ===

-- Scenario 1: Timer create + cancel (tests timer ref leak fix)
local function testTimerCancel()
    for i = 1, 10 do
        local id = lumina.setInterval(function() end, 1000)
        lumina.clearInterval(id)
    end
    for i = 1, 10 do
        local id = lumina.setTimeout(function() end, 1000)
        lumina.clearTimeout(id)
    end
end

-- Scenario 2: Component mount/unmount (tests component tree cleanup)
local function testComponentMountUnmount()
    local show = lumina.store.get("showComponent")
    lumina.store.set("showComponent", not show)
end

-- Scenario 3: Deep re-render with function props (tests propFuncRef lifecycle)
local function testPropFuncRefs()
    local c = lumina.store.get("counter") or 0
    lumina.store.set("counter", c + 1)
end

-- Scenario 4: DataGrid scroll (tests virtual scroll + event handler churn)
local function testDataGridScroll()
    local idx = lumina.store.get("gridIdx") or 1
    idx = idx + 1
    if idx > 50 then idx = 1 end
    lumina.store.set("gridIdx", idx)
end

-- Scenario 5: Theme switch (tests markAllDirty + full re-render)
local function testThemeSwitch()
    local themes = {"mocha", "latte", "nord", "dracula"}
    local current = lumina.store.get("themeIdx") or 1
    current = current % #themes + 1
    lumina.store.set("themeIdx", current)
    lumina.setTheme(themes[current])
end

-- === MAIN APP ===

lumina.app {
    id = "memtest",
    store = {
        showComponent = true,
        counter = 0,
        gridIdx = 1,
        themeIdx = 1,
        iteration = 0,
        phase = "warmup",
    },
    keys = {
        ["q"] = function() lumina.quit() end,
    },
    render = function()
        local lux = require("lux")
        local t = lumina.getTheme()
        local showComponent = lumina.useStore("showComponent")
        local counter = lumina.useStore("counter")
        local gridIdx = lumina.useStore("gridIdx")
        local iteration = lumina.useStore("iteration")
        local phase = lumina.useStore("phase")

        local children = {}

        -- Header
        children[#children + 1] = lumina.createElement("text", {
            key = "hdr",
            foreground = t.primary or "#89B4FA",
            bold = true,
            style = { height = 1 },
        }, "  Memory Leak Test -- " .. phase .. " [iter " .. tostring(iteration) .. "/" .. tostring(ITERATIONS) .. "]")

        children[#children + 1] = lumina.createElement("text", {
            key = "sp1", style = { height = 1 },
        }, "")

        -- Conditionally mounted component (tests mount/unmount cleanup)
        if showComponent then
            children[#children + 1] = lux.Alert {
                key = "alert-toggle",
                variant = "info",
                title = "Mounted #" .. tostring(counter),
                message = "This mounts/unmounts each cycle",
                width = 50,
            }
        end

        -- DataGrid with function props (tests propFuncRef churn)
        local rows = {}
        for i = 1, 50 do
            rows[i] = { name = "Row " .. tostring(i), val = tostring(i * counter) }
        end
        children[#children + 1] = lux.DataGrid {
            key = "grid",
            width = 50,
            height = 8,
            columns = {
                { id = "name", header = "Name", width = 25, key = "name" },
                { id = "val", header = "Value", width = 20, key = "val" },
            },
            rows = rows,
            selectedIndex = gridIdx,
            virtualScroll = true,
            onChangeIndex = function(i)
                lumina.store.set("gridIdx", i)
            end,
            onActivate = function(i, row) end,
        }

        -- Memory stats display
        local stats = lumina.memStats()
        children[#children + 1] = lumina.createElement("text", {
            key = "sp2", style = { height = 1 },
        }, "")
        children[#children + 1] = lumina.createElement("text", {
            key = "mem-hdr",
            foreground = t.secondary or "#CDD6F4",
            bold = true,
            style = { height = 1 },
        }, "  Live Memory:")
        children[#children + 1] = lumina.createElement("text", {
            key = "mem-go",
            foreground = t.text or "#CDD6F4",
            style = { height = 1 },
        }, "    Go Heap: " .. formatBytes(stats.goHeap) .. " | Objects: " .. tostring(stats.goObjects))
        children[#children + 1] = lumina.createElement("text", {
            key = "mem-lua",
            foreground = t.text or "#CDD6F4",
            style = { height = 1 },
        }, "    Lua: " .. formatBytes(stats.luaBytes))

        children[#children + 1] = lumina.createElement("text", {
            key = "sp3", style = { height = 1 },
        }, "")
        children[#children + 1] = lumina.createElement("text", {
            key = "footer",
            foreground = t.overlay0 or "#6C7086",
            style = { height = 1 },
        }, "  q to quit")

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 24, background = t.base or "#1E1E2E" },
        }, table.unpack(children))
    end,
}

-- === RUN TEST LOOP (async) ===

lumina.spawn(function()
    print("\n=== MEMORY LEAK TEST START ===\n")

    -- Initial sample
    local baseline = sample("baseline")
    report(0, baseline)

    for iter = 1, ITERATIONS do
        -- Run all test scenarios
        testTimerCancel()
        testComponentMountUnmount()
        testPropFuncRefs()
        testDataGridScroll()
        testThemeSwitch()

        -- Wait one frame for re-render
        lumina.sleep(16)

        -- Update iteration counter
        lumina.store.set("iteration", iter)

        if iter == WARMUP then
            lumina.store.set("phase", "testing")
        end

        -- Periodic report
        if iter % REPORT_EVERY == 0 then
            local stats = sample("iter-" .. tostring(iter))
            report(iter, stats)
        end
    end

    lumina.store.set("phase", "done")

    -- Final analysis
    print("\n=== MEMORY LEAK ANALYSIS ===\n")

    local postWarmup = nil
    local final = nil
    for i, s in ipairs(samples) do
        if s.label == "iter-" .. tostring(WARMUP) then
            postWarmup = s
        end
        final = s
    end

    if postWarmup and final then
        local goHeapDelta = final.goHeap - postWarmup.goHeap
        local luaDelta = final.luaBytes - postWarmup.luaBytes
        local objDelta = final.goObjects - postWarmup.goObjects

        print(string.format("Post-warmup Go Heap:   %s", formatBytes(postWarmup.goHeap)))
        print(string.format("Final Go Heap:         %s", formatBytes(final.goHeap)))
        print(string.format("Go Heap Delta:         %+d bytes", goHeapDelta))
        print("")
        print(string.format("Post-warmup Lua:       %s", formatBytes(postWarmup.luaBytes)))
        print(string.format("Final Lua:             %s", formatBytes(final.luaBytes)))
        print(string.format("Lua Delta:             %+d bytes", luaDelta))
        print("")
        print(string.format("Go Objects Delta:      %+d", objDelta))
        print("")

        -- Leak detection heuristic
        local goGrowthPct = (goHeapDelta / postWarmup.goHeap) * 100
        local luaGrowthPct = 0
        if postWarmup.luaBytes > 0 then
            luaGrowthPct = (luaDelta / postWarmup.luaBytes) * 100
        end

        print(string.format("Go Heap growth: %.1f%%", goGrowthPct))
        print(string.format("Lua memory growth: %.1f%%", luaGrowthPct))
        print("")

        if math.abs(goGrowthPct) < 10 and math.abs(luaGrowthPct) < 10 then
            print("PASS -- No significant memory leak detected")
        else
            print("WARNING -- Possible memory leak detected!")
            if goGrowthPct > 10 then
                print("   Go heap grew " .. string.format("%.1f%%", goGrowthPct) .. " after warmup")
            end
            if luaGrowthPct > 10 then
                print("   Lua memory grew " .. string.format("%.1f%%", luaGrowthPct) .. " after warmup")
            end
        end
    else
        print("ERROR: Not enough samples collected")
    end

    print("\n=== TEST COMPLETE ===\n")
    lumina.quit()
end)
