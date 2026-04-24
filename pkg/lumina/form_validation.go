package lumina

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// ValidationRule defines a single validation rule for a form field.
type ValidationRule struct {
	Type     string                // "required" | "minLength" | "maxLength" | "pattern" | "min" | "max" | "email" | "custom"
	Value    any                   // rule parameter (int for length, string for pattern, etc.)
	Message  string                // error message when validation fails
	Validate func(value any) bool  // custom validator function
}

// FieldState tracks the state of a single form field.
type FieldState struct {
	Value   any
	Error   string
	Touched bool
	Dirty   bool
	Valid   bool
}

// FormState tracks the state of an entire form.
type FormState struct {
	mu           sync.RWMutex
	Fields       map[string]*FieldState
	Rules        map[string][]ValidationRule
	IsValid      bool
	IsSubmitting bool
	Errors       map[string]string
}

// NewFormState creates a new FormState with default values and rules.
func NewFormState(defaults map[string]any, rules map[string][]ValidationRule) *FormState {
	fs := &FormState{
		Fields: make(map[string]*FieldState),
		Rules:  rules,
		Errors: make(map[string]string),
	}
	if rules == nil {
		fs.Rules = make(map[string][]ValidationRule)
	}
	for k, v := range defaults {
		fs.Fields[k] = &FieldState{
			Value: v,
			Valid: true,
		}
	}
	fs.IsValid = true
	return fs
}

// SetField sets a field value and marks it as dirty.
func (fs *FormState) SetField(name string, value any) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	field, ok := fs.Fields[name]
	if !ok {
		field = &FieldState{}
		fs.Fields[name] = field
	}
	field.Value = value
	field.Dirty = true
}

// TouchField marks a field as touched.
func (fs *FormState) TouchField(name string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if field, ok := fs.Fields[name]; ok {
		field.Touched = true
	}
}

// ValidateField validates a single field against its rules.
func (fs *FormState) ValidateField(name string) (bool, string) {
	fs.mu.RLock()
	field := fs.Fields[name]
	rules := fs.Rules[name]
	fs.mu.RUnlock()

	if field == nil {
		return true, ""
	}

	for _, rule := range rules {
		valid, msg := ApplyRule(field.Value, rule)
		if !valid {
			fs.mu.Lock()
			field.Valid = false
			field.Error = msg
			fs.Errors[name] = msg
			fs.mu.Unlock()
			return false, msg
		}
	}

	fs.mu.Lock()
	field.Valid = true
	field.Error = ""
	delete(fs.Errors, name)
	fs.mu.Unlock()
	return true, ""
}

// ValidateAll validates all fields and returns whether the form is valid.
func (fs *FormState) ValidateAll() bool {
	allValid := true
	fs.mu.RLock()
	fieldNames := make([]string, 0, len(fs.Fields))
	for name := range fs.Fields {
		fieldNames = append(fieldNames, name)
	}
	fs.mu.RUnlock()

	for _, name := range fieldNames {
		valid, _ := fs.ValidateField(name)
		if !valid {
			allValid = false
		}
	}

	fs.mu.Lock()
	fs.IsValid = allValid
	fs.mu.Unlock()
	return allValid
}

// GetValues returns all field values as a map.
func (fs *FormState) GetValues() map[string]any {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	values := make(map[string]any)
	for k, f := range fs.Fields {
		values[k] = f.Value
	}
	return values
}

// GetErrors returns all current errors.
func (fs *FormState) GetErrors() map[string]string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	errs := make(map[string]string)
	for k, v := range fs.Errors {
		errs[k] = v
	}
	return errs
}

// Reset resets all fields to their initial values.
func (fs *FormState) Reset(defaults map[string]any) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Errors = make(map[string]string)
	fs.IsValid = true
	fs.IsSubmitting = false
	for k, f := range fs.Fields {
		f.Touched = false
		f.Dirty = false
		f.Error = ""
		f.Valid = true
		if v, ok := defaults[k]; ok {
			f.Value = v
		}
	}
}

// ApplyRule applies a single validation rule to a value.
func ApplyRule(value any, rule ValidationRule) (bool, string) {
	switch rule.Type {
	case "required":
		return validateRequired(value, rule.Message)
	case "minLength":
		return validateMinLength(value, rule.Value, rule.Message)
	case "maxLength":
		return validateMaxLength(value, rule.Value, rule.Message)
	case "pattern":
		return validatePattern(value, rule.Value, rule.Message)
	case "min":
		return validateMin(value, rule.Value, rule.Message)
	case "max":
		return validateMax(value, rule.Value, rule.Message)
	case "email":
		return validateEmail(value, rule.Message)
	case "custom":
		if rule.Validate != nil {
			if rule.Validate(value) {
				return true, ""
			}
			return false, rule.Message
		}
		return true, ""
	}
	return true, ""
}

func validateRequired(value any, msg string) (bool, string) {
	if msg == "" {
		msg = "This field is required"
	}
	if value == nil {
		return false, msg
	}
	if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
		return false, msg
	}
	return true, ""
}

func validateMinLength(value any, param any, msg string) (bool, string) {
	minLen := formToInt(param)
	s := formToString(value)
	if len(s) < minLen {
		if msg == "" {
			msg = "Too short"
		}
		return false, msg
	}
	return true, ""
}

func validateMaxLength(value any, param any, msg string) (bool, string) {
	maxLen := formToInt(param)
	s := formToString(value)
	if len(s) > maxLen {
		if msg == "" {
			msg = "Too long"
		}
		return false, msg
	}
	return true, ""
}

func validatePattern(value any, param any, msg string) (bool, string) {
	pattern := formToString(param)
	s := formToString(value)
	matched, err := regexp.MatchString(pattern, s)
	if err != nil || !matched {
		if msg == "" {
			msg = "Invalid format"
		}
		return false, msg
	}
	return true, ""
}

func validateMin(value any, param any, msg string) (bool, string) {
	minVal := formToFloat(param)
	val := formToFloat(value)
	if val < minVal {
		if msg == "" {
			msg = "Value too small"
		}
		return false, msg
	}
	return true, ""
}

func validateMax(value any, param any, msg string) (bool, string) {
	maxVal := formToFloat(param)
	val := formToFloat(value)
	if val > maxVal {
		if msg == "" {
			msg = "Value too large"
		}
		return false, msg
	}
	return true, ""
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validateEmail(value any, msg string) (bool, string) {
	s := formToString(value)
	if !emailRegex.MatchString(s) {
		if msg == "" {
			msg = "Invalid email"
		}
		return false, msg
	}
	return true, ""
}

func formToString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		s := strconv.FormatFloat(val, 'f', -1, 64)
		return s
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return ""
	}
}

func formToInt(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

func formToFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}
