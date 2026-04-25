local lumina = require("lumina")
local theme = require("theme")

-- Pages (lazy-loaded via require)
local pages = {
    { name = "Button", path = "pages/button", key = "1" },
    -- Future: { name = "Input", path = "pages/input", key = "2" },
    -- Future: { name = "Card", path = "pages/card", key = "3" },
}

local store = lumina.createStore({
    state = { currentPage = 1 }
})

-- Key bindings for page navigation
for i, page in ipairs(pages) do
    lumina.onKey(tostring(i), function()
        store.dispatch("setState", { currentPage = i })
    end)
end
lumina.onKey("q", function() lumina.quit() end)
lumina.onKey("j", function() lumina.scrollBy("content-scroll", 1) end)
lumina.onKey("k", function() lumina.scrollBy("content-scroll", -1) end)

-- Sidebar component
local function Sidebar()
    local state = lumina.useStore(store)
    local children = {
        { type = "text", content = " 📦 Components", style = { foreground = theme.accent, bold = true } },
        { type = "text", content = " ─────────────────", style = { foreground = theme.border } },
        { type = "text", content = "", style = {} },
    }
    for i, page in ipairs(pages) do
        local selected = (state.currentPage == i)
        children[#children + 1] = {
            type = "text",
            content = selected
                and string.format(" ▸ %d. %s", i, page.name)
                or  string.format("   %d. %s", i, page.name),
            style = {
                foreground = selected and theme.accent or theme.subtext,
                bold = selected,
                background = selected and theme.surface or nil,
            },
        }
    end
    -- Footer
    children[#children + 1] = { type = "text", content = "", style = {} }
    children[#children + 1] = { type = "text", content = " [j/k] Scroll", style = { foreground = theme.muted, dim = true } }
    children[#children + 1] = { type = "text", content = " [q] Quit", style = { foreground = theme.muted, dim = true } }

    return {
        type = "vbox",
        style = {
            width = 25,
            background = theme.bg,
            border = "single",
            foreground = theme.border,
            paddingTop = 1,
        },
        children = children,
    }
end

-- Content area
local function Content()
    local state = lumina.useStore(store)
    local pageInfo = pages[state.currentPage]

    -- Load the page module
    local ok, pageModule = pcall(require, pageInfo.path)
    if not ok then
        return {
            type = "vbox",
            style = { flex = 1, padding = 2 },
            children = {
                { type = "text", content = "Error loading: " .. pageInfo.path, style = { foreground = theme.error } },
                { type = "text", content = tostring(pageModule), style = { foreground = theme.muted } },
            }
        }
    end

    -- Wrap page content in a scrollable container
    return {
        type = "vbox",
        props = { id = "content-scroll" },
        style = {
            flex = 1,
            overflow = "scroll",
            background = theme.bg,
        },
        children = { pageModule.render() }
    }
end

-- App root
local App = lumina.defineComponent({
    name = "ComponentLib",
    render = function()
        return {
            type = "hbox",
            style = { background = theme.bg },
            children = {
                Sidebar(),
                Content(),
            }
        }
    end
})

lumina.mount(App)
lumina.run()
