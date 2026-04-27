# Lumina v2 — Performance Tooling Design

**Status**: Proposal (not implemented)  
**Audience**: Maintainers extending v2 (`buffer`, `layout`, `paint`, `compositor`, `event`, `output`, future `app` / `bridge`)  
**Related**: `DESIGN.md` (architecture and incremental goals)

---

## 1. Goals

1. **Explain frame cost**: Break end-to-end latency into stages that match the v2 pipeline (layout → paint → compositor → output, plus event handling and Lua/bridge when present).
2. **Catch regressions**: Detect accidental growth in hot paths (e.g. full-screen `OcclusionMap.Build`, `ComputeLayout` on large trees) in CI or local runs.
3. **Stay optional**: Zero cost when disabled (default builds unchanged; no extra allocations on hot paths).
4. **Work headless**: Same hooks usable under `go test`, JSON output driver, and real TUI (no terminal-only profiler requirement).

Non-goals for the first iteration:

- Distributed tracing across processes.
- Automatic optimization (this document is observability only).
- Pixel-perfect GPU-style profiling.

---

## 2. What to Measure (Metrics Spine)

Align names with `DESIGN.md` data flow so results are actionable.

| Stage (logical) | Likely Go packages | Primary cost signals |
|-----------------|-------------------|----------------------|
| Layout | `layout` | `ComputeLayout` wall time; optional node count / tree depth |
| Paint | `paint` | Per-buffer or per-VNode-subtree paint time; cell writes |
| Occlusion rebuild | `compositor` | `OcclusionMap.Build` wall time; layers count; optional cells visited |
| Screen compose | `compositor` | `ComposeAll` / `ComposeDirty` / `ComposeRects` wall time; dirty area (cells) |
| Hit / dispatch | `event` | Hit-test and handler time; queue depth if applicable |
| Output | `output` | Bytes written; flush count; escape generation time |
| Bridge / Lua | `bridge` (future) | Cross-language call counts and duration |

Global rollups:

- **Frame time**: wall clock for one “tick” (definition fixed per entrypoint: e.g. one `app.Step` or one test scenario iteration).
- **Allocations**: `testing.AllocsPerRun` in benchmarks; optional `runtime.MemStats` delta in diagnostic mode only.

---

## 3. Instrumentation Strategy (Two Layers)

### 3.1 Layer A — Standard Go profiling (no Lumina-specific code)

Use immediately without new APIs:

- **CPU**: `go test -cpuprofile`, `pprof` for `layout`, `compositor`, `paint` packages under realistic `Benchmark*` workloads.
- **Trace**: `go test -trace` for goroutine / blocking analysis if the runtime loop becomes concurrent.
- **Benchmarks**: Table-driven benchmarks with `b.ReportMetric` for custom counters (e.g. “cells composited per op”).

**Pros**: No code churn, flamegraphs are familiar.  
**Cons**: Poor mapping from “one frame” to pipeline stages unless benchmarks are structured per stage.

### 3.2 Layer B — First-class v2 spans (implemented later)

Introduce a small internal package, e.g. `pkg/lumina/v2/internal/perf` (or `pkg/lumina/v2/perf` if exported to `app` only), with:

- **`perf.Enabled() bool`**: controlled by `LUMINA_PERF=1` and/or build tag `lumina_perf`.
- **`perf.Span(name string) func()`**: `defer perf.Span("layout")()` pattern; records elapsed time into a thread-local or per-frame buffer when enabled.
- **`perf.Counter(name string).Add(n int)`**: optional cheap integer counters (cells touched in `Build`, etc.).

**Implementation notes**:

- When disabled: `Enabled()` is false; spans compile to no-op inlining-friendly stubs (single branch or empty func body).
- When enabled: use `time.Since` + ring buffer or `sync.Map` keyed by frame id — avoid mutex in inner loop; prefer aggregating at frame end from a preallocated `[]sample`.
- **Do not** add `defer` in the hottest inner loops (e.g. per-cell in `Build`); only span **package boundaries**: `ComputeLayout`, `Build`, `Compose*`, top-level paint entry.

Optional later: hook `runtime/pprof.StartCPUProfile` from the same env for one-shot capture.

---

## 4. Integration Points (Where to Attach Spans)

Concrete insertion map (for implementers):

| Location | Span / counter name (suggested) |
|----------|-----------------------------------|
| `layout.ComputeLayout` entry | `layout.compute` |
| `compositor.OcclusionMap.Build` | `compositor.build`; counter `compositor.build_cells_visited` (increment in inner double loop only when `lumina_perf`) |
| `Compositor.ComposeAll` | `compositor.compose_all` |
| `Compositor.ComposeDirty` | `compositor.compose_dirty` |
| `Compositor.ComposeRects` | `compositor.compose_rects` |
| `paint` public paint entry (single choke point) | `paint vnode` or `paint.buffer` |
| `event` dispatcher main dispatch | `event.dispatch` |
| `output` adapter `Write` / flush | `output.write` |

Rule: **one span per major API**, not per VNode/cell, unless sampling (e.g. 1/N frames) is used.

---

## 5. Workloads (How to Produce Numbers)

1. **Micro-benchmarks** (in each package’s `*_test.go`):  
   - Full-screen N layers `Build` + `ComposeAll`.  
   - Large VNode tree `ComputeLayout`.  
   - Large buffer blit / paint path.

2. **Scenario tests** (under `app_test` or integration when `app` exists):  
   - “Hover one cell” path: assert dominant cost is paint+dirty compose, not full layout of unrelated subtrees (matches DESIGN goals).

3. **Replay / fixtures** (later):  
   - Recorded event sequences (mouse moves, resizes) driving the same app for A/B comparisons.

---

## 6. Reporting and UX

| Mode | Output |
|------|--------|
| CI / local benchmark | `go test -bench` text; optional `benchstat` comparison vs baseline |
| Dev diagnostic | `LUMINA_PERF=1`: end of frame print one-line summary (p50/p95 optional in phase 2) or append JSON line to stderr |
| Deep dive | Existing `pprof` on benchmark binary |

Optional JSON schema (phase 2) for machine-readable logs:

```json
{"frame":1,"spans":{"layout.compute_us":120,"compositor.build_us":340,"paint.us":80}}
```

---

## 7. Phased Rollout

| Phase | Deliverable |
|-------|-------------|
| **P0** | Documented benchmark scenarios + `benchstat` workflow in this file’s “Appendix” (commands only). No new packages. |
| **P1** | `internal/perf` no-op + real impl behind `lumina_perf`; spans on `ComputeLayout`, `Build`, `ComposeAll`, `ComposeDirty`, `ComposeRects`. |
| **P2** | Counters for cells visited in `Build`; scenario test asserting bounds on tree size or allocs for hover path. |
| **P3** | Optional `LUMINA_PERF` JSON lines; bridge/Lua spans when `bridge` lands. |

---

## 8. Risks and Constraints

- **Deferred overhead**: Even no-op `defer` has a cost in extremely tight loops — keep spans at coarse boundaries.
- **Timer resolution**: `time.Since` is enough for frame-level; sub-microsecond per-span noise is acceptable.
- **CI stability**: Use `-count` and fixed seeds; avoid timing-only assertions in flaky VMs (prefer alloc or structural bounds).

---

## 9. Appendix — P0 Commands (template)

```bash
# CPU profile while running compositor benchmarks (when benchmarks exist)
go test ./pkg/lumina/v2/compositor/... -bench=. -cpuprofile=cpu.prof -benchmem
go tool pprof -http=:0 cpu.prof

# Layout package
go test ./pkg/lumina/v2/layout/... -bench=. -benchmem
```

Replace paths as benchmarks are added per phase P0.
