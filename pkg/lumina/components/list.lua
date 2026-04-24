-- List component for Lumina — scrollable list with selection
local lumina = require("lumina")

local List = lumina.defineComponent({
    name = "List",
    init = function(props)
        return {
            items = props.items or {},
            selectedIndex = props.selectedIndex or 1,
            scrollOffset = props.scrollOffset or 0,
            onSelect = props.onSelect,
            renderItem = props.renderItem,
            style = props.style or {},
        }
    end,
    render = function(instance)
        local items = instance.items or {}
        local selected = instance.selectedIndex or 1
        local height = (instance.style and instance.style.height) or #items
        local fg = instance.style and instance.style.foreground or "#FFFFFF"
        local children = {}
        for i, item in ipairs(items) do
            local isSelected = (i == selected)
            local prefix = isSelected and "▸ " or "  "
            children[#children + 1] = {
                type = "text",
                content = prefix .. tostring(item),
                foreground = isSelected and "#00FF00" or fg,
                bold = isSelected,
            }
        end
        return {
            type = "vbox",
            style = { height = height, overflow = (#items > height) and "scroll" or "hidden" },
            children = children,
        }
    end
})

return List
