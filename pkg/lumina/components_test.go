package lumina

import (
    "testing"
    "github.com/akzj/go-lua/pkg/lua"
)

func TestCheckboxAllFields(t *testing.T) {
    app := NewApp()
    defer app.Close()
    
    CheckboxComponent(app.L)
    
    t.Logf("Stack size: %d", app.L.GetTop())
    
    // Factory is at -1
    for _, field := range []string{"name", "type", "isComponent", "render", "init", "props", "id"} {
        app.L.GetField(-1, field)
        typ := app.L.Type(-1)
        t.Logf(".%s = %s", field, app.L.TypeName(typ))
        if typ == lua.TypeString {
            s, _ := app.L.ToString(-1)
            t.Logf("  value: '%s'", s)
        } else if typ == lua.TypeBoolean {
            t.Logf("  value: %v", app.L.ToBoolean(-1))
        } else if typ == lua.TypeFunction {
            t.Logf("  is function")
        }
        app.L.Pop(1)
    }
}
