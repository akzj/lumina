# Lumina 架构设计文档

> **版本**: v2.0  
> **核心理念**: Go 渲染引擎 + Lua 声明式 UI；持久化节点树、增量布局、脏区绘制  
> **一句话**: Lua 描述 UI → Go 协调/布局/绘制 → 终端输出

---

## 1. 设计目标

| 目标 | 实现方式 |
|------|----------|
| **O(k) 渲染** | 只重新渲染脏组件、只重算脏布局、只重绘脏节点（k = 变化量） |
| **正确性** | 协调器保证节点树与 Lua 描述一致；脏标记保证不遗漏重绘 |
| **可测试性** | `TestAdapter` + 无终端运行；所有渲染逻辑纯函数 |
| **简洁** | 无 Virtual DOM 中间层；Lua 表 → Descriptor → 直接 patch Node 树 |

---

## 2. 核心架构

```
┌─────────────────────────────────────────────────────────┐
│                     Lua 用户代码                         │
│  createComponent / defineComponent / createElement       │
│  useState / onKeyDown / onClick / ...                    │
└──────────────────────┬──────────────────────────────────┘
                       │ Lua 表（Descriptor）
                       ▼
┌─────────────────────────────────────────────────────────┐
│                    Engine (Go)                           │
│                                                         │
│  ┌─────────┐  ┌────────────┐  ┌────────┐  ┌─────────┐  │
│  │ Render  │→│ Reconcile  │→│ Layout │→│  Paint  │  │
│  │ (Lua→   │  │ (Diff+Patch│  │(Flex   │  │(Dirty   │  │
│  │  Desc)  │  │  Node Tree)│  │ incr.) │  │ cells)  │  │
│  └─────────┘  └────────────┘  └────────┘  └────┬────┘  │
│                                                 │       │
│                                          CellBuffer     │
└─────────────────────────────────────────────┬───────────┘
                                              │ ToBuffer()
                                              ▼
┌─────────────────────────────────────────────────────────┐
│                  Output Adapter                          │
│            WriteDirty(buf, dirtyRects)                   │
│            ANSI / TestAdapter / ...                      │
└─────────────────────────────────────────────────────────┘
```

---

## 3. 渲染流程（一帧）

`Engine.RenderDirty()` 是每帧的核心函数，流程如下：

```
1. ResetStats()           — 重置 CellBuffer 统计
2. renderInOrder()        — 渲染所有 Dirty 组件（父→子顺序）
   ├─ 调用 Lua renderFn(props)
   ├─ readDescriptor()    — Lua 表 → Descriptor 结构体
   ├─ Reconcile()         — Descriptor vs Node 树，就地 patch
   └─ reconcileChildComponents() — 发现/创建子组件
3. graftChildComponents() — 将子组件 RootNode 嫁接到父树
4. 提前退出检查           — 无渲染 + 无脏节点 → 跳过布局/绘制
5. Layout
   ├─ LayoutFull()        — 根节点 LayoutDirty → 全量布局
   └─ LayoutIncremental() — 否则只走脏子树
6. PaintDirty()           — 只重绘 PaintDirty 节点到 CellBuffer
7. 记录统计               — 写入/清除的 cell 数、脏区面积
```

### 完整帧（初始挂载）

`Engine.RenderAll()` 用于首次渲染：

```
1. 标记所有组件 Dirty
2. renderInOrder()        — 渲染所有组件
3. graftChildComponents() — 嫁接子组件
4. LayoutFull()           — 全量布局
5. PaintFull()            — 清空 buffer + 全量绘制
6. FocusAutoFocus()       — 自动聚焦 autoFocus 节点
```

---

## 4. 组件系统

### Component 结构

```go
type Component struct {
    ID       string              // 唯一标识（如 "todo-app"）
    Type     string              // 工厂名（如 "Cell"）
    State    map[string]any      // 组件状态（useState 管理）
    Props    map[string]any      // 父组件传入的属性
    RenderFn LuaRef              // Lua 渲染函数的注册表引用
    RootNode *Node               // 渲染输出的节点子树
    Dirty    bool                // true → 需要重新渲染
    Parent   *Component          // 父组件
    Children []*Component        // 子组件列表
    ChildMap map[string]*Component // type:key → 子组件（快速查找）
}
```

### 组件生命周期

```
创建: createComponent/defineComponent → 注册到 Engine
  ↓
首次渲染: Dirty=true → renderComponent() → 创建 RootNode
  ↓
状态更新: useState setter → SetState() → Dirty=true
  ↓
增量渲染: RenderDirty() → 只渲染 Dirty 组件 → Reconcile 更新 RootNode
```

### useState 工作原理

```lua
local count, setCount = lumina.useState("count", 0)
```

1. `useState("count", 0)` → 查找当前组件的 `State["count"]`
2. 不存在 → 初始化为 `0`
3. 返回 `(当前值, setter函数)`
4. `setCount(1)` → `Component.SetState("count", 1)` → `Dirty = true`
5. 下一帧 `RenderDirty()` 重新调用 renderFn

---

## 5. 树结构

### Node 树

```go
type Node struct {
    Type     string   // "box", "vbox", "hbox", "text", "input", "textarea", "component"
    ID       string   // 用于协调匹配
    Key      string   // 用于列表协调
    Children []*Node
    Parent   *Node
    
    // 布局（缓存）
    X, Y, W, H  int
    LayoutDirty  bool
    
    // 绘制
    Content     string
    Style       Style
    PaintDirty  bool
    
    // 事件处理器（Lua 注册表引用）
    OnClick, OnMouseEnter, OnMouseLeave LuaRef
    OnKeyDown, OnChange, OnScroll       LuaRef
    
    // 组件关联
    Component     *Component  // 组件根节点指向所属组件
    ComponentType string      // type="component" 节点的工厂名
}
```

### 嫁接机制 (Graft)

子组件的 `RootNode` 需要嫁接到父组件的节点树中：

```
父组件 RootNode:
  box
  ├── text "Header"
  ├── component (type="Cell", key="0,0")    ← 占位节点
  │   └── [Cell 组件的 RootNode]            ← 嫁接到这里
  └── component (type="Cell", key="0,1")
      └── [Cell 组件的 RootNode]
```

**流程**:
1. 父组件渲染 → 产生 `type="component"` 占位节点
2. `reconcileChildComponents()` → 为占位节点创建/查找子组件
3. 子组件渲染 → 产生自己的 `RootNode`
4. `graftChildComponents()` → 将 `RootNode` 设为占位节点的子节点
5. 布局/绘制自然穿透到子组件

---

## 6. 协调算法 (Reconciliation)

`Reconcile(node *Node, desc Descriptor) bool`

**原则**: 就地修改，最小化脏标记。

```
1. 比较 Content → 不同 → 更新 + PaintDirty
2. 比较 Style → 不同 → 分析变化类型：
   ├─ 布局属性变化（width/height/flex/padding/margin/...） → LayoutDirty
   └─ 仅视觉属性变化（foreground/background/bold/...） → PaintDirty
3. 更新事件处理器引用
4. 协调子节点:
   ├─ 快速路径: 长度相同 + key 全匹配 → 逐个 Reconcile
   └─ 慢速路径: 按 type:key 建立映射 → 匹配复用/新建/删除
```

### 子节点协调

```
旧: [A, B, C]    新: [A, C, D]
  ↓ 按 type:key 匹配
  A → 复用，Reconcile
  C → 复用，Reconcile
  D → 新建
  B → 删除（parent.PaintDirty = true，清除残影）
```

---

## 7. 布局系统

### Flex 布局

支持两种容器方向：
- **vbox** — 垂直堆叠（默认）
- **hbox** — 水平排列

```
┌── vbox ──────────────┐    ┌── hbox ──────────────┐
│ ┌──────────────────┐ │    │ ┌─────┐┌─────┐┌────┐│
│ │   child 1        │ │    │ │  1  ││  2  ││ 3  ││
│ ├──────────────────┤ │    │ │     ││     ││    ││
│ │   child 2        │ │    │ └─────┘└─────┘└────┘│
│ ├──────────────────┤ │    └──────────────────────┘
│ │   child 3        │ │
│ └──────────────────┘ │
└──────────────────────┘
```

### 尺寸计算

```
固定尺寸:  width/height > 0 → 使用固定值（受 min/max 约束）
Flex 分配: flex > 0 → 按比例分配剩余空间
隐式规则:  text/input/textarea → 固定 1 行高
           容器无 flex 无 height → 隐式 flex=1
```

### 对齐

- **justify** (主轴): `start`, `center`, `end`, `space-between`, `space-around`
- **align** (交叉轴): `stretch`(默认), `start`, `center`, `end`

### 定位

- `position: "relative"` — 相对偏移（不脱离文档流）
- `position: "absolute"` — 相对父内容区定位
- `position: "fixed"` — 相对屏幕定位

### 增量布局

```
LayoutIncremental(root):
  遍历树 → 遇到 LayoutDirty 节点 → 用缓存的 (X,Y,W,H) 重算子树
  非脏节点 → 跳过（保留缓存布局）
```

**脏传播**: `MarkLayoutDirty()` 向上传播到最近的固定尺寸祖先。

---

## 8. 绘制系统

### CellBuffer

```go
type CellBuffer struct {
    cells  []Cell    // 扁平数组: cells[y*width + x]
    width  int
    height int
    // 每帧统计
    writeCount, clearCount int
    dirtyMinX, dirtyMinY, dirtyMaxX, dirtyMaxY int  // 脏区包围盒
}

type Cell struct {
    Ch        rune
    FG, BG    string   // 颜色字符串如 "#FF0000"
    Bold, Dim, Underline bool
    Wide      bool     // CJK 宽字符右半部
}
```

### 脏区绘制

```
PaintDirty(buf, root):
  遍历树:
    PaintDirty=true → ClearRect(旧区域) + paintNode(完整子树) + 清除标记
    PaintDirty=false → 递归检查子节点
```

**关键**: 父节点重绘时，先清空自己的区域，再绘制自己和所有子节点。这保证了：
- 子节点被删除时不留残影
- 节点移动时旧位置被清理

### 绘制优先级

```
box/vbox/hbox:
  1. 填充背景色
  2. 绘制边框
  3. 递归绘制子节点

text:
  1. 填充背景色（如果有）
  2. 逐字符写入（无背景色时继承父背景）

input/textarea:
  - 有内容 → 同 text
  - 无内容 → 绘制 placeholder（dim 样式）
```

---

## 9. 事件系统

### 事件分发流程

```
终端输入 → App.HandleEvent()
  ├─ click     → Engine.HandleClick(x, y)
  │              ├─ HitTest → 找到最深层节点
  │              ├─ 焦点管理（input/textarea）
  │              └─ 向上冒泡找 onClick handler → callLuaRef
  │
  ├─ mousemove → Engine.HandleMouseMove(x, y)
  │              ├─ HitTest → 找到当前节点
  │              ├─ 与上次 hoveredNode 比较
  │              ├─ 不同 → 触发 onMouseLeave(旧) + onMouseEnter(新)
  │              └─ 相同 → 无操作
  │
  ├─ keydown   → Engine.HandleKeyDown(key)
  │              ├─ Tab → FocusNext()
  │              ├─ 有焦点 input → HandleInputKeyDown()
  │              └─ 否则 → DFS 找 onKeyDown handler
  │
  └─ scroll    → Engine.HandleScroll(x, y, delta)
                 └─ HitTest + 冒泡找 onScroll handler
```

### HitTest

```go
func HitTest(root *Node, x, y int) *Node {
    // 点不在节点范围内 → nil
    // 逆序检查子节点（后绘制的在上层）
    // 子节点命中 → 返回子节点
    // 无子节点命中 → 返回自身
}
```

### 事件冒泡

点击/滚动事件从最深层节点向上冒泡，直到找到对应的事件处理器。

---

## 10. 输入系统

### 焦点管理

- `Tab` → 循环聚焦下一个 input/textarea
- 点击 input/textarea → 聚焦
- `autoFocus=true` → 初始渲染后自动聚焦

### 文本编辑

| 按键 | input | textarea |
|------|-------|----------|
| 字符 | 在光标处插入 | 在光标处插入 |
| Backspace | 删除光标前字符 | 删除光标前字符 |
| Enter | 触发 onChange | 插入换行 |
| ←/→ | 移动光标 | 移动光标 |
| ↑/↓ | — | 上下移动行 |

编辑后自动调用 `onChange` 回调并标记所属组件 Dirty。

---

## 11. 性能策略

### O(k) 渲染

| 阶段 | 优化 |
|------|------|
| 渲染 | 只调用 Dirty 组件的 Lua renderFn |
| 协调 | 快速路径：子节点数量+key 不变 → 逐个 patch |
| 布局 | 只重算 LayoutDirty 子树 |
| 绘制 | 只重绘 PaintDirty 节点 |
| 输出 | 只输出 DirtyRect 区域 |

### 空闲检测

```
rendered == 0 && !hasAnyDirty(root)
  → 跳过布局和绘制
  → 记录 0 cells painted
```

### GC 控制

渲染期间暂停 Lua GC，渲染后执行一步增量 GC：
```go
L.SetGCStopped(true)
// ... 渲染 ...
L.SetGCStopped(false)
L.GCStepAPI()
```

### 状态变化检测

`SetState` 使用 `reflect.DeepEqual` 避免无变化时标记 Dirty：
```go
func (c *Component) SetState(key string, value any) {
    if reflect.DeepEqual(c.State[key], value) {
        return // 无变化，不标记 Dirty
    }
    c.State[key] = value
    c.Dirty = true
}
```

---

## 12. 模块结构

```
pkg/lumina/v2/
├── render/           — 核心渲染引擎
│   ├── engine.go     — Engine 主体 + Lua API 注册
│   ├── node.go       — Node, Component, Style, Descriptor 类型
│   ├── reconciler.go — 协调算法
│   ├── layout.go     — Flex 布局（vbox/hbox）
│   ├── painter.go    — 绘制（full/dirty/clipped）
│   ├── cellbuffer.go — CellBuffer 2D 网格
│   ├── events.go     — HitTest + 事件分发
│   └── input.go      — 输入编辑 + 焦点管理
├── app.go            — App 组合根（集成所有模块）
├── app_run.go        — 事件循环 + 热加载
├── app_lua.go        — App 级 Lua API（quit, timer）
├── app_timer.go      — setInterval/setTimeout 管理
├── buffer/           — 通用 Buffer 类型（输出适配器接口）
├── output/           — 输出适配器（ANSI, TestAdapter）
├── event/            — 事件类型定义
├── perf/             — 性能追踪器
├── devtools/         — 开发者工具面板
├── animation/        — 动画管理器
├── router/           — 路由
├── hotreload/        — 文件监控热加载
└── store/            — 状态管理
```
