-- Lumina Mouse Test: Full-screen character canvas for mouse debugging
-- Hover: cell under cursor turns green █
-- Click: cell turns red █ (persists)
-- Shows coordinates at top
local lumina = require("lumina")

local theme = {
    bg = "#1E1E2E",
    dot = "#585B70",
    hover = "#A6E3A1",
    click = "#F38BA8",
    both = "#F9E2AF",
    accent = "#89B4FA",
    muted = "#6C7086",
    text = "#CDD6F4",
    bar = "#181825",
}

-- Global state via store
local store = lumina.createStore({
    state = {
        hoverX = -1,
        hoverY = -1,
        clickX = -1,
        clickY = -1,
        totalClicks = 0,
    }
})

-- Clicked cells: "x,y" -> true
local clickedCells = {}

-- Register global mouse events (empty componentID = wildcard, matches all)
lumina.on("mousemove", "", function(e)
    store.dispatch("setState", { hoverX = e.x, hoverY = e.y })
end)

lumina.on("mousedown", "", function(e)
    local key = e.x .. "," .. e.y
    clickedCells[key] = true
    local st = store.getState()
    store.dispatch("setState", {
        clickX = e.x,
        clickY = e.y,
        totalClicks = (st.totalClicks or 0) + 1,
    })
end)

lumina.onKey("c", function()
    clickedCells = {}
    store.dispatch("setState", { totalClicks = 0, clickX = -1, clickY = -1 })
end)

lumina.onKey("q", function() lumina.quit() end)

-- Build a single row as text segments (optimized: accumulate same-style chars)
local function buildRow(y, width, hoverX, hoverY)
    -- Check if this row has any special cells
    local isHoverRow = (y == hoverY)
    local hasClicks = false
    if not isHoverRow then
        for x = 0, width - 1 do
            if clickedCells[x .. "," .. y] then
                hasClicks = true
                break
            end
        end
    end

    -- Fast path: uniform row (most rows)
    if not isHoverRow and not hasClicks then
        return {
            type = "text",
            content = string.rep(".", width),
            style = { foreground = theme.dot, background = theme.bg },
        }
    end

    -- Slow path: build segments for rows with special cells
    local segments = {}
    local curChars = {}
    local curFg = theme.dot
    local curBg = theme.bg

    local function flush()
        if #curChars > 0 then
            segments[#segments + 1] = {
                type = "text",
                content = table.concat(curChars),
                style = { foreground = curFg, background = curBg },
            }
            curChars = {}
        end
    end

    for x = 0, width - 1 do
        local key = x .. "," .. y
        local isHover = (x == hoverX and y == hoverY)
        local isClick = clickedCells[key]

        local char = "."
        local fg = theme.dot
        local bg = theme.bg

        if isHover and isClick then
            char = "#"
            fg = theme.both
            bg = "#45475A"
        elseif isHover then
            char = "@"
            fg = theme.hover
            bg = "#313244"
        elseif isClick then
            char = "*"
            fg = theme.click
            bg = "#313244"
        end

        -- If style changed, flush current segment
        if fg ~= curFg or bg ~= curBg then
            flush()
            curFg = fg
            curBg = bg
        end
        curChars[#curChars + 1] = char
    end
    flush()

    if #segments == 1 then
        return segments[1]
    end
    return { type = "hbox", children = segments }
end

local App = lumina.defineComponent({
    name = "MouseTest",
    render = function()
        local state = lumina.useStore(store)
        local hx = state.hoverX or -1
        local hy = state.hoverY or -1
        local cx = state.clickX or -1
        local cy = state.clickY or -1
        local total = state.totalClicks or 0

        -- Get actual terminal size
        local width, height = lumina.getSize()

        -- Status bar (row 0)
        local status = string.format(
            " Mouse: (%d, %d) | Clicked: (%d, %d) | Total: %d | [c] Clear  [q] Quit",
            hx, hy, cx, cy, total
        )

        local children = {
            {
                type = "text",
                content = status,
                style = { foreground = theme.accent, background = theme.bar, bold = true },
            },
        }

        for y = 1, height - 1 do
            children[#children + 1] = buildRow(y, width, hx, hy)
        end

        return {
            type = "vbox",
            style = { background = theme.bg },
            children = children,
        }
    end,
})

lumina.mount(App)
lumina.run()
