local lumina = require("lumina")
local theme = require("theme")

local M = {}

-- Local store for this page's interactive state
local store = lumina.createStore({
    state = {
        clickCount = 0,
        lastClicked = "none",
        loading = false,
    }
})

-- Button helper function
local function Button(props)
    local id = props.id or "btn"
    local label = props.label or "Button"
    local variant = props.variant or "default"  -- default, primary, outline, ghost, destructive
    local size = props.size or "md"             -- sm, md, lg
    local disabled = props.disabled or false

    -- Size presets
    local padH = 2
    if size == "sm" then padH = 1
    elseif size == "lg" then padH = 3 end

    -- Variant styles
    local fg, bg, border = theme.text, theme.surface, "rounded"
    if variant == "primary" then
        fg, bg = theme.primary_fg, theme.primary_bg
    elseif variant == "outline" then
        fg, bg = theme.text, nil
    elseif variant == "ghost" then
        fg, bg, border = theme.text, nil, nil
    elseif variant == "destructive" then
        fg, bg = "#FFFFFF", theme.error
    end

    -- Disabled state
    if disabled then
        fg = theme.muted
        bg = theme.surface
    end

    return {
        type = "box",
        id = id,
        style = {
            foreground = fg,
            background = bg,
            border = border,
            paddingLeft = padH,
            paddingRight = padH,
        },
        children = {
            { type = "text", content = label, style = { foreground = fg, bold = (variant == "primary") } }
        }
    }
end

function M.render()
    local state = lumina.useStore(store)

    return {
        type = "vbox",
        style = { flex = 1, padding = 2, gap = 1 },
        children = {
            -- Title
            { type = "text", content = "Button", style = { foreground = theme.text, bold = true } },
            { type = "text", content = "Displays a button or a component that looks like a button.", style = { foreground = theme.subtext } },
            { type = "text", content = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━", style = { foreground = theme.border } },

            -- Click counter
            { type = "text", content = string.format("  Clicks: %d | Last: %s", state.clickCount, state.lastClicked),
              style = { foreground = theme.accent } },
            { type = "text", content = "", style = {} },

            -- Section: Variants
            { type = "text", content = "  Variants", style = { foreground = theme.text, bold = true } },
            { type = "text", content = "  Different visual styles for different contexts.", style = { foreground = theme.muted, dim = true } },
            { type = "text", content = "", style = {} },
            {
                type = "hbox",
                style = { gap = 2, paddingLeft = 2 },
                children = {
                    Button({ id = "btn-default", label = "Default", variant = "default" }),
                    Button({ id = "btn-primary", label = "Primary", variant = "primary" }),
                    Button({ id = "btn-outline", label = "Outline", variant = "outline" }),
                    Button({ id = "btn-ghost", label = "Ghost", variant = "ghost" }),
                    Button({ id = "btn-destructive", label = "Destructive", variant = "destructive" }),
                }
            },
            { type = "text", content = "", style = {} },

            -- Section: Sizes
            { type = "text", content = "  Sizes", style = { foreground = theme.text, bold = true } },
            { type = "text", content = "  Three sizes: sm, md (default), lg.", style = { foreground = theme.muted, dim = true } },
            { type = "text", content = "", style = {} },
            {
                type = "hbox",
                style = { gap = 2, paddingLeft = 2 },
                children = {
                    Button({ id = "btn-sm", label = "Small", variant = "primary", size = "sm" }),
                    Button({ id = "btn-md", label = "Medium", variant = "primary", size = "md" }),
                    Button({ id = "btn-lg", label = "Large", variant = "primary", size = "lg" }),
                }
            },
            { type = "text", content = "", style = {} },

            -- Section: States
            { type = "text", content = "  States", style = { foreground = theme.text, bold = true } },
            { type = "text", content = "  Hover (mouse over) and disabled states.", style = { foreground = theme.muted, dim = true } },
            { type = "text", content = "", style = {} },
            {
                type = "hbox",
                style = { gap = 2, paddingLeft = 2 },
                children = {
                    Button({ id = "btn-hover", label = "Hover Me", variant = "primary" }),
                    Button({ id = "btn-disabled", label = "Disabled", variant = "default", disabled = true }),
                }
            },
            { type = "text", content = "", style = {} },

            -- Section: Features
            { type = "text", content = "  Supported Features", style = { foreground = theme.text, bold = true } },
            { type = "text", content = "  ✓ Click events    — lumina.on('click', id, fn)", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Hover state     — lumina.isHovered(id)", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Focus/Tab       — lumina.isFocused(id), Tab navigation", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Keyboard        — lumina.onKey(key, fn)", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Drag & Drop     — lumina.useDrag/useDrop", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Variants        — default, primary, outline, ghost, destructive", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Sizes           — sm, md, lg", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Disabled state  — style.dim + no click handler", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Border styles   — single, double, rounded", style = { foreground = theme.success } },
            { type = "text", content = "  ✓ Custom colors   — foreground, background via style", style = { foreground = theme.success } },
        }
    }
end

return M
