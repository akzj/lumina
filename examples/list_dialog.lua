-- examples/list_dialog.lua — Lux ListView + clickable rows + Lux Dialog
--
-- Demonstrates:
--   • require("lux").ListView with rows + renderRow
--   • Row onClick opens a composable Dialog (title / body / OK)
--   • j/k or Arrow keys + Enter on the list; [Esc] closes dialog; [q] quit
--
-- Usage: lumina examples/list_dialog.lua

local lux = require("lux")
local ListView = lux.ListView
local Dialog = lux.Dialog
local Title = Dialog.Title
local Content = Dialog.Content
local Actions = Dialog.Actions

local items = {
    { label = "Apple", detail = "Malus domestica — crisp, sweet, keeps doctors away (allegedly)." },
    { label = "Banana", detail = "Musa — curved yellow energy bar. Peel from the stem end." },
    { label = "Cherry", detail = "Prunus avium — small stone fruit, pie classic." },
    { label = "Date", detail = "Phoenix dactylifera — chewy Middle Eastern staple." },
    { label = "Elderberry", detail = "Sambucus — deep purple, syrup and wine." },
    { label = "Fig", detail = "Ficus carica — soft interior, Mediterranean favorite." },
}

lumina.app {
    id = "list-dialog-demo",
    store = {
        selectedIdx = 1,
        dialogOpen = false,
        dialogTitle = "",
        dialogBody = "",
    },
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
        ["Escape"] = function()
            if lumina.store.get("dialogOpen") then
                lumina.store.set("dialogOpen", false)
            end
        end,
    },
    render = function()
        local t = lumina.getTheme()
        local selectedIdx = lumina.useStore("selectedIdx")
        local dialogOpen = lumina.useStore("dialogOpen")
        local dialogTitle = lumina.useStore("dialogTitle")
        local dialogBody = lumina.useStore("dialogBody")

        local function openForRow(i, row)
            lumina.store.set("selectedIdx", i)
            lumina.store.set("dialogOpen", true)
            lumina.store.set("dialogTitle", row.label)
            lumina.store.set("dialogBody", row.detail)
        end

        local function renderRow(row, i, ctx)
            -- Opaque row bg: empty string skips hbox fill so dirty cells show terminal default.
            local rowBg = ctx.selected and (t.surface1 or "#45475A") or (t.base or "#1E1E2E")
            return lumina.createElement("hbox", {
                style = {
                    height = 1,
                    background = rowBg,
                },
                onClick = function()
                    openForRow(i, row)
                end,
            },
                lumina.createElement("text", {
                    foreground = ctx.selected and (t.primary or "#89B4FA") or (t.text or "#CDD6F4"),
                }, "  " .. row.label)
            )
        end

        local body = {
            lumina.createElement("text", {
                bold = true,
                foreground = t.text or "#CDD6F4",
                style = { height = 1 },
            }, "  List + Dialog (Lux)"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                dim = true,
                style = { height = 1 },
            }, "  Click a row · j/k or arrows · Enter opens dialog · Esc closes · q quit"),
            ListView {
                key = "main-list",
                id = "main-list",
                rows = items,
                rowHeight = 1,
                height = 14,
                width = 76,
                selectedIndex = selectedIdx,
                onChangeIndex = function(i)
                    lumina.store.set("selectedIdx", i)
                end,
                onActivate = function(i, row)
                    openForRow(i, row)
                end,
                renderRow = renderRow,
            },
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  Tip: click the list panel (not only text) to keep keyboard focus for j/k."),
        }

        if dialogOpen then
            -- Fixed overlay must set height: without it, layout defaults h=1 and hit-test
            -- misses the OK row (only the top line receives pointer events).
            body[#body + 1] = lumina.createElement("vbox", {
                style = {
                    position = "fixed",
                    left = 14,
                    top = 4,
                    width = 52,
                    height = 14,
                    zIndex = 50,
                },
            },
                Dialog {
                    open = true,
                    width = 50,
                    Title { dialogTitle },
                    Content { dialogBody },
                    Actions {
                        lumina.createElement("text", {
                            foreground = t.primary or "#89B4FA",
                            bold = true,
                            onClick = function()
                                lumina.store.set("dialogOpen", false)
                            end,
                        }, "  [ OK ]  "),
                    },
                }
            )
        end

        return lumina.createElement("vbox", {
            style = {
                width = 80,
                height = 24,
                background = t.base or "#1E1E2E",
            },
        }, table.unpack(body))
    end,
}
