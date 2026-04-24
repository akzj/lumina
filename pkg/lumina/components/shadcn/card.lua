-- shadcn/card — Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter
local lumina = require("lumina")

local Card = lumina.defineComponent({
    name = "ShadcnCard",
    init = function(props)
        return { size = props.size or "default" }
    end,
    render = function(self)
        return {
            type = "vbox",
            style = {
                border = "rounded",
                foreground = "#E2E8F0",
                background = "#0F172A",
                padding = self.size == "sm" and 0 or 1,
            },
            children = self.props and self.props.children or {},
        }
    end
})

local CardHeader = lumina.defineComponent({
    name = "ShadcnCardHeader",
    init = function(props) return {} end,
    render = function(self)
        return {
            type = "vbox",
            style = { padding = 1 },
            children = self.props and self.props.children or {},
        }
    end
})

local CardTitle = lumina.defineComponent({
    name = "ShadcnCardTitle",
    init = function(props)
        return { title = props.title or "" }
    end,
    render = function(self)
        return {
            type = "text",
            content = self.title,
            style = { bold = true, foreground = "#F8FAFC" },
        }
    end
})

local CardDescription = lumina.defineComponent({
    name = "ShadcnCardDescription",
    init = function(props)
        return { text = props.text or "" }
    end,
    render = function(self)
        return {
            type = "text",
            content = self.text,
            style = { foreground = "#94A3B8" },
        }
    end
})

local CardContent = lumina.defineComponent({
    name = "ShadcnCardContent",
    init = function(props) return {} end,
    render = function(self)
        return {
            type = "vbox",
            style = { padding = 1 },
            children = self.props and self.props.children or {},
        }
    end
})

local CardFooter = lumina.defineComponent({
    name = "ShadcnCardFooter",
    init = function(props) return {} end,
    render = function(self)
        return {
            type = "hbox",
            style = { padding = 1, justify = "start", align = "center" },
            children = self.props and self.props.children or {},
        }
    end
})

return {
    Card = Card,
    CardHeader = CardHeader,
    CardTitle = CardTitle,
    CardDescription = CardDescription,
    CardContent = CardContent,
    CardFooter = CardFooter,
}
