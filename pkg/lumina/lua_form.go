package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// luaUseForm implements lumina.useForm(opts) -> form table
func luaUseForm(L *lua.State) int {
	defaults := make(map[string]any)
	rules := make(map[string][]ValidationRule)

	if L.Type(1) == lua.TypeTable {
		// Read defaultValues
		L.GetField(1, "defaultValues")
		if L.Type(-1) == lua.TypeTable {
			L.PushNil()
			for L.Next(-2) {
				key, ok := L.ToString(-2)
				if ok {
					defaults[key] = L.ToAny(-1)
				}
				L.Pop(1)
			}
		}
		L.Pop(1)

		// Read rules
		L.GetField(1, "rules")
		if L.Type(-1) == lua.TypeTable {
			L.PushNil()
			for L.Next(-2) {
				fieldName, ok := L.ToString(-2)
				if !ok {
					L.Pop(1)
					continue
				}
				// Value is a list of rule tables
				if L.Type(-1) == lua.TypeTable {
					var fieldRules []ValidationRule
					n := int(L.RawLen(-1))
					for i := 1; i <= n; i++ {
						L.RawGetI(-1, int64(i))
						if L.Type(-1) == lua.TypeTable {
							rule := parseValidationRule(L, -1)
							fieldRules = append(fieldRules, rule)
						}
						L.Pop(1)
					}
					rules[fieldName] = fieldRules
				}
				L.Pop(1)
			}
		}
		L.Pop(1)
	}

	fs := NewFormState(defaults, rules)

	// Store onSubmit ref
	L.GetField(1, "onSubmit")
	hasOnSubmit := L.Type(-1) == lua.TypeFunction
	onSubmitRef := 0
	if hasOnSubmit {
		onSubmitRef = L.Ref(lua.RegistryIndex)
	} else {
		L.Pop(1)
	}

	// Build form table
	L.NewTable()

	// form.setValue(field, value)
	L.PushFunction(func(L *lua.State) int {
		field := L.CheckString(1)
		value := L.ToAny(2)
		fs.SetField(field, value)
		return 0
	})
	L.SetField(-2, "setValue")

	// form.getValue(field) -> value
	L.PushFunction(func(L *lua.State) int {
		field := L.CheckString(1)
		fs.mu.RLock()
		f := fs.Fields[field]
		fs.mu.RUnlock()
		if f != nil {
			L.PushAny(f.Value)
		} else {
			L.PushNil()
		}
		return 1
	})
	L.SetField(-2, "getValue")

	// form.validate() -> bool
	L.PushFunction(func(L *lua.State) int {
		valid := fs.ValidateAll()
		L.PushBoolean(valid)
		return 1
	})
	L.SetField(-2, "validate")

	// form.validateField(name) -> bool, error
	L.PushFunction(func(L *lua.State) int {
		name := L.CheckString(1)
		valid, errMsg := fs.ValidateField(name)
		L.PushBoolean(valid)
		if errMsg != "" {
			L.PushString(errMsg)
		} else {
			L.PushNil()
		}
		return 2
	})
	L.SetField(-2, "validateField")

	// form.getErrors() -> table
	L.PushFunction(func(L *lua.State) int {
		errs := fs.GetErrors()
		L.NewTable()
		for k, v := range errs {
			L.PushString(v)
			L.SetField(-2, k)
		}
		return 1
	})
	L.SetField(-2, "getErrors")

	// form.getValues() -> table
	L.PushFunction(func(L *lua.State) int {
		vals := fs.GetValues()
		L.PushAny(vals)
		return 1
	})
	L.SetField(-2, "getValues")

	// form.handleSubmit()
	L.PushFunction(func(L *lua.State) int {
		if !fs.ValidateAll() {
			L.PushBoolean(false)
			return 1
		}
		if hasOnSubmit {
			L.RawGetI(lua.RegistryIndex, int64(onSubmitRef))
			vals := fs.GetValues()
			L.PushAny(vals)
			L.PCall(1, 0, 0)
		}
		L.PushBoolean(true)
		return 1
	})
	L.SetField(-2, "handleSubmit")

	// form.reset()
	L.PushFunction(func(L *lua.State) int {
		fs.Reset(defaults)
		return 0
	})
	L.SetField(-2, "reset")

	// form.isValid (computed on access)
	L.PushBoolean(fs.IsValid)
	L.SetField(-2, "isValid")

	return 1
}

func parseValidationRule(L *lua.State, idx int) ValidationRule {
	rule := ValidationRule{}

	L.GetField(idx, "type")
	if s, ok := L.ToString(-1); ok {
		rule.Type = s
	}
	L.Pop(1)

	L.GetField(idx, "value")
	rule.Value = L.ToAny(-1)
	L.Pop(1)

	L.GetField(idx, "message")
	if s, ok := L.ToString(-1); ok {
		rule.Message = s
	}
	L.Pop(1)

	return rule
}
