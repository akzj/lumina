# Index Convention — Lumina Lua API

## Rule: All indexes exposed to Lua are 1-based

Lua arrays start at 1. All Lumina APIs follow this convention:
- `onChange` callbacks from Go widgets pass **1-based** row/item indexes
- `selectedIndex` props are **1-based** (first item = 1)
- `onChangeIndex`, `onActivate` callbacks pass **1-based** indexes
- `rows[index]` works directly without conversion

## Go Widget Internal State

Go widgets may use 0-based indexing internally (Go convention).
The conversion happens at the **FireOnChange boundary**:

```go
// Internal: 0-based
s.SelectedIndex++
// Exposed to Lua: 1-based
event.FireOnChange = s.SelectedIndex + 1
```

## "No selection" state

- Go internal: `-1` means no selection
- Lua: `nil` or `0` means no selection (0 is not a valid array index in Lua)

## Affected Widgets

| Widget | onChange payload | Notes |
|--------|----------------|-------|
| Table | row index (1-based) | First row = 1 |
| List | item index (1-based) | First item = 1 |
| Menu | item index (1-based) | First item = 1 |
| Dropdown | option index (1-based) | First option = 1 |
| Pagination | page number (1-based) | First page = 1 |
| Checkbox | boolean | N/A |
| Switch | boolean | N/A |
| Radio | value string | N/A |
| Select | value string | N/A |

## Lux Components (Lua-only)

| Component | API | Convention |
|-----------|-----|-----------|
| ListView | `selectedIndex`, `onChangeIndex(i)` | 1-based |
| DataGrid | `selectedIndex`, `onChangeIndex(i)` | 1-based |
| Pagination | `currentPage`, `onPageChange(p)` | 1-based |
