-- shadcn/slider — Value slider with track and thumb
local lumina = require("lumina")

local Slider = lumina.defineComponent({
    name = "ShadcnSlider",
    init = function(props)
        return {
            value = props.value or 50,
            min = props.min or 0,
            max = props.max or 100,
            step = props.step or 1,
            width = props.width or 20,
            disabled = props.disabled or false,
            showValue = props.showValue or false,
        }
    end,
    render = function(self)
        local range = self.max - self.min
        local pct = range > 0 and ((self.value - self.min) / range) or 0
        pct = math.max(0, math.min(1, pct))
        local pos = math.floor(pct * (self.width - 1))

        local track = ""
        for i = 0, self.width - 1 do
            if i == pos then
                track = track .. "●"
            elseif i < pos then
                track = track .. "━"
            else
                track = track .. "─"
            end
        end

        local fg = self.disabled and "#475569" or "#3B82F6"
        local children = {
            { type = "text", content = "[" .. track .. "]", style = { foreground = fg } },
        }
        if self.showValue then
            children[#children + 1] = {
                type = "text",
                content = " " .. tostring(self.value),
                style = { foreground = "#94A3B8" },
            }
        end
        return {
            type = "hbox",
            children = children,
        }
    end
})

return Slider
