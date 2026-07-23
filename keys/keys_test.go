package keys

import "testing"

func TestMapResolvesAliases(t *testing.T) {
	m := NewMap(
		Binding{Keys: []string{"up", "k"}, Action: Up, Glyph: "↑/↓/j/k"},
		Binding{Keys: []string{"down", "j"}, Action: Down},
	)
	for _, in := range []string{"up", "k"} {
		if a, ok := m.Action(in); !ok || a != Up {
			t.Fatalf("Action(%q) = %q,%v; want %q,true", in, a, ok, Up)
		}
	}
	if a, ok := m.Action("j"); !ok || a != Down {
		t.Fatalf("Action(j) = %q,%v; want %q,true", a, ok, Down)
	}
	if _, ok := m.Action("z"); ok {
		t.Fatal("Action(z) resolved; want unbound")
	}
}

func TestMatcherOnlyBindingEmitsNoHint(t *testing.T) {
	m := NewMap(
		Binding{Keys: []string{"up", "k"}, Action: Up, Glyph: "↑/↓/j/k", Label: "select"},
		Binding{Keys: []string{"down", "j"}, Action: Down},
	)
	hints := m.Hints(Up, Down)
	if len(hints) != 1 {
		t.Fatalf("Hints len = %d; want 1 (Down is matcher-only)", len(hints))
	}
	if hints[0] != [2]string{"↑/↓/j/k", "select"} {
		t.Fatalf("Hints[0] = %v; want [↑/↓/j/k select]", hints[0])
	}
	if !m.Has(Down) {
		t.Fatal("Down should still be bound for matching")
	}
}

func TestDisplayGlyphFallsBackToKeys(t *testing.T) {
	b := Binding{Keys: []string{"y", "z"}, Action: Confirm}
	if got := b.DisplayGlyph(); got != "y/z" {
		t.Fatalf("DisplayGlyph = %q; want y/z", got)
	}
}

func TestHintLabeledOverridesLabel(t *testing.T) {
	m := NewMap(Binding{Keys: []string{"up"}, Action: Up, Glyph: "↑/↓/j/k", Label: "select"})
	if got := m.HintLabeled(Up, "scroll feed"); got != [2]string{"↑/↓/j/k", "scroll feed"} {
		t.Fatalf("HintLabeled = %v; want [↑/↓/j/k scroll feed]", got)
	}
}

func TestSchemeWithOverridesAreConfigurable(t *testing.T) {
	sc := Default().With(Binding{Keys: []string{"w"}, Action: Up, Glyph: "w"})
	m := NewMap(sc.Binding(Up))
	if a, ok := m.Action("w"); !ok || a != Up {
		t.Fatalf("Action(w) = %q,%v; want %q,true", a, ok, Up)
	}

	if _, ok := NewMap(Default().Binding(Up)).Action("w"); ok {
		t.Fatal("Default scheme mutated by With")
	}
}

func TestNamedReturnsRegisteredScheme(t *testing.T) {
	if _, ok := Named("default"); !ok {
		t.Fatal("Named(default) not found")
	}
	if _, ok := Named("does-not-exist"); ok {
		t.Fatal("Named(unknown) should report not found")
	}
}

func TestKeysDefaultFirst(t *testing.T) {
	keys := Keys()
	if len(keys) == 0 || keys[0] != "default" {
		t.Fatalf("Keys() = %v, want default first", keys)
	}
}

func TestRegisterAddsNamedScheme(t *testing.T) {
	orig := registry
	defer func() { registry = orig }()

	sc := Default().With(Binding{Keys: []string{"w"}, Action: Up, Glyph: "w"})
	Register("wasd", "WASD", sc)

	got, ok := Named("wasd")
	if !ok {
		t.Fatal("Named(wasd) not found after Register")
	}
	m := NewMap(got.Binding(Up))
	if a, ok := m.Action("w"); !ok || a != Up {
		t.Fatalf("registered scheme Action(w) = %q,%v; want %q,true", a, ok, Up)
	}
	if got := DisplayName("wasd"); got != "WASD" {
		t.Fatalf("DisplayName(wasd) = %q, want %q", got, "WASD")
	}

	var found bool
	for _, k := range Keys() {
		if k == "wasd" {
			found = true
		}
	}
	if !found {
		t.Fatal("Keys() does not include registered key")
	}
}

func TestRegisterReplacesExistingKey(t *testing.T) {
	orig := registry
	defer func() { registry = orig }()

	before := len(registry)
	Register("dup", "First", Default())
	Register("dup", "Second", Default())

	if got := len(registry); got != before+1 {
		t.Fatalf("registry len = %d, want %d (no duplicate entries)", got, before+1)
	}
	if got := DisplayName("dup"); got != "Second" {
		t.Fatalf("DisplayName(dup) = %q, want %q", got, "Second")
	}
}
