-- Lumina v2 Example: Counter
-- Demonstrates: createComponent, useState, createElement, click events
--
-- Usage: lumina examples/counter.lua
-- Quit:  Ctrl+C or Ctrl+Q

-- "lumina" is a global table registered by the Go runtime.
-- No require() needed.

lumina.createComponent({
    id = "counter",
    name = "Counter",
    x = 0, y = 0,
    w = 40, h = 10,
    zIndex = 0,

    render = function(state, props)
        local count, setCount = lumina.useState("count", 0)

        return lumina.createElement("box", {
            id = "counter-box",
            style = {background = "#1E1E2E"},
            onClick = function()
                setCount(count + 1)
            end,
        },
            lumina.createElement("text", {
                foreground = "#89B4FA",
                bold = true,
            }, "Lumina v2 Counter"),

            lumina.createElement("text", {
                foreground = "#CDD6F4",
            }, ""),

            lumina.createElement("text", {
                foreground = "#A6E3A1",
            }, "Count: " .. tostring(count)),

            lumina.createElement("text", {
                foreground = "#6C7086",
            }, ""),

            lumina.createElement("text", {
                foreground = "#6C7086",
            }, "Click anywhere to increment"),

            lumina.createElement("text", {
                foreground = "#6C7086",
            }, "Ctrl+C or Ctrl+Q to quit")
        )
    end,
})
