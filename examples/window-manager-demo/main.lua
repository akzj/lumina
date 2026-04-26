-- ============================================================================
-- Lumina Example: Window Manager Demo
-- ============================================================================
-- Demonstrates: Window manager — create, close, focus, drag, resize, tile
-- Run: lumina examples/window-manager-demo/main.lua
-- ============================================================================
local lumina = require("lumina")

local c = {
    bg = "#1E1E2E", fg = "#CDD6F4", accent = "#89B4FA",
    success = "#A6E3A1", error = "#F38BA8", muted = "#6C7086",
}

local windowCount = 0

local function createEditorWindow()
    windowCount = windowCount + 1
    local id = "editor" .. windowCount
    local App = lumina.defineComponent({
        name = "EditorWindow",
        render = function(self)
            return {
                type = "vbox",
                style = { background = c.bg },
                children = {
                    { type = "text", content = "  Editor: " .. id, style = { foreground = c.accent, bold = true, background = "#181825" } },
                    { type = "text", content = "  Content: (empty)", style = { foreground = c.fg } },
                    { type = "text", content = "  [e] new editor  [t] new terminal  [a] tile", style = { foreground = c.muted, dim = true } },
                },
            }
        end,
    })
    lumina.createWindow({ id = id, title = "Editor " .. windowCount, x = 5 + windowCount * 5, y = 3 + windowCount * 2, w = 30, h = 8, content = App })
    return id
end

local function createTerminalWindow()
    windowCount = windowCount + 1
    local id = "term" .. windowCount
    local App = lumina.defineComponent({
        name = "TerminalWindow",
        render = function(self)
            return {
                type = "vbox",
                style = { background = c.bg },
                children = {
                    { type = "text", content = "  Terminal: " .. id, style = { foreground = c.success, bold = true, background = "#181825" } },
                    { type = "text", content = "  $ _", style = { foreground = c.fg } },
                    { type = "text", content = "  [1-9] focus  [w] close  [m] max/restore", style = { foreground = c.muted, dim = true } },
                },
            }
        end,
    })
    lumina.createWindow({ id = id, title = "Terminal " .. windowCount, x = 40 + windowCount * 3, y = 2, w = 35, h = 10, content = App })
    return id
end

local App = lumina.defineComponent({
    name = "WindowManagerDemo",
    render = function(self)
        local wins = lumina.listWindows()
        return {
            type = "vbox",
            style = { background = c.bg },
            children = {
                { type = "text", content = "  Window Manager Demo", style = { foreground = c.accent, bold = true, background = "#181825" } },
                { type = "text", content = "  Windows: " .. windowCount .. "  |  F1 tile grid  F2 tile horiz  F3 tile vert", style = { foreground = c.muted } },
                { type = "text", content = "" },
                { type = "text", content = "  [e] new editor  [t] new terminal  [a] tile  [w] close  [q] quit", style = { foreground = c.fg } },
            },
        }
    end,
})

createEditorWindow()
createTerminalWindow()

lumina.onKey("e", function()
    createEditorWindow()
end)

lumina.onKey("t", function()
    createTerminalWindow()
end)

lumina.onKey("a", function()
    lumina.tileWindows()
end)

lumina.onKey("F1", function()
    lumina.tileWindows("grid")
end)

lumina.onKey("F2", function()
    lumina.tileWindows("horizontal")
end)

lumina.onKey("F3", function()
    lumina.tileWindows("vertical")
end)

lumina.onKey("q", function()
    lumina.quit()
end)

lumina.mount(App)
lumina.run()
