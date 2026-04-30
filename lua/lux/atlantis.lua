-- lua/lux/atlantis.lua — PrimeReact Atlantis–inspired TUI chrome (shell + nav + fields).
-- Demo-only form blocks live in lux.atlantis_form_demo (examples / “Form Layout” tab).
-- Styling follows docs/css-properties.md (percent/vw/vh, borderColor, grid, flex).

local Layout = require("lux.layout")
local Breadcrumb = require("lux.breadcrumb")

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

return M
