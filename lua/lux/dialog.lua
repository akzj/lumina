-- lua/lux/dialog.lua
-- Composable Dialog component with slot-based children.
-- Usage:
--   local Dialog = require("lux.dialog")
--   Dialog {
--       open = true,
--       Dialog.Title { "Confirm" },
--       Dialog.Content { "Are you sure?" },
--       Dialog.Actions { btn1, btn2 },
--   }

local Slot = require("lux.slot")

-- Create slot factories
local DialogTitle = Slot("title")
local DialogContent = Slot("content")
local DialogActions = Slot("actions")

-- Dialog component
local Dialog = lumina.defineComponent("LuxDialog", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local open = props.open
    if not open then
        return lumina.createElement("box", { style = { width = 0, height = 0 } })
    end

    local title = props.title or "Dialog"
    local width = props.width or 40
    local children = props.children or {}

    -- Extract slots from children
    local titleSlot = nil
    local contentSlot = nil
    local actionsSlot = nil
    local otherChildren = {}

    for _, child in ipairs(children) do
        if type(child) == "table" and child._slotName then
            if child._slotName == "title" then
                titleSlot = child
            elseif child._slotName == "content" then
                contentSlot = child
            elseif child._slotName == "actions" then
                actionsSlot = child
            end
        else
            otherChildren[#otherChildren + 1] = child
        end
    end

    -- Build dialog content
    local dialogChildren = {}

    -- Title
    if titleSlot and titleSlot.children and #titleSlot.children > 0 then
        local titleText = titleSlot.children[1]
        if type(titleText) == "string" then
            dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
                foreground = t.primary or "#F5C842",
                bold = true,
            }, titleText)
        else
            dialogChildren[#dialogChildren + 1] = titleText
        end
    else
        dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
            foreground = t.primary or "#F5C842",
            bold = true,
        }, title)
    end

    -- Divider after title
    local divWidth = width - 4
    if divWidth < 1 then divWidth = 1 end
    dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
        foreground = t.muted or "#8B9BB4",
        dim = true,
    }, string.rep("-", divWidth))

    -- Content
    if contentSlot and contentSlot.children and #contentSlot.children > 0 then
        for _, child in ipairs(contentSlot.children) do
            if type(child) == "string" then
                dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
                    foreground = t.text or "#E8EDF7",
                }, child)
            else
                dialogChildren[#dialogChildren + 1] = child
            end
        end
    elseif props.message then
        dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
            foreground = t.text or "#E8EDF7",
        }, props.message)
    end

    -- Other (non-slot) children
    for _, child in ipairs(otherChildren) do
        dialogChildren[#dialogChildren + 1] = child
    end

    -- Actions
    if actionsSlot and actionsSlot.children and #actionsSlot.children > 0 then
        -- Divider before actions
        dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
            foreground = t.muted or "#8B9BB4",
            dim = true,
        }, string.rep("-", divWidth))

        -- Actions in an hbox
        dialogChildren[#dialogChildren + 1] = lumina.createElement("hbox", {
            style = { gap = 1 },
        }, table.unpack(actionsSlot.children))
    end

    -- Wrap in a bordered box
    return lumina.createElement("vbox", {
        style = {
            border = "rounded",
            padding = 1,
            width = width,
            background = t.surface0 or "#141C2C",
        },
    }, table.unpack(dialogChildren))
end)

-- Attach slot factories
Dialog.Title = DialogTitle
Dialog.Content = DialogContent
Dialog.Actions = DialogActions

return Dialog
