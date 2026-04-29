package widget

import (
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func testColumns() []any {
	return []any{
		map[string]any{"header": "Name", "key": "name", "width": float64(10)},
		map[string]any{"header": "Status", "key": "status", "width": float64(8)},
	}
}

func testRows() []any {
	return []any{
		map[string]any{"name": "Alice", "status": "Active"},
		map[string]any{"name": "Bob", "status": "Inactive"},
		map[string]any{"name": "Charlie", "status": "Active"},
	}
}

func TestTableRenderEmpty(t *testing.T) {
	state := Table.NewState()
	props := map[string]any{"columns": testColumns()}
	result := Table.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Table.Render returned nil or non-*Node")
	}
	if node.Type != "vbox" {
		t.Errorf("expected root type 'vbox', got %q", node.Type)
	}
	if node.Style.Border != "rounded" {
		t.Errorf("expected border 'rounded', got %q", node.Style.Border)
	}
	// Should have 2 children: header row + separator, no data rows
	if len(node.Children) != 2 {
		t.Fatalf("empty table: expected 2 children (header + separator), got %d", len(node.Children))
	}
	// Header row
	header := node.Children[0]
	if header.Type != "hbox" {
		t.Errorf("header type: got %q, want 'hbox'", header.Type)
	}
	if len(header.Children) != 2 {
		t.Fatalf("header: expected 2 cells, got %d", len(header.Children))
	}
	// Separator
	sep := node.Children[1]
	if sep.Type != "text" {
		t.Errorf("separator type: got %q, want 'text'", sep.Type)
	}
	if !strings.Contains(sep.Content, "─") {
		t.Errorf("separator content should contain '─', got %q", sep.Content)
	}
}

func TestTableRenderWithRows(t *testing.T) {
	state := Table.NewState()
	props := map[string]any{"columns": testColumns(), "rows": testRows()}
	node := Table.Render(props, state).(*render.Node)

	// header + separator + 3 data rows = 5
	if len(node.Children) != 5 {
		t.Fatalf("expected 5 children, got %d", len(node.Children))
	}

	// Check first data row
	row0 := node.Children[2]
	if row0.Type != "hbox" {
		t.Errorf("row type: got %q, want 'hbox'", row0.Type)
	}
	if len(row0.Children) != 2 {
		t.Fatalf("row: expected 2 cells, got %d", len(row0.Children))
	}
	// "Alice" padded to width 10
	nameCell := row0.Children[0]
	if !strings.HasPrefix(nameCell.Content, "Alice") {
		t.Errorf("name cell: got %q, want prefix 'Alice'", nameCell.Content)
	}
	if len(nameCell.Content) != 10 {
		t.Errorf("name cell width: got %d, want 10", len(nameCell.Content))
	}
}

func TestTableRenderStriped(t *testing.T) {
	state := Table.NewState()
	props := map[string]any{
		"columns": testColumns(),
		"rows":    testRows(),
		"striped": true,
	}
	node := Table.Render(props, state).(*render.Node)

	// Row 0 (index 0) = no stripe
	row0 := node.Children[2]
	if row0.Style.Background != "" {
		t.Errorf("row 0 bg: got %q, want empty", row0.Style.Background)
	}
	// Row 1 (index 1) = striped
	row1 := node.Children[3]
	if row1.Style.Background != CurrentTheme.Surface0 {
		t.Errorf("row 1 bg: got %q, want %q", row1.Style.Background, CurrentTheme.Surface0)
	}
	// Row 2 (index 2) = no stripe
	row2 := node.Children[4]
	if row2.Style.Background != "" {
		t.Errorf("row 2 bg: got %q, want empty", row2.Style.Background)
	}
}

func TestTableRenderSelectedRow(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.SelectedRow = 1
	props := map[string]any{
		"columns":    testColumns(),
		"rows":       testRows(),
		"selectable": true,
	}
	node := Table.Render(props, state).(*render.Node)

	// Row 1 should have primary bg
	row1 := node.Children[3]
	if row1.Style.Background != CurrentTheme.Primary {
		t.Errorf("selected row bg: got %q, want %q", row1.Style.Background, CurrentTheme.Primary)
	}
	// Text should be PrimaryDark
	for _, cell := range row1.Children {
		if cell.Style.Foreground != CurrentTheme.PrimaryDark {
			t.Errorf("selected row cell fg: got %q, want %q", cell.Style.Foreground, CurrentTheme.PrimaryDark)
		}
	}

	// Non-selected rows should not have primary bg
	row0 := node.Children[2]
	if row0.Style.Background == CurrentTheme.Primary {
		t.Error("non-selected row should not have primary bg")
	}
}

func TestTableKeydownArrowDown(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.SelectedRow = 0
	props := map[string]any{
		"columns":    testColumns(),
		"rows":       testRows(),
		"selectable": true,
	}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	changed := Table.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowDown should return true")
	}
	if state.SelectedRow != 1 {
		t.Errorf("expected SelectedRow=1, got %d", state.SelectedRow)
	}
	if evt.FireOnChange != 2 {
		t.Errorf("FireOnChange: got %v, want 2 (1-based)", evt.FireOnChange)
	}
}

func TestTableKeydownArrowDownWraps(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.SelectedRow = 2 // last row
	props := map[string]any{
		"columns":    testColumns(),
		"rows":       testRows(),
		"selectable": true,
	}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	Table.OnEvent(props, state, evt)
	if state.SelectedRow != 0 {
		t.Errorf("ArrowDown should wrap to 0, got %d", state.SelectedRow)
	}
}

func TestTableKeydownArrowUp(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.SelectedRow = 2
	props := map[string]any{
		"columns":    testColumns(),
		"rows":       testRows(),
		"selectable": true,
	}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	changed := Table.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowUp should return true")
	}
	if state.SelectedRow != 1 {
		t.Errorf("expected SelectedRow=1, got %d", state.SelectedRow)
	}
}

func TestTableKeydownArrowUpWraps(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.SelectedRow = 0
	props := map[string]any{
		"columns":    testColumns(),
		"rows":       testRows(),
		"selectable": true,
	}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	Table.OnEvent(props, state, evt)
	if state.SelectedRow != 2 {
		t.Errorf("ArrowUp should wrap to 2, got %d", state.SelectedRow)
	}
}

func TestTableKeydownJ(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.SelectedRow = 0
	props := map[string]any{
		"columns":    testColumns(),
		"rows":       testRows(),
		"selectable": true,
	}

	evt := &render.WidgetEvent{Type: "keydown", Key: "j"}
	changed := Table.OnEvent(props, state, evt)
	if !changed {
		t.Error("j should return true")
	}
	if state.SelectedRow != 1 {
		t.Errorf("expected SelectedRow=1, got %d", state.SelectedRow)
	}
}

func TestTableKeydownK(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.SelectedRow = 2
	props := map[string]any{
		"columns":    testColumns(),
		"rows":       testRows(),
		"selectable": true,
	}

	evt := &render.WidgetEvent{Type: "keydown", Key: "k"}
	changed := Table.OnEvent(props, state, evt)
	if !changed {
		t.Error("k should return true")
	}
	if state.SelectedRow != 1 {
		t.Errorf("expected SelectedRow=1, got %d", state.SelectedRow)
	}
}

func TestTableNotSelectableIgnoresKeys(t *testing.T) {
	state := Table.NewState().(*TableState)
	props := map[string]any{
		"columns": testColumns(),
		"rows":    testRows(),
		// selectable not set (defaults to false)
	}

	keys := []string{"ArrowDown", "ArrowUp", "j", "k", "Enter"}
	for _, key := range keys {
		evt := &render.WidgetEvent{Type: "keydown", Key: key}
		changed := Table.OnEvent(props, state, evt)
		if changed {
			t.Errorf("non-selectable table should ignore key %q", key)
		}
	}
}

func TestTableThemeColors(t *testing.T) {
	state := Table.NewState()
	props := map[string]any{"columns": testColumns(), "rows": testRows()}
	node := Table.Render(props, state).(*render.Node)

	// Header should use Primary color
	header := node.Children[0]
	for _, cell := range header.Children {
		if cell.Style.Foreground != CurrentTheme.Primary {
			t.Errorf("header cell fg: got %q, want %q", cell.Style.Foreground, CurrentTheme.Primary)
		}
		if !cell.Style.Bold {
			t.Error("header cell should be bold")
		}
	}

	// Separator should use Surface1
	sep := node.Children[1]
	if sep.Style.Foreground != CurrentTheme.Surface1 {
		t.Errorf("separator fg: got %q, want %q", sep.Style.Foreground, CurrentTheme.Surface1)
	}
}

func TestTableParentPointers(t *testing.T) {
	state := Table.NewState()
	props := map[string]any{"columns": testColumns(), "rows": testRows()}
	node := Table.Render(props, state).(*render.Node)
	checkParents(t, node)
}

func TestTableColumnWidths(t *testing.T) {
	cols := []any{
		map[string]any{"header": "A", "key": "a", "width": float64(5)},
		map[string]any{"header": "B", "key": "b", "width": float64(15)},
	}
	rows := []any{
		map[string]any{"a": "short", "b": "a longer value here"},
	}
	state := Table.NewState()
	props := map[string]any{"columns": cols, "rows": rows}
	node := Table.Render(props, state).(*render.Node)

	// Header cells should have correct widths
	header := node.Children[0]
	if header.Children[0].Style.Width != 5 {
		t.Errorf("col 0 width: got %d, want 5", header.Children[0].Style.Width)
	}
	if header.Children[1].Style.Width != 15 {
		t.Errorf("col 1 width: got %d, want 15", header.Children[1].Style.Width)
	}

	// Data cells should also have correct widths
	row := node.Children[2]
	if row.Children[0].Style.Width != 5 {
		t.Errorf("data col 0 width: got %d, want 5", row.Children[0].Style.Width)
	}
	if row.Children[1].Style.Width != 15 {
		t.Errorf("data col 1 width: got %d, want 15", row.Children[1].Style.Width)
	}
}

func TestTableDefaultColumnWidth(t *testing.T) {
	cols := []any{
		map[string]any{"header": "Name", "key": "name"},
		// no width specified → default 10
	}
	state := Table.NewState()
	props := map[string]any{"columns": cols}
	node := Table.Render(props, state).(*render.Node)

	header := node.Children[0]
	if header.Children[0].Style.Width != 10 {
		t.Errorf("default col width: got %d, want 10", header.Children[0].Style.Width)
	}
}

func TestTableSelectableIsFocusable(t *testing.T) {
	state := Table.NewState()
	props := map[string]any{
		"columns":    testColumns(),
		"rows":       testRows(),
		"selectable": true,
	}
	node := Table.Render(props, state).(*render.Node)
	if !node.Focusable {
		t.Error("selectable table should be focusable")
	}
}

func TestTableNotSelectableNotFocusable(t *testing.T) {
	state := Table.NewState()
	props := map[string]any{
		"columns": testColumns(),
		"rows":    testRows(),
	}
	node := Table.Render(props, state).(*render.Node)
	if node.Focusable {
		t.Error("non-selectable table should not be focusable")
	}
}

func TestTableHoveredRow(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.HoveredRow = 0
	props := map[string]any{
		"columns": testColumns(),
		"rows":    testRows(),
	}
	node := Table.Render(props, state).(*render.Node)

	row0 := node.Children[2]
	if row0.Style.Background != CurrentTheme.Surface1 {
		t.Errorf("hovered row bg: got %q, want %q", row0.Style.Background, CurrentTheme.Surface1)
	}
}

func TestTableMouseleaveResetsHover(t *testing.T) {
	state := Table.NewState().(*TableState)
	state.HoveredRow = 1
	props := map[string]any{
		"columns": testColumns(),
		"rows":    testRows(),
	}

	evt := &render.WidgetEvent{Type: "mouseleave"}
	changed := Table.OnEvent(props, state, evt)
	if !changed {
		t.Error("mouseleave should return true when hovered")
	}
	if state.HoveredRow != -1 {
		t.Errorf("expected HoveredRow=-1, got %d", state.HoveredRow)
	}
}

func TestTableRenderAutoFocusProp(t *testing.T) {
	props := map[string]any{
		"columns":    []any{map[string]any{"header": "A", "key": "a", "width": 5.0}},
		"rows":       []any{map[string]any{"a": "x"}},
		"selectable": true,
		"autoFocus":  true,
	}
	node := Table.Render(props, Table.NewState()).(*render.Node)
	if !node.Focusable || !node.AutoFocus {
		t.Errorf("root Focusable=%v AutoFocus=%v (want both true)", node.Focusable, node.AutoFocus)
	}
}

func TestTableWidgetDefInterface(t *testing.T) {
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = Table

	if w.GetName() != "Table" {
		t.Errorf("GetName() = %q, want 'Table'", w.GetName())
	}

	state := w.GetNewState()
	if state == nil {
		t.Fatal("GetNewState() returned nil")
	}

	props := map[string]any{"columns": testColumns(), "rows": testRows()}
	result := w.DoRender(props, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
}

func TestReadColumns(t *testing.T) {
	props := map[string]any{
		"columns": []any{
			map[string]any{"header": "A", "key": "a", "width": float64(5)},
			map[string]any{"header": "B", "key": "b"},
		},
	}
	cols := readColumns(props)
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].Header != "A" || cols[0].Key != "a" || cols[0].Width != 5 {
		t.Errorf("col 0: got %+v", cols[0])
	}
	if cols[1].Header != "B" || cols[1].Key != "b" || cols[1].Width != 10 {
		t.Errorf("col 1: got %+v (expected default width 10)", cols[1])
	}
}

func TestReadColumnsNil(t *testing.T) {
	cols := readColumns(map[string]any{})
	if cols != nil {
		t.Errorf("expected nil for missing columns, got %v", cols)
	}
}

func TestReadRows(t *testing.T) {
	props := map[string]any{
		"rows": []any{
			map[string]any{"a": "1"},
			map[string]any{"a": "2"},
		},
	}
	rows := readRows(props)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestReadRowsNil(t *testing.T) {
	rows := readRows(map[string]any{})
	if rows != nil {
		t.Errorf("expected nil for missing rows, got %v", rows)
	}
}

func TestPadRight(t *testing.T) {
	if got := padRight("abc", 5); got != "abc  " {
		t.Errorf("padRight('abc', 5) = %q, want 'abc  '", got)
	}
	if got := padRight("abcdef", 3); got != "abc" {
		t.Errorf("padRight('abcdef', 3) = %q, want 'abc'", got)
	}
	if got := padRight("abc", 3); got != "abc" {
		t.Errorf("padRight('abc', 3) = %q, want 'abc'", got)
	}
}
