-- examples/windows_widget.lua — Window Widget Demo
--
-- Demonstrates the lux Window widget with:
--   • Drag title bar to move
--   • Drag bottom-right corner (◢) to resize
--   • Close button (✕)
--   • Multiple overlapping windows
--
-- See also: examples/windows.lua for the pure Lua keyboard-only version.

local Window = lumina.Window

lumina.app {
    id = "window-widget-demo",
    store = {
        win1 = {x = 2, y = 1, w = 35, h = 12, open = true},
        win2 = {x = 20, y = 5, w = 35, h = 12, open = true},
        win3 = {x = 10, y = 3, w = 30, h = 10, open = true},
    },
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
        ["r"] = function()
            -- Reset all windows
            lumina.store.set("win1", {x = 2, y = 1, w = 35, h = 12, open = true})
            lumina.store.set("win2", {x = 20, y = 5, w = 35, h = 12, open = true})
            lumina.store.set("win3", {x = 10, y = 3, w = 30, h = 10, open = true})
        end,
    },
    render = function()
        local t = lumina.getTheme()
        local win1 = lumina.useStore("win1")
        local win2 = lumina.useStore("win2")
        local win3 = lumina.useStore("win3")

        local children = {}

        -- Helper to create window onChange handler
        local function makeHandler(storeKey)
            return function(e)
                if e == "close" then
                    local s = lumina.store.get(storeKey)
                    s.open = false
                    lumina.store.set(storeKey, s)
                elseif type(e) == "table" then
                    local s = lumina.store.get(storeKey)
                    if e.type == "move" then
                        s.x = e.x
                        s.y = e.y
                    elseif e.type == "resize" then
                        s.w = e.width
                        s.h = e.height
                    end
                    lumina.store.set(storeKey, s)
                end
            end
        end

        if win1.open then
            children[#children + 1] = Window {
                title = "📝 Editor",
                x = win1.x, y = win1.y,
                width = win1.w, height = win1.h,
                key = "win1",
                onChange = makeHandler("win1"),
                lumina.createElement("text", { style = { foreground = t.text } },
                    "Welcome to the editor.\nDrag the title bar to move.\nDrag ◢ to resize.\nClick ✕ to close."),
            }
        end

        if win2.open then
            children[#children + 1] = Window {
                title = "📊 Monitor",
                x = win2.x, y = win2.y,
                width = win2.w, height = win2.h,
                key = "win2",
                onChange = makeHandler("win2"),
                lumina.createElement("text", { style = { foreground = t.text } },
                    "System monitor.\nCPU: 42%  Memory: 68%\nProcesses: 142"),
            }
        end

        if win3.open then
            children[#children + 1] = Window {
                title = "🎨 Palette",
                x = win3.x, y = win3.y,
                width = win3.w, height = win3.h,
                key = "win3",
                onChange = makeHandler("win3"),
                lumina.createElement("text", { style = { foreground = t.text } },
                    "Color palette.\nRed Green Blue\nBrush: Round 3px"),
            }
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
