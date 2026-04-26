-- Lumina Fullscreen Stress Test: 1-char cells filling entire terminal
-- Tests rendering performance with maximum cell count.
local lumina = require("lumina")

local theme = {
    bg = "#1E1E2E", dot = "#585B70", hover = "#A6E3A1",
    click = "#F38BA8", bar = "#181825", accent = "#89B4FA",
}

local store = lumina.createStore({
    state = {
        hoverX = -1,
        hoverY = -1,
        clickedCells = {},
        lastClick = "",
        renders = 0,
    },
    actions = {
        setHover = function(state, x, y)
            state.hoverX = x
            state.hoverY = y
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
        end,
        incRenders = function(state)
            state.renders = state.renders + 1
        end,
    },
})

lumina.onKey("q", function() lumina.quit() end)
lumina.onKey("c", function() store.dispatch("clearAll") end)

-- Single-char Cell component
local Cell = lumina.defineComponent({
    name = "Cell",
    render = function(props)
        local hovered, setHovered = lumina.useState("hovered", false)
        local clicked = props.clicked

        local ch = "·"
        local fg = theme.dot
        local bg = theme.bg

        if hovered and clicked then
            ch = "*"
            fg = theme.click
            bg = "#45475A"
        elseif hovered then
            ch = "█"
            fg = theme.hover
            bg = "#313244"
        elseif clicked then
            ch = "×"
            fg = theme.click
            bg = "#313244"
        end

        return {
            type = "box",
            id = props.id,
            style = { width = 1, height = 1, background = bg },
            onClick = function()
                store.dispatch("clickCell", props.id)
            end,
            onMouseEnter = function()
                setHovered(true)
                store.dispatch("setHover", props.col, props.row)
            end,
            onMouseLeave = function()
                setHovered(false)
            end,
            children = {
                { type = "text", content = ch, style = { foreground = fg } }
            }
        }
    end
})

-- App: fullscreen grid
local App = lumina.defineComponent({
    name = "StressGrid",
    render = function()
        local state = lumina.useStore(store)
        store.dispatch("incRenders")

        local width, height = lumina.getSize()
        local cols = width
        local rows = height - 1  -- reserve 1 row for status

        local cellCount = cols * rows
        local children = {}

        -- Status bar
        children[1] = {
            type = "text",
            content = string.format(
                " %dx%d=%d cells | Hover:(%d,%d) | Click:%s | Renders:%d | [c]Clear [q]Quit",
                cols, rows, cellCount,
                state.hoverX, state.hoverY,
                state.lastClick or "",
                state.renders),
            style = { foreground = theme.accent, background = theme.bar, bold = true },
        }

        -- Grid rows
        for y = 0, rows - 1 do
            local rowChildren = {}
            for x = 0, cols - 1 do
                local cellId = x .. "," .. y
                local isClicked = (state.clickedCells[cellId] == true)
                rowChildren[#rowChildren + 1] = lumina.createElement(Cell, {
                    id = cellId,
                    key = cellId,
                    col = x,
                    row = y,
                    clicked = isClicked,
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
