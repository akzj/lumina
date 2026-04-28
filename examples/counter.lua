-- Lumina v2 Example: Counter
-- Demonstrates: lumina.app, useStore, global keybindings, click events
--
-- Usage: lumina examples/counter.lua
-- Quit:  q or Ctrl+Q

lumina.app {
    id = "counter",
    store = {
        count = 0,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
    },

    render = function()
        local count = lumina.useStore("count")

        return lumina.createElement("box", {
            id = "counter-box",
            style = {background = "#1E1E2E"},
            onClick = function()
                lumina.store.set("count", count + 1)
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
            }, "q or Ctrl+Q to quit")
        )
    end,
}
