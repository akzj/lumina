package lumina

import (
	"strings"
	"sync"
)

// Route represents a single route definition.
type Route struct {
	Path       string   // e.g., "/users/:id"
	Pattern    []string // split path segments
	ParamNames []string // named params (e.g., ["id"])
}

// Router provides SPA-style routing for Lumina apps.
type Router struct {
	routes      []*Route
	currentPath string
	history     []string
	params      map[string]string
	onChange    []func(path string)
	mu          sync.RWMutex
}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{
		currentPath: "/",
		params:      make(map[string]string),
	}
}

// AddRoute adds a route definition.
func (r *Router) AddRoute(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	route := &Route{Path: path}
	segments := splitPath(path)
	route.Pattern = segments
	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			route.ParamNames = append(route.ParamNames, seg[1:])
		}
	}
	r.routes = append(r.routes, route)
}

// Navigate changes the current path and triggers onChange callbacks.
func (r *Router) Navigate(path string) {
	r.mu.Lock()

	// Push current path to history before navigating
	if r.currentPath != "" {
		r.history = append(r.history, r.currentPath)
	}
	r.currentPath = path
	r.params = make(map[string]string)

	// Match route and extract params
	for _, route := range r.routes {
		if params, ok := matchRoute(route.Pattern, splitPath(path)); ok {
			r.params = params
			break
		}
	}

	// Copy callbacks, release lock, then notify (avoids deadlock if callback navigates)
	callbacks := make([]func(string), len(r.onChange))
	copy(callbacks, r.onChange)
	r.mu.Unlock()

	for _, fn := range callbacks {
		fn(path)
	}
}

// Back navigates to the previous path in history.
// Returns true if navigation occurred, false if history is empty.
func (r *Router) Back() bool {
	r.mu.Lock()

	if len(r.history) == 0 {
		r.mu.Unlock()
		return false
	}

	// Pop from history
	prev := r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]
	r.currentPath = prev
	r.params = make(map[string]string)

	// Re-match params
	for _, route := range r.routes {
		if params, ok := matchRoute(route.Pattern, splitPath(prev)); ok {
			r.params = params
			break
		}
	}

	// Copy callbacks, release lock, then notify
	callbacks := make([]func(string), len(r.onChange))
	copy(callbacks, r.onChange)
	r.mu.Unlock()

	for _, fn := range callbacks {
		fn(prev)
	}
	return true
}

// GetCurrentPath returns the current path.
func (r *Router) GetCurrentPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentPath
}

// GetParams returns the current route parameters.
func (r *Router) GetParams() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]string, len(r.params))
	for k, v := range r.params {
		result[k] = v
	}
	return result
}

// Match finds the matching route for a path and extracts parameters.
func (r *Router) Match(path string) (*Route, map[string]string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	segments := splitPath(path)
	for _, route := range r.routes {
		if params, ok := matchRoute(route.Pattern, segments); ok {
			return route, params
		}
	}
	return nil, nil
}

// OnChange registers a callback for route changes.
func (r *Router) OnChange(fn func(path string)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onChange = append(r.onChange, fn)
}

// GetHistory returns a copy of the navigation history.
func (r *Router) GetHistory() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]string, len(r.history))
	copy(result, r.history)
	return result
}

// RouteCount returns the number of registered routes.
func (r *Router) RouteCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.routes)
}

// -----------------------------------------------------------------------
// Path matching helpers
// -----------------------------------------------------------------------

// splitPath splits a URL path into segments, ignoring empty segments.
func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// matchRoute checks if path segments match a route pattern.
// Returns extracted parameters if matched.
func matchRoute(pattern, segments []string) (map[string]string, bool) {
	if len(pattern) != len(segments) {
		return nil, false
	}

	params := make(map[string]string)
	for i, pat := range pattern {
		if strings.HasPrefix(pat, ":") {
			// Named parameter — matches anything
			params[pat[1:]] = segments[i]
		} else if pat != segments[i] {
			return nil, false
		}
	}
	return params, true
}

// globalRouter is the singleton router instance.
var globalRouter = NewRouter()
