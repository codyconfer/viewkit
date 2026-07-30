package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/layout"
)

func tableLines(t *testing.T, out string) []string {
	t.Helper()
	return strings.Split(stripANSI(out), "\n")
}

func tableCells(line string) []string { return strings.Split(line, tableCellSep) }

func TestTableAlignsWideRunes(t *testing.T) {
	headers := []string{"name", "note"}
	rows := [][]string{
		{"漢字", "ok"},
		{"abcd", "fine"},
		{"日本語テスト", "x"},
	}
	lines := tableLines(t, Table(layout.DefaultFrame(), headers, rows))
	if len(lines) != 6 {
		t.Fatalf("got %d lines, want 6 (header, rule, 3 rows, footer):\n%s", len(lines), strings.Join(lines, "\n"))
	}

	want := lipgloss.Width(lines[0])
	if sum := 12 + lipgloss.Width(tableCellSep) + 4; want != sum {
		t.Fatalf("header display width = %d, want %d", want, sum)
	}
	for i, line := range lines[:len(lines)-1] {
		if w := lipgloss.Width(line); w != want {
			t.Errorf("line %d display width = %d, want %d: %q", i, w, want, line)
		}
	}

	for _, i := range []int{0, 2, 3, 4} {
		cells := tableCells(lines[i])
		if len(cells) != 2 {
			t.Fatalf("line %d split into %d cells, want 2: %q", i, len(cells), lines[i])
		}
		if w := lipgloss.Width(cells[0]); w != 12 {
			t.Errorf("line %d column 0 width = %d, want 12: %q", i, w, cells[0])
		}
		if w := lipgloss.Width(cells[1]); w != 4 {
			t.Errorf("line %d column 1 width = %d, want 4: %q", i, w, cells[1])
		}
	}

	if got, want := lines[0], "name         │ note"; got != want {
		t.Errorf("header = %q, want %q", got, want)
	}
	if got, want := lines[2], "漢字         │ ok  "; got != want {
		t.Errorf("wide row = %q, want %q", got, want)
	}
	if got, want := lines[4], "日本語テスト │ x   "; got != want {
		t.Errorf("wide row = %q, want %q", got, want)
	}
}

func TestTableAlignsEmoji(t *testing.T) {
	rows := [][]string{{"🐔🐔", "hen"}, {"abcd", "ok"}}
	lines := tableLines(t, Table(layout.DefaultFrame(), []string{"who", "what"}, rows))
	base := lipgloss.Width(lines[0])
	for i, line := range lines[:len(lines)-1] {
		if w := lipgloss.Width(line); w != base {
			t.Errorf("line %d display width = %d, want %d: %q", i, w, base, line)
		}
	}
}

func TestTableNoColumns(t *testing.T) {
	if got := stripANSI(Table(layout.DefaultFrame(), nil, [][]string{{"a"}})); got != "(no columns)" {
		t.Fatalf("Table(nil headers) = %q, want %q", got, "(no columns)")
	}
}

func TestTableZeroRows(t *testing.T) {
	lines := tableLines(t, Table(layout.DefaultFrame(), []string{"a", "bb"}, nil))
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (header, rule, footer):\n%s", len(lines), strings.Join(lines, "\n"))
	}
	if lines[0] != "a │ bb" {
		t.Errorf("header = %q, want %q", lines[0], "a │ bb")
	}
	if lines[1] != "─"+tableRuleSep+"──" {
		t.Errorf("rule = %q", lines[1])
	}
	if lines[2] != "(0 rows)" {
		t.Errorf("footer = %q, want %q", lines[2], "(0 rows)")
	}
}

func TestTableFooterCountsRows(t *testing.T) {
	rows := [][]string{{"1"}, {"2"}, {"3"}}
	lines := tableLines(t, Table(layout.DefaultFrame(), []string{"n"}, rows))
	if got := lines[len(lines)-1]; got != "(3 rows)" {
		t.Fatalf("footer = %q, want %q", got, "(3 rows)")
	}
}

func TestTableRuleMatchesColumns(t *testing.T) {
	lines := tableLines(t, Table(layout.DefaultFrame(), []string{"aaa", "b"}, [][]string{{"x", "yy"}}))
	if got, want := lines[1], "───"+tableRuleSep+"──"; got != want {
		t.Fatalf("rule = %q, want %q", got, want)
	}
	if w, h := lipgloss.Width(lines[1]), lipgloss.Width(lines[0]); w != h {
		t.Fatalf("rule width %d != header width %d", w, h)
	}
}

func TestTableRaggedRows(t *testing.T) {
	headers := []string{"a", "b"}
	rows := [][]string{
		{"1", "2", "3", "4"},
		{"5"},
		nil,
	}
	lines := tableLines(t, Table(layout.DefaultFrame(), headers, rows))
	if len(lines) != 6 {
		t.Fatalf("got %d lines, want 6:\n%s", len(lines), strings.Join(lines, "\n"))
	}
	if got, want := lines[2], "1 │ 2"; got != want {
		t.Errorf("overlong row = %q, want %q (extra cells dropped)", got, want)
	}
	if got, want := lines[3], "5 │  "; got != want {
		t.Errorf("short row = %q, want %q (missing cell padded)", got, want)
	}
	if got, want := lines[4], "  │  "; got != want {
		t.Errorf("nil row = %q, want %q", got, want)
	}
	for i, line := range lines[:len(lines)-1] {
		if w := lipgloss.Width(line); w != 5 {
			t.Errorf("line %d width = %d, want 5: %q", i, w, line)
		}
	}
}

func TestTableOverlongRowDoesNotWidenTable(t *testing.T) {
	out := Table(layout.DefaultFrame(), []string{"a"}, [][]string{{"x", "a very long dropped cell"}})
	lines := tableLines(t, out)
	if got := lines[0]; got != "a" {
		t.Fatalf("header = %q, want %q", got, "a")
	}
	if got := lines[2]; got != "x" {
		t.Fatalf("row = %q, want %q", got, "x")
	}
}

func TestTableTruncatesWithEllipsis(t *testing.T) {
	rows := [][]string{{"abcdefghij"}, {"hi"}}
	lines := tableLines(t, Table(layout.DefaultFrame(), []string{"col"}, rows, TableOpts{MaxCell: 5}))
	if got, want := lines[2], "abcd…"; got != want {
		t.Errorf("truncated cell = %q, want %q", got, want)
	}
	if got, want := lines[3], "hi   "; got != want {
		t.Errorf("padded cell = %q, want %q", got, want)
	}
	for i, line := range lines[:len(lines)-1] {
		if w := lipgloss.Width(line); w != 5 {
			t.Errorf("line %d width = %d, want 5: %q", i, w, line)
		}
	}
}

func TestTableTruncatesWideRunesToDisplayWidth(t *testing.T) {
	rows := [][]string{{"日本語テスト"}, {"abcde"}}
	lines := tableLines(t, Table(layout.DefaultFrame(), []string{"c"}, rows, TableOpts{MaxCell: 5}))
	for i, line := range lines[:len(lines)-1] {
		if w := lipgloss.Width(line); w != 5 {
			t.Errorf("line %d width = %d, want 5: %q", i, w, line)
		}
	}
	if !strings.Contains(lines[2], "…") {
		t.Errorf("wide cell %q missing ellipsis", lines[2])
	}
}

func TestTableDefaultMaxCell(t *testing.T) {
	long := strings.Repeat("x", TableMaxCell+20)
	lines := tableLines(t, Table(layout.DefaultFrame(), []string{"c"}, [][]string{{long}}))
	if w := lipgloss.Width(lines[2]); w != TableMaxCell {
		t.Fatalf("cell width = %d, want default cap %d", w, TableMaxCell)
	}
}

func TestTableMaxCellClampedToFrame(t *testing.T) {
	f := layout.NewFrame(24)
	long := strings.Repeat("x", 60)
	lines := tableLines(t, Table(f, []string{"c"}, [][]string{{long}}, TableOpts{MaxCell: 50}))
	if w := lipgloss.Width(lines[2]); w != f.BodyWidth() {
		t.Fatalf("cell width = %d, want frame body width %d", w, f.BodyWidth())
	}
}

func TestTableFlattensCells(t *testing.T) {
	rows := [][]string{{"a\r\nb\nc\td", "plain"}}
	out := Table(layout.DefaultFrame(), []string{"multi\nhead", "x"}, rows)
	lines := tableLines(t, out)
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4 (embedded newlines must be flattened):\n%s", len(lines), strings.Join(lines, "\n"))
	}
	if got := tableCells(lines[2])[0]; got != "a b c d   " {
		t.Errorf("flattened cell = %q, want %q", got, "a b c d   ")
	}
	if got := tableCells(lines[0])[0]; got != "multi head" {
		t.Errorf("flattened header = %q, want %q", got, "multi head")
	}
}

func TestTableColumnWidthFromHeader(t *testing.T) {
	lines := tableLines(t, Table(layout.DefaultFrame(), []string{"longheader", "b"}, [][]string{{"x", "y"}}))
	cells := tableCells(lines[2])
	if w := lipgloss.Width(cells[0]); w != len("longheader") {
		t.Fatalf("column 0 width = %d, want %d (header is widest)", w, len("longheader"))
	}
}

func TestTableZeroFrameUsesDefaultBodyWidth(t *testing.T) {
	long := strings.Repeat("y", 100)
	lines := tableLines(t, Table(layout.Frame{}, []string{"c"}, [][]string{{long}}))
	if w := lipgloss.Width(lines[2]); w != TableMaxCell {
		t.Fatalf("cell width = %d, want %d", w, TableMaxCell)
	}
}
