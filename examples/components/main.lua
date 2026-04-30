-- Atlantis-style demo (TUI): Form Layout + Button showcase pages.
--
-- Run: go run ./cmd/lumina examples/components/main.lua
-- Quit: q  or  Ctrl+C

local lux = require("lux")
local Atlantis = lux.Atlantis
local Card = lux.Card
local Button = lux.Button

--- Lux Button showcase (severities, appearances, group, split) — uses docs/css-properties.md styles.
local function buttonShowcaseBlocks()
    local noop = function() end

    local function wrapRow(...)
        return lumina.createElement("hbox", {
            style = { gap = 1, flexWrap = "wrap", width = "100%" },
        }, ...)
    end

    return {
        lumina.createElement(Card, {
            key = "c-sev",
            title = "Severities (solid)",
            variant = "elevated",
            padding = 1,
        },
            wrapRow(
                Button { key = "bs-p", label = "Primary", severity = "primary", onClick = noop },
                Button { key = "bs-s", label = "Secondary", severity = "secondary", onClick = noop },
                Button { key = "bs-ok", label = "Success", severity = "success", onClick = noop },
                Button { key = "bs-i", label = "Info", severity = "info", onClick = noop },
                Button { key = "bs-h", label = "Help", severity = "help", onClick = noop },
                Button { key = "bs-w", label = "Warning", severity = "warning", onClick = noop },
                Button { key = "bs-d", label = "Danger", severity = "danger", onClick = noop }
            )
        ),

        lumina.createElement(Card, {
            key = "c-out",
            title = "Outlined",
            variant = "elevated",
            padding = 1,
        },
            wrapRow(
                Button { key = "bo-p", label = "Primary", severity = "primary", appearance = "outlined", onClick = noop },
                Button { key = "bo-s", label = "Secondary", severity = "secondary", appearance = "outlined", onClick = noop },
                Button { key = "bo-ok", label = "Success", severity = "success", appearance = "outlined", onClick = noop },
                Button { key = "bo-d", label = "Danger", severity = "danger", appearance = "outlined", onClick = noop }
            )
        ),

        lumina.createElement(Card, {
            key = "c-txt",
            title = "Text / Link",
            variant = "elevated",
            padding = 1,
        },
            wrapRow(
                Button { key = "bt-p", label = "Primary", severity = "primary", appearance = "text", onClick = noop },
                Button { key = "bt-d", label = "Danger", severity = "danger", appearance = "text", onClick = noop },
                Button { key = "bt-l", label = "Link style", severity = "info", link = true, onClick = noop }
            )
        ),

        lumina.createElement(Card, {
            key = "c-raised",
            title = "Raised",
            variant = "elevated",
            padding = 1,
        },
            wrapRow(
                Button { key = "br-p", label = "Primary", severity = "primary", appearance = "raised", onClick = noop },
                Button { key = "br-s", label = "Secondary", severity = "secondary", appearance = "raised", onClick = noop }
            )
        ),

        lumina.createElement(Card, {
            key = "c-icon",
            title = "Icon",
            variant = "elevated",
            padding = 1,
        },
            wrapRow(
                Button { key = "bi-l", label = "Run", icon = "▶", iconPosition = "left", severity = "success", onClick = noop },
                Button { key = "bi-r", label = "Next", icon = "→", iconPosition = "right", severity = "primary", onClick = noop },
                Button { key = "bi-o", icon = "⚙", iconOnly = true, severity = "secondary", onClick = noop }
            )
        ),

        lumina.createElement(Card, {
            key = "c-grp",
            title = "Button group",
            variant = "elevated",
            padding = 1,
        },
            Button.Group {
                key = "grp1",
                severity = "primary",
                items = {
                    { label = "Save", onClick = noop },
                    { label = "Delete", onClick = noop },
                    { label = "Cancel", onClick = noop },
                },
            }
        ),

        lumina.createElement(Card, {
            key = "c-spl",
            title = "Split button",
            variant = "elevated",
            padding = 1,
        },
            Button.Split {
                key = "spl1",
                label = "Action",
                severity = "secondary",
                onClick = noop,
                onMenuClick = noop,
            }
        ),

        lumina.createElement(Card, {
            key = "c-dis",
            title = "Disabled",
            variant = "elevated",
            padding = 1,
        },
            wrapRow(
                Button { key = "bd-1", label = "Disabled solid", severity = "primary", disabled = true, onClick = noop },
                Button { key = "bd-2", label = "Disabled outline", severity = "danger", appearance = "outlined", disabled = true, onClick = noop }
            )
        ),

        lumina.createElement("text", {
            foreground = (lumina.getTheme() or {}).muted or "#8B9BB4",
            style = { height = 1, width = "100%", marginTop = 1 },
        }, "require(\"lux\").Button — docs/lux-api.md, lua/lux/button.lua"),
    }
end

local pageTitles = {
    forms = "Form Layout",
    btn = "Button",
    dash = "Dashboard",
    input = "Input",
    table = "Table",
}

lumina.app {
    id = "atlantis-forms",
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
    },
    render = function()
        lumina.useEffect(function()
            Atlantis.applyTheme()
        end, {})

        local active, setActive = lumina.useState("nav", "forms")

        local navItems = {
            { id = "dash", label = "Dashboard" },
            { id = "forms", label = "Form Layout" },
            { id = "input", label = "Input" },
            { id = "btn", label = "Button" },
            { id = "table", label = "Table" },
        }

        local shellTitle = pageTitles[active] or "UI Kit"

        local mainBody = {}
        if active == "forms" then
            mainBody = Atlantis.formShowcaseBlocks()
        elseif active == "btn" then
            mainBody = buttonShowcaseBlocks()
        else
            local t = lumina.getTheme()
            mainBody = {
                lumina.createElement("text", {
                    bold = true,
                    foreground = t.text or "#E8EDF7",
                    style = { height = 1 },
                }, "Section: " .. active),
                lumina.createElement("text", {
                    foreground = t.muted or "#8B9BB4",
                    style = { height = 1 },
                }, "Placeholder — wire routes or pages here."),
            }
        end

        return Atlantis.Shell {
            title = shellTitle,
            sidebar = Atlantis.SideNav {
                brand = "ATLANTIS",
                items = navItems,
                activeId = active,
                onSelect = function(id) setActive(id) end,
                footerHint = "[q] quit",
            },
            mainChildren = mainBody,
        }
    end,
}
