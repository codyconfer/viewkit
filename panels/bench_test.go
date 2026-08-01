package panels

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

const (
	benchFrameWidth = 81
	trueColorSGR    = "38;2;"
	trueColorFgSGR  = "\x1b[38;2;"
)

var benchSink string

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	os.Exit(m.Run())
}

func requireTrueColor(tb testing.TB) {
	tb.Helper()
	if lipgloss.ColorProfile() != termenv.TrueColor {
		lipgloss.SetColorProfile(termenv.TrueColor)
	}
	if got := lipgloss.ColorProfile(); got != termenv.TrueColor {
		tb.Fatalf("lipgloss.ColorProfile() = %v, want termenv.TrueColor", got)
	}
	if out := theme.Default().Val.Render("probe"); !strings.Contains(out, trueColorFgSGR) {
		tb.Fatalf("theme render lacks %q (profile not in effect): %q", trueColorFgSGR, out)
	}
}

func TestBenchmarksMeasureTrueColorRendering(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	requireTrueColor(t)

	f := layout.NewFrame(benchFrameWidth)
	headers, rows := benchTableFixture(2, 3)
	for name, out := range map[string]string{
		"Markdown heading":  Markdown(f, "# Title"),
		"Markdown emphasis": Markdown(f, "some **bold** text"),
		"Markdown list":     Markdown(f, "- an item"),
		"Table":             Table(f, headers, rows),
	} {
		if !strings.Contains(out, trueColorSGR) {
			t.Errorf("%s output lacks %q: %q", name, trueColorSGR, out)
		}
	}

	plain := Markdown(f, "a plain sentence of prose with no markdown syntax at all")
	if strings.Contains(plain, "\x1b[") {
		t.Errorf("a plain markdown line should emit no escapes: %q", plain)
	}
}

func TestBenchmarksUnderstateCostInAsciiProfile(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	f := layout.NewFrame(benchFrameWidth)
	src := benchMarkdownDoc20Lines()

	lipgloss.SetColorProfile(termenv.Ascii)
	ascii := Markdown(f, src)
	lipgloss.SetColorProfile(termenv.TrueColor)
	truecolor := Markdown(f, src)

	if strings.Contains(ascii, "\x1b[") {
		t.Errorf("Ascii profile emitted escapes: %q", ascii)
	}
	if len(truecolor) <= len(ascii) {
		t.Errorf("TrueColor output (%d bytes) should exceed Ascii output (%d bytes)", len(truecolor), len(ascii))
	}
}

func benchMarkdownDoc20Lines() string {
	return strings.Join([]string{
		"# Release Notes",
		"",
		"## Summary",
		"The **audit** pipeline now records *every* directive in `config.yaml`.",
		"",
		"### Details",
		"- collector rewritten for streaming",
		"- formatter cache reuses templates",
		"- see [the plan](https://example.invalid/plan) for context",
		"",
		"1. gather signals",
		"2. normalize records",
		"3. emit report",
		"",
		"> Latency budget is 16ms per frame.",
		"",
		"---",
		"Plain trailing paragraph with no inline markup whatsoever here.",
		"A second plain paragraph line to round out the twenty line document.",
		"",
	}, "\n")
}

func benchTableFixture(rows, cols int) ([]string, [][]string) {
	headers := make([]string, cols)
	for c := range headers {
		headers[c] = "column" + strconv.Itoa(c)
	}
	data := make([][]string, rows)
	for r := range data {
		row := make([]string, cols)
		for c := range row {
			row[c] = fmt.Sprintf("r%02dc%d-value", r, c)
		}
		data[r] = row
	}
	return headers, data
}

func BenchmarkMarkdownPlainLine(b *testing.B) {
	requireTrueColor(b)
	f := layout.NewFrame(benchFrameWidth)
	src := "a plain sentence of prose with no markdown syntax at all"
	b.ReportAllocs()
	for b.Loop() {
		benchSink = Markdown(f, src)
	}
}

func BenchmarkMarkdownDoc20Lines(b *testing.B) {
	requireTrueColor(b)
	f := layout.NewFrame(benchFrameWidth)
	src := benchMarkdownDoc20Lines()
	if n := len(strings.Split(src, "\n")); n != 20 {
		b.Fatalf("fixture has %d lines, want 20", n)
	}
	b.ReportAllocs()
	for b.Loop() {
		benchSink = Markdown(f, src)
	}
}

func BenchmarkMarkdownPanelDoc20Lines(b *testing.B) {
	requireTrueColor(b)
	f := layout.NewFrame(benchFrameWidth)
	src := benchMarkdownDoc20Lines()
	b.ReportAllocs()
	for b.Loop() {
		benchSink = MarkdownPanel(f, "Notes", src)
	}
}

func BenchmarkTable40Rows6Cols(b *testing.B) {
	requireTrueColor(b)
	f := layout.NewFrame(benchFrameWidth)
	headers, rows := benchTableFixture(40, 6)
	b.ReportAllocs()
	for b.Loop() {
		benchSink = Table(f, headers, rows)
	}
}

func BenchmarkTable1Row6Cols(b *testing.B) {
	requireTrueColor(b)
	f := layout.NewFrame(benchFrameWidth)
	headers, rows := benchTableFixture(1, 6)
	b.ReportAllocs()
	for b.Loop() {
		benchSink = Table(f, headers, rows)
	}
}
