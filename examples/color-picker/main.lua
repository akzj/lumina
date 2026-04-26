-- ============================================================================
-- Lumina Example: Color Picker
-- ============================================================================
-- Demonstrates: ColorPicker component, HSL/RGB/Hex conversion, palette preview
-- Run: lumina examples/color-picker/main.lua
-- ============================================================================
local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

-- Catppuccin palette for presets
local catppuccin = {
    { name = "Base",       hex = "#1E1E2E", fg = "#CDD6F4" },
    { name = "Mantle",     hex = "#181825", fg = "#CDD6F4" },
    { name = "Crust",      hex = "#11111B", fg = "#CDD6F4" },
    { name = "Surface0",   hex = "#313244", fg = "#CDD6F4" },
    { name = "Surface1",   hex = "#45475A", fg = "#CDD6F4" },
    { name = "Surface2",   hex = "#585B70", fg = "#CDD6F4" },
    { name = "Overlay0",   hex = "#6C7086", fg = "#CDD6F4" },
    { name = "Overlay1",   hex = "#7F849C", fg = "#CDD6F4" },
    { name = "Overlay2",   hex = "#9399B2", fg = "#1E1E2E" },
    { name = "Subtext0",   hex = "#A6ADC8", fg = "#1E1E2E" },
    { name = "Subtext1",   hex = "#BAC2DE", fg = "#1E1E2E" },
    { name = "Text",       hex = "#CDD6F4", fg = "#1E1E2E" },
    { name = "Blue",       hex = "#89B4FA", fg = "#1E1E2E" },
    { name = "Sapphire",   hex = "#74C7EC", fg = "#1E1E2E" },
    { name = "Sky",        hex = "#89DCEB", fg = "#1E1E2E" },
    { name = "Teal",       hex = "#94E2D5", fg = "#1E1E2E" },
    { name = "Green",      hex = "#A6E3A1", fg = "#1E1E2E" },
    { name = "Yellow",     hex = "#F9E2AF", fg = "#1E1E2E" },
    { name = "Peach",      hex = "#FAB387", fg = "#1E1E2E" },
    { name = "Maroon",      hex = "#EBA0AC", fg = "#1E1E2E" },
    { name = "Red",        hex = "#F38BA8", fg = "#1E1E2E" },
    { name = "Mauve",      hex = "#CBA6F7", fg = "#1E1E2E" },
    { name = "Pink",       hex = "#F5C2E7", fg = "#1E1E2E" },
    { name = "Lavender",   hex = "#B4BEFE", fg = "#1E1E2E" },
}

local store = lumina.createStore({
    state = {
        selectedColor = "#89B4FA",
        colorName = "Blue",
        mode = "hex", -- hex, rgb, hsl
    },
})

-- Parse hex to RGB
local function hexToRgb(hex)
    local r = tonumber(hex:sub(2, 3), 16)
    local g = tonumber(hex:sub(4, 5), 16)
    local b = tonumber(hex:sub(6, 7), 16)
    return r, g, b
end

-- RGB to HSL
local function rgbToHsl(r, g, b)
    r, g, b = r / 255, g / 255, b / 255
    local max, min = math.max(r, g, b), math.min(r, g, b)
    local h, s, l = 0, 0, (max + min) / 2

    if max ~= min then
        local d = max - min
        s = l > 0.5 and d / (2 - max - min) or d / (max + min)
        if max == r then
            h = (g - b) / d + (g < b and 6 or 0)
        elseif max == g then
            h = (b - r) / d + 2
        else
            h = (r - g) / d + 4
        end
        h = h / 6
    end

    return math.floor(h * 360), math.floor(s * 100), math.floor(l * 100)
end

-- HSL to RGB
local function hslToRgb(h, s, l)
    h, s, l = h / 360, s / 100, l / 100
    local r, g, b
    if s == 0 then
        r, g, b = l, l, l
    else
        local function hue2rgb(p, q, t)
            if t < 0 then t = t + 1 end
            if t > 1 then t = t - 1 end
            if t < 1/6 then return p + (q - p) * 6 * t end
            if t < 1/2 then return q end
            if t < 2/3 then return p + (q - p) * (2/3 - t) * 6 end
            return p
        end
        local q = l < 0.5 and l * (1 + s) or l + s - l * s
        local p = 2 * l - q
        r, g, b = hue2rgb(p, q, h + 1/3), hue2rgb(p, q, h), hue2rgb(p, q, h - 1/3)
    end
    return math.floor(r * 255), math.floor(g * 255), math.floor(b * 255)
end

-- Format color based on mode
local function formatColor(hex, mode)
    local r, g, b = hexToRgb(hex)
    if mode == "hex" then
        return string.upper(hex)
    elseif mode == "rgb" then
        return string.format("rgb(%d, %d, %d)", r, g, b)
    else -- hsl
        local h, s, l = rgbToHsl(r, g, b)
        return string.format("hsl(%d°, %d%%, %d%%)", h, s, l)
    end
end

local c = {
    bg = "#1E1E2E", fg = "#CDD6F4", accent = "#89B4FA",
    success = "#A6E3A1", error = "#F38BA8", muted = "#6C7086",
    surface = "#313244", border = "#45475A",
}

local function ColorBox(props)
    return {
        type = "hbox",
        style = {
            width = 1, height = 1,
            background = props.color,
            border = "single",
            borderColor = props.selected and c.accent or c.border,
        },
        children = {},
    }
end

local App = lumina.defineComponent({
    name = "ColorPicker",
    render = function(self)
        local state = lumina.useStore(store)
        local color = state.selectedColor
        local mode = state.mode or "hex"
        local colorName = state.colorName or "Custom"

        local r, g, b = hexToRgb(color)
        local h, s, l = rgbToHsl(r, g, b)
        local formatted = formatColor(color, mode)

        return {
            type = "vbox",
            style = { background = c.bg },
            children = {
                -- Header
                { type = "text", content = " 🎨 Color Picker ", style = { foreground = c.accent, bold = true, background = "#181825" } },
                { type = "text", content = " Select a color from the Catppuccin palette ", style = { foreground = c.muted } },
                { type = "text", content = "" },

                -- Selected color preview
                {
                    type = "vbox",
                    style = { border = "rounded", borderColor = c.border, padding = 1, width = 40 },
                    children = {
                        { type = "text", content = "┌─ Selected Color ─────────────────────────┐", style = { foreground = c.border } },
                        { type = "text", content = "│  " .. string.rep(" ", 40) .. " │", style = { foreground = c.border } },
                        {
                            type = "text",
                            content = "│  " .. colorName .. string.rep(" ", math.max(1, 34 - #colorName)) .. " │",
                            style = { foreground = color, bold = true },
                        },
                        { type = "text", content = "│  " .. formatted .. string.rep(" ", math.max(1, 39 - #formatted)) .. " │", style = { foreground = c.fg } },
                        { type = "text", content = "│  " .. color .. string.rep(" ", math.max(1, 39 - #color)) .. " │", style = { foreground = c.muted } },
                        { type = "text", content = "└" .. string.rep("─", 40) .. "┘", style = { foreground = c.border } },
                    },
                },

                { type = "text", content = "" },

                -- Mode toggle
                {
                    type = "hbox",
                    children = {
                        { type = "text", content = " Mode: ", style = { foreground = c.muted } },
                        {
                            type = "text",
                            content = mode == "hex" and " [HEX] " or "  HEX  ",
                            style = { foreground = mode == "hex" and c.accent or c.muted, bold = mode == "hex" },
                        },
                        {
                            type = "text",
                            content = mode == "rgb" and " [RGB] " or "  RGB  ",
                            style = { foreground = mode == "rgb" and c.accent or c.muted, bold = mode == "rgb" },
                        },
                        {
                            type = "text",
                            content = mode == "hsl" and " [HSL] " or "  HSL  ",
                            style = { foreground = mode == "hsl" and c.accent or c.muted, bold = mode == "hsl" },
                        },
                    },
                },

                { type = "text", content = "" },

                -- Palette grid
                {
                    type = "vbox",
                    style = { border = "rounded", borderColor = c.border, padding = 1 },
                    children = (function()
                        local rows = {}
                        local cols = 6
                        for i = 1, #catppuccin, cols do
                            local row = {}
                            for j = i, math.min(i + cols - 1, #catppuccin) do
                                local swatch = catppuccin[j]
                                local selected = (swatch.hex == color)
                                local display = "[" .. swatch.name .. "]"
                                if #display > 13 then display = "[" .. swatch.name:sub(1, 10) .. "]" end
                                row[#row + 1] = {
                                    type = "text",
                                    content = display,
                                    style = {
                                        foreground = swatch.hex,
                                        bold = selected,
                                        background = selected and c.surface or "",
                                        dim = false,
                                    },
                                }
                            end
                            rows[#rows + 1] = { type = "hbox", children = row }
                        end
                        return rows
                    end)(),
                },

                { type = "text", content = "" },
                { type = "text", content = " [1-4] switch mode  [q] quit", style = { foreground = c.muted, dim = true } },
            },
        }
    end,
})

-- Key bindings
lumina.onKey("1", function() store.dispatch("setState", { mode = "hex" }) end)
lumina.onKey("2", function() store.dispatch("setState", { mode = "rgb" }) end)
lumina.onKey("3", function() store.dispatch("setState", { mode = "hsl" }) end)
lumina.onKey("4", function()
    -- Cycle through palette
    local state = store.getState()
    local current = state.selectedColor
    for i, sw in ipairs(catppuccin) do
        if sw.hex == current then
            local next = catppuccin[(i % #catppuccin) + 1]
            store.dispatch("setState", { selectedColor = next.hex, colorName = next.name })
            break
        end
    end
end)

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
