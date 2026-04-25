package lumina

import "unicode"

// RuneWidth returns the number of terminal columns a rune occupies.
// Most characters are 1 column. CJK, emoji, and other wide chars are 2 columns.
func RuneWidth(r rune) int {
	if r == 0 {
		return 0
	}
	// Control characters
	if r < 32 || r == 0x7f {
		return 0
	}
	// Zero-width joiner
	if r == 0x200D {
		return 0
	}
	// Variation selectors
	if r >= 0xFE00 && r <= 0xFE0F {
		return 0
	}
	// Combining marks
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mc, r) {
		return 0
	}

	// Emoji ranges (most emoji are 2 columns wide in modern terminals)
	if r >= 0x1F300 && r <= 0x1F9FF {
		return 2
	} // Misc Symbols, Emoticons, Supplemental Symbols
	if r >= 0x1FA00 && r <= 0x1FA6F {
		return 2
	} // Chess Symbols, Extended-A
	if r >= 0x1FA70 && r <= 0x1FAFF {
		return 2
	} // Symbols Extended-A
	if r >= 0x2702 && r <= 0x27B0 {
		return 2
	} // Dingbats subset (✂️ etc.)
	if r >= 0x2934 && r <= 0x2935 {
		return 2
	} // Arrows
	if r >= 0x231A && r <= 0x231B {
		return 2
	} // Watch, Hourglass
	if r >= 0x23E9 && r <= 0x23F3 {
		return 2
	} // Media controls
	if r >= 0x25AA && r <= 0x25AB {
		return 2
	} // Small squares
	if r >= 0x25FB && r <= 0x25FE {
		return 2
	} // Medium squares
	if r >= 0x2614 && r <= 0x2615 {
		return 2
	} // Umbrella, Hot Beverage
	if r >= 0x2648 && r <= 0x2653 {
		return 2
	} // Zodiac
	if r == 0x267F || r == 0x2693 || r == 0x26A1 {
		return 2
	} // Wheelchair, Anchor, Lightning
	if r >= 0x26AA && r <= 0x26AB {
		return 2
	} // Circles
	if r >= 0x26BD && r <= 0x26BE {
		return 2
	} // Soccer, Baseball
	if r >= 0x26C4 && r <= 0x26C5 {
		return 2
	} // Snowman, Sun
	if r == 0x26D4 || r == 0x26EA || r == 0x26F2 || r == 0x26F3 {
		return 2
	}
	if r == 0x26F5 || r == 0x26FA || r == 0x26FD {
		return 2
	}
	if r == 0x2660 || r == 0x2663 || r == 0x2665 || r == 0x2666 {
		return 2
	} // Card suits
	if r >= 0x2190 && r <= 0x21FF {
		return 1
	} // Arrows (1-wide)

	// East Asian Wide/Fullwidth
	if unicode.Is(unicode.Han, r) {
		return 2
	}
	if r >= 0x1100 && r <= 0x115F {
		return 2
	} // Hangul Jamo
	if r >= 0x2E80 && r <= 0x303E {
		return 2
	} // CJK Radicals, Kangxi, CJK Symbols
	if r >= 0x3041 && r <= 0x33BF {
		return 2
	} // Hiragana, Katakana, Bopomofo, CJK Compat
	if r >= 0x3400 && r <= 0x4DBF {
		return 2
	} // CJK Unified Extension A
	if r >= 0x4E00 && r <= 0x9FFF {
		return 2
	} // CJK Unified
	if r >= 0xA000 && r <= 0xA4CF {
		return 2
	} // Yi
	if r >= 0xAC00 && r <= 0xD7AF {
		return 2
	} // Hangul Syllables
	if r >= 0xF900 && r <= 0xFAFF {
		return 2
	} // CJK Compatibility Ideographs
	if r >= 0xFE10 && r <= 0xFE6F {
		return 2
	} // CJK Forms
	if r >= 0xFF01 && r <= 0xFF60 {
		return 2
	} // Fullwidth Forms
	if r >= 0xFFE0 && r <= 0xFFE6 {
		return 2
	} // Fullwidth Signs
	if r >= 0x20000 && r <= 0x2FFFF {
		return 2
	} // CJK Extension B+
	if r >= 0x30000 && r <= 0x3FFFF {
		return 2
	} // CJK Extension G+

	return 1
}

// StringWidth returns the display width of a string in terminal columns.
func StringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += RuneWidth(r)
	}
	return w
}
