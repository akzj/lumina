-- shadcn/input — Styled text input field
-- Note: For interactive terminal input, use lumina.registerInput() + lumina.focusInput()
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    muted_bg = "#313244",
    border = "#45475A",
    destructive = "#F38BA8",
    bg = "#1E1E2E",
}

local Input = lumina.defineComponent({
    name = "ShadcnInput",

    init = function(props)
        return {
            placeholder = props.placeholder or "",
            value = props.value or "",
            disabled = props.disabled or false,
            type = props.type or "text",
            id = props.id,
            className = props.className,
            style = props.style,
            onChange = props.onChange,
            onFocus = props.onFocus,
            onBlur = props.onBlur,
        }
    end,

    render = function(self)
        local display = self.value
        if display == "" then
            display = self.placeholder
        end

        local fg
        if self.disabled then
            fg = c.muted
        elseif display == self.placeholder then
            fg = c.muted
        else
            fg = c.fg
        end

        local style = {
            border = "rounded",
            foreground = fg,
            background = c.bg,
            borderColor = c.border,
            paddingLeft = 1,
            paddingRight = 1,
            paddingTop = 0,
            paddingBottom = 0,
            height = 3,
            align = "left",
            justify = "left",
        }

        -- Apply className overrides
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        -- Apply inline style overrides
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "hbox",
            id = self.id,
            style = style,
            onFocus = self.onFocus,
            onBlur = self.onBlur,
            children = {
                { type = "text", content = display },
            },
        }
    end,
})

return Input
