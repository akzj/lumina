-- lua/theme/init.lua
-- Theme system for Lua Shadcn components
-- Reads from Go widget theme via lumina.getTheme()

local M = {}

-- Fallback theme (Catppuccin Mocha) if Go bridge not available
local _fallback = {
    base     = "#1E1E2E",
    surface0 = "#313244",
    surface1 = "#45475A",
    surface2 = "#585B70",
    text     = "#CDD6F4",
    muted    = "#6C7086",
    primary  = "#89B4FA",
    hover    = "#B4BEFE",
    pressed  = "#74C7EC",
    success  = "#A6E3A1",
    warning  = "#F9E2AF",
    error    = "#F38BA8",
}

function M.current()
    -- Try to get from Go engine first
    if lumina and lumina.getTheme then
        return lumina.getTheme()
    end
    return _fallback
end

return M
