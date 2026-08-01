package keys

import "strings"

// Action names a user intent ("confirm", "nav.up") independent of the keys
// bound to it. Views match on Actions so key layout can change per scheme.
type Action string

// Binding ties one or more key names to an Action. Glyph and Label are
// optional footer-hint text: Glyph is the display form of the keys and Label
// describes the effect.
type Binding struct {
	Keys   []string
	Action Action
	Glyph  string
	Label  string
}

// DisplayGlyph returns Glyph, or the key names joined with "/" when no glyph
// is set.
func (b Binding) DisplayGlyph() string {
	if b.Glyph != "" {
		return b.Glyph
	}
	return strings.Join(b.Keys, "/")
}

// WithGlyph returns a copy of b with Glyph set to glyph.
func (b Binding) WithGlyph(glyph string) Binding {
	b.Glyph = glyph
	return b
}

// WithLabel returns a copy of b with Label set to label.
func (b Binding) WithLabel(label string) Binding {
	b.Label = label
	return b
}

// Map indexes a fixed set of bindings both by key name (for input dispatch)
// and by Action (for hints and Has checks). Build one with NewMap or MapFor.
type Map struct {
	byKey map[string]Action
	byAct map[Action]Binding
}

// NewMap builds a Map from bindings. When two bindings share an action or a
// key, the later one wins.
func NewMap(bindings ...Binding) *Map {
	m := &Map{
		byKey: make(map[string]Action, len(bindings)*2),
		byAct: make(map[Action]Binding, len(bindings)),
	}
	for _, b := range bindings {
		m.byAct[b.Action] = b
		for _, k := range b.Keys {
			m.byKey[k] = b.Action
		}
	}
	return m
}

// Action returns the action bound to the given input key, reporting whether
// the key is mapped.
func (m *Map) Action(input string) (a Action, ok bool) {
	a, ok = m.byKey[input]
	return a, ok
}

// Has reports whether the map carries a binding for a.
func (m *Map) Has(a Action) bool {
	_, ok := m.byAct[a]
	return ok
}

// Hint is one footer-legend entry: the key (or context name) on the left and
// its label (or value) on the right.
type Hint struct {
	Key   string
	Label string
}

// Hint returns the footer hint for a: the binding's display glyph paired with
// its label. An unbound action yields an empty Hint.
func (m *Map) Hint(a Action) Hint {
	b := m.byAct[a]
	return Hint{Key: b.DisplayGlyph(), Label: b.Label}
}

// HintLabeled is Hint with the binding's label replaced by label.
func (m *Map) HintLabeled(a Action, label string) Hint {
	return Hint{Key: m.byAct[a].DisplayGlyph(), Label: label}
}

// Hints returns footer hints for the given actions in order, keeping only
// bindings that carry an explicit Glyph; unbound or glyph-less actions are
// skipped rather than falling back to raw key names.
func (m *Map) Hints(actions ...Action) []Hint {
	out := make([]Hint, 0, len(actions))
	for _, a := range actions {
		if b, ok := m.byAct[a]; ok && b.Glyph != "" {
			out = append(out, Hint{Key: b.Glyph, Label: b.Label})
		}
	}
	return out
}

// MapFor builds a Map from this scheme's bindings for the given actions.
func (s Scheme) MapFor(actions ...Action) *Map {
	bindings := make([]Binding, 0, len(actions))
	for _, a := range actions {
		bindings = append(bindings, s.Binding(a))
	}
	return NewMap(bindings...)
}
