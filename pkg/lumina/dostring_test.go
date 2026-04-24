package lumina

import (
    "testing"
)

func TestAppEvalServerSim(t *testing.T) {
    app := NewApp()
    defer app.Close()
    
    // Now test eval — all Lua operations happen on this goroutine (safe)
    if err := app.L.DoString("return lumina.Checkbox ~= nil"); err != nil {
        t.Fatalf("DoString failed: %v", err)
    }
    
    n := app.L.GetTop()
    t.Logf("Stack size: %d", n)
    if n > 0 {
        t.Logf("Result: %v", app.L.ToBoolean(1))
    }
}
