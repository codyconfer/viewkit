package forms

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
)

func typeInto(fm *Form, s string) {
	for _, r := range s {
		fm.Insert(string(r))
	}
}

func TestAcceptSuggestionCompletesTypedToken(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{
		Key:     "filters",
		Kind:    FieldText,
		Suggest: Static("stale-prs", "standup", "review"),
		Delim:   ",",
	})

	typeInto(fm, "sta")
	if got := fm.Suggestions(); len(got) != 2 {
		t.Fatalf("Suggestions() = %v, want two", got)
	}
	if !fm.AcceptSuggestion() {
		t.Fatal("AcceptSuggestion reported nothing to accept")
	}
	if got := fm.Fields[0].Text; got != "stale-prs" {
		t.Fatalf("Text = %q, want %q", got, "stale-prs")
	}
}

func TestAcceptSuggestionCompletesSecondEntry(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{
		Key:     "queries",
		Kind:    FieldText,
		Suggest: Static("alpha", "beta"),
		Delim:   ",",
	})

	typeInto(fm, "alpha,be")
	if !fm.AcceptSuggestion() {
		t.Fatal("AcceptSuggestion reported nothing to accept")
	}
	if got := fm.Fields[0].Text; got != "alpha, beta" {
		t.Fatalf("Text = %q, want %q", got, "alpha, beta")
	}
}

func TestAcceptSuggestionChainsSpaceDelimitedTerms(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{
		Key:     "query",
		Kind:    FieldText,
		Suggest: Static("is:open", "is:pr"),
		Delim:   " ",
	})

	typeInto(fm, "is:o")
	if !fm.AcceptSuggestion() {
		t.Fatal("first accept failed")
	}
	if got := fm.Fields[0].Text; got != "is:open" {
		t.Fatalf("Text = %q, want %q", got, "is:open")
	}
	typeInto(fm, " is:p")
	if !fm.AcceptSuggestion() {
		t.Fatal("second accept failed")
	}
	if got := fm.Fields[0].Text; got != "is:open is:pr" {
		t.Fatalf("Text = %q, want %q", got, "is:open is:pr")
	}
}

func TestAcceptSuggestionFalseWithoutCandidates(t *testing.T) {
	t.Parallel()
	fm := NewForm(
		Field{Key: "include", Kind: FieldText},
		Field{Key: "name", Kind: FieldText, Suggest: Static("alpha")},
	)

	typeInto(fm, "^wip")
	if got := fm.Suggestions(); len(got) != 0 {
		t.Fatalf("a field with no Suggester offers nothing, got %v", got)
	}
	if fm.AcceptSuggestion() {
		t.Fatal("AcceptSuggestion must report false so the host can reuse the key")
	}
	if got := fm.Fields[0].Text; got != "^wip" {
		t.Fatalf("Text = %q, want it untouched", got)
	}
}

func TestCycleSuggestionWraps(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{
		Key:     "filters",
		Kind:    FieldText,
		Suggest: Static("alpha", "amber", "ash"),
	})

	typeInto(fm, "a")
	fm.CycleSuggestion(+1)
	if !fm.AcceptSuggestion() {
		t.Fatal("accept failed")
	}
	if got := fm.Fields[0].Text; got != "amber" {
		t.Fatalf("Text = %q, want the second candidate", got)
	}

	fm2 := NewForm(Field{Key: "f", Kind: FieldText, Suggest: Static("alpha", "amber", "ash")})
	typeInto(fm2, "a")
	fm2.CycleSuggestion(-1)
	if !fm2.AcceptSuggestion() {
		t.Fatal("accept failed")
	}
	if got := fm2.Fields[0].Text; got != "ash" {
		t.Fatalf("Text = %q, want the list to wrap to the last candidate", got)
	}
}

func TestSuggestionsResetOnEraseAndFieldMove(t *testing.T) {
	t.Parallel()
	fm := NewForm(
		Field{Key: "filters", Kind: FieldText, Suggest: Static("stale-prs")},
		Field{Key: "name", Kind: FieldText},
	)

	typeInto(fm, "sta")
	if len(fm.Suggestions()) != 1 {
		t.Fatal("expected a candidate after typing")
	}
	fm.Handle(keys.Down)
	if got := fm.Suggestions(); len(got) != 0 {
		t.Fatalf("moving to a field with no Suggester must clear candidates, got %v", got)
	}
	fm.Handle(keys.Up)
	if len(fm.Suggestions()) != 1 {
		t.Fatal("returning to the field must recompute candidates")
	}
	for range 3 {
		fm.Handle(keys.Erase)
	}
	if got := fm.Suggestions(); len(got) != 1 {
		t.Fatalf("an empty prefix still matches, got %v", got)
	}
}

func TestHandleCyclesThroughKeyActions(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{Key: "f", Kind: FieldText, Suggest: Static("alpha", "amber")})
	typeInto(fm, "a")

	if !fm.Handle(keys.CompleteNext) {
		t.Fatal("CompleteNext must be handled")
	}
	if !fm.AcceptSuggestion() {
		t.Fatal("accept failed")
	}
	if got := fm.Fields[0].Text; got != "amber" {
		t.Fatalf("Text = %q, want %q", got, "amber")
	}
}

func TestSecretFieldNeverSuggests(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{Key: "token", Kind: FieldText, Secret: true, Suggest: Static("hunter2")})
	typeInto(fm, "hun")
	if got := fm.Suggestions(); len(got) != 0 {
		t.Fatalf("a secret field must not leak its vocabulary, got %v", got)
	}
}

func TestRenderShowsGhostAndCandidates(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{
		Key:     "filters",
		Label:   "filters",
		Kind:    FieldText,
		Suggest: Static("stale-prs", "standup"),
	})
	typeInto(fm, "sta")

	body := fm.Render(layout.NewFrame(80), "build query")
	if !strings.Contains(body, "le-prs") {
		t.Fatalf("expected the ghost remainder in:\n%s", body)
	}
	if !strings.Contains(body, "standup") {
		t.Fatalf("expected the alternative candidate listed in:\n%s", body)
	}
}
