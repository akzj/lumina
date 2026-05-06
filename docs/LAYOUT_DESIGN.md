# Lumina 布局系统设计文档

## 概述
Lumina 布局系统采用类 CSS Flexbox 设计，遵循 React 式的响应式布局语义，同时针对终端场景做了优化，实现了高性能增量布局。

---

## 核心架构

### 双阶段布局流程
```
Measure 阶段（自底向上） → Layout 阶段（自顶向下）
```

#### 1. Measure 阶段
- 从叶子节点向上计算每个节点的固有尺寸
- 结果缓存到 `node.MeasuredW` / `node.MeasuredH`
- 考虑 `minWidth`/`maxWidth`/`minHeight`/`maxHeight` 约束

#### 2. Layout 阶段
- 从根节点向下计算每个节点的位置和尺寸
- 实现 Flex 布局逻辑
- 处理定位、间距、边框等
- 标记 `PaintDirty` 触发重绘

---

## 布局模式

### 基础布局类型

| 类型 | 描述 |
|------|------|
| `vbox` | 垂直弹性盒，子元素垂直排列 |
| `hbox` | 水平弹性盒，子元素水平排列 |
| `grid` | 网格布局（结合 `gridColumns` 属性） |
| `fragment` | 透明容器，不占空间，仅承载子元素 |
| `component` | 组件占位符，支持子组件嫁接（graft） |
| `text` | 文本节点，自动换行计算 |

### 定位模式

| 定位模式 | 描述 |
|----------|------|
| `static` (默认) | 正常流布局 |
| `absolute` | 相对于最近的非 static 父容器定位 |
| `fixed` | 相对于根视口定位 |

---

## 布局属性

### 尺寸属性
| 属性 | 类型 | 说明 |
|------|------|------|
| `width`/`height` | number/string | 固定尺寸，支持 `px`/`%`/`vw`/`vh` 单位 |
| `minWidth`/`minHeight` | number | 最小尺寸约束 |
| `maxWidth`/`maxHeight` | number | 最大尺寸约束 |
| `flexGrow` | number | 弹性增长系数 |
| `flexShrink` | number | 弹性收缩系数 |
| `flexBasis` | number | 弹性基准尺寸 |

### 间距属性
| 属性 | 类型 | 说明 |
|------|------|------|
| `padding` | number | 内边距（简写，应用到上下左右） |
| `paddingTop`/`paddingBottom`/`paddingLeft`/`paddingRight` | number | 各方向内边距 |
| `margin` | number | 外边距（简写，应用到上下左右） |
| `marginTop`/`marginBottom`/`marginLeft`/`marginRight` | number | 各方向外边距 |
| `gap` | number | 子元素之间的间距 |

### 边框属性
| 属性 | 类型 | 说明 |
|------|------|------|
| `border` | string | 边框样式：`single`/`double`/`rounded`/`none` |
| `borderColor` | string | 边框颜色（ANSI 颜色名或十六进制） |

### 对齐属性
| 属性 | 类型 | 说明 |
|------|------|------|
| `justifyContent` | string | 主轴对齐：`start`/`center`/`end`/`space-between`/`space-around`/`space-evenly` |
| `alignItems` | string | 交叉轴对齐：`start`/`center`/`end`/`stretch` |
| `alignSelf` | string | 单个子元素交叉轴对齐，覆盖 `alignItems` |
| `flexDirection` | string | 主轴方向：`row`/`column`/`row-reverse`/`column-reverse` |
| `flexWrap` | string | 换行：`nowrap`/`wrap`/`wrap-reverse` |

### 位置属性
| 属性 | 类型 | 说明 |
|------|------|------|
| `position` | string | 定位模式：`static`/`absolute`/`fixed` |
| `top`/`bottom`/`left`/`right` | number | 偏移量 |
| `zIndex` | number | 层级顺序（数字越大越靠上） |

### 溢出属性
| 属性 | 类型 | 说明 |
|------|------|------|
| `overflow` | string | 溢出处理：`hidden`/`scroll`/`visible` |
| `scrollX`/`scrollY` | number | 滚动偏移量 |
| `scrollWidth`/`scrollHeight` | number | 内容总尺寸（只读） |

---

## 高性能特性

### 增量布局
- 仅重算 `LayoutDirty` 标记的子树
- 未变化的节点保留缓存的位置和尺寸
- 大幅减少布局计算量，提升渲染性能

### 脏区绘制
- 布局变化时仅标记相关区域为 `PaintDirty`
- 绘制时仅重绘变化区域，减少终端 IO
- 支持跨帧布局缓存，避免重复计算

### 布局缓存
- 节点尺寸和位置缓存，无需每次重渲染都重新计算
- 热重载时自动失效缓存
- 支持布局快照和恢复

---

## 单位支持

| 单位 | 说明 |
|------|------|
| 无单位 | 终端字符列/行 |
| `px` | 像素，同字符单位 |
| `%` | 相对于父容器尺寸的百分比 |
| `vw` | 视口宽度的 1% |
| `vh` | 视口高度的 1% |

---

## 示例代码

### 垂直布局示例
```lua
local app = lumina.app({
    id = "demo",
    render = function()
        return lumina.createElement("vbox", {
            style = {
                width = "100%",
                height = "100%",
                padding = 1,
                gap = 1
            },
            children = {
                lumina.createElement("text", {content = "Header"}),
                lumina.createElement("hbox", {
                    style = {
                        flexGrow = 1,
                        gap = 1
                    },
                    children = {
                        lumina.createElement("box", {
                            style = {flexGrow = 1, border = "single"}
                        }),
                        lumina.createElement("box", {
                            style = {flexGrow = 1, border = "single"}
                        })
                    }
                }),
                lumina.createElement("text", {content = "Footer"})
            }
        })
    end
})
```

### 绝对定位示例
```lua
lumina.createElement("box", {
    style = {
        width = 20,
        height = 10,
        position = "absolute",
        top = 2,
        left = 5,
        border = "single"
    }
})
```

---

## 性能最佳实践

1. **避免频繁修改布局属性**：修改 `width`/`height`/`padding`/`margin` 会触发重布局
2. **使用 `flexGrow` 代替动态计算尺寸**：弹性布局性能更高
3. **长列表使用 `lux.vlist` 组件**：虚拟滚动大幅减少节点数
4. **批量修改状态**：减少布局重算次数
5. **避免过深的嵌套**：布局深度超过 500 会被强制中断，防止无限递归

---

## 约束与限制

- 终端字符是最小单位，不支持小数像素
- Flexbox 语义与 CSS 基本一致，但不支持部分复杂特性（如 `order`/`flex-flow`）
- 百分比计算基于父容器的可用尺寸，不考虑边框和内边距
- 固定定位元素基于根视口，不支持嵌套 fixed 定位
