package render

import (
	"strconv"
	"strings"
)

// --- CSS Grid layout ---

// gridTrack represents a single track in a grid template.
type gridTrack struct {
	fr int // fractional unit (0 = not fr)
	px int // pixel/cell value (0 = auto)
}

// parseGridTemplate parses a grid template string like "1fr 2fr 100" into track sizes.
// available is the total space to distribute among tracks.
// gapSize is the gap between tracks.
func parseGridTemplate(template string, available int, gapSize int) []int {
	if template == "" {
		return nil
	}
	parts := strings.Fields(template)
	if len(parts) == 0 {
		return nil
	}

	tracks := make([]gridTrack, len(parts))
	fixedTotal := 0
	frTotal := 0

	for i, part := range parts {
		if strings.HasSuffix(part, "fr") {
			nStr := strings.TrimSuffix(part, "fr")
			n, err := strconv.Atoi(nStr)
			if err != nil || n < 1 {
				n = 1
			}
			tracks[i] = gridTrack{fr: n}
			frTotal += n
		} else if part == "auto" {
			tracks[i] = gridTrack{px: 0} // auto = 0, will get remaining space
			frTotal += 1
			tracks[i] = gridTrack{fr: 1} // treat auto as 1fr
		} else {
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 {
				n = 0
			}
			tracks[i] = gridTrack{px: n}
			fixedTotal += n
		}
	}

	// Subtract gaps from available space
	totalGaps := 0
	if len(tracks) > 1 {
		totalGaps = gapSize * (len(tracks) - 1)
	}
	remainW := available - fixedTotal - totalGaps
	if remainW < 0 {
		remainW = 0
	}

	sizes := make([]int, len(tracks))
	for i, t := range tracks {
		if t.fr > 0 && frTotal > 0 {
			sizes[i] = (remainW * t.fr) / frTotal
		} else {
			sizes[i] = t.px
		}
		if sizes[i] < 0 {
			sizes[i] = 0
		}
	}
	return sizes
}

// parseGridSpan parses "1 / 3" into (start=1, end=3) or "2" into (start=2, end=3).
// Returns 1-based start and exclusive end.
func parseGridSpan(s string) (int, int) {
	if s == "" {
		return 0, 0
	}
	parts := strings.Split(s, "/")
	if len(parts) == 2 {
		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 == nil && err2 == nil && start >= 1 && end > start {
			return start, end
		}
	}
	// Single number: occupies one cell
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err == nil && n >= 1 {
		return n, n + 1
	}
	return 0, 0
}

// layoutGrid lays out children in a CSS Grid.
func layoutGrid(node *Node, contentX, contentY, contentW, contentH int, style Style, depth int) {
	if len(node.Children) == 0 {
		return
	}

	// Determine gaps
	colGap := style.GridColumnGap
	if colGap == 0 {
		colGap = style.Gap
	}
	rowGap := style.GridRowGap
	if rowGap == 0 {
		rowGap = style.Gap
	}

	// Parse column template
	colSizes := parseGridTemplate(style.GridTemplateColumns, contentW, colGap)
	if len(colSizes) == 0 {
		// Default: single column
		colSizes = []int{contentW}
	}
	numCols := len(colSizes)

	// Collect flow children
	type gridChild struct {
		childIdx int
		colStart int // 1-based
		colEnd   int // exclusive
		rowStart int // 1-based
		rowEnd   int // exclusive
	}
	var flowChildren []gridChild
	for i, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			continue
		}

		colStart, colEnd := 0, 0
		rowStart, rowEnd := 0, 0

		// Explicit placement via gridColumn/gridRow strings
		if cs.GridColumn != "" {
			colStart, colEnd = parseGridSpan(cs.GridColumn)
		}
		if cs.GridRow != "" {
			rowStart, rowEnd = parseGridSpan(cs.GridRow)
		}

		// Explicit placement via gridColumnStart/End fields
		if colStart == 0 && cs.GridColumnStart > 0 {
			colStart = cs.GridColumnStart
			colEnd = cs.GridColumnEnd
			if colEnd <= colStart {
				colEnd = colStart + 1
			}
		}
		if rowStart == 0 && cs.GridRowStart > 0 {
			rowStart = cs.GridRowStart
			rowEnd = cs.GridRowEnd
			if rowEnd <= rowStart {
				rowEnd = rowStart + 1
			}
		}

		flowChildren = append(flowChildren, gridChild{
			childIdx: i,
			colStart: colStart,
			colEnd:   colEnd,
			rowStart: rowStart,
			rowEnd:   rowEnd,
		})
	}

	if len(flowChildren) == 0 {
		return
	}

	// Auto-place children that don't have explicit positions
	// Track which cells are occupied
	// First pass: determine number of rows needed
	maxRow := 0
	for _, gc := range flowChildren {
		if gc.rowEnd > maxRow {
			maxRow = gc.rowEnd - 1
		}
	}
	// Estimate rows from auto-placement
	autoPlaceCount := 0
	for _, gc := range flowChildren {
		if gc.colStart == 0 {
			autoPlaceCount++
		}
	}
	estimatedRows := maxRow
	if autoPlaceCount > 0 {
		neededRows := (autoPlaceCount + numCols - 1) / numCols
		if neededRows > estimatedRows {
			estimatedRows = neededRows
		}
	}
	if estimatedRows < 1 {
		estimatedRows = 1
	}

	// Build occupancy grid and place explicitly positioned children
	numRows := estimatedRows + maxRow // overestimate to be safe
	if numRows < 1 {
		numRows = 1
	}
	occupied := make([][]bool, numRows)
	for r := range occupied {
		occupied[r] = make([]bool, numCols)
	}

	// Place explicitly positioned children
	for i := range flowChildren {
		gc := &flowChildren[i]
		if gc.colStart > 0 && gc.rowStart > 0 {
			// Mark cells as occupied
			for r := gc.rowStart - 1; r < gc.rowEnd-1 && r < numRows; r++ {
				for c := gc.colStart - 1; c < gc.colEnd-1 && c < numCols; c++ {
					occupied[r][c] = true
				}
			}
		}
	}

	// Auto-place remaining children
	autoRow, autoCol := 0, 0
	for i := range flowChildren {
		gc := &flowChildren[i]
		if gc.colStart > 0 && gc.rowStart > 0 {
			continue // already placed
		}

		// Find next available cell
		span := 1
		if gc.colStart > 0 && gc.colEnd > gc.colStart {
			span = gc.colEnd - gc.colStart
		}

		for {
			if autoCol+span <= numCols {
				// Check if cells are free
				free := true
				for c := autoCol; c < autoCol+span; c++ {
					if autoRow < numRows && occupied[autoRow][c] {
						free = false
						break
					}
				}
				if free {
					break
				}
			}
			autoCol++
			if autoCol+span > numCols {
				autoCol = 0
				autoRow++
				// Expand occupied grid if needed
				if autoRow >= numRows {
					numRows++
					occupied = append(occupied, make([]bool, numCols))
				}
			}
		}

		gc.colStart = autoCol + 1
		gc.colEnd = autoCol + span + 1
		if gc.colEnd > numCols+1 {
			gc.colEnd = numCols + 1
		}
		gc.rowStart = autoRow + 1
		gc.rowEnd = autoRow + 2
		if gc.rowEnd-1 > gc.rowStart-1 {
			// multi-row not supported in auto-place
		}

		// Mark occupied
		for c := autoCol; c < autoCol+span && c < numCols; c++ {
			if autoRow < numRows {
				occupied[autoRow][c] = true
			}
		}

		// Advance auto cursor
		autoCol += span
		if autoCol >= numCols {
			autoCol = 0
			autoRow++
			if autoRow >= numRows {
				numRows++
				occupied = append(occupied, make([]bool, numCols))
			}
		}
	}

	// Determine actual number of rows used
	actualRows := 0
	for _, gc := range flowChildren {
		if gc.rowEnd-1 > actualRows {
			actualRows = gc.rowEnd - 1
		}
	}
	if actualRows < 1 {
		actualRows = 1
	}

	// Parse row template
	rowSizes := parseGridTemplate(style.GridTemplateRows, contentH, rowGap)
	// If row template doesn't cover all rows, extend with equal distribution
	if len(rowSizes) < actualRows {
		// Calculate remaining height after defined rows
		definedH := 0
		for _, h := range rowSizes {
			definedH += h
		}
		definedGaps := 0
		if len(rowSizes) > 0 {
			definedGaps = rowGap * len(rowSizes)
		}
		remainH := contentH - definedH - definedGaps
		if remainH < 0 {
			remainH = 0
		}
		extraRows := actualRows - len(rowSizes)
		extraGaps := 0
		if extraRows > 1 {
			extraGaps = rowGap * (extraRows - 1)
		}
		perRow := 1
		if extraRows > 0 && remainH > extraGaps {
			perRow = (remainH - extraGaps) / extraRows
		}
		if perRow < 1 {
			perRow = 1
		}
		for len(rowSizes) < actualRows {
			rowSizes = append(rowSizes, perRow)
		}
	}

	// Compute column X positions
	colX := make([]int, numCols)
	cx := contentX
	for c := 0; c < numCols; c++ {
		colX[c] = cx
		cx += colSizes[c]
		if c < numCols-1 {
			cx += colGap
		}
	}

	// Compute row Y positions
	rowY := make([]int, actualRows)
	ry := contentY
	for r := 0; r < actualRows; r++ {
		rowY[r] = ry
		ry += rowSizes[r]
		if r < actualRows-1 {
			ry += rowGap
		}
	}

	// Position each child in its grid cell(s)
	for _, gc := range flowChildren {
		child := node.Children[gc.childIdx]

		c0 := gc.colStart - 1
		c1 := gc.colEnd - 2 // inclusive end column
		r0 := gc.rowStart - 1
		r1 := gc.rowEnd - 2 // inclusive end row

		// Clamp to grid bounds
		if c0 < 0 {
			c0 = 0
		}
		if c1 >= numCols {
			c1 = numCols - 1
		}
		if r0 < 0 {
			r0 = 0
		}
		if r1 >= actualRows {
			r1 = actualRows - 1
		}

		// Calculate cell position and size
		cellX := colX[c0]
		cellY := rowY[r0]

		// Width spans from colStart to colEnd (inclusive of gaps between spanned columns)
		cellW := 0
		for c := c0; c <= c1; c++ {
			cellW += colSizes[c]
			if c < c1 {
				cellW += colGap
			}
		}

		// Height spans from rowStart to rowEnd
		cellH := 0
		for r := r0; r <= r1; r++ {
			cellH += rowSizes[r]
			if r < r1 {
				cellH += rowGap
			}
		}

		if cellW < 1 {
			cellW = 1
		}
		if cellH < 1 {
			cellH = 1
		}

		computeFlex(child, cellX, cellY, cellW, cellH, depth+1)
		applyRelativeOffset(child, child.Style)
	}

	// For scroll containers, store the total content height
	if style.Overflow == "scroll" || style.Overflow == "auto" {
		maxBottom := 0
		for _, child := range node.Children {
			if isPositioned(child.Style) || child.Style.Display == "none" {
				continue
			}
			bottom := child.Y + child.H - contentY
			if bottom > maxBottom {
				maxBottom = bottom
			}
		}
		node.ScrollHeight = maxBottom
	}
}
