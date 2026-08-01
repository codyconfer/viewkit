package keys

import (
	"sync"
	"sync/atomic"
)

// Standard actions every scheme is expected to bind: cursor and focus
// movement, confirm/cancel/quit, value adjustment (Inc/Dec), paging, and the
// text-completion trio. Default binds all of them.
const (
	Up        Action = "nav.up"
	Down      Action = "nav.down"
	Left      Action = "nav.left"
	Right     Action = "nav.right"
	Confirm   Action = "confirm"
	Cancel    Action = "cancel"
	Quit      Action = "quit"
	FocusNext Action = "focus.next"
	FocusPrev Action = "focus.prev"
	Inc       Action = "inc"
	Dec       Action = "dec"
	Erase     Action = "erase"
	PageUp    Action = "page.up"
	PageDown  Action = "page.down"
	Open      Action = "open"
	Reload    Action = "reload"

	// Complete accepts the suggestion a text field is showing. It shares tab
	// with FocusNext: a host that binds both resolves the conflict by build
	// order, and one that binds Complete alone gets tab-to-accept for free.
	Complete     Action = "complete"
	CompleteNext Action = "complete.next"
	CompletePrev Action = "complete.prev"
)

// Scheme is a complete keymap: one Binding per Action. Schemes are treated as
// immutable; With and WithDefaults return modified copies.
type Scheme struct {
	bindings map[Action]Binding
}

// Binding returns the binding for a, or the zero Binding when a is unbound.
func (s Scheme) Binding(a Action) Binding {
	return s.bindings[a]
}

// With returns a copy of s with overrides applied, each replacing any
// existing binding for its action. s itself is not modified.
func (s Scheme) With(overrides ...Binding) Scheme {
	next := Scheme{bindings: make(map[Action]Binding, len(s.bindings))}
	for k, v := range s.bindings {
		next.bindings[k] = v
	}
	for _, b := range overrides {
		next.bindings[b.Action] = b
	}
	return next
}

// WithDefaults fills every action missing from s (or bound without keys) from
// defaults; bindings already in s win. Use it to graft app-specific actions
// onto a user-selected scheme instead of falling back at each lookup.
func (s Scheme) WithDefaults(defaults ...Binding) Scheme {
	next := Scheme{bindings: make(map[Action]Binding, len(s.bindings)+len(defaults))}
	for k, v := range s.bindings {
		next.bindings[k] = v
	}
	for _, b := range defaults {
		if have, ok := next.bindings[b.Action]; !ok || len(have.Keys) == 0 {
			next.bindings[b.Action] = b
		}
	}
	return next
}

// Default returns the built-in scheme: arrow keys doubled with vi-style
// h/j/k/l, enter/space to confirm, esc to cancel, ctrl+c to quit, tab focus
// cycling, and bracket/plus/minus value adjustment.
func Default() Scheme {
	return Scheme{bindings: map[Action]Binding{
		Up:        {Keys: []string{"up", "k"}, Action: Up, Glyph: "↑/↓/j/k"},
		Down:      {Keys: []string{"down", "j"}, Action: Down},
		Left:      {Keys: []string{"left", "h"}, Action: Left, Glyph: "←/→/h/l"},
		Right:     {Keys: []string{"right", "l"}, Action: Right},
		Confirm:   {Keys: []string{"enter", " ", "spacebar"}, Action: Confirm, Glyph: "enter/space"},
		Cancel:    {Keys: []string{"esc"}, Action: Cancel},
		Quit:      {Keys: []string{"ctrl+c"}, Action: Quit},
		FocusNext: {Keys: []string{"tab"}, Action: FocusNext, Glyph: "tab/⇧tab", Label: "focus panel"},
		FocusPrev: {Keys: []string{"shift+tab"}, Action: FocusPrev},
		Inc:       {Keys: []string{"]", "+", "="}, Action: Inc, Glyph: "[ ]/-/+"},
		Dec:       {Keys: []string{"[", "-", "_"}, Action: Dec},
		Erase:     {Keys: []string{"backspace", "ctrl+h"}, Action: Erase, Glyph: "backspace"},
		PageUp:    {Keys: []string{"pgup"}, Action: PageUp, Glyph: "pgup/pgdn"},
		PageDown:  {Keys: []string{"pgdown"}, Action: PageDown},
		Open:      {Keys: []string{"o"}, Action: Open, Glyph: "o", Label: "open"},
		Reload:    {Keys: []string{"r", "f5"}, Action: Reload, Glyph: "r", Label: "reload"},

		Complete:     {Keys: []string{"tab"}, Action: Complete, Glyph: "tab", Label: "accept"},
		CompleteNext: {Keys: []string{"ctrl+n"}, Action: CompleteNext, Glyph: "ctrl+n/ctrl+p", Label: "suggestion"},
		CompletePrev: {Keys: []string{"ctrl+p"}, Action: CompletePrev, Glyph: "ctrl+p"},
	}}
}

var current = func() *atomic.Pointer[Scheme] {
	p := new(atomic.Pointer[Scheme])
	s := Default()
	p.Store(&s)
	return p
}()

// Cur returns the active scheme by value (matching theme.Cur); a later Use
// does not change a Scheme already returned.
func Cur() Scheme { return *current.Load() }

// Use makes s the active scheme. Safe to call while other goroutines read Cur.
func Use(s Scheme) { current.Store(&s) }

type registryEntry struct {
	key    string
	name   string
	scheme Scheme
}

var (
	registryMu sync.RWMutex
	registry   = []registryEntry{
		{key: "default", name: "Default", scheme: Default()},
	}
)

// Register adds scheme s to the registry under key with the given display
// name. A later call with the same key replaces the earlier name and scheme
// in place, so apps can override the built-in entry. Safe for concurrent use.
func Register(key, name string, s Scheme) {
	registryMu.Lock()
	defer registryMu.Unlock()
	for i := range registry {
		if registry[i].key == key {
			registry[i].name = name
			registry[i].scheme = s
			return
		}
	}
	registry = append(registry, registryEntry{key: key, name: name, scheme: s})
}

// Keys returns the registered scheme keys in registration order, "default"
// first.
func Keys() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, len(registry))
	for i, e := range registry {
		out[i] = e.key
	}
	return out
}

// Activate looks up the named scheme and makes it active, replacing the
// Named-then-Use two-step. It reports whether key was registered; the active
// scheme is left unchanged when it was not.
func Activate(key string) bool {
	s, ok := Named(key)
	if !ok {
		return false
	}
	Use(s)
	return true
}

// Named returns the scheme registered under key. When key is not registered
// it returns Default() and false.
func Named(key string) (Scheme, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	for _, e := range registry {
		if e.key == key {
			return e.scheme, true
		}
	}
	return Default(), false
}

// DisplayName returns the human-readable name registered for key, or key
// itself when it is not registered.
func DisplayName(key string) string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	for _, e := range registry {
		if e.key == key {
			return e.name
		}
	}
	return key
}
