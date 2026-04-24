package lumina

import (
    "testing"
)

func TestAppEvalServerSim(t *testing.T) {
    app := NewApp()
    
    // Simulate what server does - start render loop
    app.RenderLoop.Start()
    defer app.RenderLoop.Stop()
    
    // Now test eval
    if err := app.L.DoString("return lumina.Checkbox ~= nil"); err != nil {
        t.Fatalf("DoString failed: %v", err)
    }
    
    n := app.L.GetTop()
    t.Logf("Stack size: %d", n)
    if n > 0 {
        t.Logf("Result: %v", app.L.ToBoolean(1))
    }
    
    app.Close()
}
