-- button_group.lua — Group of buttons with shared styling
local lumina = require("lumina")

local ButtonGroup = lumina.defineComponent({
    name = "ShadcnButtonGroup",
    render = function(self)
        local direction = self.props.direction or "row"
        local gap = self.props.gap or 0
        local variant = self.props.variant or "outline"
        local children = self.props.children or {}

        -- Wrap children to pass variant through
        local wrapped = {}
        for i, child in ipairs(children) do
            if type(child) == "table" then
                local c = {}
                for k, v in pairs(child) do c[k] = v end
                if not c.variant then c.variant = variant end
                -- First/last get different border treatment
                if i == 1 then
                    c._groupPosition = "first"
                elseif i == #children then
                    c._groupPosition = "last"
                else
                    c._groupPosition = "middle"
                end
                wrapped[#wrapped + 1] = c
            else
                wrapped[#wrapped + 1] = child
            end
        end

        return {
            type = direction == "row" and "hbox" or "vbox",
            style = {
                gap = gap,
                border = self.props.border or "rounded",
            },
            children = wrapped,
        }
    end,
})

return ButtonGroup
