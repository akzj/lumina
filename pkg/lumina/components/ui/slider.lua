-- shadcn/slider — Value slider with track and thumb
local lumina = require("lumina")

local c = {
    muted = "#6C7086",
    primary = "#89B4FA",
    fg = "#CDD6F4",
}

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
            id = props.id,
            className = props.className,
            style = props.style,
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

        local children = {
            { type = "text", content = "[" .. track .. "]", style = { foreground = self.disabled and c.muted or c.primary } },
        }
        if self.showValue then
            children[#children + 1] = {
                type = "text",
                content = " " .. tostring(self.value),
                style = { foreground = c.muted },
            }
        end

        return {
            type = "hbox",
            id = self.id,
            style = { align = "center" },
            children = children,
        }
    end,
})

return Slider
