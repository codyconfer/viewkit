package deck

import (
	"fmt"
	"sort"
	"sync"
)

// Component is a reusable UI fragment registered by id.
// Components are not full Views — they paint a region when asked.
type Component interface {
	Render(width, height int) string
}

var (
	compMu     sync.RWMutex
	components = map[string]func() Component{}
)

// RegisterComponent associates id with a component constructor.
func RegisterComponent(id string, ctor func() Component) {
	if id == "" || ctor == nil {
		return
	}
	compMu.Lock()
	defer compMu.Unlock()
	if _, ok := components[id]; ok {
		panic(fmt.Sprintf("deck: duplicate component id %q", id))
	}
	components[id] = ctor
}

// LookupComponent returns a fresh Component for id.
func LookupComponent(id string) (Component, bool) {
	compMu.RLock()
	ctor, ok := components[id]
	compMu.RUnlock()
	if !ok {
		return nil, false
	}
	return ctor(), true
}

// ComponentIDs returns registered component ids sorted.
func ComponentIDs() []string {
	compMu.RLock()
	defer compMu.RUnlock()
	out := make([]string, 0, len(components))
	for id := range components {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
