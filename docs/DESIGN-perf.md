# Perf 工具设计分析

> **状态**：设计评审（与实现可不同步；本文档以仓库代码为准，**最后对齐**：`pkg/perf`、`pkg/app.go`、`pkg/render/engine.go`、`pkg/app_devtools_v2.go`、`pkg/devtools/devtools.go`）。  
> **代码入口**：`pkg/perf/`（`Tracker`、`FrameStats`、`Report` / `Assert`）、`pkg/app.go`（`BeginFrame`/`EndFrame`、输出计数）、`pkg/render/engine.go`（V2 计数）、`pkg/devtools/` + `pkg/app_devtools_v2.go`（F12 面板 Perf 页）。

---

## 1. 目标角色（完整工具应覆盖）

| 角色 | 需求 | 当前覆盖度 |
|------|------|------------|
| **框架开发者** | 布局/协调/绘制/事件各阶段成本、回归防退化 | 部分：V2 有「组件渲染次数 + 画布写/清 + 脏区面积」；**无** layout/reconcile 节点级计数、**无**分阶段耗时 |
| **应用 / Lua 作者** | 哪次 `render` 重、store 更新频率、Lua 耗时 | 弱：`Panel` 上有 `LuaCPUTime` / `LuaMemBytes` 字段与 `UpdateLuaMetrics` API，**仓库内无调用方**，Perf 表也未展示 |
| **终端 / 宿主** | 写出带宽、flush 次数、脏区面积 | 部分：`WriteFull`/`WriteDirty`/`Flush`、`DirtyRectsOut`；**无**字节带宽 |
| **CI / 自动化** | 可断言的阈值、稳定报告 | 有 `perf.AssertLastFrame`、`CheckV2*` 等，依赖 `Tracker.Enable()`（F12 打开 DevTools 时会 `Enable`） |

---

## 2. 现状架构（数据流）

```
App.RenderAll / RenderDirty
    tracker.BeginFrame()              // 清零 current，合并上一帧 End 之后产生的 pending
    engine.RenderAll / RenderDirty    // 见 §2.1；可能 early-return 不写 V2 计数
    [RenderDirty:] adapter.Write* / Flush → Record(Write*, DirtyRects*, Flush)
    tracker.EndFrame()                // Duration = 整段 wall time；写入环形历史 + Total + Alert

DevTools：F12 打开时 tracker.Enable()；SnapshotPerf() 复制 LastFrame/TotalStats + FPS → paintPerfTab 只读快照
tickDevToolsV2：TickFPS()；可见时每 ~300ms SnapshotPerf + 重画面板（与 FPS EMA 窗口独立）
```

### 2.1 `RenderDirty` 与 perf 的边界情况

- **`engine.needsRender == false`**：`RenderDirty` 在 **`ResetStats` 之后**立刻 `return`，**本帧不产生任何 `V2*` `Record`**（`V2ComponentsRendered` 等保持上一帧 `BeginFrame` 合并进来的 pending 计数规则下的值；通常与「无渲染」帧组合时需注意断言时机）。
- **有脏但 `rendered==0` 且无 `hasAnyDirty`**：写 **`V2PaintCells/V2PaintClearCells/V2DirtyRectArea = 0`** 后 `return`（不跑 layout/paint）。
- **`V2DirtyRectArea`**：来自 **`CellBuffer.Stats()`** 的 **`DirtyW * DirtyH`**（本帧画布的脏矩形包围盒面积），**不是** `app` 传给 `WriteDirty` 的 adapter 矩形面积；二者可相关但不等价。

### 2.2 帧语义

- **帧边界**：与 **`App.RenderAll` / `App.RenderDirty` 单次调用** 对齐，不是 OS vsync。
- **`FrameStats.Duration`**：`EndFrame` 时 `time.Since(BeginFrame 的 StartTime)`，包含 **engine + ToBuffer + adapter + Flush**（以及 `RenderDirty` 里 DevTools 合成进脏矩形的额外成本）。
- **历史**：`NewTracker(0)` 默认 **60** 帧环形缓冲；`LastFrame()` / `History()` / `TotalStats()`。

---

## 3. 指标清单：定义 vs 实际写入

### 3.1 全量 `Metric` 枚举（`pkg/perf/tracker.go`）

| Metric | 设计含义 | 生产路径是否写入 |
|--------|----------|------------------|
| `Renders` | `RecordComponent` 每次 +1 | **否**（仅测试） |
| `Layouts` / `Paints` | 旧管线 | **否** |
| `OcclusionBuilds` / `OcclusionUpdates` | 旧合成器 | **否** |
| `ComposeFull` / `ComposeDirty` / `ComposeRects` | 旧合成 | **否** |
| `DirtyRectsOut` | 本帧 `WriteDirty` 传入的矩形 **个数** | **是**：`app.RenderDirty`，每帧最多 +1 |
| `HitTesterRebuilds` / `HandlerFullSyncs` / `HandlerDirtySyncs` | 旧管线 | **否** |
| `EventsDispatched` / `EventsMissed` | 事件命中 | **否**（`RecordEvent` 无调用方） |
| `ComponentsRegistered` / `ComponentsUnregistered` / `MovesPositionOnly` / `MovesWithResize` / `StateSets` | 组件/状态 | **否**（V2 主路径未 `Record`） |
| `WriteDirtyCalls` / `WriteFullCalls` / `FlushCalls` | 适配器写出 | **是**：`app.RenderAll` / `RenderDirty` |
| **`V2ComponentsRendered`** | 本帧 `renderInOrder` 实际调用 `renderComponent` 的次数 | **是**：`RenderAll` / `RenderDirty` |
| **`V2LayoutNodes`** | 参与布局的结点数 | **否**，恒 0 |
| **`V2PaintNodes`** | 参与绘制的结点数 | **否**，恒 0 |
| **`V2PaintCells`** | `CellBuffer` 单元格写入次数（`SetChar`/`Set` 等累计） | **是**：paint 后 `buffer.Stats().WriteCount` |
| **`V2PaintClearCells`** | `ClearRect` / `Clear` 清除的 cell 数 | **是**：`buffer.Stats().ClearCount` |
| **`V2DirtyRectArea`** | 上述 Stats 的 **`DirtyW * DirtyH`**（无写入时为 0） | **是** |
| **`V2ReconcileChanges`** | reconcile 变更结点数 | **否**，恒 0；**`Report()`/`TotalReport()` 仍打印该字段** → 易误解 |

### 3.2 `RecordComponent` / `RecordEvent`

- **`RecordComponent(id)`**：递增 `Renders`，并把 `id` 追加到 `FrameStats.RenderComponents`（用于「按组件 ID 列渲染」类断言）。
- **`RecordEvent(type, dispatched)`**：维护 `EventsByType` 与 `EventsDispatched`/`EventsMissed`。
- **现状**：二者**仅在 `pkg/perf/*_test.go` 等测试中出现**；V2 应用路径**未调用**。

### 3.3 `Report()` / `TotalReport()`（`pkg/perf/report.go`）

- **始终打印**一大段旧管线计数（Renders、Occlusion、Compose、HitTester、Handler、Events、Components…）。
- **仅当** `V2ComponentsRendered > 0 || V2PaintCells > 0` 时追加一行 **`V2 Engine: ... reconcileChanges=...`**；其中 **`reconcileChanges` 当前恒为 0**。

### 3.4 DevTools Perf 页（`paintPerfTab`，`pkg/app_devtools_v2.go`）

**实际渲染的行（label → 数据来源）**：

| 展示项 | 数据来源 |
|--------|----------|
| FPS | `Panel.TickFPS()` 的 EMA（约每 **300ms** 用 `fpsFrameCount` 更新一次） |
| Frame Duration | `Snapshot.Last.Duration` |
| Max Frame | `Snapshot.Total.MaxFrameDuration` |
| Total Frames | `Snapshot.Total.Frames` |
| Renders / Layouts / Paints | `last/total.Get(perf.Renders|Layouts|Paints)` → 真实应用多为 **0** |
| V2 Rendered | `last.Get(V2ComponentsRendered)` |
| V2 Paint Cells | `last.Get(V2PaintCells)` |
| DirtyRects | `last.Get(DirtyRectsOut)` |
| Events Hit / Missed | `last.Get(EventsDispatched/EventsMissed)` → 多为 **0** |
| Runtime 段 | `runtime.MemStats`、`runtime.NumGoroutine()` |

**当前未在 Perf 表中展示、但已有数据或 API**：

- `V2PaintClearCells`、`V2DirtyRectArea`、`WriteFullCalls`/`WriteDirtyCalls`/`FlushCalls`
- `Panel.LuaCPUTime` / `LuaMemBytes`（**`UpdateLuaMetrics` 无仓库内调用方**，数值通常不更新）

---

## 4. 主要缺口（「不够完整」指什么）

1. **指标与实现两套皮**：枚举 + Report 很全，**生产路径只填一小子集**；`V2ReconcileChanges` 等应 **接线**或从 Report/DevTools **隐藏并标注未实现**。
2. **缺分阶段耗时**：仅有 `FrameStats.Duration`（整段 `Render*`），没有 **graft / layout / paint / Lua PCall** 等 span。
3. **Lua / 业务层不可见**：无 `render` 直方图、无 store 次数；Lua 指标字段未接线到主循环。
4. **事件 perf 未贯通**：`RecordEvent` 未在 `HandleEvent` 调用 → **Events** 列无诊断价值。
5. **观测闭环弱**：`SetAlert` 有；无默认阈值策略、无 JSON/Chrome trace 导出、无标准 CI 片段（除手写 `AssertLastFrame`）。
6. **FPS 与渲染帧**：`TickFPS` 基于主循环 tick + **300ms EMA**，与 **`BeginFrame` 次数**可偏离（例如有 tick 但 `RenderDirty` early-out）。
7. **DevTools 刷新**：`tickDevToolsV2` 对整面板 **~300ms** 节流；短尖峰在 Perf 表上可能被平滑。

---

## 5. 演进建议（分阶段）

| 阶段 | 内容 |
|------|------|
| **P0 对齐** | 对 **`V2LayoutNodes` / `V2PaintNodes` / `V2ReconcileChanges`**：在 `layout`/`PaintDirty`/`Reconcile` 路径 **`Record`**，或从 **Metric / Report / DevTools** 移除/标注 **UNIMPLEMENTED**，避免假数。 |
| **P1 展示** | Perf 表增加 **V2 Clear、DirtyArea、WriteFull/WriteDirty/Flush**；**折叠或灰显**恒为 0 的旧列（Renders/Layouts/Paints/Events）。 |
| **P1b Lua** | 在 `cmd/lumina` 或 `App` 主循环调用 **`UpdateLuaMetrics`**（需定义统计来源，如 VM hook 或采样）。 |
| **P2 耗时** | 在 `RenderDirty`/`renderInOrder` 内用 `time.Since` 拆 **graft / layout / paint**（采样开关），写入新 `Metric` 或 `FrameStats` 扩展字段。 |
| **P3 事件** | `RecordEvent` 接入 `app.HandleEvent`（注意高频路径开销）。 |
| **P4 导出** | `Tracker.ExportJSON()` 或文件 dump，供 CI diff。 |

---

## 6. 断言与测试（`pkg/perf/assert.go` + `pkg/app_v2engine_perf_test.go`）

- **`AssertLastFrame(t, checks...)`**：对 **`LastFrame()`** 跑一组 `FrameCheck`。
- **V2 常用**：`CheckV2ComponentsRendered`、`CheckV2PaintCells` / `CheckV2PaintCellsMax`、`CheckV2PaintClearCells`、`CheckV2ComponentsRenderedMax`。
- **旧接口仍可用**：`CheckRenders`、`CheckLayouts`、`CheckPaints`、`CheckEventsDispatched` 等——在 **未接线** 的指标上断言会得到 **0**，易写成「永远通过」或误用。
- **改 Metric 语义或 early-return 分支时**：需同步 **`app_v2engine_perf_test.go`** 与 DevTools 文档预期。

---

## 7. 与 `docs/TESTING.md` 的关系

- 性能回归建议：**`Tracker.Enable()`** 后跑脚本/交互，再 **`AssertLastFrame(CheckV2PaintCellsMax(N), ...)`** 约束过绘。
- 更完整 perf 落地后，可在 **`docs/TESTING.md`** 增加「性能回归」小节（阈值 + 跑法 + 注意 FPS 与 `BeginFrame` 非同一语义）。

---

## 8. 小结

当前 **perf 核心**是：**以单次 `App.Render*` 为帧的计数器 + 60 帧环形历史 + F12 简表 + 文本 Report**；对 **V2 画布的写格与清格、粗脏区面积、组件 render 调用次数** 有参考价值。**完整 perf 工具**通常还要求：**无死指标、与 Report/UI 一致、分阶段耗时、Lua/事件维度、可导出与 CI**。优先 **P0 对齐（接线或下架假数）+ P1 展示补齐**，收益最大、误导最少。

---

## 9. 附录：关键符号索引

| 符号 / 文件 | 说明 |
|---------------|------|
| `perf.Metric` | `pkg/perf/tracker.go` |
| `Tracker.BeginFrame` / `EndFrame` / `Record` | `pkg/perf/tracker.go`；帧外 `Record` 进 **pending**，下一帧 `BeginFrame` 合并 |
| `App.RenderAll` / `RenderDirty` | `pkg/app.go` |
| `Engine.RenderAll` / `RenderDirty` | `pkg/render/engine.go` |
| `CellBuffer.ResetStats` / `Stats` | `pkg/render/cellbuffer.go` |
| `PaintDirty` / `PaintFull` | `pkg/render/painter.go` |
| `paintPerfTab` | `pkg/app_devtools_v2.go` |
| `Panel.TickFPS` / `SnapshotPerf` | `pkg/devtools/devtools.go` |
