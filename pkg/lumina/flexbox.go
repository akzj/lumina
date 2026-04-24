package lumina

// FlexLayout defines the layout properties of a flex container.
type FlexLayout struct {
	Direction      string  // "row" | "column"
	Wrap           string  // "nowrap" | "wrap"
	JustifyContent string  // "flex-start" | "center" | "flex-end" | "space-between" | "space-around" | "space-evenly"
	AlignItems     string  // "flex-start" | "center" | "flex-end" | "stretch"
	AlignContent   string  // same as justify
	Gap            int     // gap between items
}

// FlexItem defines the flex properties of a child item.
type FlexItem struct {
	Grow      float64 // flex-grow (default 0)
	Shrink    float64 // flex-shrink (default 1)
	Basis     int     // flex-basis (0 = auto, use content size)
	Order     int     // display order (default 0)
	AlignSelf string  // override parent's alignItems ("" = inherit)
	MinWidth  int     // minimum width
	MinHeight int     // minimum height
}

// CalculateFlexLayout computes positioned rectangles for flex items.
// container defines the available space, items define flex properties,
// contentSizes provides natural (content-based) sizes for each item.
// Uses the existing Rect type from output.go (X, Y, W, H).
func CalculateFlexLayout(container Rect, items []FlexItem, contentSizes []Rect, layout FlexLayout) []Rect {
	n := len(items)
	if n == 0 {
		return nil
	}

	// Default layout values
	if layout.Direction == "" {
		layout.Direction = "row"
	}
	if layout.Wrap == "" {
		layout.Wrap = "nowrap"
	}
	if layout.JustifyContent == "" {
		layout.JustifyContent = "flex-start"
	}
	if layout.AlignItems == "" {
		layout.AlignItems = "stretch"
	}

	isRow := layout.Direction == "row"

	// Get main and cross axis sizes from container
	var mainSize, crossSize int
	if isRow {
		mainSize = container.W
		crossSize = container.H
	} else {
		mainSize = container.H
		crossSize = container.W
	}

	// Build flex lines (for wrapping)
	type flexLine struct {
		indices   []int
		mainUsed  int
		crossMax  int
	}

	lines := []flexLine{{}}
	currentLine := 0

	for i := 0; i < n; i++ {
		itemMainSize := getItemMainSize(items[i], contentSizes[i], isRow)
		gap := 0
		if len(lines[currentLine].indices) > 0 {
			gap = layout.Gap
		}

		// Check if item fits on current line (wrapping)
		if layout.Wrap == "wrap" && len(lines[currentLine].indices) > 0 &&
			lines[currentLine].mainUsed+gap+itemMainSize > mainSize {
			// Start new line
			lines = append(lines, flexLine{})
			currentLine++
			gap = 0
		}

		lines[currentLine].indices = append(lines[currentLine].indices, i)
		lines[currentLine].mainUsed += gap + itemMainSize

		itemCrossSize := getItemCrossSize(items[i], contentSizes[i], isRow)
		if itemCrossSize > lines[currentLine].crossMax {
			lines[currentLine].crossMax = itemCrossSize
		}
	}

	// Allocate cross-axis space to lines
	totalCross := 0
	for i := range lines {
		if lines[i].crossMax == 0 {
			lines[i].crossMax = 1 // minimum 1 cell
		}
		totalCross += lines[i].crossMax
	}
	if len(lines) > 1 {
		totalCross += layout.Gap * (len(lines) - 1)
	}

	// Position items
	results := make([]Rect, n)
	crossOffset := 0

	// If single line, expand it to fill container cross size
	if len(lines) == 1 {
		lines[0].crossMax = crossSize
	}

	for _, line := range lines {
		lineItems := line.indices
		lineN := len(lineItems)

		// Calculate total basis/content sizes and total grow/shrink
		totalMain := 0
		totalGrow := 0.0
		totalShrink := 0.0
		itemMains := make([]int, lineN)

		for j, idx := range lineItems {
			sz := getItemMainSize(items[idx], contentSizes[idx], isRow)
			itemMains[j] = sz
			totalMain += sz
			totalGrow += items[idx].Grow
			totalShrink += items[idx].Shrink
		}

		// Add gaps to totalMain
		gapSpace := 0
		if lineN > 1 {
			gapSpace = layout.Gap * (lineN - 1)
		}
		totalMainWithGaps := totalMain + gapSpace
		freeSpace := mainSize - totalMainWithGaps

		// Distribute free space via grow/shrink
		finalMains := make([]int, lineN)
		if freeSpace > 0 && totalGrow > 0 {
			// Grow
			remainder := freeSpace
			for j, idx := range lineItems {
				if items[idx].Grow > 0 {
					share := int(float64(freeSpace) * items[idx].Grow / totalGrow)
					finalMains[j] = itemMains[j] + share
					remainder -= share
				} else {
					finalMains[j] = itemMains[j]
				}
			}
			// Distribute rounding remainder to first growing item
			for j, idx := range lineItems {
				if remainder > 0 && items[idx].Grow > 0 {
					finalMains[j] += remainder
					break
				}
			}
		} else if freeSpace < 0 && totalShrink > 0 {
			// Shrink
			deficit := -freeSpace
			for j, idx := range lineItems {
				if items[idx].Shrink > 0 {
					share := int(float64(deficit) * items[idx].Shrink / totalShrink)
					finalMains[j] = itemMains[j] - share
					if finalMains[j] < 0 {
						finalMains[j] = 0
					}
				} else {
					finalMains[j] = itemMains[j]
				}
			}
		} else {
			copy(finalMains, itemMains)
		}

		// Calculate justify offset and spacing
		usedMain := 0
		for _, m := range finalMains {
			usedMain += m
		}
		usedMain += gapSpace
		justifyFree := mainSize - usedMain
		if justifyFree < 0 {
			justifyFree = 0
		}

		mainOffset := 0
		justifyGap := layout.Gap

		switch layout.JustifyContent {
		case "flex-start":
			mainOffset = 0
		case "flex-end":
			mainOffset = justifyFree
		case "center":
			mainOffset = justifyFree / 2
		case "space-between":
			if lineN > 1 {
				justifyGap = justifyFree / (lineN - 1)
			}
		case "space-around":
			if lineN > 0 {
				spacing := justifyFree / (lineN * 2)
				mainOffset = spacing
				justifyGap = spacing * 2
			}
		case "space-evenly":
			if lineN > 0 {
				spacing := justifyFree / (lineN + 1)
				mainOffset = spacing
				justifyGap = spacing
			}
		}

		// Position each item in this line
		pos := mainOffset
		for j, idx := range lineItems {
			itemCross := getItemCrossSize(items[idx], contentSizes[idx], isRow)

			// Determine cross-axis alignment
			align := layout.AlignItems
			if items[idx].AlignSelf != "" {
				align = items[idx].AlignSelf
			}

			var crossPos int
			lineCross := line.crossMax
			if lineCross > crossSize {
				lineCross = crossSize
			}

			switch align {
			case "flex-start":
				crossPos = crossOffset
			case "flex-end":
				crossPos = crossOffset + lineCross - itemCross
			case "center":
				crossPos = crossOffset + (lineCross-itemCross)/2
			case "stretch":
				crossPos = crossOffset
				itemCross = lineCross
			default:
				crossPos = crossOffset
			}

			if isRow {
				results[idx] = Rect{
					X: container.X + pos,
					Y: container.Y + crossPos,
					W: finalMains[j],
					H: itemCross,
				}
			} else {
				results[idx] = Rect{
					X: container.X + crossPos,
					Y: container.Y + pos,
					W: itemCross,
					H: finalMains[j],
				}
			}

			pos += finalMains[j]
			if j < lineN-1 {
				if layout.JustifyContent == "space-between" || layout.JustifyContent == "space-around" || layout.JustifyContent == "space-evenly" {
					pos += justifyGap
				} else {
					pos += layout.Gap
				}
			}
		}

		crossOffset += line.crossMax + layout.Gap
	}

	return results
}

func getItemMainSize(item FlexItem, content Rect, isRow bool) int {
	if item.Basis > 0 {
		return item.Basis
	}
	if isRow {
		if content.W > 0 {
			return content.W
		}
		return 1
	}
	if content.H > 0 {
		return content.H
	}
	return 1
}

func getItemCrossSize(item FlexItem, content Rect, isRow bool) int {
	if isRow {
		if content.H > 0 {
			return content.H
		}
		return 1
	}
	if content.W > 0 {
		return content.W
	}
	return 1
}
