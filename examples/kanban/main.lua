-- Kanban Board — A Trello-like task management board
-- Usage: lumina examples/kanban/main.lua
--
-- Features:
--   • 3 columns: Todo, In Progress, Done
--   • Cards with title, description, tags
--   • Add new cards with 'n' key
--   • Move cards between columns with h/l
--   • Navigate with j/k
--   • Delete cards with 'd'
--   • Store-based state management

local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

-- ─── Store ────────────────────────────────────────────────────────────

local nextId = 7

local store = lumina.createStore({
    state = {
        columns = {
            {
                id = "todo",
                title = "📋 Todo",
                color = "#89B4FA",
                cards = {
                    { id = 1, title = "Design landing page", desc = "Create wireframes and mockups", tags = {"design", "ui"} },
                    { id = 2, title = "Write API docs", desc = "Document all REST endpoints", tags = {"docs"} },
                    { id = 3, title = "Set up CI/CD", desc = "GitHub Actions pipeline", tags = {"devops"} },
                },
            },
            {
                id = "progress",
                title = "🔄 In Progress",
                color = "#F9E2AF",
                cards = {
                    { id = 4, title = "Build auth system", desc = "JWT + OAuth2 integration", tags = {"backend", "security"} },
                    { id = 5, title = "Component library", desc = "Reusable UI components", tags = {"frontend"} },
                },
            },
            {
                id = "done",
                title = "✅ Done",
                color = "#A6E3A1",
                cards = {
                    { id = 6, title = "Project setup", desc = "Initialize repo and deps", tags = {"devops"} },
                },
            },
        },
        selectedCol = 1,
        selectedCard = 1,
        showNewCardDialog = false,
        newCardTitle = "",
        newCardDesc = "",
    },
    actions = {
        selectCol = function(state, col)
            state.selectedCol = math.max(1, math.min(col, #state.columns))
            local cards = state.columns[state.selectedCol].cards
            state.selectedCard = math.max(1, math.min(state.selectedCard, #cards))
        end,
        selectCard = function(state, card)
            local cards = state.columns[state.selectedCol].cards
            state.selectedCard = math.max(1, math.min(card, #cards))
        end,
        moveCardRight = function(state)
            local col = state.selectedCol
            if col >= #state.columns then return end
            local cards = state.columns[col].cards
            if #cards == 0 then return end
            local card = table.remove(cards, state.selectedCard)
            table.insert(state.columns[col + 1].cards, card)
            if state.selectedCard > #cards and #cards > 0 then
                state.selectedCard = #cards
            end
        end,
        moveCardLeft = function(state)
            local col = state.selectedCol
            if col <= 1 then return end
            local cards = state.columns[col].cards
            if #cards == 0 then return end
            local card = table.remove(cards, state.selectedCard)
            table.insert(state.columns[col - 1].cards, card)
            if state.selectedCard > #cards and #cards > 0 then
                state.selectedCard = #cards
            end
        end,
        deleteCard = function(state)
            local cards = state.columns[state.selectedCol].cards
            if #cards == 0 then return end
            table.remove(cards, state.selectedCard)
            if state.selectedCard > #cards and #cards > 0 then
                state.selectedCard = #cards
            end
        end,
        toggleNewCard = function(state)
            state.showNewCardDialog = not state.showNewCardDialog
            state.newCardTitle = ""
            state.newCardDesc = ""
        end,
        setNewCardTitle = function(state, title)
            state.newCardTitle = title
        end,
        setNewCardDesc = function(state, desc)
            state.newCardDesc = desc
        end,
        addCard = function(state)
            if state.newCardTitle == "" then return end
            local card = {
                id = nextId,
                title = state.newCardTitle,
                desc = state.newCardDesc,
                tags = {"new"},
            }
            nextId = nextId + 1
            table.insert(state.columns[state.selectedCol].cards, card)
            state.showNewCardDialog = false
            state.newCardTitle = ""
            state.newCardDesc = ""
        end,
    },
})

-- ─── Tag Badge ────────────────────────────────────────────────────────

local tagColors = {
    design = "#F5C2E7",
    ui = "#89DCEB",
    docs = "#89B4FA",
    devops = "#A6E3A1",
    backend = "#CBA6F7",
    security = "#F38BA8",
    frontend = "#F9E2AF",
    new = "#FAB387",
}

local TagBadge = lumina.defineComponent({
    name = "TagBadge",
    render = function(self)
        local tag = self.props.tag or ""
        local color = tagColors[tag] or "#585B70"
        return {
            type = "text",
            content = " " .. tag .. " ",
            style = { foreground = "#1E1E2E", background = color },
        }
    end,
})

-- ─── Kanban Card ──────────────────────────────────────────────────────

local KanbanCard = lumina.defineComponent({
    name = "KanbanCard",
    render = function(self)
        local card = self.props.card
        local isSelected = self.props.selected

        local tagNodes = {}
        for _, tag in ipairs(card.tags or {}) do
            table.insert(tagNodes, lumina.createElement(TagBadge, { tag = tag }))
        end

        return {
            type = "vbox",
            style = {
                border = isSelected and "double" or "rounded",
                padding = 1,
                background = isSelected and "#313244" or "#1E1E2E",
                marginBottom = 1,
            },
            children = {
                { type = "text", content = card.title, style = { bold = true, foreground = "#CDD6F4" } },
                { type = "text", content = card.desc or "", style = { foreground = "#A6ADC8" } },
                { type = "hbox", style = { gap = 1 }, children = tagNodes },
            },
        }
    end,
})

-- ─── Column ───────────────────────────────────────────────────────────

local KanbanColumn = lumina.defineComponent({
    name = "KanbanColumn",
    render = function(self)
        local col = self.props.column
        local isActiveCol = self.props.active
        local selectedCard = self.props.selectedCard

        local cardNodes = {}
        for i, card in ipairs(col.cards) do
            table.insert(cardNodes, lumina.createElement(KanbanCard, {
                card = card,
                selected = isActiveCol and i == selectedCard,
            }))
        end

        if #col.cards == 0 then
            table.insert(cardNodes, {
                type = "text",
                content = "  (empty)",
                style = { foreground = "#585B70" },
            })
        end

        return {
            type = "vbox",
            style = {
                border = isActiveCol and "double" or "rounded",
                padding = 1,
                width = "33%",
                background = "#1E1E2E",
            },
            children = {
                -- Column header
                {
                    type = "hbox",
                    children = {
                        { type = "text", content = col.title, style = { bold = true, foreground = col.color } },
                        { type = "text", content = " (" .. #col.cards .. ")", style = { foreground = "#585B70" } },
                    },
                },
                { type = "text", content = "" },
                -- Cards
                { type = "vbox", children = cardNodes },
            },
        }
    end,
})

-- ─── New Card Dialog ──────────────────────────────────────────────────

local NewCardDialog = lumina.defineComponent({
    name = "NewCardDialog",
    render = function(self)
        local state = store.getState()

        if not state.showNewCardDialog then
            return { type = "text", content = "" }
        end

        return lumina.createElement(shadcn.Dialog, {
            open = true,
            title = "New Card",
            children = {
                { type = "text", content = "Title:", style = { foreground = "#A6ADC8", bold = true } },
                lumina.createElement(shadcn.Input, {
                    value = state.newCardTitle,
                    placeholder = "Card title...",
                    onChange = function(v) store.dispatch("setNewCardTitle", v) end,
                }),
                { type = "text", content = "" },
                { type = "text", content = "Description:", style = { foreground = "#A6ADC8", bold = true } },
                lumina.createElement(shadcn.Input, {
                    value = state.newCardDesc,
                    placeholder = "Card description...",
                    onChange = function(v) store.dispatch("setNewCardDesc", v) end,
                }),
                { type = "text", content = "" },
                {
                    type = "hbox",
                    style = { gap = 1 },
                    children = {
                        lumina.createElement(shadcn.Button, {
                            label = "✓ Add Card",
                            variant = "default",
                            onClick = function() store.dispatch("addCard") end,
                        }),
                        lumina.createElement(shadcn.Button, {
                            label = "✕ Cancel",
                            variant = "outline",
                            onClick = function() store.dispatch("toggleNewCard") end,
                        }),
                    },
                },
            },
        })
    end,
})

-- ─── App Layout ───────────────────────────────────────────────────────

local App = lumina.defineComponent({
    name = "KanbanBoard",
    render = function(self)
        local state = store.getState()

        local columnNodes = {}
        for i, col in ipairs(state.columns) do
            table.insert(columnNodes, lumina.createElement(KanbanColumn, {
                column = col,
                active = i == state.selectedCol,
                selectedCard = state.selectedCard,
            }))
        end

        return {
            type = "vbox",
            style = { background = "#1E1E2E" },
            children = {
                -- Title bar
                {
                    type = "hbox",
                    style = { padding = 1, background = "#313244" },
                    children = {
                        { type = "text", content = "📌 Lumina Kanban", style = { bold = true, foreground = "#F5C2E7" } },
                        { type = "text", content = "  h/l=col  j/k=card  m=move→  M=move←  n=new  d=del  q=quit",
                          style = { foreground = "#585B70" } },
                    },
                },
                -- Board columns
                {
                    type = "hbox",
                    style = { padding = 1 },
                    children = columnNodes,
                },
                -- Dialog overlay
                lumina.createElement(NewCardDialog, {}),
            },
        }
    end,
})

-- ─── Keyboard Shortcuts ───────────────────────────────────────────────

lumina.onKey("h", function()
    local state = store.getState()
    store.dispatch("selectCol", state.selectedCol - 1)
end)

lumina.onKey("l", function()
    local state = store.getState()
    store.dispatch("selectCol", state.selectedCol + 1)
end)

lumina.onKey("j", function()
    local state = store.getState()
    store.dispatch("selectCard", state.selectedCard + 1)
end)

lumina.onKey("k", function()
    local state = store.getState()
    store.dispatch("selectCard", state.selectedCard - 1)
end)

lumina.onKey("m", function()
    store.dispatch("moveCardRight")
end)

lumina.onKey("M", function()
    store.dispatch("moveCardLeft")
end)

lumina.onKey("n", function()
    store.dispatch("toggleNewCard")
end)

lumina.onKey("d", function()
    store.dispatch("deleteCard")
end)

lumina.onKey("q", function()
    lumina.quit()
end)

lumina.mount(App)
lumina.run()
