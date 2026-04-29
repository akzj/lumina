# 开发指南

> 面向 Lumina 框架开发者的指南。

---

## 环境要求

- Go 1.22+
- Git

---

## 项目结构

```
lumina/
├── cmd/lumina/           CLI 入口
├── pkg/                  核心框架（package v2）
│   ├── render/           渲染引擎
│   │   ├── engine.go     Engine + Lua API
│   │   ├── node.go       Node, Component, Style 类型
│   │   ├── reconciler.go 协调算法
│   │   ├── layout.go     Flex 布局
│   │   ├── painter.go    绘制
│   │   ├── cellbuffer.go CellBuffer
│   │   ├── events.go     事件分发
│   │   └── input.go      输入编辑 + 焦点
│   ├── app.go            App 组合根
│   ├── app_run.go        事件循环
│   ├── app_lua.go        App 级 Lua API
│   ├── app_timer.go      定时器
│   ├── buffer/           Buffer 类型
│   ├── output/           输出适配器
│   ├── event/            事件类型
│   ├── perf/             性能追踪
│   ├── devtools/         开发者工具
│   ├── animation/        动画
│   ├── router/           路由
│   ├── hotreload/        热加载
│   ├── store/            状态管理
│   └── terminal/         终端 I/O
├── examples/             示例 Lua 应用
└── docs/                 文档
```

---

## 构建

```bash
go build ./...
```

---

## 测试

```bash
# 全部测试
go test ./pkg/...

# 渲染引擎单元测试
go test ./pkg/render/...

# App 集成测试
go test ./pkg/ -run TestV2Engine

# E2E 测试（完整 Lua 脚本执行）
go test ./pkg/ -run TestE2E

# Lua 测试框架（pkg/testdata/lua_tests/**/*_test.lua）
go test ./pkg/ -run TestLuaTestFramework -count=1

# 性能测试
go test ./pkg/ -run TestPerf

# 压力测试 benchmark
go test ./pkg/ -bench BenchmarkStress -benchtime 5s

# 带详细输出
go test ./pkg/render/... -v -run TestReconcile
```

**Lua 组件 / 示例应用怎么测**：分层策略、`test.createApp` API、断言习惯与目录约定见 **[TESTING.md](./TESTING.md)**。

---

## 测试策略

### 分层测试

| 层 | 位置 | 说明 |
|---|------|------|
| 渲染引擎 | `render/*_test.go` | 协调、布局、绘制、事件的纯单元测试 |
| App 集成 | `app_v2engine_test.go` | Lua 代码 → 渲染 → 输出的集成测试 |
| Lua 集成 | `lua_test.go` + `testdata/lua_tests/` | `test.describe` / `test.createApp`，覆盖 `lumina.app`、store、hooks、示例 |
| E2E | `lua_e2e_v2_test.go` | 完整 Lua 脚本文件执行 |
| 性能 | `app_v2engine_perf_test.go` | 性能指标验证 |
| Benchmark | `stress_bench_test.go` | 压力测试基准 |

### TestAdapter

测试中使用 `output.TestAdapter` 代替真实终端输出：

```go
app, ta := v2.NewTestApp(80, 24)
app.RunString(`
    lumina.createComponent({
        id = "test",
        render = function(props)
            return lumina.createElement("text", {foreground = "#89B4FA"}, "Hello")
        end,
    })
`)
app.RenderAll()

// 验证输出
screen := app.Screen()
cell := screen.Get(0, 0)
assert.Equal(t, 'H', cell.Char)
assert.Equal(t, "#89B4FA", cell.Foreground)
```

---

## 常见开发任务

### 添加新的元素类型

1. **定义类型字符串** — 在 `render/node.go` 中约定类型名
2. **布局** — 在 `render/layout.go` 的 `computeFlex()` 中添加 case
3. **绘制** — 在 `render/painter.go` 的 `paintNode()` 中添加 case
4. **读取属性** — 在 `render/engine.go` 的 `readDescriptor()` 中读取特有属性
5. **写测试** — 在 `render/` 下添加对应测试
6. **验证**: `go build ./... && go test ./pkg/render/...`

### 添加新的样式属性

1. **添加字段** — 在 `render/node.go` 的 `Style` 结构体中添加
2. **读取** — 在 `render/engine.go` 的 `readStyle()` 中读取 Lua 表字段
3. **布局/绘制** — 在对应的布局/绘制函数中使用
4. **协调** — 在 `render/reconciler.go` 的 `reconcileStyle()` 中判断是否影响布局
5. **写测试**

### 添加新的 Lua API

**Engine 级 API**（渲染相关）:
- 在 `render/engine.go` 中实现 `lua*` 方法
- 在 `RegisterLuaAPI()` 中注册到 `lumina` 表

**App 级 API**（运行时相关）:
- 在 `app_lua.go` 中实现 `lua*` 方法
- 在 `registerAppLuaAPIs()` 中注册

### 添加新的事件类型

1. **Node 字段** — 在 `render/node.go` 的 `Node` 中添加 `OnXxx LuaRef`
2. **Descriptor 字段** — 在 `Descriptor` 中添加对应字段
3. **读取** — 在 `readDescriptor()` 中用 `getRefField()` 读取
4. **协调** — 在 `Reconcile()` 中用 `updateRef()` 更新
5. **分发** — 在 `render/events.go` 中添加分发逻辑
6. **App 集成** — 在 `app.go` 的 `HandleEvent()` 中路由

---

## 调试

### 开发者工具

运行时按 `F12` 切换 DevTools 面板，显示：
- **Elements** — 组件树结构
- **Perf** — 实时性能指标

### 性能追踪

```go
app.Tracker().Enable()
// ... 运行 ...
stats := app.Tracker().Stats()
// V2ComponentsRendered, V2PaintCells, V2DirtyRectArea, ...
```

### 热加载

使用 `--watch` 标志运行，修改 Lua 文件后自动重载：

```bash
lumina --watch examples/counter.lua
```

---

## 代码规范

- 渲染引擎（`render/`）不依赖 `app`、`output` 等外部包
- 所有 Lua 交互通过 `LuaRef`（注册表引用），不持有 Lua 值
- 渲染期间暂停 Lua GC，渲染后执行增量 GC
- 事件处理器存储为 `LuaRef`，不在每帧重新注册
- 使用 `reflect.DeepEqual` 避免无变化的状态更新触发重渲染
