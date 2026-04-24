-- Spinner component for Lumina — loading animation
local lumina = require("lumina")

local spinnerFrames = {
    dots    = {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
    line    = {"-", "\\", "|", "/"},
    arc     = {"◜", "◠", "◝", "◞", "◡", "◟"},
    braille = {"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
}

local Spinner = lumina.defineComponent({
    name = "Spinner",
    init = function(props)
        return {
            spinnerStyle = props.spinnerStyle or "dots",
            label = props.label or "",
            frame = props.frame or 1,
            style = props.style or {},
        }
    end,
    render = function(instance)
        local styleName = instance.spinnerStyle or "dots"
        local frames = spinnerFrames[styleName] or spinnerFrames.dots
        local frame = ((instance.frame or 1) - 1) % #frames + 1
        local spinChar = frames[frame]
        local fg = instance.style and instance.style.foreground or "#00FFFF"
        local content = spinChar
        if instance.label and instance.label ~= "" then
            content = content .. " " .. instance.label
        end
        return { type = "text", content = content, foreground = fg }
    end
})

return Spinner
