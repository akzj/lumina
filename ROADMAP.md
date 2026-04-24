# Lumina Development Roadmap v2.0

> **Vision**: Lumina = 下一代应用开发范式  
> 一套 Lua 代码 → 桌面终端 + Web (xterm.js) + AI Agent (MCP)  
> **Status**: React API 全对标完成 (Phase 1-15), 进入组件生态建设

---

## Phase 16: 浮层系统 (Overlay/Popup/Z-Index) 🔴 P0

**目标**: 支持 Dialog, Dropdown, Popover, Tooltip 等弹出层组件

**实现**:
- `position: "absolute"` / `"fixed"` 样式属性
- `zIndex` 层叠顺序
- Overlay 渲染层 (base layer → overlay layers, 后渲染覆盖先渲染)
- `lumina.showOverlay(vnode, {x, y, width, height, zIndex})`
- `lumina.hideOverlay(id)`

**测试**: 10+ tests

---

## Phase 17: 动画系统 (Animation/Transition) 🔴 P0

**目标**: 组件交互反馈 (hover, active, transition, spinner)

**实现**:
- `useAnimation(config)` hook — 返回当前帧值
- `transition` 样式属性 — 属性变化时自动插值
- 内置动画: fadeIn, fadeOut, slideIn, slideOut, pulse, spin
- 帧调度器 — 与 60fps ticker 集成

**测试**: 8+ tests

---

## Phase 18: 热加载 + 焦点 Trap + 路由 🔴 P0

**目标**: 开发效率 + Dialog 焦点 + SPA 导航

**实现**:
- 热加载: 文件监控 (os.Stat 轮询) → 快照状态 → 重载 Lua → 恢复状态
- 焦点 Trap: `FocusScope` 组件, Tab 循环限制在 scope 内
- 路由: `lumina.createRouter({routes})`, `lumina.navigate(path)`, `useRoute()`

**测试**: 12+ tests

---

## Phase 19-20: 状态管理 + 开发者工具

**Phase 19**: 全局状态管理库
- `lumina.createStore({state, actions})` — 类 Zustand
- 配合 `useSyncExternalStore` 使用

**Phase 20**: CLI 工具链
- `lumina init` — 项目脚手架
- `lumina dev` — 热重载开发服务器
- `lumina build` — 打包

---

## Phase 21-23: shadcn/ui 组件库 (纯 Lua)

**Phase 21**: 基础组件 (~20 个)
- Alert, Badge, Card, Label, Separator, Skeleton, Spinner, Avatar
- Breadcrumb, Kbd, Empty, Button (variants), ButtonGroup, AspectRatio

**Phase 22**: 表单组件 (~15 个)
- Input, InputGroup, InputOTP, Select, NativeSelect, Checkbox
- RadioGroup, Switch, Slider, Toggle, ToggleGroup, Field, Textarea, Combobox

**Phase 23**: 复合组件 (~20 个)
- Dialog, AlertDialog, Sheet, Drawer, DropdownMenu, ContextMenu
- Popover, Tooltip, HoverCard, Accordion, Collapsible, Tabs
- Table, Carousel, Pagination, NavigationMenu, Sidebar, Command
- Menubar, ScrollArea, Resizable, Sonner/Toast

---

## Phase 24-25: Web 运行时

**Phase 24**: WebSocket 服务器
- `lumina serve app.lua --port 8080`
- Go HTTP server + WebSocket, 推送 ANSI 流到浏览器

**Phase 25**: xterm.js 前端 + 双向交互
- 极简 HTML + xterm.js
- 浏览器键盘/鼠标 → WebSocket → Go → Lua
- 多会话管理 (每连接独立 Lua VM)

---

## Phase 26-29: AI 原生 + 生产化

**Phase 26**: MCP 工具完善
- AI 可创建/修改组件, 读取 UI 树, 触发事件

**Phase 27**: 文档 + README
- API 文档, 教程, React 迁移指南

**Phase 28**: 安全沙箱
- Lua 脚本沙箱, 资源配额

**Phase 29**: 性能优化
- 渲染批次合并, 大列表虚拟滚动, VDom diff 优化

---

## 统计 (Phase 15 完成时)

| 指标 | 数值 |
|------|------|
| Go 源码 | 11,643 行 |
| 测试代码 | 10,395 行 |
| 测试用例 | 404 个 |
| React Hooks | 15 个 |
| Component APIs | 18 个 |
| Lua 组件 | 11 个 |
| Lua 示例 | 6 个 |
