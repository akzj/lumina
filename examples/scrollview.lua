-- examples/scrollview.lua — ScrollView Widget Demo
--
-- Demonstrates the ScrollView widget with:
--   • Automatic vertical scrolling for overflowing content
--   • Keyboard scrolling: ↑/↓ (1 line), PageUp/PageDown (1 page), Home/End
--   • Mouse wheel scrolling (built-in engine support)
--   • Visual scrollbar (thumb + track)
--
-- Press q or Ctrl+C to quit.

local ScrollView = lumina.ScrollView

lumina.app {
    id = "scrollview-demo",
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme()

        -- Generate 50 lines of content
        local items = {}
        for i = 1, 50 do
            items[#items + 1] = lumina.createElement("text", {
                style = {
                    foreground = (i % 2 == 0) and t.text or t.subtext1,
                    background = (i % 2 == 0) and t.surface0 or t.base,
                },
            }, string.format("  %2d. %s", i, "Line " .. i .. " — scroll me with ↑↓ PgUp PgDn Home End or mouse wheel"))
        end

        return lumina.createElement("box", {
            style = { width = 80, height = 24, background = t.base },
        },
            -- Header
            lumina.createElement("text", {
                style = { foreground = t.lavender, bold = true, height = 1 },
            }, " 📜 ScrollView Demo — Use ↑↓ PageUp/PageDown Home/End to scroll"),

            -- ScrollView with 50 lines in a 20-row visible area
            ScrollView {
                style = {
                    flex = 1,
                    border = "single",
                    background = t.base,
                },
                table.unpack(items),
            },

            -- Footer
            lumina.createElement("text", {
                style = { foreground = t.muted, height = 1 },
            }, " [↑↓=scroll] [PgUp/PgDn=page] [Home/End=top/bottom] [q=quit]")
        )
    end,
}
