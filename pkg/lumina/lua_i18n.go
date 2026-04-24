package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// registerI18nModule registers the lumina.i18n subtable.
func registerI18nModule(L *lua.State) {
	// Create lumina.i18n table
	L.NewTable()

	L.PushFunction(luaI18nAddTranslation)
	L.SetField(-2, "addTranslation")

	L.PushFunction(luaI18nSetLocale)
	L.SetField(-2, "setLocale")

	L.PushFunction(luaI18nGetLocale)
	L.SetField(-2, "getLocale")

	L.PushFunction(luaI18nT)
	L.SetField(-2, "t")

	// Set lumina.i18n
	L.SetField(-2, "i18n")
}

// luaI18nAddTranslation implements lumina.i18n.addTranslation(locale, table).
func luaI18nAddTranslation(L *lua.State) int {
	locale := L.CheckString(1)
	if L.Type(2) != lua.TypeTable {
		L.PushString("addTranslation: expected table as second argument")
		L.Error()
		return 0
	}

	translations := make(map[string]string)
	L.PushNil()
	for L.Next(2) {
		if k, ok := L.ToString(-2); ok {
			if v, ok := L.ToString(-1); ok {
				translations[k] = v
			}
		}
		L.Pop(1)
	}

	globalI18n.AddTranslation(locale, translations)
	return 0
}

// luaI18nSetLocale implements lumina.i18n.setLocale(locale).
func luaI18nSetLocale(L *lua.State) int {
	locale := L.CheckString(1)
	globalI18n.SetLocale(locale)
	return 0
}

// luaI18nGetLocale implements lumina.i18n.getLocale() -> string.
func luaI18nGetLocale(L *lua.State) int {
	L.PushString(globalI18n.GetLocale())
	return 1
}

// luaI18nT implements lumina.i18n.t(key, ...) -> string.
func luaI18nT(L *lua.State) int {
	key := L.CheckString(1)
	// Collect optional interpolation args
	nArgs := L.GetTop()
	args := make([]string, 0, nArgs-1)
	for i := 2; i <= nArgs; i++ {
		if s, ok := L.ToString(i); ok {
			args = append(args, s)
		}
	}
	result := globalI18n.T(key, args...)
	L.PushString(result)
	return 1
}

// luaUseTranslation implements lumina.useTranslation() -> t function.
// Returns a function that calls i18n.t().
func luaUseTranslation(L *lua.State) int {
	// Return the t function directly
	L.PushFunction(luaI18nT)
	return 1
}
