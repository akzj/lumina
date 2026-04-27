-- shadcn/table — Data table
local lumina = require("lumina")

local Table = lumina.defineComponent({
    name = "ShadcnTable",
    init = function(props)
        return {
            headers = props.headers or {},
            rows = props.rows or {},
            colWidth = props.colWidth or 15,
        }
    end,
    render = function(self)
        local children = {}
        local cw = self.colWidth

        -- Helper: pad/truncate string to width
        local function fit(s, w)
            s = tostring(s or "")
            if #s > w then return string.sub(s, 1, w - 1) .. "…" end
            return s .. string.rep(" ", w - #s)
        end

        -- Header row
        if #self.headers > 0 then
            local hdr = ""
            for _, h in ipairs(self.headers) do
                hdr = hdr .. fit(h, cw)
            end
            children[#children + 1] = {
                type = "text",
                content = hdr,
                style = { foreground = "#94A3B8", bold = true },
            }
            children[#children + 1] = {
                type = "text",
                content = string.rep("─", cw * #self.headers),
                style = { foreground = "#334155" },
            }
        end

        -- Data rows
        for _, row in ipairs(self.rows) do
            local line = ""
            for _, cell in ipairs(row) do
                line = line .. fit(cell, cw)
            end
            children[#children + 1] = {
                type = "text",
                content = line,
                style = { foreground = "#E2E8F0" },
            }
        end

        return {
            type = "vbox",
            children = children,
        }
    end
})

return Table
