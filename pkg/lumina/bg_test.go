package lumina

import (
    "testing"
)

func TestBackgroundFill(t *testing.T) {
    app := NewApp()
    defer app.Close()
    
    tio := NewMockTermIO(120, 40)
    SetOutputAdapter(NewANSIAdapter(tio))
    
    err := app.LoadScript("../../examples/components-showcase/main.lua", tio)
    if err != nil {
        t.Fatalf("LoadScript: %v", err)
    }
    
    app.RenderOnce()
    
    if app.lastFrame == nil {
        t.Fatal("No frame rendered")
    }
    
    frame := app.lastFrame
    t.Logf("Frame: %dx%d", frame.Width, frame.Height)
    
    emptyBg := 0
    totalCells := 0
    for y := 0; y < frame.Height; y++ {
        for x := 0; x < frame.Width; x++ {
            totalCells++
            if frame.Cells[y][x].Background == "" {
                emptyBg++
            }
        }
    }
    t.Logf("Total cells: %d, Empty background: %d (%.1f%%)", totalCells, emptyBg, float64(emptyBg)/float64(totalCells)*100)
    
    // Show some empty bg cells
    count := 0
    for y := 0; y < frame.Height && count < 5; y++ {
        for x := 0; x < frame.Width && count < 5; x++ {
            c := frame.Cells[y][x]
            if c.Background == "" {
                t.Logf("  Empty bg at [%d,%d] char='%c' fg='%s' owner='%s' role='%s'", x, y, c.Char, c.Foreground, c.OwnerID, c.CellRole)
                count++
            }
        }
    }
    
    // Show cells from last row
    lastY := frame.Height - 1
    for x := 0; x < 5; x++ {
        c := frame.Cells[lastY][x]
        t.Logf("  Last row [%d,%d] char='%c' bg='%s' fg='%s'", x, lastY, c.Char, c.Background, c.Foreground)
    }
    
    if emptyBg > 0 {
        t.Logf("WARNING: %d cells (%.1f%%) have empty background — these will show terminal default color", emptyBg, float64(emptyBg)/float64(totalCells)*100)
    }
}
