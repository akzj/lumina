-- Atlantis-style form layout demo (TUI). Adaptive via flex + scrollable Main.
--
-- Run: go run ./cmd/lumina examples/components/main.lua
-- Quit: q  or  Ctrl+C
--
-- Note: There is no lumina.getTerminalSize() for Lua; width/height follow the
-- engine root layout after terminal resize. Shell uses flex (no Main scroll by
-- default) due to scroll+height quirks in pkg/render/layout.go for plain vboxes.

local lux = require("lux")
local Atlantis = lux.Atlantis

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

        local mainBody = {}
        if active == "forms" then
            mainBody = Atlantis.formShowcaseBlocks()
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
            title = "Form Layout",
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
