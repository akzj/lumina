package lumina

import (
	"fmt"
	"sync"
)

// Plugin represents an extension to the Lumina framework.
type Plugin struct {
	Name       string
	Version    string
	InitFn     func() error          // Go-level init
	Hooks      map[string]any        // custom hooks (name -> impl)
	Components map[string]string     // name -> lua source
	Metadata   map[string]string     // arbitrary metadata
	initialized bool
}

// PluginRegistry manages all registered plugins.
type PluginRegistry struct {
	plugins map[string]*Plugin
	order   []string // insertion order
	mu      sync.RWMutex
}

var globalPluginRegistry = NewPluginRegistry()

// NewPluginRegistry creates a new plugin registry.
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]*Plugin),
	}
}

// GetPluginRegistry returns the global plugin registry.
func GetPluginRegistry() *PluginRegistry {
	return globalPluginRegistry
}

// Register adds a plugin to the registry.
func (pr *PluginRegistry) Register(plugin *Plugin) error {
	if plugin.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if existing, ok := pr.plugins[plugin.Name]; ok {
		if existing.Version != plugin.Version {
			return fmt.Errorf("plugin %q version conflict: %s vs %s",
				plugin.Name, existing.Version, plugin.Version)
		}
		return nil // same version, idempotent
	}
	pr.plugins[plugin.Name] = plugin
	pr.order = append(pr.order, plugin.Name)
	return nil
}

// Get returns a plugin by name.
func (pr *PluginRegistry) Get(name string) (*Plugin, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	p, ok := pr.plugins[name]
	return p, ok
}

// InitAll initializes all registered plugins in order.
func (pr *PluginRegistry) InitAll() error {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	for _, name := range pr.order {
		p := pr.plugins[name]
		if p.initialized {
			continue
		}
		if p.InitFn != nil {
			if err := p.InitFn(); err != nil {
				return fmt.Errorf("plugin %q init failed: %w", name, err)
			}
		}
		p.initialized = true
	}
	return nil
}

// InitPlugin initializes a single plugin by name.
func (pr *PluginRegistry) InitPlugin(name string) error {
	pr.mu.RLock()
	p, ok := pr.plugins[name]
	pr.mu.RUnlock()
	if !ok {
		return fmt.Errorf("plugin %q not found", name)
	}
	if p.initialized {
		return nil
	}
	if p.InitFn != nil {
		if err := p.InitFn(); err != nil {
			return fmt.Errorf("plugin %q init failed: %w", name, err)
		}
	}
	pr.mu.Lock()
	p.initialized = true
	pr.mu.Unlock()
	return nil
}

// List returns all registered plugin names.
func (pr *PluginRegistry) List() []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	result := make([]string, len(pr.order))
	copy(result, pr.order)
	return result
}

// Count returns the number of registered plugins.
func (pr *PluginRegistry) Count() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	return len(pr.plugins)
}

// Clear removes all plugins (for testing).
func (pr *PluginRegistry) Clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.plugins = make(map[string]*Plugin)
	pr.order = nil
}
