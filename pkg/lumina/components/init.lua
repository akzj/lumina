-- Component library initialization
local lumina = require("lumina")

-- Load components
local Button = dofile(lumina._getComponentPath and lumina._getComponentPath("button") or "pkg/lumina/components/button.lua")
local Input = dofile(lumina._getComponentPath and lumina._getComponentPath("input") or "pkg/lumina/components/input.lua")
local Dialog = dofile(lumina._getComponentPath and lumina._getComponentPath("dialog") or "pkg/lumina/components/dialog.lua")

-- Alternative: inline definitions if dofile fails
if not Button then
    Button = lumina.defineComponent({
        name = "Button",
        init = function(props)
            return {
                label = props.label or "Button",
                variant = props.variant or "default",
                disabled = props.disabled or false
            }
        end,
        render = function(instance)
            local theme = lumina.hooks.useTheme and lumina.hooks.useTheme() or {}
            local colors = theme.colors or {primary="cyan", text="white"}
            local color = instance.disabled and "gray" or colors.primary
            return {
                type = "hbox",
                children = {
                    {
                        type = "text",
                        content = "[ " .. instance.label .. " ]",
                        color = color,
                        bold = not instance.disabled
                    }
                }
            }
        end
    })
end

return {
    Button = Button,
    Input = Input,
    Dialog = Dialog
}
