package forms

import (
	"slices"
	"strings"
	"testing"
)

func TestStr(t *testing.T) {
	t.Parallel()
	vals := map[string]any{
		"name":   "  munin  ",
		"blank":  "   ",
		"plain":  "plain",
		"number": 42,
		"nil":    nil,
		"slice":  []string{"a"},
		"tabs":   "\t\nspaced\n\t",
	}
	cases := []struct {
		key  string
		want string
	}{
		{"name", "munin"},
		{"plain", "plain"},
		{"blank", ""},
		{"tabs", "spaced"},
		{"number", ""},
		{"nil", ""},
		{"slice", ""},
		{"missing", ""},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			if got := Str(vals, c.key); got != c.want {
				t.Errorf("Str(%q) = %q, want %q", c.key, got, c.want)
			}
		})
	}
	if got := Str(nil, "name"); got != "" {
		t.Errorf("Str(nil map) = %q, want empty", got)
	}
}

func TestBool(t *testing.T) {
	t.Parallel()
	vals := map[string]any{
		"on":     true,
		"off":    false,
		"text":   "true",
		"number": 1,
		"nil":    nil,
	}
	cases := []struct {
		key  string
		want bool
	}{
		{"on", true},
		{"off", false},
		{"text", false},
		{"number", false},
		{"nil", false},
		{"missing", false},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			if got := Bool(vals, c.key); got != c.want {
				t.Errorf("Bool(%q) = %v, want %v", c.key, got, c.want)
			}
		})
	}
	if Bool(nil, "on") {
		t.Error("Bool(nil map) = true, want false")
	}
}

func TestInt(t *testing.T) {
	t.Parallel()
	vals := map[string]any{
		"n":        "42",
		"padded":   "  7  ",
		"negative": "-13",
		"zero":     "0",
		"plus":     "+5",
		"float":    "1.5",
		"words":    "twelve",
		"blank":    "",
		"spaces":   "   ",
		"native":   42,
		"boolish":  true,
		"big":      "2147483647",
	}
	cases := []struct {
		key  string
		want int
	}{
		{"n", 42},
		{"padded", 7},
		{"negative", -13},
		{"zero", 0},
		{"plus", 5},
		{"float", 0},
		{"words", 0},
		{"blank", 0},
		{"spaces", 0},
		{"native", 42},
		{"boolish", 0},
		{"big", 2147483647},
		{"missing", 0},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			if got := Int(vals, c.key); got != c.want {
				t.Errorf("Int(%q) = %d, want %d", c.key, got, c.want)
			}
		})
	}
	if got := Int(nil, "n"); got != 0 {
		t.Errorf("Int(nil map) = %d, want 0", got)
	}
}

func TestStrings(t *testing.T) {
	t.Parallel()
	want := []string{"a", "b", "c"}
	vals := map[string]any{
		"tags":   want,
		"empty":  []string{},
		"single": "a",
		"anys":   []any{"a", "b"},
		"nil":    nil,
	}

	got := Strings(vals, "tags")
	if !slices.Equal(got, want) {
		t.Errorf("Strings(tags) = %q, want %q", got, want)
	}
	if got := Strings(vals, "empty"); got == nil || len(got) != 0 {
		t.Errorf("Strings(empty) = %#v, want empty non-nil slice", got)
	}
	for _, key := range []string{"single", "anys", "nil", "missing"} {
		if got := Strings(vals, key); got != nil {
			t.Errorf("Strings(%q) = %#v, want nil", key, got)
		}
	}
	if got := Strings(nil, "tags"); got != nil {
		t.Errorf("Strings(nil map) = %#v, want nil", got)
	}
}

func TestSelectIndex(t *testing.T) {
	t.Parallel()
	options := []string{"alpha", "beta", "gamma"}
	cases := []struct {
		name    string
		options []string
		want    string
		idx     int
	}{
		{"first", options, "alpha", 0},
		{"middle", options, "beta", 1},
		{"last", options, "gamma", 2},
		{"empty want", options, "", 0},
		{"unknown want", options, "delta", 0},
		{"case sensitive miss", options, "Beta", 0},
		{"no options", nil, "alpha", 0},
		{"no options empty want", []string{}, "", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := SelectIndex(c.options, c.want); got != c.idx {
				t.Errorf("SelectIndex(%q, %q) = %d, want %d", c.options, c.want, got, c.idx)
			}
		})
	}
}

func TestSelectIndexPicksFirstOfDuplicates(t *testing.T) {
	t.Parallel()
	if got := SelectIndex([]string{"a", "b", "a"}, "a"); got != 0 {
		t.Errorf("SelectIndex duplicates = %d, want 0", got)
	}
}

func TestSelectFirstMovesCurrentToFront(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		options []string
		cur     string
		want    []string
	}{
		{"middle", []string{"a", "b", "c"}, "b", []string{"b", "a", "c"}},
		{"last", []string{"a", "b", "c"}, "c", []string{"c", "a", "b"}},
		{"already first", []string{"a", "b", "c"}, "a", []string{"a", "b", "c"}},
		{"single option", []string{"a"}, "a", []string{"a"}},
		{"empty option present", []string{"a", "", "c"}, "", []string{"", "a", "c"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SelectFirst(c.options, c.cur)
			if !slices.Equal(got, c.want) {
				t.Errorf("SelectFirst(%q, %q) = %q, want %q", c.options, c.cur, got, c.want)
			}
		})
	}
}

func TestSelectFirstPreservesRelativeOrderOfTheRest(t *testing.T) {
	t.Parallel()
	options := []string{"a", "b", "c", "d", "e"}
	got := SelectFirst(options, "d")
	want := []string{"d", "a", "b", "c", "e"}
	if !slices.Equal(got, want) {
		t.Fatalf("SelectFirst = %q, want %q", got, want)
	}
	if rest := strings.Join(got[1:], ""); rest != "abce" {
		t.Errorf("tail order = %q, want abce", rest)
	}
}

func TestSelectFirstReturnsOptionsUnchangedWhenCurrentIsAbsent(t *testing.T) {
	t.Parallel()
	options := []string{"a", "b", "c"}
	cases := []struct {
		name string
		cur  string
	}{
		{"absent", "zz"},
		{"empty cur", ""},
		{"case mismatch", "A"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SelectFirst(options, c.cur)
			if !slices.Equal(got, options) {
				t.Fatalf("SelectFirst(%q, %q) = %q, want unchanged", options, c.cur, got)
			}
			if len(got) > 0 && &got[0] == &options[0] {
				t.Errorf("SelectFirst(%q, %q) aliased the input, want a fresh slice", options, c.cur)
			}
		})
	}

	if got := SelectFirst(nil, "a"); len(got) != 0 {
		t.Errorf("SelectFirst(nil, a) = %#v, want empty", got)
	}
	if got := SelectFirst([]string{}, "a"); len(got) != 0 {
		t.Errorf("SelectFirst(empty, a) = %#v, want empty", got)
	}
}

func TestSelectFirstDoesNotMutateInput(t *testing.T) {
	t.Parallel()
	options := []string{"a", "b", "c"}
	before := slices.Clone(options)

	got := SelectFirst(options, "c")

	if !slices.Equal(options, before) {
		t.Fatalf("input mutated: %q, want %q", options, before)
	}
	got[0] = "MUTATED"
	if options[0] != "a" {
		t.Errorf("result aliases input: options = %q", options)
	}
	if !slices.Equal(options, before) {
		t.Errorf("writing to the result changed the input: %q, want %q", options, before)
	}
}

func TestSelectFirstKeepsDuplicateCurrentValues(t *testing.T) {
	t.Parallel()
	options := []string{"a", "b", "a", "c"}
	got := SelectFirst(options, "a")
	want := []string{"a", "b", "a", "c"}
	if !slices.Equal(got, want) {
		t.Errorf("SelectFirst(%q, %q) = %q, want %q (only the first match moves)", options, "a", got, want)
	}
}
