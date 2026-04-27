# Dashboard Example (Lumina v2)

Admin-style TUI: sidebar (home / users / settings), stat cards, user table,
theme + locale toggles (local `useState` only).

This example uses the **v2 Lua bridge** (`lumina.createElement`, `lumina.useState`,
`lumina.createComponent`, `lumina.quit`). It does **not** use v1 APIs (`require("lumina")`,
`createStore`, `defineComponent`, `mount`, `run`, `onKey`, `i18n`, `setTheme`).

## Run

```bash
lumina-v2 examples/dashboard/main.lua
```

Web:

```bash
lumina-v2 --web :8080 examples/dashboard/main.lua
```

From repo root (if `lumina-v2` is not on `PATH`):

```bash
go run ./cmd/lumina-v2 examples/dashboard/main.lua
```

## Keys

Focus the dashboard root (click the hint bar) if shortcuts do not fire.

- `1` / `2` / `3` — Home, Users, Settings  
- `T` — Toggle theme palette (mocha / latte)  
- `L` — Toggle English / Chinese strings  
- `Q` — Quit  
