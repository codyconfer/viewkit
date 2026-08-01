package deck

import (
	"sort"
	"sync"

	"github.com/codyconfer/viewkit/layout"
)

// Component is a reusable UI fragment registered by id.
// Components are not full Views — they paint a region when asked.
type Component interface {
	Render(f layout.Frame) string
}

var (
	compMu     sync.RWMutex
	components = map[string]func() Component{}
)

// RegisterComponent registers ctor under id. First registration wins; duplicate,
// empty, or nil registrations change nothing and return false.
func RegisterComponent(id string, ctor func() Component) bool {
	if id == "" || ctor == nil {
		return false
	}
	compMu.Lock()
	defer compMu.Unlock()
	if _, ok := components[id]; ok {
		return false
	}
	components[id] = ctor
	return true
}

// NamedComponent returns a fresh Component for id.
func NamedComponent(id string) (Component, bool) {
	compMu.RLock()
	ctor, ok := components[id]
	compMu.RUnlock()
	if !ok {
		return nil, false
	}
	return ctor(), true
}

// ComponentKeys returns registered component ids sorted.
func ComponentKeys() []string {
	compMu.RLock()
	defer compMu.RUnlock()
	out := make([]string, 0, len(components))
	for id := range components {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
