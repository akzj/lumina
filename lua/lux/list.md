# Lux ListView 组件 — 设计草案（Review）

> **状态**：仅设计文档，尚未实现。实现后路径拟为 `lua/lux/list.lua`，并在 `init.lua` 中导出。  
> **命名**：建议使用 **`ListView`**（或 `ScrollList`），**避免**与 Go 内置 `lumina.List`（`pkg/widget/list.go`）同名混用。

---

## 1. 动机与边界

### 1.1 为什么放在 Lux，而不是扩展 Go `List`

| 能力 | Go `lumina.List` | Lux `ListView`（拟） |
|------|------------------|----------------------|
| 数据源 | `items` 仅字符串 / `{ label }` | **任意行 vnode**（`createElement`、子组件） |
| 行高 | 单行文本 | **可变行高**（多行预览、图标+hbox） |
| 滚动 | 无（需外层自己包） | **内置** `ScrollView` 或 `overflow=scroll` + `scrollY` |
| 搜索 / 过滤 | 无 | **可选**（首版可简化为无，或 props 注入 `filter`） |
| 迭代速度 | 需改 Go + 引擎契约 | **Lua `defineComponent`**，与 `CommandPalette` 同层 |

结论：与仓库既有分层一致 — **引擎 Widget 保持窄能力；Lux 承载「shadcn 式」可组合厚组件**。

### 1.2 与现有 Lux 组件的关系

- **`CommandPalette`**：固定「搜索框 + 命令行 + Enter 执行」，**不**替代通用列表。
- **`examples/notes.lua`**：手写 `vbox` + scroll + 选中态 — **ListView 应吸收该模式**，减少复制粘贴。

### 1.3 非目标（首版可不承诺）

- 虚拟化（万行级）— 若性能不足再评估 **第二期** 或文档推荐「大数据用简化行 + 分页」。
- 行内可聚焦控件（每行一个 `input`）与列表键盘导航 **同时** 完美共存 — 首版可约定 **行内无 focusable**，或文档写明限制。
- 拖拽排序、多列表头 — 超出「List」范畴，归 **Table / 专用组件**。

---

## 2. 组件职责（首版 MVP）

1. **渲染**：在固定 **高度** 内展示 **子行**（由调用方传入的 vnode 数组或 render prop）。
2. **选中**：**单选**、`selectedIndex`（1-based 与 Lua 习惯一致，**或** 0-based 在文档与 API 中二选一并全文统一）。
3. **键盘**：`↑` / `↓` / `j` / `k` 移动选中；`Enter` 触发 `onActivate(index, row)`（若提供）。
4. **滚动**：选中行变化时 **滚入视口**（与 notes 类似：`scrollY` 按累计行高或按「每行固定高度」简化）。
5. **主题**：通过 `lumina.getTheme()` 设置选中行背景/前景（与 `CommandPalette` 一致风格）。

---

## 3. API 草案

### 3.1 模块与引用

```lua
local ListView = require("lux.list")
-- 或（实现并挂 init 后）
local lux = require("lux")
lux.ListView { ... }
```

### 3.2 Props（拟）

| Prop | 类型 | 默认 | 说明 |
|------|------|------|------|
| `rows` | `table` | `{}` | **数据源**（任意 Lua 表）；与 `renderRow` 配合 |
| `renderRow` | `function(row, index, ctx) -> vnode` | **必填**（若不用 `children` 模式） | `ctx.selected`、`ctx.dim` 等 |
| `selectedIndex` | `number` | `1` | 当前选中下标（**全文需统一 1-based 或 0-based**） |
| `onChangeIndex` | `function(index)` | `nil` | 键盘/点击导致选中变化时回调 |
| `onActivate` | `function(index, row)` | `nil` | `Enter` 或未来「行点击」时 |
| `height` | `number` | `10` | 列表视口高度（行数/cells） |
| `width` | `number` / `nil` | `nil` | 可选，传给外层 `ScrollView` / `box` |
| `empty` | `vnode` / `string` | `"No items"` | 无数据时展示 |
| `key` / `id` | `string` | `nil` | 透传 reconciler |

**备选 API（二选一，Review 时定）**

- **A. `rows` + `renderRow`**：数据与视图分离，易测试、易做过滤扩展。  
- **B. 仅 `children` 数组**：与 `Card` 一致，但选中索引与 **子节点顺序** 强绑定，过滤时需调用方重建 children。

**建议首版采用 A**；B 可作为薄封装或文档示例。

### 3.3 与 `lumina.List` 的对照（文档层）

在 `API.md` / 本文件末尾增加「**何时用 `lumina.List`，何时用 `lux.ListView`**」一段即可，无需改名引擎组件。

---

## 4. 实现要点（实现阶段参考）

1. **根节点**：`lumina.ScrollView`（固定 `style.height`）+ 内层 `vbox` 包裹各行，便于键盘滚动与引擎滚动条一致。  
2. **行包装**：每行外包一层 `box`/`hbox`（可选），用于 **整行高亮背景**、未来 `onClick` 命中。  
3. **`scrollY`**：受控或内联 `useState`；根据 `selectedIndex` 与 **行高策略** 更新：  
   - **简化策略（推荐 MVP）**：约定 `renderRow` 返回的根节点带 `style.height = n`（n≥1），或 ListView 接受 `rowHeight = 1` 常量近似 notes。  
4. **焦点**：根 `focusable = true`，`onKeyDown` 处理方向键；**不**把事件冒泡到行内（首版无行内 input）。  
5. **文件**：`lua/lux/list.lua`；`init.lua` 增加 `M.ListView = require("lux.list")`。

---

## 5. 使用示例（目标形态）

### 5.1 简单富行（标题 + 副标题）

```lua
local ListView = require("lux.list")
local t = lumina.getTheme()

local notes = {
    { id = 1, title = "A", subtitle = "line a" },
    { id = 2, title = "B", subtitle = "line b" },
}

local function renderRow(row, i, ctx)
    local sel = ctx.selected
    return lumina.createElement("vbox", {
        style = {
            height = 2,
            background = sel and (t.surface1 or "#45475A") or "",
        },
    },
        lumina.createElement("text", { bold = sel, foreground = t.text }, (sel and "▸ " or "  ") .. row.title),
        lumina.createElement("text", { foreground = t.muted, dim = true }, "    " .. row.subtitle)
    )
end

ListView {
    rows = notes,
    renderRow = renderRow,
    height = 12,
    onChangeIndex = function(i) /* sync store */ end,
    onActivate = function(i, row) /* open detail */ end,
}
```

### 5.2 与 `useStore` 组合（受控选中）

由父组件 `useStore("selectedIdx")`，在 `onChangeIndex` 里 `store.set`，`selectedIndex` 从 store 传入 — 与 `notes.lua` 一致，仅把滚动+键盘收进 ListView。

---

## 6. 交付阶段（建议）

| 阶段 | 内容 |
|------|------|
| **P0** | `rows` + `renderRow`、单选、`↑↓jk`、固定 `rowHeight` 或行内声明 `height`、ScrollView、`onChangeIndex` / `onActivate` |
| **P1** | 行点击选中（`mousedown` 相对坐标算行号，需行高一致或每行记录 Y） |
| **P2** | 可选 `filter` / 搜索框 slot；或单独 `FilterableListView` 避免 API 膨胀 |
| **P3** | 虚拟化（若确有长列表场景） |

---

## 7. Review 清单（请拍板）

1. **组件名**：`ListView` vs `ScrollList` vs 其他？  
2. **下标**：`selectedIndex` **1-based**（Lua 习惯）还是 **0-based**（与 Go `List` onChange 对齐）？  
3. **API 形态**：是否确认 **`rows` + `renderRow`** 为 P0 唯一路径？  
4. **行高**：MVP 是否强制 **`renderRow` 根节点 `style.height`**，文档写明？  
5. **`init.lua` 导出键名**：`ListView` 是否可接受（避免 `lux.List` 与 `lumina.List` 混淆）？

---

## 8. 参考代码路径

- `lua/lux/command_palette.lua` — 选中行 + 键盘 + 主题  
- `examples/notes.lua` — `scrollY` 与多行 `NotePreview`  
- `pkg/widget/scrollview.go` — 滚动与 PageUp/PageDown 行为  
- `pkg/widget/list.go` — 勿混用：纯文本列表仍用引擎 `List`
