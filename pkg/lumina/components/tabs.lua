-- Tabs component for Lumina — tab switching
local lumina = require("lumina")

local Tabs = lumina.defineComponent({
    name = "Tabs",
    init = function(props)
        return {
            tabs = props.tabs or {},
            activeTab = props.activeTab or 1,
            onTabChange = props.onTabChange,
            style = props.style or {},
        }
    end,
    render = function(instance)
        local tabs = instance.tabs or {}
        local active = instance.activeTab or 1
        local bg = instance.style and instance.style.background
        local children = {}
        for i, tab in ipairs(tabs) do
            local isActive = (i == active)
            local label = isActive and ("[ " .. tostring(tab) .. " ]") or ("  " .. tostring(tab) .. "  ")
            children[#children + 1] = {
                type = "text",
                content = label,
                foreground = isActive and "#00FFFF" or "#888888",
                bold = isActive,
                background = bg,
            }
        end
        return { type = "hbox", style = { background = bg }, children = children }
    end
})

return Tabs
