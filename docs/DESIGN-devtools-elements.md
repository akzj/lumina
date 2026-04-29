# DevTools Elements 向体验（设计备忘）

> **目标**：接近 Chrome **Elements** 的日常价值：**完整真相树 + 命中选中 + 单节点详情**，在 TUI 约束下取舍。  
> **现状**：`F12` → Elements 为拍平快照，见 `pkg/app_devtools_v2.go`（`updateDevToolsElements` / `paintElementsTab`）。

---

## MVP（建议第一版）

1. **树**：提高或取消过严的 `maxDepth` / `maxNodes`，或改为可配置；支持 **折叠状态**（存在 `devtools.Panel`）。
2. **选中**：**Inspect 模式**（如 Shift+点击或快捷键）→ `HitTest` → 记录 `selectedId`，树 **滚动到该项** 并展开祖先路径。
3. **详情**：选中节点展示 **几何（x,y,w,h）+ 主要 `Style` 字段 + `type` / `id` / `content` / `Component`**（第二列或下半屏）。

---

## 后续阶段（可选）

| 阶段 | 内容 |
|------|------|
| P1 | 按 `type` / `id` / `content` **搜索过滤**；**Copy 路径**（如 `/root/box[0]/…`）。 |
| P2 | 命中与面板区域 **互斥**（避免面板挡住 hit-test）；buffer 上 **简化高亮框**。 |
| P3 | 与 **Perf** 联动（选中节点本帧 dirty 等，需额外状态）。 |

---

## 约束（相对 Chrome）

- 以 **`Node` 树** 为真相；Lua 描述符为源，不等价于 DOM。
- **Computed 样式** = 当前 `Style` + 布局结果，无完整 CSS 级联。
- **实时改样式**：仅建议调试后门，且不反写 Lua，避免语义分叉。

---

## 代码入口

- `pkg/devtools/devtools.go` — `Panel`、`NodeInfo`（若扩展字段在此或邻接类型）。
- `pkg/app_devtools_v2.go` — 树 walk、`paintElementsTab`、`HandleEvent` 与 Inspect 接线点。
