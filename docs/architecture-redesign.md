# Lumina Architecture Redesign

## 1. Problem Statement

Lumina is a terminal UI framework with a React-like Lua API. It works correctly for small UIs but **cannot scale**:

| Scenario | Cells | CPU | FPS |
|----------|-------|-----|-----|
| 450-cell grid (30×15) | 450 | 28% | ~30 |
| Fullscreen grid (80×23) | 1840 | 100% | <1 |
| 4K terminal (200×50) | 10000 | ∞ | 0 |

**Root cause:** Every child state change triggers a full tree rebuild — O(n) work for O(1) changes.

### Per-Frame Cost When ONE Cell's Hover Changes (1840 cells)

| Step | What happens | Allocations | Time |
|------|-------------|-------------|------|
| Root Lua PCall | 1840 `createElement()` calls | 1840 Lua tables | 1.1s |
| LuaVNodeToVNode | Convert Lua → Go objects | 1840 VNode structs | 1.4s |
| bridgeVNodeEvents | Walk tree, register handlers | 5520 map entries | 1.7s |
| DiffVNode | Compare old vs new tree | — | 0.7s |
| layout (hbox/vbox) | Recompute all positions | Frame slices | 1.1s |
| Go GC | Scan/sweep all above | — | 4.3s |
| Go malloc | Allocate all above | — | 3.0s |
| **Total** | | **12.8 GB/20s** | **12.7s/20s** |

**Ideal cost:** Re-render 2 cells, update 2 cells in frame buffer, write 2 ANSI sequences. **O(1).**

---

## 2. Architecture Principles

### 2.1 Persistent Component Tree

**Current:** Component tree is rebuilt from Lua every frame. Root's render function runs, creating 1840 `createElement` tables. Go side converts them to VNodes. Every frame.

**New:** Component tree is a **persistent Go data structure**. Lua `createElement` runs once during mount. State changes trigger **targeted re-renders** of only the affected component subtree. The Go-side component tree persists across frames.

### 2.2 Dirty Subtree Rendering

**Current:** `SetState` in any child → marks root dirty → root re-renders → all children re-render.

**New:** `SetState` marks only the component itself dirty. `renderAllDirty` finds dirty components and re-renders only them. Parent is NOT re-rendered unless its own state/props changed.

```
Before: Cell hover → root dirty → render 1840 cells
After:  Cell hover → Cell dirty → render 1 cell → patch Frame
```

### 2.3 Event Delegation

**Current:** `bridgeVNodeEvents` walks entire VNode tree every frame, registers per-element handlers. 2.4 GB allocations in 20s.

**New:** Single event listener on root. Events dispatched by **hit-testing** the VNode tree (already have `HitTestVNode`). Event handlers stored on VNode/Component, not in a separate registry. No per-frame re-registration.

### 2.4 Zero-Allocation Render Path

**Current:** Every frame allocates: VNodes, Frame slices, layout arrays, event handler maps, props maps. 12.8 GB in 20s → Go GC spends 57% of CPU.

**New:** Object pools for VNodes, Frames. In-place mutation where possible. Layout results cached on VNode. Event handlers stored persistently.

### 2.5 Incremental Layout

**Current:** `computeFlexLayout` + `layoutHBox/VBox` recomputes positions for entire tree every frame.

**New:** Layout results cached on each VNode (`X`, `Y`, `W`, `H` — already stored). Only recompute for dirty subtrees. Parent layout only invalidated if child size changes (not for content-only changes like hover color).

---

## 3. Redesign Phases

### Phase 1: Dirty Subtree Rendering (Biggest Impact)

**Goal:** When a child's state changes, only re-render that child — not root.

**Changes:**

1. **`Component.SetState`** — Don't mark root dirty. Add `HasDirtyChild` flag up the ancestor chain.

2. **`renderAllDirty`** — Process components with `HasDirtyChild`:
   - If root is self-dirty → full render (current path)
   - If root has dirty children only → call `renderDirtySubtrees(root.LastVNode)`

3. **`renderDirtySubtrees(vnode)`** — Walk the persistent VNode tree:
   - Find VNodes with dirty `ComponentRef`
   - Re-render only those components (Lua PCall)
   - Replace the subtree in-place
   - Re-layout only the affected subtree
   - Patch the Frame at the affected region

4. **`bridgeVNodeEvents`** — Skip when only children re-rendered (event handlers don't change for hover)

**Expected impact:** CPU drops from O(n) to O(k) where k = number of dirty components (typically 2 for hover).

### Phase 2: Persistent Event System

**Goal:** Eliminate per-frame event re-registration.

**Changes:**

1. **Remove `bridgeVNodeEvents`** — the biggest single function cost (1.7s/13%)

2. **Store handlers on VNode/Component** — When `luaComponentToVNode` creates a VNode with `onClick`, store the Lua ref directly on the VNode. Persists across frames.

3. **Event dispatch via hit-test** — On click/hover, `HitTestVNode` finds the target VNode, call its handler directly. No EventBus lookup needed.

4. **Hover via direct VNode lookup** — `stageHover` already does hit-test. Instead of emitting synthetic mouseenter/mouseleave events through EventBus, directly call the VNode's onMouseEnter/onMouseLeave handlers.

**Expected impact:** Eliminates 13% CPU + 2.4 GB allocations.

### Phase 3: Zero-Allocation VNode & Layout

**Goal:** Eliminate Go GC pressure from render path.

**Changes:**

1. **VNode pool** — `sync.Pool` for VNode objects. `NewVNode` pulls from pool, end of frame returns old VNodes to pool.

2. **Frame reuse** — Already partially done with `CloneInto`. Extend to never allocate new Frames — always write into existing Frame buffer.

3. **Layout caching** — Store layout results (X, Y, W, H) on VNode. Mark layout dirty only when size/position constraints change. Hover color change → no layout invalidation.

4. **Props map reuse** — Instead of creating new `map[string]any` for props every frame, reuse the existing map and update changed keys.

**Expected impact:** Eliminates 57% CPU (Go GC + malloc).

### Phase 4: Incremental Terminal Output

**Goal:** Minimize terminal I/O.

**Changes:**

1. **Region-based dirty tracking** — Track which rectangular regions of the Frame changed. Only write ANSI for those regions.

2. **Batch ANSI output** — Collect all changes in a buffer, write once per frame.

3. **Already partially implemented** — `writeDiffFrame` already does cell-level diff. Extend to skip unchanged rows entirely.

---

## 4. Implementation Order

```
Phase 1.1: HasDirtyChild + skip root re-render when only children dirty
Phase 1.2: renderDirtySubtrees — re-render dirty children in-place
Phase 1.3: Incremental layout for dirty subtrees only
Phase 2.1: Store event handlers on VNode (remove bridgeVNodeEvents)
Phase 2.2: Direct hit-test dispatch (remove EventBus for mouse events)
Phase 3.1: VNode pool
Phase 3.2: Frame buffer reuse
Phase 3.3: Layout caching
Phase 4.1: Region-based dirty output
```

Each phase is independently testable and deployable. Phase 1 alone should make fullscreen usable.

---

## 5. Performance Targets

| Scenario | Current | After Phase 1 | After Phase 2 | After Phase 3 |
|----------|---------|--------------|--------------|--------------|
| 1840 cells, hover | 100% CPU, <1 FPS | ~20% CPU, ~30 FPS | ~10% CPU, ~50 FPS | ~5% CPU, 60 FPS |
| 10000 cells, hover | ∞ | ~30% CPU | ~15% CPU | ~5% CPU |

---

## 6. Risk Mitigation

1. **Lua API compatibility** — All phases preserve the existing Lua API (`createElement`, `useState`, `useStore`, etc.). No Lua-side changes needed.

2. **Test coverage** — All existing tests must pass after each phase. Run `go test ./pkg/lumina/... -count=1` after every change.

3. **Incremental rollout** — Each phase is a separate set of commits. If a phase causes regressions, it can be reverted independently.

4. **Fallback** — If dirty subtree rendering causes correctness issues, the full-render path remains as fallback (just mark root dirty like before).