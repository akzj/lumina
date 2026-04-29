# Perf 工具设计分析

> **状态**：与实现同步时以仓库代码为准。  
> **代码入口**：`pkg/perf/`（`Tracker`、`FrameStats`、`Report` / `Assert`）、`pkg/app.go`（`BeginFrame`/`EndFrame`、adapter 输出计数）、`pkg/render/engine.go`（引擎 paint 后计数）、`pkg/devtools/` + `pkg/app_devtools_v2.go`（F12 面板 Perf 页）。

---

## 1. 目标角色（完整工具应覆盖）

| 角色 | 需求 | 当前覆盖度 |
|------|------|------------|
| **框架开发者** | 绘制与写出成本、回归防退化 | 有「组件 render 次数 + 画布写/清 + 脏区面积 + adapter 写出」；**无** layout/reconcile 节点级计数、**无**分阶段耗时 |
| **应用 / Lua 作者** | 哪次 `render` 重、store 更新频率、Lua 耗时 | 弱：`Panel` 上有 `LuaCPUTime` / `LuaMemBytes` 与 `UpdateLuaMetrics` API，**仓库内无调用方**，Perf 表也未展示 |
| **终端 / 宿主** | 写出带宽、flush 次数、脏区面积 | 部分：`WriteFull`/`WriteDirty`/`Flush`、`DirtyRectsOut`；**无**字节带宽 |
| **CI / 自动化** | 可断言的阈值、稳定报告 | 有 `perf.AssertLastFrame`、`CheckComponentsRendered`、`CheckPaintCellsMax` 等，依赖 `Tracker.Enable()`（F12 打开 DevTools 时会 `Enable`） |

---

## 2. 现状架构（数据流）

```
App.RenderAll / RenderDirty
    tracker.BeginFrame()              // 合并上一帧 End 之后产生的 pending
    engine.RenderAll / RenderDirty    // 见 §2.1；可能 early-return 不写引擎计数
    [RenderDirty:] adapter.Write* / Flush → Record(Write*, DirtyRects*, Flush)
    tracker.EndFrame()                // Duration；环形历史；**Total 按指标累加**；Alert

DevTools：F12 打开时 tracker.Enable()；SnapshotPerf() 复制 LastFrame + **TotalStats** + FPS → paintPerfTab 只读快照  
tickDevToolsV2：TickFPS()；可见时每 ~300ms SnapshotPerf + 重画面板（与 FPS EMA 窗口独立）
```

### 2.1 `RenderDirty` 与 perf 的边界情况

- **`engine.needsRender == false`**：`RenderDirty` 在 **`ResetStats` 之后**立刻 `return`，**本帧不产生引擎侧 `Record(ComponentsRendered|Paint*)`**（pending 合并规则仍适用；断言时注意帧边界）。
- **有脏但 `rendered==0` 且无 `hasAnyDirty`**：写 **`PaintCells`/`PaintClearCells`/`DirtyRectArea = 0`** 后 `return`（不跑 layout/paint）。
- **`DirtyRectArea`**：来自 **`CellBuffer.Stats()`** 的 **`DirtyW * DirtyH`**（本帧画布脏矩形包围盒面积），**不是** `app` 传给 `WriteDirty` 的 adapter 矩形面积。

### 2.2 帧语义

- **帧边界**：与 **`App.RenderAll` / `App.RenderDirty` 单次调用** 对齐。
- **`FrameStats.Duration`**：`EndFrame` 时相对 `BeginFrame` 的 wall time。
- **历史**：默认 **60** 帧环形缓冲；`LastFrame()` / `History()` / **`TotalStats()`**（自 Enable 或上次 `Reset` 起的**累计**计数，适合观察尖峰间「看不见」的总量）。

---

## 3. 指标清单（`pkg/perf/tracker.go`）

| Metric | 含义 | 生产路径 |
|--------|------|----------|
| `DirtyRectsOut` | 本帧 `WriteDirty` 传入的矩形个数 | **是**：`app.RenderDirty` |
| `WriteDirtyCalls` / `WriteFullCalls` / `FlushCalls` | 适配器写出 | **是**：`app.RenderAll` / `RenderDirty` |
| `ComponentsRendered` | **渲染次数**（`renderComponent` 调用次数） | **是**：`engine.RenderAll` / `RenderDirty` |
| `PaintCells` | `CellBuffer` 单元格写入次数 | **是**：paint 后 `buffer.Stats().WriteCount` |
| `PaintClearCells` | `ClearRect` / `Clear` 清除的 cell 数 | **是**：`buffer.Stats().ClearCount` |
| `DirtyRectArea` | `DirtyW * DirtyH`（无写入时为 0） | **是** |

### `RecordComponent` / `RecordEvent`

- **`RecordComponent(id)`**：递增渲染次数计数 `ComponentsRendered`，并追加 `RenderComponents`（**测试/插桩**；引擎路径不调用）。
- **`RecordEvent(type, dispatched)`**：仅维护 **`EventsByType`**（`dispatched` 忽略）；应用主路径通常不调用。

### `Report()` / `TotalReport()`（`pkg/perf/report.go`）

紧凑文本：**Output**（dirtyRects、writeDirty、writeFull、flush）；若本帧有渲染或 paint，再打印 **Render**（`renderCount`、paintCells、clearCells、dirtyArea）；可选 `RecordComponent` / `EventsByType` 行。

### DevTools Perf 页（`paintPerfTab`）

- 引擎与 adapter 相关计数：**`total N  last M`**（**total 在前**，便于在刷新节流下阅读累计值）。
- Runtime 段：`runtime.MemStats`、`NumGoroutine()`。

---

## 4. 主要缺口

1. **分阶段耗时**：仅有整帧 `Duration`。
2. **Lua / 业务**：`UpdateLuaMetrics` 未接线到主循环。
3. **事件**：`RecordEvent` 未在 `HandleEvent` 贯通。
4. **FPS 与渲染帧**：`TickFPS` 与 `BeginFrame` 次数可偏离。
5. **DevTools 刷新**：~300ms 节流，单帧尖峰在表上可能被平滑；**Total** 列缓解该问题。

---

## 5. 演进建议（分阶段）

| 阶段 | 内容 |
|------|------|
| **P1 Lua** | 主循环调用 **`UpdateLuaMetrics`**（定义统计来源）。 |
| **P2 耗时** | `RenderDirty` 内拆 graft / layout / paint span（采样开关）。 |
| **P3 事件** | `RecordEvent` 接入 `HandleEvent`（注意高频开销）。 |
| **P4 导出** | JSON / trace dump 供 CI。 |

---

## 6. 断言与测试

- **`AssertLastFrame(t, checks...)`**：对 `LastFrame()` 跑 `FrameCheck`。
- **常用**：`CheckComponentsRendered`、`CheckPaintCells` / `CheckPaintCellsMax`、`CheckPaintClearCells`、`CheckDirtyRectArea`、`CheckMetric`、`CheckRenderComponents`。
- 引擎与 perf 行为变更时同步 **`app_v2engine_perf_test.go`** 与本文档。

---

## 7. 小结

Perf 以 **`App.Render*` 为帧**：环形历史 + **累计 Total** + F12 表（total/last）+ 文本 Report；对画布写清、粗脏区面积、组件 render 次数与 adapter 写出有参考价值。

---

## 8. 附录：关键符号索引

| 符号 / 文件 | 说明 |
|---------------|------|
| `perf.Metric` | `pkg/perf/tracker.go` |
| `Tracker.BeginFrame` / `EndFrame` / `Record` | 帧外 `Record` 进 **pending**，下一帧 `BeginFrame` 合并 |
| `App.RenderAll` / `RenderDirty` | `pkg/app.go` |
| `Engine.RenderAll` / `RenderDirty` | `pkg/render/engine.go` |
| `CellBuffer.ResetStats` / `Stats` | `pkg/render/cellbuffer.go` |
| `paintPerfTab` | `pkg/app_devtools_v2.go` |
| `Panel.TickFPS` / `SnapshotPerf` | `pkg/devtools/devtools.go` |
