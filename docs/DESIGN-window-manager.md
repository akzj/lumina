# Window Manager 设计方案（多窗口 / MDI）

> **状态**：设计文档（与仓库实现可不同步落地）。  
> **相关**：[DESIGN-widgets.md](./DESIGN-widgets.md)（Widget 与事件）、[render-engine-v2.md](./render-engine-v2.md)（引擎、Layer、脏区）。

## 1. 背景与问题

当前多窗口示例（如 `examples/windows_widget.lua`）通常在 **全局 `lumina.store`** 里维护 `windows[]`，用 **数组顺序表示 z-order**，在 `onChange` 里手写 **`bringToFront`（`table.remove` + 追加）** 并写回 **x/y/w/h**。

这样可行，但缺少 **单一职责的「窗口管理」抽象**，易踩以下坑：

| 问题 | 说明 |
|------|------|
| **身份与下标混淆** | 闭包捕获 `ipairs` 的索引 `i`；`bringToFront` 改变顺序后，**同一闭包内的索引不再对应同一扇窗**，在 **多次输入事件先于一次完整重渲染** 时可能把坐标写到 **错误 id**（表现为拖 A 时 B 跳一下）。 |
| **重复约定** | 每个应用各自约定「数组末尾 = 最前」「谁负责置顶」；难复用、难测。 |
| **与引擎边界不清** | **命中顺序 / 绘制顺序** 已由节点树决定；业务若再维护一套「逻辑顺序」易与渲染顺序漂移。 |

本方案定义 **Window Manager（WM）**：在 **不替代** Go `lumina.Window` 控件的前提下，为 **多实例叠放** 提供 **稳定 id、有序列表、统一写入口**，并与现有 **Layer / 焦点 / 渲染** 模型对齐。

## 2. 目标与非目标

### 2.1 目标

- **稳定身份**：一切变更（移动、缩放、置顶、关闭）以 **`windowId`（字符串或业务主键）** 为参数，**禁止**以「当前 render 循环里的数组下标」作为跨事件身份。
- **单一真相**：**叠放顺序** 只维护一份（建议 `order: string[]`），**渲染子节点顺序** 与之一致（见 §5）。
- **可复用 API**：注册 / 关闭 / 激活 / 设置几何；可选 **activeId**（活动窗）与引擎 **focusedNode**（键盘焦点在控件上）分离。
- **与现有架构兼容**：首版 **可不改 Go 引擎**；以 **Lua 模块或 Lux** 实现即可。

### 2.2 非目标（首版可不做的）

- 操作系统级多进程窗口、多显示器坐标系。
- 平铺式 tiling 布局引擎（可作为后续扩展）。
- 替代 **`CreateLayer` / Modal** 的模态栈（见 §7，二者分工不同）。

## 3. 职责划分

```
┌─────────────────────────────────────────────────────────────┐
│  Window Manager（建议：Lua 单例或每桌面一实例）              │
│  • order[]、frames[id]、activeId、register/close/activate    │
│  • 唯一写入口：禁止业务散落 table.remove(i)                 │
├─────────────────────────────────────────────────────────────┤
│  lumina.Window（Go Widget，`pkg/widget/window.go`）          │
│  • 单窗：标题栏拖拽、resize 手柄、关闭、FireOnChange         │
│  • 不知道「兄弟窗口」与全局 z-order                          │
├─────────────────────────────────────────────────────────────┤
│  Engine（`pkg/render/`）                                     │
│  • 节点树布局、绘制顺序、命中测试、焦点、capture             │
│  • 不识别「业务窗口 id」                                     │
└─────────────────────────────────────────────────────────────┘
```

| 能力 | WM | `lumina.Window` | Engine |
|------|----|-----------------|--------|
| 单窗内拖拽/缩放 | 否 | 是 | 分发事件、capture |
| 多窗谁在上 | **是** | 否 | 仅体现为子节点顺序 |
| 写回 x,y,w,h 到「哪条记录」 | **是**（按 id） | 通过 `onChange` 传出增量 | 否 |
| 模态挡后面 | 否（用 Layer） | 否 | Modal + hit-test |

## 4. 数据模型（建议）

对外可序列化、便于 `useStore` 与 DevTools 观察：

```lua
-- 概念结构（字段名可微调）
{
  order   = { "editor", "palette", "monitor" },  -- 从底到顶；最后一个 = 最前
  frames  = {
    editor  = { x = 2,  y = 1,  w = 35, h = 12, open = true,  title = "…", ... },
    palette = { x = 10, y = 3,  w = 30, h = 10, open = true,  ... },
    ...
  },
  activeId = "editor",  -- 可选：最后激活，用于快捷键作用域、状态栏等
}
```

- **`order`**：仅含 **打开** 且应参与叠放的 id；关闭时从 `order` 移除，保留 `frames[id].open = false`（支持重新打开并记住位置）。
- **`frames[id]`**：几何 + 业务字段；**不要**用数组下标当 id。
- **`activeId`**：可选；**不必**等于当前 `focusedNode` 所在窗（输入焦点可能在某窗内的 `input`）。

## 5. 与渲染、命中测试的对齐

引擎规则：**同一父节点下，子节点在 `Children` 中越靠后，越后绘制、命中越优先**（与现有 vbox/hbox 子顺序一致）。

因此 **WM → UI** 的映射应为：

```text
for _, id in ipairs(state.order) do
  -- 按顺序 append Window；最后一个 Window 对应最顶层
  children[#children+1] = Window { key = id, ... }
end
```

**禁止**：`order` 与 `render` 子列表顺序不一致的第二套顺序。

## 6. API 草案（Lua，概念层）

以下为 **应用侧** 可调用的操作集合（实现可为表 + 函数，或 Lux 封装）：

| 操作 | 行为 |
|------|------|
| `register(id, frame)` | 注册 `frames[id]`，默认追加到 `order` 末尾（最前）。 |
| `close(id)` | 从 `order` 移除 + `frames[id].open = false`（保留位置记忆）。 |
| `reopen(id)` | `frames[id].open = true` + append 到 `order` 末尾（置顶重新打开）。 |
| `activate(id)` | 从 `order` 中移除 `id` 再 `append`，使该窗置顶；更新 `activeId`。 |
| `setFrame(id, patch)` | 合并 `patch`（如 `x,y` 或 `w,h`）；**内部用 id 查找**，不用下标。 |
| `snapshot()` / `subscribe` | 若走 store：提供 `apply` 或只暴露 reducer 式更新，避免手写散落逻辑。 |

**`onChange` 对接（推荐）**：

- `Window` 的 `onChange` 收到 `{ type = "move", x, y }` / `resize` 时，调用 **`wm.setFrame(currentWindowId, { ... })`**；`currentWindowId` 来自 **创建该 `Window` 时的 `key` 或显式 `id` prop**，**不要**使用 `for` 循环的 `i`。
- **`activate` 只在 `"activate"` 事件时调用**（Window widget 在 mousedown 时单独发 `"activate"`）。**`move` / `resize` 事件只调用 `setFrame`，不再重复 `activate`**。
- **可选优化**：仅在 **mousedown** 或首次 **mousemove** 时 `activate`，避免每个像素都触发全量 `store.set`（减少重排与重渲染）。

**标准 onChange 模板**：

```lua
onChange = function(e)
    if e == "close" then
        mgr.close(win.id)
    elseif e == "activate" then
        mgr.activate(win.id)   -- 仅在 mousedown 时触发
    elseif type(e) == "table" then
        -- move/resize 只更新位置，不 activate
        if e.type == "move" then
            mgr.setFrame(win.id, { x = e.x, y = e.y })
        elseif e.type == "resize" then
            mgr.setFrame(win.id, { w = e.width, h = e.height })
        end
    end
end
```

## 6.1 语义规则（硬约束）

以下规则在实现中必须遵守：

**close 语义**：
- `close(id)` = 从 `order` 移除 + `frames[id].open = false`
- `frames[id]` 保留（记住位置/大小），支持后续 `reopen`
- `getWindows()` 只返回 `order` 中的窗口

**reopen 语义**：
- `reopen(id)` = `frames[id].open = true` + append 到 `order` 末尾
- 窗口重新打开时恢复上次的位置/大小
- 若 `frames[id]` 不存在，`reopen` 无效

**activate 语义**：
- `activate(id)` 仅在 **mousedown**（Window widget 发 `"activate"` 事件）时调用
- 拖拽过程中的 `move` / `resize` 事件 **不应** 再调用 `activate`
- `activate` = 从 `order` 中移除 `id` 再 append 到末尾 + 更新 `activeId`

**数据模型预留字段**（首版可不实现）：
- `frames[id].state`: `"normal"` | `"minimized"` | `"maximized"`（P2）
- `wm.arrange("cascade")` / `wm.arrange("tile")`（P2）

## 7. 与 Layer / Modal 的关系

| 场景 | 建议 |
|------|------|
| MDI：多文档在主内容区内叠放 | **主 Layer 根** 下挂 WM 生成的多个 `Window` 即可。 |
| 真模态（必须挡后面、点外关闭等） | 继续使用 **`WidgetEvent.CreateLayer`** 与引擎 **Modal** 语义（见 `pkg/render/events.go`）；**不要**把模态窗混进 MDI 的 `order` 与主窗一起做 `HitTest` 穿透，避免两套「谁挡谁」冲突。 |

约定：**一层 Layer 内一个 WM 实例** 通常足够；跨 Layer 的「全局 WM」非首版必要。

## 8. 与运行时刷新节奏的关系

应用主循环中，**鼠标事件路径未必每个事件都触发 `RenderDirty`**（常见为 ticker 合帧）。因此：

- 用 **id** 更新 store 可避免「下标与顺序在两次渲染之间错位」。
- 若仍出现高频 `store.set`，可通过 **activate 降频** 或 **拖拽结束再 commit** 等策略优化（属实现细节，不改变 WM 抽象）。

## 9. 落地阶段（建议）

| 阶段 | 内容 |
|------|------|
| **P0** | Lua 模块：`order` + `frames` + 仅 id 的 API；改写 `examples/windows_widget.lua` 为推荐范例。 |
| **P1** | 纳入 Lux（如 `lux.WindowManager`）+ 主题/状态栏与 `activeId` 联动示例。 |
| **P2** | 若 profiling 需要：Go 侧可选「只读快照」或 DevTools 专用查询接口；**仍保持** WM 策略在 Lua 可调。 |

## 10. 验收要点（设计层）

- 快速连续拖拽、多窗交替拖拽后，**任意窗的 `frames[id]` 仅在该 id 的 `onChange` 链路上被写入**。
- **`order` 与屏幕上「谁压谁」一致**。
- 模态 Layer 与 MDI WM **文档化分工**，示例不混用两种语义于同一交互。

---

**文档维护**：若将来在 `widget.All()` 中增加与 WM 强绑定的 Go 类型，在本节追加 **实现状态** 与代码路径即可。
