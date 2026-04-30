package render

// style_parse.go — Canonical style field parsing.
// Each style property is defined in ONE place (applyStyleField).
// styleFromMap uses this. readStyle/readStyleFields use Lua-specific
// accessors for performance but follow the same property names.

// toInt converts any numeric value to int.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case int:
		return n, true
	}
	return 0, false
}

// applyStyleField sets a single style field by name from a generic value.
// This is the canonical definition of all style property names and their types.
func applyStyleField(s *Style, name string, value any) {
	switch name {
	// --- Sizing ---
	case "width":
		if n, ok := toInt(value); ok {
			s.Width = n
		} else if str, ok := value.(string); ok && str != "" {
			if pct, pok := parsePercent(str); pok {
				s.WidthPercent = pct
			} else if v, unit, vok := parseViewport(str); vok && unit == "vw" {
				s.WidthVW = v
			}
		}
	case "height":
		if n, ok := toInt(value); ok {
			s.Height = n
		} else if str, ok := value.(string); ok && str != "" {
			if pct, pok := parsePercent(str); pok {
				s.HeightPercent = pct
			} else if v, unit, vok := parseViewport(str); vok && unit == "vh" {
				s.HeightVH = v
			}
		}
	case "minWidth":
		if n, ok := toInt(value); ok {
			s.MinWidth = n
		} else if str, ok := value.(string); ok && str != "" {
			if pct, pok := parsePercent(str); pok {
				s.MinWidthPercent = pct
			}
		}
	case "maxWidth":
		if n, ok := toInt(value); ok {
			s.MaxWidth = n
		} else if str, ok := value.(string); ok && str != "" {
			if pct, pok := parsePercent(str); pok {
				s.MaxWidthPercent = pct
			}
		}
	case "minHeight":
		if n, ok := toInt(value); ok {
			s.MinHeight = n
		} else if str, ok := value.(string); ok && str != "" {
			if pct, pok := parsePercent(str); pok {
				s.MinHeightPercent = pct
			}
		}
	case "maxHeight":
		if n, ok := toInt(value); ok {
			s.MaxHeight = n
		} else if str, ok := value.(string); ok && str != "" {
			if pct, pok := parsePercent(str); pok {
				s.MaxHeightPercent = pct
			}
		}

	// --- Flex ---
	case "flex":
		if n, ok := toInt(value); ok {
			s.Flex = n
		}
	case "flexShrink":
		if n, ok := toInt(value); ok {
			s.FlexShrink = n
		}
	case "flexBasis":
		if n, ok := toInt(value); ok {
			s.FlexBasis = n
		}
	case "flexWrap":
		if str, ok := value.(string); ok {
			s.FlexWrap = str
		}

	// --- Spacing ---
	case "padding":
		if n, ok := toInt(value); ok {
			s.Padding = n
		}
	case "paddingTop":
		if n, ok := toInt(value); ok {
			s.PaddingTop = n
		}
	case "paddingBottom":
		if n, ok := toInt(value); ok {
			s.PaddingBottom = n
		}
	case "paddingLeft":
		if n, ok := toInt(value); ok {
			s.PaddingLeft = n
		}
	case "paddingRight":
		if n, ok := toInt(value); ok {
			s.PaddingRight = n
		}
	case "margin":
		if n, ok := toInt(value); ok {
			s.Margin = n
		}
	case "marginTop":
		if n, ok := toInt(value); ok {
			s.MarginTop = n
		}
	case "marginBottom":
		if n, ok := toInt(value); ok {
			s.MarginBottom = n
		}
	case "marginLeft":
		if n, ok := toInt(value); ok {
			s.MarginLeft = n
		}
	case "marginRight":
		if n, ok := toInt(value); ok {
			s.MarginRight = n
		}
	case "gap":
		if n, ok := toInt(value); ok {
			s.Gap = n
		}

	// --- Alignment ---
	case "justify":
		if str, ok := value.(string); ok {
			s.Justify = str
		}
	case "align":
		if str, ok := value.(string); ok {
			s.Align = str
		}
	case "alignSelf":
		if str, ok := value.(string); ok {
			s.AlignSelf = str
		}
	case "order":
		if n, ok := toInt(value); ok {
			s.Order = n
		}

	// --- Visual ---
	case "border":
		if str, ok := value.(string); ok {
			s.Border = str
		}
	case "borderColor":
		if str, ok := value.(string); ok {
			s.BorderColor = str
		}
	case "foreground", "fg":
		if str, ok := value.(string); ok {
			s.Foreground = str
		}
	case "background", "bg":
		if str, ok := value.(string); ok {
			s.Background = str
		}
	case "bold":
		if b, ok := value.(bool); ok {
			s.Bold = b
		}
	case "dim":
		if b, ok := value.(bool); ok {
			s.Dim = b
		}
	case "underline":
		if b, ok := value.(bool); ok {
			s.Underline = b
		}
	case "italic":
		if b, ok := value.(bool); ok {
			s.Italic = b
		}
	case "strikethrough":
		if b, ok := value.(bool); ok {
			s.Strikethrough = b
		}
	case "inverse":
		if b, ok := value.(bool); ok {
			s.Inverse = b
		}

	// --- Text ---
	case "textAlign":
		if str, ok := value.(string); ok {
			s.TextAlign = str
		}
	case "textOverflow":
		if str, ok := value.(string); ok {
			s.TextOverflow = str
		}
	case "whiteSpace":
		if str, ok := value.(string); ok {
			s.WhiteSpace = str
		}

	// --- Display/Visibility ---
	case "display":
		if str, ok := value.(string); ok {
			s.Display = str
		}
	case "visibility":
		if str, ok := value.(string); ok {
			s.Visibility = str
		}

	// --- Overflow ---
	case "overflow":
		if str, ok := value.(string); ok {
			s.Overflow = str
		}

	// --- Positioning ---
	case "position":
		if str, ok := value.(string); ok {
			s.Position = str
		}
	case "top":
		if n, ok := toInt(value); ok {
			s.Top = n
		}
	case "left":
		if n, ok := toInt(value); ok {
			s.Left = n
		}
	case "right":
		if n, ok := toInt(value); ok {
			s.Right = n
		}
	case "bottom":
		if n, ok := toInt(value); ok {
			s.Bottom = n
		}
	case "zIndex":
		if n, ok := toInt(value); ok {
			s.ZIndex = n
		}

	// --- Grid container ---
	case "gridTemplateColumns":
		if str, ok := value.(string); ok {
			s.GridTemplateColumns = str
		}
	case "gridTemplateRows":
		if str, ok := value.(string); ok {
			s.GridTemplateRows = str
		}
	case "gridColumnGap":
		if n, ok := toInt(value); ok {
			s.GridColumnGap = n
		}
	case "gridRowGap":
		if n, ok := toInt(value); ok {
			s.GridRowGap = n
		}

	// --- Grid item ---
	case "gridColumn":
		if str, ok := value.(string); ok {
			s.GridColumn = str
		}
	case "gridRow":
		if str, ok := value.(string); ok {
			s.GridRow = str
		}
	case "gridColumnStart":
		if n, ok := toInt(value); ok {
			s.GridColumnStart = n
		}
	case "gridColumnEnd":
		if n, ok := toInt(value); ok {
			s.GridColumnEnd = n
		}
	case "gridRowStart":
		if n, ok := toInt(value); ok {
			s.GridRowStart = n
		}
	case "gridRowEnd":
		if n, ok := toInt(value); ok {
			s.GridRowEnd = n
		}
	}
}

// styleFromMap creates a Style from a map of property names to values.
// Uses applyStyleField for each property — single source of truth.
func styleFromMap(m map[string]any) Style {
	var s Style
	s.Right = -1
	s.Bottom = -1

	for name, value := range m {
		applyStyleField(&s, name, value)
	}
	return s
}

// StyleFromMap is the exported version of styleFromMap.
// Creates a Style from a map of property names to values.
func StyleFromMap(m map[string]any) Style {
	return styleFromMap(m)
}

// MergeStyleFromMap applies style fields from a map onto an existing Style.
// Fields present in the map override the corresponding fields in s.
func MergeStyleFromMap(s *Style, m map[string]any) {
	for name, value := range m {
		applyStyleField(s, name, value)
	}
}
