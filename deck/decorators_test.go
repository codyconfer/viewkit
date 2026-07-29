package deck

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type decoView struct {
	title   string
	hints   [][2]string
	ctx     [][2]string
	updated int
}

func (d *decoView) Title() string        { return d.title }
func (d *decoView) Init() tea.Cmd        { return nil }
func (d *decoView) Body(int, int) string { return "inner-body" }
func (d *decoView) Hints() [][2]string   { return d.hints }
func (d *decoView) Context() [][2]string { return d.ctx }

func (d *decoView) Update(*Model, tea.Msg) tea.Cmd {
	d.updated++
	return nil
}

func TestWithExtraHintsAppendsAfterInnerHints(t *testing.T) {
	inner := &decoView{hints: [][2]string{{"↑/↓", "move"}}}
	got := WithExtraHints(inner, [][2]string{{"g", "go"}, {"r", "run"}}).Hints()
	want := [][2]string{{"↑/↓", "move"}, {"g", "go"}, {"r", "run"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hints = %v, want %v", got, want)
	}
}

func TestWithExtraHintsEmptyReturnsInnerUnchanged(t *testing.T) {
	inner := &decoView{}
	if got := WithExtraHints(inner, nil); got != View(inner) {
		t.Errorf("nil extras should return the same view, got %#v", got)
	}
	if got := WithExtraHints(inner, [][2]string{}); got != View(inner) {
		t.Errorf("empty extras should return the same view, got %#v", got)
	}
}

func TestWithExtraHintsLeavesInnerSliceAlone(t *testing.T) {
	// A slice with spare capacity is the case a naive append corrupts.
	base := make([][2]string, 1, 4)
	base[0] = [2]string{"↑/↓", "move"}
	inner := &decoView{hints: base}
	wrapped := WithExtraHints(inner, [][2]string{{"g", "go"}})

	first := wrapped.Hints()
	second := wrapped.Hints()
	if !reflect.DeepEqual(first, second) {
		t.Errorf("repeated Hints differ: %v then %v", first, second)
	}
	if len(inner.hints) != 1 || inner.hints[0] != [2]string{"↑/↓", "move"} {
		t.Errorf("inner hints mutated: %v", inner.hints)
	}
	if cap(base) > 1 && base[:cap(base)][1] != [2]string{} {
		t.Errorf("extras were written into the inner backing array: %v", base[:cap(base)])
	}
}

func TestWithExtraHintsPassesEverythingElseThrough(t *testing.T) {
	inner := &decoView{title: "roles", ctx: [][2]string{{"role", "sre"}}}
	wrapped := WithExtraHints(inner, [][2]string{{"g", "go"}})

	if wrapped.Title() != "roles" {
		t.Errorf("title = %q, want roles", wrapped.Title())
	}
	if wrapped.Body(20, 5) != "inner-body" {
		t.Errorf("body = %q", wrapped.Body(20, 5))
	}
	if !reflect.DeepEqual(wrapped.Context(), inner.ctx) {
		t.Errorf("context = %v, want %v", wrapped.Context(), inner.ctx)
	}
	wrapped.Update(nil, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if inner.updated != 1 {
		t.Errorf("update reached inner %d times, want 1", inner.updated)
	}
}

func TestWithLiveContextReEvaluatesEveryRead(t *testing.T) {
	role := "sre"
	inner := &decoView{ctx: [][2]string{{"role", "stale"}}}
	wrapped := WithLiveContext(inner, func() [][2]string { return [][2]string{{"role", role}} })

	if got := wrapped.Context(); !reflect.DeepEqual(got, [][2]string{{"role", "sre"}}) {
		t.Fatalf("context = %v, want role sre", got)
	}
	role = "oncall"
	if got := wrapped.Context(); !reflect.DeepEqual(got, [][2]string{{"role", "oncall"}}) {
		t.Fatalf("context = %v, want role oncall after the value changed", got)
	}
}

func TestWithLiveContextNilReturnsInnerUnchanged(t *testing.T) {
	inner := &decoView{}
	if got := WithLiveContext(inner, nil); got != View(inner) {
		t.Errorf("nil fn should return the same view, got %#v", got)
	}
}

func TestWithLiveContextPassesEverythingElseThrough(t *testing.T) {
	inner := &decoView{title: "home", hints: [][2]string{{"g", "go"}}}
	wrapped := WithLiveContext(inner, func() [][2]string { return nil })

	if wrapped.Title() != "home" {
		t.Errorf("title = %q, want home", wrapped.Title())
	}
	if wrapped.Body(20, 5) != "inner-body" {
		t.Errorf("body = %q", wrapped.Body(20, 5))
	}
	if !reflect.DeepEqual(wrapped.Hints(), inner.hints) {
		t.Errorf("hints = %v, want %v", wrapped.Hints(), inner.hints)
	}
	wrapped.Update(nil, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if inner.updated != 1 {
		t.Errorf("update reached inner %d times, want 1", inner.updated)
	}
}

func TestDecoratorsStackOnEachOther(t *testing.T) {
	inner := &decoView{hints: [][2]string{{"↑/↓", "move"}}}
	wrapped := WithExtraHints(WithLiveContext(inner, func() [][2]string {
		return [][2]string{{"role", "sre"}}
	}), [][2]string{{"g", "go"}})

	if got := wrapped.Hints(); !reflect.DeepEqual(got, [][2]string{{"↑/↓", "move"}, {"g", "go"}}) {
		t.Errorf("hints = %v", got)
	}
	if got := wrapped.Context(); !reflect.DeepEqual(got, [][2]string{{"role", "sre"}}) {
		t.Errorf("context = %v", got)
	}
}
