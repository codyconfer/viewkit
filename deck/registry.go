package deck

import (
	"fmt"
	"sort"
	"sync"
)

var (
	regMu sync.RWMutex
	views = map[string]func() View{}
)

// RegisterView associates id with a view constructor (M5 registries).
func RegisterView(id string, ctor func() View) {
	if id == "" || ctor == nil {
		return
	}
	regMu.Lock()
	defer regMu.Unlock()
	if _, ok := views[id]; ok {
		panic(fmt.Sprintf("deck: duplicate view id %q", id))
	}
	views[id] = ctor
}

// LookupView returns a fresh View for id.
func LookupView(id string) (View, bool) {
	regMu.RLock()
	ctor, ok := views[id]
	regMu.RUnlock()
	if !ok {
		return nil, false
	}
	return ctor(), true
}

// ViewIDs returns registered view ids sorted.
func ViewIDs() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]string, 0, len(views))
	for id := range views {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
