# Lumina 架构设计文档 v0.3

> **版本**: v0.3（与仓库主分支实现对齐）  
> **状态**: 持续演进  
> **日期**: 2026-04-26  
> **核心理念**: Go 运行时 + Lua 声明式 UI；真实 TUI（布局、事件、差分输出），面向人机协作与 MCP 调试。

**配套文档**: API 细节以 [`docs/API.md`](docs/API.md) 为准（与 `pkg/lumina/lumina.go` 中 `luaLoader` 同步）。

---

## 1. 项目定位

**Lumina** = 终端上的类 React 体验：Lua 描述 VDOM，Go 负责布局、栅格、输入与输出。

| 维度 | 说明 |
|------|------|
| 目标 | 高效开发终端 UI；支持本机运行与 WebSocket/xterm 同源体验 |
| 核心技术栈 | **Go**（运行时、渲染、I/O）+ **Lua**（`github.com/akzj/go-lua`）+ **可选 MCP**（JSON-RPC HTTP 等） |
| 输出 | **ANSI** 终端栅格为主；可切换 **JSON** 等模式供机器消费 |
| 类比 | Web 的 React（组件、Hooks、VDOM），但布局为 **字符栅格 Flex**，非浏览器 CSS 全量 |

---

## 2. 核心架构

```
┌────────────────────────────────────────────────────────────────────┐
│                         Lumina 运行时                               │
├────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐     JSON-RPC / 工具      ┌───────────────────┐   │
│  │ MCP / 调试   │◄────────────────────────►│  lumina.inspect   │   │
│  │ 客户端       │                            │  simulate* 等     │   │
│  └──────────────┘                            └─────────┬─────────┘   │
│                                                         │            │
│  ┌─────────────────────────────────────────────────────▼────────┐   │
│  │ App（单线程事件循环）                                            │   │
│  │  ticker / SIGWINCH → renderAllDirty → renderComponent(Lua)     │   │
│  │  input → handleEvent（键盘、鼠标、resize…）                      │   │
│  └───────────────┬────────────────────────────────────────────────┘   │
│                  │                                                    │
│         ┌────────▼────────┐         ┌─────────────────────────┐     │
│         │ Component 注册表 │◄───────►│ go-lua State            │     │
│         │ + Hook 状态      │  PCall  │ defineComponent / hooks │     │
│         └────────┬────────┘         └─────────────────────────┘     │
│                  │                                                    │
│         ┌────────▼────────────────────────────────────────────────┐   │
│         │ 渲染管线（Go）                                            │   │
│         │ Lua VNode 表 → VNode 树 → computeFlexLayout               │   │
│         │ → VNodeToFrame 或 DiffVNode + ApplyPatches（增量）        │   │
│         │ → Frame（Cells + DirtyRects）→ OutputAdapter.Write        │   │
│         └──────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────────┘
```

**要点**:

- **单 VM、单线程**：组件渲染与 Lua 调用在同一条逻辑线程上（避免在回调里并发动 VM）。
- **根组件**通过 `lumina.mount` 注册；`lumina.run()` 进入终端 raw 模式 + 事件循环。
- **全局注册表**保存 `Component` 指针；子组件在渲染树展开时挂载/通过 `ReconcileComponents` 清理。

---

## 3. 组件模型

### 3.1 三层分工

| 层 | 实现 | 职责 |
|---|------|------|
| Go `*Component` | `pkg/lumina/component.go` | 唯一 ID、`Props`、`State` map、Hook 槽位、脏标记、父子链 |
| Lua 工厂表 | `defineComponent` 生成 | `init`、`render` 等函数引用在 Registry |
| VNode 树 | Lua 返回 plain table | `type` / `style` / `children` / `props` / 事件字段；由 Go 转为 `*VNode` |

组件 **不是** 每个实例一个 Lua userdata 元表模型；逻辑上仍是 **Go 持有实例 + Lua 闭包渲染**。

### 3.2 生命周期（与实现对应）

| 阶段 | 行为 |
|------|------|
| 首次渲染 | 创建 `Component`，可选执行 Lua `init`，再 `render` 得到 VNode |
| 更新 | `SetState` / store 订阅 / 事件标记 `Dirty`；下一帧 `renderComponent` 再次执行 `render` |
| 卸载 | VNode 树对比后 `ReconcileComponents` 对消失的子组件 `Cleanup`（解绑 Lua ref、effect cleanup） |

### 3.3 Hooks：仅在 `render` 中调用

与 React 16.8+ 一致：**所有 Hooks 必须在组件的 `render` 函数体内同步调用**（`hooks.go` + `ResetHookIndex` 每帧重置索引）。

### 3.4 `useState`（必须带 key）

状态挂在 `Component.State[string]` 上；Lua API 为：

```lua
local count, setCount = lumina.useState("count", 0)
setCount(count + 1)
```

第一个参数为 **稳定字符串 key**（同组件多段 `useState` 靠 key 区分槽位）。

### 3.5 最小计数器示例（与实现对齐）

```lua
local Counter = lumina.defineComponent({
    name = "Counter",
    render = function()
        local n, setN = lumina.useState("n", 0)
        return {
            type = "vbox",
            children = {
                { type = "text", content = "Count: " .. tostring(n) },
                {
                    type = "box",
                    id = "inc",
                    style = { border = "rounded" },
                    onClick = function() setN(n + 1) end,
                    children = {
                        { type = "text", content = " +1 " },
                    },
                },
            },
        }
    end,
})
```

（具体可点击节点类型以项目内示例为准，如 `examples/counter.lua`。）

### 3.6 脏更新路径

`Component.SetState`：写入 `State` → `Dirty.Store(true)` → 沿父链标记根组件脏（子组件需根重渲染才能再次执行 Lua 子树）。

全局 **`createStore` + `useStore`** 在 `dispatch` 时通知订阅的组件 `Dirty`。

---

## 4. 渲染引擎

### 4.1 数据流

```
Lua 返回 VNode 表  →  LuaVNodeToVNode  →  *VNode 树
        →  computeFlexLayout（整树，字符栅格 Flex 语义）
        →  VNodeToFrame（全量）或 DiffVNode + ApplyPatches（条件增量）
        →  Frame.Cells + DirtyRects
        →  bridgeVNodeEvents（事件目标绑定）
        →  OutputAdapter.Write（ANSI：整帧或按 DirtyRects 与 prevFrame  cell diff）
```

### 4.2 帧缓冲（`Frame`）

定义见 `pkg/lumina/output.go`：

- `Cells [][]Cell`：与终端同尺寸的字符单元（含颜色、OwnerNode 等）。
- `DirtyRects`：本帧变更矩形集合，供 ANSI 适配器缩小扫描范围。
- **无**早期设计稿中的 `Frame.Metadata` 字段；元数据通过其它调试通道（MCP / inspect）获取。

`Cell` 含 `Char`、`Foreground`、`Background`、样式位、`OwnerNode` / `OwnerID`（命中测试与 inspector）。

### 4.3 增量策略（概览）

实现分散在 `app.go`、`vdom_diff.go` 等：

1. **组件级**：仅 `Dirty` 为真的组件执行 Lua `render`。
2. **VDOM**：`DiffVNode(old, new)` 产出 `Patch` 列表；无 patch 且无视口滚动等特殊情况时可跳过绘制。
3. **帧级增量**：在 patch 较少、`lastFrame` 存在且未强制全量条件时，`ApplyPatches` 在**旧 Frame** 上按「受影响的父容器」重画子树，并维护 `DirtyRects`。
4. **终端输出**：`ANSIAdapter` 在尺寸变化或 `Invalidate` 后走全量写；否则对 dirty 区域与 `prevFrame` 做 cell 级 diff。

**限制（有意为之）**：`ApplyPatches` 以**父容器**为单位重画，避免兄弟位移错误；大改动时退回 `VNodeToFrame` 全量。

### 4.4 布局（`layout.go`）

- 显式节点类型：`fragment`、`text`、`vbox`、`hbox`、`input`、`textarea`；其它 `type`（如 `box`）走**默认竖直栈**，语义接近 `vbox`。
- `style`：`flex`、`gap`、`justify`（`start|center|end|space-between|space-around`）、`align`、`overflow=scroll`（与 viewport / 滚动 API 配合）等。
- **非**完整 CSS Grid/Flexbox；不要按浏览器假设等价。

### 4.5 输出适配器

- **`ANSIAdapter`**：`Write` 缓冲整帧输出，隐藏光标、单 `Write` 系统调用减少撕裂。
- **JSON / MCP**：通过 `setOutputMode` 等与 MCP 帧抓取（如 `getMCPFrame`）配合；详见 `docs/API.md`。

---

## 5. 输入与事件

- **键盘 / 鼠标 / 定时器**：输入 goroutine 将消息投递到 `App.events`，主循环 `handleEvent` 统一处理（`app.go` + `input.go`）。
- **合成事件**：例如 `mousemove` 上根据 `Target` 变化合成 `mouseenter` / `mouseleave`（`EventBus`）。
- **焦点**：全局可聚焦 ID 列表 + `setFocus` / Tab 顺序；另有 `pushFocusScope` / `popFocusScope`。
- **注意**：`resize` 时会清空 `hoveredID`；若 Lua 侧用 `useState` 维护悬停，需与引擎策略一致，避免出现「尺寸变化后悬停态残留」（见 issue 讨论与后续改进方向）。

---

## 6. MCP 与调试面

### 6.1 当前形态（以仓库为准）

- **`cmd/lumina-mcp-http`**：Streamable HTTP 上 **JSON-RPC 2.0** 风格 MCP 工具注册（Go `encoding/json`），用于远程多客户端调试。
- **Lua 侧**：`lumina.inspect*`、`simulate*`、`diff`、`patch`、`eval` 等挂在模块上，由 MCP 或本进程调试入口调用（见 `inspect_api.go`、`mcp_debug.go` 等）。
- **`proto/v1/lumina.proto`**：存在历史/扩展 protobuf 定义；**线上 MCP 主路径以 JSON-RPC + 结构化结果为主**，不要把「全链路 Protobuf」写死为当前唯一实现。

### 6.2 设计取舍（记录）

| 方向 | 说明 |
|------|------|
| JSON 便于人类与脚本调试 | 与 MCP 生态常见传输一致 |
| Proto 可选 | 适合未来高性能或强 schema 场景 |

---

## 7. 模块与 UI 库

- **`require("lumina")`**：`luaLoader` 注册大量 API（Hooks、路由、窗口管理、`lumina.ui` preload 等），权威列表见 **`docs/API.md`** 的 Module index。
- **`require("lumina.ui")`**：`pkg/lumina/components/ui/init.lua` 聚合的终端版 shadcn 风格组件；单文件为 `require("lumina.ui.button")` 等。
- **`require("shadcn")` / `require("shadcn.button")`**：兼容别名（`RegisterShadcn`），与 `lumina.ui` 同源文件。

---

## 8. 测试与质量

- **单元 / 集成测试**：`pkg/lumina/*_test.go` 覆盖布局、diff、输入、composition 等。
- **E2E**：`e2e_full_test.go`、`e2e_showcase_test.go` 等验证 Lua 管线。
- **Headless 测试辅助**：`lumina.createTestRenderer()`（VDOM 级，非完整 App），见 `lua_accessibility.go`。

---

## 9. 演进与已知边界

### 9.1 已相对成熟的能力

- Flex 布局、滚动视口、overlay、窗口管理、热重载入口、路由全局表、数据 `fetch`/`useQuery`、大量 Hooks、DevTools / Inspector（F12）等（以 git 与 `docs/API.md` 为准）。

### 9.2 仍在演进或简化的点

- 布局语义与浏览器 CSS 不完全等价；复杂排版需按 TUI 约束设计。
- 增量渲染与悬停状态在极端 resize 下的边界情况可继续收紧。
- 插件、多进程隔离等大特性：**按需迭代**，本文件不逐条承诺时间表。

### 9.3 历史附录（v0.1 快照）

早期「Phase 1 完成状态」中的部分限制（如「无 Diff Engine」）**已被后续实现超越**；保留旧清单易产生误解，故 **不再逐条复制**。若需考古，请 `git log` / `git show` 查看历史 `DESIGN.md`。

---

## 10. 远期设想（非当前承诺）

以下条目来自原设计文档中的扩展思考，**尚未**作为整体落地保证；实现时以 Issue/PR 为准：

- 更细的分层 `ComponentID`（scope:type:version:instance）与跨作用域隔离。
- 通过 `providers` 表声明式 Context（当前已有 `createContext` / `useContext` API，形态与旧伪代码不同，见 `docs/API.md`）。
- 渲染器插件接口标准化（当前以 `OutputAdapter` 实现为主）。

---

## 附录

| 资源 | 路径 |
|------|------|
| Lua API 参考 | `docs/API.md` |
| 主运行时入口 | `pkg/lumina/app.go`、`pkg/lumina/lumina.go` |
| 布局 | `pkg/lumina/layout.go` |
| VDOM diff / patch | `pkg/lumina/vdom_diff.go` |
| ANSI 输出 | `pkg/lumina/ansi_adapter.go` |
| go-lua | `github.com/akzj/go-lua`（见 `go.mod`） |

---

*文档结束。*
