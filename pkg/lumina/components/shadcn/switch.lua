-- shadcn/switch — Toggle switch
local lumina = require("lumina")

local Switch = lumina.defineComponent({
    name = "ShadcnSwitch",
    init = function(props)
        return {
            checked = props.checked or false,
            disabled = props.disabled or false,
            size = props.size or "default",
            label = props.label or "",
        }
    end,
    render = function(self)
        local track, thumb
        if self.size == "sm" then
            track = self.checked and "━●" or "●━"
        else
            track = self.checked and "━━●" or "●━━"
        end
        local fg = self.disabled and "#475569" or (self.checked and "#22C55E" or "#64748B")
        local children = {
            { type = "text", content = track, style = { foreground = fg, bold = true } },
        }
        if self.label ~= "" then
            children[#children + 1] = {
                type = "text",
                content = " " .. self.label,
                style = { foreground = self.disabled and "#475569" or "#E2E8F0" },
            }
        end
        return {
            type = "hbox",
            children = children,
        }
    end
})

return Switch
