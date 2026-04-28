package v2

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/akzj/lumina/pkg/lumina/v2/devtools"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
	"github.com/akzj/lumina/pkg/lumina/v2/render"
)

// toggleDevToolsV2 shows or hides the DevTools panel in V2 engine mode.
// Instead of registering a component (old pipeline), it paints directly
// onto the engine's CellBuffer as an overlay.
func (a *App) toggleDevToolsV2() {
	a.devtools.Toggle()
	if a.devtools.Visible {
		a.tracker.Enable()

		panelH := a.height * 4 / 10
		if panelH < 8 {
			panelH = 8
		}
		a.devtools.Width = a.width
		a.devtools.Height = panelH
	}
	a.paintDevToolsV2()
}

// refreshDevToolsV2 repaints the DevTools panel immediately (for tab switching).
func (a *App) refreshDevToolsV2() {
	a.devtools.SnapshotPerf()
	a.paintDevToolsV2()
}

// tickDevToolsV2 is called every frame tick when V2 engine is active.
// It updates FPS and repaints the devtools overlay if visible (throttled to ~3Hz).
func (a *App) tickDevToolsV2() {
	a.devtools.TickFPS()
	if !a.devtools.Visible {
		return
	}

	// Only refresh devtools overlay every 300ms to avoid 60Hz full-screen writes
	now := time.Now()
	if now.Sub(a.devtoolsLastRefresh) < 300*time.Millisecond {
		return
	}
	a.devtoolsLastRefresh = now

	a.updateDevToolsElements()
	a.devtools.SnapshotPerf()
	a.paintDevToolsV2()
}

// updateDevToolsElements walks the engine's render node tree and builds a
// flattened snapshot for the Elements tab (Chrome DevTools-style DOM tree).
func (a *App) updateDevToolsElements() {
	root := a.engine.Root()
	if root == nil || root.RootNode == nil {
		a.devtools.UpdateNodeTree(nil)
		return
	}

	// Cap total nodes to avoid huge slices; depth limit keeps tree readable
	const maxNodes = 500
	const maxDepth = 5

	var infos []devtools.NodeInfo
	var walk func(node *render.Node, depth int)
	walk = func(node *render.Node, depth int) {
		if len(infos) >= maxNodes {
			return
		}
		info := devtools.NodeInfo{
			Type:    node.Type,
			X:       node.X,
			Y:       node.Y,
			W:       node.W,
			H:       node.H,
			BG:      node.Style.Background,
			FG:      node.Style.Foreground,
			Content: node.Content,
			Depth:   depth,
		}
		infos = append(infos, info)
		if depth < maxDepth {
			for _, child := range node.Children {
				if len(infos) >= maxNodes {
					return
				}
				walk(child, depth+1)
			}
		}
	}
	walk(root.RootNode, 0)
	a.devtools.UpdateNodeTree(infos)
}

// paintDevToolsV2 paints the devtools panel directly onto the engine's CellBuffer.
// When visible, it draws a panel at the bottom of the screen.
// When hidden, it clears the panel area and triggers a full re-render.
func (a *App) paintDevToolsV2() {
	cb := a.engine.Buffer()

	if !a.devtools.Visible {
		// Clear the panel area and re-render the app content.
		a.engine.RenderAll()
		screen := a.engine.ToBuffer()
		_ = a.adapter.WriteFull(screen)
		_ = a.adapter.Flush()
		return
	}

	panelH := a.devtools.Height
	panelY := a.height - panelH

	// First render the app normally.
	a.engine.RenderDirty()

	// Then paint the devtools overlay on top.
	paintDevToolsOverlay(cb, a.devtools, panelY, a.width, panelH)

	// Output the full screen (devtools toggle is a major visual change).
	screen := a.engine.ToBuffer()
	_ = a.adapter.WriteFull(screen)
	_ = a.adapter.Flush()
}

// paintDevToolsOverlay renders the devtools panel content directly onto a CellBuffer.
func paintDevToolsOverlay(cb *render.CellBuffer, panel *devtools.Panel, startY, width, height int) {
	// Colors
	const (
		bgColor      = "#1E1E2E"
		fgColor      = "#CDD6F4"
		tabBgColor   = "#313244"
		activeColor  = "#89B4FA"
		dimColor     = "#6C7086"
		titleColor   = "#F9E2AF"
		greenColor   = "#A6E3A1"
	)

	// Clear the panel area with background.
	for y := startY; y < startY+height && y < cb.Height(); y++ {
		for x := 0; x < width && x < cb.Width(); x++ {
			cb.Set(x, y, render.Cell{Ch: ' ', BG: bgColor})
		}
	}

	row := startY

	// --- Tab bar ---
	tabLine := buildTabBar(panel)
	paintLine(cb, row, 0, width, tabLine, tabBgColor, fgColor, panel)
	row++

	// --- Content ---
	row++ // blank line after tab bar

	switch panel.ActiveTab {
	case devtools.TabElements:
		row = paintElementsTab(cb, panel, row, width, fgColor, bgColor, greenColor, dimColor)
	case devtools.TabPerf:
		row = paintPerfTab(cb, panel, row, width, fgColor, bgColor, titleColor)
	}
	_ = row
}

// buildTabBar creates the tab bar string.
func buildTabBar(panel *devtools.Panel) string {
	elemMark := " "
	perfMark := " "
	if panel.ActiveTab == devtools.TabElements {
		elemMark = "▸"
	}
	if panel.ActiveTab == devtools.TabPerf {
		perfMark = "▸"
	}
	return fmt.Sprintf(" %sElements  %sPerf   %d FPS  [F12 close] [1 Elements] [2 Perf]",
		elemMark, perfMark, panel.FPS())
}

// paintLine writes a string at a given row with specified colors.
func paintLine(cb *render.CellBuffer, row, startX, maxWidth int, text, bg, fg string, _ *devtools.Panel) {
	if row >= cb.Height() {
		return
	}
	x := startX
	for _, ch := range text {
		if x >= maxWidth || x >= cb.Width() {
			break
		}
		cb.Set(x, row, render.Cell{Ch: ch, FG: fg, BG: bg})
		x++
	}
	// Fill rest of line with background.
	for ; x < maxWidth && x < cb.Width(); x++ {
		cb.Set(x, row, render.Cell{Ch: ' ', BG: bg})
	}
}

// paintTextLine writes a simple text line.
func paintTextLine(cb *render.CellBuffer, row, maxWidth int, text, fg, bg string) {
	if row >= cb.Height() {
		return
	}
	x := 0
	for _, ch := range text {
		if x >= maxWidth || x >= cb.Width() {
			break
		}
		cb.Set(x, row, render.Cell{Ch: ch, FG: fg, BG: bg})
		x++
	}
	// Fill rest with background.
	for ; x < maxWidth && x < cb.Width(); x++ {
		cb.Set(x, row, render.Cell{Ch: ' ', BG: bg})
	}
}

// paintElementsTab renders the Elements tab content as a node tree.
func paintElementsTab(cb *render.CellBuffer, panel *devtools.Panel, startRow, width int, fgColor, bgColor, greenColor, dimColor string) int {
	row := startRow

	nodes := panel.NodeTree()
	if len(nodes) == 0 {
		paintTextLine(cb, row, width, "  No nodes", dimColor, bgColor)
		row++
		return row
	}

	scrollY := panel.ElementsScrollY()
	visibleLines := panel.Height - 3 // tab bar + blank + margin
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Show scroll indicator if scrolled
	if scrollY > 0 {
		indicator := fmt.Sprintf("  ↑ %d more above", scrollY)
		paintTextLine(cb, row, width, indicator, dimColor, bgColor)
		row++
		visibleLines--
	}

	end := scrollY + visibleLines
	if end > len(nodes) {
		end = len(nodes)
	}

	for i := scrollY; i < end; i++ {
		if row >= cb.Height() {
			break
		}
		node := nodes[i]
		indent := strings.Repeat("  ", node.Depth)
		var line string
		if node.Type == "text" {
			content := node.Content
			if len(content) > 30 {
				content = content[:27] + "..."
			}
			line = fmt.Sprintf("%s<text> %q", indent, content)
		} else {
			line = fmt.Sprintf("%s▸ <%s> %dx%d", indent, node.Type, node.W, node.H)
			if node.BG != "" {
				line += " bg=" + node.BG
			}
		}
		paintTextLine(cb, row, width, "  "+line, greenColor, bgColor)
		row++
	}

	// Show scroll indicator if more below
	if end < len(nodes) {
		remaining := len(nodes) - end
		indicator := fmt.Sprintf("  ↓ %d more below", remaining)
		paintTextLine(cb, row, width, indicator, dimColor, bgColor)
		row++
	}

	return row
}

// paintPerfTab renders the Perf tab content.
func paintPerfTab(cb *render.CellBuffer, panel *devtools.Panel, startRow, width int, fgColor, bgColor, titleColor string) int {
	row := startRow

	// Section: Frame Stats
	paintTextLine(cb, row, width, "  ── Frame Stats ──", titleColor, bgColor)
	row++

	snap := panel.Snapshot()
	last := snap.Last
	total := snap.Total
	fps := snap.FPS

	lines := []struct{ label, value string }{
		{"FPS", fmt.Sprintf("%d", fps)},
		{"Frame Duration", fmt.Sprintf("%v", last.Duration)},
		{"Max Frame", fmt.Sprintf("%v", total.MaxFrameDuration)},
		{"Total Frames", fmt.Sprintf("%d", total.Frames)},
		{"Renders", fmt.Sprintf("%d (total: %d)", last.Get(perf.Renders), total.Get(perf.Renders))},
		{"Layouts", fmt.Sprintf("%d (total: %d)", last.Get(perf.Layouts), total.Get(perf.Layouts))},
		{"Paints", fmt.Sprintf("%d (total: %d)", last.Get(perf.Paints), total.Get(perf.Paints))},
		{"V2 Rendered", fmt.Sprintf("%d", last.Get(perf.V2ComponentsRendered))},
		{"V2 Paint Cells", fmt.Sprintf("%d", last.Get(perf.V2PaintCells))},
		{"DirtyRects", fmt.Sprintf("%d", last.Get(perf.DirtyRectsOut))},
		{"Events Hit", fmt.Sprintf("%d", last.Get(perf.EventsDispatched))},
		{"Events Missed", fmt.Sprintf("%d", last.Get(perf.EventsMissed))},
	}

	for _, l := range lines {
		if row >= cb.Height() {
			break
		}
		text := fmt.Sprintf("    %-20s %s", l.label, l.value)
		paintTextLine(cb, row, width, text, fgColor, bgColor)
		row++
	}

	// Section: Runtime
	row++
	paintTextLine(cb, row, width, "  ── Runtime ──", titleColor, bgColor)
	row++

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	rtLines := []struct{ label, value string }{
		{"Go Heap Alloc", formatBytesV2(memStats.HeapAlloc)},
		{"Go Heap Objects", fmt.Sprintf("%d", memStats.HeapObjects)},
		{"Go GC Cycles", fmt.Sprintf("%d", memStats.NumGC)},
		{"Go Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine())},
	}

	for _, l := range rtLines {
		if row >= cb.Height() {
			break
		}
		text := fmt.Sprintf("    %-20s %s", l.label, l.value)
		paintTextLine(cb, row, width, text, fgColor, bgColor)
		row++
	}

	return row
}

// formatBytesV2 formats a byte count into a human-readable string.
func formatBytesV2(b uint64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GiB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KiB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}


