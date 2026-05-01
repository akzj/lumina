-- examples/windows.lua — Multi-Window Manager (keyboard-only)
--
-- Demonstrates overlapping windows with z-order management:
--   • Absolute positioning for window placement
--   • Z-order via child ordering (last child = on top)
--   • Click to bring window to front
--   • Keyboard movement of active window
--   • Per-window click counters
--
-- Pure Lua window management (no Go Window widget needed)
-- with mouse drag-to-move and drag-to-resize support.
--
-- Features showcased:
--   • position = "absolute" with left/top/width/height
--   • onClick handlers on windows and buttons
--   • lumina.store for state management
--   • Global keybindings (1/2/3=select, arrows=move, q=quit)
--
-- Usage: lumina examples/windows.lua

local initialWindows = {
    {
        id = "win1",
        title = "📝 Editor",
        x = 2, y = 1, w = 30, h = 12,
        clicks = 0,
        content = "Welcome to the editor window.\nType your code here.\nLine 3 of content.",
    },
    {
        id = "win2",
        title = "📊 Monitor",
        x = 15, y = 5, w = 30, h = 12,
        clicks = 0,
        content = "CPU: 42%  MEM: 1.2GB\nProcesses: 128\nUptime: 3d 14h",
    },
    {
        id = "win3",
        title = "🎨 Palette",
        x = 28, y = 3, w = 30, h = 12,
        clicks = 0,
        content = "Colors: Red, Green, Blue\nBrush: Round 3px\nOpacity: 80%",
    },
}

-- Helper: bring a window to the front of the order
local function bringToFront(winIdx)
    local order = lumina.store.get("windowOrder")
    local newOrder = {}
    for _, idx in ipairs(order) do
        if idx ~= winIdx then
            newOrder[#newOrder + 1] = idx
        end
    end
    newOrder[#newOrder + 1] = winIdx
    lumina.store.set("windowOrder", newOrder)
    lumina.store.set("activeIdx", winIdx)
end

-- Helper: increment click counter for a window
local function incrementClicks(winIdx)
    local windows = lumina.store.get("windows")
    local newWindows = {}
    for i, w in ipairs(windows) do
        if i == winIdx then
            local copy = {}
            for k, v in pairs(w) do copy[k] = v end
            copy.clicks = w.clicks + 1
            newWindows[i] = copy
        else
            newWindows[i] = w
        end
    end
    lumina.store.set("windows", newWindows)
end

-- Create a single window element
local function createWindowElement(win, isActive, winIdx)
    local t = lumina.getTheme()
    local borderColor = isActive and t.primary or t.surface1
    local titleBg = isActive and t.primary or t.surface1
    local titleFg = isActive and t.base or t.text
    local bg = isActive and t.surface0 or t.base

    return lumina.createElement("vbox", {
        key = win.id,
        style = {
            position = "absolute",
            left = win.x,
            top = win.y,
            width = win.w,
            height = win.h,
            border = "rounded",
            background = bg,
        },
        onClick = function()
            bringToFront(winIdx)
        end,
    },
        -- Title bar
        lumina.createElement("text", {
            bold = true,
            foreground = titleFg,
            background = titleBg,
        }, " " .. win.title .. string.rep(" ", math.max(0, win.w - #win.title - 4))),
        -- Content area
        lumina.createElement("text", {
            foreground = t.text,
            style = { height = win.h - 5 },
        }, win.content),
        -- Button
        lumina.createElement("text", {
            key = win.id .. "-btn",
            foreground = isActive and t.primary or t.accent,
            bold = true,
            onClick = function()
                incrementClicks(winIdx)
                bringToFront(winIdx)
            end,
        }, " [ Click Me (" .. win.clicks .. ") ]")
    )
end

lumina.app {
    id = "windows-app",
    store = {
        windows = initialWindows,
        windowOrder = {1, 2, 3},  -- indices into windows[], last = top/front
        activeIdx = 3,            -- which window is "active"
    },
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
        ["1"] = function()
            bringToFront(1)
        end,
        ["2"] = function()
            bringToFront(2)
        end,
        ["3"] = function()
            bringToFront(3)
        end,
        ["ArrowLeft"] = function()
            local idx = lumina.store.get("activeIdx")
            local windows = lumina.store.get("windows")
            local newWindows = {}
            for i, w in ipairs(windows) do
                if i == idx then
                    local copy = {}
                    for k, v in pairs(w) do copy[k] = v end
                    copy.x = math.max(0, w.x - 2)
                    newWindows[i] = copy
                else
                    newWindows[i] = w
                end
            end
            lumina.store.set("windows", newWindows)
        end,
        ["ArrowRight"] = function()
            local idx = lumina.store.get("activeIdx")
            local windows = lumina.store.get("windows")
            local newWindows = {}
            for i, w in ipairs(windows) do
                if i == idx then
                    local copy = {}
                    for k, v in pairs(w) do copy[k] = v end
                    copy.x = math.min(50, w.x + 2)
                    newWindows[i] = copy
                else
                    newWindows[i] = w
                end
            end
            lumina.store.set("windows", newWindows)
        end,
        ["ArrowUp"] = function()
            local idx = lumina.store.get("activeIdx")
            local windows = lumina.store.get("windows")
            local newWindows = {}
            for i, w in ipairs(windows) do
                if i == idx then
                    local copy = {}
                    for k, v in pairs(w) do copy[k] = v end
                    copy.y = math.max(0, w.y - 2)
                    newWindows[i] = copy
                else
                    newWindows[i] = w
                end
            end
            lumina.store.set("windows", newWindows)
        end,
        ["ArrowDown"] = function()
            local idx = lumina.store.get("activeIdx")
            local windows = lumina.store.get("windows")
            local newWindows = {}
            for i, w in ipairs(windows) do
                if i == idx then
                    local copy = {}
                    for k, v in pairs(w) do copy[k] = v end
                    copy.y = math.min(12, w.y + 2)
                    newWindows[i] = copy
                else
                    newWindows[i] = w
                end
            end
            lumina.store.set("windows", newWindows)
        end,
    },

    render = function()
        local t = lumina.getTheme()
        local windows = lumina.useStore("windows")
        local windowOrder = lumina.useStore("windowOrder")
        local activeIdx = lumina.useStore("activeIdx")

        -- Build window elements in windowOrder (last = top = painted last)
        local windowElements = {}
        for _, orderIdx in ipairs(windowOrder) do
            local win = windows[orderIdx]
            local isActive = (orderIdx == activeIdx)
            windowElements[#windowElements + 1] = createWindowElement(win, isActive, orderIdx)
        end

        -- Status bar at the bottom
        local statusText = " [1/2/3] Select  [←→↑↓] Move  [q] Quit  |  Active: " ..
            windows[activeIdx].title

        return lumina.createElement("vbox", {
            style = { width = 80, height = 24 },
        },
            -- Window container (relative positioning context)
            lumina.createElement("box", {
                style = { width = 80, height = 23 },
            }, table.unpack(windowElements)),
            -- Status bar
            lumina.createElement("text", {
                foreground = t.muted,
                background = t.surface0,
            }, statusText)
        )
    end,
}
