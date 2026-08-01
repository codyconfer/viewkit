package forms

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

const (
	kelvinSign  = "\u212a"
	dottedCapI  = "\u0130"
	angstromCap = "\u00c5"
	oDiaeresis  = "\u00d6"
)

func TestGhostOfCaseFoldingChangesByteLength(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		pick, tail string
		want       string
	}{
		{"kelvin sign shrinks to one byte", "kx", kelvinSign, "x"},
		{"kelvin sign with nothing left", "k", kelvinSign, ""},
		{"dotted capital I pair", "iix", dottedCapI + dottedCapI, "x"},
		{"dotted capital I keeps the delimiter", "istanbul-1", dottedCapI + "stanbul", "-1"},
		{"tail longer than candidate", "ab", "abcd", ""},
		{"multibyte tail stays aligned", angstromCap + "NGSTR" + oDiaeresis + "M-unit", angstromCap + "ngstr" + oDiaeresis + "m", "-unit"},
		{"non prefix", "stale-prs", "review", ""},
		{"empty tail keeps the whole candidate", "stale-prs", "", "stale-prs"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ghostOf(c.pick, c.tail); got != c.want {
				t.Fatalf("ghostOf(%q, %q) = %q, want %q", c.pick, c.tail, got, c.want)
			}
		})
	}
}

func TestRenderFoldingKeystrokeDoesNotPanic(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{Key: "k", Suggest: Static("kx")})
	fm.Insert(kelvinSign)
	if got := fm.Values()["k"]; got != kelvinSign {
		t.Fatalf("inserted text = %q, want the kelvin sign", got)
	}
	out := stripANSI(fm.Render(layout.DefaultFrame(), "form"))
	if !strings.Contains(out, kelvinSign+"▎"+"x") {
		t.Fatalf("want typed rune, cursor, then the ghost remainder:\n%s", out)
	}
	if !strings.Contains(out, "kx") {
		t.Fatalf("candidate missing from render:\n%s", out)
	}
}

func TestRenderFoldingKeystrokeAcrossDelimiter(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{Key: "k", Delim: ",", Suggest: Static("kelvin")})
	fm.Insert("a," + kelvinSign)
	out := stripANSI(fm.Render(layout.DefaultFrame(), "form"))
	if !strings.Contains(out, "▎"+"elvin") {
		t.Fatalf("ghost should complete the folded token:\n%s", out)
	}
}
