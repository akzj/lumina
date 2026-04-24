-- shadcn/input_otp — One-time password input (separate digit boxes)
local lumina = require("lumina")

local InputOTP = lumina.defineComponent({
    name = "ShadcnInputOTP",
    init = function(props)
        return {
            length = props.length or 6,
            value = props.value or "",
            mask = props.mask or false,
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
                    foreground = filled and "#E2E8F0" or "#475569",
                    bold = filled,
                },
            }
            -- Add separator every 3 digits
            if i < self.length and i % 3 == 0 then
                children[#children + 1] = {
                    type = "text",
                    content = " - ",
                    style = { foreground = "#475569" },
                }
            end
        end
        return {
            type = "hbox",
            children = children,
        }
    end
})

return InputOTP
