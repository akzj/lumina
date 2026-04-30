package v2

import (
    "os"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/akzj/go-lua/pkg/lua"
    "github.com/akzj/lumina/pkg/output"
)

func TestReloadScript_ComponentsExample(t *testing.T) {
    // Use absolute path
    script, _ := filepath.Abs("../examples/components/main.lua")
    if _, err := os.Stat(script); err != nil {
        // Try from project root
        script, _ = filepath.Abs("examples/components/main.lua")
        if _, err := os.Stat(script); err != nil {
            t.Skip("examples/components/main.lua not found:", err)
        }
    }
    t.Log("Script:", script)

    L := lua.NewState()
    defer L.Close()

    adapter := output.NewTestAdapter()
    app := NewApp(L, 80, 24, adapter)

    events := make(chan InputEvent, 16)
    
    errCh := make(chan error, 1)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                t.Errorf("PANIC during Run: %v", r)
                errCh <- nil
            }
        }()
        errCh <- app.Run(RunConfig{
            ScriptPath: script,
            Events:     events,
            Watch:      true,
        })
    }()

    // Wait for initial render
    time.Sleep(1 * time.Second)
    t.Log("App started with components example")

    // Touch the file to trigger reload
    now := time.Now()
    if err := os.Chtimes(script, now, now); err != nil {
        t.Fatal("Chtimes error:", err)
    }
    t.Log("File touched")

    // Wait for watcher to detect and reload
    time.Sleep(2 * time.Second)
    t.Log("After reload wait")

    // Quit
    close(app.quit)
    
    select {
    case err := <-errCh:
        if err != nil {
            t.Fatal("App.Run error:", err)
        }
    case <-time.After(3 * time.Second):
        t.Fatal("App.Run didn't exit")
    }
    t.Log("Test passed - no crash")
}
