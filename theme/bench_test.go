package theme

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	trueColorSGR   = "38;2;"
	trueColorFgSGR = "\x1b[38;2;"
)

var (
	benchSink  string
	benchColor lipgloss.TerminalColor
)

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
	if out := Cur().Val.Render("probe"); !strings.Contains(out, trueColorFgSGR) {
		tb.Fatalf("theme render lacks %q (profile not in effect): %q", trueColorFgSGR, out)
	}
}

func TestBenchmarksMeasureTrueColorRendering(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	requireTrueColor(t)

	fg := Cur().Val.GetForeground()
	for name, out := range map[string]string{
		"StripText":  StripText(fg, "probe"),
		"StripBold":  StripBold(fg, "probe"),
		"StripBlock": StripBlock(RuleWidth, StripText(fg, "probe")),
	} {
		if !strings.Contains(out, trueColorSGR) {
			t.Errorf("%s output lacks %q: %q", name, trueColorSGR, out)
		}
	}
}

func TestBenchmarksUnderstateCostInAsciiProfile(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	fg := Cur().Val.GetForeground()

	lipgloss.SetColorProfile(termenv.Ascii)
	ascii := StripText(fg, "probe")
	lipgloss.SetColorProfile(termenv.TrueColor)
	truecolor := StripText(fg, "probe")

	if strings.Contains(ascii, "\x1b[") {
		t.Errorf("Ascii profile emitted escapes: %q", ascii)
	}
	if len(truecolor) <= len(ascii) {
		t.Errorf("TrueColor output (%d bytes) should exceed Ascii output (%d bytes)", len(truecolor), len(ascii))
	}
}

func BenchmarkStripText(b *testing.B) {
	requireTrueColor(b)
	fg := Cur().Val.GetForeground()
	s := "munin · audit"
	b.ReportAllocs()
	for b.Loop() {
		benchSink = StripText(fg, s)
	}
}

func BenchmarkStripBold(b *testing.B) {
	requireTrueColor(b)
	fg := Cur().Accent.GetForeground()
	s := "munin · audit"
	b.ReportAllocs()
	for b.Loop() {
		benchSink = StripBold(fg, s)
	}
}

func BenchmarkStripTextEmptyString(b *testing.B) {
	requireTrueColor(b)
	fg := Cur().Val.GetForeground()
	b.ReportAllocs()
	for b.Loop() {
		benchSink = StripText(fg, "")
	}
}

func BenchmarkStripBg(b *testing.B) {
	requireTrueColor(b)
	b.ReportAllocs()
	for b.Loop() {
		benchColor = StripBg()
	}
}

func BenchmarkStripBlockTwoLines(b *testing.B) {
	requireTrueColor(b)
	fg := Cur().Val.GetForeground()
	left := StripText(fg, " brand · deck")
	right := StripBold(fg, "12:34:56 ")
	b.ReportAllocs()
	for b.Loop() {
		benchSink = StripBlock(RuleWidth, left, right)
	}
}

func BenchmarkStripCalls20PerFrame(b *testing.B) {
	requireTrueColor(b)
	fg := Cur().Val.GetForeground()
	accent := Cur().Accent.GetForeground()
	s := "munin · audit"
	b.ReportAllocs()
	for b.Loop() {
		var sb strings.Builder
		for range 15 {
			sb.WriteString(StripText(fg, s))
		}
		for range 5 {
			sb.WriteString(StripBold(accent, s))
		}
		benchSink = sb.String()
	}
}
