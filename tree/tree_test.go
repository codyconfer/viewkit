package tree

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/glyph"
)

func withMode(t *testing.T, m glyph.Mode) {
	t.Helper()
	prev := glyph.CurrentMode()
	glyph.SetMode(m)
	t.Cleanup(func() { glyph.SetMode(prev) })
}

func plain(s string) string { return ansi.Strip(s) }

func plainLines(r Row) []string {
	out := make([]string, len(r.Lines))
	for i, l := range r.Lines {
		out[i] = plain(l)
	}
	return out
}

func TestDefaultConnectorsFollowsGlyphMode(t *testing.T) {
	withMode(t, glyph.ModeNone)
	if got := DefaultConnectors(); got != asciiConnectors {
		t.Fatalf("ModeNone connectors = %+v, want ASCII set", got)
	}
	for _, m := range []glyph.Mode{glyph.ModeNerd, glyph.ModeUnicode} {
		glyph.SetMode(m)
		if got := DefaultConnectors(); got != unicodeConnectors {
			t.Fatalf("mode %v connectors = %+v, want Unicode set", m, got)
		}
	}
}

func TestConnectorWidthsMatch(t *testing.T) {
	for name, c := range map[string]Connectors{"ascii": asciiConnectors, "unicode": unicodeConnectors} {
		width := len([]rune(c.Mid))
		for field, s := range map[string]string{"End": c.End, "Vert": c.Vert, "Space": c.Space} {
			if len([]rune(s)) != width {
				t.Errorf("%s %s width = %d, want %d", name, field, len([]rune(s)), width)
			}
		}
	}
}

func TestEdgePicksMidThenEnd(t *testing.T) {
	c := unicodeConnectors
	if conn, stem := c.Edge(false); conn != "├─ " || stem != "│  " {
		t.Fatalf("Edge(false) = %q, %q; want mid and vert", conn, stem)
	}
	if conn, stem := c.Edge(true); conn != "└─ " || stem != "   " {
		t.Fatalf("Edge(true) = %q, %q; want end and space", conn, stem)
	}
}

func TestLeafConnectorChoice(t *testing.T) {
	withMode(t, glyph.ModeUnicode)
	c := DefaultConnectors()

	mid := Leaf(c, "", false, []string{"a"}, "")
	if got := plainLines(mid)[0]; got != "├─ a" {
		t.Fatalf("non-last leaf = %q, want %q", got, "├─ a")
	}

	end := Leaf(c, "", true, []string{"a"}, "")
	if got := plainLines(end)[0]; got != "└─ a" {
		t.Fatalf("last leaf = %q, want %q", got, "└─ a")
	}
}

func TestLeafMultiLineBodyGetsContinuationStem(t *testing.T) {
	withMode(t, glyph.ModeUnicode)
	c := DefaultConnectors()

	mid := plainLines(Leaf(c, "│  ", false, []string{"one", "two", "three"}, ""))
	want := []string{"│  ├─ one", "│  │  two", "│  │  three"}
	for i := range want {
		if mid[i] != want[i] {
			t.Fatalf("mid line %d = %q, want %q", i, mid[i], want[i])
		}
	}

	last := plainLines(Leaf(c, "│  ", true, []string{"one", "two"}, ""))
	wantLast := []string{"│  └─ one", "│     two"}
	for i := range wantLast {
		if last[i] != wantLast[i] {
			t.Fatalf("last line %d = %q, want %q", i, last[i], wantLast[i])
		}
	}
}

func TestLeafEmptyBodyYieldsNoLines(t *testing.T) {
	withMode(t, glyph.ModeUnicode)
	r := Leaf(DefaultConnectors(), "", false, nil, "")
	if len(r.Lines) != 0 {
		t.Fatalf("Lines = %q, want none", r.Lines)
	}
}

func TestLeafKeyDrivesSelectable(t *testing.T) {
	withMode(t, glyph.ModeUnicode)
	c := DefaultConnectors()

	if r := Leaf(c, "", false, []string{"a"}, "u1"); !r.Selectable || r.Key != "u1" {
		t.Fatalf("keyed leaf = %+v, want selectable with Key u1", r)
	}
	if r := Leaf(c, "", false, []string{"a"}, ""); r.Selectable {
		t.Fatal("keyless leaf should not be selectable")
	}
}

func TestBranchIsALeafWithNoStem(t *testing.T) {
	withMode(t, glyph.ModeNone)
	c := DefaultConnectors()

	branch := Branch(c, false, []string{"github  (2)"}, "")
	leaf := Leaf(c, "", false, []string{"github  (2)"}, "")
	if plainLines(branch)[0] != plainLines(leaf)[0] {
		t.Fatalf("Branch = %q, Leaf with empty stem = %q", branch.Lines[0], leaf.Lines[0])
	}
	if got := plain(branch.GapStem); got != "|  " {
		t.Fatalf("branch GapStem = %q, want %q", got, "|  ")
	}
	if got := plain(Branch(c, true, []string{"cal"}, "").GapStem); got != "   " {
		t.Fatalf("last branch GapStem = %q, want three spaces", got)
	}
}

func TestGapStemContinuesConnectors(t *testing.T) {
	withMode(t, glyph.ModeNone)
	c := DefaultConnectors()

	_, stem := c.Edge(false)
	if got := plain(Leaf(c, stem, false, []string{"a"}, "u1").GapStem); got != "|  |  " {
		t.Fatalf("mid leaf GapStem = %q, want %q", got, "|  |  ")
	}
	if got := plain(Leaf(c, stem, true, []string{"b"}, "u2").GapStem); got != "|     " {
		t.Fatalf("last leaf GapStem = %q, want %q", got, "|     ")
	}

	_, lastStem := c.Edge(true)
	if got := plain(Leaf(c, lastStem, false, []string{"c"}, "u3").GapStem); got != "   |  " {
		t.Fatalf("leaf under a last branch GapStem = %q, want %q", got, "   |  ")
	}
}

func TestGapStemMatchesNextLineOfTheSameRow(t *testing.T) {
	withMode(t, glyph.ModeUnicode)
	c := DefaultConnectors()

	r := Leaf(c, "│  ", false, []string{"one", "two"}, "")
	gap := plain(r.GapStem)
	if cont := plainLines(r)[1]; !strings.HasPrefix(cont, gap) {
		t.Fatalf("continuation line %q does not start with GapStem %q", cont, gap)
	}
}

func TestNestedIndentationStacksStems(t *testing.T) {
	withMode(t, glyph.ModeUnicode)
	c := DefaultConnectors()

	branch := Branch(c, false, []string{"root child"}, "")
	_, stem := c.Edge(false)
	child := Leaf(c, stem, false, []string{"child"}, "")
	_, childStem := c.Edge(false)
	grandchild := Leaf(c, stem+childStem, true, []string{"grandchild"}, "")

	got := []string{plainLines(branch)[0], plainLines(child)[0], plainLines(grandchild)[0]}
	want := []string{"├─ root child", "│  ├─ child", "│  │  └─ grandchild"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("depth %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestASCIIFallbackWhenGlyphsOff(t *testing.T) {
	withMode(t, glyph.ModeNone)
	c := DefaultConnectors()

	rows := []Row{
		Branch(c, false, []string{"github"}, ""),
		Leaf(c, "|  ", false, []string{"a", "wrapped"}, "u1"),
		Leaf(c, "|  ", true, []string{"b"}, "u2"),
		Branch(c, true, []string{"cal"}, ""),
		Leaf(c, "   ", true, []string{c.Empty + "nothing to show"}, ""),
	}

	var lines []string
	for _, r := range rows {
		lines = append(lines, plainLines(r)...)
	}
	want := []string{
		"+- github",
		"|  +- a",
		"|  |  wrapped",
		"|  `- b",
		"`- cal",
		"   `- o nothing to show",
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, lines[i], want[i])
		}
	}
	for _, l := range lines {
		if strings.ContainsAny(l, "├└│●∅") {
			t.Fatalf("line %q leaked a Unicode connector in ASCII mode", l)
		}
	}
}

func TestIndentCountsStemPlusConnector(t *testing.T) {
	withMode(t, glyph.ModeUnicode)
	c := DefaultConnectors()

	if got := c.Indent(""); got != 3 {
		t.Fatalf("Indent(\"\") = %d, want 3", got)
	}
	if got := c.Indent("│  "); got != 6 {
		t.Fatalf("Indent(vert) = %d, want 6", got)
	}
	if got := c.Indent("│  │  "); got != 9 {
		t.Fatalf("Indent(vert vert) = %d, want 9", got)
	}

	body := "x"
	line := plain(Leaf(c, "│  ", false, []string{body}, "").Lines[0])
	if got := len([]rune(line)) - len([]rune(body)); got != c.Indent("│  ") {
		t.Fatalf("rendered prefix width = %d, want Indent %d", got, c.Indent("│  "))
	}
}
