-- shadcn/progress — Progress bar
local lumina = require("lumina")

local c = {
    muted = "#313244",
    primary = "#89B4FA",
    fg = "#CDD6F4",
}

local Progress = lumina.defineComponent({
    name = "ShadcnProgress",

    init = function(props)
        return {
            value = props.value or 0,
            max = props.max or 100,
            width = props.width or 20,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local pct = self.max > 0 and math.min(1, math.max(0, self.value / self.max)) or 0
        local filled = math.floor(pct * self.width)
        local empty = self.width - filled

        local track = string.rep("█", filled) .. string.rep("░", empty)

        local style = {
            foreground = c.primary,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "hbox",
            id = self.id,
            style = { align = "center" },
            children = {
                {
                    type = "text",
                    content = "[" .. track .. "]",
                    style = style,
                },
                {
                    type = "text",
                    content = " " .. math.floor(pct * 100) .. "%",
                    style = { foreground = c.fg },
                },
            },
        }
    end,
})

return Progress