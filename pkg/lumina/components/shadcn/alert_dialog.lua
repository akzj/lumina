-- shadcn/alert_dialog — Confirmation dialog with action/cancel
local lumina = require("lumina")

local AlertDialog = lumina.defineComponent({
    name = "ShadcnAlertDialog",
    init = function(props)
        return {
            open = props.open or false,
            title = props.title or "Are you sure?",
            description = props.description or "",
            confirmLabel = props.confirmLabel or "Continue",
            cancelLabel = props.cancelLabel or "Cancel",
            variant = props.variant or "default",
        }
    end,
    render = function(self)
        if not self.open then
            return { type = "fragment", children = {} }
        end
        local confirmFg = self.variant == "destructive" and "#DC2626" or "#3B82F6"
        return {
            type = "vbox",
            style = {
                border = "rounded",
                background = "#1E1E2E",
                foreground = "#CDD6F4",
                padding = 1,
                width = 50,
                height = 12,
            },
            children = {
                { type = "text", content = self.title, style = { bold = true, foreground = "#F8FAFC" } },
                { type = "text", content = self.description, style = { foreground = "#94A3B8" } },
                { type = "text", content = string.rep("─", 46), style = { foreground = "#334155" } },
                {
                    type = "hbox",
                    style = { justify = "end" },
                    children = {
                        { type = "text", content = "[ " .. self.cancelLabel .. " ]", style = { foreground = "#94A3B8" } },
                        { type = "text", content = "  " },
                        { type = "text", content = "[ " .. self.confirmLabel .. " ]", style = { foreground = confirmFg, bold = true } },
                    },
                },
            },
        }
    end
})

return AlertDialog
