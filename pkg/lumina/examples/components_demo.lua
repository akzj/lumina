-- Components Demo: Button, Input, Dialog
local lumina = require("lumina")

print("=== Lumina Components Demo ===")
print()

-- Button component
local Button = lumina.defineComponent({
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
        local colors = theme.colors or {primary="cyan", error="red", success="green"}
        
        local color
        if instance.disabled then
            color = "gray"
        elseif instance.variant == "danger" then
            color = "error"
        elseif instance.variant == "success" then
            color = "success"
        else
            color = colors.primary
        end
        
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

-- Dialog component
local Dialog = lumina.defineComponent({
    name = "Dialog",
    init = function(props)
        return {
            title = props.title or "Dialog",
            content = props.content or "",
            buttons = props.buttons or {{label = "OK"}},
            visible = props.visible ~= false
        }
    end,
    render = function(instance)
        if not instance.visible then
            return { type = "box", children = {} }
        end
        
        local theme = lumina.hooks.useTheme and lumina.hooks.useTheme() or {}
        local colors = theme.colors or {primary="cyan", text="white", secondary="gray"}
        
        local width = math.max(#instance.title + 4, #instance.content + 4)
        for _, btn in ipairs(instance.buttons) do
            width = math.max(width, #btn.label + 6)
        end
        
        local border = "─"
        local hLine = string.rep(border, width - 2)
        
        local function pad(str)
            if #str >= width - 2 then
                return str:sub(1, width - 2)
            end
            return str .. string.rep(" ", width - 2 - #str)
        end
        
        local buttonRow = ""
        for i, btn in ipairs(instance.buttons) do
            if i > 1 then buttonRow = buttonRow .. " " end
            buttonRow = buttonRow .. "[ " .. btn.label .. " ]"
        end
        
        return {
            type = "vbox",
            children = {
                {
                    type = "text",
                    content = "┌" .. hLine .. "┐",
                    color = colors.secondary
                },
                {
                    type = "text",
                    content = "│ " .. pad(instance.title) .. " │",
                    color = colors.primary,
                    bold = true
                },
                {
                    type = "text",
                    content = "├" .. hLine .. "┤",
                    color = colors.secondary
                },
                {
                    type = "text",
                    content = "│ " .. pad(instance.content) .. " │",
                    color = colors.text
                },
                {
                    type = "text",
                    content = "├" .. hLine .. "┤",
                    color = colors.secondary
                },
                {
                    type = "text",
                    content = "│ " .. pad(buttonRow) .. " │",
                    color = colors.primary
                },
                {
                    type = "text",
                    content = "└" .. hLine .. "┘",
                    color = colors.secondary
                }
            }
        }
    end
})

-- Input component
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
            placeholder = props.placeholder or ""
        }
    end,
    render = function(instance)
        local theme = lumina.hooks.useTheme and lumina.hooks.useTheme() or {}
        local colors = theme.colors or {primary="cyan", text="white", secondary="gray"}
        
        local displayValue = instance.value
        if displayValue == "" and instance.placeholder ~= "" then
            displayValue = instance.placeholder
        end
        
        local cursor = instance.focused and "█" or "_"
        
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
                    color = instance.value == "" and colors.secondary or colors.text,
                    dim = instance.value == ""
                },
                {
                    type = "text",
                    content = cursor,
                    color = colors.primary
                }
            }
        }
    end
})

-- Demo app using all components
local App = lumina.defineComponent({
    name = "App",
    init = function(props)
        return {}
    end,
    render = function(instance)
        return {
            type = "vbox",
            children = {
                { type = "text", content = "═══ Components Demo ═══" },
                { type = "text", content = "" },
                { type = "text", content = "Button variants:" },
                {
                    type = "hbox",
                    children = {
                        { type = "button", label = "Default" },
                        { type = "text", content = " " },
                        { type = "button", label = "Danger", variant = "danger" },
                        { type = "text", content = " " },
                        { type = "button", label = "Success", variant = "success" }
                    }
                },
                { type = "text", content = "" },
                { type = "text", content = "Disabled button:" },
                { type = "button", label = "Disabled", disabled = true },
                { type = "text", content = "" },
                { type = "text", content = "Input field:" },
                { type = "input", placeholder = "Type here..." },
                { type = "text", content = "" },
                { type = "text", content = "Dialog:" },
                {
                    type = "dialog",
                    title = "Welcome",
                    content = "Lumina v0.2 is ready!",
                    buttons = {{label = "OK"}, {label = "Cancel"}}
                }
            }
        }
    end
})

lumina.render(App)
print()
print("Components demo complete!")
