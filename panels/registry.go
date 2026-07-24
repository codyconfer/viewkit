package panels

import (
	"fmt"
	"sort"
	"sync"
)

var (
	regMu  sync.RWMutex
	panels = map[string]func() DualHost{}
)

// Register associates id with a DualHost panel constructor (M5 panel registry).
func Register(id string, ctor func() DualHost) {
	if id == "" || ctor == nil {
		return
	}
	regMu.Lock()
	defer regMu.Unlock()
	if _, ok := panels[id]; ok {
		panic(fmt.Sprintf("panels: duplicate id %q", id))
	}
	panels[id] = ctor
}

// Lookup returns a fresh DualHost for id.
func Lookup(id string) (DualHost, bool) {
	regMu.RLock()
	ctor, ok := panels[id]
	regMu.RUnlock()
	if !ok {
		return nil, false
	}
	return ctor(), true
}

// IDs returns registered panel ids sorted.
func IDs() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]string, 0, len(panels))
	for id := range panels {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
