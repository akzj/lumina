-- shadcn/alert_dialog — Confirmation dialog with action/cancel
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    destructive = "#F38BA8",
    border = "#45475A",
    bg = "#181825",
}

local AlertDialog = lumina.defineComponent({
    name = "ShadcnAlertDialog",

    init = function(props)
        return {
            open = props.open or false,
            title = props.title or "Are you sure?",
            description = props.description or "",
            confirmLabel = props.confirmLabel or "Continue",
            cancelLabel = props.cancelLabel or "Cancel",
            variant = props.variant or "default", -- default, destructive
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        if not self.open then
            return { type = "empty" }
        end

        local confirmFg = (self.variant == "destructive") and c.destructive or c.primary
        local w = 50

        local style = {
            border = "rounded",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
            padding = 1,
            width = w,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local children = {
            { type = "text", content = self.title, style = { bold = true, foreground = c.fg } },
        }
        if self.description ~= "" then
            children[#children + 1] = {
                type = "text",
                content = self.description,
                style = { foreground = c.muted },
            }
        end
        children[#children + 1] = {
            type = "text",
            content = string.rep("─", w - 2),
            style = { foreground = c.border },
        }
        children[#children + 1] = {
            type = "hbox",
            style = { justify = "end" },
            children = {
                { type = "text", content = "[ " .. self.cancelLabel .. " ]", style = { foreground = c.muted } },
                { type = "text", content = "  " },
                { type = "text", content = "[ " .. self.confirmLabel .. " ]", style = { foreground = confirmFg, bold = true } },
            },
        }

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return AlertDialog