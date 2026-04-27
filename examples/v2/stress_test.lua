-- Lumina v2 Example: Fullscreen Stress Test
-- Demonstrates: 1840+ individual elements, hover tracking, click toggling,
--               maximum element count rendering performance.
--
-- Usage: lumina-v2 examples/v2/stress_test.lua
-- Quit:  q or Ctrl+Q
--
-- Keyboard:
--   q / Ctrl+Q  - Quit
--   c            - Clear all clicked cells
--
-- Mouse:
--   Hover        - Highlights individual cells
--   Click        - Toggles cell on/off

-- Theme (Catppuccin Mocha)
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
local ROWS = 23  -- 24 total minus 1 for status bar

lumina.createComponent({
    id = "stress",
    name = "StressTest",
    x = 0, y = 0,
    w = COLS, h = ROWS + 1,
    zIndex = 0,

    render = function(state, props)
        local clickedCells, setClickedCells = lumina.useState("clicked", {})
        local clickCount, setClickCount = lumina.useState("clickCount", 0)
        local lastClick, setLastClick = lumina.useState("lastClick", "")
        local hoveredCell, setHoveredCell = lumina.useState("hovered", "")

        -- Keyboard handler
        local function handleKey(e)
            if e.key == "q" then
                lumina.quit()
            elseif e.key == "c" then
                setClickedCells({})
                setClickCount(0)
                setLastClick("")
            end
        end

        -- Toggle a cell on click
        local function toggleCell(cellId)
            local newClicked = {}
            for k, v in pairs(clickedCells) do
                newClicked[k] = v
            end
            if newClicked[cellId] then
                newClicked[cellId] = nil
            else
                newClicked[cellId] = true
            end
            setClickedCells(newClicked)
            setClickCount(clickCount + 1)
            setLastClick(cellId)
        end

        -- Build grid: each row is an hbox of individual 1-char box cells
        local rowElements = {}
        for y = 0, ROWS - 1 do
            local cellsInRow = {}
            for x = 0, COLS - 1 do
                local cellId = x .. "," .. y
                local isHovered = (hoveredCell == cellId)
                local isClicked = (clickedCells[cellId] == true)

                local ch, fg, bg
                if isHovered and isClicked then
                    ch = "*"
                    fg = theme.click
                    bg = theme.bothBg
                elseif isHovered then
                    ch = "█"
                    fg = theme.hover
                    bg = theme.hoverBg
                elseif isClicked then
                    ch = "×"
                    fg = theme.click
                    bg = theme.hoverBg
                else
                    ch = "·"
                    fg = theme.dot
                    bg = theme.bg
                end

                -- Capture cellId in closure
                local cid = cellId
                cellsInRow[#cellsInRow + 1] = {
                    type = "box",
                    id = cid,
                    style = {width = 1, height = 1, background = bg},
                    onMouseEnter = function() setHoveredCell(cid) end,
                    onMouseLeave = function() setHoveredCell("") end,
                    onClick = function() toggleCell(cid) end,
                    children = {
                        {type = "text", content = ch, style = {foreground = fg}},
                    },
                }
            end

            rowElements[#rowElements + 1] = {
                type = "hbox",
                id = "row-" .. y,
                style = {height = 1},
                children = cellsInRow,
            }
        end

        -- Status bar
        local statusText = string.format(
            " %dx%d=%d cells | Click:%s | Clicked:%d | [c]Clear [q]Quit",
            COLS, ROWS, COLS * ROWS,
            lastClick,
            clickCount
        )
        rowElements[#rowElements + 1] = lumina.createElement("text", {
            foreground = theme.accent,
            bold = true,
            style = {background = theme.bar, height = 1},
        }, statusText)

        -- Root container: vbox with all rows + status bar
        return {
            type = "vbox",
            id = "stress-root",
            style = {background = theme.bg},
            onKeyDown = handleKey,
            focusable = true,
            children = rowElements,
        }
    end,
})
