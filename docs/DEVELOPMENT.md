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
# 单文件（路径相对 pkg/）或封装脚本
LUMINA_LUA_TEST=testdata/lua_tests/examples/scrollview_test.lua go test ./pkg/ -run TestLuaTestFramework -count=1
./scripts/lua-test.sh scrollview

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

Elements 向「Chrome 程度」体验的 **MVP 与阶段规划** 见 **[DESIGN-devtools-elements.md](./DESIGN-devtools-elements.md)**。

### 性能追踪

```go
app.Tracker().Enable()
// ... 运行 ...
f := app.Tracker().LastFrame()
_ = f.Get(perf.ComponentsRendered) // renderComponent invocations per frame
_ = f.Get(perf.PaintCells)
_ = f.Get(perf.DirtyRectArea)
// 或 app.Tracker().Report() / TotalReport()
```

**设计说明（指标是否接线、DevTools 展示缺口、演进阶段）** 见 **[DESIGN-perf.md](./DESIGN-perf.md)**。

### 热加载

使用 `--watch` 标志运行，修改 Lua 文件后自动重载：

```bash
lumina --watch examples/counter.lua
```

---

## 协议字符串速查（Lua ↔ 引擎）

以下为 **代码与脚本里出现的约定字符串**（非穷举每个分支值），含义与典型使用场景。实现细节以 `pkg/render/engine.go`（`readDescriptor` / `readStyle`）、`pkg/render/events.go`、`pkg/render/node.go` 为准；对外 API 摘要见 **[API.md](./API.md)**。

### 应用输入 `event.Event.Type`（`pkg/app.go` → `Engine`）

| 字符串 | 含义 | 场景 |
|--------|------|------|
| `mousedown` / `mouseup` / `mousemove` | 指针事件 | 终端 / WebSocket 适配器上报 |
| `click` | 点击（或引擎在 mouseup 同格合成） | 路由到 `HandleClick` |
| `keydown` / `keyup` | 键盘 | `HandleKeyDown` 等 |
| `scroll` | 滚轮 | `HandleScroll`；`Event.Key` 常承载 `"up"`/`"down"` |
| `resize` | 终端尺寸变化 | `Resize` + 全量重绘 |

### 元素描述符表字段（`lumina.createElement` / `render` 返回的表）

Lua 表 **键名**（字符串）与 Go 侧 `readDescriptor` 读取一致。

| 字符串 | 含义 | 场景 |
|--------|------|------|
| `type` | 节点种类 | 见下表「节点 `type`」；省略时默认 **`box`** |
| `id` | 节点标识 | 测试 `app:click("id")`、调试、部分 MCP 查询 |
| `key` | 稳定键 | 列表/reconciler；子 **组件** 占位映射到 `ChildMap` |
| `content` | 主文本内容 | `text`；`input`/`textarea` 受控内容 |
| `value` | 与 `content` 等价（输入） | 受控 `input` / `textarea` |
| `placeholder` | 占位文案 | `input` / `textarea` |
| `scrollY` | 垂直滚动偏移（整数） | 带 `overflow: scroll` 的容器初始/受控滚动 |
| `style` | 嵌套样式子表 | 任意元素 |
| `children` | 子元素数组 | 容器；元素可为表或字符串（字符串→**`text`** 子节点） |
| `focusable` / `disabled` / `autoFocus` | 布尔 | 焦点环、禁用、首焦 |
| `onClick` / `onMouseEnter` / … | Lua 回调（registry ref） | 见下表「节点事件名」 |
| `_factoryName` / `_props` | 组件工厂 | `lumina.Button` 等返回的占位；`type` 在内部变为 **`component`** |

### 样式子表 `style = { ... }` 常用键名

数值多为 **cell 单位**（整数或 Lua number）。字符串键在 `readStyle` / `readStyleFromMap` / `readStyleFields` 中读取。

| 字符串 | 含义 | 场景 |
|--------|------|------|
| `width` / `height` / `flex` | 尺寸与伸缩 | 布局、`flex` 子分配 |
| `padding` / `paddingTop` / `paddingBottom` / `paddingLeft` / `paddingRight` | 内边距 | 文本区、滚动可视区 |
| `margin` / `marginTop` / … | 外边距 | 与 flex gap 等组合 |
| `gap` | 主轴间隙 | `vbox`/`hbox` |
| `minWidth` / `maxWidth` / `minHeight` / `maxHeight` | 约束 | 防止撑破/收缩 |
| `justify` / `align` | flex 对齐 | `start`/`end`/`center`/`stretch` 等（以 `layout.go` 解析为准） |
| `border` | 边框样式 | **`none`** / **`single`** / **`rounded`** 等；影响占位与裁剪 |
| `foreground` / `background` | 颜色 | **`fg`** / **`bg`** 为别名 |
| `bold` / `dim` / `underline` | 文本样式 | 布尔 |
| `overflow` | 溢出行为 | 见下表「`overflow` 取值」 |
| `position` | 定位模式 | 见下表「`position` 取值」 |
| `top` / `left` / `right` / `bottom` | 偏移 | 与 `position` 配合；`right`/`bottom` 常用 **`-1`** 表「贴边」约定 |
| `zIndex` | 叠放（有限使用） | 层内顺序辅助 |

### 节点 `type`（基元与占位）

| 字符串 | 含义 | 场景 |
|--------|------|------|
| `box` | 通用块 | **默认**；无 `type` 时 |
| `vbox` / `hbox` | 纵向/横向 flex | 布局、Go Widget 根节点常用 `vbox` |
| `text` | 文本叶 | `content` 或 `children` 字符串简写 |
| `input` / `textarea` | 可编辑 | 键盘由 `HandleInputKeyDown` 优先处理 |
| `component` | 子组件占位 | 工厂节点；真实类型在 `ComponentType` + Lua `_props` |

### 样式语义值（字符串，常用）

| 字符串 | 字段 | 含义 | 场景 |
|--------|------|------|------|
| `scroll` | `overflow` | 可滚动容器 | 裁剪内容、`ScrollY`、`autoScroll`、滚动条列 |
| `hidden` | `overflow` | 裁剪、不滚动 | 如 Window 内容区默认 |
| `none` / `single` / `rounded` | `border` | 边框绘制与占位 | `hitTest` 内可视区裁剪 |
| `fixed` / `absolute` | `position` | 定位 | 与 `top`/`left` 等配合 |

### Lua 节点事件属性名 ↔ 引擎 `hasHandler` / 分发

Lua 表字段 **`onXxx`**（小驼峰）对应节点上 `OnXxx` ref。  
**`HitTestWithHandler` / `hasHandler` 的 `eventType` 字符串**（用于带 handler 的命中）：

| eventType | 对应 Lua 字段 | 场景 |
|-----------|----------------|------|
| `click` | `onClick` | 点击冒泡 |
| `mousedown` / `mouseup` | `onMouseDown` / `onMouseUp` | 按下/抬起 |
| `mouseenter` / `mouseleave` | `onMouseEnter` / `onMouseLeave` | 悬停（引擎 + Widget 协同） |
| `keydown` | `onKeyDown` | 未被子组件消费时的 Lua 快捷键 |
| `scroll` | `onScroll` | 自定义滚轮处理（优先于 `autoScroll`） |
| `submit` | `onSubmit` | 如输入框内 Enter 向上冒泡 |
| `outsideclick` | `onOutsideClick` | 设计上有此分支；Modal 外点击等见 `events.go` |

### `dispatchWidgetEvent` → Go Widget 的 `WidgetEvent.Type`

引擎向 **`WidgetDef.OnEvent`** 传入的 **`event.Type`**（与 Lua 节点 `onXxx` 不是同一套名字）：

| Type | 场景 |
|------|------|
| `mousemove` / `mouseenter` / `mouseleave` | 指针与悬停 |
| `mousedown` / `mouseup` / `click` | 按键与点击 |
| `keydown` | 焦点在组件子树内时的键盘（`event.Key` 为键名） |

部分 Widget 在单元测试或内部还会使用 **`focus` / `blur`**（见 `WidgetEvent` 注释）；与 **Lua 节点 `onFocus`/`onBlur`**（由 `setFocus` 直接 `callLuaRefSimple`）路径不同，勿混用。

### 其它约定

| 字符串 | 含义 | 场景 |
|--------|------|------|
| `_childNodes` | Go Widget 子节点切片 | 引擎在渲染 Lua 子树时注入 `props` |
| `ScrollHeight` | 非 Lua 字符串；为 **Node 字段** | layout 写入，供 `computeMaxScrollY` |

**说明**：未在表中列出的键仍可能存在于扩展样式或未来字段；加新协议时建议同步本表与 **API.md**。

---

## 代码规范

- 渲染引擎（`render/`）不依赖 `app`、`output` 等外部包
- 所有 Lua 交互通过 `LuaRef`（注册表引用），不持有 Lua 值
- 渲染期间暂停 Lua GC，渲染后执行增量 GC
- 事件处理器存储为 `LuaRef`，不在每帧重新注册
- 使用 `reflect.DeepEqual` 避免无变化的状态更新触发重渲染
