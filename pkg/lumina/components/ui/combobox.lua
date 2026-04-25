-- shadcn/combobox — Searchable select (input + dropdown)
local lumina = require("lumina")

local Combobox = lumina.defineComponent({
    name = "ShadcnCombobox",
    init = function(props)
        return {
            options = props.options or {},
            value = props.value or "",
            search = props.search or "",
            placeholder = props.placeholder or "Search...",
            disabled = props.disabled or false,
            open = false,
        }
    end,
    render = function(self)
        -- Display: selected value or search text
        local display = self.value ~= "" and self.value or self.search
        if display == "" then display = self.placeholder end
        local fg = (self.value == "" and self.search == "") and "#64748B" or "#E2E8F0"
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
                { type = "text", content = "🔍 " .. display },
                { type = "text", content = " ▾", style = { foreground = "#64748B" } },
            },
        }
    end
})

return Combobox
