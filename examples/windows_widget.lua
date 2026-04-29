-- examples/windows_widget.lua — Window Widget Demo
--
-- Demonstrates the lux Window widget with:
--   • Drag title bar to move
--   • Drag bottom-right corner (◢) to resize
--   • Close button (✕)
--   • Click/drag brings window to front (z-order)
--   • Multiple overlapping windows
--
-- See also: examples/windows.lua for the pure Lua keyboard-only version.

local Window = lumina.Window

lumina.app {
    id = "window-widget-demo",
    store = {
        windows = {
            { id = "win1", title = "📝 Editor",  x = 2,  y = 1, w = 35, h = 12, open = true,
              content = "Welcome to the editor.\nDrag the title bar to move.\nDrag ◢ to resize.\nClick ✕ to close." },
            { id = "win2", title = "📊 Monitor", x = 20, y = 5, w = 35, h = 12, open = true,
              content = "System monitor.\nCPU: 42%  Memory: 68%\nProcesses: 142" },
            { id = "win3", title = "🎨 Palette", x = 10, y = 3, w = 30, h = 10, open = true,
              content = "Color palette.\nRed Green Blue\nBrush: Round 3px" },
        },
    },
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
        ["r"] = function()
            -- Reset all windows
            lumina.store.set("windows", {
                { id = "win1", title = "📝 Editor",  x = 2,  y = 1, w = 35, h = 12, open = true,
                  content = "Welcome to the editor.\nDrag the title bar to move.\nDrag ◢ to resize.\nClick ✕ to close." },
                { id = "win2", title = "📊 Monitor", x = 20, y = 5, w = 35, h = 12, open = true,
                  content = "System monitor.\nCPU: 42%  Memory: 68%\nProcesses: 142" },
                { id = "win3", title = "🎨 Palette", x = 10, y = 3, w = 30, h = 10, open = true,
                  content = "Color palette.\nRed Green Blue\nBrush: Round 3px" },
            })
        end,
    },
    render = function()
        local t = lumina.getTheme()
        local windows = lumina.useStore("windows")

        -- Helper: bring window at index to front (end of array = on top)
        local function bringToFront(idx)
            local wins = lumina.store.get("windows")
            local win = table.remove(wins, idx)
            wins[#wins + 1] = win
            lumina.store.set("windows", wins)
        end

        local children = {}

        for i, win in ipairs(windows) do
            if win.open then
                local winIdx = i  -- capture for closure
                children[#children + 1] = Window {
                    title = win.title,
                    x = win.x, y = win.y,
                    width = win.w, height = win.h,
                    key = win.id,
                    onChange = function(e)
                        if e == "close" then
                            local wins = lumina.store.get("windows")
                            wins[winIdx].open = false
                            lumina.store.set("windows", wins)
                        elseif type(e) == "table" then
                            bringToFront(winIdx)
                            local wins = lumina.store.get("windows")
                            -- After bringToFront, this window is now at end
                            local w = wins[#wins]
                            if e.type == "move" then
                                w.x = e.x
                                w.y = e.y
                            elseif e.type == "resize" then
                                w.w = e.width
                                w.h = e.height
                            end
                            lumina.store.set("windows", wins)
                        end
                    end,
                    lumina.createElement("text", { style = { foreground = t.text } }, win.content),
                }
            end
        end

        -- Status bar at bottom
        children[#children + 1] = lumina.createElement("text", {
            style = { foreground = t.muted, position = "fixed", left = 0, top = 23, height = 1 },
        }, " [Drag title=move] [Drag ◢=resize] [✕=close] [r=reset] [q=quit]")

        return lumina.createElement("box", {
            style = { width = 80, height = 24, background = t.base },
        }, table.unpack(children))
    end,
}
