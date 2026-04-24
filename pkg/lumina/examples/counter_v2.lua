-- Counter with Button component demo
local lumina = require("lumina")

-- Define Button component
local Button = lumina.defineComponent({
    name = "Button",
    
    init = function(props)
        return {
            label = props.label or "Button",
            variant = props.variant or "default",
            onClick = props.onClick,
            disabled = props.disabled or false
        }
    end,
    
    render = function(instance)
        local theme = lumina.hooks.useTheme and lumina.hooks.useTheme() or {}
        local colors = theme.colors or {primary="cyan", text="white"}
        
        local prefix, suffix
        if instance.disabled then
            prefix, suffix = "[ ", " ]"
        else
            prefix, suffix = "[ ", " ]"
        end
        
        return {
            type = "hbox",
            children = {
                {
                    type = "text",
                    content = prefix .. instance.label .. suffix,
                    color = instance.disabled and "gray" or colors.primary,
                    bold = not instance.disabled
                }
            },
            -- Store click handler
            _onClick = instance.onClick,
            _disabled = instance.disabled
        }
    end
})

-- Define Counter component with buttons
local Counter = lumina.defineComponent({
    name = "Counter",
    
    init = function(props)
        local count, setCount = lumina.useState(props.initial or 0)
        return { count = count, setCount = setCount }
    end,
    
    render = function(instance)
        return {
            type = "vbox",
            children = {
                { type = "text", content = "╔══════════════════════╗" },
                { type = "text", content = "║   Counter App v0.2   ║" },
                { type = "text", content = "╠══════════════════════╣" },
                {
                    type = "text",
                    content = "║  Count: " .. string.format("%-10d", instance.count) .. "║"
                },
                { type = "text", content = "╠══════════════════════╣" },
                {
                    type = "hbox",
                    children = {
                        {
                            type = "button",
                            label = "-1",
                            onClick = function()
                                instance.setCount(instance.count - 1)
                            end
                        },
                        { type = "text", content = " " },
                        {
                            type = "button",
                            label = "Reset",
                            onClick = function()
                                instance.setCount(0)
                            end
                        },
                        { type = "text", content = " " },
                        {
                            type = "button",
                            label = "+1",
                            onClick = function()
                                instance.setCount(instance.count + 1)
                            end
                        }
                    }
                },
                { type = "text", content = "╚══════════════════════╝" }
            }
        }
    end
})

print("Lumina Counter with Buttons")
print("============================")
lumina.render(Counter, { initial = 0 })
print()
print("Components loaded: Button, Input, Dialog")
