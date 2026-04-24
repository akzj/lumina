-- Table component for Lumina — data table with columns
local lumina = require("lumina")

local function fitWidth(str, width)
    str = tostring(str or "")
    if #str > width then return str:sub(1, width - 1) .. "…" end
    return str .. string.rep(" ", width - #str)
end

local Table = lumina.defineComponent({
    name = "Table",
    init = function(props)
        return {
            columns = props.columns or {},
            data = props.data or {},
            selectedRow = props.selectedRow or 0,
            onSelect = props.onSelect,
            style = props.style or {},
        }
    end,
    render = function(instance)
        local columns = instance.columns or {}
        local data = instance.data or {}
        local selected = instance.selectedRow or 0
        local fg = instance.style and instance.style.foreground or "#FFFFFF"
        local bg = instance.style and instance.style.background
        local children = {}
        local header = ""
        for i, col in ipairs(columns) do
            if i > 1 then header = header .. "│" end
            header = header .. fitWidth(col.label or col.key or "", col.width or 10)
        end
        children[#children + 1] = { type = "text", content = header, foreground = "#00FFFF", bold = true }
        local sep = ""
        for i, col in ipairs(columns) do
            if i > 1 then sep = sep .. "┼" end
            sep = sep .. string.rep("─", col.width or 10)
        end
        children[#children + 1] = { type = "text", content = sep, foreground = "#555555" }
        for rowIdx, row in ipairs(data) do
            local rc = ""
            for i, col in ipairs(columns) do
                if i > 1 then rc = rc .. "│" end
                rc = rc .. fitWidth(row[col.key or ""] or "", col.width or 10)
            end
            local isSel = (rowIdx == selected)
            children[#children + 1] = {
                type = "text", content = rc,
                foreground = isSel and "#00FF00" or fg,
                bold = isSel, background = isSel and "#333333" or bg,
            }
        end
        return { type = "vbox", style = { border = instance.style and instance.style.border }, children = children }
    end
})

return Table
