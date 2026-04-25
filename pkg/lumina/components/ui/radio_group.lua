-- shadcn/radio_group — Radio button group
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
}

local RadioGroup = lumina.defineComponent({
    name = "ShadcnRadioGroup",

    init = function(props)
        return {
            options = props.options or {},
            value = props.value or "",
            disabled = props.disabled or false,
            orientation = props.orientation or "vertical",
            id = props.id,
            className = props.className,
            style = props.style,
            onChange = props.onChange,
        }
    end,

    render = function(self)
        local children = {}
        for _, opt in ipairs(self.options) do
            local val = type(opt) == "table" and opt.value or opt
            local label = type(opt) == "table" and (opt.label or opt.value) or opt
            local selected = (val == self.value)
            local icon = selected and "(●)" or "( )"
            local fg = self.disabled and c.muted or (selected and c.primary or c.muted)

            local item = {
                type = "hbox",
                id = self.id and (self.id .. "-" .. val) or nil,
                style = { align = "center", height = 1, dim = self.disabled },
                children = {
                    { type = "text", content = icon, style = { foreground = fg, bold = selected } },
                    { type = "text", content = " " .. label, style = { foreground = self.disabled and c.muted or c.fg } },
                },
            }
            if not self.disabled and self.onChange then
                item.onClick = function() self.onChange(val) end
            end
            children[#children + 1] = item
        end

        return {
            type = self.orientation == "horizontal" and "hbox" or "vbox",
            id = self.id,
            style = self.style or {},
            children = children,
        }
    end,
})

return RadioGroup
