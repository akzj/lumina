-- ============================================================================
-- Lumina Example: DevTools Demo
-- ============================================================================
-- Demonstrates: Built-in DevTools inspector — F12 to toggle, element selection
-- Run: lumina examples/devtools-demo/main.lua
-- Press F12 to open DevTools panel
-- ============================================================================
local lumina = require("lumina")

local c = {
    bg = "#1E1E2E", fg = "#CDD6F4", accent = "#89B4FA",
    success = "#A6E3A1", error = "#F38BA8", muted = "#6C7086",
    surface = "#313244",
}

local showInfo = false

local InfoPanel = lumina.defineComponent({
    name = "InfoPanel",
    render = function()
        return {
            type = "vbox",
            style = { background = c.surface, border = c.accent },
            children = {
                { type = "text", content = "  Component Inspector", style = { foreground = c.accent, bold = true } },
                { type = "text", content = "  ─────────────────────", style = { foreground = c.muted } },
                { type = "text", content = "  Hover elements to inspect", style = { foreground = c.fg } },
                { type = "text", content = "  Click to select", style = { foreground = c.fg } },
                { type = "text", content = "" },
                { type = "text", content = "  F12: toggle DevTools", style = { foreground = c.muted } },
            },
        }
    end,
})

local Counter = lumina.defineComponent({
    name = "Counter",
    state = { count = 0 },
    render = function(self)
        return {
            type = "vbox",
            style = { background = c.surface, padding = 1 },
            children = {
                { type = "text", content = "  Counter: " .. self.count, style = { foreground = c.accent, bold = true } },
                {
                    type = "hbox",
                    children = {
                        {
                            type = "button",
                            id = "dec-btn",
                            label = " - ",
                            onClick = function()
                                self:setState({ count = self.count - 1 })
                            end,
                        },
                        { type = "text", content = "  ", style = { minWidth = 2 } },
                        {
                            type = "button",
                            id = "inc-btn",
                            label = " + ",
                            onClick = function()
                                self:setState({ count = self.count + 1 })
                            end,
                        },
                    },
                },
            },
        }
    end,
})

local TodoApp = lumina.defineComponent({
    name = "TodoApp",
    state = { items = { "Learn Lumina", "Build TUI apps", "Ship it!" }, input = "" },
    render = function(self)
        local children = {
            { type = "text", content = "  Todo List", style = { foreground = c.accent, bold = true } },
            { type = "separator" },
        }
        for i, item in ipairs(self.items) do
            table.insert(children, {
                type = "text",
                id = "todo-" .. i,
                content = "  " .. i .. ". " .. item,
                style = { foreground = c.fg },
            })
        end
        return {
            type = "vbox",
            style = { background = c.surface, padding = 1 },
            children = children,
        }
    end,
})

local App = lumina.defineComponent({
    name = "App",
    render = function()
        return {
            type = "vbox",
            style = { background = c.bg },
            children = {
                { type = "text", content = "  DevTools Demo — Press F12 to inspect", style = { foreground = c.accent, bold = true, background = "#181825" } },
                { type = "text", content = "" },
                {
                    type = "hbox",
                    children = {
                        { type = "box", id = "counter-box", children = { Counter }, style = { minWidth = 20 } },
                        { type = "text", content = "  ", style = { minWidth = 2 } },
                        { type = "box", id = "todo-box", children = { TodoApp }, style = { minWidth = 25 } },
                    },
                },
                { type = "text", content = "" },
                { type = "text", content = "  Click + / - to change counter, see DevTools react", style = { foreground = c.muted, dim = true } },
            },
        }
    end,
})

lumina.mount(App)
lumina.run()
