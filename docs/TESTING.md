# Lumina 测试方案

> 说明如何为 **Go 引擎 / Go Widget / Lua 应用与组件** 选择测试层次、编写用例与避免常见脆弱点。  
> **实现入口**：`pkg/testing.go`（`test.*` API）、`pkg/lua_test.go`（`TestLuaTestFramework`）、`pkg/testdata/lua_tests/**/*_test.lua`。

---

## 1. 测试分层（建议金字塔）

```
                    ┌─────────────┐
                    │ 少量 E2E   │  完整脚本、示例应用、真实交互链
                    ├─────────────┤
                    │ Lua 集成   │  test.createApp + loadString/loadFile
                    ├─────────────┤
                    │ Go 单元    │  render/、widget/*_test.go、纯逻辑
                    └─────────────┘
```

| 层级 | 适用对象 | 典型路径 | 优点 | 注意 |
|------|----------|----------|------|------|
| **Go 单元** | `pkg/render`、layout、reconciler、纯算法；**Go Widget 的 `Render`/`OnEvent`**（不启 Lua） | `*_test.go` | 快、稳定、易调试 | 不覆盖 Lua API、不覆盖 `lumina.app` 生命周期 |
| **Lua 集成（本框架主力）** | `lumina.app`、`defineComponent`、store、hooks、**业务 Lua 组件** | `pkg/testdata/lua_tests/**/_test.lua` | 覆盖真实宿主与事件管线 | 像素/布局敏感时需设计断言 |
| **E2E / 大脚本** | 完整 `examples/*.lua`、多模块 | `lua_e2e_v2_test.go`、`examples_*` | 防回归整条产品路径 | 慢、失败定位难，宜少而精 |

**原则**：能在 **Go 单测** 断言的（例如某 Widget 收到 `ArrowDown` 后 `ScrollBy`），不要拖到 Lua；涉及 **store + render + 输入** 的，用 **Lua 集成**。

---

## 2. Lua 测试框架（`test` 全局表）

由 `pkg.TestRunner` 加载每个 `*_test.lua` 前注册，**与业务 App 使用独立 Lua state**（`test.createApp` 内部再建 `appL`），避免污染。

### 2.1 结构

```lua
test.describe("套件名", function()
    test.beforeEach(function() ... end)
    test.afterEach(function() ... end)

    test.it("用例名", function()
        test.assert.eq(a, b)
        test.assert.neq(a, b)
        test.assert.notNil(v)
        test.log("调试信息")
    end)
end)
```

### 2.2 `app = test.createApp(w, h)` 常用 API

| API | 用途 |
|-----|------|
| `app:loadFile(path)` | 执行脚本（相对 `pkg/` 时常用 `../examples/foo.lua`），并 `RenderAll` |
| `app:loadString(code)` | 内联最小复现，适合小组件 |
| `app:click(x, y)` / `app:click("vnode-id")` | 点击；**按 id 点击**可减少魔法坐标（需在描述符上设 `id`） |
| `app:mouseDown` / `mouseMove` / `mouseUp` | 拖拽、capture 等分步事件 |
| `app:keyPress(key)` | 与终端一致时优先 **`ArrowDown`** 等真实键名，而非仅 `Down` |
| `app:scroll(x, y, delta)` | 滚轮 |
| `app:render()` | 显式 `RenderDirty`（定时器 / effect 改状态后需要时） |
| `app:tick()` / `app:waitAsync(ms)` | 异步、`spawn` |
| `app:screenText()` | 整屏文本（含换行） |
| `app:screenContains(sub)` | 子串存在（粗断言） |
| `app:cellAt(x, y)` | 单格 `{char, fg, bg, ...}`，适合精确像素断言 |
| `app:find(id)` / `app:vnodeTree()` / `app:findAll(type)` | 结构、布局调试用 |
| `app:getState(compID, key)` / `app:setState(...)` | 组件状态 |
| `app:destroy()` | `beforeEach` 里创建的 app 在 `afterEach` 释放 |

**说明**：`click` / `mouse*` / `keyPress` / `scroll` 在 harness 内已跟 **`RenderDirty()`**（与真实 `handleInputEvent` 不完全相同，但多数用例足够）。若依赖 **多帧 ticker**，可额外 `app:render()`。

### 2.3 运行方式（全量 / 单文件 / 子用例）

**全量**（递归跑 `pkg/testdata/lua_tests/**/*_test.lua`）：

```bash
go test ./pkg/ -run TestLuaTestFramework -count=1
# 或（仓库根目录）
./scripts/lua-test.sh
./scripts/lua-test.sh -- -v
```

**只跑一个 Lua 文件**（改测迭代最快；路径相对于 **`pkg/`** 包目录）：

```bash
LUMINA_LUA_TEST=testdata/lua_tests/examples/scrollview_test.lua go test ./pkg/ -run TestLuaTestFramework -count=1
```

**封装脚本**（在仓库根执行；自动 `cd pkg`；支持路径子串唯一匹配）：

```bash
./scripts/lua-test.sh examples/scrollview_test.lua
./scripts/lua-test.sh scrollview                    # 唯一匹配 *_test.lua*scrollview*
./scripts/lua-test.sh examples/wm_test.lua -- -v  # 额外 go test 参数放在 -- 之后
```

实现：`pkg/lua_test.go` 读取环境变量 **`LUMINA_LUA_TEST`**；未设置时行为与原来一致（`RunDir`）。

**只跑套件里的某几条（Go 子测试名过滤）**：`TestLuaTestFramework` 会为每个 `it` 注册子测试 `Suite/itName`，可用正则过滤，例如只跑描述块名里带 `ScrollView` 的用例：

```bash
go test ./pkg/ -run 'TestLuaTestFramework/ScrollView' -count=1
```

---


## 3. 开发 Lua 组件时如何测「正确性」

### 3.1 先定义「正确」的观测面

| 观测面 | 手段 | 适用 |
|--------|------|------|
| **屏幕** | `screenContains` / `screenText` / `cellAt` | UI 文案、是否出现按钮、某格字符 |
| **VNode / id** | `app:find("id")` 检查 `x,y,w,h` | 布局、可见性（比纯像素少碎一点） |
| **store** | `lumina.store` 若可从脚本读；或通过 **key 暴露到 render** 再 `screenContains` | 业务状态机 |
| **组件 state** | `app:getState("compId", key)` | `useState` / 绑定 |

组合：**关键业务用 store 或稳定文案断言；纯位置用 `find(id)` 或常量坐标并注释布局依据。**

### 3.2 推荐写法：由内到外

1. **最小复现（`loadString`）**  
   只 `lumina.app { render = function() ... end }` 或只注册一个 `defineComponent`，去掉无关路由/大示例，失败时栈更短。

2. **再挂真实示例（`loadFile`）**  
   与 `examples/*.lua` 对齐，防「示例改了、文档没改」类回归（如 `wm_test.lua` 加载 `windows_widget.lua`）。

3. **断言由粗到细**  
   - 先 `screenContains` 保证主路径；  
   - 对 z-order、颜色等再用 `cellAt` 或 `find` 收紧。

### 3.3 降低脆弱性

- **优先 `app:click("my-button-id")`**，避免 `(37, 4)` 随边框/主题漂移。  
- 坐标不可避免时：**文件顶部常量 + 注释**（对应示例里哪个窗、哪一列是关闭钮）。  
- **键名**：终端为 `ArrowUp`/`ArrowDown` 时，测试里与之一致（见 `pkg/terminal/parse.go`）。  
- **一个 `it` 一个行为**，避免「渲染 + 点击 + 再断言另一件事」混在同一用例名里。  
- **异步**：`tick` / `waitAsync` / 额外 `render` 写进用例注释。

### 3.4 `defineComponent` vs `lumina.app` 示例

| 形态 | 测法建议 |
|------|----------|
| **单组件库** | `loadString` 里 `defineComponent` + `lumina.createElement` 挂到最小 `box`；用 `click`/`keyPress` + 屏幕断言 |
| **全页应用** | `loadFile("../examples/xxx.lua")`；或把核心逻辑抽到 `require("myapp.core")` 再在测试里 `loadString` 只测 core |
| **Lux / `require("lux.*")`** | 与 widget 例测一致：`loadFile` 或 `loadString` 内 `require`；注意 `package.path` 已由 `loadFile` 脚本目录配置 |

### 3.5 与 Go Widget 测试的分工

- **Go**（`pkg/widget/*_test.go`）：`WidgetEvent` 进、`ScrollBy`/状态 出；**不跑** `lumina.app`。  
- **Lua**：工厂参数、`onChange`、`store`、`useEffect` 与引擎合流。  
同一功能 **两边都测一层** 最理想：Go 锁行为，Lua 锁集成。

---

## 4. 目录与命名约定

| 路径 | 用途 |
|------|------|
| `pkg/testdata/lua_tests/` | 根级：框架、store、router、hooks 等 |
| `pkg/testdata/lua_tests/widgets/` | Go Radix 控件与 Lua 工厂交互 |
| `pkg/testdata/lua_tests/lux/` | Lux 模块 |
| `pkg/testdata/lua_tests/examples/` | 与 `examples/` 对齐的冒烟 / 回归 |
| `pkg/testdata/lua_tests/hooks/` | `useEffect` 等 |

文件命名：**`*_test.lua`**，否则 `TestRunner.RunDir` 不会收集。

---

## 5. 与现有文档的关系

- **环境、构建、简版测试命令**：[`DEVELOPMENT.md`](./DEVELOPMENT.md)  
- **Widget / 事件语义**：[`DESIGN-widgets.md`](./DESIGN-widgets.md)  
- **引擎行为**：[`render-engine-v2.md`](./render-engine-v2.md)  

---

## 6. 对 AI / 自动化助手是否友好

**整体：友好。** 原因包括：可复制的 **`go test`** 命令、**固定目录** `pkg/testdata/lua_tests/`、**`*_test.lua`** 收集规则、**`test.*` / `app:*` 方法名** 与 `pkg/testing.go` 一致，便于检索与对齐实现。

**仍建议人类/AI 共同注意：**

| 点 | 说明 |
|----|------|
| **实现为准** | API 以 `pkg/testing.go` 的 `luaCreateApp` 注册为准；文档未列出的方法勿假设存在。 |
| **改测必跑** | 修改或新增 `*_test.lua` 后执行：`go test ./pkg/ -run TestLuaTestFramework -count=1`。 |
| **键名** | 终端输入多为 `ArrowUp`/`ArrowDown`，与 `pkg/terminal/parse.go` 一致；勿只写 `Down`。 |
| **路径** | `app:loadFile` 相对 **进程 cwd**（常为 `pkg/`），故常用 `../examples/...` 或相对 `testdata` 的路径。 |
| **观测断言** | 优先写清「断言对象」（屏幕子串 / `cellAt` / `find(id)`），减少含糊的「应该对了」。 |

**给 AI 的短指令模板（可贴进任务说明）：**

```text
在 pkg/testdata/lua_tests/<子目录>/ 新增或修改 *_test.lua；
使用 test.describe / test.it / test.createApp；
交互后依赖 app:render() 时用 app:render()；
迭代时优先：./scripts/lua-test.sh <路径或子串> -- -count=1
全量回归：./scripts/lua-test.sh
```

---

## 7. 可选后续（仓库演进）

- 为 `test.assert` 增加 `truthy`、`matches(str, pat)` 等，减少 fragile 子串判断。  
- 文档化 **从单文件跑 Lua 测试** 的官方入口（flag 或 `go test` 子测试名）。  
- 对关键 **纯 Lua 模块**（如 `lux.wm`）增加 **不启动 App 的单元测试**（仅 `require` + 表操作），与 `createApp` E2E 互补。
