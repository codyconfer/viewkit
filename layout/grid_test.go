package layout

import (
	"strings"
	"testing"
)

func fixedPane(name string, interactive bool, pos *GridPos) Pane {
	return Pane{
		Name:        name,
		Interactive: interactive,
		Pos:         pos,
		Render: func(f Frame) string {
			body := name
			if f.Focused {
				body = name + "*"
			}
			if f.Height < 1 {
				return body
			}
			lines := make([]string, f.Height)
			for i := range lines {
				lines[i] = body
			}
			return FitBlock(strings.Join(lines, "\n"), f.Width, f.Height)
		},
	}
}

func lineAt(out string, y int) string {
	lines := strings.Split(out, "\n")
	if y < 0 || y >= len(lines) {
		return ""
	}
	return lines[y]
}

func TestGridTilesTwoColumnsGapFree(t *testing.T) {
	t.Parallel()
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("left", false, &GridPos{Col: 0, Row: 0}),
			fixedPane("right", false, &GridPos{Col: 1, Row: 0}),
		},
	}
	out := scr.Render(Frame{Width: 81, Height: 4}, TierTall, "")
	rows := strings.Split(out, "\n")
	if len(rows) != 4 {
		t.Fatalf("grid height = %d rows, want 4:\n%s", len(rows), out)
	}
	for _, r := range rows {
		if w := len([]rune(r)); w != 81 {
			t.Fatalf("row width = %d, want 81 (gap-free tiling):\n%q", w, r)
		}
	}
	if !strings.Contains(rows[0], "left") || !strings.Contains(rows[0], "right") {
		t.Fatalf("both columns should occupy row 0:\n%s", out)
	}
	if strings.Index(rows[0], "left") >= strings.Index(rows[0], "right") {
		t.Fatalf("left column should precede right:\n%s", out)
	}
}

func TestGridEdgeTilingHasNoOddWidthGap(t *testing.T) {
	t.Parallel()
	cells := []gridCell{{x: 0, w: 1}, {x: 1, w: 1}, {x: 2, w: 1}}
	total := 0
	prevEnd := 0
	for _, c := range cells {
		r := pixelRect(c, 3, 1, 81, 1)
		if r.x != prevEnd {
			t.Fatalf("column %d starts at %d, want %d (no gap/overlap)", c.x, r.x, prevEnd)
		}
		prevEnd = r.x + r.w
		total += r.w
	}
	if total != 81 {
		t.Fatalf("columns sum to %d, want 81", total)
	}
}

func TestGridRowSpanStacksVertically(t *testing.T) {
	t.Parallel()
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("tall", false, &GridPos{Col: 0, Row: 0, RowSpan: 2}),
			fixedPane("top", false, &GridPos{Col: 1, Row: 0}),
			fixedPane("bot", false, &GridPos{Col: 1, Row: 1}),
		},
	}
	out := scr.Render(Frame{Width: 40, Height: 2}, TierTall, "")
	if !strings.Contains(lineAt(out, 0), "tall") || !strings.Contains(lineAt(out, 1), "tall") {
		t.Fatalf("row-spanning pane should appear on both rows:\n%s", out)
	}
	if !strings.Contains(lineAt(out, 0), "top") {
		t.Fatalf("top pane missing from row 0:\n%s", out)
	}
	if !strings.Contains(lineAt(out, 1), "bot") {
		t.Fatalf("bot pane missing from row 1:\n%s", out)
	}
	if strings.Contains(lineAt(out, 0), "bot") {
		t.Fatalf("bot pane leaked into row 0:\n%s", out)
	}
}

func TestGridAutoFlowFillsFreeCells(t *testing.T) {
	t.Parallel()
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("a", false, nil),
			fixedPane("b", false, nil),
			fixedPane("c", false, nil),
		},
	}
	out := scr.Render(Frame{Width: 40, Height: 2}, TierTall, "")
	if !strings.Contains(lineAt(out, 0), "a") || !strings.Contains(lineAt(out, 0), "b") {
		t.Fatalf("a and b should auto-flow into row 0:\n%s", out)
	}
	if !strings.Contains(lineAt(out, 1), "c") {
		t.Fatalf("c should auto-flow into row 1:\n%s", out)
	}
}

func TestGridCollidingPositionsKeepEveryPane(t *testing.T) {
	t.Parallel()
	panes := []Pane{
		fixedPane("A", false, &GridPos{Col: 0, Row: 0}),
		fixedPane("B", false, &GridPos{Col: 0, Row: 0}),
		fixedPane("C", false, &GridPos{Col: 1, Row: 0}),
	}
	out := Grid{Cols: 2}.Arrange(Frame{Width: 80, Height: 4}, TierTall, panes, "")
	for i, r := range strings.Split(out, "\n") {
		if w := ansiWidth(r); w != 80 {
			t.Fatalf("row %d width = %d, want 80:\n%s", i, w, stripANSI(out))
		}
	}
	for _, name := range []string{"A", "B", "C"} {
		if !strings.Contains(stripANSI(out), name) {
			t.Errorf("pane %q vanished from a colliding grid:\n%s", name, stripANSI(out))
		}
	}
}

func TestGridCollidingPositionsCascadeToNextRow(t *testing.T) {
	t.Parallel()
	panes := []Pane{
		fixedPane("A", false, &GridPos{Col: 0, Row: 0}),
		fixedPane("B", false, &GridPos{Col: 0, Row: 0}),
	}
	out := Grid{Cols: 1}.Arrange(Frame{Width: 40, Height: 4}, TierTall, panes, "")
	rows := strings.Split(out, "\n")
	if !strings.Contains(stripANSI(rows[0]), "A") {
		t.Fatalf("pane A missing from row 0:\n%s", stripANSI(out))
	}
	if strings.Contains(stripANSI(rows[0]), "B") {
		t.Fatalf("pane B drew into A's row in a single-column grid:\n%s", stripANSI(out))
	}
	if !strings.Contains(stripANSI(rows[len(rows)-1]), "B") {
		t.Fatalf("pane B never cascaded to a free row:\n%s", stripANSI(out))
	}
}

func TestGridReflowKeepsColSpanIntent(t *testing.T) {
	t.Parallel()
	panes := []Pane{
		fixedPane("HDR", false, &GridPos{Col: 0, Row: 0, ColSpan: 3}),
		fixedPane("x", false, nil),
		fixedPane("y", false, nil),
		fixedPane("z", false, nil),
	}
	out := Grid{Cols: 3}.Arrange(Frame{Width: 59, Height: 6}, TierTall, panes, "")
	for i, r := range strings.Split(out, "\n") {
		if w := ansiWidth(r); w != 59 {
			t.Fatalf("row %d width = %d, want 59:\n%s", i, w, stripANSI(out))
		}
	}
	row0 := stripANSI(lineAt(out, 0))
	if !strings.Contains(row0, "HDR") {
		t.Fatalf("header missing from row 0:\n%s", stripANSI(out))
	}
	for _, name := range []string{"x", "y", "z"} {
		if strings.Contains(row0, name) {
			t.Errorf("pane %q shares row 0 with the full-width header (ColSpan discarded):\n%s",
				name, stripANSI(out))
		}
	}
	for _, name := range []string{"x", "y", "z"} {
		if !strings.Contains(stripANSI(out), name) {
			t.Errorf("pane %q vanished:\n%s", name, stripANSI(out))
		}
	}
}

func TestGridReflowKeepsRowSpanIntent(t *testing.T) {
	t.Parallel()
	panes := []Pane{
		fixedPane("A", false, &GridPos{Col: 0, Row: 0, RowSpan: 2}),
		fixedPane("B", false, nil),
	}
	out := Grid{Cols: 2}.Arrange(Frame{Width: 39, Height: 6}, TierTall, panes, "")
	rows := strings.Split(out, "\n")
	tall, short := 0, 0
	for _, r := range rows {
		s := stripANSI(r)
		if strings.Contains(s, "A") {
			tall++
		}
		if strings.Contains(s, "B") {
			short++
		}
	}
	if tall <= short {
		t.Fatalf("row-spanning pane got %d rows and its neighbour %d; the span was discarded:\n%s",
			tall, short, stripANSI(out))
	}
}

// TestGridReflowCapsAbsurdRowSpan makes sure a reflowed span stays searchable:
// autoFlow used to zero every span, so an absurd RowSpan was silently ignored;
// now that spans survive, the free-block search must not walk them forever.
func TestGridReflowCapsAbsurdRowSpan(t *testing.T) {
	t.Parallel()
	panes := []Pane{
		fixedPane("A", false, &GridPos{Col: 0, Row: 0, RowSpan: 1 << 30}),
		fixedPane("B", false, nil),
	}
	out := Grid{Cols: 2}.Arrange(Frame{Width: 39, Height: 4}, TierTall, panes, "")
	rows := strings.Split(out, "\n")
	if len(rows) != 4 {
		t.Fatalf("grid height = %d rows, want 4", len(rows))
	}
	for i, r := range rows {
		if w := ansiWidth(r); w != 39 {
			t.Fatalf("row %d width = %d, want 39:\n%s", i, w, stripANSI(out))
		}
	}
	for _, name := range []string{"A", "B"} {
		if !strings.Contains(stripANSI(out), name) {
			t.Errorf("pane %q vanished:\n%s", name, stripANSI(out))
		}
	}
}

// TestGridCellsNeverOverlap is the invariant composite depends on: it lays rows
// out left to right and pads them to the frame width, so any pair of overlapping
// rects would push a pane off the end of its row and delete it outright.
func TestGridCellsNeverOverlap(t *testing.T) {
	t.Parallel()
	specs := [][]*GridPos{
		{{Col: 0, Row: 0}, {Col: 0, Row: 0}, {Col: 1, Row: 0}},
		{{Col: 0, Row: 0, ColSpan: 3}, {Col: 1, Row: 0}, nil},
		{{Col: 0, Row: 0, RowSpan: 3}, {Col: 0, Row: 1}, {Col: 0, Row: 2}},
		{{Col: 2, Row: 1, ColSpan: 2}, {Col: 3, Row: 1}, nil, nil},
		{nil, {Col: 0, Row: 0}, nil, {Col: 0, Row: 0, ColSpan: 2}},
		{{Col: -4, Row: -4}, {Col: 0, Row: 0}, {Col: 9, Row: 0}, {Col: 9, Row: 0}},
	}
	for si, spec := range specs {
		panes := make([]Pane, 0, len(spec))
		for i, pos := range spec {
			panes = append(panes, fixedPane(string(rune('A'+i)), false, pos))
		}
		for _, cols := range []int{1, 2, 3, 4} {
			for _, width := range []int{20, 40, 59, 80, 120} {
				g := Grid{Cols: FitCols(cols, width)}
				slots := gridSlots(panes)
				if g.Cols < cols {
					slots = autoFlow(slots, g.Cols)
				}
				cells, gc, gr := g.place(slots)
				for a := range cells {
					for b := a + 1; b < len(cells); b++ {
						if cellsOverlap(cells[a], cells[b]) {
							t.Fatalf("spec %d cols %d width %d: cells %d %+v and %d %+v overlap",
								si, cols, width, a, cells[a], b, cells[b])
						}
					}
				}
				out := Grid{Cols: cols}.Arrange(Frame{Width: width, Height: 8}, TierTall, panes, "")
				for i, r := range strings.Split(out, "\n") {
					if w := ansiWidth(r); w != width {
						t.Fatalf("spec %d cols %d width %d: row %d width %d:\n%s",
							si, cols, width, i, w, stripANSI(out))
					}
				}
				for i := range spec {
					if !strings.Contains(stripANSI(out), string(rune('A'+i))) {
						t.Fatalf("spec %d cols %d width %d (grid %dx%d): pane %q vanished:\n%s",
							si, cols, width, gc, gr, string(rune('A'+i)), stripANSI(out))
					}
				}
			}
		}
	}
}

func cellsOverlap(a, b gridCell) bool {
	return a.x < b.x+b.w && b.x < a.x+a.w && a.y < b.y+b.h && b.y < a.y+a.h
}

func TestGridColSpanClampsToGrid(t *testing.T) {
	t.Parallel()
	c := cellFor(GridPos{Col: 1, ColSpan: 5}, 3)
	if c.x != 1 || c.w != 2 {
		t.Fatalf("cellFor clamp = {x:%d w:%d}, want {x:1 w:2}", c.x, c.w)
	}
}

func TestGridFocusesRingSelection(t *testing.T) {
	t.Parallel()
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("a", true, &GridPos{Col: 0}),
			fixedPane("b", true, &GridPos{Col: 1}),
		},
	}
	out := scr.Render(Frame{Width: 40, Height: 1}, TierTall, "b")
	if !strings.Contains(out, "b*") {
		t.Fatalf("focused pane b should render focused:\n%s", out)
	}
	if strings.Contains(out, "a*") {
		t.Fatalf("unfocused pane a rendered focused:\n%s", out)
	}
}

func TestGridFallsBackToStackWithoutHeight(t *testing.T) {
	t.Parallel()
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("a", false, &GridPos{Col: 0}),
			fixedPane("b", false, &GridPos{Col: 1}),
		},
	}
	out := scr.Render(NewFrame(40), TierTall, "")
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("fallback should still render both panes:\n%s", out)
	}
}

func demoPanel(title string, body ...string) Pane {
	return Pane{
		Title: title,
		Render: func(f Frame) string {
			inner := NewFrame(f.Width - 4)
			return FitBlock(inner.Panel(title, body...), f.Width, f.Height)
		},
	}
}

func TestGridRendersBorderedDashboard(t *testing.T) {
	t.Parallel()
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			demoPanel("STATUS", "tokens 1.2M", "eggs 340", "Lv.7"),
			demoPanel("MARKET", "price 12.4", "trend up", "demand hi"),
			{
				Pos:   &GridPos{Col: 0, Row: 1, ColSpan: 2},
				Title: "FEED",
				Render: func(f Frame) string {
					inner := NewFrame(f.Width - 4)
					return FitBlock(inner.Panel("FEED", "honk", "honk", "sold 50 eggs"), f.Width, f.Height)
				},
			},
		},
	}
	out := scr.Render(Frame{Width: 64, Height: 14}, TierTall, "")
	rows := strings.Split(out, "\n")
	if len(rows) != 14 {
		t.Fatalf("dashboard height = %d rows, want 14", len(rows))
	}
	for i, r := range rows {
		if w := ansiWidth(r); w != 64 {
			t.Fatalf("row %d width = %d, want 64:\n%q", i, w, r)
		}
	}
	if !strings.Contains(out, "STATUS") || !strings.Contains(out, "MARKET") || !strings.Contains(out, "FEED") {
		t.Fatalf("all three panels should render:\n%s", out)
	}
	t.Log("\n" + out)
}

func ansiWidth(s string) int {
	return len([]rune(stripANSI(s)))
}

func TestFitBlockClipsAndPads(t *testing.T) {
	t.Parallel()
	got := FitBlock("abcdef\nxy", 4, 3)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d: %q", len(lines), got)
	}
	if lines[0] != "abcd" {
		t.Fatalf("line 0 = %q, want abcd (clipped)", lines[0])
	}
	if lines[1] != "xy  " {
		t.Fatalf("line 1 = %q, want 'xy  ' (padded)", lines[1])
	}
	if lines[2] != "    " {
		t.Fatalf("line 2 = %q, want 4 spaces (blank row)", lines[2])
	}
}

func TestGridSlimNarrowsAndDonates(t *testing.T) {
	t.Parallel()
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			func() Pane { p := fixedPane("slim", false, &GridPos{Col: 0, Row: 0}); p.Slim = true; return p }(),
			fixedPane("wide", false, &GridPos{Col: 1, Row: 0}),
		},
	}
	out := scr.Render(Frame{Width: 80, Height: 1}, TierTall, "")
	row := lineAt(out, 0)
	if w := ansiWidth(row); w != 80 {
		t.Fatalf("slim row width = %d, want gap-free 80:\n%q", w, row)
	}
	wideStart := strings.Index(row, "wide")
	if wideStart < 0 {
		t.Fatalf("wide pane missing:\n%q", row)
	}
	if wideStart >= 40 {
		t.Fatalf("slim pane did not narrow: wide starts at %d, want < 40:\n%q", wideStart, row)
	}
	if wideStart < 20 {
		t.Fatalf("slim pane shrank past the 20-col floor: wide starts at %d:\n%q", wideStart, row)
	}
}

func TestGridSlimFloorAt20(t *testing.T) {
	t.Parallel()
	if got := slimWidth(40); got != 20 {
		t.Fatalf("slimWidth(40) = %d, want 20", got)
	}
	if got := slimWidth(30); got != 20 {
		t.Fatalf("slimWidth(30) = %d, want 20 (floored)", got)
	}
	if got := slimWidth(50); got != 25 {
		t.Fatalf("slimWidth(50) = %d, want 25 (half)", got)
	}
	if got := slimWidth(16); got != 16 {
		t.Fatalf("slimWidth(16) = %d, want 16 (already narrower than floor)", got)
	}
}

func TestGridSoleAutoPaneFillsWidth(t *testing.T) {
	t.Parallel()
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("a", false, &GridPos{Col: 0, Row: 0, ColSpan: 2}),
			fixedPane("b", false, nil),
		},
	}
	out := scr.Render(Frame{Width: 60, Height: 2}, TierTall, "")
	if w := ansiWidth(lineAt(out, 1)); w != 60 {
		t.Fatalf("sole auto pane row width = %d, want 60:\n%s", w, out)
	}
}
