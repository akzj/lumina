# V2 渲染引擎深度解析

> 本文档详细描述 Lumina V2 渲染引擎的内部实现。  
> 适合想要理解或修改引擎的开发者。

---

## 1. 总览

V2 引擎的核心理念是 **持久化节点树 + 脏标记驱动**：

- 节点树创建后**永不丢弃**，只在协调时就地修改
- 三种脏标记独立工作：`Component.Dirty`、`Node.LayoutDirty`、`Node.PaintDirty`
- 每一帧只处理脏的部分，实现 O(k) 复杂度

```
传统 React 模型:                    V2 模型:
render() → new VNode tree           render() → Descriptor
diff(old, new) → patches            Reconcile(node, desc) → patch in-place
apply patches → DOM mutations       node 就是 "DOM"，直接修改
```

---

## 2. 数据类型

### Descriptor

Lua 渲染函数的输出，由 `readDescriptor()` 从 Lua 表转换而来：

```go
type Descriptor struct {
    Type           string       // "box", "vbox", "hbox", "text", ...
    ID, Key        string       // 用于协调匹配
    Content        string       // 文本内容
    Placeholder    string       // input/textarea 占位文本
    AutoFocus      bool
    ScrollY        int
    Style          Style
    Children       []Descriptor
    
    // 事件处理器（Lua 注册表引用）
    OnClick, OnMouseEnter, OnMouseLeave LuaRef
    OnKeyDown, OnChange, OnScroll       LuaRef
    
    // 子组件
    ComponentType  string       // 工厂名（如 "Cell"）
    ComponentProps map[string]any
}
```

**Descriptor 是临时的** — 只存在于一帧的渲染过程中，协调完成后即可丢弃。

### Node

持久化的 UI 节点，构成渲染树：

```go
type Node struct {
    Type     string   // 元素类型
    ID, Key  string   // 身份标识
    Parent   *Node
    Children []*Node
    
    // 布局缓存
    X, Y, W, H  int
    LayoutDirty  bool
    
    // 绘制状态
    Content     string
    Style       Style
    PaintDirty  bool
    
    // 事件
    OnClick, OnMouseEnter, OnMouseLeave LuaRef
    OnKeyDown, OnChange, OnScroll       LuaRef
    
    // 组件关联
    Component     *Component
    ComponentType string
}
```

**Node 是持久的** — 创建后只通过协调修改，不会被重建。

### Component

有状态的 UI 单元：

```go
type Component struct {
    ID       string
    Type     string              // 工厂名
    State    map[string]any      // useState 管理的状态
    Props    map[string]any
    RenderFn LuaRef              // Lua 渲染函数
    RootNode *Node               // 渲染输出的节点子树
    Dirty    bool                // 需要重新渲染
    Parent   *Component
    Children []*Component
    ChildMap map[string]*Component
}
```

---

## 3. 渲染流程详解

### 3.1 renderInOrder()

按依赖顺序渲染脏组件：**父组件先于子组件**。

```
1. 如果 root.Dirty → renderComponent(root)
2. 遍历所有其他组件 → 如果 Dirty 且非 root → renderComponent(comp)
```

**为什么父先于子？** 父组件渲染时会产生 `type="component"` 占位节点，
`reconcileChildComponents()` 需要这些占位节点来发现/创建子组件。

### 3.2 renderComponent(comp)

单个组件的渲染流程：

```
1. 暂停 Lua GC
2. 设置 currentComp = comp（useState 需要知道当前组件）
3. 从 Lua 注册表获取 renderFn
4. 调用 renderFn(props) → 得到 Lua 表
5. readDescriptor() → 将 Lua 表转为 Descriptor
6. 如果 RootNode == nil（首次挂载）:
   └─ createNodeFromDesc(desc) → 创建整棵节点树
7. 如果 RootNode != nil（更新）:
   └─ Reconcile(RootNode, desc) → 就地 patch
8. reconcileChildComponents() → 处理子组件
9. comp.Dirty = false, comp.Mounted = true
10. 恢复 Lua GC + 执行一步增量 GC
```

### 3.3 readDescriptor()

将 Lua 表递归转换为 Descriptor 结构体：

```
Lua 表:
{
    type = "box",
    id = "main",
    style = {background = "#1E1E2E", flex = 1},
    onClick = function(e) ... end,
    children = {
        {type = "text", content = "Hello"},
    },
}

→ Descriptor{
    Type: "box", ID: "main",
    Style: Style{Background: "#1E1E2E", Flex: 1},
    OnClick: <LuaRef>,
    Children: []Descriptor{{Type: "text", Content: "Hello"}},
}
```

**特殊处理**:
- `_factoryName` 字段 → 转为 `type="component"` + `ComponentType`
- `_props` 字段 → 读取为 `ComponentProps`
- `value` 字段 → 作为 `Content` 的备选（input/textarea）
- 字符串子节点 → 自动转为 `{type: "text", content: s}`

### 3.4 reconcileChildComponents()

遍历节点树，为 `type="component"` 节点创建/查找子组件：

```
walkTree(node):
  if node.Type == "component" && node.ComponentType != "":
    child = parent.FindChild(factoryName, lookupKey)
    if child == nil:
      // 新子组件
      child = NewComponent(...)
      child.RenderFn = factories[factoryName]
      child.Dirty = true
      parent.AddChild(child)
      engine.components[childID] = child
    node.Component = child
  else:
    // 递归检查子节点
    for each child in node.Children:
      walkTree(child)
```

### 3.5 graftChildComponents()

将子组件的 RootNode 嫁接到父树的占位节点：

```
graftWalk(node):
  for each child in node.Children:
    if child.Type == "component" && child.Component != nil:
      comp = child.Component
      if comp.RootNode != nil:
        alreadyGrafted = (child.Children[0] == comp.RootNode)
        if !alreadyGrafted:
          child.Children = [comp.RootNode]
          comp.RootNode.Parent = child
          child.LayoutDirty = true
          child.PaintDirty = true
    graftWalk(child)  // 递归（嵌套组件）
```

**嫁接前后**:
```
嫁接前:                          嫁接后:
  box                              box
  ├── text "Header"                ├── text "Header"
  └── component(Cell)              └── component(Cell)
      (无子节点)                       └── box (Cell 的 RootNode)
                                           └── text "·"
```

---

## 4. 协调算法

### 4.1 Reconcile(node, desc)

```go
func Reconcile(node *Node, desc Descriptor) bool {
    // 1. 内容变化 → PaintDirty
    // 2. 样式变化 → 分析是布局属性还是视觉属性
    //    布局属性变化 → LayoutDirty（传播到祖先）
    //    视觉属性变化 → PaintDirty
    // 3. 更新事件处理器引用
    // 4. 协调子节点（component 类型跳过，子节点由 graft 管理）
}
```

### 4.2 reconcileChildren(parent, descs)

子节点协调有两条路径：

**快速路径**（最常见）:
```
条件: len(old) == len(new) && 所有 key 按序匹配
操作: 逐个 Reconcile(old[i], new[i])
```

**慢速路径**:
```
1. 建立 oldByKey 映射: "type:key" → index
2. 遍历新 Descriptor:
   ├─ key 匹配 + type 匹配 → 复用旧节点，Reconcile
   └─ 无匹配 → createNodeFromDesc() 创建新节点
3. 清理未使用的旧节点（parent.PaintDirty = true）
4. 检测重排序
5. 如果有变化 → parent.LayoutDirty = true
```

### 4.3 样式协调

```go
func reconcileStyle(node *Node, newStyle Style) bool {
    if old == newStyle { return false }  // 快速相等检查
    
    // 布局属性: width, height, flex, padding, margin, gap,
    //          min/max, justify, align, border, overflow,
    //          position, top/left/right/bottom
    layoutChanged = ... // 逐字段比较
    
    node.Style = newStyle
    if layoutChanged { node.MarkLayoutDirty() }
    node.PaintDirty = true
    return true
}
```

---

## 5. 布局系统

### 5.1 computeFlex(node, x, y, w, h)

核心布局函数，递归计算每个节点的位置和尺寸：

```
1. 应用 margin（缩小可用空间）
2. 应用固定尺寸（width/height）+ min/max 约束
3. 检测位置/尺寸变化 → PaintDirty
4. 计算内容区（减去 border + padding）
5. 如果 overflow=scroll → 预留 1 列给滚动条
6. 根据 node.Type 分发:
   ├─ "fragment"  → layoutVBox (透明容器)
   ├─ "component" → 直接传递给子节点
   ├─ "text"      → layoutText (计算换行高度)
   ├─ "vbox"      → layoutVBox
   ├─ "hbox"      → layoutHBox
   └─ 默认(box)   → layoutVBox
7. 处理 absolute/fixed 定位的子节点
```

### 5.2 layoutVBox — 垂直布局

```
1. 统计流式子节点（排除 absolute/fixed）
2. 计算 gap 总量
3. 分类每个子节点:
   ├─ 有固定 height → fixedH
   ├─ 有 flex → flexGrow
   ├─ text/input/textarea 无 flex 无 height → fixedH=1
   └─ 容器无 flex 无 height → 隐式 flex=1
4. 分配剩余空间给 flex 子节点（按比例）
5. 根据 justify 计算起始 Y:
   ├─ "start"         → 顶部
   ├─ "center"        → 居中
   ├─ "end"           → 底部
   ├─ "space-between" → 均匀分布
   └─ "space-around"  → 两端留半间距
6. 逐个定位子节点（考虑 align 交叉轴对齐）
7. 应用 relative 偏移
```

### 5.3 增量布局

```go
func LayoutIncremental(root *Node) {
    layoutDirtyWalk(root)
}

func layoutDirtyWalk(node *Node) {
    if !node.LayoutDirty {
        // 缓存有效，只检查子节点
        for _, child := range node.Children {
            layoutDirtyWalk(child)
        }
        return
    }
    // 用缓存的 (X,Y,W,H) 重算子树
    normalizeSpacing(node)
    computeFlex(node, node.X, node.Y, node.W, node.H)
    node.LayoutDirty = false
    clearLayoutDirtyBelow(node)
}
```

### 5.4 脏传播

```go
func (n *Node) MarkLayoutDirty() {
    n.LayoutDirty = true
    p := n.Parent
    for p != nil {
        if p.LayoutDirty { break }
        // 固定尺寸节点是布局边界
        if p.Style.Width > 0 && p.Style.Height > 0 {
            p.LayoutDirty = true
            break
        }
        p.LayoutDirty = true
        p = p.Parent
    }
}
```

**布局边界**: 有固定 width+height 的节点是布局边界。
子节点的布局变化不会传播到边界之外，因为边界节点的尺寸不会改变。

---

## 6. 绘制系统

### 6.1 CellBuffer

```go
type CellBuffer struct {
    cells  []Cell    // 扁平数组: cells[y*width + x]
    width, height int
    
    // 每帧统计
    writeCount, clearCount int
    dirtyMinX, dirtyMinY   int  // 脏区包围盒
    dirtyMaxX, dirtyMaxY   int
}
```

**特点**:
- 预分配，不产生 GC 压力
- 每次写入自动更新脏区包围盒
- `ResetStats()` 在每帧开始时重置统计

### 6.2 PaintDirty

```go
func PaintDirty(buf *CellBuffer, root *Node) {
    paintDirtyWalk(buf, root)
}

func paintDirtyWalk(buf *CellBuffer, node *Node) {
    if node.PaintDirty {
        buf.ClearRect(node.X, node.Y, node.W, node.H)  // 清除旧内容
        paintNode(buf, node)                              // 重绘节点+所有子节点
        node.PaintDirty = false
        return  // 子节点已在 paintNode 中绘制，不再递归
    }
    // 非脏节点 → 检查子节点
    for _, child := range node.Children {
        paintDirtyWalk(buf, child)
    }
}
```

**关键设计**: 
- 父节点 PaintDirty → 清除整个区域 + 重绘整个子树
- 这保证了删除子节点时不留残影
- 子节点不需要单独清除（父节点已清除）

### 6.3 paintNode 分发

```
paintNode(buf, node):
  switch node.Type:
    "text"      → paintText()     — 逐字符写入，继承父背景
    "box/vbox/hbox" → paintBox()  — 背景 → 边框 → 递归子节点
    "input/textarea" → paintInput() — 有内容同 text，无内容绘制 placeholder
    "component" → 直接递归子节点（透明容器）
```

### 6.4 滚动绘制

当 `overflow="scroll"` 且 `scrollY != 0` 时：

```
paintScrollChildren(buf, node):
  1. 将所有子节点的 Y 坐标偏移 -scrollY
  2. 使用 paintNodeClipped() 绘制（裁剪到容器边界）
  3. 恢复子节点 Y 坐标
```

### 6.5 背景继承

文本节点没有背景色时，**继承已绘制的父背景**：

```go
// paintText
bg := node.Style.Background
if bg == "" {
    existing := buf.Get(x, y)
    bg = existing.BG  // 读取 CellBuffer 中已有的背景色
}
buf.SetChar(x, y, ch, fg, bg, bold)
```

这依赖于绘制顺序：父节点先绘制背景，子文本后绘制时读取父背景。

---

## 7. 事件系统

### 7.1 HitTest

```go
func HitTest(root *Node, x, y int) *Node {
    // 点不在节点范围内 → nil
    if x < root.X || x >= root.X+root.W || ... { return nil }
    
    // 逆序检查子节点（后绘制的在上层）
    for i := len(root.Children) - 1; i >= 0; i-- {
        if hit := HitTest(root.Children[i], x, y); hit != nil {
            return hit
        }
    }
    return root  // 无子节点命中 → 返回自身
}
```

### 7.2 事件冒泡

```go
func HitTestWithHandler(root *Node, x, y int, eventType string) *Node {
    node := HitTest(root, x, y)  // 找到最深层节点
    // 向上冒泡找处理器
    for n := node; n != nil; n = n.Parent {
        if hasHandler(n, eventType) {
            return n
        }
    }
    return nil
}
```

### 7.3 鼠标 Enter/Leave 追踪

```go
func (e *Engine) HandleMouseMove(x, y int) {
    target := HitTest(root, x, y)
    if target == e.hoveredNode { return }  // 同一节点，无变化
    
    old := e.hoveredNode
    e.hoveredNode = target
    
    // 触发 onMouseLeave（旧节点，冒泡）
    // 触发 onMouseEnter（新节点，冒泡）
}
```

---

## 8. 输出管道

### 8.1 ToBuffer()

将 CellBuffer 转换为通用 Buffer 格式：

```go
func (e *Engine) ToBuffer() *buffer.Buffer {
    buf := buffer.New(e.width, e.height)
    for y := 0; y < e.height; y++ {
        for x := 0; x < e.width; x++ {
            c := cb.Get(x, y)
            if c.Ch == 0 && c.FG == "" && c.BG == "" {
                continue  // 跳过空 cell
            }
            buf.Set(x, y, buffer.Cell{...})
        }
    }
    return buf
}
```

### 8.2 DirtyRect

```go
func (e *Engine) DirtyRect() buffer.Rect {
    stats := e.buffer.Stats()
    return buffer.Rect{
        X: stats.DirtyX, Y: stats.DirtyY,
        W: stats.DirtyW, H: stats.DirtyH,
    }
}
```

### 8.3 App.RenderDirty()

```go
func (a *App) RenderDirty() {
    a.tracker.BeginFrame()
    a.engine.RenderDirty()
    
    // DevTools 覆盖层
    if a.devtools.Visible { paintDevToolsOverlay(...) }
    
    dirtyRect := a.engine.DirtyRect()
    if dirtyRect.W > 0 && dirtyRect.H > 0 {
        screen := a.engine.ToBuffer()
        a.adapter.WriteDirty(screen, []buffer.Rect{dirtyRect})
        a.adapter.Flush()
    }
    a.tracker.EndFrame()
}
```

**关键优化**: 只有脏区非空时才执行 ToBuffer + WriteDirty。
空闲帧（无状态变化）完全跳过输出。

---

## 9. 事件循环

```go
func (a *App) eventLoop(cfg RunConfig) error {
    ticker := time.NewTicker(frameDuration)  // 默认 60fps
    
    for {
        select {
        case <-a.quit:
            return nil
            
        case ie := <-events:
            a.handleInputEvent(ie)  // 立即处理输入
            
        case path := <-reloadCh:
            a.reloadScript(path)    // 热加载
            
        case <-ticker.C:
            // 每帧 tick:
            a.animMgr.Tick(now)     // 动画
            a.fireTimers()          // setInterval/setTimeout
            a.tickDevTools()        // DevTools 刷新
            a.RenderDirty()         // 渲染脏组件
        }
    }
}
```

**输入处理时机**: 输入事件在 select 中**立即**处理（不等待下一帧），
但渲染在 ticker 触发时才执行。这意味着多个快速输入事件会合并到同一帧渲染。

---

## 10. 性能特征

### 压力测试数据

80×23 = 1840 个独立 Cell 组件，鼠标 hover 时：

| 指标 | 值 |
|------|-----|
| 组件重渲染 | 1-2 个（进入+离开的 Cell） |
| 布局重算 | 1-2 个节点子树 |
| Cell 重绘 | ~4 个 cell（2 个 Cell × 2 cell/Cell） |
| 帧时间 | < 1ms |

### 空闲检测

```go
if rendered == 0 && !hasAnyDirty(root) {
    // 跳过布局和绘制
    return
}
```

无状态变化时，每帧只执行：
1. `ResetStats()` — O(1)
2. `renderInOrder()` — 遍历组件列表检查 Dirty，全部为 false → 返回 0
3. 空闲检测 → 跳过布局/绘制
4. 无脏区 → 跳过输出

---

## 11. Lua 交互细节

### GC 控制

渲染期间暂停 GC，避免在 Lua 栈操作中触发收集：

```go
L.SetGCStopped(true)
defer func() {
    L.SetGCStopped(false)
    L.GCStepAPI()
}()
```

### 事件处理器存储

事件处理器以 Lua 注册表引用（`LuaRef`）存储在 Node 上：

```go
desc.OnClick = getRefField(L, absIdx, "onClick")
// getRefField: L.GetField → L.IsFunction → L.Ref(RegistryIndex) → int64
```

**生命周期**: 引用在 `readDescriptor` 时创建，在协调时通过 `updateRef` 更新。
当前实现中旧引用未显式释放（标记为 TODO）。

### useState 实现

```go
func (e *Engine) luaUseState(L *lua.State) int {
    comp := e.currentComp  // renderComponent 设置的当前组件
    key := L.CheckString(1)
    
    // 初始化
    if _, exists := comp.State[key]; !exists {
        comp.State[key] = L.ToAny(2)
    }
    
    // 返回值
    L.PushAny(comp.State[key])
    
    // 返回 setter（闭包捕获 compID 和 key）
    L.PushFunction(func(L *lua.State) int {
        newValue := L.ToAny(1)
        e.SetState(compID, key, newValue)
        return 0
    })
    
    return 2  // 返回 (value, setter)
}
```
