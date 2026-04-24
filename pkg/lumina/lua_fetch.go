package lumina

import (
	"fmt"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// luaFetch implements lumina.fetch(url) -> data, error
// Synchronous HTTP GET (for use inside useQuery fetcher functions).
func luaFetch(L *lua.State) int {
	url := L.CheckString(1)
	data, err := Fetch(url)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}
	L.PushString(data)
	L.PushNil()
	return 2
}

// luaUseFetch implements lumina.useFetch(url) -> { data, loading, error }
// Returns a table with the fetch state.
func luaUseFetch(L *lua.State) int {
	url := L.CheckString(1)

	// Check cache first
	cache := GetQueryCache()
	entry := cache.Get("fetch:" + url)

	if entry != nil && !entry.IsStale() {
		// Return cached result
		L.NewTable()
		if entry.Data != nil {
			L.PushAny(entry.Data)
		} else {
			L.PushNil()
		}
		L.SetField(-2, "data")
		L.PushBoolean(false)
		L.SetField(-2, "loading")
		if entry.Error != nil {
			L.PushString(entry.Error.Error())
		} else {
			L.PushNil()
		}
		L.SetField(-2, "error")
		return 1
	}

	// Perform fetch (synchronous for now)
	data, err := Fetch(url)
	cache.Set("fetch:"+url, data, err, 60*time.Second)

	L.NewTable()
	if err != nil {
		L.PushNil()
		L.SetField(-2, "data")
		L.PushBoolean(false)
		L.SetField(-2, "loading")
		L.PushString(err.Error())
		L.SetField(-2, "error")
	} else {
		L.PushString(data)
		L.SetField(-2, "data")
		L.PushBoolean(false)
		L.SetField(-2, "loading")
		L.PushNil()
		L.SetField(-2, "error")
	}
	return 1
}

// luaUseQuery implements lumina.useQuery(key, fetcherFn, options) -> { data, loading, error, refetch }
// Cached data fetching with stale-while-revalidate.
func luaUseQuery(L *lua.State) int {
	key := L.CheckString(1)
	// arg2: fetcher function (Lua function)
	if L.Type(2) != lua.TypeFunction {
		L.PushString("useQuery: expected function as second argument")
		L.Error()
		return 0
	}

	// Read options (arg3, optional table)
	staleTime := 60 * time.Second
	if L.Type(3) == lua.TypeTable {
		L.GetField(3, "staleTime")
		if st, ok := L.ToInteger(-1); ok && st > 0 {
			staleTime = time.Duration(st) * time.Second
		}
		L.Pop(1)
	}

	cache := GetQueryCache()
	entry := cache.Get(key)

	// If cached and not stale, return cached
	if entry != nil && !entry.IsStale() {
		pushQueryResult(L, entry)
		return 1
	}

	// Call the fetcher function
	L.PushValue(2) // push fetcher
	if status := L.PCall(0, 1, 0); status != 0 {
		// Fetcher errored — error message is on stack
		errMsg := "query fetcher error"
		if s, ok := L.ToString(-1); ok {
			errMsg = s
		}
		L.Pop(1)
		cache.Set(key, nil, fmt.Errorf("%s", errMsg), staleTime)
		L.NewTable()
		L.PushNil()
		L.SetField(-2, "data")
		L.PushBoolean(false)
		L.SetField(-2, "loading")
		L.PushString(errMsg)
		L.SetField(-2, "error")
		return 1
	}

	// Get result from stack
	data := L.ToAny(-1)
	L.Pop(1)

	cache.Set(key, data, nil, staleTime)

	L.NewTable()
	L.PushAny(data)
	L.SetField(-2, "data")
	L.PushBoolean(false)
	L.SetField(-2, "loading")
	L.PushNil()
	L.SetField(-2, "error")
	return 1
}

// luaInvalidateQuery implements lumina.invalidateQuery(key)
func luaInvalidateQuery(L *lua.State) int {
	key := L.CheckString(1)
	GetQueryCache().Invalidate(key)
	return 0
}

// luaInvalidateAllQueries implements lumina.invalidateAllQueries()
func luaInvalidateAllQueries(L *lua.State) int {
	GetQueryCache().InvalidateAll()
	return 0
}

func pushQueryResult(L *lua.State, entry *QueryEntry) {
	L.NewTable()
	if entry.Data != nil {
		L.PushAny(entry.Data)
	} else {
		L.PushNil()
	}
	L.SetField(-2, "data")
	L.PushBoolean(entry.Loading)
	L.SetField(-2, "loading")
	if entry.Error != nil {
		L.PushString(entry.Error.Error())
	} else {
		L.PushNil()
	}
	L.SetField(-2, "error")
}
