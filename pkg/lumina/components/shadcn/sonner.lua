-- shadcn/sonner — Toast notifications
local lumina = require("lumina")

local Sonner = lumina.defineComponent({
    name = "ShadcnSonner",
    init = function(props)
        return {
            toasts = props.toasts or {},
            position = props.position or "bottom-right",
            maxVisible = props.maxVisible or 3,
        }
    end,
    render = function(self)
        local children = {}
        local visible = math.min(#self.toasts, self.maxVisible)

        for i = 1, visible do
            local toast = self.toasts[i]
            local variant = toast.variant or "default"
            local styles = {
                default     = { bg = "#1E293B", fg = "#E2E8F0", icon = "ℹ" },
                success     = { bg = "#166534", fg = "#BBF7D0", icon = "✓" },
                error       = { bg = "#991B1B", fg = "#FCA5A5", icon = "✕" },
                warning     = { bg = "#854D0E", fg = "#FDE68A", icon = "⚠" },
                info        = { bg = "#1E3A5F", fg = "#93C5FD", icon = "ℹ" },
            }
            local s = styles[variant] or styles.default

            local toastChildren = {}
            toastChildren[#toastChildren + 1] = {
                type = "text",
                content = s.icon .. " " .. (toast.title or ""),
                style = { foreground = s.fg, bold = true },
            }
            if toast.description then
                toastChildren[#toastChildren + 1] = {
                    type = "text",
                    content = "  " .. toast.description,
                    style = { foreground = s.fg },
                }
            end

            children[#children + 1] = {
                type = "vbox",
                style = {
                    border = "rounded",
                    background = s.bg,
                    foreground = s.fg,
                    padding = 1,
                    width = 40,
                },
                children = toastChildren,
            }
        end

        if #self.toasts > self.maxVisible then
            children[#children + 1] = {
                type = "text",
                content = "  +" .. (#self.toasts - self.maxVisible) .. " more",
                style = { foreground = "#64748B" },
            }
        end

        return {
            type = "vbox",
            style = { gap = 1 },
            children = children,
        }
    end
})

return Sonner
