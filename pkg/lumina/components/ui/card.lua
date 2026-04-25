-- shadcn/card — Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    surface = "#313244",
    bg = "#181825",
}

-- Card: Main container
local Card = lumina.defineComponent({
    name = "ShadcnCard",

    init = function(props)
        return {
            className = props.className,
            style = props.style,
            children = props.children or {},
        }
    end,

    render = function(self)
        local style = {
            border = "rounded",
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

        return {
            type = "vbox",
            style = style,
            children = self.children,
        }
    end,
})

-- CardHeader
local CardHeader = lumina.defineComponent({
    name = "ShadcnCardHeader",

    init = function(props)
        return {
            children = props.children or {},
        }
    end,

    render = function(self)
        return {
            type = "vbox",
            style = { paddingTop = 1, paddingBottom = 1 },
            children = self.children,
        }
    end,
})

-- CardTitle
local CardTitle = lumina.defineComponent({
    name = "ShadcnCardTitle",

    init = function(props)
        return {
            title = props.title or "",
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local style = { bold = true, foreground = c.fg }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end
        return {
            type = "text",
            content = self.title,
            style = style,
        }
    end,
})

-- CardDescription
local CardDescription = lumina.defineComponent({
    name = "ShadcnCardDescription",

    init = function(props)
        return {
            description = props.description or props.text or "",
        }
    end,

    render = function(self)
        return {
            type = "text",
            content = self.description,
            style = { foreground = c.muted },
        }
    end,
})

-- CardContent
local CardContent = lumina.defineComponent({
    name = "ShadcnCardContent",

    init = function(props)
        return {
            children = props.children or {},
        }
    end,

    render = function(self)
        return {
            type = "vbox",
            style = { paddingTop = 1 },
            children = self.children,
        }
    end,
})

-- CardFooter
local CardFooter = lumina.defineComponent({
    name = "ShadcnCardFooter",

    init = function(props)
        return {
            children = props.children or {},
        }
    end,

    render = function(self)
        return {
            type = "hbox",
            style = { paddingTop = 1, justify = "start", align = "center" },
            children = self.children,
        }
    end,
})

return {
    Card = Card,
    CardHeader = CardHeader,
    CardTitle = CardTitle,
    CardDescription = CardDescription,
    CardContent = CardContent,
    CardFooter = CardFooter,
}
