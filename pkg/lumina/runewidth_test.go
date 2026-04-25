package lumina_test

import (
	"bytes"
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

func TestRuneWidth_ASCII(t *testing.T) {
	tests := []struct {
		r    rune
		want int
	}{
		{'A', 1},
		{'z', 1},
		{'0', 1},
		{' ', 1},
		{'!', 1},
		{'~', 1},
	}
	for _, tt := range tests {
		got := lumina.RuneWidth(tt.r)
		if got != tt.want {
			t.Errorf("RuneWidth(%q) = %d, want %d", tt.r, got, tt.want)
		}
	}
}

func TestRuneWidth_Emoji(t *testing.T) {
	tests := []struct {
		r    rune
		want int
	}{
		{'🏠', 2},
		{'👥', 2},
		{'🔧', 2},
		{'🌟', 2},
		{'🤖', 2},
		{'⚙', 1}, // U+2699 - varies by terminal, 1 is safe default
	}
	for _, tt := range tests {
		got := lumina.RuneWidth(tt.r)
		if got != tt.want {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want %d", tt.r, tt.r, got, tt.want)
		}
	}
}

func TestRuneWidth_CJK(t *testing.T) {
	tests := []struct {
		r    rune
		want int
	}{
		{'中', 2},
		{'文', 2},
		{'日', 2},
		{'本', 2},
		{'語', 2},
		{'한', 2}, // Hangul
	}
	for _, tt := range tests {
		got := lumina.RuneWidth(tt.r)
		if got != tt.want {
			t.Errorf("RuneWidth(%q) = %d, want %d", tt.r, got, tt.want)
		}
	}
}

func TestRuneWidth_Control(t *testing.T) {
	if w := lumina.RuneWidth(0); w != 0 {
		t.Errorf("RuneWidth(0) = %d, want 0", w)
	}
	if w := lumina.RuneWidth('\t'); w != 0 {
		t.Errorf("RuneWidth(tab) = %d, want 0", w)
	}
	if w := lumina.RuneWidth(0x7f); w != 0 {
		t.Errorf("RuneWidth(DEL) = %d, want 0", w)
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"Hello", 5},
		{"🏠 Home", 7},   // 2 + 1 + 4 = 7
		{"中文", 4},        // 2 + 2 = 4
		{"A🌟B", 4},       // 1 + 2 + 1 = 4
		{"", 0},
		{"abc", 3},
	}
	for _, tt := range tests {
		got := lumina.StringWidth(tt.s)
		if got != tt.want {
			t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestRenderText_WideChar(t *testing.T) {
	app := lumina.NewApp()
	defer app.Close()

	var buf bytes.Buffer
	tio := lumina.NewBufferTermIO(40, 10, &buf)
	lumina.SetOutputAdapter(lumina.NewANSIAdapter(tio))

	err := app.L.DoString(`
		local lumina = require("lumina")
		local App = lumina.defineComponent({
			name = "WideCharApp",
			render = function(self)
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "🏠 Home" },
						{ type = "text", content = "Normal" },
					}
				}
			end
		})
		lumina.mount(App)
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	app.RenderOnce()
	output := stripANSICodes(buf.String())

	// "Normal" should appear on its own line, not shifted
	if len(output) == 0 {
		t.Fatal("no output produced")
	}

	// Verify both texts appear in output
	if !containsSubstring(output, "Home") {
		t.Error("output should contain 'Home'")
	}
	if !containsSubstring(output, "Normal") {
		t.Error("output should contain 'Normal'")
	}
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
