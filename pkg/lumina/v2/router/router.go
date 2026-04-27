// Package router provides SPA-style routing for Lumina v2 apps.
//
// It supports static paths, named parameters (":id"), and wildcard
// segments ("*path"). Pure Go, zero v2 dependencies.
package router

import (
	"strings"
)

// Route represents a route definition with pattern matching.
type Route struct {
	path       string   // original path e.g. "/users/:id"
	segments   []string // split segments
	paramNames []string // e.g. ["id"]
}

// Path returns the original route pattern.
func (r *Route) Path() string { return r.path }

// MatchResult contains the matched route and extracted params.
type MatchResult struct {
	Route  *Route
	Params map[string]string
}

// Router provides SPA-style routing with history and onChange callbacks.
type Router struct {
	routes      []*Route
	currentPath string
	history     []string
	params      map[string]string
	onChange     []func(path string, params map[string]string)
}

// New creates a new Router with initial path "/".
func New() *Router {
	return &Router{
		currentPath: "/",
		params:      make(map[string]string),
	}
}

// AddRoute registers a route pattern.
// Patterns support:
//   - Static segments: "/users/list"
//   - Named params: "/users/:id"
//   - Wildcard: "/files/*path" (catches rest of path)
func (r *Router) AddRoute(pattern string) {
	route := &Route{path: pattern}
	segs := splitPath(pattern)
	route.segments = segs
	for _, seg := range segs {
		if strings.HasPrefix(seg, ":") {
			route.paramNames = append(route.paramNames, seg[1:])
		} else if strings.HasPrefix(seg, "*") {
			route.paramNames = append(route.paramNames, seg[1:])
		}
	}
	r.routes = append(r.routes, route)
}

// Navigate changes the current path, matches route, extracts params,
// pushes previous path to history, and triggers onChange callbacks.
func (r *Router) Navigate(path string) {
	// Push current path to history before navigating.
	if r.currentPath != "" {
		r.history = append(r.history, r.currentPath)
	}
	r.currentPath = path
	r.params = make(map[string]string)

	if result := r.Match(path); result != nil {
		r.params = result.Params
	}

	// Copy params for callbacks (callers should not mutate internal state).
	cbParams := copyParams(r.params)
	for _, fn := range r.onChange {
		fn(path, cbParams)
	}
}

// Back pops the last path from history and navigates to it.
// Returns false if history is empty.
func (r *Router) Back() bool {
	if len(r.history) == 0 {
		return false
	}

	prev := r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]
	r.currentPath = prev
	r.params = make(map[string]string)

	if result := r.Match(prev); result != nil {
		r.params = result.Params
	}

	cbParams := copyParams(r.params)
	for _, fn := range r.onChange {
		fn(prev, cbParams)
	}
	return true
}

// Match matches a path against registered routes.
// Returns nil if no route matches.
func (r *Router) Match(path string) *MatchResult {
	segments := splitPath(path)
	for _, route := range r.routes {
		if params, ok := matchRoute(route, segments); ok {
			return &MatchResult{Route: route, Params: params}
		}
	}
	return nil
}

// CurrentPath returns the current path.
func (r *Router) CurrentPath() string {
	return r.currentPath
}

// Params returns a copy of the current route parameters.
func (r *Router) Params() map[string]string {
	return copyParams(r.params)
}

// History returns a copy of the navigation history.
func (r *Router) History() []string {
	result := make([]string, len(r.history))
	copy(result, r.history)
	return result
}

// RouteCount returns the number of registered routes.
func (r *Router) RouteCount() int {
	return len(r.routes)
}

// OnChange registers a callback for path changes.
// The callback receives the new path and extracted parameters.
func (r *Router) OnChange(fn func(path string, params map[string]string)) {
	r.onChange = append(r.onChange, fn)
}

// -----------------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------------

// splitPath splits a URL path into non-empty segments.
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

// matchRoute checks if path segments match a route's pattern.
// Supports static segments, named params (":name"), and wildcards ("*name").
func matchRoute(route *Route, pathSegments []string) (map[string]string, bool) {
	pattern := route.segments

	// Check for wildcard as last segment.
	hasWildcard := len(pattern) > 0 && strings.HasPrefix(pattern[len(pattern)-1], "*")

	if hasWildcard {
		// Wildcard route: path must have at least len(pattern)-1 segments.
		if len(pathSegments) < len(pattern)-1 {
			return nil, false
		}
	} else {
		// Non-wildcard: exact segment count required.
		if len(pattern) != len(pathSegments) {
			return nil, false
		}
	}

	params := make(map[string]string)
	for i, seg := range pattern {
		if strings.HasPrefix(seg, "*") {
			// Wildcard: capture the rest of the path.
			params[seg[1:]] = strings.Join(pathSegments[i:], "/")
			return params, true
		}
		if i >= len(pathSegments) {
			return nil, false
		}
		if strings.HasPrefix(seg, ":") {
			params[seg[1:]] = pathSegments[i]
		} else if seg != pathSegments[i] {
			return nil, false
		}
	}
	return params, true
}

// copyParams returns a shallow copy of a params map.
func copyParams(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
