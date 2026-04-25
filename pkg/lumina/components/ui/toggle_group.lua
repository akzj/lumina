-- shadcn/toggle_group — Group of toggles
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    surface = "#313244",
}

local ToggleGroup = lumina.defineComponent({
    name = "ShadcnToggleGroup",

    init = function(props)
        return {
            items = props.items or {},
            value = props.value or {},
            selectionMode = props.type or "single",
            variant = props.variant or "default",
            size = props.size or "default",
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}

        -- Check if a value is selected
        local function isSelected(val)
            if self.selectionMode == "single" then
                return self.value == val
            end
            if type(self.value) == "table" then
                for _, v in ipairs(self.value) do
                    if v == val then return true end
                end
            end
            return false
        end

        for i, item in ipairs(self.items) do
            local selected = isSelected(item.value or i)
            local fg = selected and c.fg or c.muted
            local bg = selected and c.surface or ""

            children[#children + 1] = {
                type = "text",
                content = " " .. (item.label or tostring(item.value or i)) .. " ",
                style = {
                    foreground = fg,
                    background = bg,
                    bold = selected,
                    border = self.variant == "outline" and "rounded" or "",
                    borderColor = selected and c.primary or "",
                },
            }
        end

        local style = { gap = 1 }
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
            children = children,
        }
    end,
})

return ToggleGroup