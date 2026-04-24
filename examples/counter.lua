-- Counter example for testing MCP server
local Counter = lumina.defineComponent({
    name = "Counter",
    init = function(props)
        return { count = props.initial or 0 }
    end,
    render = function(instance)
        return {
            type = "vbox",
            children = {
                { type = "text", content = "Count: " .. tostring(instance.count) }
            }
        }
    end
})

lumina.render(Counter, { initial = 42 })
