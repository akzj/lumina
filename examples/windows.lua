-- examples/windows.lua — Multi-Window Manager using lux.wm + lux.window
--
-- Demonstrates overlapping windows with z-order management:
--   • Absolute positioning for window placement
--   • Z-order via child ordering (last child = on top)
--   • Click to bring window to front
--   • Drag via title bar (onMouseDown + onMouseMove)
--   • Resize via bottom-right corner
--   • Keyboard movement of active window
--   • Per-window click counters
--
-- Uses lux.wm for window state + lux.window for drag/resize.
--
-- Features showcased:
--   • position = "absolute" with left/top/width/height
--   • onClick, onMouseDown, onMouseMove, onMouseUp handlers
--   • lumina.store for state management
--   • Global keybindings (1/2/3=select, arrows=move, q=quit)
--
-- Usage: lumina examples/windows.lua

local WM = require("lux.wm")
local Window = require("lux.window")
local mgr = WM.create("wm_state", {
    {id = "win1", title = "📝 Editor", x = 2, y = 1, w = 30, h = 12},
    {id = "win2", title = "📊 Monitor", x = 15, y = 5, w = 30, h = 12},
    {id = "win3", title = "🎨 Palette", x = 28, y = 3, w = 30, h = 12},
})

-- Per-window content (static)
local windowContent = {
    win1 = "Welcome to the editor window.\nType your code here.\nLine 3 of content.",
    win2 = "CPU: 42%  MEM: 1.2GB\nProcesses: 128\nUptime: 3d 14h",
    win3 = "Colors: Red, Green, Blue\nBrush: Round 3px\nOpacity: 80%",
}

-- Create a single window element using lux.window
local function createWindowElement(win, isActive)
    local t = lumina.getTheme()
    local clicks = lumina.useStore("wm_clicks")
    local clickCount = (clicks and clicks[win.id]) or 0
    local content = windowContent[win.id] or ""

    return lumina.createElement(Window, {
        id = win.id,
        title = win.title,
        x = win.x, y = win.y,
        w = win.w, h = win.h,
        isActive = isActive,
        onActivate = function()
            mgr.activate(win.id)
        end,
        onMove = function(newX, newY)
            mgr.setFrame(win.id, {x = newX, y = newY})
        end,
        onResize = function(newW, newH)
            mgr.setFrame(win.id, {w = newW, h = newH})
        end,
    },
        -- Content area
        lumina.createElement("text", {
            foreground = t.text,
            style = { height = math.max(0, win.h - 5) },
        }, content),
        -- Button
        lumina.createElement("text", {
            key = win.id .. "-btn",
            foreground = isActive and t.primary or t.accent,
            bold = true,
            onClick = function()
                local c = lumina.store.get("wm_clicks") or {}
                c[win.id] = (c[win.id] or 0) + 1
                lumina.store.set("wm_clicks", c)
                mgr.activate(win.id)
            end,
        }, " [ Click Me (" .. clickCount .. ") ]")
    )
end

lumina.app {
    id = "windows-app",
    store = {
        wm_clicks = {win1 = 0, win2 = 0, win3 = 0},
    },
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
        ["1"] = function()
            mgr.activate("win1")
        end,
        ["2"] = function()
            mgr.activate("win2")
        end,
        ["3"] = function()
            mgr.activate("win3")
        end,
        ["ArrowLeft"] = function()
            local s = lumina.store.get("wm_state")
            local id = s.activeId
            if id and s.frames[id] then
                local f = s.frames[id]
                mgr.setFrame(id, {x = math.max(0, f.x - 2)})
            end
        end,
        ["ArrowRight"] = function()
            local s = lumina.store.get("wm_state")
            local id = s.activeId
            if id and s.frames[id] then
                local f = s.frames[id]
                mgr.setFrame(id, {x = math.min(50, f.x + 2)})
            end
        end,
        ["ArrowUp"] = function()
            local s = lumina.store.get("wm_state")
            local id = s.activeId
            if id and s.frames[id] then
                local f = s.frames[id]
                mgr.setFrame(id, {y = math.max(0, f.y - 2)})
            end
        end,
        ["ArrowDown"] = function()
            local s = lumina.store.get("wm_state")
            local id = s.activeId
            if id and s.frames[id] then
                local f = s.frames[id]
                mgr.setFrame(id, {y = math.min(12, f.y + 2)})
            end
        end,
    },

    render = function()
        local t = lumina.getTheme()
        local windows = mgr.getWindows()
        local activeId = mgr.getActiveId()

        -- Build window elements in order (last = top = painted last)
        local windowElements = {}
        for _, win in ipairs(windows) do
            local isActive = (win.id == activeId)
            windowElements[#windowElements + 1] = createWindowElement(win, isActive)
        end

        -- Find active window title for status bar
        local activeTitle = ""
        for _, win in ipairs(windows) do
            if win.id == activeId then
                activeTitle = win.title
                break
            end
        end

        -- Status bar at the bottom
        local statusText = " [1/2/3] Select  [←→↑↓] Move  [q] Quit  |  Active: " .. activeTitle

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
