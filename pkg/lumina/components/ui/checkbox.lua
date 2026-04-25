-- shadcn/checkbox — Checkbox with checked/unchecked/indeterminate state
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
}

local Checkbox = lumina.defineComponent({
    name = "ShadcnCheckbox",

    init = function(props)
        return {
            checked = props.checked or false,
            indeterminate = props.indeterminate or false,
            disabled = props.disabled or false,
            label = props.label or "",
            id = props.id,
            className = props.className,
            style = props.style,
            onChange = props.onChange,
        }
    end,

    render = function(self)
        local checked = self.checked
        local indeterminate = self.indeterminate
        local disabled = self.disabled

        local icon
        if indeterminate then icon = "[-]"
        elseif checked then icon = "[✓]"
        else icon = "[ ]"
        end

        local iconFg = disabled and c.muted or (checked or indeterminate) and c.primary or c.muted
        local labelFg = disabled and c.muted or c.fg

        local children = {
            { type = "text", content = icon, style = { foreground = iconFg, bold = checked or indeterminate } },
        }
        if self.label ~= "" then
            children[#children + 1] = {
                type = "text",
                content = " " .. self.label,
                style = { foreground = labelFg },
            }
        end

        local style = { align = "center", height = 1 }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local checkbox = {
            type = "hbox",
            id = self.id,
            style = style,
            children = children,
        }

        if not disabled and self.onChange then
            checkbox.onClick = function()
                self.onChange(not checked)
            end
        end

        return checkbox
    end,
})

return Checkbox
