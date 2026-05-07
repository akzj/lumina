package render

import (
	"testing"
)

// TestTextareaLayoutAutoHeight verifies that a textarea inside an hbox
// correctly auto-sizes based on content lines.
func TestTextareaLayoutAutoHeight(t *testing.T) {
	// Build tree: vbox > hbox > [text, textarea]
	// The hbox has minHeight=1, maxHeight=8, no explicit height
	// The textarea has 3 lines of content

	textarea := &Node{
		Type:    "textarea",
		Content: "line1\nline2\nline3",
		Style: Style{
			Flex:      1,
			MinHeight: 1,
			MaxHeight: 8,
		},
	}

	prompt := &Node{
		Type:    "text",
		Content: "> ",
		Style: Style{
			Width: 2,
		},
	}

	hbox := &Node{
		Type:     "hbox",
		Children: []*Node{prompt, textarea},
		Style: Style{
			Width:     80,
			MinHeight: 1,
			MaxHeight: 8,
		},
	}
	prompt.Parent = hbox
	textarea.Parent = hbox

	root := &Node{
		Type:     "vbox",
		Children: []*Node{hbox},
		Style: Style{
			Width:  80,
			Height: 40,
		},
	}
	hbox.Parent = root

	// Run layout
	LayoutFull(root, 0, 0, 80, 40)

	t.Logf("Root: W=%d H=%d", root.W, root.H)
	t.Logf("Hbox: W=%d H=%d", hbox.W, hbox.H)
	t.Logf("Prompt: W=%d H=%d", prompt.W, prompt.H)
	t.Logf("Textarea: W=%d H=%d", textarea.W, textarea.H)

	// The textarea has 3 lines, so it should be height 3
	if textarea.H != 3 {
		t.Errorf("Textarea height should be 3 (3 lines), got %d", textarea.H)
	}

	// The hbox should match its tallest child
	if hbox.H < 3 {
		t.Errorf("Hbox height should be >= 3, got %d", hbox.H)
	}
}

// TestTextareaLayoutSingleLine verifies textarea with no newlines stays at height 1.
func TestTextareaLayoutSingleLine(t *testing.T) {
	textarea := &Node{
		Type:    "textarea",
		Content: "hello world",
		Style: Style{
			Flex:      1,
			MinHeight: 1,
			MaxHeight: 8,
		},
	}

	hbox := &Node{
		Type:     "hbox",
		Children: []*Node{textarea},
		Style: Style{
			Width:     80,
			MinHeight: 1,
			MaxHeight: 8,
		},
	}
	textarea.Parent = hbox

	root := &Node{
		Type:     "vbox",
		Children: []*Node{hbox},
		Style: Style{
			Width:  80,
			Height: 40,
		},
	}
	hbox.Parent = root

	LayoutFull(root, 0, 0, 80, 40)

	t.Logf("Textarea: W=%d H=%d", textarea.W, textarea.H)
	t.Logf("Hbox: W=%d H=%d", hbox.W, hbox.H)

	if textarea.H != 1 {
		t.Errorf("Textarea with single line should be height 1, got %d", textarea.H)
	}
}
