-- lua/lux/atlantis.lua — PrimeReact Atlantis–inspired TUI chrome.
-- Styling follows docs/css-properties.md (percent/vw/vh, borderColor, grid, flex).

local Layout = require("lux.layout")
local Card = require("lux.card")
local Breadcrumb = require("lux.breadcrumb")
local TextInput = require("lux.text_input")
local Button = require("lux.button")

local M = {}

function M.themeTable()
    return {
        base = "#0B1220",
        surface0 = "#141C2C",
        surface1 = "#1B2639",
        surface2 = "#2A3A56",
        text = "#E8EDF7",
        muted = "#8B9BB4",
        primary = "#F5C842",
        primaryDark = "#0B1220",
        hover = "#FFD35A",
        pressed = "#D4A82E",
        success = "#4ADE80",
        warning = "#F5C842",
        error = "#F87171",
    }
end

function M.applyTheme()
    if lumina and lumina.setTheme then
        lumina.setTheme(M.themeTable())
    end
end

--- App header: full width, flex row (menu + trail + actions).
function M.TopBar(props)
    local t = lumina.getTheme and lumina.getTheme() or M.themeTable()
    local bg = props.background or t.surface0 or "#141C2C"
    local pageTitle = props.title or "Form Layout"
    local items = props.breadcrumbItems or {
        { id = "ui", label = "UI Kit" },
        -- Id must track the title so breadcrumb segment keys change when only the
        -- label updates (avoids stale last crumb under reconciliation).
        { id = "page-" .. pageTitle, label = pageTitle },
    }
    local right = props.rightSlot

    return lumina.createElement("hbox", {
        id = props.id,
        key = props.key,
        style = {
            width = "100%",
            height = 1,
            paddingLeft = 1,
            paddingRight = 1,
            background = bg,
            gap = 1,
            align = "center",
        },
    },
        lumina.createElement("text", {
            foreground = t.muted or "#8B9BB4",
            background = bg,
            style = { height = 1 },
        }, props.menuGlyph or "☰"),
        lumina.createElement(Breadcrumb, {
            id = props.breadcrumbId,
            items = items,
            barBackground = bg,
            style = { flex = 1, height = 1, minWidth = "15%" },
            onNavigate = props.onNavigate,
        }),
        right or lumina.createElement("text", {
            foreground = t.primary or "#F5C842",
            background = bg,
            bold = true,
            style = { height = 1, alignSelf = "center" },
        }, props.todayLabel or "Today")
    )
end

--- Sidebar list: optional bordered pills + left rail for active (borderColor + box model).
function M.SideNav(props)
    local t = lumina.getTheme and lumina.getTheme() or M.themeTable()
    local bg = props.background or t.base or "#0B1220"
    local active = props.activeId
    local onSelect = props.onSelect
    local items = props.items or {}
    local navBorder = props.itemBorder or "rounded"
    local railW = props.railWidth or 1
    -- Vertical gap between nav rows (cell count). Default 0 so items stack tight;
    -- set itemGap = 1 if you want a clear strip between pills.
    local itemGap = props.itemGap or 0
    local brandGap = props.brandGap or 1

    local rows = {}
    if props.brand then
        rows[#rows + 1] = lumina.createElement("text", {
            bold = true,
            foreground = t.primary or "#F5C842",
            background = bg,
            style = { height = 1, paddingLeft = 1, marginBottom = brandGap, width = "100%" },
        }, props.brand)
    end

    for _, it in ipairs(items) do
        local sel = (it.id == active)
        local label = it.label or it.id
        local railBg = sel and (t.primary or "#F5C842") or bg
        local railFg = sel and (t.primary or "#F5C842") or bg

        rows[#rows + 1] = lumina.createElement("hbox", {
            key = "nav-" .. tostring(it.id),
            style = {
                width = "100%",
                height = 3,
                gap = 0,
                marginBottom = itemGap,
            },
        },
            lumina.createElement("text", {
                style = {
                    width = railW,
                    height = 3,
                    background = railBg,
                    foreground = railFg,
                },
            }, " "),
            lumina.createElement("box", {
                style = {
                    flex = 1,
                    height = 3,
                    border = navBorder,
                    borderColor = sel and (t.primary or "#F5C842") or (t.surface2 or "#2A3A56"),
                    background = sel and (t.surface1 or "#1B2639") or (t.surface0 or "#141C2C"),
                    paddingLeft = 1,
                    justify = "center",
                },
                onClick = onSelect and function()
                    onSelect(it.id)
                end or nil,
            },
                lumina.createElement("text", {
                    foreground = sel and (t.primary or "#F5C842") or (t.text or "#E8EDF7"),
                    bold = sel,
                    style = { height = 1 },
                }, label)
            )
        )
    end

    rows[#rows + 1] = lumina.createElement("vbox", {
        style = { flex = 1, width = "100%", background = bg },
    })

    if props.footerHint then
        rows[#rows + 1] = lumina.createElement("text", {
            foreground = t.muted or "#8B9BB4",
            background = bg,
            style = { height = 1, paddingLeft = 1, marginTop = 1, width = "100%" },
        }, props.footerHint)
    end

    return lumina.createElement("vbox", {
        id = props.id,
        key = props.key,
        style = { background = bg, gap = 0, width = "100%", height = "100%" },
    }, table.unpack(rows))
end

--- Horizontal label + input row (flexBasis label column, flex-grow field).
function M.HorizontalField(props)
    local t = lumina.getTheme and lumina.getTheme() or M.themeTable()
    local lw = props.labelWidth or 14
    return lumina.createElement("hbox", {
        key = props.key,
        id = props.id,
        style = {
            width = "100%",
            height = 1,
            gap = 1,
            align = "center",
        },
    },
        lumina.createElement("text", {
            foreground = t.text or "#E8EDF7",
            style = {
                height = 1,
                width = lw,
            },
        }, props.label or ""),
        lumina.createElement("vbox", {
            style = { flex = 1, height = 1, minWidth = "20%" },
        },
            lumina.createElement("input", {
                id = props.inputId,
                value = props.value or "",
                placeholder = props.placeholder or "",
                foreground = t.text or "#E8EDF7",
                background = t.surface0 or "#141C2C",
                focusable = props.focusable ~= false,
                style = { height = 1, width = "100%" },
                onChange = props.onChange,
            })
        )
    )
end

function M.Shell(props)
    local t = lumina.getTheme and lumina.getTheme() or M.themeTable()
    local sidebarW = props.sidebarWidth or 26
    local mainScroll = props.mainScroll ~= false
    local innerOverflow = props.innerOverflow
    local top = props.topBar or M.TopBar({ title = props.title or "Form Layout" })

    local mainInnerStyle = {
        width = "100%",
        minHeight = "10%",
        padding = 1,
        gap = 1,
        background = t.base or "#0B1220",
    }
    if innerOverflow and innerOverflow ~= "" then
        mainInnerStyle.overflow = innerOverflow
    end

    return lumina.createElement(Layout, {
        id = props.layoutId,
        width = props.width or "100%",
        height = props.height or "100%",
    },
        Layout.Header {
            top,
            height = 1,
            background = t.surface0 or "#141C2C",
        },

        Layout.Sidebar {
            props.sidebar,
            width = sidebarW,
            background = t.base or "#0B1220",
            border = "single",
            borderColor = t.surface2 or "#2A3A56",
        },

        Layout.Main {
            lumina.createElement("vbox", {
                style = mainInnerStyle,
            }, table.unpack(props.mainChildren or {})),
            scroll = mainScroll,
        }
    )
end

function M.formShowcaseBlocks()
    local t = lumina.getTheme and lumina.getTheme() or M.themeTable()
    local HF = M.HorizontalField

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
