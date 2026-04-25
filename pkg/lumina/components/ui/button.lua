-- shadcn/button — Terminal-adapted Button component
-- Supports: variant, size, disabled, id, className, onClick, style
local lumina = require("lumina")

-- Color palette (Catppuccin Mocha)
local c = {
    primary      = "#89B4FA",
    primary_fg   = "#1E1E2E",
    destructive  = "#F38BA8",
    destructive_fg = "#1E1E2E",
    outline      = "#313244",
    outline_fg   = "#CDD6F4",
    secondary    = "#313244",
    secondary_fg = "#CDD6F4",
    ghost_fg     = "#CDD6F4",
    link         = "#89B4FA",
    muted_bg     = "#313244",
    muted_fg     = "#6C7086",
}

local Button = lumina.defineComponent({
    name = "ShadcnButton",

    init = function(props)
        return {
            variant = props.variant or "default",
            size = props.size or "default",
            disabled = props.disabled or false,
            label = props.label or props.children or "Button",
            id = props.id,
            className = props.className,
            onClick = props.onClick,
            style = props.style,
        }
    end,

    render = function(self)
        local variant = self.variant
        local disabled = self.disabled

        -- Variant styles: bg, fg
        local styles = {
            default     = { bg = c.primary,      fg = c.primary_fg },
            destructive = { bg = c.destructive,   fg = c.destructive_fg },
            outline     = { bg = c.outline,       fg = c.outline_fg },
            secondary   = { bg = c.secondary,      fg = c.secondary_fg },
            ghost       = { bg = "",              fg = c.ghost_fg },
            link        = { bg = "",              fg = c.link },
        }
        local s = styles[variant] or styles.default

        -- Size styles: padding, height, width
        local sizeMap = {
            default = { paddingH = 1, paddingV = 0, height = 3, width = nil },
            sm      = { paddingH = 0, paddingV = 0, height = 1, width = nil },
            lg      = { paddingH = 2, paddingV = 0, height = 5, width = nil },
            icon    = { paddingH = 0, paddingV = 0, height = 3, width = 5,  minWidth = 5 },
            xs      = { paddingH = 0, paddingV = 0, height = 1, width = nil },
        }
        local sz = sizeMap[self.size] or sizeMap.default

        local fg = disabled and c.muted_fg or s.fg
        local bg = s.bg
        if variant == "outline" or variant == "ghost" then
            bg = disabled and c.muted_bg or (bg ~= "" and bg or "")
        elseif variant == "link" then
            bg = ""
        end

        -- Build the button VNode
        local btn = {
            type = "hbox",
            id = self.id,
            style = {
                background = bg,
                foreground = fg,
                border = variant == "outline" and "rounded" or (variant == "default" and "rounded" or ""),
                paddingLeft = sz.paddingH,
                paddingRight = sz.paddingH,
                paddingTop = sz.paddingV,
                paddingBottom = sz.paddingV,
                height = sz.height,
                justify = "center",
                align = "center",
                dim = disabled,
            },
            onClick = disabled and nil or self.onClick,
            children = {
                {
                    type = "text",
                    content = self.label,
                },
            },
        }

        -- Apply className overrides (simple: className is a table of style overrides)
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do
                btn.style[k] = v
            end
        end

        -- Apply inline style overrides
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do
                btn.style[k] = v
            end
        end

        -- Width constraint for icon buttons
        if sz.width and sz.minWidth then
            btn.style.minWidth = sz.minWidth
        end

        return btn
    end,
})

return Button
