package render

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HighlightLine holds spans for one line of highlighted code.
type HighlightLine struct {
	Spans []Span
}

// Highlight tokenizes code using chroma and returns per-line spans with colors.
// language can be a language name (e.g. "go", "python") or file extension.
// If language is empty, chroma will try to detect it.
// styleName is the chroma style (e.g. "monokai", "dracula"). If empty, uses "monokai".
func Highlight(code, language, styleName string) []HighlightLine {
	// Get lexer
	var lexer chroma.Lexer
	if language != "" {
		lexer = lexers.Get(language)
	}
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Get style
	if styleName == "" {
		styleName = "monokai"
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// Fallback: return plain text spans per line
		return plainLines(code)
	}

	// Build per-line spans
	var result []HighlightLine
	var currentLine []Span

	for _, token := range iterator.Tokens() {
		entry := style.Get(token.Type)
		fg := ""
		if entry.Colour.IsSet() {
			fg = "#" + entry.Colour.String()
		}
		bold := entry.Bold == chroma.Yes
		italic := entry.Italic == chroma.Yes
		underline := entry.Underline == chroma.Yes

		// Split token text by newlines
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				// New line
				result = append(result, HighlightLine{Spans: currentLine})
				currentLine = nil
			}
			if part != "" {
				span := Span{
					Text:       part,
					Foreground: fg,
				}
				if bold {
					b := true
					span.Bold = &b
				}
				if italic {
					b := true
					span.Italic = &b
				}
				if underline {
					b := true
					span.Underline = &b
				}
				currentLine = append(currentLine, span)
			}
		}
	}
	// Don't forget the last line
	if currentLine != nil {
		result = append(result, HighlightLine{Spans: currentLine})
	}

	return result
}

func plainLines(code string) []HighlightLine {
	lines := strings.Split(code, "\n")
	result := make([]HighlightLine, len(lines))
	for i, line := range lines {
		if line != "" {
			result[i] = HighlightLine{Spans: []Span{{Text: line}}}
		}
	}
	return result
}
