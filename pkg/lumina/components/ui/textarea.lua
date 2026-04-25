-- shadcn/textarea — Multi-line text input display
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    border = "#45475A",
    bg = "#1E1E2E",
}

local Textarea = lumina.defineComponent({
    name = "ShadcnTextarea",

    init = function(props)
        return {
            value = props.value or "",
            placeholder = props.placeholder or "",
            disabled = props.disabled or false,
            rows = props.rows or 4,
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
            height = self.rows + 2,
            align = "left",
            justify = "left",
        }

        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "vbox",
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

return Textarea
