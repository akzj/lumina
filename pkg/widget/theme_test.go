package widget

import (
	"reflect"
	"testing"

	"github.com/akzj/lumina/pkg/render"
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

func TestSwitchThemeButton(t *testing.T) {
	// Switch to Nord, render a button, verify it uses Nord colors
	old := CurrentTheme
	defer func() { CurrentTheme = old }()
	CurrentTheme = NordTheme

	state := Button.NewState()
	node := Button.Render(map[string]any{"label": "Test"}, state).(*render.Node)
	if node.Style.Background != NordTheme.Primary {
		t.Errorf("expected Nord primary %q, got %q", NordTheme.Primary, node.Style.Background)
	}
	text := node.Children[0]
	if text.Style.Foreground != NordTheme.PrimaryDark {
		t.Errorf("expected Nord PrimaryDark %q, got %q", NordTheme.PrimaryDark, text.Style.Foreground)
	}
}

func TestSwitchThemeCheckbox(t *testing.T) {
	old := CurrentTheme
	defer func() { CurrentTheme = old }()
	CurrentTheme = DraculaTheme

	state := Checkbox.NewState()
	node := Checkbox.Render(map[string]any{"label": "Check"}, state).(*render.Node)
	// Indicator should use Dracula primary
	indicator := node.Children[0]
	if indicator.Style.Foreground != DraculaTheme.Primary {
		t.Errorf("expected Dracula primary %q, got %q", DraculaTheme.Primary, indicator.Style.Foreground)
	}
	// Label should use Dracula text
	label := node.Children[1]
	if label.Style.Foreground != DraculaTheme.Text {
		t.Errorf("expected Dracula text %q, got %q", DraculaTheme.Text, label.Style.Foreground)
	}
}

func TestSwitchThemeLabel(t *testing.T) {
	old := CurrentTheme
	defer func() { CurrentTheme = old }()
	CurrentTheme = NordTheme

	state := Label.NewState()
	node := Label.Render(map[string]any{"text": "Hello"}, state).(*render.Node)
	if node.Style.Foreground != NordTheme.Text {
		t.Errorf("expected Nord text %q, got %q", NordTheme.Text, node.Style.Foreground)
	}
}

func TestSwitchThemeSwitch(t *testing.T) {
	old := CurrentTheme
	defer func() { CurrentTheme = old }()
	CurrentTheme = NordTheme

	state := &SwitchState{Checked: true}
	node := Switch.Render(map[string]any{}, state).(*render.Node)
	track := node.Children[0]
	// Checked track bg should be Nord primary
	if track.Style.Background != NordTheme.Primary {
		t.Errorf("expected Nord primary %q, got %q", NordTheme.Primary, track.Style.Background)
	}
}

func TestSwitchThemeRadio(t *testing.T) {
	old := CurrentTheme
	defer func() { CurrentTheme = old }()
	CurrentTheme = DraculaTheme

	state := &RadioState{}
	node := Radio.Render(map[string]any{"label": "Opt"}, state).(*render.Node)
	indicator := node.Children[0]
	if indicator.Style.Foreground != DraculaTheme.Primary {
		t.Errorf("expected Dracula primary %q, got %q", DraculaTheme.Primary, indicator.Style.Foreground)
	}
}

func TestSwitchThemeSelect(t *testing.T) {
	old := CurrentTheme
	defer func() { CurrentTheme = old }()
	CurrentTheme = NordTheme

	state := Select.NewState()
	node := Select.Render(map[string]any{}, state).(*render.Node)
	if node.Style.Background != NordTheme.Surface0 {
		t.Errorf("expected Nord Surface0 %q, got %q", NordTheme.Surface0, node.Style.Background)
	}
}

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
