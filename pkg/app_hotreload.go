package v2

import (
	"log"
	"path/filepath"
	"strings"
)

// reloadModule attempts a module-level hot reload using go-lua's ReloadModule.
// This swaps function prototypes in-place while preserving upvalues (state),
// so all existing references to old functions automatically see new code.
// Returns true if module reload succeeded, false if fallback is needed.
func (a *App) reloadModule(path string) bool {
	// Convert file path to Lua module name
	moduleName := a.pathToModuleName(path)
	if moduleName == "" {
		return false
	}

	// Check if this module is loaded in package.loaded
	if !a.isModuleLoaded(moduleName) {
		return false
	}

	// Try two-phase reload for safety
	plan, err := a.luaState.PrepareReload(moduleName)
	if err != nil {
		log.Printf("[hotreload] prepare failed for %q: %v", moduleName, err)
		return false
	}

	// Check for incompatible functions
	if plan.HasIncompatible() {
		log.Printf("[hotreload] %q has %d incompatible functions, falling back to full reload",
			moduleName, plan.IncompatibleCount())
		plan.Abort()
		return false
	}

	// Commit the reload
	result := plan.Commit()
	log.Printf("[hotreload] module %q: replaced=%d, skipped=%d, added=%d, removed=%d",
		moduleName, result.Replaced, result.Skipped, result.Added, result.Removed)

	for _, w := range result.Warnings {
		log.Printf("[hotreload] warning: %s", w)
	}

	// Mark all components dirty and re-render.
	// Since Proto is swapped in-place, all components that use functions from
	// this module will automatically get new code on next render.
	a.engine.MarkAllComponentsDirty()
	a.RenderDirty()

	return true
}

// pathToModuleName converts a file path to a Lua module name.
// e.g., "/project/lua/lux/card.lua" → "lux.card"
//
//	"/project/mylib.lua" → "mylib"
//	"/project/lib/utils.lua" → "lib.utils"
func (a *App) pathToModuleName(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ""
	}

	ext := filepath.Ext(absPath)
	if ext != ".lua" {
		return ""
	}

	// Remove .lua extension
	noExt := strings.TrimSuffix(absPath, ext)

	// Handle init.lua → parent directory name
	base := filepath.Base(noExt)
	if base == "init" {
		noExt = filepath.Dir(noExt)
	}

	// Try to find a matching module name by checking package.loaded.
	// Strategy: try progressively shorter path suffixes as module names.
	parts := strings.Split(filepath.ToSlash(noExt), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.Join(parts[i:], ".")
		if a.isModuleLoaded(candidate) {
			return candidate
		}
	}

	// Fallback: just use the filename without extension
	return filepath.Base(noExt)
}

// isModuleLoaded checks if a module name exists in Lua's package.loaded table.
func (a *App) isModuleLoaded(name string) bool {
	L := a.luaState
	top := L.GetTop()
	defer L.SetTop(top)

	L.GetGlobal("package")
	if L.IsNil(-1) {
		return false
	}
	L.GetField(-1, "loaded")
	if L.IsNil(-1) {
		return false
	}
	L.GetField(-1, name)
	return !L.IsNil(-1)
}
