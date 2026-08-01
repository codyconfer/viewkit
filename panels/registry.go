package panels

import (
	"sort"
	"sync"
)

var (
	regMu  sync.RWMutex
	panels = map[string]func() DualHost{}
)

// Register registers ctor under id. First registration wins; duplicate,
// empty, or nil registrations change nothing and return false.
func Register(id string, ctor func() DualHost) bool {
	if id == "" || ctor == nil {
		return false
	}
	regMu.Lock()
	defer regMu.Unlock()
	if _, ok := panels[id]; ok {
		return false
	}
	panels[id] = ctor
	return true
}

// Named returns a fresh DualHost for id.
func Named(id string) (DualHost, bool) {
	regMu.RLock()
	ctor, ok := panels[id]
	regMu.RUnlock()
	if !ok {
		return nil, false
	}
	return ctor(), true
}

// Keys returns registered panel ids sorted.
func Keys() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]string, 0, len(panels))
	for id := range panels {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
