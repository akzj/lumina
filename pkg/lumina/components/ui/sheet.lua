-- shadcn/sheet — Side panel (slide in from edge)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
    destructive = "#F38BA8",
}

local Sheet = lumina.defineComponent({
    name = "ShadcnSheet",

    init = function(props)
        return {
            open = props.open or false,
            side = props.side or "right", -- right, left, top, bottom
            title = props.title or "",
            width = props.width or 40,
            height = props.height,
            id = props.id,
            className = props.className,
            style = props.style,
            children = props.children or {},
        }
    end,

    render = function(self)
        if not self.open then
            return { type = "empty" }
        end

        local style = {
            border = "single",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
            padding = 1,
            width = self.width,
            height = self.height,
            justify = "left",
        }

        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local children = {}
        if self.title ~= "" then
            children[#children + 1] = {
                type = "text",
                content = "┌─ " .. self.title .. " " .. string.rep("─", math.max(0, self.width - #self.title - 4)) .. "┐",
                style = { foreground = c.primary, bold = true },
            }
            children[#children + 1] = {
                type = "text",
                content = "│" .. string.rep(" ", self.width - 2) .. "│",
                style = { foreground = c.border },
            }
        end

        -- Add content children
        local contentChildren = self.children
        if type(contentChildren) == "table" then
            if contentChildren.type then
                children[#children + 1] = contentChildren
            else
                for _, child in ipairs(contentChildren) do
                    children[#children + 1] = child
                end
            end
        end

        -- Close border if we have a title
        if self.title ~= "" then
            children[#children + 1] = {
                type = "text",
                content = "└" .. string.rep("─", self.width - 2) .. "┘",
                style = { foreground = c.border },
            }
        end

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return Sheet
