-- Input component for Lumina
local lumina = require("lumina")

local Input = lumina.defineComponent({
    name = "Input",
    
    init = function(props)
        local value, setValue = lumina.useState(props.value or "")
        local focused, setFocused = lumina.useState(false)
        
        return {
            value = value,
            setValue = setValue,
            focused = focused,
            setFocused = setFocused,
            placeholder = props.placeholder or "",
            maxLength = props.maxLength or 100,
            onChange = props.onChange,
            onSubmit = props.onSubmit,
            type = props.type or "text",
            password = props.password or false
        }
    end,
    
    render = function(instance)
        local theme = lumina.hooks.useTheme and lumina.hooks.useTheme() or {}
        local colors = theme.colors or {primary="cyan", text="white", secondary="gray"}
        
        -- Determine display value
        local displayValue
        if instance.password then
            displayValue = string.rep("*", #instance.value)
        else
            displayValue = instance.value
        end
        
        -- Show placeholder if empty
        if displayValue == "" and instance.placeholder ~= "" then
            displayValue = instance.placeholder
        end
        
        -- Colors based on state
        local cursorColor = instance.focused and colors.primary or colors.secondary
        local textColor
        if instance.value == "" and instance.placeholder ~= "" then
            textColor = colors.secondary
        else
            textColor = colors.text
        end
        
        -- Calculate cursor position
        local cursor = "_"
        if instance.focused then
            cursor = "█"
        end
        
        return {
            type = "hbox",
            children = {
                {
                    type = "text",
                    content = instance.focused and "> " or "  ",
                    color = colors.primary
                },
                {
                    type = "text",
                    content = displayValue,
                    color = textColor,
                    dim = instance.value == ""
                },
                {
                    type = "text",
                    content = cursor,
                    color = cursorColor
                }
            }
        }
    end
})

return Input
