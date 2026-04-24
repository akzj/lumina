# Lumina Examples

## Running Examples

### Local Terminal
```bash
go run ./cmd/lumina examples/dashboard/main.lua
go run ./cmd/lumina examples/todo/main.lua
go run ./cmd/lumina examples/chat/main.lua
go run ./cmd/lumina examples/file-explorer/main.lua
go run ./cmd/lumina examples/components-showcase/main.lua
```

### Web Browser
```bash
go run ./cmd/lumina-server examples/dashboard/main.lua
# Then open http://localhost:8080 (if web server mode is configured)
```

## Examples

| Example | Description |
|---------|-------------|
| **dashboard** | Admin dashboard with router, stat cards, tables, charts, theme & i18n switcher |
| **todo** | Classic todo app with store, filters, keyboard shortcuts |
| **file-explorer** | File browser with tree, preview panel, breadcrumbs |
| **chat** | Chat UI with messages, input, auto-scroll, timestamps |
| **components-showcase** | Gallery of all 47 shadcn components with interactive demos |

## Architecture

All examples use the same Lumina API:

```lua
local lumina = require("lumina")
local shadcn = require("shadcn")

-- Define components with React-style hooks
local App = lumina.defineComponent({
    render = function(self)
        local count, setCount = lumina.useState(0)
        return {
            type = "vbox",
            children = {
                shadcn.Button({ label = "Count: " .. count, onClick = function() setCount(count + 1) end }),
            }
        }
    end
})

lumina.mount(App)
lumina.run()  -- local terminal
-- OR: lumina.serve(8080)  -- web browser
```
