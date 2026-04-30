package render

import "testing"

// --- parsePercent tests ---

func TestParsePercent(t *testing.T) {
	tests := []struct {
		input string
		val   int
		ok    bool
	}{
		{"50%", 50, true},
		{"100%", 100, true},
		{"0%", 0, true},
		{"25%", 25, true},
		{" 50%", 50, true}, // TrimSpace handles leading space
		{"50", 0, false},
		{"abc%", 0, false},
		{"", 0, false},
		{"-10%", 0, false},
	}
	for _, tt := range tests {
		v, ok := parsePercent(tt.input)
		if ok != tt.ok || v != tt.val {
			t.Errorf("parsePercent(%q) = (%d, %v), want (%d, %v)", tt.input, v, ok, tt.val, tt.ok)
		}
	}
}

func TestParseViewport(t *testing.T) {
	tests := []struct {
		input string
		val   int
		unit  string
		ok    bool
	}{
		{"50vw", 50, "vw", true},
		{"100vh", 100, "vh", true},
		{"0vw", 0, "vw", true},
		{"25vh", 25, "vh", true},
		{"50", 0, "", false},
		{"50%", 0, "", false},
		{"abc", 0, "", false},
		{"", 0, "", false},
	}
	for _, tt := range tests {
		v, unit, ok := parseViewport(tt.input)
		if ok != tt.ok || v != tt.val || unit != tt.unit {
			t.Errorf("parseViewport(%q) = (%d, %q, %v), want (%d, %q, %v)",
				tt.input, v, unit, ok, tt.val, tt.unit, tt.ok)
		}
	}
}

// --- Percentage layout tests ---

func TestPercentWidth(t *testing.T) {
	// Parent has 80 cols, child has width="50%" → child.W = 40
	root := NewNode("box")
	child := NewNode("box")
	child.Style.WidthPercent = 50
	child.Style.Height = 5
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	if child.W != 40 {
		t.Errorf("child.W = %d, want 40 (50%% of 80)", child.W)
	}
}

func TestPercentHeight(t *testing.T) {
	// Parent has 24 rows, child has height="50%" → child.H = 12
	root := NewNode("box")
	child := NewNode("box")
	child.Style.HeightPercent = 50
	child.Style.Width = 10
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	if child.H != 12 {
		t.Errorf("child.H = %d, want 12 (50%% of 24)", child.H)
	}
}

func TestPercentWidthAndHeight(t *testing.T) {
	// Both width and height as percentages
	root := NewNode("box")
	child := NewNode("box")
	child.Style.WidthPercent = 25
	child.Style.HeightPercent = 75
	root.AddChild(child)

	LayoutFull(root, 0, 0, 100, 40)

	if child.W != 25 {
		t.Errorf("child.W = %d, want 25 (25%% of 100)", child.W)
	}
	if child.H != 30 {
		t.Errorf("child.H = %d, want 30 (75%% of 40)", child.H)
	}
}

func TestPercentPrecedence(t *testing.T) {
	// Absolute width takes precedence over WidthPercent
	root := NewNode("box")
	child := NewNode("box")
	child.Style.Width = 20
	child.Style.WidthPercent = 50 // should be ignored because Width > 0
	child.Style.Height = 5
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	if child.W != 20 {
		t.Errorf("child.W = %d, want 20 (absolute takes precedence)", child.W)
	}
}

func TestPercentMinMax(t *testing.T) {
	// MinWidthPercent and MaxWidthPercent without explicit Width
	root := NewNode("box")
	child := NewNode("box")
	child.Style.MinWidthPercent = 25 // 25% of 100 = 25
	child.Style.MaxWidthPercent = 75 // 75% of 100 = 75
	child.Style.Height = 5
	child.Style.WidthPercent = 90 // 90% of 100 = 90, exceeds max → clamped to 75
	root.AddChild(child)

	LayoutFull(root, 0, 0, 100, 24)

	if child.W != 75 {
		t.Errorf("child.W = %d, want 75 (90%% clamped by maxWidth 75%%)", child.W)
	}
}

func TestPercentMinWidthClamp(t *testing.T) {
	// WidthPercent=10 but MinWidthPercent=25 (25% of 100 = 25) → clamped up to 25
	root := NewNode("box")
	child := NewNode("box")
	child.Style.WidthPercent = 10 // 10% of 100 = 10
	child.Style.MinWidthPercent = 25
	child.Style.Height = 5
	root.AddChild(child)

	LayoutFull(root, 0, 0, 100, 24)

	if child.W != 25 {
		t.Errorf("child.W = %d, want 25 (10%% clamped up by minWidth 25%%)", child.W)
	}
}

func TestViewportWidth(t *testing.T) {
	// Child with width="50vw" in viewport 80 → child.W = 40
	root := NewNode("box")
	child := NewNode("box")
	child.Style.WidthVW = 50
	child.Style.Height = 5
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	if child.W != 40 {
		t.Errorf("child.W = %d, want 40 (50vw of viewport 80)", child.W)
	}
}

func TestViewportHeight(t *testing.T) {
	// Child with height="50vh" in viewport 24 → child.H = 12
	root := NewNode("box")
	child := NewNode("box")
	child.Style.HeightVH = 50
	child.Style.Width = 10
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	if child.H != 12 {
		t.Errorf("child.H = %d, want 12 (50vh of viewport 24)", child.H)
	}
}

func TestViewportInNestedChild(t *testing.T) {
	// Viewport units should use root dimensions, not parent dimensions
	root := NewNode("box")
	root.Style.Width = 80
	root.Style.Height = 24

	parent := NewNode("box")
	parent.Style.Width = 40
	parent.Style.Height = 12
	root.AddChild(parent)

	child := NewNode("box")
	child.Style.WidthVW = 50  // 50% of viewport (80) = 40, not 50% of parent (40) = 20
	child.Style.HeightVH = 50 // 50% of viewport (24) = 12, not 50% of parent (12) = 6
	parent.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	if child.W != 40 {
		t.Errorf("child.W = %d, want 40 (50vw uses viewport 80, not parent 40)", child.W)
	}
	if child.H != 12 {
		t.Errorf("child.H = %d, want 12 (50vh uses viewport 24, not parent 12)", child.H)
	}
}

func TestPercentInHBox(t *testing.T) {
	// In an hbox, child with width="30%" of parent contentW=100 → child.W = 30
	root := NewNode("hbox")
	child1 := NewNode("box")
	child1.Style.WidthPercent = 30
	root.AddChild(child1)

	child2 := NewNode("box")
	child2.Style.WidthPercent = 70
	root.AddChild(child2)

	LayoutFull(root, 0, 0, 100, 24)

	if child1.W != 30 {
		t.Errorf("child1.W = %d, want 30 (30%% of 100)", child1.W)
	}
	if child2.W != 70 {
		t.Errorf("child2.W = %d, want 70 (70%% of 100)", child2.W)
	}
}

func TestPercentInVBox(t *testing.T) {
	// In a vbox, child with height="25%" of parent contentH=40 → child.H = 10
	root := NewNode("vbox")
	child := NewNode("box")
	child.Style.HeightPercent = 25
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 40)

	if child.H != 10 {
		t.Errorf("child.H = %d, want 10 (25%% of 40)", child.H)
	}
}

func TestPercentInScrollContainer(t *testing.T) {
	// Child in scroll container with height="50%" uses parent contentH
	root := NewNode("vbox")
	root.Style.Overflow = "scroll"

	child := NewNode("box")
	child.Style.HeightPercent = 50
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 20)

	// In scroll container, contentH is 20 (full parent), child gets 50% = 10
	if child.H != 10 {
		t.Errorf("child.H = %d, want 10 (50%% of 20 in scroll container)", child.H)
	}
}

func TestPercentWithBorder(t *testing.T) {
	// Parent with border reduces content area
	root := NewNode("box")
	root.Style.Border = "single" // 1 cell border on each side

	child := NewNode("box")
	child.Style.WidthPercent = 50
	child.Style.Height = 3
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	// Content area = 80 - 2 (border) = 78
	// 50% of 78 = 39
	if child.W != 39 {
		t.Errorf("child.W = %d, want 39 (50%% of content width 78)", child.W)
	}
}

func TestPercentWithPadding(t *testing.T) {
	// Parent with padding reduces content area
	root := NewNode("box")
	root.Style.PaddingLeft = 5
	root.Style.PaddingRight = 5

	child := NewNode("box")
	child.Style.WidthPercent = 50
	child.Style.Height = 3
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	// Content area = 80 - 10 (padding) = 70
	// 50% of 70 = 35
	if child.W != 35 {
		t.Errorf("child.W = %d, want 35 (50%% of content width 70)", child.W)
	}
}

func TestPercentZeroIsNoOp(t *testing.T) {
	// WidthPercent=0 means not set, should behave as before (flex fills)
	root := NewNode("box")
	child := NewNode("box")
	child.Style.WidthPercent = 0
	child.Style.HeightPercent = 0
	root.AddChild(child)

	LayoutFull(root, 0, 0, 80, 24)

	// Without any size specification, a single box child fills the parent
	if child.W != 80 {
		t.Errorf("child.W = %d, want 80 (no percent set, should fill parent)", child.W)
	}
}

func TestResolveWidthHelper(t *testing.T) {
	// Test resolveWidth function directly
	s := Style{Width: 20}
	if v := resolveWidth(s, 100); v != 20 {
		t.Errorf("resolveWidth with Width=20 = %d, want 20", v)
	}

	s = Style{WidthPercent: 50}
	if v := resolveWidth(s, 100); v != 50 {
		t.Errorf("resolveWidth with WidthPercent=50, parent=100 = %d, want 50", v)
	}

	layoutViewportW = 80
	s = Style{WidthVW: 25}
	if v := resolveWidth(s, 100); v != 20 {
		t.Errorf("resolveWidth with WidthVW=25, viewport=80 = %d, want 20", v)
	}

	s = Style{}
	if v := resolveWidth(s, 100); v != 0 {
		t.Errorf("resolveWidth with no size = %d, want 0", v)
	}
}

func TestResolveHeightHelper(t *testing.T) {
	s := Style{Height: 15}
	if v := resolveHeight(s, 50); v != 15 {
		t.Errorf("resolveHeight with Height=15 = %d, want 15", v)
	}

	s = Style{HeightPercent: 50}
	if v := resolveHeight(s, 50); v != 25 {
		t.Errorf("resolveHeight with HeightPercent=50, parent=50 = %d, want 25", v)
	}

	layoutViewportH = 24
	s = Style{HeightVH: 50}
	if v := resolveHeight(s, 50); v != 12 {
		t.Errorf("resolveHeight with HeightVH=50, viewport=24 = %d, want 12", v)
	}

	s = Style{}
	if v := resolveHeight(s, 50); v != 0 {
		t.Errorf("resolveHeight with no size = %d, want 0", v)
	}
}
