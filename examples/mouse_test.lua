-- Lumina Mouse Grid Test: React declarative style
-- Each cell is an independent component with its own useState for hover/click.
-- Uses createElement + defineComponent + onMouseEnter/onMouseLeave/onClick props.
-- No global store, no lumina.on(), no lumina.isHovered() — pure React pattern.
local lumina = require("lumina")

local theme = {
    bg = "#1E1E2E", dot = "#585B70", hover = "#A6E3A1",
    click = "#F38BA8", both = "#F9E2AF", bar = "#181825",
    accent = "#89B4FA",
}

-- ============================================================
-- Cell: independent component with own hover/click state
-- ============================================================
local Cell = lumina.defineComponent({
    name = "Cell",
    render = function(props)
        local hovered, setHovered = lumina.useState("hovered", false)
        local clicked, setClicked = lumina.useState("clicked", false)

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

        return {
            type = "box",
            id = props.id,
            style = { width = 3, height = 1, background = bg },
            -- Declarative event props — React style
            onClick = function()
                setClicked(not clicked)
                if props.onCellClick then
                    props.onCellClick(props.id, not clicked)
                end
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
        local lastHover, setLastHover = lumina.useState("lastHover", "")
        local lastClick, setLastClick = lumina.useState("lastClick", "")

        local width, height = lumina.getSize()
        -- Reduce grid for performance (each cell is an independent component)
        local cols = math.min(math.floor(width / 3), 30)
        local rows = math.min(height - 1, 15)

        local children = {}

        -- Status bar
        children[1] = {
            type = "text",
            content = string.format(
                " Hover: %s | Click: %s | [q] Quit",
                lastHover, lastClick),
            style = { foreground = theme.accent, background = theme.bar, bold = true },
        }

        -- Grid of Cell components via createElement
        for y = 0, rows - 1 do
            local rowChildren = {}
            for x = 0, cols - 1 do
                local cellId = "c-" .. x .. "-" .. y
                rowChildren[#rowChildren + 1] = lumina.createElement(Cell, {
                    id = cellId,
                    key = cellId,
                    onCellHover = function(id) setLastHover(id) end,
                    onCellClick = function(id, isClicked)
                        if isClicked then setLastClick(id) end
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

lumina.onKey("q", function() lumina.quit() end)
lumina.mount(App)
lumina.run()
