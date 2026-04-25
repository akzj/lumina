-- shadcn/label — Form label
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    destructive = "#F38BA8",
}

local Label = lumina.defineComponent({
    name = "ShadcnLabel",

    init = function(props)
        return {
            children = props.children or props.label or "",
            htmlFor = props.htmlFor,
            required = props.required or false,
            disabled = props.disabled or false,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local content = self.children
        if type(content) ~= "string" then
            content = tostring(content)
        end
        if self.required then
            content = content .. " *"
        end

        local fg = self.disabled and c.muted or c.fg
        if self.htmlFor then
            fg = c.destructive
        end

        local style = {
            foreground = fg,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "text",
            id = self.id,
            content = content,
            style = style,
        }
    end,
})

return Label