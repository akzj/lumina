-- shadcn/switch — Toggle switch
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    success = "#A6E3A1",
    primary = "#89B4FA",
}

local Switch = lumina.defineComponent({
    name = "ShadcnSwitch",

    init = function(props)
        return {
            checked = props.checked or false,
            disabled = props.disabled or false,
            size = props.size or "default",
            label = props.label or "",
            id = props.id,
            className = props.className,
            style = props.style,
            onCheckedChange = props.onCheckedChange,
        }
    end,

    render = function(self)
        local checked = self.checked
        local disabled = self.disabled
        local track = self.size == "sm" and (checked and "━●" or "●━") or (checked and "━━●" or "●━━")
        local fg = disabled and c.muted or (checked and c.success or c.muted)

        local children = {
            { type = "text", content = "[" .. track .. "]", style = { foreground = fg, bold = true } },
        }
        if self.label ~= "" then
            children[#children + 1] = {
                type = "text",
                content = " " .. self.label,
                style = { foreground = disabled and c.muted or c.fg },
            }
        end

        local sw = {
            type = "hbox",
            id = self.id,
            style = { align = "center" },
            children = children,
        }
        if not disabled and self.onCheckedChange then
            sw.onClick = function() self.onCheckedChange(not checked) end
        end
        return sw
    end,
})

return Switch
