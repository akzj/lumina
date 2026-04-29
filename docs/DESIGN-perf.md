# Perf 工具设计分析

> **状态**：设计评审（与实现可不同步）。  
> **代码入口**：`pkg/perf/`（`Tracker`、`FrameStats`、`Report` / `Assert`）、`pkg/app.go`（`BeginFrame`/`EndFrame`、输出计数）、`pkg/render/engine.go`（V2 计数）、`pkg/devtools/` + `pkg/app_devtools_v2.go`（F12 面板 Perf 页）。

---

## 1. 目标角色（完整工具应覆盖）

| 角色 | 需求 | 当前覆盖度 |
|------|------|------------|
| **框架开发者** | 布局/协调/绘制/事件各阶段成本、回归防退化 | 部分（有计数，缺分阶段耗时） |
| **应用 / Lua 作者** | 哪次 `render` 重、store 更新频率、Lua 耗时 | 弱（无 Lua 级埋点） |
| **终端 / 宿主** | 写出带宽、flush 次数、脏区面积 | 部分 |
| **CI / 自动化** | 可断言的阈值、稳定报告 | 有 `perf.AssertLastFrame` 等，依赖 `Tracker.Enable()` |

---

## 2. 现状架构（数据流）

```
App.RenderAll / RenderDirty
    tracker.BeginFrame()
    engine.RenderAll / RenderDirty   →  tracker.Record(V2*)
    adapter.WriteFull / WriteDirty   →  tracker.Record(Write*, DirtyRects*, Flush)
    tracker.EndFrame()               →  环形历史 + Total + 可选 Alert

DevTools (F12) 打开时 tracker.Enable()；SnapshotPerf() 冻结 Last/Total + FPS 再画 Perf 页
```

- **帧边界**：与 **`RenderAll` / `RenderDirty` 调用** 对齐，不是严格 OS vsync；事件循环里 ticker 与输入交错时，「一帧」语义 = 一次完整渲染管线调用。
- **历史**：默认约 60 帧环形缓冲；`LastFrame()` / `History()` / `TotalStats()`。

---

## 3. 指标清单：定义 vs 实际写入

### 3.1 V2 引擎与输出（**已接线**，可信）

| Metric | 含义 | 写入位置（概要） |
|--------|------|------------------|
| `V2ComponentsRendered` | 本帧执行了 `renderComponent` 的组件数 | `engine.RenderDirty` / early-out 分支 |
| `V2PaintCells` / `V2PaintClearCells` | 单元格写入 / Clear 次数 | 来自 `buffer.Stats()` _paint 后 |
| `V2DirtyRectArea` | 脏矩形面积 Σ(W×H) | 同上 |
| `WriteFullCalls` / `WriteDirtyCalls` / `FlushCalls` | 适配器输出 | `app.RenderAll` / `RenderDirty` |
| `DirtyRectsOut` | 本帧 `WriteDirty` 传入的矩形个数 | `RenderDirty` |

### 3.2 V2 枚举中存在、**当前未 `Record`** 的指标

| Metric | 设计意图 | 现状 |
|--------|----------|------|
| `V2LayoutNodes` | 参与布局的节点数 | **未接线**，恒为 0 |
| `V2PaintNodes` | 参与绘制的节点数 | **未接线**，恒为 0 |
| `V2ReconcileChanges` | 协调时变更节点数 | **未接线**；`Report()` 仍打印，易误解 |

### 3.3 旧管线指标（Renders / Layouts / Paints / Occlusion / Compose / HitTester / Handler / Events / Components…）

- **设计**：面向早期合成器 + `RecordComponent` + `RecordEvent` 的模型。
- **现状**：`RecordComponent` / `RecordEvent` **仅在 `perf` 包测试中出现**，V2 主路径 **未调用**；故 DevTools 里 **Renders / Layouts / Paints、Events Hit/Miss** 等在真实应用中 **多为 0**，与「有无负载」无关，**易误导**。
- **Report() / TotalReport()**：仍大段打印上述计数，**与 V2 实际热点不对齐**。

### 3.4 DevTools Perf 页实际展示（`paintPerfTab`）

当前约展示：**FPS、帧耗时、Max、TotalFrames、Renders/Layouts/Paints（旧）、V2 Rendered、V2 Paint Cells、DirtyRects、Events Hit/Miss、Go 内存/协程**。

缺失示例：**V2PaintClearCells、V2DirtyRectArea、WriteDirty/WriteFull 次数、Lua 时间（Panel 有字段但未在 Perf 表列出）** 等。

---

## 4. 主要缺口（「不够完整」指什么）

1. **指标与实现两套皮**：枚举 + Report 很全，**生产路径只填一小子集**；未填指标应标注 deprecated 或接线，避免「有数无源」。
2. **缺分阶段耗时**：仅有 `FrameStats.Duration`（整段 Render），没有 **layout-only / paint-only / Lua PCall** 等 span，难以做 Chrome Performance 式瓶颈定位。
3. **Lua / 业务层不可见**：无 `render` 耗时直方图、无 store 更新次数、无组件级标签（除非再扩展 API）。
4. **事件 perf 未贯通**：`RecordEvent` 未在 `HandleEvent` 路径调用，**EventsDispatched/Missed** 无意义。
5. **观测与反馈闭环弱**：`SetAlert` 有，但无内置阈值策略、无导出 JSON/Chrome trace、无 CI 模板。
6. **FPS 与「渲染帧」**：`TickFPS` 基于 ticker，**与 `BeginFrame` 次数不一定一致**；文档应说明「FPS ≈ 主循环节拍」而非 GPU 帧率。
7. **DevTools 刷新**：约 300ms 节流，适合 TUI；**细粒度卡顿**可能被平滑掉。

---

## 5. 演进建议（分阶段，不要求一次做完）

| 阶段 | 内容 |
|------|------|
| **P0 对齐** | 要么在 layout/paint/reconcile 接线 `V2LayoutNodes` 等，要么从 **Metric / Report / DevTools** 中移除或标记「未实现」，避免假数。 |
| **P1 展示** | Perf 页补齐 **V2Clear、DirtyArea、Write/Flush**；**隐藏或折叠**恒为 0 的旧列。 |
| **P2 耗时** | 在 `RenderDirty` 内用 `time.Since` 拆 **graft / layout / paint / compose**（或采样开关），写入新字段或子结构。 |
| **P3 语义** | 可选：`RecordEvent` 接入 `app.HandleEvent`；Lua 侧可选 hook（`debug` 或专用 API）记 **PCall 累计时间**（注意开销）。 |
| **P4 导出** | `Tracker.ExportJSON()` / 文件 dump，便于 CI 与对比。 |

---

## 6. 与测试体系的关系

- `app_v2engine_perf_test.go`、`perf.AssertLastFrame`、`CheckV2*` 依赖当前接线；**改 Metric 语义时需同步测试**。
- 更完整 perf 后，可在 **`docs/TESTING.md`** 增加「性能回归」小节（阈值断言 + 跑法）。

---

## 7. 小结

当前 **perf 核心**是：**按次 Render 为帧的计数器 + 环形历史 + F12 简表 + 文本 Report**，对 **V2 绘制量与脏区面积** 有一定价值；**完整 perf 工具**通常还要求：**真实接线、分阶段耗时、Lua/业务维度、无死指标、可导出与 CI**。优先做 **P0 对齐 + P1 展示**，收益最大、误导最少。
