# Lumina 架构设计文档 v0.2

> **版本**: v0.2 (Phase 2 Complete)  
> **状态**: ✅ 实现完成  
> **日期**: 2024-04-24  
> **核心理念**: 基础牢固，可持续迭代，真实的 TUI 框架，不是 demo

---

## 1. 项目定位

**Lumina** = Terminal React for AI Agents

| 维度 | 说明 |
|------|------|
| 目标 | 让 AI Agent 能高效研发 TUI 应用 |
| 核心技术栈 | Go + Lua + MCP |
| 复用 | go-lua 解释器（100% 控制权） |
| 目标用户 | AI Agent（机器可解析输出优先） |
| 类比 | Web 端的 React |

---

## 2. 核心架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         Lumina 架构                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐   │
│  │  AI Agent   │◄──►│    MCP      │◄──►│   Lua 组件层         │   │
│  │  (调试方)   │    │   协议      │    │  (用户代码)          │   │
│  └─────────────┘    └─────────────┘    └──────────┬──────────┘   │
│                                                  │               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────▼──────────┐   │
│  │   Go 运行时  │◄──►│  组件模型    │◄──►│   go-lua VM         │   │
│  │  (渲染引擎)  │    │  (Userdata) │    │                     │   │
│  └──────┬──────┘    └─────────────┘    └─────────────────────┘   │
│         │                                                           │
│         ▼                                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    渲染引擎 (Go)                              │   │
│  │  ┌─────────────┐   ┌──────────┐   ┌──────────────┐           │   │
│  │  │ Layout      │──►│ Virtual  │──►│ Diff Engine  │           │   │
│  │  │ Engine      │   │ Terminal │   │ (双缓冲差分)  │           │   │
│  │  │ (Flexbox)   │   │ Buffer   │   │              │           │   │
│  │  └─────────────┘   └──────────┘   └───────┬──────┘           │   │
│  └───────────────────────────────────────────┼──────────────────┘   │
│                                               ▼                      │
│                                    ┌──────────────────┐             │
│                                    │ Output Adapter   │             │
│                                    │ (ANSI / JSON)    │             │
│                                    └──────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. 组件模型

### 3.1 混合架构（Go Userdata + Lua Metatable）

| 层 | 技术 | 职责 |
|---|------|------|
| Go Userdata | Component 实例 | Props, State, Methods |
| Lua Metatable | __index 元方法 | 暴露方法给 Lua |
| Lua 表 | 用户组件定义 | init(), render() 函数 |

**优势**：
- Go 侧类型安全 + 方法分发
- Lua 侧灵活性 + 动态性
- 可逐步将性能关键路径迁移到 Go

### 3.2 生命周期

| React | 终端 | 实现 |
|-------|------|------|
| mount | init() | NewUserdata → 调用组件 init() |
| render | draw() | Component:render() → 终端 buffer |
| update | diff() + patch() | 比较 prev/next state，emit delta |
| unmount | cleanup() | __gc 元方法 或显式 unmount() |

### 3.3 Hooks 执行时机

**决策：选项 B（render 中调用）**

```go
func (c *Component) Render(vm *lua.State) int {
    // hooks 在 render 中调用
    c.mountPhase(vm)  // 首次调用时自动初始化 useState/useEffect
    // ... 组件逻辑
    return 1
}
```

**理由**：
- 实现简单，不需要维护"首次/更新"状态标志
- 与 React 16.8+ hooks 模式一致
- go-lua upvalue 在闭包创建时捕获，符合 hooks 语义

### 3.4 状态管理（Hooks）

```lua
-- useState
local count, setCount = lumina.useState(0)

-- useEffect  
lumina.useEffect(function()
  print("Count: " .. count)
  return function() end  -- cleanup
end, { count })

-- useMemo
local doubled = lumina.useMemo(function() return count * 2 end, {count})

-- useCallback
local handler = lumina.useCallback(function() ... end, deps)
```

### 3.5 Counter 组件示例

```lua
local Counter = lumina.defineComponent({
  name = "Counter",
  
  init = function(props)
    local count, setCount = lumina.useState(props.initial or 0)
    return { count = count, setCount = setCount }
  end,
  
  render = function(instance)
    return {
      type = "vbox",
      children = {
        { type = "text", content = "Count: " .. instance.count },
        { 
          type = "button", 
          label = "+1",
          onClick = function() instance.setCount(instance.count + 1) end
        }
      }
    }
  end
})
```

### 3.6 响应式更新

**方案 A：Lua setter → Go dirty flag**

```go
func (c *Component) SetState(L *lua.State) int {
    key := L.CheckString(1)
    value := L.GetAny(-1)
    c.State[key] = value
    c.Dirty = true
    c.App.ScheduleUpdate(c)
    return 0
}
```

---

## 4. 渲染引擎

### 4.1 双缓冲 + 差异刷新

```
组件树 (Lua) → VirtualTerminal (内存 2D Cell) → DiffEngine → 终端
```

- **避免闪烁**：先画后台缓冲区，完成后一次性显示
- **内存开销**：终端 ≤ 80x200 = 16,000 cells，可忽略

### 4.2 增量更新粒度

**决策：组件级 diffing**

```go
type DirtyTracker struct {
    mu    sync.Mutex
    dirty map[*Component]struct{}  // 组件级脏标记
}

// 状态变化 → 标记组件为 dirty → 只 diff 该组件子树
func (t *DirtyTracker) MarkDirty(c *Component) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.dirty[c] = struct{}{}
}
```

**理由**：
- Flexbox 边界天然与组件边界对齐
- 局部状态变化场景（TUI 中大多数是单组件交互）
- VirtualTerminal 按区域刷新比全局刷新更高效

### 4.3 布局系统（Flexbox 映射）

```go
type LayoutNode struct {
    Direction   Direction   // horizontal | vertical
    MainAlign  Align       // start | center | end | space-between
    CrossAlign Align       // start | center | end | stretch
    Gap        int         // 字符间距
    Flex       float64     // 弹性系数
    Children   []LayoutNode
}
```

- 两遍计算：自底向上算尺寸，自顶向下分配空间
- 支持 `flex: 1` 动态宽度

### 4.4 渲染原语

| 原语 | ANSI 实现 |
|------|----------|
| Cell | `\x1b[{y};{x}H{content}\x1b[{attr}m` |
| Move | `\x1b[{y};{x}H` |
| Clear | `\x1b[{n}J` |
| Style | `\x1b[{attrs}m` |

### 4.5 双输出抽象层

```
组件树 → VDOM → Layout → Style → Frame → OutputAdapter → 终端
                                              ↓
                              ┌───────────────┴───────────────┐
                              ↓                               ↓
                        ANSIAdapter                     JSONAdapter
                              ↓                               ↓
                      终端写入 (ANSI)                  MCP 传输 (JSON)
```

**Frame 结构（统一中间表示）**：

```go
type Frame struct {
    Cells      [][]Cell  // 虚拟终端 buffer
    DirtyRects []Rect    // 脏区（优化：不输出全量）
    Timestamp  int64
    Metadata   map[string]any  // 组件树快照等
}

type Cell struct {
    Char       rune
    Foreground string  // "#rrggbb" 格式
    Background string
    Bold, Dim  bool
    Underline  bool
}
```

**OutputAdapter 接口**：

```go
type OutputAdapter interface {
    Write(frame *Frame) error
    Flush() error
    Close() error
    Mode() OutputMode
}

type OutputMode int
const (
    ModeANSI OutputMode = iota
    ModeJSON
)
```

**ANSI 适配器**：

```go
type ANSIAdapter struct {
    writer io.Writer
    buf    bytes.Buffer
}

func (a *ANSIAdapter) Write(frame *Frame) error {
    for _, rect := range frame.DirtyRects {
        for y := rect.Y; y < rect.Y+rect.H; y++ {
            for x := rect.X; x < rect.X+rect.W; x++ {
                a.writeCell(x, y, frame.Cells[y][x])
            }
        }
    }
    return a.writer.Write(a.buf.Bytes())
}
```

**JSON 适配器**：

```go
type JSONAdapter struct {
    encoder *json.Encoder
}

type JSONFrame struct {
    Type      string  `json:"type"`
    Timestamp int64   `json:"timestamp"`
    Patches   []Patch `json:"patches,omitempty"`  // diff 模式
    Cells     [][]Cell `json:"cells,omitempty"`   // 全量模式（调试用）
}

type Patch struct {
    Op string `json:"op"`  // "set"
    X  int    `json:"x"`
    Y  int    `json:"y"`
    C  Cell   `json:"c"`
}
```

**模式切换**：

```lua
-- 运行时切换
lumina.setOutputMode("ansi")  -- 人类可读
lumina.setOutputMode("json")  -- AI 可解析

-- 启动时选择（优先级：参数 > 环境变量 > 默认值）
LUMINA_OUTPUT=ansi  # 或 json
```

**ANSI 样式码映射**：

| 属性 | ANSI 代码 |
|------|----------|
| Bold | 1 |
| Dim | 2 |
| Underline | 4 |
| fg color | 38;2;r;g;b |
| bg color | 48;2;r;g;b |

---

## 5. MCP 协议

### 5.1 协议决策

**决策：Protobuf**

**理由**：
- AI Agent 是机器，类型安全 > 人类可读
- 跨语言代码生成（Go + Lua）
- 高频更新场景性能重要
- Schema 防止 API 漂移

**调试折中**：调试时用 JSON 序列化 protobuf 消息，但协议本身是 Protobuf。

### 5.2 Schema 组织

```
mcp/
├── proto/
│   ├── v1/
│   │   ├── lumina.proto      # 核心消息
│   │   ├── component.proto   # 组件树
│   │   ├── event.proto       # 输入事件
│   │   └── render.proto      # 渲染命令
│   └── buf.yaml
└── generated/                 # 生成的 Go/Lua 代码
```

### 5.3 核心命令

| 命令 | 用途 |
|------|------|
| `lumina/create` | 创建组件 |
| `lumina/update` | 更新组件属性 |
| `lumina/delete` | 删除组件 |
| `lumina/query` | 查询状态 |
| `lumina/batch` | 批量操作 |

### 5.4 调试命令

| 命令 | 用途 |
|------|------|
| `lumina/debug/tree` | 组件树结构 |
| `lumina/debug/inspect` | 组件详情 |
| `lumina/debug/vm` | Lua VM 状态 |
| `lumina/debug/patch` | 实时修改 |
| `lumina/debug/history` | 状态历史 |
| `lumina/hotreload/reload` | 热重载 |

---

## 6. go-lua 集成

### 6.1 已验证模式

| 模式 | 证据 |
|------|------|
| Userdata + Metatable | example_test.go, iolib.go |
| SetFuncs 函数注册 | baselib.go |
| Upvalue 闭包状态 | example_test.go |
| Coroutine 事件循环 | example_test.go |
| Registry Ref/Unref | example_test.go |

### 6.2 关键集成点

```go
// 创建组件 userdata
L.NewUserdata(0, 0)
L.SetUserdataValue(-1, &Component{Type: "Button"})
L.NewMetatable("Component")
L.SetMetatable(-2)

// 注册 lumina 函数
L.SetFuncs(map[string]lua.Function{
    "useState": useStateFn,
    "useEffect": useEffectFn,
}, 0)
```

---

## 7. 可持续性设计

### 7.1 版本演进策略

```
v1.0（基础稳固）:
├─ 组件模型（Userdata + Metatable）
├─ 响应式系统（dirty flag + batch）
├─ 渲染引擎（双缓冲 + diff）
├─ 基础 hooks（useState, useEffect）
├─ MCP 核心命令
├─ Protobuf 协议
└─ 测试覆盖 >80%

v1.x（增量完善）:
├─ 更多 hooks（useMemo, useCallback）
├─ 官方组件库
└─ 性能优化

v2.0（架构演进）:
├─ CSS 布局系统（完整）
├─ 动画系统
└─ 插件系统（如需要）
```

### 7.2 API 稳定性

```go
// API 级别标识
const (
    APILevelPublic       = 0  // 稳定版，遵守 semver
    APILevelInternal     = 1  // 可能有变更
    APILevelExperimental = 2  // 随时可能变更
)

// 渐进式弃用策略
// 1. v1.x: 标记弃用 + 警告日志
// 2. v2.0: 移除旧 API + 迁移指南
```

### 7.3 扩展点预留

```go
// 组件注册表
type ComponentRegistry interface {
    Register(name string, factory ComponentFactory)
    Get(name string) (ComponentFactory, bool)
}

// Hook 扩展点
type HookExtension interface {
    Name() string
    Execute(ctx *HookContext) error
}

// 渲染器插件
type RendererAdapter interface {
    Init(term Terminal) error
    Render(frame *Frame) error
}
```

### 7.4 React 成功的关键借鉴

| 原则 | Lumina 实现 |
|------|-------------|
| 虚拟 DOM | VirtualTerminal 解耦终端 |
| 不可变数据 | 状态浅拷贝 + deepEqual |
| 单向数据流 | 事件 → 状态 → 渲染 |
| Hooks 扩展性 | 自定义 hooks 机制 |

---

## 8. 测试策略

```
┌─────────────────────┐
│     E2E Tests       │  ← 关键路径验证
├─────────────────────┤
│   Integration Tests │  ← Go ↔ Lua 交互
├─────────────────────┤
│    Unit Tests       │  ← 独立组件 >80%
└─────────────────────┘
```

---

## 9. 已决策问题清单

| 决策点 | 决策 | 理由 |
|--------|------|------|
| 组件实例 | Go Userdata + Lua Metatable | 类型安全 + 动态性 |
| hooks 时机 | render 中调用 | 实现简单，与 React 16.8+ 一致 |
| 响应式更新 | 方案 A（Lua setter → Go dirty flag） | 显式更新，React 风格 |
| 更新粒度 | 组件级 diffing | Flexbox 边界清晰，局部刷新高效 |
| MCP 协议 | Protobuf + JSON 调试 | AI 是机器，类型安全 > 可读 |
| CSS 布局 | v2.0 实现，v1.0 预留接口 | 避免过度设计 |

---

## 10. MVP 范围

### Phase 1: 最小可行产品

- [ ] go-lua 集成（复用 VM）
- [ ] Protobuf 协议定义 + 代码生成
- [ ] 基础组件：Box, Text, Button
- [ ] 简单布局：vbox, hbox
- [ ] MCP 核心命令（create/update/delete）
- [ ] 组件级 diffing + 双缓冲渲染
- [ ] hooks：useState, useEffect
- [ ] ANSI 输出（调试用 JSON 序列化）

### Phase 2: 增强

- [ ] 更多 hooks（useMemo, useCallback）
- [ ] 热重载
- [ ] MCP 调试命令（debug/tree, debug/inspect）
- [ ] 官方组件库

### Phase 3: 完善

- [ ] 完整 CSS 布局系统
- [ ] 虚拟列表
- [ ] 性能分析
- [ ] 插件系统

---

## 14. 组件作用域 & 隔离模型

### 14.1 作用域模型

**推荐：模块化作用域（基于 Lua require）**

```lua
-- 推荐模式
local Button = require("lumina/components/button")
local Dialog = require("my-dialog")

-- 不污染全局命名空间
```

### 14.2 组件 ID 系统

**分层 ID 格式**：
```
{scope}:{type}:{version}:{instance_id}
Example: "main:Button:1.2.0:abc123"
```

**Go struct 扩展**：
```go
type Component struct {
    // 已有字段
    Type   string
    Props  map[string]interface{}
    State  map[string]interface{}
    Dirty  bool
    
    // 新增字段
    ID      ComponentID  // 分层唯一 ID
    Version string       // 组件版本（热重载用）
    Scope   string       // "main" | "modal" | "plugin:xyz"
}
```

### 14.3 状态隔离

| 级别 | 实现 | 状态 |
|------|------|------|
| 实例隔离 | Hooks + upvalues | ✅ 已实现 |
| 模块隔离 | Lua `require()` | ✅ 自然支持 |
| 作用域隔离 | `Component.Scope` | 新增 |
| 进程隔离 | 独立 VM | 未来 |

```lua
-- 每个实例独立状态（已有）
local Counter = lumina.defineComponent({
    init = function()
        local count, setCount = lumina.useState(0)  -- 实例独立
        return { count = count, setCount = setCount }
    end
})
```

### 14.4 Context 系统（React Context 模式）

```lua
-- Provider
local ThemeProvider = lumina.defineComponent({
    providers = { { name = "theme", value = "dark" } }
})

-- Consumer
local theme = lumina.useContext("theme")

-- 嵌套 Provider
local App = lumina.defineComponent({
    providers = {
        { name = "theme", value = "dark" },
        { name = "locale", value = "zh-CN" }
    },
    children = {...}
})
```

### 14.5 热重载机制

**原理：实例存活 + 代码重载 + 状态快照**

```
1. 用户修改 Button.lua
2. Lumina 重新编译 Lua 源码
3. 组件实例（Go struct）保持不变
4. 状态通过快照恢复
5. 版本不匹配时触发迁移回调
```

**集成点**：
- `DirtyTracker` 增加版本跟踪
- 组件注册表映射 types → factories
- `lumina/hotreload/reload` 命令使用此系统

### 14.6 状态持久化

```lua
local counter = lumina.defineComponent({
    persist = true,  -- 状态持久化
    init = function()
        local saved = lumina.loadState("counter") or { count = 0 }
        return saved
    end
})
```

### 14.7 与现有架构集成

**无需破坏性改动**：
- VirtualTerminal buffer ❌ 不变
- Diff Engine ❌ 不变
- Event delegation ❌ 不变
- Output adapters ❌ 不变

**仅扩展**：
- `Component` struct 增加 3 字段
- `DirtyTracker` 增加版本跟踪
- 新增 `useContext` hook

---

## 附录：相关文档

- go-lua 项目：/home/ubuntu/workspace/go-lua
- go-lua 设计文档：/home/ubuntu/workspace/go-lua/DESIGN.md

---

# Phase 1 完成状态 (v0.1)

## ✅ 已实现功能

| 功能 | 状态 | 文件 |
|------|------|------|
| go-lua 集成 | ✅ | `pkg/lumina/lumina.go` |
| 组件模型 | ✅ | `lumina.go` (defineComponent, hooks) |
| OutputAdapter | ✅ | `pkg/lumina/output.go` |
| ANSI 渲染 | ✅ | `pkg/lumina/ansi_adapter.go` |
| VNode → Frame | ✅ | `pkg/lumina/renderer.go` |
| Proto 生成 | ✅ | `pkg/lumina/v1/lumina.pb.go` |
| Counter 示例 | ✅ | `examples/counter.lua` |
| E2E 测试 | ✅ | `pkg/lumina/e2e_test.go` |
| CI/CD | ✅ | `.github/workflows/ci.yml` |

## 已知限制

| 限制 | 说明 | 未来方向 |
|------|------|----------|
| 无热重载 | 需要手动重启 | Week 5+ |
| 无 Diff Engine | 全量重绘 | 差分渲染 |
| 无 Event 系统 | 按钮点击未实现 | 事件委托 |
| 单 VM | 状态隔离在 VM 内 | 进程隔离 |
| 固定 terminal size | 80x24 | 动态检测 |

## 项目结构

```
lumina/
├── pkg/lumina/
│   ├── lumina.go           # 核心: Open, render, hooks
│   ├── lumina_test.go      # 单元测试
│   ├── e2e_test.go         # E2E 测试
│   ├── output.go          # OutputAdapter 接口
│   ├── ansi_adapter.go     # ANSI 终端输出
│   ├── renderer.go         # VNode → Frame
│   └── v1/lumina.pb.go     # Proto 生成代码
├── examples/
│   └── counter.lua        # Counter 示例
├── proto/v1/
│   └── lumina.proto       # Proto 定义
├── .github/workflows/
│   └── ci.yml            # CI/CD 流水线
├── DESIGN.md              # 架构设计
├── DEVELOPMENT.md         # 开发文档
└── go.mod
```