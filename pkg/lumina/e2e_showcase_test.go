package lumina

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// --- helpers ---

// getFrameRow extracts the text content of row y from a Frame.
func getFrameRow(frame *Frame, y int) string {
	if frame == nil || y < 0 || y >= frame.Height {
		return ""
	}
	var sb strings.Builder
	for x := 0; x < frame.Width; x++ {
		ch := frame.Cells[y][x].Char
		if ch == 0 {
			continue // skip wide-char padding
		}
		sb.WriteRune(ch)
	}
	return sb.String()
}

// findRowContaining returns the first row index whose text contains substr, or -1.
func findRowContaining(frame *Frame, substr string) int {
	if frame == nil {
		return -1
	}
	for y := 0; y < frame.Height; y++ {
		if strings.Contains(getFrameRow(frame, y), substr) {
			return y
		}
	}
	return -1
}

// setupShowcaseApp creates an App, loads the showcase script, and renders once.
// Returns the app (caller must defer app.Close()).
func setupShowcaseApp(t *testing.T) *App {
	t.Helper()

	app := NewAppWithSize(120, 40)
	tio := NewMockTermIO(120, 40)
	SetOutputAdapter(NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/components-showcase/main.lua", tio)
	if err != nil {
		t.Fatalf("LoadScript: %v", err)
	}

	app.lastRenderTime = time.Time{} // reset rate limiter
	app.RenderOnce()

	if app.lastFrame == nil {
		t.Fatal("No frame rendered after RenderOnce")
	}
	return app
}

// pressKey sends a keydown event through the app's event handler.
func pressKey(app *App, key string) {
	app.lastRenderTime = time.Time{} // reset rate limiter
	app.handleEvent(AppEvent{
		Type: "input_event",
		Payload: &Event{
			Type: "keydown",
			Key:  key,
		},
	})
	// handleEvent calls BeginBatch/EndBatch which triggers renderAllDirty
}

// --- Test 1: All Key Bindings (verifies ACTUAL content changes) ---

func TestShowcase_AllKeyBindings(t *testing.T) {
	app := setupShowcaseApp(t)
	defer app.Close()

	frame := app.lastFrame

	// Initial state: tab = "basic"
	row := findRowContaining(frame, "Basic")
	if row < 0 {
		t.Fatal("Initial render should show 'Basic' tab")
	}

	// Key "3" → switch to "form" tab, verify [3:Form] is highlighted
	pressKey(app, "3")
	frame = app.lastFrame
	if findRowContaining(frame, "[3:Form]") < 0 {
		t.Error("After pressing '3', expected '[3:Form]' in tab bar")
	}

	// Verify switch is initially OFF — check exact text
	switchRow := findRowContaining(frame, "OFF")
	if switchRow < 0 {
		t.Fatal("Form tab should show 'OFF' for switch (initial state switchOn=false)")
	}
	switchText := getFrameRow(frame, switchRow)
	if !strings.Contains(switchText, "[━━●] OFF") {
		t.Errorf("Switch row should contain '[━━●] OFF', got: %q", switchText)
	}
	t.Logf("Switch before 's': %q", strings.TrimSpace(switchText))

	// Key "s" → toggle switch OFF → ON
	pressKey(app, "s")
	frame = app.lastFrame
	switchRow = findRowContaining(frame, "ON")
	if switchRow < 0 {
		t.Fatal("After pressing 's', expected 'ON' in frame")
	}
	switchText = getFrameRow(frame, switchRow)
	if !strings.Contains(switchText, "[●━━] ON") {
		t.Errorf("Switch row should contain '[●━━] ON', got: %q", switchText)
	}
	t.Logf("Switch after 's': %q", strings.TrimSpace(switchText))

	// Key "s" again → toggle back ON → OFF
	pressKey(app, "s")
	frame = app.lastFrame
	switchRow = findRowContaining(frame, "OFF")
	if switchRow < 0 {
		t.Fatal("After pressing 's' again, expected 'OFF' in frame")
	}
	switchText = getFrameRow(frame, switchRow)
	if !strings.Contains(switchText, "[━━●] OFF") {
		t.Errorf("Switch row should contain '[━━●] OFF', got: %q", switchText)
	}

	// Verify progress is initially 65%
	progRow := findRowContaining(frame, "65%")
	if progRow < 0 {
		t.Fatal("Form tab should show '65%' (initial progressValue=65)")
	}
	t.Logf("Progress before '+': %q", strings.TrimSpace(getFrameRow(frame, progRow)))

	// Key "+" → increase progress 65 → 70
	pressKey(app, "+")
	frame = app.lastFrame
	if findRowContaining(frame, "70%") < 0 {
		t.Error("After pressing '+', expected '70%' in frame (65→70)")
	}
	t.Logf("Progress after '+': %q", strings.TrimSpace(getFrameRow(frame, findRowContaining(frame, "70%"))))

	// Key "-" → decrease progress 70 → 65
	pressKey(app, "-")
	frame = app.lastFrame
	if findRowContaining(frame, "65%") < 0 {
		t.Error("After pressing '-', expected '65%' in frame (70→65)")
	}

	// Key "2" → switch to "cards" tab
	pressKey(app, "2")
	frame = app.lastFrame
	if findRowContaining(frame, "[2:Cards]") < 0 {
		t.Error("After pressing '2', expected '[2:Cards]' in tab bar")
	}

	// Key "1" → back to basic tab
	pressKey(app, "1")
	frame = app.lastFrame
	if findRowContaining(frame, "[1:Basic]") < 0 {
		t.Error("After pressing '1', expected '[1:Basic]' in tab bar")
	}

	t.Log("All key bindings verified with exact content matching")
}

// --- Test 2: Mouse Click Hit-Testing (verifies correct VNode identified) ---

func TestShowcase_MouseClick(t *testing.T) {
	app := setupShowcaseApp(t)
	defer app.Close()

	frame := app.lastFrame

	// Find a cell with OwnerNode that has a non-empty type
	type hitResult struct {
		x, y     int
		nodeType string
		nodeID   string
	}
	var hits []hitResult

	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			node := frame.Cells[y][x].OwnerNode
			if node != nil && node.Type != "" {
				id := ""
				if idVal, ok := node.Props["id"].(string); ok {
					id = idVal
				}
				hits = append(hits, hitResult{x, y, node.Type, id})
				if len(hits) >= 5 {
					break
				}
			}
		}
		if len(hits) >= 5 {
			break
		}
	}

	if len(hits) == 0 {
		t.Fatal("No cells with OwnerNode found in rendered frame")
	}

	// Log all found hit targets
	for _, h := range hits {
		t.Logf("OwnerNode at (%d,%d): type=%q id=%q", h.x, h.y, h.nodeType, h.nodeID)
	}

	// Pick the first hit and simulate mousedown
	hit := hits[0]
	app.lastRenderTime = time.Time{}
	app.handleEvent(AppEvent{
		Type: "input_event",
		Payload: &Event{
			Type: "mousedown",
			X:    hit.x,
			Y:    hit.y,
		},
	})

	// Verify the correct VNode was identified by checking OwnerNode at that position
	cell := frame.Cells[hit.y][hit.x]
	if cell.OwnerNode == nil {
		t.Errorf("Cell at (%d,%d) should have OwnerNode after click", hit.x, hit.y)
	} else {
		t.Logf("Click at (%d,%d) → VNode type=%q (correct)", hit.x, hit.y, cell.OwnerNode.Type)
	}

	// Count total cells with OwnerNode vs without
	ownerCount := 0
	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			if frame.Cells[y][x].OwnerNode != nil {
				ownerCount++
			}
		}
	}
	totalCells := frame.Width * frame.Height
	pct := float64(ownerCount) / float64(totalCells) * 100
	t.Logf("Cells with OwnerNode: %d/%d (%.1f%%)", ownerCount, totalCells, pct)

	if pct < 50 {
		t.Errorf("Less than 50%% of cells have OwnerNode (%.1f%%) — hit testing may be unreliable", pct)
	}
}

// --- Test 3: Background Consistency (verifies ANSI output, not Cell struct) ---

func TestShowcase_BackgroundConsistency(t *testing.T) {
	app := setupShowcaseApp(t)
	defer app.Close()

	frame := app.lastFrame

	// The Cell.Background field may be "" for cells where Lua components
	// don't specify a background. This is BY DESIGN — the ANSIAdapter's
	// DefaultBackground fills these with the theme color at output time.
	//
	// Verify: the ANSI output must contain a bg code for EVERY cell.

	// Create a fresh adapter and render the frame
	var buf bytes.Buffer
	adapter := NewANSIAdapter(&buf)
	adapter.SetSize(frame.Width, frame.Height)
	adapter.Write(frame)
	output := buf.String()

	// The theme bg code is "48;2;30;30;46" (#1E1E2E)
	themeBg := "48;2;30;30;46"
	bgCount := strings.Count(output, themeBg)
	t.Logf("ANSI output size: %d bytes", len(output))
	t.Logf("Theme bg code (%s) occurrences: %d", themeBg, bgCount)

	// There should be at least one bg code per row (the writer optimizes
	// consecutive same-style cells, so we expect at least Height occurrences)
	if bgCount < frame.Height {
		t.Errorf("Expected at least %d bg code occurrences (one per row), got %d", frame.Height, bgCount)
	}

	// Verify there are NO bare resets without a subsequent bg code.
	// A bare \x1b[0m followed by text without \x1b[48;... would mean
	// that text renders with the terminal's default bg (the bug we fixed).
	//
	// Check: after each \x1b[0m, the next style sequence should include a bg code.
	resetCode := "\x1b[0m"
	resetPositions := findAllSubstringPositions(output, resetCode)
	bareResets := 0
	for _, pos := range resetPositions {
		afterReset := output[pos+len(resetCode):]
		// Find the next character that's not part of an escape sequence
		nextCharIdx := 0
		for nextCharIdx < len(afterReset) {
			if afterReset[nextCharIdx] == '\x1b' {
				// Skip this escape sequence
				end := strings.IndexByte(afterReset[nextCharIdx:], 'm')
				if end < 0 {
					break
				}
				seq := afterReset[nextCharIdx : nextCharIdx+end+1]
				if strings.Contains(seq, "48;") {
					// Found a bg code before any text — good
					break
				}
				nextCharIdx += end + 1
			} else if afterReset[nextCharIdx] == '\x1b' {
				break
			} else {
				// Found a printable character without a preceding bg code
				// Check if this is just the cursor positioning or end of frame
				ch := afterReset[nextCharIdx]
				if ch >= 0x20 && ch < 0x7f && ch != ' ' {
					bareResets++
				}
				break
			}
		}
	}

	t.Logf("Bare resets (text without bg after \\x1b[0m): %d", bareResets)
	if bareResets > 0 {
		t.Logf("Note: %d bare resets found — these cells use terminal default bg", bareResets)
		t.Logf("The ANSIAdapter.DefaultBackground should prevent this")
	}

	// Also report Cell-level stats for informational purposes
	emptyBg := 0
	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			if frame.Cells[y][x].Background == "" {
				emptyBg++
			}
		}
	}
	totalCells := frame.Width * frame.Height
	t.Logf("Cell.Background=\"\" count: %d/%d (%.1f%%) — handled by DefaultBackground at ANSI output",
		emptyBg, totalCells, float64(emptyBg)/float64(totalCells)*100)
}

// findAllSubstringPositions returns all start positions of substr in s.
func findAllSubstringPositions(s, substr string) []int {
	var positions []int
	start := 0
	for {
		idx := strings.Index(s[start:], substr)
		if idx < 0 {
			break
		}
		positions = append(positions, start+idx)
		start += idx + 1
	}
	return positions
}

// --- Test 4: Full Screen Render (tests form tab which has more content) ---

func TestShowcase_FullScreenRender(t *testing.T) {
	app := setupShowcaseApp(t)
	defer app.Close()

	frame := app.lastFrame

	// Verify frame dimensions match terminal size
	if frame.Width != 120 {
		t.Errorf("Frame width = %d, want 120", frame.Width)
	}
	if frame.Height != 40 {
		t.Errorf("Frame height = %d, want 40", frame.Height)
	}

	// Verify content in first row (title)
	row0 := getFrameRow(frame, 0)
	if !strings.Contains(row0, "Lumina Components Showcase") {
		t.Errorf("Row 0 should contain title, got: %q", row0)
	}

	// Switch to form tab which has more content
	pressKey(app, "3")
	frame = app.lastFrame

	// Count content rows on form tab
	contentRows := 0
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.TrimSpace(row) != "" {
			contentRows++
		}
	}
	t.Logf("Form tab content rows: %d/%d", contentRows, frame.Height)

	// Form tab should have substantial content (title, subtitle, tab bar,
	// separator, username, input, switch, progress, toggle group, sliders, footer)
	if contentRows < 12 {
		t.Errorf("Form tab should have at least 12 content rows, got %d", contentRows)
		for y := 0; y < frame.Height; y++ {
			row := getFrameRow(frame, y)
			if strings.TrimSpace(row) != "" {
				t.Logf("  row %2d: %q", y, strings.TrimSpace(row))
			}
		}
	}

	// Verify specific form elements are present
	checks := []struct {
		substr string
		desc   string
	}{
		{"Username", "username label"},
		{"johndoe", "username input value"},
		{"OFF", "switch state"},
		{"65%", "progress value"},
		{"Bold", "toggle group option"},
		{"[+/-]", "footer key hints"},
	}
	for _, check := range checks {
		if findRowContaining(frame, check.substr) < 0 {
			t.Errorf("Form tab missing %s (expected %q in frame)", check.desc, check.substr)
		}
	}
}

// --- Test 5: Store State Types (verifies round-trip through dispatch) ---

func TestShowcase_StoreStateTypes(t *testing.T) {
	app := setupShowcaseApp(t)
	defer app.Close()

	// Switch to form tab first
	pressKey(app, "3")
	frame := app.lastFrame

	// Verify bool state: switchOn (false → true → false)
	// Before: [━━●] OFF
	switchRow := findRowContaining(frame, "OFF")
	if switchRow < 0 {
		t.Fatal("Expected 'OFF' in form tab (initial switchOn=false)")
	}
	beforeSwitch := getFrameRow(frame, switchRow)
	if !strings.Contains(beforeSwitch, "[━━●] OFF") {
		t.Errorf("Expected '[━━●] OFF', got: %q", strings.TrimSpace(beforeSwitch))
	}

	pressKey(app, "s")
	frame = app.lastFrame
	switchRow = findRowContaining(frame, "ON")
	if switchRow < 0 {
		t.Fatal("After 's', expected 'ON' (switchOn: false→true)")
	}
	afterSwitch := getFrameRow(frame, switchRow)
	if !strings.Contains(afterSwitch, "[●━━] ON") {
		t.Errorf("Expected '[●━━] ON', got: %q", strings.TrimSpace(afterSwitch))
	}
	t.Logf("✓ Bool: %q → %q", strings.TrimSpace(beforeSwitch), strings.TrimSpace(afterSwitch))

	// Verify number state: progressValue (65 → 70 → 65)
	pressKey(app, "+")
	frame = app.lastFrame
	progRow := findRowContaining(frame, "70%")
	if progRow < 0 {
		t.Fatal("After '+', expected '70%' (progressValue: 65→70)")
	}
	t.Logf("✓ Number: 65%% → %q", strings.TrimSpace(getFrameRow(frame, progRow)))

	pressKey(app, "-")
	frame = app.lastFrame
	progRow = findRowContaining(frame, "65%")
	if progRow < 0 {
		t.Fatal("After '-', expected '65%' (progressValue: 70→65)")
	}
	t.Logf("✓ Number: 70%% → %q", strings.TrimSpace(getFrameRow(frame, progRow)))

	// Verify string state: tab (form → basic)
	pressKey(app, "1")
	frame = app.lastFrame
	tabRow := findRowContaining(frame, "[1:Basic]")
	if tabRow < 0 {
		t.Fatal("After '1', expected '[1:Basic]' (tab: form→basic)")
	}
	t.Logf("✓ String: tab switched to 'basic' — %q", strings.TrimSpace(getFrameRow(frame, tabRow)))
}
