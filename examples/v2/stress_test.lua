-- Lumina v2: Stress Test with Per-Cell Components
-- Each cell is an independent component with its own state.
-- Hover triggers re-render of only 1-2 cells, not all 1840.
--
-- Usage: lumina-v2 examples/v2/stress_test.lua
-- Quit:  q or Ctrl+Q
--
-- Mouse:
--   Hover  - Highlights individual cells (per-cell re-render)
--   Click  - Toggles cell on/off

local theme = {
    bg      = "#1E1E2E",
    dot     = "#585B70",
    hover   = "#A6E3A1",
    click   = "#F38BA8",
    bar     = "#181825",
    accent  = "#89B4FA",
    hoverBg = "#313244",
    bothBg  = "#45475A",
}

local COLS = 80
local ROWS = 23

-- Each Cell is an independent component with its own hover/click state.
-- Hover only re-renders 1-2 cells instead of all 1840.
local Cell = lumina.defineComponent("Cell", function(props)
    local hovered, setHovered = lumina.useState("h", false)
    local clicked, setClicked = lumina.useState("c", false)

    local ch, fg, bg
    if hovered and clicked then
        ch = "*"
        fg = theme.click
        bg = theme.bothBg
    elseif hovered then
        ch = "█"
        fg = theme.hover
        bg = theme.hoverBg
    elseif clicked then
        ch = "×"
        fg = theme.click
        bg = theme.hoverBg
    else
        ch = "·"
        fg = theme.dot
        bg = theme.bg
    end

    return lumina.createElement("box", {
        style = {width = 1, height = 1, background = bg},
        onMouseEnter = function() setHovered(true) end,
        onMouseLeave = function() setHovered(false) end,
        onClick = function() setClicked(not clicked) end,
    }, lumina.createElement("text", {
        style = {foreground = fg},
    }, ch))
end)

-- Root component: creates the grid structure.
-- This only re-renders when the grid structure changes (never on hover).
lumina.createComponent({
    id = "stress",
    name = "StressTest",

    render = function(props)
        local rowElements = {}
        for y = 0, ROWS - 1 do
            local cellsInRow = {}
            for x = 0, COLS - 1 do
                local cellId = x .. "," .. y
                cellsInRow[#cellsInRow + 1] = lumina.createElement(Cell, {
                    key = cellId,
                    id = cellId,
                })
            end
            rowElements[#rowElements + 1] = {
                type = "hbox",
                id = "row-" .. y,
                style = {height = 1},
                children = cellsInRow,
            }
        end

        -- Status bar
        rowElements[#rowElements + 1] = lumina.createElement("text", {
            foreground = theme.accent,
            bold = true,
            style = {background = theme.bar, height = 1},
        }, string.format(
            " %dx%d=%d cells | Per-cell components | [q]Quit",
            COLS, ROWS, COLS * ROWS
        ))

        return {
            type = "vbox",
            id = "stress-root",
            style = {background = theme.bg},
            onKeyDown = function(e)
                if e.key == "q" then lumina.quit() end
            end,
            focusable = true,
            children = rowElements,
        }
    end,
})
