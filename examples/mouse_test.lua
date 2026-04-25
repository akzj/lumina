-- Lumina Mouse Grid Test: per-component grid with independent hover/click
-- Each cell is a box with unique ID, uses lumina.isHovered() and e.target
-- Hover: green [O], Click: red [X], Both: yellow [*]
local lumina = require("lumina")

local store = lumina.createStore({
    state = {
        clicked = {},    -- id -> true
        lastClick = "",
        hoverTarget = "",
    }
})

-- Global mousemove handler — track hover target
lumina.on("mousemove", "", function(e)
    store.dispatch("setState", { hoverTarget = e.target or "" })
end)

-- Global click handler — reads e.target to know which cell
lumina.on("mousedown", "", function(e)
    if e.target and e.target ~= "" then
        local st = store.getState()
        local clicked = st.clicked or {}
        clicked[e.target] = true
        store.dispatch("setState", {
            clicked = clicked,
            lastClick = e.target,
        })
    end
end)

lumina.onKey("c", function()
    store.dispatch("setState", { clicked = {}, lastClick = "" })
end)
lumina.onKey("q", function() lumina.quit() end)

local App = lumina.defineComponent({
    name = "MouseGrid",
    render = function()
        local state = lumina.useStore(store)
        local width, height = lumina.getSize()

        -- Each cell is 3 chars wide
        local cols = math.floor(width / 3)
        local rows = height - 1  -- reserve 1 row for status

        local children = {}

        -- Status bar
        children[1] = {
            type = "text",
            content = string.format(" Hover: %s | Click: %s | [c] Clear [q] Quit",
                state.hoverTarget or "none", state.lastClick or "none"),
            style = { foreground = "#89B4FA", background = "#181825", bold = true },
        }

        -- Grid rows
        for y = 0, rows - 1 do
            local rowChildren = {}
            for x = 0, cols - 1 do
                local id = "c-" .. x .. "-" .. y
                local isHover = lumina.isHovered(id)
                local isClick = (state.clicked or {})[id]

                local char = " . "
                local fg = "#585B70"
                local bg = "#1E1E2E"

                if isHover and isClick then
                    char = "[*]"
                    fg = "#F9E2AF"
                    bg = "#45475A"
                elseif isHover then
                    char = "[O]"
                    fg = "#A6E3A1"
                    bg = "#313244"
                elseif isClick then
                    char = "[X]"
                    fg = "#F38BA8"
                    bg = "#313244"
                end

                rowChildren[#rowChildren + 1] = {
                    type = "box",
                    props = { id = id },
                    style = { width = 3, height = 1, background = bg },
                    children = {
                        { type = "text", content = char, style = { foreground = fg } }
                    }
                }
            end

            children[#children + 1] = {
                type = "hbox",
                children = rowChildren,
            }
        end

        return {
            type = "vbox",
            style = { background = "#1E1E2E" },
            children = children,
        }
    end,
})

lumina.mount(App)
lumina.run()
