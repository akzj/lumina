-- lua/theme/init.lua
-- Theme system for Lua UI components
-- Reads from Go widget theme via lumina.getTheme()

local M = {}

-- Fallback theme (Catppuccin Mocha) if Go bridge not available
local _fallback = {
    base        = "#1E1E2E",
    surface0    = "#313244",
    surface1    = "#45475A",
    surface2    = "#585B70",
    text        = "#CDD6F4",
    muted       = "#6C7086",
    primary     = "#89B4FA",
    primaryDark = "#1E1E2E",
    hover       = "#B4BEFE",
    pressed     = "#74C7EC",
    success     = "#A6E3A1",
    warning     = "#F9E2AF",
    error       = "#F38BA8",
}

-- Built-in theme names (for reference in Lua)
M.themes = { "mocha", "latte", "nord", "dracula" }

function M.current()
    -- Try to get from Go engine first
    if lumina and lumina.getTheme then
        return lumina.getTheme()
    end
    return _fallback
end

function M.setTheme(nameOrTable)
    if lumina and lumina.setTheme then
        lumina.setTheme(nameOrTable)
    end
end

return M
