-- examples/messages.lua
-- A message list app demonstrating nested scrolling (2 layers).
--
-- Demonstrates:
--   • Outer scroll: message list scrolls when there are many messages
--   • Inner scroll: each message body scrolls independently if text is long
--   • lumina.defineComponent for MessageItem
--   • lumina.useStore for state management
--   • lumina.getTheme() for themed colors
--   • Global keybindings (j/k=select, q=quit)
--
-- Usage: lumina examples/messages.lua

local lux = require("lux")
local Divider = lux.Divider
local Badge = lux.Badge

-- Message item component: shows sender + content in a fixed-height box.
-- Long content scrolls within the box (inner scroll).
local MessageItem = lumina.defineComponent("MessageItem", function(props)
    local t = lumina.getTheme()
    local msg = props.message or {}
    local selected = props.selected
    local scrollY = props.scrollY or 0

    local bg = selected and t.surface0 or t.base
    local fg = selected and t.primary or t.text
    local prefix = selected and "▸ " or "  "

    -- Build content lines from message text
    local contentLines = {}
    for line in (msg.content or ""):gmatch("[^\n]+") do
        contentLines[#contentLines + 1] = lumina.createElement("text", {
            foreground = t.text,
        }, "    " .. line)
    end
    -- If no newlines, just use the content as a single line
    if #contentLines == 0 then
        contentLines[1] = lumina.createElement("text", {
            foreground = t.text,
        }, "    " .. (msg.content or ""))
    end

    return lumina.createElement("vbox", {
        style = {
            height = 5,
            overflow = "scroll",
            background = bg,
        },
        scrollY = scrollY,
    },
        lumina.createElement("text", { bold = true, foreground = fg },
            prefix .. (msg.sender or "Unknown") .. "  " .. (msg.time or "")),
        table.unpack(contentLines)
    )
end)

-- Sample messages: mix of short and long content
local initialMessages = {
    {
        id = 1,
        sender = "Alice",
        time = "10:30",
        content = "Hey! How's the project going? I was thinking we could refactor the scroll system to support nested containers. Each container would clip its children independently and handle scroll events at the correct depth. This would enable complex layouts like chat apps, code editors with split panes, and dashboard widgets with internal scrolling.",
    },
    {
        id = 2,
        sender = "Bob",
        time = "10:35",
        content = "Quick update: tests pass!",
    },
    {
        id = 3,
        sender = "Charlie",
        time = "10:42",
        content = "Here's my detailed analysis of the rendering pipeline:\n1. Parse Lua descriptor into VNode tree\n2. Reconcile against previous tree (diff)\n3. Layout pass: compute positions and sizes\n4. Paint pass: write cells to buffer\n5. Flush: output ANSI to terminal\n6. Each step can be optimized independently\n7. The scroll container clips at step 4\n8. Nested scrolls accumulate offsets\n9. Hit testing reverses the offset chain\n10. This enables arbitrary nesting depth",
    },
    {
        id = 4,
        sender = "Diana",
        time = "11:00",
        content = "Meeting at 3pm today.",
    },
    {
        id = 5,
        sender = "Eve",
        time = "11:15",
        content = "I've been working on the widget system and found several interesting patterns:\n- Controlled vs uncontrolled inputs\n- Event bubbling through component boundaries\n- Focus management across scroll containers\n- State preservation during re-renders\n- Lazy rendering for off-screen content\n- Accessibility hooks for screen readers\nLet me know if you want to discuss any of these topics in detail.",
    },
    {
        id = 6,
        sender = "Frank",
        time = "11:30",
        content = "LGTM 👍",
    },
    {
        id = 7,
        sender = "Grace",
        time = "12:00",
        content = "Performance report for the new renderer:\nFrame time: 2.1ms average (down from 8.4ms)\nMemory: 12MB steady state\nGC pauses: <0.5ms p99\nScroll latency: <16ms (60fps target met)\nThe optimization of the paint pass made the biggest difference. By only repainting dirty rectangles, we avoid redundant ANSI output. Combined with the cursor parking fix for IME, the rendering is now rock solid across Terminal.app, iTerm2, and Cursor.",
    },
}

lumina.app {
    id = "messages-app",
    store = {
        messages = initialMessages,
        selectedIdx = 1,
        innerScrollY = {},  -- per-message scroll offsets
    },
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
        ["j"] = function()
            local msgs = lumina.store.get("messages")
            local idx = lumina.store.get("selectedIdx")
            if idx < #msgs then
                lumina.store.set("selectedIdx", idx + 1)
            end
        end,
        ["k"] = function()
            local idx = lumina.store.get("selectedIdx")
            if idx > 1 then
                lumina.store.set("selectedIdx", idx - 1)
            end
        end,
        -- Scroll inner message content
        ["ArrowDown"] = function()
            local idx = lumina.store.get("selectedIdx")
            local scrolls = lumina.store.get("innerScrollY") or {}
            local newScrolls = {}
            for k, v in pairs(scrolls) do newScrolls[k] = v end
            newScrolls[idx] = (newScrolls[idx] or 0) + 1
            lumina.store.set("innerScrollY", newScrolls)
        end,
        ["ArrowUp"] = function()
            local idx = lumina.store.get("selectedIdx")
            local scrolls = lumina.store.get("innerScrollY") or {}
            local newScrolls = {}
            for k, v in pairs(scrolls) do newScrolls[k] = v end
            newScrolls[idx] = math.max(0, (newScrolls[idx] or 0) - 1)
            lumina.store.set("innerScrollY", newScrolls)
        end,
    },

    render = function()
        local t = lumina.getTheme()
        local messages = lumina.useStore("messages")
        local selectedIdx = lumina.useStore("selectedIdx")
        local innerScrollY = lumina.useStore("innerScrollY") or {}

        -- Header
        local header = lumina.createElement("hbox", { style = { height = 1 } },
            lumina.createElement("text", { foreground = t.primary, bold = true },
                " 💬 Messages "),
            Badge { label = tostring(#messages), variant = "default", key = "msg-count" }
        )

        -- Build message list items
        local listItems = {}
        for i, msg in ipairs(messages) do
            listItems[#listItems + 1] = MessageItem {
                message = msg,
                selected = (i == selectedIdx),
                scrollY = innerScrollY[i] or 0,
                key = "msg-" .. msg.id,
            }
        end

        -- Outer scroll: auto-scroll to keep selected message visible
        local msgH = 5  -- each MessageItem has height=5
        local outerScrollY = 0
        if selectedIdx > 1 then
            outerScrollY = (selectedIdx - 1) * msgH
        end

        local messageList = lumina.createElement("vbox", {
            style = { flex = 1, overflow = "scroll" },
            scrollY = outerScrollY,
        }, table.unpack(listItems))

        -- Footer
        local footer = lumina.createElement("text", {
            foreground = t.muted, dim = true,
            style = { height = 1, background = t.surface0 },
        }, " [j/k] Select  [↑/↓] Scroll message  [q] Quit")

        -- Main layout
        return lumina.createElement("vbox", {
            id = "messages-root",
            style = { background = t.base, border = "rounded" },
            focusable = true,
        },
            header,
            Divider { width = 60, key = "div-top" },
            messageList,
            Divider { width = 60, key = "div-bottom" },
            footer
        )
    end,
}
