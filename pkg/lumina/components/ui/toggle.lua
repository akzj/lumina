-- shadcn/toggle — Toggle button (pressed/unpressed)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    primary_fg = "#1E1E2E",
    surface = "#313244",
}

local Toggle = lumina.defineComponent({
    name = "ShadcnToggle",

    init = function(props)
        return {
            pressed = props.pressed or false,
            variant = props.variant or "default",
            size = props.size or "default",
            label = props.label or "",
            disabled = props.disabled or false,
            id = props.id,
            className = props.className,
            style = props.style,
            onPressedChange = props.onPressedChange,
        }
    end,

    render = function(self)
        local pressed = self.pressed
        local disabled = self.disabled

        local fg, bg
        if disabled then
            fg = c.muted; bg = ""
        elseif pressed then
            if self.variant == "outline" then fg = c.fg; bg = c.surface
            else fg = c.primary_fg; bg = c.primary end
        else
            fg = c.muted; bg = ""
        end

        local sz = self.size == "sm" and {h=0, v=0} or self.size == "lg" and {h=2, v=0} or {h=1, v=0}

        local style = {
            foreground = fg,
            background = bg,
            border = self.variant == "outline" and "rounded" or "",
            borderColor = pressed and c.primary or "",
            paddingLeft = sz.h, paddingRight = sz.h,
            paddingTop = sz.v, paddingBottom = sz.v,
            justify = "center", align = "center",
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local toggle = {
            type = "hbox",
            id = self.id,
            style = style,
            children = { { type = "text", content = self.label } },
        }
        if not disabled and self.onPressedChange then
            toggle.onClick = function() self.onPressedChange(not pressed) end
        end
        return toggle
    end,
})

return Toggle
