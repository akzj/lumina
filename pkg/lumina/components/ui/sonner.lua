-- shadcn/sonner — Toast notifications (terminal-friendly)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    success = "#A6E3A1",
    destructive = "#F38BA8",
    warning = "#F9E2AF",
    border = "#45475A",
    bg = "#181825",
}

local Sonner = lumina.defineComponent({
    name = "ShadcnSonner",

    init = function(props)
        return {
            toasts = props.toasts or {},
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}

        for i, toast in ipairs(self.toasts) do
            local variant = toast.variant or "default"
            local title = toast.title or ""
            local description = toast.description or ""

            local colors = {
                default = { icon = "ℹ", iconFg = c.primary, fg = c.fg },
                success = { icon = "✓", iconFg = c.success, fg = c.success },
                destructive = { icon = "✕", iconFg = c.destructive, fg = c.destructive },
                warning = { icon = "!", iconFg = c.warning, fg = c.warning },
            }
            local s = colors[variant] or colors.default

            local content = " " .. s.icon .. " " .. title
            if description ~= "" then
                content = content .. "\n  " .. description
            end

            children[#children + 1] = {
                type = "vbox",
                style = {
                    border = "rounded",
                    borderColor = c.border,
                    background = c.bg,
                    foreground = s.fg,
                    padding = 1,
                },
                children = {
                    { type = "text", content = content },
                },
            }
        end

        return {
            type = "vbox",
            id = self.id,
            style = self.style or {},
            children = children,
        }
    end,
})

return Sonner