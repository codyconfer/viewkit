package layout

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

func cellBoxPane(name string) Pane {
	return Pane{
		Name:   name,
		Render: func(f Frame) string { return f.CellBox(name, "x") },
	}
}

// titledBoxPane is the documented pane shape for TitledBox: the pane converts
// its rect width to a body width. TitledBox routes through Frame.BodyWidth, so
// it can only be exact while the body stays >= theme.MinBodyWidth, i.e. tracks
// of at least MinBodyWidth+4 columns.
func titledBoxPane(name string) Pane {
	return Pane{
		Name:   name,
		Render: func(f Frame) string { return NewFrame(f.Width-4).TitledBox(name, "x") },
	}
}

// cellTitledBoxPane is the pane shape for tracks of any width: CellTitledBox
// takes the rect width directly and clamps its body down to fit.
func cellTitledBoxPane(name string) Pane {
	return Pane{
		Name:   name,
		Render: func(f Frame) string { return f.CellTitledBox(name, "x") },
	}
}

// assertBoxedGridIntact checks the two invariants a tiled layout owes its
// caller: every row is exactly the frame width, and no box loses an edge — a
// sliced right border shows up as a row with more ╭ than ╮.
func assertBoxedGridIntact(t *testing.T, label, out string, width, wantBoxes int) {
	t.Helper()
	rows := strings.Split(out, "\n")
	for i, r := range rows {
		s := stripANSI(r)
		if w := ansi.StringWidth(r); w != width {
			t.Fatalf("%s: row %d width = %d, want %d:\n%s", label, i, w, width, stripANSI(out))
		}
		for _, pair := range [][2]string{{"╭", "╮"}, {"╰", "╯"}} {
			open, close := strings.Count(s, pair[0]), strings.Count(s, pair[1])
			if open != close {
				t.Fatalf("%s: row %d has %d %s but %d %s (border sliced off):\n%s",
					label, i, open, pair[0], close, pair[1], stripANSI(out))
			}
		}
		if n := strings.Count(s, "│"); n%2 != 0 {
			t.Fatalf("%s: row %d has %d │ (odd — an edge was cut):\n%s", label, i, n, stripANSI(out))
		}
	}
	if got := strings.Count(stripANSI(rows[0]), "╭"); got != wantBoxes {
		t.Fatalf("%s: top row has %d boxes, want %d:\n%s", label, got, wantBoxes, stripANSI(out))
	}
	closed := 0
	for _, r := range rows {
		closed += strings.Count(stripANSI(r), "╯")
	}
	if closed == 0 {
		t.Fatalf("%s: no box ever closed:\n%s", label, stripANSI(out))
	}
}

func TestGridNarrowColumnsKeepBordersIntact(t *testing.T) {
	cases := []struct {
		name      string
		cols      int
		width     int
		height    int
		pane      func(string) Pane
		names     []string
		wantBoxes int
	}{
		{"4 cols at 81 with CellBox", 4, 81, 4, cellBoxPane, []string{"a", "b", "c", "d"}, 4},
		{"3 cols at 60 with CellBox", 3, 60, 4, cellBoxPane, []string{"a", "b", "c"}, 3},
		{"2 cols at 60 with CellBox", 2, 60, 4, cellBoxPane, []string{"a", "b"}, 2},
		{"2 cols at 60 with TitledBox", 2, 60, 4, titledBoxPane, []string{"a", "b"}, 2},
		{"2 cols at 81 with TitledBox", 2, 81, 4, titledBoxPane, []string{"a", "b"}, 2},
		{"8 cols at 81 with CellBox", 8, 81, 8, cellBoxPane, []string{"a", "b", "c", "d", "e", "f", "g", "h"}, 4},
		{"4 cols at 40 with CellBox", 4, 40, 8, cellBoxPane, []string{"a", "b", "c", "d"}, 2},
		{"2 cols at 24 with CellBox", 2, 24, 8, cellBoxPane, []string{"a", "b"}, 1},
		{"1 col at 27 with CellTitledBox", 1, 27, 4, cellTitledBoxPane, []string{"a"}, 1},
		{"2 cols at 43 with CellTitledBox", 2, 43, 4, cellTitledBoxPane, []string{"a", "b"}, 2},
		{"3 cols at 67 with CellTitledBox", 3, 67, 4, cellTitledBoxPane, []string{"a", "b", "c"}, 3},
		{"4 cols at 95 with CellTitledBox", 4, 95, 4, cellTitledBoxPane, []string{"a", "b", "c", "d"}, 4},
		{"8 cols at 184 with CellTitledBox", 8, 184, 4, cellTitledBoxPane,
			[]string{"a", "b", "c", "d", "e", "f", "g", "h"}, 8},
	}
	for _, c := range cases {
		panes := make([]Pane, 0, len(c.names))
		for _, n := range c.names {
			panes = append(panes, c.pane(n))
		}
		out := Grid{Cols: c.cols}.Arrange(Frame{Width: c.width, Height: c.height}, TierTall, panes, "")
		assertBoxedGridIntact(t, c.name, out, c.width, c.wantBoxes)
	}
}

func TestFitColsDropsColumnsBelowMinTrack(t *testing.T) {
	cases := []struct{ cols, width, want int }{
		{4, 81, 4}, {4, 79, 3}, {4, 40, 2}, {4, 20, 1}, {4, 19, 1},
		{2, 40, 2}, {2, 39, 1}, {8, 81, 4}, {1, 1, 1}, {0, 81, 1}, {-3, 81, 1},
		{4, 0, 1}, {4, -5, 1},
	}
	for _, c := range cases {
		if got := FitCols(c.cols, c.width); got != c.want {
			t.Errorf("FitCols(%d, %d) = %d, want %d", c.cols, c.width, got, c.want)
		}
	}
}

func TestGridNeverBuildsColumnNarrowerThanMinTrack(t *testing.T) {
	for _, width := range []int{28, 40, 55, 60, 81, 83, 111, 200} {
		for _, cols := range []int{1, 2, 3, 4, 6, 8} {
			names := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
			panes := make([]Pane, 0, cols)
			for i := 0; i < cols; i++ {
				panes = append(panes, cellBoxPane(names[i]))
			}
			out := Grid{Cols: cols}.Arrange(Frame{Width: width, Height: 6}, TierTall, panes, "")
			top := stripANSI(strings.Split(out, "\n")[0])
			if opens, closes := strings.Count(top, "╭"), strings.Count(top, "╮"); opens != closes {
				t.Fatalf("width %d cols %d: %d ╭ but %d ╮ on top row:\n%s", width, cols, opens, closes, stripANSI(out))
			}
			for _, r := range strings.Split(out, "\n") {
				if w := ansi.StringWidth(r); w != width {
					t.Fatalf("width %d cols %d: row width %d:\n%s", width, cols, w, stripANSI(out))
				}
			}
		}
	}
}

func TestFlexRowsNarrowTracksKeepBordersIntact(t *testing.T) {
	panes := []Pane{cellBoxPane("a"), cellBoxPane("b"), cellBoxPane("c"), cellBoxPane("d")}
	for _, width := range []int{28, 40, 60, 81, 120} {
		out := FlexRows{FlexBounds: FlexBounds{MinWidth: 20, MaxCols: 4}}.Arrange(Frame{Width: width}, TierTall, panes, "")
		for _, r := range strings.Split(out, "\n") {
			if r == "" {
				continue
			}
			if w := ansi.StringWidth(r); w != width {
				t.Fatalf("FlexRows width %d: row width %d:\n%s", width, w, stripANSI(out))
			}
		}
		for _, r := range strings.Split(out, "\n") {
			s := stripANSI(r)
			if opens, closes := strings.Count(s, "╭"), strings.Count(s, "╮"); opens != closes {
				t.Fatalf("FlexRows width %d: %d ╭ but %d ╮:\n%s", width, opens, closes, stripANSI(out))
			}
		}
	}
}

func TestFlexColumnsNarrowTracksKeepBordersIntact(t *testing.T) {
	panes := []Pane{cellBoxPane("a"), cellBoxPane("b"), cellBoxPane("c"), cellBoxPane("d")}
	for _, width := range []int{28, 40, 60, 81, 120} {
		out := FlexColumns{FlexBounds: FlexBounds{MinWidth: 20, MaxCols: 4}}.Arrange(Frame{Width: width}, TierTall, panes, "")
		for _, r := range strings.Split(out, "\n") {
			if w := ansi.StringWidth(r); w != width {
				t.Fatalf("FlexColumns width %d: row width %d:\n%s", width, w, stripANSI(out))
			}
			s := stripANSI(r)
			if opens, closes := strings.Count(s, "╭"), strings.Count(s, "╮"); opens != closes {
				t.Fatalf("FlexColumns width %d: %d ╭ but %d ╮:\n%s", width, opens, closes, stripANSI(out))
			}
		}
	}
}

func TestCellBoxNeverExceedsFrameWidth(t *testing.T) {
	for width := 1; width <= 40; width++ {
		box := Frame{Width: width}.CellBox("title", "body")
		for _, r := range strings.Split(box, "\n") {
			if w := ansi.StringWidth(r); w > width {
				t.Fatalf("CellBox at width %d produced a %d-wide row: %q", width, w, stripANSI(r))
			}
		}
	}
}

func TestCellPanelNeverExceedsFrameWidth(t *testing.T) {
	lines := []string{"one", "two", "three", "four", "five", "six"}
	for width := 1; width <= 40; width++ {
		for _, height := range []int{0, 1, 6} {
			block := Frame{Width: width, Height: height}.CellPanel("title", lines, 2)
			for _, r := range strings.Split(block, "\n") {
				if w := ansi.StringWidth(r); w > width {
					t.Fatalf("CellPanel at width %d height %d produced a %d-wide row: %q",
						width, height, w, stripANSI(r))
				}
			}
		}
	}
}

func TestDegenerateFramesStillClamp(t *testing.T) {
	panes := []Pane{cellBoxPane("a"), cellBoxPane("b")}
	checks := []struct {
		name string
		out  func() string
	}{
		{"zero frame panel", func() string { return Frame{}.Panel("t", "body") }},
		{"negative frame titled box", func() string { return Frame{Width: -5}.TitledBox("t", "body") }},
		{"grid at w=0 h=0", func() string { return Grid{Cols: 2}.Arrange(Frame{}, TierTall, panes, "") }},
		{"grid cols 8 at width 1", func() string {
			return Grid{Cols: 8}.Arrange(Frame{Width: 1, Height: 3}, TierTall, panes, "")
		}},
		{"flex rows at width 1", func() string {
			return FlexRows{}.Arrange(Frame{Width: 1}, TierTall, panes, "")
		}},
		{"flex columns at width 1", func() string {
			return FlexColumns{}.Arrange(Frame{Width: 1}, TierTall, panes, "")
		}},
		{"cell box at width 1", func() string { return Frame{Width: 1, Height: 1}.CellBox("t", "body") }},
	}
	for _, c := range checks {
		if c.out() == "" && c.name != "grid at w=0 h=0" {
			t.Errorf("%s returned empty output", c.name)
		}
	}

	out := Grid{Cols: 8}.Arrange(Frame{Width: 1, Height: 3}, TierTall, panes, "")
	for i, r := range strings.Split(out, "\n") {
		if w := ansi.StringWidth(r); w != 1 {
			t.Errorf("grid cols 8 at width 1: row %d width %d, want 1: %q", i, w, stripANSI(r))
		}
	}
}

// TestSpreadFamilyIsWidthExact covers both spreaders: they right-align the same
// way, so they owe their callers the same width guarantee.
func TestSpreadFamilyIsWidthExact(t *testing.T) {
	bg := theme.Cur().Dim.GetForeground()
	long := "identity.that.is.long@example.com   12:34:56"
	cases := []struct {
		name        string
		left, right string
		width       int
	}{
		{"left overflows", " BRAND · deck", long, 40},
		{"right alone overflows", "", long, 20},
		{"both overflow", strings.Repeat("breadcrumb / ", 8), long, 30},
		{"exact fit", "abc", "def", 6},
		{"one short of fit", "abc", "def", 7},
		{"tiny width", "abc", "def", 3},
		{"width one", "abc", "def", 1},
		{"fits with slack", "left", "right", 40},
		{"wide right alone at two", "", "日本語", 2},
		{"wide right alone at three", "", "日本語", 3},
		{"wide right alone at four", "", "日本語", 4},
		{"wide both", "日本語", "日本語", 7},
		{"wide left narrow right", "日本語ですよ", "x", 5},
	}
	for _, c := range cases {
		for name, got := range map[string]string{
			"Spread":   Spread(c.left, c.right, c.width),
			"SpreadBG": SpreadBG(bg, c.left, c.right, c.width),
		} {
			if w := ansi.StringWidth(got); w != c.width {
				t.Errorf("%s(%s) width = %d, want %d: %q", name, c.name, w, c.width, stripANSI(got))
			}
			if strings.Contains(got, "\n") {
				t.Errorf("%s(%s) wrapped onto multiple lines: %q", name, c.name, stripANSI(got))
			}
		}
	}
}

func TestSpreadFamilyWidthSweep(t *testing.T) {
	bg := theme.Cur().Dim.GetForeground()
	parts := []string{"", "a", "abc", "日本語", "日", "a日b", "abc def", strings.Repeat("x", 12), "…"}
	for _, left := range parts {
		for _, right := range parts {
			for width := 1; width <= 14; width++ {
				for name, got := range map[string]string{
					"Spread":   Spread(left, right, width),
					"SpreadBG": SpreadBG(bg, left, right, width),
				} {
					if w := ansi.StringWidth(got); w != width {
						t.Fatalf("%s(%q, %q, %d) width = %d, want %d: %q",
							name, left, right, width, w, width, stripANSI(got))
					}
				}
			}
		}
	}
}

// TestCellTitledBoxNeverExceedsFrameWidth is the TitledBox counterpart of
// TestCellBoxNeverExceedsFrameWidth: MinTrackWidth (20) is below the 28 columns
// TitledBox needs, so tracks in [20,27] used to lose a right border. The
// clamp-down variant must fit any track.
func TestCellTitledBoxNeverExceedsFrameWidth(t *testing.T) {
	for width := 1; width <= 40; width++ {
		for _, box := range []string{
			Frame{Width: width}.CellTitledBox("title", "body"),
			Frame{Width: width}.CellTitledBoxIcon("✱", "title", "body"),
		} {
			for _, r := range strings.Split(box, "\n") {
				if w := ansi.StringWidth(r); w > width {
					t.Fatalf("CellTitledBox at width %d produced a %d-wide row: %q",
						width, w, stripANSI(r))
				}
			}
		}
	}
}

func TestIndentLinesClampsNegative(t *testing.T) {
	for _, n := range []int{-5, -1, 0} {
		if got := IndentLines("x\ny", n); got != "x\ny" {
			t.Errorf("IndentLines(n=%d) = %q, want unchanged", n, got)
		}
	}
	if got := IndentLines("x\ny", 1); got != " x\n y" {
		t.Errorf("IndentLines(n=1) = %q, want %q", got, " x\n y")
	}
}
