-- lua/lux/spinner.lua
-- Spinner component for Lux
-- Usage: local Spinner = require("lux.spinner")

local Spinner = lumina.defineComponent("Spinner", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local frames = {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    local frame, setFrame = lumina.useState("frame", 1)
    local label = props.label or "Loading..."

    lumina.useEffect("spin", function()
        local timer = lumina.setInterval(function()
            setFrame(function(f) return (f % #frames) + 1 end)
        end, 80)
        return function() lumina.clearInterval(timer) end
    end, {})

    return lumina.createElement("hbox", {},
        lumina.createElement("text", {
            foreground = t.primary or "#89B4FA",
        }, frames[frame] .. " "),
        lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
        }, label)
    )
end)

return Spinner
