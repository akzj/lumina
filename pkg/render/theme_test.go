package render

import (
	"reflect"
	"testing"
)

func TestDefaultThemeHasAllFields(t *testing.T) {
	th := DefaultTheme
	v := reflect.ValueOf(th).Elem()
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		val := v.Field(i).String()
		if val == "" {
			t.Errorf("DefaultTheme.%s is empty", field.Name)
		}
	}
}

func TestNordThemeHasAllFields(t *testing.T) {
	th := NordTheme
	v := reflect.ValueOf(th).Elem()
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		val := v.Field(i).String()
		if val == "" {
			t.Errorf("NordTheme.%s is empty", field.Name)
		}
	}
}

func TestDraculaThemeHasAllFields(t *testing.T) {
	th := DraculaTheme
	v := reflect.ValueOf(th).Elem()
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		val := v.Field(i).String()
		if val == "" {
			t.Errorf("DraculaTheme.%s is empty", field.Name)
		}
	}
}

// All widget-specific theme tests have been removed (widgets deleted).
// Theme system is still tested via ThemeToMap, SetThemeByName, etc.

func TestCurrentThemeDefaultsToDefault(t *testing.T) {
	if CurrentTheme != DefaultTheme {
		t.Error("CurrentTheme should default to DefaultTheme")
	}
}

func TestThemesAreDifferent(t *testing.T) {
	if DefaultTheme.Primary == NordTheme.Primary {
		t.Error("DefaultTheme and NordTheme should have different Primary colors")
	}
	if DefaultTheme.Primary == DraculaTheme.Primary {
		t.Error("DefaultTheme and DraculaTheme should have different Primary colors")
	}
	if NordTheme.Primary == DraculaTheme.Primary {
		t.Error("NordTheme and DraculaTheme should have different Primary colors")
	}
}

func TestLatteThemeHasAllFields(t *testing.T) {
	th := LatteTheme
	v := reflect.ValueOf(th).Elem()
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		val := v.Field(i).String()
		if val == "" {
			t.Errorf("LatteTheme.%s is empty", field.Name)
		}
	}
}

func TestSetThemeByName(t *testing.T) {
	old := CurrentTheme
	defer func() { CurrentTheme = old }()

	if !SetThemeByName("nord") {
		t.Fatal("SetThemeByName(\"nord\") returned false")
	}
	if CurrentTheme != NordTheme {
		t.Error("CurrentTheme should be NordTheme after SetThemeByName(\"nord\")")
	}

	if !SetThemeByName("dracula") {
		t.Fatal("SetThemeByName(\"dracula\") returned false")
	}
	if CurrentTheme != DraculaTheme {
		t.Error("CurrentTheme should be DraculaTheme after SetThemeByName(\"dracula\")")
	}

	if !SetThemeByName("latte") {
		t.Fatal("SetThemeByName(\"latte\") returned false")
	}
	if CurrentTheme != LatteTheme {
		t.Error("CurrentTheme should be LatteTheme after SetThemeByName(\"latte\")")
	}

	if !SetThemeByName("mocha") {
		t.Fatal("SetThemeByName(\"mocha\") returned false")
	}
	if CurrentTheme != DefaultTheme {
		t.Error("CurrentTheme should be DefaultTheme after SetThemeByName(\"mocha\")")
	}

	if SetThemeByName("nonexistent") {
		t.Error("SetThemeByName(\"nonexistent\") should return false")
	}
}

func TestThemeToMap(t *testing.T) {
	m := ThemeToMap(DefaultTheme)
	if m["base"] != DefaultTheme.Base {
		t.Errorf("ThemeToMap base: got %q, want %q", m["base"], DefaultTheme.Base)
	}
	if m["primary"] != DefaultTheme.Primary {
		t.Errorf("ThemeToMap primary: got %q, want %q", m["primary"], DefaultTheme.Primary)
	}
	if m["error"] != DefaultTheme.Error {
		t.Errorf("ThemeToMap error: got %q, want %q", m["error"], DefaultTheme.Error)
	}
	// Verify all 13 fields are present
	expected := 13
	if len(m) != expected {
		t.Errorf("ThemeToMap returned %d keys, want %d", len(m), expected)
	}
}

func TestBuiltinThemesMap(t *testing.T) {
	if len(BuiltinThemes) != 4 {
		t.Errorf("BuiltinThemes has %d entries, want 4", len(BuiltinThemes))
	}
	names := []string{"mocha", "latte", "nord", "dracula"}
	for _, name := range names {
		if _, ok := BuiltinThemes[name]; !ok {
			t.Errorf("BuiltinThemes missing %q", name)
		}
	}
}
