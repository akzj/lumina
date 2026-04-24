package lumina

// GridLayout defines a CSS Grid-like layout configuration.
type GridLayout struct {
	Columns    []GridTrack // column track definitions
	Rows       []GridTrack // row track definitions
	Gap        int         // gap between all cells
	ColumnGap  int         // column-specific gap (overrides Gap if > 0)
	RowGap     int         // row-specific gap (overrides Gap if > 0)
	AutoRows   GridTrack   // implicit row size
	AutoCols   GridTrack   // implicit column size
	AutoFlow   string      // "row" | "column" | "dense"
}

// GridTrack defines a single track (column or row) size.
type GridTrack struct {
	Type  string // "fixed" | "fr" | "auto" | "minmax"
	Value int    // fixed size in cells, or fr units
	Min   int    // for minmax
	Max   int    // for minmax
}

// GridItem defines how an item is placed in the grid.
type GridItem struct {
	ColumnStart int // 1-based start column (0 = auto)
	ColumnEnd   int // 1-based end column (0 = auto)
	RowStart    int // 1-based start row (0 = auto)
	RowEnd      int // 1-based end row (0 = auto)
	ColumnSpan  int // number of columns to span (default 1)
	RowSpan     int // number of rows to span (default 1)
}

// GridRect represents a positioned rectangle in the grid.
type GridRect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// CalculateGridLayout computes positions for all items in a grid.
func CalculateGridLayout(containerWidth, containerHeight int, layout GridLayout, items []GridItem) []GridRect {
	if len(items) == 0 {
		return nil
	}

	colGap := layout.Gap
	if layout.ColumnGap > 0 {
		colGap = layout.ColumnGap
	}
	rowGap := layout.Gap
	if layout.RowGap > 0 {
		rowGap = layout.RowGap
	}

	// Determine number of columns
	numCols := len(layout.Columns)
	if numCols == 0 {
		numCols = 1
		layout.Columns = []GridTrack{{Type: "fr", Value: 1}}
	}

	// Determine number of rows needed
	numRows := len(layout.Rows)
	neededRows := calculateNeededRows(numCols, items)
	if neededRows > numRows {
		// Add implicit rows
		autoRow := layout.AutoRows
		if autoRow.Type == "" {
			autoRow = GridTrack{Type: "auto", Value: 0}
		}
		for numRows < neededRows {
			layout.Rows = append(layout.Rows, autoRow)
			numRows++
		}
	}
	if numRows == 0 {
		numRows = 1
		layout.Rows = []GridTrack{{Type: "auto", Value: 0}}
	}

	// Resolve column widths
	colWidths := resolveTrackSizes(layout.Columns, containerWidth, colGap)
	// Resolve row heights
	rowHeights := resolveTrackSizes(layout.Rows, containerHeight, rowGap)

	// Calculate column positions (x offsets)
	colPositions := make([]int, len(colWidths))
	x := 0
	for i, w := range colWidths {
		colPositions[i] = x
		x += w + colGap
	}

	// Calculate row positions (y offsets)
	rowPositions := make([]int, len(rowHeights))
	y := 0
	for i, h := range rowHeights {
		rowPositions[i] = y
		y += h + rowGap
	}

	// Place items
	results := make([]GridRect, len(items))
	autoCol, autoRow := 0, 0

	for i, item := range items {
		colStart := item.ColumnStart - 1 // convert to 0-based
		rowStart := item.RowStart - 1
		colSpan := item.ColumnSpan
		rowSpan := item.RowSpan
		if colSpan < 1 {
			colSpan = 1
		}
		if rowSpan < 1 {
			rowSpan = 1
		}

		// Auto-placement
		if item.ColumnStart == 0 && item.RowStart == 0 {
			colStart = autoCol
			rowStart = autoRow
			autoCol += colSpan
			if autoCol >= numCols {
				autoCol = 0
				autoRow++
			}
		} else if item.ColumnStart == 0 {
			colStart = 0
		} else if item.RowStart == 0 {
			rowStart = 0
		}

		// Column end from span
		if item.ColumnEnd > 0 {
			colSpan = item.ColumnEnd - item.ColumnStart
			if colSpan < 1 {
				colSpan = 1
			}
		}
		if item.RowEnd > 0 {
			rowSpan = item.RowEnd - item.RowStart
			if rowSpan < 1 {
				rowSpan = 1
			}
		}

		// Clamp to grid bounds
		if colStart < 0 {
			colStart = 0
		}
		if rowStart < 0 {
			rowStart = 0
		}
		if colStart >= numCols {
			colStart = numCols - 1
		}
		if rowStart >= numRows {
			rowStart = numRows - 1
		}
		colEnd := colStart + colSpan
		if colEnd > numCols {
			colEnd = numCols
		}
		rowEnd := rowStart + rowSpan
		if rowEnd > numRows {
			rowEnd = numRows
		}

		// Calculate rect
		rx := 0
		if colStart < len(colPositions) {
			rx = colPositions[colStart]
		}
		ry := 0
		if rowStart < len(rowPositions) {
			ry = rowPositions[rowStart]
		}

		// Width = sum of spanned columns + gaps between them
		rw := 0
		for c := colStart; c < colEnd && c < len(colWidths); c++ {
			rw += colWidths[c]
			if c > colStart {
				rw += colGap
			}
		}
		// Height = sum of spanned rows + gaps between them
		rh := 0
		for r := rowStart; r < rowEnd && r < len(rowHeights); r++ {
			rh += rowHeights[r]
			if r > rowStart {
				rh += rowGap
			}
		}

		results[i] = GridRect{X: rx, Y: ry, Width: rw, Height: rh}
	}

	return results
}

// resolveTrackSizes resolves track sizes given available space.
func resolveTrackSizes(tracks []GridTrack, available int, gap int) []int {
	n := len(tracks)
	if n == 0 {
		return nil
	}

	totalGap := gap * (n - 1)
	remaining := available - totalGap
	if remaining < 0 {
		remaining = 0
	}

	sizes := make([]int, n)
	totalFr := 0
	fixedUsed := 0

	// First pass: allocate fixed and auto sizes
	for i, t := range tracks {
		switch t.Type {
		case "fixed":
			sizes[i] = t.Value
			fixedUsed += t.Value
		case "auto":
			// Auto gets minimum size (1 cell)
			autoSize := 1
			if t.Value > 0 {
				autoSize = t.Value
			}
			sizes[i] = autoSize
			fixedUsed += autoSize
		case "fr":
			fr := t.Value
			if fr < 1 {
				fr = 1
			}
			totalFr += fr
		case "minmax":
			sizes[i] = t.Min
			fixedUsed += t.Min
		}
	}

	// Second pass: distribute remaining space to fr tracks
	frSpace := remaining - fixedUsed
	if frSpace < 0 {
		frSpace = 0
	}

	if totalFr > 0 {
		for i, t := range tracks {
			if t.Type == "fr" {
				fr := t.Value
				if fr < 1 {
					fr = 1
				}
				sizes[i] = (frSpace * fr) / totalFr
			}
		}
	}

	// Third pass: handle minmax max constraints
	for i, t := range tracks {
		if t.Type == "minmax" && t.Max > 0 && sizes[i] < t.Max {
			// Could grow up to max if space allows
			extra := t.Max - sizes[i]
			if extra > frSpace {
				extra = frSpace
			}
			sizes[i] += extra
		}
	}

	return sizes
}

// calculateNeededRows determines how many rows are needed for auto-placed items.
func calculateNeededRows(numCols int, items []GridItem) int {
	if numCols <= 0 {
		numCols = 1
	}
	maxRow := 0
	autoCol, autoRow := 0, 0

	for _, item := range items {
		rowSpan := item.RowSpan
		if rowSpan < 1 {
			rowSpan = 1
		}
		colSpan := item.ColumnSpan
		if colSpan < 1 {
			colSpan = 1
		}

		if item.RowStart > 0 {
			end := item.RowStart - 1 + rowSpan
			if end > maxRow {
				maxRow = end
			}
		} else {
			// Auto-placed
			end := autoRow + rowSpan
			if end > maxRow {
				maxRow = end
			}
			autoCol += colSpan
			if autoCol >= numCols {
				autoCol = 0
				autoRow++
			}
		}
	}
	return maxRow
}
