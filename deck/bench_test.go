package deck

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/theme"
)

const (
	benchScreenWidth  = 120
	benchScreenHeight = 40
	trueColorSGR      = "38;2;"
	trueColorFgSGR    = "\x1b[38;2;"
)

var benchSink string

func requireTrueColor(tb testing.TB) {
	tb.Helper()
	if lipgloss.ColorProfile() != termenv.TrueColor {
		lipgloss.SetColorProfile(termenv.TrueColor)
	}
	if got := lipgloss.ColorProfile(); got != termenv.TrueColor {
		tb.Fatalf("lipgloss.ColorProfile() = %v, want termenv.TrueColor", got)
	}
	if out := theme.Cur().Val.Render("probe"); !strings.Contains(out, trueColorFgSGR) {
		tb.Fatalf("theme render lacks %q (profile not in effect): %q", trueColorFgSGR, out)
	}
}

type benchView struct{}

func (benchView) Title() string                  { return "Audit" }
func (benchView) Init() tea.Cmd                  { return nil }
func (benchView) Update(*Model, tea.Msg) tea.Cmd { return nil }
func (benchView) Body(int, int) string           { return "body" }

func (benchView) Hints() []keys.Hint {
	return []keys.Hint{
		{Key: "↑↓", Label: "move"},
		{Key: "enter", Label: "open"},
		{Key: "r", Label: "refresh"},
		{Key: "f", Label: "filter"},
	}
}

func (benchView) Context() []keys.Hint {
	return []keys.Hint{
		{Key: "env", Label: "prod"},
		{Key: "scope", Label: "all"},
	}
}

func benchModel() *Model {
	return &Model{
		stack:  []View{benchView{}, benchView{}},
		width:  benchScreenWidth,
		height: benchScreenHeight,
		clock:  "12:34:56",
		chrome: Chrome{
			Brand:      "MUNIN",
			BrandGlyph: "◆",
			Subtitle:   "deck",
			ClockGlyph: "◷",
		},
		hasStatus: true,
		status: StatusInfo{
			Identity: "operator",
			Services: []ServiceStatus{
				{Name: "collector", Detail: "ok", Glyph: "●"},
				{Name: "store", Detail: "ok", Glyph: "●"},
				{Name: "queue", Detail: "lag 2s", Glyph: "●"},
			},
		},
		quitCheck: schemeQuit,
	}
}

func TestBenchmarksMeasureTrueColorRendering(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	requireTrueColor(t)

	h := benchModel()
	v := h.top()
	for name, out := range map[string]string{
		"header":         h.header(v),
		"footer":         h.footer(v),
		"breadcrumbs":    h.breadcrumbs(),
		"statusSegments": h.statusSegments(),
		"View":           h.View(),
	} {
		if !strings.Contains(out, trueColorSGR) {
			t.Errorf("%s output lacks %q: %q", name, trueColorSGR, out)
		}
	}
}

func TestBenchmarksUnderstateCostInAsciiProfile(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	h := benchModel()
	v := h.top()

	lipgloss.SetColorProfile(termenv.Ascii)
	ascii := h.header(v) + h.footer(v)
	lipgloss.SetColorProfile(termenv.TrueColor)
	truecolor := h.header(v) + h.footer(v)

	if strings.Contains(ascii, "\x1b[") {
		t.Errorf("Ascii profile emitted escapes: %q", ascii)
	}
	if len(truecolor) <= len(ascii) {
		t.Errorf("TrueColor output (%d bytes) should exceed Ascii output (%d bytes)", len(truecolor), len(ascii))
	}
}

func BenchmarkChromeHeader(b *testing.B) {
	requireTrueColor(b)
	h := benchModel()
	v := h.top()
	b.ReportAllocs()
	for b.Loop() {
		benchSink = h.header(v)
	}
}

func BenchmarkChromeFooter(b *testing.B) {
	requireTrueColor(b)
	h := benchModel()
	v := h.top()
	b.ReportAllocs()
	for b.Loop() {
		benchSink = h.footer(v)
	}
}

func BenchmarkChromeHeaderPlusFooterPerFrame(b *testing.B) {
	requireTrueColor(b)
	h := benchModel()
	v := h.top()
	b.ReportAllocs()
	for b.Loop() {
		benchSink = h.header(v) + h.footer(v)
	}
}

func BenchmarkChromeBreadcrumbs(b *testing.B) {
	requireTrueColor(b)
	h := benchModel()
	b.ReportAllocs()
	for b.Loop() {
		benchSink = h.breadcrumbs()
	}
}

func BenchmarkChromeStatusSegments(b *testing.B) {
	requireTrueColor(b)
	h := benchModel()
	b.ReportAllocs()
	for b.Loop() {
		benchSink = h.statusSegments()
	}
}

func BenchmarkHostViewWholeFrame(b *testing.B) {
	requireTrueColor(b)
	h := benchModel()
	b.ReportAllocs()
	for b.Loop() {
		benchSink = h.View()
	}
}
