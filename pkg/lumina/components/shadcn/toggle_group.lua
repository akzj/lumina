-- shadcn/toggle-group — Group of toggles
local lumina = require("lumina")

local ToggleGroup = lumina.defineComponent({
    name = "ShadcnToggleGroup",
    init = function(props)
        return {
            items = props.items or {},
            value = props.value or {},
            selectionMode = props.type or "single",
            variant = props.variant or "default",
            size = props.size or "default",
        }
    end,
    render = function(self)
        local children = {}

        -- Check if a value is selected
        local function isSelected(val)
            if self.selectionMode == "single" then
                return self.value == val
            end
            -- multiple
            if type(self.value) == "table" then
                for _, v in ipairs(self.value) do
                    if v == val then return true end
                end
            end
            return false
        end

        for i, item in ipairs(self.items) do
            local selected = isSelected(item.value or i)
            local fg = selected and "#E2E8F0" or "#64748B"
            local bg = selected and "#334155" or ""

            children[#children + 1] = {
                type = "text",
                content = " " .. (item.label or tostring(item.value or i)) .. " ",
                style = {
                    foreground = fg,
                    background = bg,
                    bold = selected,
                    border = self.variant == "outline" and "rounded" or "",
                },
            }
        end

        return {
            type = "hbox",
            style = { gap = 1 },
            children = children,
        }
    end
})

return ToggleGroup
