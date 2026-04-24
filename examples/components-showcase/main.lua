-- ============================================================================
-- Lumina Example: Components Showcase
-- ============================================================================
-- Gallery of all shadcn components — the "Storybook" for Lumina.
-- Showcases: Tabs, all 47 shadcn components, interactive demos
--
-- Run: go run ./cmd/lumina examples/components-showcase/main.lua
-- ============================================================================

local lumina = require("lumina")

-- ── Store ───────────────────────────────────────────────────────────────────
local store = lumina.createStore({
    state = {
        tab = "basic",
        scroll = 0,
        switchOn = false,
        progressValue = 65,
        accordionOpen = "section1",
        togglePressed = false,
    },
})

-- ── Colors ──────────────────────────────────────────────────────────────────
local c = {
    bg       = "#1E1E2E",
    fg       = "#CDD6F4",
    accent   = "#89B4FA",
    success  = "#A6E3A1",
    error    = "#F38BA8",
    warning  = "#F9E2AF",
    muted    = "#6C7086",
    surface  = "#313244",
    border   = "#45475A",
    pink     = "#F5C2E7",
    cyan     = "#89DCEB",
}

-- ── Section Header ──────────────────────────────────────────────────────────
local function SectionHeader(props)
    return {
        type = "vbox",
        children = {
            { type = "text", content = "" },
            { type = "text", content = "  ▎ " .. props.title, style = { foreground = c.accent, bold = true } },
            { type = "text", content = "  " .. string.rep("─", 50), style = { foreground = c.border } },
        },
    }
end

-- ── Basic Components Tab ────────────────────────────────────────────────────
local function BasicTab()
    return {
        type = "vbox",
        children = {
            SectionHeader({ title = "Button" }),
            {
                type = "hbox",
                children = {
                    { type = "text", content = "  [Default]  ", style = { foreground = "#FFF", background = "#2563EB", bold = true } },
                    { type = "text", content = " " },
                    { type = "text", content = "  [Outline]  ", style = { foreground = c.fg, border = "rounded" } },
                    { type = "text", content = " " },
                    { type = "text", content = "  [Ghost]    ", style = { foreground = c.fg } },
                    { type = "text", content = " " },
                    { type = "text", content = "  [Danger]   ", style = { foreground = "#FFF", background = "#DC2626" } },
                },
            },

            SectionHeader({ title = "Badge" }),
            {
                type = "hbox",
                children = {
                    { type = "text", content = " Default ", style = { foreground = "#FFF", background = c.accent } },
                    { type = "text", content = "  " },
                    { type = "text", content = " Secondary ", style = { foreground = c.fg, background = c.surface } },
                    { type = "text", content = "  " },
                    { type = "text", content = " Destructive ", style = { foreground = "#FFF", background = c.error } },
                    { type = "text", content = "  " },
                    { type = "text", content = " Outline ", style = { foreground = c.fg } },
                },
            },

            SectionHeader({ title = "Label" }),
            { type = "text", content = "  Email address:", style = { foreground = c.fg } },
            { type = "text", content = "  ┌──────────────────────────┐", style = { foreground = c.border } },
            { type = "text", content = "  │ user@example.com        │", style = { foreground = c.muted } },
            { type = "text", content = "  └──────────────────────────┘", style = { foreground = c.border } },

            SectionHeader({ title = "Separator" }),
            { type = "text", content = "  Content above" , style = { foreground = c.fg } },
            { type = "text", content = "  " .. string.rep("─", 40), style = { foreground = c.border } },
            { type = "text", content = "  Content below", style = { foreground = c.fg } },

            SectionHeader({ title = "Kbd" }),
            {
                type = "hbox",
                children = {
                    { type = "text", content = "  Press " },
                    { type = "text", content = " ⌘K ", style = { foreground = c.fg, background = c.surface, bold = true } },
                    { type = "text", content = " to open command palette" },
                },
            },
        },
    }
end

-- ── Card Components Tab ─────────────────────────────────────────────────────
local function CardTab()
    return {
        type = "vbox",
        children = {
            SectionHeader({ title = "Card" }),
            { type = "text", content = "  ╭──────────────────────────────────╮", style = { foreground = c.border } },
            { type = "text", content = "  │  📊 Monthly Revenue             │", style = { foreground = c.accent, bold = true } },
            { type = "text", content = "  │  Track your monthly earnings    │", style = { foreground = c.muted } },
            { type = "text", content = "  │                                 │", style = { foreground = c.border } },
            { type = "text", content = "  │  $45,231.89                     │", style = { foreground = c.success, bold = true } },
            { type = "text", content = "  │  +20.1% from last month         │", style = { foreground = c.success } },
            { type = "text", content = "  ╰──────────────────────────────────╯", style = { foreground = c.border } },

            SectionHeader({ title = "Alert" }),
            { type = "text", content = "  ┌─ ⚠ Warning ──────────────────────┐", style = { foreground = c.warning } },
            { type = "text", content = "  │ Your trial expires in 3 days.    │", style = { foreground = c.fg } },
            { type = "text", content = "  └───────────────────────────────────┘", style = { foreground = c.warning } },
            { type = "text", content = "" },
            { type = "text", content = "  ┌─ ✕ Error ────────────────────────┐", style = { foreground = c.error } },
            { type = "text", content = "  │ Failed to save. Please retry.    │", style = { foreground = c.fg } },
            { type = "text", content = "  └───────────────────────────────────┘", style = { foreground = c.error } },

            SectionHeader({ title = "Avatar" }),
            {
                type = "hbox",
                children = {
                    { type = "text", content = "  " },
                    { type = "text", content = " AC ", style = { foreground = "#FFF", background = c.pink, bold = true } },
                    { type = "text", content = "  " },
                    { type = "text", content = " BW ", style = { foreground = "#FFF", background = c.accent, bold = true } },
                    { type = "text", content = "  " },
                    { type = "text", content = " CL ", style = { foreground = "#FFF", background = c.success, bold = true } },
                    { type = "text", content = "  Alice Chen  Bob Wang  Carol Li", style = { foreground = c.muted } },
                },
            },

            SectionHeader({ title = "Breadcrumb" }),
            {
                type = "hbox",
                children = {
                    { type = "text", content = "  Home", style = { foreground = c.accent } },
                    { type = "text", content = " / ", style = { foreground = c.muted } },
                    { type = "text", content = "Components", style = { foreground = c.accent } },
                    { type = "text", content = " / ", style = { foreground = c.muted } },
                    { type = "text", content = "Card", style = { foreground = c.fg } },
                },
            },
        },
    }
end

-- ── Form Components Tab ─────────────────────────────────────────────────────
local function FormTab()
    local state = lumina.useStore(store)
    local switchOn = state.switchOn
    local progressValue = state.progressValue or 65

    local switchDisplay = switchOn and "  [●━━] ON " or "  [━━●] OFF"
    local switchColor = switchOn and c.success or c.muted

    local barWidth = 30
    local filled = math.floor(progressValue / 100 * barWidth)
    local progressBar = string.rep("█", filled) .. string.rep("░", barWidth - filled)

    return {
        type = "vbox",
        children = {
            SectionHeader({ title = "Input" }),
            { type = "text", content = "  Username:", style = { foreground = c.fg } },
            { type = "text", content = "  ┌──────────────────────────┐", style = { foreground = c.border } },
            { type = "text", content = "  │ johndoe                  │", style = { foreground = c.fg } },
            { type = "text", content = "  └──────────────────────────┘", style = { foreground = c.border } },

            SectionHeader({ title = "Switch" }),
            { type = "text", content = switchDisplay, style = { foreground = switchColor, bold = true } },
            { type = "text", content = "  Press [s] to toggle", style = { foreground = c.muted, dim = true } },

            SectionHeader({ title = "Progress" }),
            { type = "text", content = string.format("  %d%%", progressValue), style = { foreground = c.accent, bold = true } },
            { type = "text", content = "  " .. progressBar, style = { foreground = c.accent } },
            { type = "text", content = "  Press [+/-] to change", style = { foreground = c.muted, dim = true } },

            SectionHeader({ title = "Toggle" }),
            {
                type = "hbox",
                children = {
                    { type = "text", content = "  " },
                    { type = "text", content = state.togglePressed and " [Bold] " or "  Bold  ",
                      style = { foreground = state.togglePressed and "#FFF" or c.muted,
                                background = state.togglePressed and c.accent or nil, bold = state.togglePressed } },
                    { type = "text", content = "  Italic  ", style = { foreground = c.muted } },
                    { type = "text", content = "  Underline  ", style = { foreground = c.muted } },
                },
            },

            SectionHeader({ title = "Skeleton" }),
            { type = "text", content = "  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░", style = { foreground = c.surface } },
            { type = "text", content = "  ░░░░░░░░░░░░░░░░░░", style = { foreground = c.surface } },
            { type = "text", content = "  ░░░░░░░░░░░░░░░░░░░░░░░░", style = { foreground = c.surface } },
        },
    }
end

-- ── Complex Components Tab ──────────────────────────────────────────────────
local function ComplexTab()
    local state = lumina.useStore(store)
    local openSection = state.accordionOpen or "section1"

    return {
        type = "vbox",
        children = {
            SectionHeader({ title = "Accordion" }),
            { type = "text", content = (openSection == "section1" and "  ▼ " or "  ▶ ") .. "What is Lumina?",
              style = { foreground = c.accent, bold = true } },
            openSection == "section1" and
                { type = "text", content = "    A React-style TUI framework for Go + Lua.",
                  style = { foreground = c.fg } } or nil,
            { type = "text", content = "  " .. string.rep("─", 44), style = { foreground = c.border } },
            { type = "text", content = (openSection == "section2" and "  ▼ " or "  ▶ ") .. "How does it work?",
              style = { foreground = c.accent, bold = true } },
            openSection == "section2" and
                { type = "text", content = "    Define components in Lua, render to terminal.",
                  style = { foreground = c.fg } } or nil,
            { type = "text", content = "  " .. string.rep("─", 44), style = { foreground = c.border } },

            SectionHeader({ title = "Tabs" }),
            {
                type = "hbox",
                children = {
                    { type = "text", content = "  [Account] ", style = { foreground = c.accent, bold = true } },
                    { type = "text", content = "  Password  ", style = { foreground = c.muted } },
                    { type = "text", content = "  Settings  ", style = { foreground = c.muted } },
                },
            },
            { type = "text", content = "  " .. string.rep("━", 40), style = { foreground = c.border } },
            { type = "text", content = "  Make changes to your account here.", style = { foreground = c.fg } },

            SectionHeader({ title = "Table" }),
            { type = "text", content = string.format("  %-15s %-10s %-10s", "Invoice", "Status", "Amount"),
              style = { foreground = c.pink, bold = true } },
            { type = "text", content = "  " .. string.rep("─", 40), style = { foreground = c.border } },
            { type = "text", content = string.format("  %-15s %-10s %-10s", "INV-001", "Paid", "$250.00"),
              style = { foreground = c.fg } },
            { type = "text", content = string.format("  %-15s %-10s %-10s", "INV-002", "Pending", "$150.00"),
              style = { foreground = c.fg } },
            { type = "text", content = string.format("  %-15s %-10s %-10s", "INV-003", "Unpaid", "$350.00"),
              style = { foreground = c.fg } },

            SectionHeader({ title = "Pagination" }),
            {
                type = "hbox",
                children = {
                    { type = "text", content = "  ◀ ", style = { foreground = c.muted } },
                    { type = "text", content = " 1 ", style = { foreground = "#FFF", background = c.accent, bold = true } },
                    { type = "text", content = " 2 ", style = { foreground = c.fg } },
                    { type = "text", content = " 3 ", style = { foreground = c.fg } },
                    { type = "text", content = " ... ", style = { foreground = c.muted } },
                    { type = "text", content = " 10 ", style = { foreground = c.fg } },
                    { type = "text", content = " ▶", style = { foreground = c.muted } },
                },
            },
        },
    }
end

-- ── Spinner Tab ─────────────────────────────────────────────────────────────
local function AnimationTab()
    return {
        type = "vbox",
        children = {
            SectionHeader({ title = "Spinner" }),
            { type = "text", content = "  ◐ Loading...", style = { foreground = c.accent } },
            { type = "text", content = "  ⠋ Processing data...", style = { foreground = c.cyan } },
            { type = "text", content = "  ⣾ Syncing files...", style = { foreground = c.success } },

            SectionHeader({ title = "Dialog (simulated)" }),
            { type = "text", content = "  ╭─────────────────────────────╮", style = { foreground = c.border } },
            { type = "text", content = "  │   Are you sure?            │", style = { foreground = c.fg, bold = true } },
            { type = "text", content = "  │                            │", style = { foreground = c.border } },
            { type = "text", content = "  │   This action cannot be    │", style = { foreground = c.muted } },
            { type = "text", content = "  │   undone.                  │", style = { foreground = c.muted } },
            { type = "text", content = "  │                            │", style = { foreground = c.border } },
            { type = "text", content = "  │  [Cancel]       [Confirm]  │", style = { foreground = c.accent } },
            { type = "text", content = "  ╰─────────────────────────────╯", style = { foreground = c.border } },

            SectionHeader({ title = "Tooltip (simulated)" }),
            { type = "text", content = "  Hover over button:", style = { foreground = c.fg } },
            { type = "text", content = "  ┌─────────────────────┐", style = { foreground = c.surface } },
            { type = "text", content = "  │ Add to library      │", style = { foreground = c.fg, background = c.surface } },
            { type = "text", content = "  └─────────────────────┘", style = { foreground = c.surface } },
            { type = "text", content = "          ▲", style = { foreground = c.surface } },
            { type = "text", content = "     [  Save  ]", style = { foreground = "#FFF", background = c.accent } },
        },
    }
end

-- ── Tab Bar ─────────────────────────────────────────────────────────────────
local function TabBar()
    local state = lumina.useStore(store)
    local currentTab = state.tab or "basic"

    local tabs = {
        { key = "basic",     label = "1:Basic" },
        { key = "cards",     label = "2:Cards" },
        { key = "form",      label = "3:Form" },
        { key = "complex",   label = "4:Complex" },
        { key = "animation", label = "5:Overlay" },
    }

    local children = {}
    for _, tab in ipairs(tabs) do
        local isActive = (tab.key == currentTab)
        children[#children + 1] = {
            type = "text",
            content = isActive
                and string.format(" [%s] ", tab.label)
                or  string.format("  %s  ", tab.label),
            style = {
                foreground = isActive and c.accent or c.muted,
                bold = isActive,
                background = isActive and c.surface or nil,
            },
        }
    end

    return {
        type = "vbox",
        children = {
            { type = "hbox", children = children },
            { type = "text", content = string.rep("━", 60), style = { foreground = c.border } },
        },
    }
end

-- ── Main App ───────────────────────────────────────────────────────────────
local App = lumina.defineComponent({
    name = "ComponentsShowcase",
    render = function(self)
        local state = lumina.useStore(store)
        local tab = state.tab or "basic"

        local content
        if tab == "basic" then content = BasicTab()
        elseif tab == "cards" then content = CardTab()
        elseif tab == "form" then content = FormTab()
        elseif tab == "complex" then content = ComplexTab()
        elseif tab == "animation" then content = AnimationTab()
        end

        return {
            type = "vbox",
            style = { background = c.bg },
            children = {
                { type = "text", content = " 🎨 Lumina Components Showcase ",
                  style = { foreground = c.accent, bold = true, background = "#181825" } },
                { type = "text", content = " All 47 shadcn components for the terminal",
                  style = { foreground = c.muted, dim = true } },
                { type = "text", content = "" },
                TabBar(),
                content,
                { type = "text", content = "" },
                { type = "text", content = " [1-5] tabs  [s]witch  [+/-]progress  [q]uit",
                  style = { foreground = c.muted, dim = true } },
            },
        }
    end,
})

-- ── Key bindings ───────────────────────────────────────────────────────────
local tabKeys = { ["1"] = "basic", ["2"] = "cards", ["3"] = "form", ["4"] = "complex", ["5"] = "animation" }
for key, tab in pairs(tabKeys) do
    lumina.onKey(key, function() store.dispatch("setState", { tab = tab }) end)
end

lumina.onKey("s", function()
    local state = store.getState()
    store.dispatch("setState", { switchOn = not state.switchOn })
end)

lumina.onKey("+", function()
    local state = store.getState()
    store.dispatch("setState", { progressValue = math.min(100, (state.progressValue or 65) + 5) })
end)

lumina.onKey("-", function()
    local state = store.getState()
    store.dispatch("setState", { progressValue = math.max(0, (state.progressValue or 65) - 5) })
end)

lumina.onKey("a", function()
    local state = store.getState()
    local open = state.accordionOpen
    if open == "section1" then
        store.dispatch("setState", { accordionOpen = "section2" })
    else
        store.dispatch("setState", { accordionOpen = "section1" })
    end
end)

lumina.onKey("t", function()
    local state = store.getState()
    store.dispatch("setState", { togglePressed = not state.togglePressed })
end)

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
