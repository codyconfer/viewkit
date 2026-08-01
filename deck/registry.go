package deck

import (
	"sort"
	"sync"
)

var (
	regMu sync.RWMutex
	views = map[string]func() View{}
)

// RegisterView registers ctor under id. First registration wins; duplicate,
// empty, or nil registrations change nothing and return false.
func RegisterView(id string, ctor func() View) bool {
	if id == "" || ctor == nil {
		return false
	}
	regMu.Lock()
	defer regMu.Unlock()
	if _, ok := views[id]; ok {
		return false
	}
	views[id] = ctor
	return true
}

// NamedView returns a fresh View for id.
func NamedView(id string) (View, bool) {
	regMu.RLock()
	ctor, ok := views[id]
	regMu.RUnlock()
	if !ok {
		return nil, false
	}
	return ctor(), true
}

// ViewKeys returns registered view ids sorted.
func ViewKeys() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]string, 0, len(views))
	for id := range views {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
