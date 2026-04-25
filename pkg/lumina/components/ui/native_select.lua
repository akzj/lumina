-- shadcn/native_select — Simple inline select (no overlay)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    border = "#45475A",
    bg = "#181825",
}

local NativeSelect = lumina.defineComponent({
    name = "ShadcnNativeSelect",

    init = function(props)
        return {
            options = props.options or {},
            value = props.value or "",
            disabled = props.disabled or false,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local display = self.value
        if display == "" and #self.options > 0 then
            local first = self.options[1]
            display = type(first) == "table" and (first.label or first.value) or first
        end
        local fg = self.disabled and c.muted or c.fg

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
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "hbox",
            id = self.id,
            style = style,
            children = {
                { type = "text", content = "◄ " .. tostring(display) .. " ►", style = { foreground = fg } },
            },
        }
    end,
})

return NativeSelect