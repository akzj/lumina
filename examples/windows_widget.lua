-- examples/windows_widget.lua — Window Widget Demo (using WM module)
--
-- Demonstrates the Window widget with WindowManager:
--   • Drag title bar to move
--   • Drag bottom-right corner (◢) to resize
--   • Close button (✕)
--   • Click/drag brings window to front (z-order)
--   • Multiple overlapping windows
--   • Press 'r' to reset, 'o' to reopen closed windows
--
-- See also: examples/windows.lua for the pure Lua keyboard-only version.

local Window = lumina.Window
local WM = require("lux.wm")

lumina.app {
    id = "window-widget-demo",
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
        ["r"] = function()
            -- Reset: clear store so WM.create re-initializes
            lumina.store.set("wm", nil)
        end,
        ["o"] = function()
            -- Reopen all closed windows
            local s = lumina.store.get("wm")
            if s and s.frames then
                for id, f in pairs(s.frames) do
                    if not f.open then
                        -- Inline reopen (can't use mgr here since it's in render scope)
                        f.open = true
                        s.order[#s.order + 1] = id
                        s.activeId = id
                    end
                end
                lumina.store.set("wm", s)
            end
        end,
    },
    render = function()
        local t = lumina.getTheme()

        -- Create WM instance (idempotent — only initializes on first call)
        local mgr = WM.create("wm", {
            { id = "win1", title = "📝 Editor",  x = 2,  y = 1, w = 35, h = 12 },
            { id = "win2", title = "📊 Monitor", x = 20, y = 5, w = 35, h = 12 },
            { id = "win3", title = "🎨 Palette", x = 10, y = 3, w = 30, h = 10 },
        })

        -- Get ordered open windows (bottom to top) — reactive via useStore
        local windows = mgr.getWindows()

        -- Window content (could be dynamic per-window)
        local content = {
            win1 = "Welcome to the editor.\nDrag the title bar to move.\nDrag ◢ to resize.\nClick ✕ to close.",
            win2 = "System monitor.\nCPU: 42%  Memory: 68%\nProcesses: 142",
            win3 = "Color palette.\nRed Green Blue\nBrush: Round 3px",
        }

        local children = {}

        for _, win in ipairs(windows) do
            local winId = win.id  -- stable id, never an array index
            children[#children + 1] = Window {
                title = win.title,
                x = win.x, y = win.y,
                width = win.w, height = win.h,
                key = winId,
                onChange = function(e)
                    if e == "close" then
                        mgr.close(winId)
                    elseif e == "activate" then
                        mgr.activate(winId)
                    elseif type(e) == "table" then
                        -- move/resize: only update frame, do NOT activate
                        if e.type == "move" then
                            mgr.setFrame(winId, { x = e.x, y = e.y })
                        elseif e.type == "resize" then
                            mgr.setFrame(winId, { w = e.width, h = e.height })
                        end
                    end
                end,
                lumina.createElement("text", {
                    style = { foreground = t.text },
                }, content[winId] or ""),
            }
        end

        -- Status bar at bottom
        children[#children + 1] = lumina.createElement("text", {
            style = { foreground = t.muted, position = "fixed", left = 0, top = 23, height = 1 },
        }, " [Drag title=move] [Drag ◢=resize] [✕=close] [r=reset] [o=reopen] [q=quit]")

        return lumina.createElement("box", {
            style = { width = 80, height = 24, background = t.base },
        }, table.unpack(children))
    end,
}
