-- Theme Demo: Shows Lumina's CSS DSL and Theme system
local lumina = require("lumina")

-- Define custom styles using CSS DSL pattern
lumina.defineGlobalStyles({
    primary = { color = "cyan", bold = true },
    secondary = { color = "gray" },
    surface = { background = "black" },
    error = { color = "red" },
    success = { color = "green" }
})

-- Custom theme
lumina.defineTheme("night", {
    colors = {
        primary = "magenta",
        secondary = "lightblue",
        background = "#1a1a2e",
        text = "white",
        surface = "#16213e"
    },
    spacing = {
        xs = 1,
        sm = 2,
        md = 4,
        lg = 8
    }
})

-- Theme demo component
local ThemeDemo = lumina.defineComponent({
    name = "ThemeDemo",
    
    init = function(props)
        return { theme = props.theme or "dark" }
    end,
    
    render = function(instance)
        -- Get theme info
        local theme = lumina.hooks.useTheme()
        local colors = theme.colors
        local spacing = theme.spacing
        
        return {
            type = "vbox",
            children = {
                {
                    type = "text",
                    content = "================================",
                    color = colors.secondary
                },
                {
                    type = "text",
                    content = "  Lumina Theme Demo",
                    color = colors.primary,
                    bold = true
                },
                {
                    type = "text",
                    content = "================================",
                    color = colors.secondary
                },
                { type = "text", content = "" },
                {
                    type = "text",
                    content = "Current Theme: " .. theme.name,
                    color = colors.text
                },
                { type = "text", content = "" },
                {
                    type = "text",
                    content = "Available Colors:",
                    color = colors.text
                },
                {
                    type = "text",
                    content = "  primary:   " .. colors.primary,
                    color = colors.primary
                },
                {
                    type = "text",
                    content = "  secondary: " .. colors.secondary,
                    color = colors.secondary
                },
                {
                    type = "text",
                    content = "  text:      " .. colors.text,
                    color = colors.text
                },
                {
                    type = "text",
                    content = "  error:     " .. colors.error,
                    color = colors.error
                },
                {
                    type = "text",
                    content = "  success:   " .. colors.success,
                    color = colors.success
                },
                { type = "text", content = "" },
                {
                    type = "text",
                    content = "Spacing:",
                    color = colors.text
                },
                {
                    type = "text",
                    content = "  xs: " .. spacing.xs .. "  sm: " .. spacing.sm .. "  md: " .. spacing.md .. "  lg: " .. spacing.lg,
                    color = colors.secondary
                },
                { type = "text", content = "" },
                {
                    type = "text",
                    content = "--------------------------------",
                    color = colors.secondary
                }
            }
        }
    end
})

-- Run with default theme
print("=== Default (dark) theme ===")
lumina.render(ThemeDemo)

-- Switch to light theme
print("\n=== Switched to light theme ===")
lumina.setTheme("light")
lumina.render(ThemeDemo)

-- Switch to custom theme
print("\n=== Custom 'night' theme ===")
lumina.setTheme("night")
lumina.render(ThemeDemo)

print("\nTheme system working!")
