# Lumina 开发路线图

> **状态**: V2 渲染引擎已完成，进入功能扩展阶段

---

## ✅ 已完成

### V2 渲染引擎
- 持久化节点树 + 脏标记驱动
- Lua → Descriptor → Reconcile → Layout → Paint 流程
- 增量布局（只重算 LayoutDirty 子树）
- 脏区绘制（只重绘 PaintDirty 节点）
- O(k) 渲染复杂度

### 组件系统
- `createComponent` — 根组件
- `defineComponent` — 可复用子组件（工厂模式）
- `createElement` — 元素创建
- `useState` — 组件状态管理
- 子组件嫁接（graft）机制

### 组件库（Lua lux）
- 基础组件（Button, Card, Dialog, Select, ...）
- 纯 Lua 实现，基于 defineComponent

### Go Widget 清理
- 完全移除 Go widget 基础设施（WidgetEvent, WidgetDef, widgets map, widgetStates）
- Theme 系统迁移至 `pkg/render/theme.go`
- `pkg/widget/` 包已删除

### 布局
- Flex 布局（vbox 垂直 / hbox 水平）
- flex 分配、justify、align
- padding、margin、gap
- min/max 约束
- position: relative / absolute / fixed
- border: single / double / rounded
- overflow: hidden / scroll

### 事件
- onClick、onMouseEnter/Leave、onKeyDown、onScroll、onChange
- HitTest + 事件冒泡
- 焦点管理（Tab 循环、点击聚焦、autoFocus）
- input/textarea 文本编辑

### 运行时
- 60fps 事件循环
- setInterval / setTimeout / clearInterval / clearTimeout
- 热加载（文件监控 + 自动重载）
- 开发者工具（F12 切换，组件树 + 性能指标）
- 性能追踪器

### 输出
- ANSI 终端输出
- TestAdapter（测试用）
- 脏区输出（WriteDirty）

---

## 🔜 计划中

### Web 运行时
- WebSocket 服务器
- xterm.js 前端
- 多会话管理

### AI 集成
- MCP 工具（读取 UI 树、触发事件、修改组件）

### 性能优化
- 大列表虚拟滚动
- 渲染批次合并
- Lua ref 生命周期管理（当前 TODO）
