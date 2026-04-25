-- shadcn/input_otp — One-time password input (separate digit boxes)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
}

local InputOTP = lumina.defineComponent({
    name = "ShadcnInputOTP",

    init = function(props)
        return {
            length = props.length or 6,
            value = props.value or "",
            mask = props.mask or false,
            disabled = props.disabled or false,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}
        for i = 1, self.length do
            local ch = ""
            if i <= #self.value then
                ch = self.mask and "●" or string.sub(self.value, i, i)
            end
            local filled = ch ~= ""

            children[#children + 1] = {
                type = "text",
                content = "[" .. (filled and ch or " ") .. "]",
                style = {
                    foreground = self.disabled and c.muted or (filled and c.fg or c.muted),
                    bold = filled,
                },
            }

            if i < self.length and i % 3 == 0 then
                children[#children + 1] = {
                    type = "text",
                    content = " - ",
                    style = { foreground = c.muted },
                }
            end
        end

        return {
            type = "hbox",
            id = self.id,
            children = children,
        }
    end,
})

return InputOTP
