-- shadcn/combobox — Searchable select (input + dropdown)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
}

local Combobox = lumina.defineComponent({
    name = "ShadcnCombobox",

    init = function(props)
        return {
            options = props.options or {},
            value = props.value or "",
            placeholder = props.placeholder or "Search...",
            open = props.open or false,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        -- Input display
        local display = self.value ~= "" and self.value or self.placeholder
        local fg = self.value == "" and c.muted or c.fg

        local children = {
            {
                type = "text",
                content = display,
                style = { foreground = fg },
            },
        }

        -- Dropdown when open
        if self.open then
            for i, opt in ipairs(self.options) do
                local label = type(opt) == "table" and (opt.label or opt) or opt
                local val = type(opt) == "table" and opt.value or opt
                local selected = val == self.value
                children[#children + 1] = {
                    type = "text",
                    content = (selected and "● " or "  ") .. tostring(label),
                    style = { foreground = selected and c.primary or c.muted },
                }
            end
        end

        return {
            type = "vbox",
            id = self.id,
            style = self.style or {},
            children = children,
        }
    end,
})

return Combobox