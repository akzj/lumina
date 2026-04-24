-- Dialog component for Lumina
local lumina = require("lumina")

local Dialog = lumina.defineComponent({
    name = "Dialog",
    
    init = function(props)
        return {
            title = props.title or "Dialog",
            content = props.content or "",
            buttons = props.buttons or {{label = "OK"}},
            visible = props.visible ~= false,
            onClose = props.onClose
        }
    end,
    
    render = function(instance)
        if not instance.visible then
            return { type = "box" }
        end
        
        local theme = lumina.hooks.useTheme and lumina.hooks.useTheme() or {}
        local colors = theme.colors or {primary="cyan", text="white", secondary="gray"}
        
        -- Calculate dimensions
        local maxWidth = #instance.title + 4
        for _, btn in ipairs(instance.buttons) do
            maxWidth = math.max(maxWidth, #btn.label + 6)
        end
        maxWidth = math.max(maxWidth, #instance.content + 4)
        
        -- Create border line
        local function borderLine(char)
            return char .. string.rep("─", maxWidth - 2) .. char
        end
        
        -- Pad string to width
        local function pad(str, width)
            if #str >= width then
                return str:sub(1, width)
            end
            return str .. string.rep(" ", width - #str)
        end
        
        -- Build button row
        local buttonRow = ""
        for i, btn in ipairs(instance.buttons) do
            if i > 1 then
                buttonRow = buttonRow .. " "
            end
            buttonRow = buttonRow .. "[ " .. btn.label .. " ]"
        end
        
        -- Center content
        local contentPadded = "│ " .. pad(instance.content, maxWidth - 4) .. " │"
        local titlePadded = "│ " .. pad(" " .. instance.title .. " ", maxWidth - 4) .. "│"
        local buttonsPadded = "│ " .. pad(buttonRow, maxWidth - 4) .. " │"
        
        return {
            type = "vbox",
            children = {
                -- Top border
                {
                    type = "text",
                    content = "┌" .. borderLine("─"):sub(2) .. "┐",
                    color = colors.secondary
                },
                -- Title
                {
                    type = "text",
                    content = titlePadded,
                    color = colors.primary,
                    bold = true
                },
                -- Separator
                {
                    type = "text",
                    content = "├" .. borderLine("─"):sub(2) .. "┤",
                    color = colors.secondary
                },
                -- Content
                {
                    type = "text",
                    content = contentPadded,
                    color = colors.text
                },
                -- Separator
                {
                    type = "text",
                    content = "├" .. borderLine("─"):sub(2) .. "┤",
                    color = colors.secondary
                },
                -- Buttons
                {
                    type = "text",
                    content = buttonsPadded,
                    color = colors.primary
                },
                -- Bottom border
                {
                    type = "text",
                    content = "└" .. borderLine("─"):sub(2) .. "┘",
                    color = colors.secondary
                }
            }
        }
    end
})

return Dialog
