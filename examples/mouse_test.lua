-- Lumina Mouse Grid Test: React declarative style + keyboard navigation
-- Each cell is an independent component with its own useState for hover.
-- Click state and cursor position are managed via a global store for keyboard access.
-- Uses createElement + defineComponent + onMouseEnter/onMouseLeave/onClick props.
-- Keyboard: h/j/k/l or arrows to move cursor, Enter/Space to select, c to clear.
local lumina = require("lumina")

local theme = {
    bg = "#1E1E2E", dot = "#585B70", hover = "#A6E3A1",
    click = "#F38BA8", both = "#F9E2AF", bar = "#181825",
    accent = "#89B4FA", cursor = "#CBA6F7",
}

-- ============================================================
-- Global store: shared state accessible from onKey handlers
-- ============================================================
-- Grid dimensions (updated by App on each render, used by onKey for clamping)
local gridCols = 30
local gridRows = 15

local store = lumina.createStore({
    state = {
        cursorX = 0,
        cursorY = 0,
        lastHover = "",
        lastClick = "",
        clickedCells = {},  -- table of cellId -> true
    },
    actions = {
        moveCursor = function(state, dir)
            if dir == "left" then
                state.cursorX = math.max(0, state.cursorX - 1)
            elseif dir == "right" then
                state.cursorX = math.min(gridCols - 1, state.cursorX + 1)
            elseif dir == "up" then
                state.cursorY = math.max(0, state.cursorY - 1)
            elseif dir == "down" then
                state.cursorY = math.min(gridRows - 1, state.cursorY + 1)
            end
        end,
        toggleCell = function(state)
            local id = "c-" .. state.cursorX .. "-" .. state.cursorY
            if state.clickedCells[id] then
                state.clickedCells[id] = nil
            else
                state.clickedCells[id] = true
            end
            state.lastClick = id
        end,
        clickCell = function(state, id)
            if state.clickedCells[id] then
                state.clickedCells[id] = nil
            else
                state.clickedCells[id] = true
            end
            state.lastClick = id
        end,
        clearAll = function(state)
            state.clickedCells = {}
            state.lastClick = ""
        end,
    },
})

-- ============================================================
-- Key bindings (global — outside components)
-- ============================================================
lumina.onKey("h", function() store.dispatch("moveCursor", "left") end)
lumina.onKey("l", function() store.dispatch("moveCursor", "right") end)
lumina.onKey("k", function() store.dispatch("moveCursor", "up") end)
lumina.onKey("j", function() store.dispatch("moveCursor", "down") end)
lumina.onKey("left", function() store.dispatch("moveCursor", "left") end)
lumina.onKey("right", function() store.dispatch("moveCursor", "right") end)
lumina.onKey("up", function() store.dispatch("moveCursor", "up") end)
lumina.onKey("down", function() store.dispatch("moveCursor", "down") end)
lumina.onKey("enter", function() store.dispatch("toggleCell") end)
lumina.onKey(" ", function() store.dispatch("toggleCell") end)
lumina.onKey("c", function() store.dispatch("clearAll") end)
lumina.onKey("q", function() lumina.quit() end)

-- ============================================================
-- Cell: independent component with own hover state
-- Click state comes from the global store via props.
-- ============================================================
local Cell = lumina.defineComponent({
    name = "Cell",
    render = function(props)
        local hovered, setHovered = lumina.useState("hovered", false)
        local clicked = props.clicked
        local cursorActive = props.cursorActive

        local label = " . "
        local fg = theme.dot
        local bg = theme.bg

        if hovered and clicked then
            label = "[*]"
            fg = theme.both
            bg = "#45475A"
        elseif hovered then
            label = "[O]"
            fg = theme.hover
            bg = "#313244"
        elseif clicked then
            label = "[X]"
            fg = theme.click
            bg = "#313244"
        end

        -- Keyboard cursor overlay: distinct border/background
        if cursorActive then
            if not hovered and not clicked then
                label = "[ ]"
                fg = theme.cursor
            end
            bg = "#45475A"
        end

        return {
            type = "box",
            id = props.id,
            style = { width = 3, height = 1, background = bg },
            onClick = function()
                store.dispatch("clickCell", props.id)
            end,
            onMouseEnter = function()
                setHovered(true)
                if props.onCellHover then
                    props.onCellHover(props.id)
                end
            end,
            onMouseLeave = function()
                setHovered(false)
            end,
            children = {
                { type = "text", content = label, style = { foreground = fg } }
            }
        }
    end
})

-- ============================================================
-- App: composes Cell components in a grid
-- ============================================================
local App = lumina.defineComponent({
    name = "MouseGrid",
    render = function()
        local state = lumina.useStore(store)

        local width, height = lumina.getSize()
        local cols = math.min(math.floor(width / 3), 30)
        local rows = math.min(height - 1, 15)

        -- Update global grid dimensions for onKey cursor clamping
        gridCols = cols
        gridRows = rows

        -- Clamp cursor to current grid bounds (in case terminal resized)
        local cursorX = math.min(state.cursorX, cols - 1)
        local cursorY = math.min(state.cursorY, rows - 1)

        local children = {}

        -- Status bar
        children[1] = {
            type = "text",
            content = string.format(
                " Hover: %s | Click: %s | Cursor: (%d,%d) | [hjkl] Move [Enter] Select [c] Clear [q] Quit",
                state.lastHover, state.lastClick, cursorX, cursorY),
            style = { foreground = theme.accent, background = theme.bar, bold = true },
        }

        -- Grid of Cell components
        for y = 0, rows - 1 do
            local rowChildren = {}
            for x = 0, cols - 1 do
                local cellId = "c-" .. x .. "-" .. y
                local isCursorHere = (x == cursorX and y == cursorY)
                local isClicked = (state.clickedCells[cellId] == true)
                rowChildren[#rowChildren + 1] = lumina.createElement(Cell, {
                    id = cellId,
                    key = cellId,
                    clicked = isClicked,
                    cursorActive = isCursorHere,
                    onCellHover = function(id)
                        store.dispatch("setState", { lastHover = id })
                    end,
                })
            end
            children[#children + 1] = { type = "hbox", children = rowChildren }
        end

        return {
            type = "vbox",
            style = { background = theme.bg },
            children = children,
        }
    end
})

lumina.mount(App)
lumina.run()
