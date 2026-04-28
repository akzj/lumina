-- lua/lux/slot.lua
-- Slot marker factory creator for composable components.
-- Usage:
--   local Slot = require("lux.slot")
--   local Title = Slot("title")
--   Title { "Hello" }  → { type = "_slot", _slotName = "title", children = {"Hello"} }

local function Slot(name)
    return setmetatable({}, {
        __call = function(self, arg)
            local children = {}
            local props = {}

            if type(arg) == "table" then
                for k, v in pairs(arg) do
                    if type(k) == "number" then
                        children[k] = v
                    else
                        props[k] = v
                    end
                end
            elseif type(arg) == "string" then
                children[1] = arg
            end

            return {
                type = "_slot",
                _slotName = name,
                children = children,
                props = props,
            }
        end,
    })
end

return Slot
