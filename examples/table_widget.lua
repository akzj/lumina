-- examples/table_widget.lua — Go widget Table (pkg/widget/table.go)
--
-- Run: lumina examples/table_widget.lua
-- Keys: q / Ctrl+C quit. Click the bordered table to focus, then j/k or arrows.
--
-- Documents behavior implemented in table.go (plain text cells; no per-cell components).

lumina.app {
    id = "table-widget-demo",
    store = {
        hint = "Table uses autoFocus: j/k and arrows work without clicking first.",
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme()
        local hint = lumina.useStore("hint")

        local function line(fg, dim, h, s)
            return lumina.createElement("text", {
                foreground = fg or (t.text or "#CDD6F4"),
                dim = dim or false,
                style = { height = h or 1 },
            }, s)
        end

        local legend = {
            line(t.primary, false, 1, "  Go Table widget — props / behavior (see pkg/widget/table.go)"),
            line(t.muted, true, 1, "  columns: { header, key, width? } — width defaults to 10; pad/truncate to width."),
            line(t.muted, true, 1, "  rows: array of maps; value = tostring(row[key]); missing key => empty cell."),
            line(t.muted, true, 1, "  selectable: focusable root; j/Down & k/Up wrap; Enter also sets onChange payload."),
            line(t.muted, true, 1, "  striped: rows where 0-based index is odd get surface0 (1st row plain, 2nd striped, ...)."),
            line(t.muted, true, 1, "  Chrome: header row bold primary; rule line; rounded border; text cells only."),
            line(t.muted, true, 1, "  Selection: primary background + primaryDark text on the active row."),
            line(t.muted, true, 1, "  onChange(idx): idx is 0-based row index (number)."),
            line(t.muted, true, 1, "  autoFocus + selectable: table grabs focus after mount (otherwise click first)."),
            line(nil, false, 1, "  " .. hint),
        }

        local columns = {
            { header = "Name", key = "name", width = 12 },
            { header = "Role", key = "role", width = 10 },
            { header = "Status", key = "status", width = 8 },
            { header = "Note", key = "note", width = 14 },
            { header = "Qty", key = "qty", width = 5 },
        }

        local rows = {
            { name = "Alice", role = "Admin", status = "Active", note = "Full access", qty = 1 },
            { name = "Bob", role = "Dev", status = "Away", note = "Long text is clipped to column width.", qty = 3 },
            { name = "Carol", role = "QA", status = "Busy", note = "", qty = 12 },
            { name = "Dan", role = "Ops", status = "Idle", note = "no qty col", },
        }

        local tableEl = lumina.createElement("Table", {
            id = "demo-table",
            key = "demo-table",
            columns = columns,
            rows = rows,
            selectable = true,
            striped = true,
            autoFocus = true,
            onChange = function(idx)
                lumina.store.set("hint", string.format("onChange: row index=%s (0-based)", tostring(idx)))
            end,
        })

        local children = {}
        for i = 1, #legend do
            children[#children + 1] = legend[i]
        end
        children[#children + 1] = line(t.muted, true, 1, "  ")
        children[#children + 1] = tableEl

        return lumina.createElement("vbox", {
            style = {
                width = 80,
                height = 24,
                background = t.base or "#1E1E2E",
            },
        }, table.unpack(children))
    end,
}
