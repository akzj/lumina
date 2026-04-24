package lumina

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// FetchState represents the state of a fetch operation.
type FetchState struct {
	Data    any
	Loading bool
	Error   error
}

// QueryCache manages cached query results.
type QueryCache struct {
	entries map[string]*QueryEntry
	mu      sync.RWMutex
}

// QueryEntry is a single cached query result.
type QueryEntry struct {
	Key        string
	Data       any
	Error      error
	FetchedAt  time.Time
	StaleTime  time.Duration
	Loading    bool
	Refetching bool
}

// IsStale returns whether the entry is stale.
func (e *QueryEntry) IsStale() bool {
	if e.StaleTime <= 0 {
		return true
	}
	return time.Since(e.FetchedAt) > e.StaleTime
}

var globalQueryCache = &QueryCache{
	entries: make(map[string]*QueryEntry),
}

// GetQueryCache returns the global query cache.
func GetQueryCache() *QueryCache {
	return globalQueryCache
}

// Get returns a cached entry, or nil if not found.
func (qc *QueryCache) Get(key string) *QueryEntry {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.entries[key]
}

// Set stores a query result in the cache.
func (qc *QueryCache) Set(key string, data any, err error, staleTime time.Duration) {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	qc.entries[key] = &QueryEntry{
		Key:       key,
		Data:      data,
		Error:     err,
		FetchedAt: time.Now(),
		StaleTime: staleTime,
	}
}

// Invalidate removes a cache entry.
func (qc *QueryCache) Invalidate(key string) {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	delete(qc.entries, key)
}

// InvalidateAll clears all cache entries.
func (qc *QueryCache) InvalidateAll() {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	qc.entries = make(map[string]*QueryEntry)
}

// Size returns the number of cache entries.
func (qc *QueryCache) Size() int {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return len(qc.entries)
}

// Reset clears the cache (for testing).
func (qc *QueryCache) Reset() {
	qc.InvalidateAll()
}

// Fetch performs an HTTP GET request and returns the body as a string.
func Fetch(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("fetch error: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fetch read error: %w", err)
	}
	return string(body), nil
}

// FetchWithOptions performs an HTTP request with options.
type FetchOptions struct {
	Method  string
	Headers map[string]string
	Body    io.Reader
	Timeout time.Duration
}

// FetchAdvanced performs an HTTP request with full options.
func FetchAdvanced(url string, opts FetchOptions) (string, int, error) {
	if opts.Method == "" {
		opts.Method = "GET"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	client := &http.Client{Timeout: opts.Timeout}
	req, err := http.NewRequest(opts.Method, url, opts.Body)
	if err != nil {
		return "", 0, err
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}

	return string(body), resp.StatusCode, nil
}
