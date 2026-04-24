-- Modal component for Lumina — modal dialog with overlay
local lumina = require("lumina")

local Modal = lumina.defineComponent({
    name = "Modal",
    init = function(props)
        return {
            visible = props.visible ~= false,
            title = props.title or "Dialog",
            children = props.children or {},
            onClose = props.onClose,
            buttons = props.buttons or {},
            style = props.style or {},
        }
    end,
    render = function(instance)
        if not instance.visible then return { type = "box" } end
        local title = instance.title or "Dialog"
        local fg = instance.style and instance.style.foreground or "#FFFFFF"
        local bg = instance.style and instance.style.background or "#222222"
        local bodyChildren = {}
        bodyChildren[#bodyChildren + 1] = {
            type = "text", content = " " .. title .. " ",
            foreground = "#00FFFF", bold = true,
        }
        for _, child in ipairs(instance.children or {}) do
            bodyChildren[#bodyChildren + 1] = child
        end
        local buttonChildren = {}
        for _, btn in ipairs(instance.buttons or {}) do
            buttonChildren[#buttonChildren + 1] = {
                type = "text", content = "[ " .. (btn.label or "OK") .. " ]",
                foreground = (btn.style and btn.style.foreground) or fg, bold = true,
            }
        end
        if #buttonChildren > 0 then
            bodyChildren[#bodyChildren + 1] = { type = "hbox", style = { gap = 2 }, children = buttonChildren }
        end
        return {
            type = "vbox",
            style = { border = "rounded", padding = 1, background = bg },
            children = bodyChildren,
        }
    end
})

return Modal
