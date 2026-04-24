package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestValidateRequired(t *testing.T) {
	ok, _ := ApplyRule("", ValidationRule{Type: "required", Message: "required"})
	if ok {
		t.Fatal("empty string should fail required")
	}
	ok2, _ := ApplyRule("hello", ValidationRule{Type: "required"})
	if !ok2 {
		t.Fatal("non-empty should pass required")
	}
	ok3, _ := ApplyRule(nil, ValidationRule{Type: "required", Message: "required"})
	if ok3 {
		t.Fatal("nil should fail required")
	}
}

func TestValidateMinLength(t *testing.T) {
	ok, _ := ApplyRule("ab", ValidationRule{Type: "minLength", Value: 3, Message: "too short"})
	if ok {
		t.Fatal("'ab' should fail minLength 3")
	}
	ok2, _ := ApplyRule("abc", ValidationRule{Type: "minLength", Value: 3})
	if !ok2 {
		t.Fatal("'abc' should pass minLength 3")
	}
}

func TestValidateMaxLength(t *testing.T) {
	ok, _ := ApplyRule("abcdef", ValidationRule{Type: "maxLength", Value: 5, Message: "too long"})
	if ok {
		t.Fatal("'abcdef' should fail maxLength 5")
	}
	ok2, _ := ApplyRule("abc", ValidationRule{Type: "maxLength", Value: 5})
	if !ok2 {
		t.Fatal("'abc' should pass maxLength 5")
	}
}

func TestValidatePattern(t *testing.T) {
	ok, _ := ApplyRule("abc123", ValidationRule{Type: "pattern", Value: `^\d+$`, Message: "numbers only"})
	if ok {
		t.Fatal("'abc123' should fail digits-only pattern")
	}
	ok2, _ := ApplyRule("12345", ValidationRule{Type: "pattern", Value: `^\d+$`})
	if !ok2 {
		t.Fatal("'12345' should pass digits-only pattern")
	}
}

func TestValidateEmail(t *testing.T) {
	ok, _ := ApplyRule("notanemail", ValidationRule{Type: "email", Message: "bad email"})
	if ok {
		t.Fatal("'notanemail' should fail email validation")
	}
	ok2, _ := ApplyRule("user@example.com", ValidationRule{Type: "email"})
	if !ok2 {
		t.Fatal("'user@example.com' should pass email validation")
	}
}

func TestValidateMinMax(t *testing.T) {
	ok, _ := ApplyRule(-1, ValidationRule{Type: "min", Value: 0, Message: "must be >= 0"})
	if ok {
		t.Fatal("-1 should fail min 0")
	}
	ok2, _ := ApplyRule(200, ValidationRule{Type: "max", Value: 150, Message: "must be <= 150"})
	if ok2 {
		t.Fatal("200 should fail max 150")
	}
	ok3, _ := ApplyRule(50, ValidationRule{Type: "min", Value: 0})
	if !ok3 {
		t.Fatal("50 should pass min 0")
	}
}

func TestValidateCustom(t *testing.T) {
	rule := ValidationRule{
		Type:    "custom",
		Message: "must be even",
		Validate: func(v any) bool {
			n := formToInt(v)
			return n%2 == 0
		},
	}
	ok, _ := ApplyRule(3, rule)
	if ok {
		t.Fatal("3 should fail 'must be even'")
	}
	ok2, _ := ApplyRule(4, rule)
	if !ok2 {
		t.Fatal("4 should pass 'must be even'")
	}
}

func TestFormStateValidation(t *testing.T) {
	fs := NewFormState(
		map[string]any{"name": "", "email": ""},
		map[string][]ValidationRule{
			"name":  {{Type: "required", Message: "Name required"}},
			"email": {{Type: "required", Message: "Email required"}, {Type: "email", Message: "Bad email"}},
		},
	)

	valid := fs.ValidateAll()
	if valid {
		t.Fatal("form with empty required fields should be invalid")
	}
	errs := fs.GetErrors()
	if errs["name"] != "Name required" {
		t.Fatalf("expected 'Name required', got '%s'", errs["name"])
	}

	fs.SetField("name", "John")
	fs.SetField("email", "john@example.com")
	valid2 := fs.ValidateAll()
	if !valid2 {
		t.Fatalf("form should be valid, errors: %v", fs.GetErrors())
	}
}

func TestFormStateTouchedDirty(t *testing.T) {
	fs := NewFormState(map[string]any{"name": "initial"}, nil)
	field := fs.Fields["name"]
	if field.Touched || field.Dirty {
		t.Fatal("field should not be touched or dirty initially")
	}
	fs.TouchField("name")
	if !field.Touched {
		t.Fatal("field should be touched after TouchField")
	}
	fs.SetField("name", "changed")
	if !field.Dirty {
		t.Fatal("field should be dirty after SetField")
	}
}

func TestLuaUseFormAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local submitted = false
		local form = lumina.useForm({
			defaultValues = { name = "", email = "" },
			rules = {
				name = {
					{ type = "required", message = "Name is required" },
					{ type = "minLength", value = 2, message = "Too short" },
				},
				email = {
					{ type = "required", message = "Email required" },
					{ type = "email", message = "Invalid email" },
				},
			},
			onSubmit = function(values)
				submitted = true
			end,
		})

		-- Should fail validation with empty defaults
		local ok = form.handleSubmit()
		assert(ok == false, "should fail with empty fields")
		assert(submitted == false, "should not call onSubmit on invalid form")

		-- Set valid values
		form.setValue("name", "John")
		form.setValue("email", "john@example.com")

		local ok2 = form.handleSubmit()
		assert(ok2 == true, "should succeed with valid fields")
		assert(submitted == true, "should call onSubmit")

		-- Validate individual field
		form.setValue("name", "J")
		local valid, err = form.validateField("name")
		assert(valid == false, "single char should fail minLength")
		assert(err == "Too short", "expected 'Too short', got " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("Lua useForm: %v", err)
	}
}
