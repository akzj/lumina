package lumina

import (
	"fmt"
	"strings"
	"sync"
)

// I18n manages internationalization — translations, locale switching, fallback.
type I18n struct {
	mu           sync.RWMutex
	locale       string
	fallback     string
	translations map[string]map[string]string // locale -> key -> value
	version      int                          // incremented on locale change
}

var globalI18n = NewI18n("en")

// NewI18n creates a new I18n instance with the given default locale.
func NewI18n(defaultLocale string) *I18n {
	return &I18n{
		locale:       defaultLocale,
		fallback:     defaultLocale,
		translations: make(map[string]map[string]string),
	}
}

// GetI18n returns the global I18n instance.
func GetI18n() *I18n {
	return globalI18n
}

// AddTranslation adds translations for a locale.
func (i *I18n) AddTranslation(locale string, translations map[string]string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.translations[locale] == nil {
		i.translations[locale] = make(map[string]string)
	}
	for k, v := range translations {
		i.translations[locale][k] = v
	}
}

// SetLocale sets the current locale.
func (i *I18n) SetLocale(locale string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.locale = locale
	i.version++
}

// GetLocale returns the current locale.
func (i *I18n) GetLocale() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.locale
}

// T translates a key using the current locale, falling back to the fallback locale.
// Supports simple interpolation: T("hello.name", "World") replaces {1} with "World".
func (i *I18n) T(key string, args ...string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Try current locale
	if t, ok := i.translations[i.locale]; ok {
		if v, ok := t[key]; ok {
			return interpolate(v, args)
		}
	}
	// Try fallback locale
	if i.locale != i.fallback {
		if t, ok := i.translations[i.fallback]; ok {
			if v, ok := t[key]; ok {
				return interpolate(v, args)
			}
		}
	}
	// Return key as-is
	return key
}

// HasTranslation checks if a key exists for the current locale.
func (i *I18n) HasTranslation(key string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if t, ok := i.translations[i.locale]; ok {
		if _, ok := t[key]; ok {
			return true
		}
	}
	return false
}

// Version returns the version counter (incremented on locale change).
func (i *I18n) Version() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.version
}

// Reset resets to default state (for testing).
func (i *I18n) Reset() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.locale = "en"
	i.fallback = "en"
	i.translations = make(map[string]map[string]string)
	i.version = 0
}

// interpolate replaces {1}, {2}, etc. with args.
func interpolate(template string, args []string) string {
	result := template
	for idx, arg := range args {
		placeholder := fmt.Sprintf("{%d}", idx+1)
		result = strings.ReplaceAll(result, placeholder, arg)
	}
	return result
}
