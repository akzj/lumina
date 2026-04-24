package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestI18nAddTranslation(t *testing.T) {
	i := NewI18n("en")
	i.AddTranslation("en", map[string]string{
		"app.title":     "My App",
		"button.submit": "Submit",
	})
	if v := i.T("app.title"); v != "My App" {
		t.Fatalf("expected 'My App', got '%s'", v)
	}
	if v := i.T("button.submit"); v != "Submit" {
		t.Fatalf("expected 'Submit', got '%s'", v)
	}
}

func TestI18nGetTranslation(t *testing.T) {
	i := NewI18n("en")
	i.AddTranslation("en", map[string]string{"hello": "Hello"})
	i.AddTranslation("zh", map[string]string{"hello": "你好"})

	// Default locale is "en"
	if v := i.T("hello"); v != "Hello" {
		t.Fatalf("expected 'Hello', got '%s'", v)
	}

	// Switch to zh
	i.SetLocale("zh")
	if v := i.T("hello"); v != "你好" {
		t.Fatalf("expected '你好', got '%s'", v)
	}
}

func TestI18nSetLocale(t *testing.T) {
	i := NewI18n("en")
	if l := i.GetLocale(); l != "en" {
		t.Fatalf("expected 'en', got '%s'", l)
	}
	i.SetLocale("zh")
	if l := i.GetLocale(); l != "zh" {
		t.Fatalf("expected 'zh', got '%s'", l)
	}
}

func TestI18nFallback(t *testing.T) {
	i := NewI18n("en")
	i.AddTranslation("en", map[string]string{
		"hello":   "Hello",
		"goodbye": "Goodbye",
	})
	i.AddTranslation("zh", map[string]string{
		"hello": "你好",
		// "goodbye" not translated in zh
	})
	i.SetLocale("zh")

	// "hello" exists in zh
	if v := i.T("hello"); v != "你好" {
		t.Fatalf("expected '你好', got '%s'", v)
	}
	// "goodbye" falls back to en
	if v := i.T("goodbye"); v != "Goodbye" {
		t.Fatalf("expected 'Goodbye', got '%s'", v)
	}
	// Unknown key returns key
	if v := i.T("unknown.key"); v != "unknown.key" {
		t.Fatalf("expected 'unknown.key', got '%s'", v)
	}
}

func TestI18nInterpolation(t *testing.T) {
	i := NewI18n("en")
	i.AddTranslation("en", map[string]string{
		"greeting": "Hello, {1}!",
		"complex":  "{1} has {2} items",
	})
	if v := i.T("greeting", "World"); v != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got '%s'", v)
	}
	if v := i.T("complex", "Cart", "5"); v != "Cart has 5 items" {
		t.Fatalf("expected 'Cart has 5 items', got '%s'", v)
	}
}

func TestLuaI18nAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalI18n.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.i18n.addTranslation("en", {
			["app.title"] = "My App",
			["button.submit"] = "Submit",
		})
		lumina.i18n.addTranslation("zh", {
			["app.title"] = "我的应用",
			["button.submit"] = "提交",
		})

		-- Default locale is "en"
		local locale = lumina.i18n.getLocale()
		assert(locale == "en", "expected en, got " .. tostring(locale))

		local v = lumina.i18n.t("app.title")
		assert(v == "My App", "expected 'My App', got '" .. tostring(v) .. "'")

		-- Switch to zh
		lumina.i18n.setLocale("zh")
		local v2 = lumina.i18n.t("app.title")
		assert(v2 == "我的应用", "expected '我的应用', got '" .. tostring(v2) .. "'")

		local v3 = lumina.i18n.t("button.submit")
		assert(v3 == "提交", "expected '提交', got '" .. tostring(v3) .. "'")
	`)
	if err != nil {
		t.Fatalf("Lua i18n API: %v", err)
	}
	globalI18n.Reset()
}

func TestLuaUseTranslation(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalI18n.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		lumina.i18n.addTranslation("en", {
			["hello"] = "Hello",
			["greeting"] = "Hi, {1}!",
		})

		local t = lumina.useTranslation()
		assert(type(t) == "function", "expected function, got " .. type(t))

		local v = t("hello")
		assert(v == "Hello", "expected 'Hello', got '" .. tostring(v) .. "'")
	`)
	if err != nil {
		t.Fatalf("useTranslation: %v", err)
	}
	globalI18n.Reset()
}
