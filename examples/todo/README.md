# Todo App Example

Classic todo application demonstrating:
- useState for local state
- createStore for global state
- Keyboard shortcuts (a=add, d=delete, space=toggle, tab=filter)
- Filtered views (all/active/completed)

## Run
```bash
go run ./cmd/lumina examples/todo/main.lua
```

## Keys
- `a` — Add new todo
- `d` — Delete selected todo
- `space` — Toggle done/undone
- `j/k` — Navigate up/down
- `tab` — Cycle filter (All → Active → Completed)
- `q` — Quit
