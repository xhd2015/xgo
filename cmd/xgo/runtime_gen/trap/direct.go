package trap

import "sync"

var bypassMapping sync.Map // <goroutine key> -> struct{}{}

// Direct make a call to fn, without
// any trap and mock interceptors
func Direct(fn func()) {
	key := uintptr(__xgo_link_getcurg())
	bypassMapping.Store(key, struct{}{})
	defer bypassMapping.Delete(key)
	fn()
}

func isByPassing() bool {
	key := uintptr(__xgo_link_getcurg())
	_, ok := bypassMapping.Load(key)
	return ok
}
