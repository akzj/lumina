-- shadcn/select — Dropdown select with overlay popup
local lumina = require("lumina")

local Select = lumina.defineComponent({
    name = "ShadcnSelect",
    init = function(props)
        return {
            options = props.options or {},
            value = props.value or "",
            placeholder = props.placeholder or "Select...",
            disabled = props.disabled or false,
            open = false,
            selectedIndex = 0,
        }
    end,
    render = function(self)
        -- Display text
        local display = self.placeholder
        for i, opt in ipairs(self.options) do
            local val = type(opt) == "table" and opt.value or opt
            local label = type(opt) == "table" and (opt.label or opt.value) or opt
            if val == self.value then
                display = label
                break
            end
        end

        local fg = self.value == "" and "#64748B" or "#E2E8F0"
        if self.disabled then fg = "#475569" end

        return {
            type = "hbox",
            style = {
                border = "rounded",
                foreground = fg,
                padding = 1,
                height = 3,
            },
            children = {
                { type = "text", content = display },
                { type = "text", content = " ▾", style = { foreground = "#64748B" } },
            },
        }
    end
})

return Select
