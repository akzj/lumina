package devtools

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
)

// Render produces the VNode tree for the DevTools panel.
// Signature matches component.RenderFunc.
func (p *Panel) Render(state map[string]any, props map[string]any) *layout.VNode {
	runtime.ReadMemStats(&p.GoMemStats)

	root := layout.NewVNode("box")
	root.ID = "__devtools_root"
	root.Style.Background = "#1E1E2E"
	root.Style.Border = "single"
	root.Style.Foreground = "#CDD6F4"

	// Tab bar
	tabBar := p.renderTabBar()
	root.AddChild(tabBar)

	// Content based on active tab
	switch p.ActiveTab {
	case TabElements:
		root.AddChild(p.renderElements())
	case TabPerf:
		root.AddChild(p.renderPerf())
	}

	return root
}

func (p *Panel) renderTabBar() *layout.VNode {
	bar := layout.NewVNode("hbox")
	bar.ID = "__devtools_tabbar"
	bar.Style.Background = "#313244"
	bar.Style.Height = 1

	elemTab := layout.NewVNode("text")
	elemTab.ID = "__devtools_tab_elements"
	if p.ActiveTab == TabElements {
		elemTab.Style.Background = "#1E1E2E"
		elemTab.Style.Foreground = "#89B4FA"
		elemTab.Style.Bold = true
	} else {
		elemTab.Style.Foreground = "#6C7086"
	}
	elemTab.Content = " Elements "
	bar.AddChild(elemTab)

	perfTab := layout.NewVNode("text")
	perfTab.ID = "__devtools_tab_perf"
	if p.ActiveTab == TabPerf {
		perfTab.Style.Background = "#1E1E2E"
		perfTab.Style.Foreground = "#89B4FA"
		perfTab.Style.Bold = true
	} else {
		perfTab.Style.Foreground = "#6C7086"
	}
	perfTab.Content = " Perf "
	bar.AddChild(perfTab)

	// Close hint
	hint := layout.NewVNode("text")
	hint.Style.Foreground = "#6C7086"
	hint.Content = "  [F12 close] [1 Elements] [2 Perf]"
	bar.AddChild(hint)

	return bar
}

func (p *Panel) renderElements() *layout.VNode {
	box := layout.NewVNode("box")
	box.ID = "__devtools_elements"

	if len(p.components) == 0 {
		txt := layout.NewVNode("text")
		txt.Content = "No components registered"
		txt.Style.Foreground = "#6C7086"
		box.AddChild(txt)
		return box
	}

	for _, comp := range p.components {
		header := layout.NewVNode("text")
		header.Style.Foreground = "#A6E3A1"
		header.Style.Bold = true
		header.Content = fmt.Sprintf("▸ %s [%s] (%d,%d %dx%d z=%d)",
			comp.Name, comp.ID, comp.X, comp.Y, comp.W, comp.H, comp.ZIndex)
		box.AddChild(header)

		// VNode tree (indented)
		if comp.VNodeTree != nil {
			renderVNodeTree(box, comp.VNodeTree, 1)
		}
	}

	return box
}

// renderVNodeTree recursively renders VNode tree lines into parent.
func renderVNodeTree(parent *layout.VNode, node *layout.VNode, depth int) {
	indent := strings.Repeat("  ", depth)
	line := layout.NewVNode("text")
	line.Style.Foreground = "#CDD6F4"

	desc := fmt.Sprintf("%s<%s", indent, node.Type)
	if node.ID != "" {
		desc += fmt.Sprintf(` id="%s"`, node.ID)
	}
	if node.Content != "" {
		content := node.Content
		if len(content) > 20 {
			content = content[:20] + "..."
		}
		desc += fmt.Sprintf(">%s</%s>", content, node.Type)
	} else {
		desc += ">"
	}
	line.Content = desc
	parent.AddChild(line)

	for _, child := range node.Children {
		if depth < 5 { // limit depth to avoid huge trees
			renderVNodeTree(parent, child, depth+1)
		}
	}
}

func (p *Panel) renderPerf() *layout.VNode {
	box := layout.NewVNode("box")
	box.ID = "__devtools_perf"

	// Section: Frame Stats
	title := layout.NewVNode("text")
	title.Style.Foreground = "#F9E2AF"
	title.Style.Bold = true
	title.Content = "── Frame Stats ──"
	box.AddChild(title)

	last := p.tracker.LastFrame()
	total := p.tracker.TotalStats()

	lines := []struct{ label, value string }{
		{"Renders", fmt.Sprintf("%d (total: %d)", last.Get(perf.Renders), total.Get(perf.Renders))},
		{"Layouts", fmt.Sprintf("%d (total: %d)", last.Get(perf.Layouts), total.Get(perf.Layouts))},
		{"Paints", fmt.Sprintf("%d (total: %d)", last.Get(perf.Paints), total.Get(perf.Paints))},
		{"ComposeFull", fmt.Sprintf("%d", last.Get(perf.ComposeFull))},
		{"ComposeDirty", fmt.Sprintf("%d", last.Get(perf.ComposeDirty))},
		{"DirtyRects", fmt.Sprintf("%d", last.Get(perf.DirtyRectsOut))},
		{"Events Hit", fmt.Sprintf("%d", last.Get(perf.EventsDispatched))},
		{"Events Missed", fmt.Sprintf("%d", last.Get(perf.EventsMissed))},
		{"Frame Duration", fmt.Sprintf("%v", last.Duration)},
		{"Total Frames", fmt.Sprintf("%d", total.Frames)},
	}

	for _, l := range lines {
		row := layout.NewVNode("text")
		row.Style.Foreground = "#CDD6F4"
		row.Content = fmt.Sprintf("  %-20s %s", l.label, l.value)
		box.AddChild(row)
	}

	// Section: Runtime
	rtTitle := layout.NewVNode("text")
	rtTitle.Style.Foreground = "#F9E2AF"
	rtTitle.Style.Bold = true
	rtTitle.Content = "── Runtime ──"
	box.AddChild(rtTitle)

	rtLines := []struct{ label, value string }{
		{"Lua CPU Time", fmt.Sprintf("%v", p.LuaCPUTime)},
		{"Lua Memory", formatBytes(p.LuaMemBytes)},
		{"Go Heap Alloc", formatBytes(p.GoMemStats.HeapAlloc)},
		{"Go Heap Objects", fmt.Sprintf("%d", p.GoMemStats.HeapObjects)},
		{"Go GC Cycles", fmt.Sprintf("%d", p.GoMemStats.NumGC)},
		{"Go Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine())},
	}

	for _, l := range rtLines {
		row := layout.NewVNode("text")
		row.Style.Foreground = "#CDD6F4"
		row.Content = fmt.Sprintf("  %-20s %s", l.label, l.value)
		box.AddChild(row)
	}

	return box
}

// formatBytes formats a byte count into a human-readable string.
func formatBytes(b uint64) string {
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
