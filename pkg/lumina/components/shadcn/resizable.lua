-- resizable.lua — Resizable split panels
local lumina = require("lumina")

local ResizablePanel = lumina.defineComponent({
    name = "ShadcnResizablePanel",
    render = function(self)
        local size = self.props.size or 50 -- percentage
        local minSize = self.props.minSize or 10
        local maxSize = self.props.maxSize or 90

        return {
            type = "vbox",
            style = {
                flex = size,
                minWidth = self.props.minWidth,
                minHeight = self.props.minHeight,
            },
            children = self.props.children or {},
        }
    end,
})

local ResizableHandle = lumina.defineComponent({
    name = "ShadcnResizableHandle",
    render = function(self)
        local direction = self.props.direction or "horizontal"
        local withHandle = self.props.withHandle ~= false

        if direction == "horizontal" then
            return {
                type = "vbox",
                style = {
                    width = 1,
                    background = "#45475A",
                    foreground = "#6C7086",
                },
                children = withHandle and {
                    { type = "text", content = "┃" },
                } or {},
            }
        else
            return {
                type = "hbox",
                style = {
                    height = 1,
                    background = "#45475A",
                    foreground = "#6C7086",
                },
                children = withHandle and {
                    { type = "text", content = "━" },
                } or {},
            }
        end
    end,
})

local ResizablePanelGroup = lumina.defineComponent({
    name = "ShadcnResizablePanelGroup",
    render = function(self)
        local direction = self.props.direction or "horizontal"
        return {
            type = direction == "horizontal" and "hbox" or "vbox",
            style = {
                width = self.props.width or "100%",
                height = self.props.height or "100%",
                border = self.props.border or "",
            },
            children = self.props.children or {},
        }
    end,
})

return {
    Panel = ResizablePanel,
    Handle = ResizableHandle,
    PanelGroup = ResizablePanelGroup,
}
