package keys

import (
	"slices"
	"testing"
)

func TestNormalizeFoldsCaseAndTrims(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"esc", "esc"},
		{"ESC", "esc"},
		{"  Ctrl+S  ", "ctrl+s"},
		{"\tTab\n", "tab"},
		{"", ""},
		{"   ", ""},
		{"Shift+Tab", "shift+tab"},
		{"a b", "a b"},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := Normalize(c.in); got != c.want {
				t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestResolveExactAndFuzzyHits(t *testing.T) {
	binds := map[string]string{
		"ctrl+s":  "save",
		" F5 ":    "  reload  ",
		"Shift+G": "bottom",
	}
	cases := []struct {
		name   string
		key    string
		target string
		ok     bool
	}{
		{"exact key", "ctrl+s", "save", true},
		{"uppercase query", "CTRL+S", "save", true},
		{"padded query", "  ctrl+s  ", "save", true},
		{"padded table key", "f5", "reload", true},
		{"padded table key uppercase", "F5", "reload", true},
		{"mixed case table key", "shift+g", "bottom", true},
		{"unknown key", "ctrl+x", "", false},
		{"empty query", "", "", false},
		{"whitespace query", "   ", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			target, ok := Resolve(binds, c.key)
			if target != c.target || ok != c.ok {
				t.Errorf("Resolve(%q) = (%q, %v), want (%q, %v)", c.key, target, ok, c.target, c.ok)
			}
		})
	}
}

func TestResolveTrimsTargetAndRejectsBlankTargets(t *testing.T) {
	binds := map[string]string{
		"a":       "  run  ",
		"b":       "",
		"c":       "   ",
		" D ":     "\t\n",
		"padded ": " go ",
	}
	cases := []struct {
		name   string
		key    string
		target string
		ok     bool
	}{
		{"trimmed target", "a", "run", true},
		{"empty target", "b", "", false},
		{"whitespace target", "c", "", false},
		{"whitespace target via scan", "d", "", false},
		{"trimmed target via scan", "padded", "go", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			target, ok := Resolve(binds, c.key)
			if target != c.target || ok != c.ok {
				t.Errorf("Resolve(%q) = (%q, %v), want (%q, %v)", c.key, target, ok, c.target, c.ok)
			}
		})
	}
}

func TestResolveEmptyTable(t *testing.T) {
	for _, binds := range []map[string]string{nil, {}} {
		if target, ok := Resolve(binds, "ctrl+s"); ok || target != "" {
			t.Errorf("Resolve(%#v, ctrl+s) = (%q, %v), want (\"\", false)", binds, target, ok)
		}
	}
}

func TestResolveDoesNotMutateTable(t *testing.T) {
	binds := map[string]string{" F5 ": "  reload  "}
	if _, ok := Resolve(binds, "f5"); !ok {
		t.Fatal("Resolve(f5) missed")
	}
	if len(binds) != 1 {
		t.Errorf("table grew to %d entries: %#v", len(binds), binds)
	}
	if binds[" F5 "] != "  reload  " {
		t.Errorf("table value rewritten: %q", binds[" F5 "])
	}
}

func TestControlKeysDropsSingleRunes(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"mixed", []string{"tab", "q", "ctrl+s", "i", "esc"}, []string{"tab", "ctrl+s", "esc"}},
		{"all control", []string{"enter", "shift+tab", "pgdown"}, []string{"enter", "shift+tab", "pgdown"}},
		{"all single", []string{"q", "j", "]", "+", " "}, []string{}},
		{"multibyte single rune", []string{"é", "ü", "→", "日"}, []string{}},
		{"multibyte kept when multi rune", []string{"éé", "ctrl+é"}, []string{"éé", "ctrl+é"}},
		{"empty string dropped", []string{"", "tab"}, []string{"tab"}},
		{"order preserved", []string{"esc", "x", "tab"}, []string{"esc", "tab"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := controlKeys(c.in)
			if !slices.Equal(got, c.want) {
				t.Errorf("controlKeys(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestControlKeysCountsRunesNotBytes(t *testing.T) {
	if got := controlKeys([]string{"é"}); len(got) != 0 {
		t.Errorf("controlKeys([é]) = %q, want empty (é is one rune, two bytes)", got)
	}
}

func TestControlKeysEmptyInputReturnsNonNilSlice(t *testing.T) {
	for _, in := range [][]string{nil, {}} {
		got := controlKeys(in)
		if got == nil {
			t.Errorf("controlKeys(%#v) = nil, want empty non-nil slice", in)
		}
		if len(got) != 0 {
			t.Errorf("controlKeys(%#v) = %q, want empty", in, got)
		}
	}
}

func TestControlKeysDoesNotMutateInput(t *testing.T) {
	in := []string{"tab", "q", "esc"}
	before := slices.Clone(in)
	controlKeys(in)
	if !slices.Equal(in, before) {
		t.Errorf("input mutated: %q, want %q", in, before)
	}
}

func TestSchemeEditorBindingsKeepsOnlyControlKeys(t *testing.T) {
	s := Default()
	got := s.EditorBindings(Confirm, FocusNext, Cancel)

	want := []struct {
		action Action
		keys   []string
	}{
		{Confirm, []string{"enter", "spacebar"}},
		{FocusNext, []string{"tab"}},
		{Cancel, []string{"esc"}},
	}
	if len(got) != len(want) {
		t.Fatalf("EditorBindings returned %d bindings (%+v), want %d", len(got), got, len(want))
	}
	for i, w := range want {
		if got[i].Action != w.action {
			t.Errorf("binding %d action = %q, want %q", i, got[i].Action, w.action)
		}
		if !slices.Equal(got[i].Keys, w.keys) {
			t.Errorf("binding %d keys = %q, want %q", i, got[i].Keys, w.keys)
		}
	}
}

func TestSchemeEditorBindingsOmitsSingleCharOnlyActions(t *testing.T) {
	s := Default()
	if got := s.EditorBindings(Inc, Dec); len(got) != 0 {
		t.Fatalf("EditorBindings(Inc, Dec) = %+v, want none (every key is a single character)", got)
	}

	nav := s.EditorBindings(Up, Down, Left, Right)
	wantKeys := [][]string{{"up"}, {"down"}, {"left"}, {"right"}}
	if len(nav) != len(wantKeys) {
		t.Fatalf("EditorBindings(nav) = %+v, want %d bindings", nav, len(wantKeys))
	}
	for i, w := range wantKeys {
		if !slices.Equal(nav[i].Keys, w) {
			t.Errorf("nav binding %d keys = %q, want %q (vim letters stripped)", i, nav[i].Keys, w)
		}
	}

	mixed := s.EditorBindings(Inc, FocusNext, Dec)
	if len(mixed) != 1 || mixed[0].Action != FocusNext {
		t.Fatalf("EditorBindings(Inc, FocusNext, Dec) = %+v, want only FocusNext", mixed)
	}
}

func TestSchemeEditorBindingsFollowsArgumentOrder(t *testing.T) {
	s := Default()
	got := s.EditorBindings(Cancel, FocusNext, Quit)
	want := []Action{Cancel, FocusNext, Quit}
	for i, w := range want {
		if i >= len(got) {
			t.Fatalf("EditorBindings = %+v, want %v", got, want)
		}
		if got[i].Action != w {
			t.Errorf("position %d = %q, want %q", i, got[i].Action, w)
		}
	}
	if len(got) != len(want) {
		t.Errorf("EditorBindings returned %d bindings, want %d", len(got), len(want))
	}
}

func TestSchemeEditorBindingsNoArgsAndUnknownAction(t *testing.T) {
	s := Default()
	if got := s.EditorBindings(); got == nil || len(got) != 0 {
		t.Errorf("EditorBindings() = %#v, want empty non-nil slice", got)
	}
	if got := s.EditorBindings(Action("nope")); len(got) != 0 {
		t.Errorf("EditorBindings(unknown) = %+v, want empty", got)
	}
}

func TestSchemeEditorBindingsKeepsGlyphAndLabel(t *testing.T) {
	s := Default()
	got := s.EditorBindings(FocusNext)
	if len(got) != 1 {
		t.Fatalf("EditorBindings(FocusNext) = %+v, want one binding", got)
	}
	if got[0].Glyph != "tab/⇧tab" {
		t.Errorf("Glyph = %q, want %q", got[0].Glyph, "tab/⇧tab")
	}
	if got[0].Label != "focus panel" {
		t.Errorf("Label = %q, want %q", got[0].Label, "focus panel")
	}
}

func TestSchemeEditorBindingsDoesNotMutateScheme(t *testing.T) {
	s := Default()
	before := slices.Clone(s.Binding(Confirm).Keys)

	s.EditorBindings(Confirm, Inc)

	if after := s.Binding(Confirm).Keys; !slices.Equal(after, before) {
		t.Errorf("scheme binding mutated: %q, want %q", after, before)
	}
	if after := Default().Binding(Confirm).Keys; len(after) != 3 {
		t.Errorf("Default() Confirm keys = %q, want three keys", after)
	}
}
