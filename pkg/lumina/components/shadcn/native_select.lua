-- shadcn/native_select — Simple inline select (no overlay)
local lumina = require("lumina")

local NativeSelect = lumina.defineComponent({
    name = "ShadcnNativeSelect",
    init = function(props)
        return {
            options = props.options or {},
            value = props.value or "",
            disabled = props.disabled or false,
        }
    end,
    render = function(self)
        local display = self.value
        if display == "" and #self.options > 0 then
            local first = self.options[1]
            display = type(first) == "table" and (first.label or first.value) or first
        end
        local fg = self.disabled and "#475569" or "#E2E8F0"

        return {
            type = "hbox",
            style = {
                border = "rounded",
                foreground = fg,
                padding = 1,
                height = 3,
            },
            children = {
                { type = "text", content = "◄ " .. tostring(display) .. " ►" },
            },
        }
    end
})

return NativeSelect
