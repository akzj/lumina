-- examples/hover_demo.lua — Hover Style Demo
--
-- Demonstrates hoverStyle: text changes appearance on mouse hover.
-- No animation — purely tests hover visual feedback.
--
-- Press q or Ctrl+C to quit.

lumina.app {
    id = "hover-demo",
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
    },
    render = function()
        return lumina.createElement("vbox", {
            style = { width = "100%", height = "100%", justify = "center", align = "center" },
        },
            lumina.createElement("text", {
                foreground = "#6C7086",
                style = { height = 1, textAlign = "center" },
            }, "Hover Demo — move mouse over items below"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = "#666666",
                hoverStyle = { foreground = "#ffffff", bold = true, underline = true },
                onClick = function() end,
                style = { height = 1, textAlign = "center" },
            }, "[ White + Bold + Underline on hover ]"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = "#666666",
                hoverStyle = { foreground = "#00ff88", bold = true },
                onClick = function() end,
                style = { height = 1, textAlign = "center" },
            }, "[ Green + Bold on hover ]"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = "#666666",
                hoverStyle = { foreground = "#ff5555", inverse = true },
                onClick = function() end,
                style = { height = 1, textAlign = "center" },
            }, "[ Red + Inverse on hover ]"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = "#666666",
                hoverStyle = { background = "#333333", foreground = "#ffffff" },
                onClick = function() end,
                style = { height = 1, textAlign = "center" },
            }, "[ White on dark background on hover ]"),
            lumina.createElement("text", { style = { height = 2 } }, ""),
            lumina.createElement("text", {
                foreground = "#6C7086",
                dim = true,
                style = { height = 1, textAlign = "center" },
            }, "Press q to quit")
        )
    end,
}
