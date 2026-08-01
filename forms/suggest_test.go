package forms

import "testing"

func TestSplitTail(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		text       string
		delim      string
		head, tail string
	}{
		{"no delim keeps whole text", "githu", "", "", "githu"},
		{"comma splits trailing entry", "alpha, bet", ",", "alpha,", "bet"},
		{"comma with no space", "alpha,bet", ",", "alpha,", "bet"},
		{"trailing comma starts a fresh entry", "alpha, ", ",", "alpha,", ""},
		{"no delim present yet", "alpha", ",", "", "alpha"},
		{"space splits trailing term", "is:open is:p", " ", "is:open ", "is:p"},
		{"empty text", "", ",", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			head, tail := splitTail(c.text, c.delim)
			if head != c.head || tail != c.tail {
				t.Fatalf("splitTail(%q, %q) = (%q, %q), want (%q, %q)", c.text, c.delim, head, tail, c.head, c.tail)
			}
		})
	}
}

func TestJoinTail(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name              string
		head, pick, delim string
		want              string
	}{
		{"first entry", "", "alpha", ",", "alpha"},
		{"comma gains a space", "alpha,", "beta", ",", "alpha, beta"},
		{"existing space is kept", "alpha, ", "beta", ",", "alpha, beta"},
		{"space delim adds nothing", "is:open ", "is:pr", " ", "is:open is:pr"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := joinTail(c.head, c.pick, c.delim); got != c.want {
				t.Fatalf("joinTail(%q, %q, %q) = %q, want %q", c.head, c.pick, c.delim, got, c.want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()
	vals := []string{"stale-prs", "standup", "Stalled", "review"}

	got := Match(vals, "sta")
	want := []string{"Stalled", "stale-prs", "standup"}
	if len(got) != len(want) {
		t.Fatalf("Match(sta) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Match(sta) = %v, want %v", got, want)
		}
	}

	if got := Match(vals, "standup"); len(got) != 0 {
		t.Fatalf("an exact match has nothing left to complete, got %v", got)
	}
	if got := Match(vals, "zzz"); len(got) != 0 {
		t.Fatalf("Match(zzz) = %v, want none", got)
	}
	if got := Match(vals, ""); len(got) != 4 {
		t.Fatalf("Match(\"\") = %v, want all four", got)
	}
}

func TestGhostOf(t *testing.T) {
	t.Parallel()
	if got := ghostOf("stale-prs", "sta"); got != "le-prs" {
		t.Fatalf("ghostOf = %q, want %q", got, "le-prs")
	}
	if got := ghostOf("Stalled", "sta"); got != "lled" {
		t.Fatalf("ghostOf is case-insensitive on the typed part, got %q", got)
	}
	if got := ghostOf("stale-prs", "review"); got != "" {
		t.Fatalf("a non-prefix has no ghost, got %q", got)
	}
}
