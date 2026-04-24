-- Counter example
local Counter = lumina.defineComponent({
    name = "Counter",
    render = function()
        return { type = "text", content = "Hello World" }
    end
})
lumina.render(Counter)
