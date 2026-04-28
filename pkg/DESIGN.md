# V2 Package 设计

> 本文档描述 `pkg/lumina/v2/` 包的内部设计和模块职责。

---

## 模块架构

```
pkg/lumina/v2/
│
├── render/           核心渲染引擎（无外部依赖，纯计算）
│   ├── engine.go     Engine 结构体 + Lua API + 帧渲染
│   ├── node.go       Node, Component, Style, Descriptor 类型
│   ├── reconciler.go 协调算法（Descriptor vs Node 树）
│   ├── layout.go     Flex 布局（vbox/hbox, 增量）
│   ├── painter.go    绘制（full/dirty/clipped/scroll）
│   ├── cellbuffer.go CellBuffer 2D 网格 + 脏区追踪
│   ├── events.go     HitTest + 鼠标/键盘/滚动分发
│   └── input.go      输入编辑 + 焦点管理
│
├── app.go            App 组合根 — 集成所有模块
├── app_run.go        事件循环 + 热加载 + 输入转换
├── app_lua.go        App 级 Lua API（quit, timer）
├── app_timer.go      setInterval/setTimeout 管理
├── app_devtools_v2.go DevTools 集成
│
├── buffer/           通用 Buffer 类型（输出适配器接口）
├── output/           输出适配器
│   ├── ansi.go       ANSI 终端输出
│   └── test.go       TestAdapter（测试用）
├── event/            事件类型定义
├── perf/             性能追踪器
├── devtools/         开发者工具面板
├── animation/        动画管理器
├── router/           路由
├── hotreload/        文件监控热加载
├── store/            状态管理
└── terminal/         终端 I/O
```

---

## 依赖关系

```
                    render (核心，无外部依赖)
                   /   |   \
                  /    |    \
            buffer  event   perf
              |
           output (依赖 buffer)
              |
             app (组合根，依赖所有模块)
            / | \
     devtools animation router hotreload store
```

**设计原则**: `render/` 包不依赖 `app`、`output`、`devtools` 等。
它只依赖 `buffer`（用于 `ToBuffer()`）和 `perf`（性能追踪）。

---

## App 组合根

`App` 是组合根，负责：

1. **创建** Engine、Tracker、DevTools、AnimationManager、Router、TimerManager
2. **注册** Engine 的 Lua API + App 级 Lua API
3. **运行** 事件循环（输入 → 处理 → 渲染 → 输出）
4. **集成** DevTools 覆盖层绘制

```go
type App struct {
    engine    *render.Engine      // 渲染引擎
    adapter   output.Adapter      // 输出适配器
    tracker   *perf.Tracker       // 性能追踪
    devtools  *devtools.Panel     // 开发者工具
    animMgr   *animation.Manager  // 动画
    routerMgr *router.Router      // 路由
    timerMgr  *timerManager       // 定时器
    luaState  *lua.State          // Lua VM
}
```

---

## Lua API 注册

API 分两层注册：

### Engine 层（render/engine.go）

```lua
lumina.createElement(type, props, ...children)
lumina.defineComponent(name, renderFn)
lumina.createComponent(config)
lumina.useState(key, defaultValue)
```

### App 层（app_lua.go）

```lua
lumina.quit()
lumina.setInterval(fn, ms)
lumina.setTimeout(fn, ms)
lumina.clearInterval(id)
lumina.clearTimeout(id)
```

---

## 事件循环

```
事件循环 (60fps ticker):
  ├─ 输入事件 → handleInputEvent() → Engine.Handle*()
  ├─ 热加载事件 → reloadScript() → RenderAll()
  └─ Tick:
      ├─ 动画 tick
      ├─ 定时器 fire
      ├─ DevTools tick
      └─ RenderDirty() → Engine → ToBuffer → WriteDirty → Flush
```

**输入处理**: 输入事件在 select 中立即处理（状态变化标记 Dirty），
渲染在下一个 ticker tick 时执行。多个输入合并到同一帧。

---

## 测试策略

### TestAdapter

`output.TestAdapter` 实现 `output.Adapter` 接口，记录所有输出调用：

```go
type TestAdapter struct {
    LastFull  *buffer.Buffer
    LastDirty *buffer.Buffer
    DirtyRects []buffer.Rect
    // ...
}
```

### 测试模式

```go
app, ta := v2.NewTestApp(80, 24)
app.RunString(`lumina.createComponent({...})`)
app.RenderAll()

// 验证输出
screen := app.Screen()
cell := screen.Get(10, 5)
assert(cell.Char == 'H')
```

### 测试分层

| 层 | 文件 | 测试内容 |
|---|------|---------|
| render | `render/*_test.go` | 协调、布局、绘制、事件 |
| app | `app_v2engine_test.go` | Lua→渲染→输出集成 |
| e2e | `lua_e2e_v2_test.go` | 完整 Lua 脚本执行 |
| perf | `app_v2engine_perf_test.go` | 性能指标验证 |
| bench | `stress_bench_test.go` | 压力测试基准 |

---

## 性能追踪

`perf.Tracker` 记录每帧的关键指标：

```go
const (
    V2ComponentsRendered  // 本帧渲染的组件数
    V2PaintCells          // 本帧写入的 cell 数
    V2PaintClearCells     // 本帧清除的 cell 数
    V2DirtyRectArea       // 脏区面积
    WriteFullCalls        // WriteFull 调用次数
    WriteDirtyCalls       // WriteDirty 调用次数
    DirtyRectsOut         // 输出的脏区数
    FlushCalls            // Flush 调用次数
)
```

启用方式: `app.Tracker().Enable()`

DevTools 面板使用 Tracker 数据显示实时性能指标。
