# Lux Component Library — API Reference

> Lux is the high-level component library for [Lumina](https://github.com/akzj/lumina).
> All components are pure Lua, built on `lumina.defineComponent` and `lumina.createElement`.

---

## Installation

```lua
-- Import entire library
local lux = require("lux")

-- Or import individual components
local Card = require("lux.card")
local Badge = require("lux.badge")
local DataGrid = require("lux.data_grid")
```

All Lux modules are embedded in the Go binary from `lua/lux/*.lua` (`lua/lux/embed.go`) and registered in `pkg/lux_modules.go` — no external file dependencies at runtime.

---

## Components

### Card

A bordered container with optional title. Ideal for grouping related content.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `title` | string | `nil` | Bold title text displayed at the top |
| `border` | string | `"rounded"` | Border style: `"rounded"`, `"single"`, `"double"`, `"none"` |
| `padding` | number | `1` | Internal padding |
| `bg` | string | `""` | Background color (hex) |
| `children` | table | `{}` | Child elements |

**Example:**
```lua
local Card = require("lux.card")

Card {
    title = "System Status",
    border = "rounded",
    padding = 1,
    lumina.createElement("text", {}, "All systems operational"),
}
```

---

### Badge

A small themed label for status indicators.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `label` | string | `""` | Badge text content |
| `variant` | string | `"default"` | Color variant: `"default"`, `"success"`, `"warning"`, `"error"` |

**Variant colors** (Catppuccin Mocha defaults):
- `default` — primary blue (`#89B4FA`)
- `success` — green (`#A6E3A1`)
- `warning` — yellow (`#F9E2AF`)
- `error` — red (`#F38BA8`)

**Example:**
```lua
local Badge = require("lux.badge")

Badge { label = "Online", variant = "success" }
Badge { label = "3 errors", variant = "error" }
```

---

### Divider

A horizontal line separator.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `char` | string | `"─"` | Character used to draw the line |
| `width` | number | `40` | Width in columns |
| `fg` | string | theme `surface1` | Foreground color |

**Example:**
```lua
local Divider = require("lux.divider")

Divider { width = 60, char = "═" }
```

---

### Progress

A horizontal progress bar with percentage label.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `value` | number | `0` | Progress percentage (0–100) |
| `width` | number | `20` | Width of the bar in columns (excluding label) |

**Example:**
```lua
local Progress = require("lux.progress")

Progress { value = 75, width = 30 }
-- Renders: ██████████████████████░░░░░░░░░ 75%
```

---

### Spinner

An animated loading indicator with customizable label.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `label` | string | `"Loading..."` | Text displayed next to the spinner |

**Notes:**
- Uses `lumina.useEffect` + `lumina.setInterval` (80ms) for animation
- Braille dot frames: `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏`
- Cleanup is automatic on unmount

**Example:**
```lua
local Spinner = require("lux.spinner")

Spinner { label = "Fetching data..." }
```

---

### TextInput

A themed text input field with optional label, error, and helper text. Wraps the native `input` element.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `id` | string | `nil` | Root element ID |
| `inputId` | string | `id .. "-input"` | Native input element ID |
| `value` | string | `""` | Current input value |
| `placeholder` | string | `""` | Placeholder text |
| `label` | string | `nil` | Label text above the input |
| `error` | string | `nil` | Error message (red, replaces helper) |
| `helperText` | string | `nil` | Helper text below the input |
| `width` | number | `30` | Input width in columns |
| `disabled` | boolean | `false` | Disable input (dims text, removes focus) |
| `autoFocus` | boolean | `false` | Auto-focus on mount |
| `onChange` | function | `nil` | `function(value)` — called on text change |
| `onSubmit` | function | `nil` | `function(value)` — called on Enter |
| `onFocus` | function | `nil` | Called when input gains focus |
| `onBlur` | function | `nil` | Called when input loses focus |

**Example:**
```lua
local TextInput = require("lux.text_input")

TextInput {
    label = "Username",
    value = username,
    placeholder = "Enter username...",
    helperText = "3-20 characters",
    onChange = function(val) setUsername(val) end,
    onSubmit = function(val) handleLogin(val) end,
}
```

---

### Dialog

A composable modal dialog with slot-based Title, Content, and Actions.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `open` | boolean | `false` | Whether dialog is visible |
| `title` | string | `"Dialog"` | Fallback title (used if no `Dialog.Title` slot) |
| `message` | string | `nil` | Fallback content text |
| `width` | number | `40` | Dialog width in columns |
| `children` | table | `{}` | Slot children (see below) |

**Slot Components:**
- `Dialog.Title { "text" }` — title slot
- `Dialog.Content { "text", element, ... }` — content slot
- `Dialog.Actions { btn1, btn2, ... }` — actions row (hbox, gap=1)

**Example:**
```lua
local Dialog = require("lux.dialog")

Dialog {
    open = showDialog,
    Dialog.Title { "Confirm Delete" },
    Dialog.Content { "Are you sure you want to delete this item?" },
    Dialog.Actions {
        lumina.createElement("button", { label = "Cancel", onClick = cancel }),
        lumina.createElement("button", { label = "Delete", onClick = doDelete }),
    },
}
```

---

### Layout

A standard TUI app structure with Header, Sidebar, Main, and Footer slots.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `width` | number | `nil` | Root width (auto if nil) |
| `height` | number | `nil` | Root height (auto if nil) |
| `children` | table | `{}` | Slot children (see below) |

**Slot Components:**

| Slot | Props | Description |
|------|-------|-------------|
| `Layout.Header { ... }` | `height` (default 1), `background`/`bg` | Top bar |
| `Layout.Footer { ... }` | `height` (default 1), `background`/`bg` | Bottom bar |
| `Layout.Sidebar { ... }` | `width` (default 20), `border`, `background`/`bg` | Left panel |
| `Layout.Main { ... }` | — | Main content area (flex: 1) |

**Example:**
```lua
local Layout = require("lux.layout")

Layout {
    Layout.Header { height = 1,
        lumina.createElement("text", { bold = true }, " My App"),
    },
    Layout.Sidebar { width = 20,
        NavMenu {},
    },
    Layout.Main {
        ContentArea {},
    },
    Layout.Footer { height = 1,
        lumina.createElement("text", { dim = true }, " Ready"),
    },
}
```

---

### CommandPalette

A searchable command list overlay with keyboard navigation.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `commands` | table | `{}` | Array of `{ title = string, action = function }` |
| `width` | number | `50` | Palette width |
| `maxHeight` | number | `15` | Maximum height before truncation |
| `onClose` | function | `nil` | Called on Escape or after command execution |

**Keyboard:**
| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Execute selected command |
| `Escape` | Close palette |
| `Backspace` | Delete last character from query |
| Any char | Append to search query |

**Example:**
```lua
local CommandPalette = require("lux.command_palette")

CommandPalette {
    commands = {
        { title = "Save File", action = function() save() end },
        { title = "Open File", action = function() openDialog() end },
        { title = "Quit", action = function() lumina.quit() end },
    },
    onClose = function() setPaletteOpen(false) end,
}
```

---

### ListView

A scrollable list with rich row rendering and keyboard navigation.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `id` | string | `nil` | Element ID |
| `rows` | table | `{}` | Array of row data objects |
| `renderRow` | function | **required** | `function(row, index, ctx)` → element. `ctx.selected` is boolean |
| `height` | number | `10` | Visible height in rows |
| `width` | number | `nil` | Width (auto if nil) |
| `rowHeight` | number | `1` | Height of each row |
| `selectedIndex` | number | `1` | Currently selected row (1-based) |
| `empty` | string | `"No items"` | Text shown when `rows` is empty |
| `onChangeIndex` | function | `nil` | `function(newIndex)` — called on ↑/↓ navigation |
| `onActivate` | function | `nil` | `function(index, row)` — called on Enter |

**Keyboard:**
| Key | Action |
|-----|--------|
| `↑` / `k` | Select previous row |
| `↓` / `j` | Select next row |
| `Enter` | Activate selected row |

**Example:**
```lua
local ListView = require("lux.list")

ListView {
    rows = items,
    height = 15,
    selectedIndex = selected,
    renderRow = function(row, i, ctx)
        local prefix = ctx.selected and "> " or "  "
        return lumina.createElement("text", {
            bold = ctx.selected,
            foreground = ctx.selected and "#89B4FA" or "#CDD6F4",
            style = { height = 1 },
        }, prefix .. row.name)
    end,
    onChangeIndex = function(idx) setSelected(idx) end,
    onActivate = function(idx, row) openItem(row) end,
}
```

---

### DataGrid

A full-featured data table with fixed header, scrollable body, sorting, and multi-select.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `columns` | table | `{}` | Array of column definitions (see below) |
| `rows` | table | `{}` | Array of row data objects |
| `height` | number | `12` | Total grid height (header + separator + body) |
| `width` | number | `nil` | Total grid width |
| `rowHeight` | number | `1` | Height per row |
| `selectedIndex` | number | `1` | Focused row (1-based) |
| `selectionMode` | string | `"single"` | `"single"` or `"multi"` |
| `selectedIds` | table | `{}` | Array of selected row IDs (multi-select) |
| `getRowId` | function | `tostring(index)` | `function(row, index)` → unique ID string |
| `sort` | table | `nil` | `{ column = "col_id", direction = "asc"|"desc" }` |
| `renderCell` | function | `nil` | `function(row, rowIndex, column, ctx)` → element |
| `renderHeaderCell` | function | `nil` | `function(column, ctx)` → element |
| `onChangeIndex` | function | `nil` | `function(newIndex)` |
| `onSelectionChange` | function | `nil` | `function(selectedIds)` |
| `onSortChange` | function | `nil` | `function(sort)` |

**Column Definition:**
```lua
{
    id = "name",         -- unique column identifier
    header = "Name",     -- display header text
    key = "name",        -- row field to access (default: same as id)
    width = 20,          -- column width in characters
    align = "left",      -- "left", "center", "right"
    sortable = true,     -- enable sort on this column
}
```

**Keyboard:**
| Key | Action |
|-----|--------|
| `↑` / `k` | Move focus up |
| `↓` / `j` | Move focus down |
| `Enter` | Activate row |
| `Space` | Toggle selection (multi-select mode) |
| `a` | Select all / deselect all (multi-select mode) |

**Example:**
```lua
local DataGrid = require("lux.data_grid")

DataGrid {
    columns = {
        { id = "name", header = "Name", width = 20, sortable = true },
        { id = "status", header = "Status", width = 10 },
        { id = "cpu", header = "CPU %", width = 8, align = "right" },
    },
    rows = processes,
    height = 20,
    selectedIndex = focusedRow,
    selectionMode = "multi",
    selectedIds = checkedIds,
    sort = { column = "cpu", direction = "desc" },
    onChangeIndex = function(idx) setFocusedRow(idx) end,
    onSelectionChange = function(ids) setCheckedIds(ids) end,
    onSortChange = function(s) setSort(s) end,
}
```

---

### Tabs

Tabbed navigation with keyboard support and disabled tab handling.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `id` | string | `nil` | Element ID |
| `tabs` | table | `{}` | Array of `{ id, label, disabled }` |
| `activeTab` | string | `nil` | ID of the active tab |
| `onTabChange` | function | `nil` | `function(tabId)` — called when tab changes |
| `renderContent` | function | `nil` | `function(activeTabId)` → element for content area |
| `width` | number | `40` | Width (for separator line) |
| `height` | number | `10` | Total height |
| `autoFocus` | boolean | `false` | Auto-focus on mount |

**Keyboard:**
| Key | Action |
|-----|--------|
| `←` / `h` | Previous non-disabled tab |
| `→` / `l` | Next non-disabled tab |
| `Home` | First non-disabled tab |
| `End` | Last non-disabled tab |

**Example:**
```lua
local Tabs = require("lux.tabs")

Tabs {
    tabs = {
        { id = "general", label = "General" },
        { id = "network", label = "Network" },
        { id = "advanced", label = "Advanced", disabled = true },
    },
    activeTab = currentTab,
    onTabChange = function(id) setCurrentTab(id) end,
    renderContent = function(tabId)
        if tabId == "general" then return GeneralPanel {} end
        if tabId == "network" then return NetworkPanel {} end
    end,
}
```

---

### Alert

A themed notification banner with optional dismiss button.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `id` | string | `nil` | Element ID |
| `variant` | string | `"info"` | `"info"`, `"success"`, `"warning"`, `"error"` |
| `title` | string | variant name | Alert title |
| `message` | string | `nil` | Body text |
| `dismissible` | boolean | `false` | Show dismiss (✕) button |
| `onDismiss` | function | `nil` | Called when dismiss is clicked |
| `width` | number | `nil` | Width (auto if nil) |
| `height` | number | auto | Height (1 if no message, 2 with message) |

**Variant icons:**
- `info` — `ℹ`
- `success` — `✓`
- `warning` — `⚠`
- `error` — `✗`

**Example:**
```lua
local Alert = require("lux.alert")

Alert {
    variant = "error",
    title = "Connection Failed",
    message = "Unable to reach the server. Retrying in 5s...",
    dismissible = true,
    onDismiss = function() setShowAlert(false) end,
}
```

---

### Accordion

Collapsible panels with single or multi-expand modes.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `id` | string | `nil` | Element ID |
| `items` | table | `{}` | Array of panel definitions (see below) |
| `mode` | string | `"single"` | `"single"` (one open at a time) or `"multi"` |
| `openItems` | table | `{}` | Array of open panel IDs |
| `selectedIndex` | number | `1` | Focused panel (1-based, for keyboard) |
| `onToggle` | function | `nil` | `function(id, isNowOpen, newOpenItems)` |
| `onSelectedChange` | function | `nil` | `function(newIndex)` — keyboard focus change |
| `width` | number | `nil` | Width |
| `height` | number | `nil` | Height |
| `autoFocus` | boolean | `false` | Auto-focus on mount |

**Item Definition:**
```lua
{
    id = "section1",
    title = "Section Title",
    content = "Plain text content",   -- OR:
    render = function() return ... end, -- render function for rich content
    disabled = false,
}
```

**Keyboard:**
| Key | Action |
|-----|--------|
| `↑` / `k` | Focus previous non-disabled panel |
| `↓` / `j` | Focus next non-disabled panel |
| `Enter` / `Space` | Toggle focused panel |

**Example:**
```lua
local Accordion = require("lux.accordion")

Accordion {
    mode = "single",
    items = {
        { id = "info", title = "Information", content = "Basic details here" },
        { id = "config", title = "Configuration", render = function()
            return ConfigPanel {}
        end },
        { id = "logs", title = "Logs", disabled = true },
    },
    openItems = openPanels,
    selectedIndex = focusedPanel,
    onToggle = function(id, isOpen, newOpen) setOpenPanels(newOpen) end,
    onSelectedChange = function(idx) setFocusedPanel(idx) end,
}
```

---

### Breadcrumb

A horizontal navigation trail with clickable ancestors.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `id` | string | `nil` | Element ID |
| `items` | table | `{}` | Array of `{ id, label }` — last item is current (non-clickable) |
| `separator` | string | `" › "` | Separator between items |
| `onNavigate` | function | `nil` | `function(itemId, index)` — called on ancestor click |
| `width` | number | `nil` | Width |

**Example:**
```lua
local Breadcrumb = require("lux.breadcrumb")

Breadcrumb {
    items = {
        { id = "home", label = "Home" },
        { id = "settings", label = "Settings" },
        { id = "network", label = "Network" },
    },
    separator = " / ",
    onNavigate = function(id, idx)
        lumina.router.navigate("/" .. id)
    end,
}
```

---

### Pagination

Page navigation control with ellipsis breaks and keyboard support.

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `id` | string | `nil` | Element ID |
| `pageCount` | number | `1` | Total number of pages |
| `currentPage` | number | `1` | Currently active page (1-based) |
| `onPageChange` | function | `nil` | `function(page)` — called on page change |
| `pageRangeDisplayed` | number | `3` | Number of pages shown around current |
| `marginPagesDisplayed` | number | `1` | Pages shown at start/end margins |
| `previousLabel` | string | `"‹"` | Previous button label |
| `nextLabel` | string | `"›"` | Next button label |
| `breakLabel` | string | `"…"` | Ellipsis label |
| `width` | number | `nil` | Width |
| `autoFocus` | boolean | `false` | Auto-focus on mount |

**Keyboard:**
| Key | Action |
|-----|--------|
| `←` / `h` | Previous page |
| `→` / `l` | Next page |
| `Home` | First page |
| `End` | Last page |

**Example:**
```lua
local Pagination = require("lux.pagination")

Pagination {
    pageCount = 20,
    currentPage = page,
    pageRangeDisplayed = 5,
    marginPagesDisplayed = 2,
    onPageChange = function(p) setPage(p) end,
}
-- Renders: ‹  1  …  4  5  [6]  7  8  …  20  ›
```

---

### WindowManager (WM)

A state-management module for overlapping window systems. Not a visual component — provides methods to manage window z-order, position, and open/close state.

**Creation:**
```lua
local WM = require("lux.wm")

local mgr = WM.create("wm_state", {
    { id = "editor", title = "Editor", x = 2, y = 1, w = 35, h = 12 },
    { id = "palette", title = "Palette", x = 10, y = 3, w = 30, h = 10 },
})
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `storeKey` | string | Key in `lumina.store` for WM state |
| `initialWindows` | table | Array of `{ id, title, x, y, w, h }` |

**Manager Methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `mgr.register(id, frame)` | `(string, {x,y,w,h,title})` | Add a new window at top of z-order |
| `mgr.close(id)` | `(string)` | Close window (preserves frame for reopen) |
| `mgr.reopen(id)` | `(string)` | Reopen a closed window at top |
| `mgr.activate(id)` | `(string)` | Bring window to front (z-order top) |
| `mgr.setFrame(id, patch)` | `(string, table)` | Update position/size without changing z-order |
| `mgr.getWindows()` | `() → table` | Get open windows ordered bottom-to-top (reactive) |
| `mgr.getActiveId()` | `() → string|nil` | Get active window ID (reactive) |

**Notes:**
- Uses `lumina.store` for reactivity — `getWindows()` and `getActiveId()` call `lumina.useStore` internally
- Idempotent: calling `WM.create` with an existing `storeKey` does not overwrite state
- Call `activate()` only on mousedown/focus events, NOT on every move/resize

---

### Slot (utility)

A factory creator for slot-based composition patterns (used internally by Dialog, Layout).

```lua
local Slot = require("lux.slot")
local Title = Slot("title")

-- Usage:
Title { "Hello" }
-- Returns: { type = "_slot", _slotName = "title", children = {"Hello"} }
```

---

## Theming

All Lux components read theme colors from `lumina.getTheme()`. Default colors follow the **Catppuccin Mocha** palette:

| Token | Default | Usage |
|-------|---------|-------|
| `primary` | `#89B4FA` | Active/focused elements |
| `text` | `#CDD6F4` | Default text |
| `muted` | `#6C7086` | Disabled/dim text |
| `surface0` | `#313244` | Elevated surfaces |
| `surface1` | `#45475A` | Active backgrounds |
| `base` | `#1E1E2E` | Base background |
| `success` | `#A6E3A1` | Success variant |
| `warning` | `#F9E2AF` | Warning variant |
| `error` | `#F38BA8` | Error variant |

---

## Conventions

- **Indexing**: All Lua-facing APIs use **1-based** indexing
- **Callbacks**: Receive the new value directly (e.g., `onChange(value)`, `onPageChange(page)`)
- **Controlled components**: State is owned by the parent; components call callbacks to request changes
- **Keyboard**: Components with `focusable = true` support keyboard navigation
- **Children**: Passed via the array part of the props table or via named slots
