-- Button component for Lumina
local lumina = require("lumina")

local Button = lumina.defineComponent({
    name = "Button",
    
    init = function(props)
        return {
            label = props.label or "Button",
            variant = props.variant or "default",
            onClick = props.onClick,
            disabled = props.disabled or false,
            variant_ = props.variant_ or "default"
        }
    end,
    
    render = function(instance)
        local theme = lumina.hooks.useTheme and lumina.hooks.useTheme() or {}
        local colors = theme.colors or {primary="cyan", text="white", secondary="gray"}
        
        -- Style based on variant
        local prefix, suffix, color
        if instance.disabled then
            prefix, suffix = "[ ", " ]"
            color = "gray"
        elseif instance.variant_ == "danger" then
            prefix, suffix = "[ ", " ]"
            color = "error"
        elseif instance.variant_ == "success" then
            prefix, suffix = "[ ", " ]"
            color = "success"
        else
            prefix, suffix = "[ ", " ]"
            color = colors.primary
        end
        
        return {
            type = "hbox",
            children = {
                {
                    type = "text",
                    content = prefix .. instance.label .. suffix,
                    color = color,
                    bold = not instance.disabled
                }
            }
        }
    end
})

return Button
