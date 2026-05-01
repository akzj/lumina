-- lua/lux/spinner.lua
-- Spinner component for Lux
-- Usage: local Spinner = require("lux.spinner")

local Spinner = lumina.defineComponent("Spinner", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local frames = {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    local frame, setFrame = lumina.useState("frame", 1)
    local label = props.label or "Loading..."

    lumina.useEffect(function()
        local currentFrame = frame
        local timer = lumina.setInterval(function()
            currentFrame = (currentFrame % #frames) + 1
            setFrame(currentFrame)
        end, 80)
        return function() lumina.clearInterval(timer) end
    end, {})

    return lumina.createElement("hbox", {},
        lumina.createElement("text", {
            foreground = t.primary or "#F5C842",
        }, frames[frame] .. " "),
        lumina.createElement("text", {
            foreground = t.text or "#E8EDF7",
        }, label)
    )
end)

return Spinner
