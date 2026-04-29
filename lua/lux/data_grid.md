# Lux DataGrid — 高级表格设计（对标 Web 复杂场景）

> **状态**：已实现为 `lua/lux/data_grid.lua`，`require("lux.data_grid")` 或 `require("lux").DataGrid`；与 `lua/lux/embed.go` / `pkg/lux_modules.go` 注册一致。  
> **命名**：**`DataGrid`**，与 Go `lumina.createElement("Table", …)` 的 **纯文本、固定列宽** widget 明确区分。  
> **相关**：行级富 UI 参考 **`list.md` / `ListView`**；简单矩阵参考 **`pkg/widget/table.go`**。

---

## 1. 动机：为何要 Lux 级「高级表」

| 维度 | Go `Table` widget | Web Data Grid（AG / TanStack / MUI DataGrid） | Lux **DataGrid**（目标） |
|------|-------------------|-----------------------------------------------|---------------------------|
| 单元格 | 仅 `tostring` + 定宽截断 | 任意 React 节点、按钮、徽标、进度条 | **`renderCell(row, rowIndex, column, ctx)`** → 任意 vnode（签名与 `ListView.renderRow(row, i, ctx)` 对齐：**行在前、索引第二、列第三**） |
| 表头 | 纯文本 | 排序图标、筛选 popover、拖拽调宽 | **`renderHeaderCell(col, ctx)`** 可选；默认文本头 |
| 列定义 | `header` / `key` / `width` | `accessor`、meta、`size`/`minSize`/`maxSize` | **列描述表**：`id`、`accessor` 或 `key`、宽度、`align`、可选 `sortable` |
| 选择 | 整行（Go 侧 `onChange` 为 **0-based** 行索引） | 单选 / 多选 / checkbox 列 / Shift 范围 | **P0**：与 ListView 一致的 **1-based** `selectedIndex` / `onChangeIndex`；**P1**：`selectedIds` / 多选 / `getRowId` |
| 排序 / 筛选 | 无 | 服务端或客户端 | **受控模式**：父组件或 store 持有 `sort` / `filters`，表只负责 UI 与回调 |
| 虚拟滚动 | 无（`ScrollOffset` 未接线） | 大行数必备 | **P2+**：视口裁剪或引擎能力扩展前，先 **行数上限 + 文档警告** |
| 热更 / 定制 | 改 Go | 改 JS | **Lua-only**，与 Card / ListView 同一分发模型 |

**一句话**：Go `Table` 适合「日志型、只读、纯文本」；Lux **DataGrid** 适合「运维控制台、可编辑单元格、复杂表头」等 **Web Data Grid 子集**，在终端约束下渐进落地。

---

## 2. 与现有栈的边界

### 2.1 非目标（v1 / 设计期明确不写进首版）

- **真正 Excel**：合并单元格、公式栏、百万行 — 非目标。  
- **列拖拽重排、列宽拖拽**：依赖精确指针与 hit-box；**P3** 再评估（可先键盘调宽或配置写死）。  
- **表内嵌套完整「子表格」展开行**：**P2**；首版可用「展开行渲染多行 `text`/`vbox`」替代。  
- **与表头并发的全局快捷键**：仍服从 App `keys`；表内 `onKeyDown` 消费规则见 §6。

### 2.2 必须遵守的引擎事实（设计约束）

- 布局：`hbox` / `vbox`、`overflow = "scroll"`、`scrollY` 与 **`ListView` 同源**（滚轮由引擎处理）。  
- **不要用空字符串当大面积背景**（见 `examples/list_dialog.lua` 教训）：未选中 / 默认格用 **不透明 `t.base` / `t.surface0`**，避免脏区重绘后透出终端底色。  
- **Go widget 与 Lux 混排**：`createElement("Table", …)` 已是 **component 占位 + 嫁接**；Lux DataGrid **不依赖**扩展 Go `Table`， entirely `defineComponent` + 基元。

---

## 3. Web 能力对标矩阵（分阶段）

| Web 常见能力 | 说明 | Lux 阶段 |
|--------------|------|----------|
| 多列 + **表头固定在视口顶部、仅表体滚动** | 双容器：表头在 `scroll` 外（**非** Web `position: sticky` 也能做到） | **P0** |
| 多列 + 表体分隔视觉 | 与 shadcn Table 类似 | **P0** |
| 固定行高或「行高由行根节点高度推导」 | 与 ListView 一致 | **P0** |
| 列宽：`width` / `minWidth` / `flex` | 终端以 cell 为单位；`flex` 映射为比例或剩余空间算法（文档化） | **P0–P1** |
| 行选择：单行（`selectedIndex`） | 与 ListView 一致 | **P0** |
| 行选择：多选（Ctrl/Space）、全选列 | Web 标配 | **P1** |
| 排序指示：列上点击 → `onSortChange(meta)` | 数据由父层排序 | **P1** |
| 筛选：列头筛选 / 全局搜索框 | 回调 + 外部过滤 `rows` | **P1** |
| 分页条 | 由父级 `Pagination` Go widget 或纯 Lua 拼 | **P1**（组合式） |
| **横向**粘性列、表头内二级行冻结等 | 接近 Web `sticky` 的更强语义 | **P2**（引擎或布局增强，待定） |
| 虚拟列表 | 仅渲染视口行 | **P2**（需稳定行高 + `scrollY` 与数据窗口对齐；**`getRowId` 见 P1**） |
| 列显示/隐藏、列顺序配置 | 配置面板 | **P3** |

---

## 4. 组件 API 草案（P0 / P1）

以下 **非最终实现**，供评审与对齐 `list.lua` 风格。

### 4.1 列描述 `columns[i]`

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `string` | 稳定列 id，用于排序状态与 reconcile `key` |
| `accessor` | `function(row) -> any` **或** `key` 字符串 | 取值；字符串时等价 `row[key]` |
| `header` | `string` **或** 省略则用 `id` | 无 `renderHeaderCell` 时的默认标题 |
| `width` | `number` | 列宽（cell 数），默认与引擎约定一致 |
| `minWidth` / `maxWidth` | `number` | **P1** 弹性布局 |
| `align` | `"start"` \| `"end"` \| `"center"` | 文本对齐（映射到 pad 或子 `text` 样式） |
| `sortable` | `bool` | **P1**：表头可点击，触发 `onSortChange` |

### 4.2 根 Props（节选）

| Prop | 说明 |
|------|------|
| `rows` | 行数据数组（只读引用；排序筛选由父组件换新表或同表） |
| `columns` | 上表 |
| **`selectedIndex`** | **P0**：当前选中行，**1-based**，与 `ListView` 一致；受控 |
| **`onChangeIndex`** | **P0**：`function(index)`，选中行变化时回调 **新**下标（1-based） |
| **`onActivate`** | **P0**：`function(index, row)`，`Enter` 时调用（与 `ListView.onActivate` 对齐） |
| `height` / `width` | **整体**视口尺寸：根 `vbox` **不**设 `overflow=scroll`；**表体子树**占剩余高度并单独 `overflow=scroll` + `scrollY`（见 §5） |
| `rowHeight` | 默认 `1`；多行单元格时与 ListView 相同约定 |
| `renderCell` | **P0**：**`function(row, rowIndex, column, ctx)`** → vnode；`ctx` 含 `selected` 等；**P1** 扩展 `focusedColumnId`、编辑态等 |
| `renderHeaderCell` | **`function(column, ctx)`** 可选；默认 `text` 头（**P1** 可与 `sortable` 组合） |
| `getRowId` | **P1**：**`function(row, index) -> string`**；多选与虚拟化行的稳定 key。**P0** 对行 vnode 使用 **`tostring(rowIndex)`**（或列级 `key`）即可 |
| `selectedIds` / `onSelectionChange` | **P1** 多选受控（Lua table 表示集合） |
| `sort` / `onSortChange` | **P1** `{ columnId, direction }` |
| `empty` | 无行文案 |
| `autoFocus` | 根 `focusable` + 首帧聚焦（与 Go `Table` 的 `autoFocus` 语义对齐，Lux 自管） |

### 4.3 索引约定

与 **`ListView` 一致**：对 Lua 作者暴露的 **行索引 1-based**（`ipairs`）。与 Go `Table` 的 `onChange`（**0-based** 行索引）互操作时，在边界 **单次转换** 并文档化。

---

## 5. 结构草案（实现时）

**表头必须在滚动容器之外**，否则表头会随内容滚出视口；这是 **P0** 基线体验（双 `vbox` 即可），**不等价于** §3 中 **P2** 的「横向粘性列 / 表头内复杂冻结」等进阶能力。

```
vbox (root: focusable, autoFocus?; 占满 height/width)
├── hbox (header row) — 固定在视口顶部，不滚动
│   └── [ renderHeaderCell(col) 或默认 text ]
├── text (separator ───…) — 可选
└── vbox (body: overflow=scroll, scrollY; flex 占剩余高度)
    └── vbox
        └── 每行：hbox of renderCell(row, rowIndex, col, ctx) …
```

- **横向**：每行一个 `hbox`，子列为 `renderCell` 根（宽度由列定义约束）。  
- **纵向**：**仅 body** 使用 `scrollY` 与 `ListView` 相同公式族（选中行居中、仅滚轮跟随等 **`scrollMode`** 放 **P1** 亦可，P0 可先「滚轮 + 简单跟选中」）。

---

## 6. 键盘与焦点

### 6.1 P0 键盘行为（首版只做这些）

| 按键 | 行为 |
|------|------|
| `↑` `↓` / `k` `j` | 移动 **当前选中行**（与 `onChangeIndex` / `selectedIndex` 联动） |
| `Enter` | 调用 **`onActivate(index, row)`** |

不与 **列内 `input` 编辑态** 混用规则写在 **P1**（见下节）。

### 6.2 P1 及以后

| 按键 / 模式 | 行为（草案） |
|-------------|----------------|
| `↑` `↓` / `k` `j` | 若存在 **编辑态单元格**，优先由该单元格消费 |
| `←` `→` / `h` `l` | 列焦点（单元格导航）或与横向滚动 **二选一**，须文档化 |
| `Space` | 多选切换（与 `selectedIds` 组合） |
| `Enter` | 在非编辑态：行激活；编辑态：交给 `input` |
| `Tab` | 引擎焦点环；表根与单元格 `focusable` **互斥规则**须文档化 |

**原则**：默认 **「表级焦点」** 优先（与 ListView 一致）；仅当显式进入「单元格编辑」才把焦点交给 `input`（参考 `list.md` 行内勿用 `focusable` 的警告；DataGrid 用 **模式位** 收紧）。

---

## 7. 性能与数据规模

- **P0**：`#rows * rowHeight` 全量挂载；建议文档写 **软上限**（如 ≤500 行或按列宽乘行数估算 cell 数）。  
- **P2**：**窗口化**：只 `renderCell` 视口内 ±buffer 行，`scrollY` 变化时重算 `windowStart`；依赖 **P1** 的 `getRowId` 与稳定 `key`。  
- 与 **DevTools / perf**：大表调试时关注 `PaintCells` 与 `renderComponent` 次数（见 `docs/DESIGN-perf.md`）。

---

## 8. 测试与示例（实现后）

| 类型 | 内容 |
|------|------|
| E2E | `examples/data_grid_*.lua`：P0 表头固定 + 滚动体、单行选中、`Enter`；P1+ 排序、多选、长文本、主题 |
| 单元 | 若逻辑抽纯函数，可 Go 测；否则 Lua `test.describe`（与现有 `pkg/testdata/lua_tests` 对齐） |

---

## 9. 路线图摘要

| 阶段 | 交付物 |
|------|--------|
| **P0** | `DataGrid`：`columns` + `rows` + **`renderCell(row, rowIndex, column, ctx)`** + **表头固定 + 仅表体 scroll** + **`selectedIndex` / `onChangeIndex` / `onActivate`** + 空态 + 主题安全背景 |
| **P1** | `getRowId`、多选 `selectedIds` / `onSelectionChange`、排序/筛选回调、`renderHeaderCell`（含 sortable）、`scrollMode`、分页组合示例 |
| **P2** | 横向粘性列 / 更强冻结语义、或虚拟窗口、展开行 |
| **P3** | 列拖拽、列菜单、列宽指针交互 |

---

## 10. 参考与修订

- **ListView**：`lua/lux/list.md`、`lua/lux/list.lua`  
- **Go Table**：`pkg/widget/table.go`（简单场景继续用）  
- **架构总览**：`docs/DESIGN-widgets.md`  
- **修订记录**：在 PR / commit 中更新本文件；**2026-04** 评审纳入：`renderCell` 参数顺序、P0 固定表头双容器、P0 单行选中 API、`getRowId` 推迟 P1、§6 拆分 P0/P1 键盘。
