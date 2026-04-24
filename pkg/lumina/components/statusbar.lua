-- StatusBar component for Lumina — bottom status bar
local lumina = require("lumina")

local StatusBar = lumina.defineComponent({
    name = "StatusBar",
    init = function(props)
        return {
            left = props.left or "",
            center = props.center or "",
            right = props.right or "",
            style = props.style or {},
        }
    end,
    render = function(instance)
        local fg = instance.style and instance.style.foreground or "#FFFFFF"
        local bg = instance.style and instance.style.background or "#333333"
        local children = {}
        children[#children + 1] = {
            type = "text", content = " " .. (instance.left or ""),
            foreground = fg, background = bg, style = { flex = 1 },
        }
        if instance.center and instance.center ~= "" then
            children[#children + 1] = {
                type = "text", content = instance.center,
                foreground = fg, background = bg, bold = true,
            }
        end
        children[#children + 1] = {
            type = "text", content = (instance.right or "") .. " ",
            foreground = fg, background = bg, style = { flex = 1 },
        }
        return { type = "hbox", style = { height = 1, background = bg }, children = children }
    end
})

return StatusBar
