-- shadcn/select — Dropdown select display
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    border = "#45475A",
    bg = "#1E1E2E",
}

local Select = lumina.defineComponent({
    name = "ShadcnSelect",

    init = function(props)
        return {
            options = props.options or {},
            value = props.value or "",
            placeholder = props.placeholder or "Select...",
            disabled = props.disabled or false,
            id = props.id,
            className = props.className,
            style = props.style,
            onOpenChange = props.onOpenChange,
        }
    end,

    render = function(self)
        local display = self.placeholder
        for _, opt in ipairs(self.options) do
            local val = type(opt) == "table" and opt.value or opt
            local label = type(opt) == "table" and (opt.label or opt.value) or opt
            if val == self.value then display = label; break end
        end

        local fg = self.disabled and c.muted or (self.value ~= "" and c.fg or c.muted)

        local style = {
            border = "rounded",
            foreground = fg,
            background = c.bg,
            borderColor = c.border,
            paddingLeft = 1,
            paddingRight = 1,
            height = 3,
            align = "left",
            justify = "left",
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
            style = style,
            children = {
                { type = "text", content = display },
                { type = "text", content = " ▾", style = { foreground = c.muted } },
            },
        }
    end,
})

return Select
