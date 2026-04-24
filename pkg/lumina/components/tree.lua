-- Tree component for Lumina — collapsible tree view
local lumina = require("lumina")

local Tree = lumina.defineComponent({
    name = "Tree",
    init = function(props)
        return {
            data = props.data or {},
            onSelect = props.onSelect,
            expanded = props.expanded or {},
            indent = props.indent or 2,
            style = props.style or {},
        }
    end,
    render = function(instance)
        local data = instance.data or {}
        local expanded = instance.expanded or {}
        local indent = instance.indent or 2
        local fg = instance.style and instance.style.foreground or "#FFFFFF"
        local children = {}
        local function addNodes(nodes, depth)
            for _, node in ipairs(nodes) do
                local prefix = string.rep(" ", depth * indent)
                local hasKids = node.children and #node.children > 0
                local icon = hasKids and (expanded[node.label] and "▾ " or "▸ ") or "  "
                children[#children + 1] = {
                    type = "text",
                    content = prefix .. icon .. (node.label or ""),
                    foreground = hasKids and "#00FFFF" or fg,
                }
                if hasKids and expanded[node.label] then
                    addNodes(node.children, depth + 1)
                end
            end
        end
        addNodes(data, 0)
        if #children == 0 then
            children[#children + 1] = { type = "text", content = "(empty)", foreground = "#555555" }
        end
        return { type = "vbox", style = instance.style, children = children }
    end
})

return Tree
