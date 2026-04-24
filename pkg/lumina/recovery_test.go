package lumina

import (
	"errors"
	"strings"
	"testing"
)

func TestRecoverFunc_NoPanic(t *testing.T) {
	err := RecoverFunc("test", func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestRecoverFunc_WithError(t *testing.T) {
	err := RecoverFunc("test", func() error {
		return errors.New("normal error")
	})
	if err == nil || err.Error() != "normal error" {
		t.Fatalf("expected 'normal error', got: %v", err)
	}
}

func TestRecoverFunc_WithPanic(t *testing.T) {
	err := RecoverFunc("test-op", func() error {
		panic("something went wrong")
	})
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	if !strings.Contains(err.Error(), "panic recovered") {
		t.Fatalf("expected 'panic recovered' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "test-op") {
		t.Fatalf("expected operation name in error, got: %v", err)
	}
}

func TestRecoverValue_NoPanic(t *testing.T) {
	val, err := RecoverValue[int]("test", func() (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if val != 42 {
		t.Fatalf("expected 42, got: %d", val)
	}
}

func TestRecoverValue_WithPanic(t *testing.T) {
	val, err := RecoverValue[int]("render", func() (int, error) {
		panic("render crash")
	})
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got: %d", val)
	}
}

func TestSafeRenderComponent(t *testing.T) {
	// Normal render
	node, err := SafeRenderComponent("TestComp", func() (VNode, error) {
		return VNode{Type: "text", Content: "hello"}, nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if node.Content != "hello" {
		t.Fatalf("expected 'hello', got: %s", node.Content)
	}

	// Panic render
	_, err = SafeRenderComponent("CrashComp", func() (VNode, error) {
		panic("nil pointer")
	})
	if err == nil {
		t.Fatal("expected error from panic")
	}
	if !strings.Contains(err.Error(), "CrashComp") {
		t.Fatalf("expected component name in error: %v", err)
	}
}

func TestSafeHandleEvent(t *testing.T) {
	err := SafeHandleEvent("click", func() error {
		panic("handler crash")
	})
	if err == nil {
		t.Fatal("expected error from panic")
	}
	if !strings.Contains(err.Error(), "click") {
		t.Fatalf("expected event type in error: %v", err)
	}
}

func TestErrorNode(t *testing.T) {
	node := ErrorNode("MyComponent", errors.New("render failed"))
	if node.Type != "text" {
		t.Fatalf("expected 'text', got: %s", node.Type)
	}
	if !strings.Contains(node.Content, "MyComponent") {
		t.Fatalf("expected component name in content: %s", node.Content)
	}
	if !strings.Contains(node.Content, "render failed") {
		t.Fatalf("expected error message in content: %s", node.Content)
	}
}
