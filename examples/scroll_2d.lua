-- examples/scroll_2d.lua — 2D Scroll + Resize Demo
--
-- Demonstrates:
--   • Vertical scrolling: mouse wheel ↑/↓
--   • Horizontal scrolling: Shift + mouse wheel
--   • Drag-to-resize: drag the divider between panels
--   • Independent scrolling in each pane
--   • Toggle scrollbar visibility (scrollbar = "none")
--
-- Press q or Ctrl+C to quit.

local SplitPane = require("lux.split_pane")

lumina.app {
    id = "scroll-2d-demo",
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme()

        -- Scrollbar visibility state
        local showScrollbar, setShowScrollbar = lumina.useState("showScrollbar", true)
        local scrollbarStyle = showScrollbar and "" or "none"

        -- LEFT PANEL: Wide+tall grid (300 rows × 20 columns)
        local rows = {}
        for r = 1, 300 do
            local line = string.format(" %3d │", r)
            for c = 1, 20 do
                line = line .. string.format(" Cell(%d,%d) ", r, c)
            end
            rows[#rows + 1] = lumina.createElement("text", {
                key = "row" .. r,
                style = {
                    whiteSpace = "nowrap",
                    foreground = (r % 2 == 0) and t.text or t.subtext1,
                    background = (r % 2 == 0) and t.surface0 or t.base,
                },
            }, line)
        end

        local leftPanel = lumina.createElement("vbox", {
            key = "left-panel",
            style = { flex = 1, background = t.base },
        },
            lumina.createElement("text", {
                key = "left-header",
                style = { bold = true, foreground = t.blue, background = t.surface1 },
            }, " Grid (300×20) — Scroll both ways"),
            lumina.createElement("vbox", {
                key = "left-scroll",
                style = {
                    flex = 1,
                    overflow = "scroll",
                    scrollbar = scrollbarStyle,
                },
            }, table.unpack(rows))
        )

        -- RIGHT PANEL: Item list (100 items)
        local items = {}
        for i = 1, 100 do
            local status = (i % 5 == 0) and "●" or (i % 3 == 0) and "◐" or "○"
            local color = (i % 5 == 0) and t.green or (i % 3 == 0) and t.yellow or t.subtext0
            items[#items + 1] = lumina.createElement("text", {
                key = "item" .. i,
                style = {
                    foreground = color,
                    background = (i % 2 == 0) and t.surface0 or t.base,
                },
            }, string.format(" %s Task #%03d — %s", status, i,
                (i % 5 == 0) and "completed" or (i % 3 == 0) and "running" or "pending"))
        end

        local rightPanel = lumina.createElement("vbox", {
            key = "right-panel",
            style = { flex = 1, background = t.base },
        },
            lumina.createElement("text", {
                key = "right-header",
                style = { bold = true, foreground = t.mauve, background = t.surface1 },
            }, " Tasks (100) — Scroll vertically"),
            lumina.createElement("vbox", {
                key = "right-scroll",
                style = {
                    flex = 1,
                    overflow = "scroll",
                    scrollbar = scrollbarStyle,
                },
            }, table.unpack(items))
        )

        return lumina.createElement("vbox", {
            style = { width = 80, height = 24, background = t.base },
        },
            -- Title bar with toggle button
            lumina.createElement("hbox", {
                key = "title-bar",
                style = { background = t.base },
            },
                lumina.createElement("text", {
                    key = "title",
                    style = { flex = 1, bold = true, foreground = t.blue },
                }, " 📜 2D Scroll + Resize Demo"),
                lumina.createElement("text", {
                    key = "toggle-btn",
                    style = { bold = true, foreground = t.green },
                    onClick = function()
                        setShowScrollbar(not showScrollbar)
                    end,
                }, showScrollbar and " [Hide Scrollbar] " or " [Show Scrollbar] ")
            ),

            -- SplitPane with two scrollable panels
            lumina.createElement(SplitPane, {
                direction = "horizontal",
                sizes = { 50, 0 },
                minSizes = { 20, 15 },
                maxSizes = { 65, 0 },
                borderColor = t.blue,
            }, leftPanel, rightPanel),

            -- Footer
            lumina.createElement("text", {
                key = "footer",
                style = { foreground = t.subtext0 },
            }, " [↕ wheel] [Shift+↕ horiz] [drag divider] [click toggle] [q=quit]")
        )
    end,
}
