-- Lumina Fullscreen Stress Test: 1-char cells filling entire terminal
-- Tests rendering performance with maximum cell count.
local lumina = require("lumina")

local theme = {
    bg = "#1E1E2E", dot = "#585B70", hover = "#A6E3A1",
    click = "#F38BA8", bar = "#181825", accent = "#89B4FA",
}

local store = lumina.createStore({
    state = {
        clickedCells = {},
        lastClick = "",
        clickCount = 0,
    },
    actions = {
        clickCell = function(state, id)
            if state.clickedCells[id] then
                state.clickedCells[id] = nil
            else
                state.clickedCells[id] = true
            end
            state.lastClick = id
            state.clickCount = state.clickCount + 1
        end,
        clearAll = function(state)
            state.clickedCells = {}
            state.clickCount = 0
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

-- StatusBar: separate component that re-renders on ANY store change
-- (no selector = always dirty on dispatch). Cheap to render (1 text node).
local StatusBar = lumina.defineComponent({
    name = "StatusBar",
    render = function()
        local state = lumina.useStore(store)
        local fps = lumina.getFPS and lumina.getFPS() or 0
        local width, height = lumina.getSize()
        local cols = width
        local rows = height - 1

        return {
            type = "text",
            content = string.format(
                " %dx%d=%d cells | FPS:%d | Click:%s | Clicked:%d | [c]Clear [q]Quit",
                cols, rows, cols * rows, fps,
                state.lastClick or "",
                state.clickCount or 0),
            style = { foreground = theme.accent, background = theme.bar, bold = true },
        }
    end
})

-- Grid: uses selector so it only re-renders when clickCount changes
-- (not on every hover, which only affects individual Cell components)
local Grid = lumina.defineComponent({
    name = "Grid",
    render = function()
        local state = lumina.useStore(store, function(s)
            return s.clickCount
        end)

        local width, height = lumina.getSize()
        local cols = width
        local rows = height - 1

        local children = {}
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

-- App: just a vbox with StatusBar + Grid
local App = lumina.defineComponent({
    name = "StressApp",
    render = function()
        return {
            type = "vbox",
            style = { background = theme.bg },
            children = {
                lumina.createElement(StatusBar, { key = "status" }),
                lumina.createElement(Grid, { key = "grid" }),
            }
        }
    end
})

lumina.mount(App)
lumina.run()
