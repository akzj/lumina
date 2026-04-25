-- shadcn/resizable — Resizable panel layout
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    border = "#45475A",
    bg = "#181825",
}

-- Panel: Single resizable panel
local Panel = lumina.defineComponent({
    name = "ShadcnResizablePanel",

    init = function(props)
        return {
            defaultSize = props.defaultSize or 50,
            minSize = props.minSize or 10,
            maxSize = props.maxSize or 100,
            id = props.id,
            className = props.className,
            style = props.style,
            children = props.children or {},
        }
    end,

    render = function(self)
        local style = {
            width = self.defaultSize,
            border = "right",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
            padding = 1,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local contentChildren = self.children
        local children = {}
        if type(contentChildren) == "table" then
            if contentChildren.type then
                children[1] = contentChildren
            else
                for _, child in ipairs(contentChildren) do
                    children[#children + 1] = child
                end
            end
        end

        return {
            type = "hbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

-- Handle: Resize handle divider
local Handle = lumina.defineComponent({
    name = "ShadcnResizableHandle",

    init = function(props)
        return {
            direction = props.direction or "horizontal",
            id = props.id,
        }
    end,

    render = function(self)
        local content = self.direction == "horizontal" and "│" or "─"
        return {
            type = "text",
            id = self.id,
            content = content,
            style = { foreground = c.border },
        }
    end,
})

-- PanelGroup: Container for resizable panels
local PanelGroup = lumina.defineComponent({
    name = "ShadcnResizablePanelGroup",

    init = function(props)
        return {
            direction = props.direction or "horizontal",
            id = props.id,
            className = props.className,
            style = props.style,
            children = props.children or {},
        }
    end,

    render = function(self)
        local style = {
            border = "single",
            borderColor = c.border,
            background = c.bg,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = self.direction == "vertical" and "vbox" or "hbox",
            id = self.id,
            style = style,
            children = self.children,
        }
    end,
})

return {
    Panel = Panel,
    Handle = Handle,
    PanelGroup = PanelGroup,
}