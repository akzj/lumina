-- shadcn/collapsible — Collapsible content section
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    surface = "#313244",
}

local Collapsible = lumina.defineComponent({
    name = "ShadcnCollapsible",

    init = function(props)
        return {
            open = props.open or false,
            disabled = props.disabled or false,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local chevron = self.open and "▼" or "▶"

        local header = {
            type = "hbox",
            id = self.id and (self.id .. "-trigger") or nil,
            style = { padding = 0 },
            children = {
                { type = "text", content = " " .. chevron .. " ", style = { foreground = c.muted } },
            },
        }

        local content = {}
        if self.open then
            -- Content children would be passed via slot/children
        end

        return {
            type = "vbox",
            id = self.id,
            style = self.style or {},
            children = {
                header,
            },
        }
    end,
})

return Collapsible
