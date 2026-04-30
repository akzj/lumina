-- lua/lux/atlantis_form_demo.lua — Form layout showcase blocks (demo / examples only).
-- Keeps lux.atlantis limited to reusable shell chrome (Shell, SideNav, TopBar, …).

local Card = require("lux.card")
local TextInput = require("lux.text_input")
local Button = require("lux.button")
local Atlantis = require("lux.atlantis")

local M = {}

--- Children array suitable for Atlantis.Shell { mainChildren = … } "Form Layout" tab.
function M.formShowcaseBlocks()
    local t = lumina.getTheme and lumina.getTheme() or Atlantis.themeTable()
    local HF = Atlantis.HorizontalField

    local function rule()
        return lumina.createElement("text", {
            foreground = t.surface2 or "#2A3A56",
            dim = true,
            style = {
                height = 1,
                width = "100%",
                whiteSpace = "nowrap",
                textOverflow = "clip",
            },
        }, string.rep("─", 400))
    end

    local grid2 = {
        display = "grid",
        gridTemplateColumns = "1fr 1fr",
        gap = 1,
        width = "100%",
    }

    return {
        lumina.createElement(Card, {
            key = "c-vert",
            title = "Vertical",
            variant = "elevated",
            padding = 1,
        },
            TextInput { id = "v-name", label = "Name", value = "", placeholder = "Your name", fill = true },
            TextInput { id = "v-email", label = "Email", value = "", placeholder = "you@example.com", fill = true },
            TextInput { id = "v-age", label = "Age", value = "", placeholder = "30", fill = true }
        ),

        lumina.createElement(Card, {
            key = "c-horiz",
            title = "Horizontal",
            variant = "elevated",
            padding = 1,
        },
            HF { id = "h-name", label = "Name", inputId = "h-name-i", placeholder = "Name" },
            HF { id = "h-email", label = "Email", inputId = "h-email-i", placeholder = "Email" }
        ),

        lumina.createElement(Card, {
            key = "c-inline",
            title = "Inline",
            variant = "elevated",
            padding = 1,
        },
            lumina.createElement("hbox", {
                style = {
                    width = "100%",
                    height = 1,
                    gap = 1,
                    align = "center",
                    justify = "start",
                    flexWrap = "nowrap",
                },
            },
                lumina.createElement("input", {
                    id = "in-first",
                    placeholder = "Firstname",
                    foreground = t.text or "#E8EDF7",
                    background = t.surface0 or "#141C2C",
                    style = { height = 1, flex = 1, minWidth = "10%" },
                }),
                lumina.createElement("input", {
                    id = "in-last",
                    placeholder = "Lastname",
                    foreground = t.text or "#E8EDF7",
                    background = t.surface0 or "#141C2C",
                    style = { height = 1, flex = 1, minWidth = "10%" },
                }),
                lumina.createElement(Button, {
                    id = "btn-submit",
                    label = "Submit",
                    variant = "primary",
                    style = { width = 10, flexShrink = 0, alignSelf = "center" },
                })
            )
        ),

        lumina.createElement(Card, {
            key = "c-grid",
            title = "Vertical Grid",
            variant = "elevated",
            padding = 1,
        },
            lumina.createElement("vbox", { style = grid2 },
                TextInput { id = "g-name", label = "Name", placeholder = "Name", fill = true },
                TextInput { id = "g-email", label = "Email", placeholder = "Email", fill = true }
            )
        ),

        lumina.createElement(Card, {
            key = "c-help",
            title = "Help Text",
            variant = "elevated",
            padding = 1,
        },
            TextInput {
                id = "help-user",
                label = "Username",
                placeholder = "",
                helperText = "Enter your username to reset your password.",
                fill = true,
            }
        ),

        lumina.createElement(Card, {
            key = "c-adv",
            title = "Advanced",
            variant = "elevated",
            padding = 1,
        },
            lumina.createElement("vbox", {
                style = {
                    display = "grid",
                    gridTemplateColumns = "1fr 1fr",
                    gap = 1,
                    width = "100%",
                },
            },
                TextInput { id = "a-fn", label = "Firstname", placeholder = "First", fill = true },
                TextInput { id = "a-ln", label = "Lastname", placeholder = "Last", fill = true },
                lumina.createElement("vbox", {
                    style = { gridColumn = "1 / 3", width = "100%" },
                },
                    TextInput {
                        id = "a-addr",
                        label = "Address",
                        placeholder = "Street, city",
                        fill = true,
                    }
                )
            )
        ),

        rule(),

        lumina.createElement("text", {
            foreground = t.muted or "#8B9BB4",
            style = { height = 1, width = "100%" },
        }, "Resize terminal — layout uses % / vw / vh and grid per docs/css-properties.md."),
    }
end

return M
