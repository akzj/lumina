-- shadcn/dialog — Modal dialog with overlay
-- Note: Requires overlay manager. For terminal, use with sheet component instead.
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
    destructive = "#F38BA8",
}

local Dialog = lumina.defineComponent({
    name = "ShadcnDialog",

    init = function(props)
        return {
            open = props.open or false,
            title = props.title or "Dialog",
            description = props.description,
            content = props.content or {},
            onOpenChange = props.onOpenChange,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        if not self.open then
            return { type = "empty" }
        end

        local style = {
            background = c.bg,
            border = "rounded",
            borderColor = c.border,
            paddingLeft = 1,
            paddingRight = 1,
            paddingTop = 0,
            paddingBottom = 0,
            justify = "center",
            align = "center",
        }

        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        -- Dialog content
        local children = {
            -- Header
            {
                type = "hbox",
                style = { paddingTop = 1 },
                children = {
                    { type = "text", content = self.title, style = { foreground = c.fg, bold = true } },
                    { type = "spacer" },
                },
            },
        }

        if self.description then
            children[#children + 1] = {
                type = "text",
                content = self.description,
                style = { foreground = c.muted, paddingTop = 1 },
            }
        end

        children[#children + 1] = {
            type = "text",
            content = "────────────────────────────────",
            style = { foreground = c.border, paddingTop = 1 },
        }

        -- Content
        if self.content then
            if type(self.content) == "table" and self.content.type then
                children[#children + 1] = self.content
            else
                children[#children + 1] = {
                    type = "text",
                    content = tostring(self.content),
                    style = { foreground = c.fg },
                }
            end
        end

        children[#children + 1] = {
            type = "text",
            content = "────────────────────────────────",
            style = { foreground = c.border, paddingTop = 1 },
        }

        -- Footer hint
        children[#children + 1] = {
            type = "text",
            content = "[Esc] to close",
            style = { foreground = c.muted, dim = true },
        }

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return Dialog
