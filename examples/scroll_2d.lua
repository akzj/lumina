-- examples/scroll_2d.lua — 2D Scroll Demo (Vertical + Horizontal)
--
-- Demonstrates:
--   • Vertical scrolling: mouse wheel ↑/↓
--   • Horizontal scrolling: Shift + mouse wheel
--   • Combined: content overflows in both directions
--
-- Press q or Ctrl+C to quit.

lumina.app {
    id = "scroll-2d-demo",
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme()

        -- Generate a wide+tall grid (80 rows × 20 columns)
        local rows = {}
        for r = 1, 300 do
            -- Each row is a long text line (wider than viewport)
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

        -- Header row (also wide)
        local header = "  #  │"
        for c = 1, 20 do
            header = header .. string.format(" Col %-5d ", c)
        end

        return lumina.createElement("vbox", {
            style = { width = 80, height = 24, background = t.base },
        },
            -- Title
            lumina.createElement("text", {
                key = "title",
                style = { bold = true, foreground = t.blue },
            }, " 📜 2D Scroll Demo — Scroll: ↕ wheel │ ↔ Shift+wheel"),

            -- Column header (fixed, not scrolled)
            lumina.createElement("text", {
                key = "header",
                style = {
                    foreground = t.blue,
                    bold = true,
                    background = t.surface1,
                },
            }, string.sub(header, 1, 78)),

            -- Scrollable area (both directions)
            lumina.createElement("vbox", {
                key = "scroll-area",
                style = {
                    flex = 1,
                    overflow = "scroll",
                    border = "single",
                    borderColor = t.blue,
                },
            }, table.unpack(rows)),

            -- Footer
            lumina.createElement("text", {
                key = "footer",
                style = { foreground = t.subtext0 },
            }, " [↕ wheel=vertical] [Shift+↕=horizontal] [q=quit]")
        )
    end,
}
