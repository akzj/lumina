//go:build linux || darwin

package lumina

import (
	"bytes"
	"os"
	"testing"
)

func TestDetectColorMode_TrueColor(t *testing.T) {
	// Save and restore env.
	origCT := os.Getenv("COLORTERM")
	origTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("COLORTERM", origCT)
		os.Setenv("TERM", origTerm)
	}()

	os.Setenv("COLORTERM", "truecolor")
	os.Setenv("TERM", "xterm")
	if got := DetectColorMode(); got != ColorTrue {
		t.Errorf("COLORTERM=truecolor: got %v, want ColorTrue", got)
	}

	os.Setenv("COLORTERM", "24bit")
	if got := DetectColorMode(); got != ColorTrue {
		t.Errorf("COLORTERM=24bit: got %v, want ColorTrue", got)
	}
}

func TestDetectColorMode_256(t *testing.T) {
	origCT := os.Getenv("COLORTERM")
	origTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("COLORTERM", origCT)
		os.Setenv("TERM", origTerm)
	}()

	os.Setenv("COLORTERM", "")
	os.Setenv("TERM", "xterm-256color")
	if got := DetectColorMode(); got != Color256 {
		t.Errorf("TERM=xterm-256color: got %v, want Color256", got)
	}
}

func TestDetectColorMode_16(t *testing.T) {
	origCT := os.Getenv("COLORTERM")
	origTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("COLORTERM", origCT)
		os.Setenv("TERM", origTerm)
	}()

	os.Setenv("COLORTERM", "")
	os.Setenv("TERM", "xterm")
	if got := DetectColorMode(); got != Color16 {
		t.Errorf("TERM=xterm: got %v, want Color16", got)
	}
}

func TestDetectColorMode_None(t *testing.T) {
	origCT := os.Getenv("COLORTERM")
	origTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("COLORTERM", origCT)
		os.Setenv("TERM", origTerm)
	}()

	os.Setenv("COLORTERM", "")

	os.Setenv("TERM", "dumb")
	if got := DetectColorMode(); got != ColorNone {
		t.Errorf("TERM=dumb: got %v, want ColorNone", got)
	}

	os.Setenv("TERM", "")
	if got := DetectColorMode(); got != ColorNone {
		t.Errorf("TERM='': got %v, want ColorNone", got)
	}
}

func TestColorModeString(t *testing.T) {
	tests := []struct {
		mode ColorMode
		want string
	}{
		{ColorNone, "none"},
		{Color16, "16"},
		{Color256, "256"},
		{ColorTrue, "truecolor"},
		{ColorMode(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("ColorMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestRgbTo256(t *testing.T) {
	// Pure black → 16 (start of 6x6x6 cube)
	if got := rgbTo256(0, 0, 0); got != 16 {
		t.Errorf("rgbTo256(0,0,0) = %d, want 16", got)
	}
	// Pure white → 231
	if got := rgbTo256(255, 255, 255); got != 231 {
		t.Errorf("rgbTo256(255,255,255) = %d, want 231", got)
	}
	// Gray → grayscale ramp range (232-255)
	got := rgbTo256(128, 128, 128)
	if got < 232 || got > 255 {
		t.Errorf("rgbTo256(128,128,128) = %d, want in [232,255]", got)
	}
}

func TestRgbTo16(t *testing.T) {
	// Black → 30 (black fg)
	if got := rgbTo16(0, 0, 0); got != 30 {
		t.Errorf("rgbTo16(0,0,0) = %d, want 30", got)
	}
	// White → bright white (90+7=97)
	if got := rgbTo16(255, 255, 255); got != 97 {
		t.Errorf("rgbTo16(255,255,255) = %d, want 97", got)
	}
}

func TestHexToRGB(t *testing.T) {
	r, g, b := hexToRGB("#FF8000")
	if r != 255 || g != 128 || b != 0 {
		t.Errorf("hexToRGB(#FF8000) = (%d,%d,%d), want (255,128,0)", r, g, b)
	}
}

func TestANSIAdapterColorMode(t *testing.T) {
	var buf bytes.Buffer
	a := NewANSIAdapterWithColorMode(&buf, Color256)
	if a.ColorMode() != Color256 {
		t.Errorf("got %v, want Color256", a.ColorMode())
	}
	a.SetColorMode(ColorTrue)
	if a.ColorMode() != ColorTrue {
		t.Errorf("after SetColorMode: got %v, want ColorTrue", a.ColorMode())
	}
}

func TestANSIAdapterColorNone_NoEscapes(t *testing.T) {
	var buf bytes.Buffer
	a := NewANSIAdapterWithColorMode(&buf, ColorNone)
	cell := &Cell{Char: 'X', Foreground: "#FF0000", Bold: true}
	codes := a.styleCodes(cell)
	// ColorNone should only emit reset, no color or attribute codes.
	if codes != "\x1b[0m" {
		t.Errorf("ColorNone styleCodes = %q, want only reset", codes)
	}
}

func TestANSIAdapterTrueColor_HexOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewANSIAdapterWithColorMode(&buf, ColorTrue)
	cell := &Cell{Char: 'X', Foreground: "#FF8000"}
	codes := a.styleCodes(cell)
	// Should contain 38;2;255;128;0
	expected := "\x1b[38;2;255;128;0m"
	if !bytes.Contains([]byte(codes), []byte(expected)) {
		t.Errorf("ColorTrue styleCodes = %q, want to contain %q", codes, expected)
	}
}

func TestANSIAdapter256_HexOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewANSIAdapterWithColorMode(&buf, Color256)
	cell := &Cell{Char: 'X', Foreground: "#FF0000"}
	codes := a.styleCodes(cell)
	// Should contain 38;5; (256-color escape)
	if !bytes.Contains([]byte(codes), []byte("38;5;")) {
		t.Errorf("Color256 styleCodes = %q, want to contain '38;5;'", codes)
	}
}

func TestANSIAdapter16_HexOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewANSIAdapterWithColorMode(&buf, Color16)
	cell := &Cell{Char: 'X', Foreground: "#FF0000"}
	codes := a.styleCodes(cell)
	// Should NOT contain 38;2; or 38;5; (no true/256 color)
	if bytes.Contains([]byte(codes), []byte("38;2;")) || bytes.Contains([]byte(codes), []byte("38;5;")) {
		t.Errorf("Color16 styleCodes = %q, should not contain 38;2; or 38;5;", codes)
	}
}
