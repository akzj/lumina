-- shadcn/table — Data table
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    surface = "#313244",
}

local Table = lumina.defineComponent({
    name = "ShadcnTable",

    init = function(props)
        return {
            headers = props.headers or {},
            rows = props.rows or {},
            colWidth = props.colWidth or 15,
            id = props.id,
            className = props.className,
            style = props.style,
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
                style = { foreground = c.primary, bold = true },
            }
            children[#children + 1] = {
                type = "text",
                content = string.rep("─", cw * #self.headers),
                style = { foreground = c.border },
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
                style = { foreground = c.fg },
            }
        end

        local style = {}
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return Table