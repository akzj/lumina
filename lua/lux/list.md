# Lux ListView — 设计与使用

> **实现**：`lua/lux/list.lua`，`require("lux.list")` 或 `require("lux").ListView`。运行时由 `lua/lux/embed.go` + `pkg/lux_modules.go` 的 `registerLuxModules` 加载本文件。  
> **命名**：**`ListView`**，与 Go `lumina.List`（纯文本 widget）区分。

---

## 1. 动机与边界

### 1.1 为何在 Lux 而非扩展 Go `List`

| 能力 | Go `lumina.List` | Lux `ListView` |
|------|------------------|----------------|
| 数据源 | `items` 仅字符串 / `{ label }` | **`rows` + `renderRow`** → 任意行 vnode |
| 行高 | 单行 | **固定 `rowHeight`（MVP）**，与 `renderRow` 根节点高度一致 |
| 滚动 | 无 | **`vbox` + `overflow = "scroll"` + `scrollY`**（与 `examples/notes.lua` 一致） |
| 鼠标滚轮 | — | **由引擎**对 `overflow=scroll` 区域处理，ListView 不额外写滚轮逻辑 |
| 迭代 | 改 Go / 引擎契约 | **`defineComponent`**，与 `CommandPalette`、`Card` 同层 |

### 1.2 非目标（当前版本）

- 虚拟化、拖拽排序、表头式多列 Table（多列高级表见 **`data_grid.md`** 设计稿）  
- 行内 **可聚焦控件**（如每行一个 `input`）与列表键盘导航并存 — 首版约定行内勿用 `focusable`，否则焦点与 `↑↓` 冲突  

---

## 2. 已定稿约定（Review 结论）

| 项 | 决定 |
|----|------|
| 组件名 | **`ListView`** |
| `selectedIndex` / `onChangeIndex` / `onActivate` 下标 | **1-based**，与 Lua `ipairs` / `#rows` 一致；Go `List` 的 0-based 不映射到本组件 |
| P0 API | 仅 **`rows` + `renderRow`** |
| 滚动实现 | **不用** `lumina.ScrollView` Go widget；用 **`vbox` + `overflow="scroll"` + `scrollY`** |
| 行高 | **`rowHeight` prop**（默认 `1`）；`renderRow` 返回的根节点 **`style.height` 必须等于 `rowHeight`**（文档约束，运行时未逐项校验） |
| `init.lua` / `lux` 包导出 | **`M.ListView = require("lux.list")`** |

---

## 3. Props

| Prop | 类型 | 默认 | 说明 |
|------|------|------|------|
| `rows` | `table` | `{}` | 数据行 |
| `renderRow` | `function(row, index, ctx)` | 无（`#rows>0` 时必填） | `index` **1-based**；`ctx.selected` 为布尔 |
| `selectedIndex` | `number` | `1` | 当前选中；受控时由父组件传入 |
| `onChangeIndex` | `function(index)` | `nil` | `↑`/`↓`/`k`/`j`（及 `Up`/`Down`）时调用 **新**下标 |
| `onActivate` | `function(index, row)` | `nil` | `Enter` 时调用 |
| `rowHeight` | `number` | `1` | 逻辑行高（用于 `scrollY` 与内容高度） |
| `height` | `number` | `10` | 视口高度（cell 行数） |
| `width` | `number` | `nil` | 可选，写入根 `vbox.style.width` |
| `empty` | `string` | `"No items"` | 无行时的提示文案 |
| `id` / `key` | `string` | `nil` | 透传 |

---

## 4. 滚动与滚轮

- **`scrollY`**：`max(0, min(maxScroll, ideal))`，其中 `contentLines = #rows * rowHeight`，`maxScroll = max(0, contentLines - height)`，`ideal = (selectedIndex - 1) * rowHeight - floor(height / 2)`，用于大致把选中行放在视口中部附近。  
- **鼠标滚轮**：根节点使用引擎 **`overflow = "scroll"`** 时，由现有引擎路径处理视口滚动（与 `notes.lua` 相同）；ListView 不重复实现。

---

## 5. 使用示例

```lua
local ListView = require("lux.list")
local t = lumina.getTheme()

local rows = {
	{ id = 1, title = "Alpha", sub = "a" },
	{ id = 2, title = "Beta",  sub = "b" },
}

local function renderRow(row, i, ctx)
	local sel = ctx.selected
	return lumina.createElement("vbox", {
		style = {
			height = 2, -- 必须等于 ListView 的 rowHeight
			background = sel and (t.surface1 or "#45475A") or "",
		},
	},
		lumina.createElement("text", { bold = sel, foreground = t.text },
			(sel and "▸ " or "  ") .. row.title),
		lumina.createElement("text", { foreground = t.muted, dim = true },
			"    " .. row.sub)
	)
end

-- 受控：selectedIndex / onChangeIndex 由 store 或父 state 驱动
ListView {
	rows = rows,
	rowHeight = 2,
	height = 8,
	selectedIndex = selectedIdx,
	renderRow = renderRow,
	onChangeIndex = function(i) setSelectedIdx(i) end,
	onActivate = function(i, row) /* open detail */ end,
}
```

---

## 6. 与 `lumina.List` 的选择

- **`lumina.List`**：纯文本 `items`，引擎内键盘选中，适合简单菜单。  
- **`lux.ListView`**：富行、固定行高滚动、Lua 侧受控选中，适合 `notes` 类界面。

---

## 7. 交付阶段（路线图）

| 阶段 | 内容 |
|------|------|
| **P0** | 当前实现：`rows` + `renderRow`、`rowHeight`、`↑↓`/`Up`/`Down`/`j`/`k`、`Enter`、`scrollY` |
| **P1** | 行点击选中（需稳定行高或命中盒） |
| **P2** | 搜索 / `filter` 或独立 `FilterableListView` |
| **P3** | 虚拟化（大数据） |

---

## 8. 参考

- `lua/lux/command_palette.lua` — 选中行样式参考  
- `examples/notes.lua` — `overflow` + `scrollY` 模式  
- `pkg/widget/list.go` — Go 文本 `List`，勿混用  

---

## 9. 实现说明（引擎）

子组件 `ComponentProps` 中的 **函数**（如 `renderRow`、`onChangeIndex`）在 `readDescriptor` → `readMapFromTable` 中存为 **`propFuncRef`（registry ref）**；`renderComponent` 里 `pushMap` 通过 **`RawGetI(RegistryIndex, ref)`** 压回 Lua，这样 `type(renderRow) == "function"`。普通 `int64` 仍按数值推入，避免与 ref 混淆。

Go `lumina.List` 的 `FireOnChange` 对 Lua 为 **1-based** 下标；Lux `ListView` 的 `selectedIndex` / 回调参数为 **1-based**，二者在 Lua 侧一致，但与 `ListState.SelectedIndex`（0-based）不同，文档区分即可。
