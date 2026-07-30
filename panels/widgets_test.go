package panels

import (
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

func TestMeterClampsToWidth(t *testing.T) {
	out := stripANSI(Meter(2, 4))
	if got := strings.Count(out, "█"); got != 4 {
		t.Fatalf("Meter filled cells = %d, want 4", got)
	}
}

func TestProgressBarNonFiniteFraction(t *testing.T) {
	for _, tc := range []struct {
		name string
		frac float64
	}{
		{"nan", math.NaN()},
		{"posinf", math.Inf(1)},
		{"neginf", math.Inf(-1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := stripANSI(ProgressBar(tc.frac, 20))
			if got := utf8.RuneCountInString(out); got != 20 {
				t.Fatalf("ProgressBar(%v, 20) rendered %d cells (%d bytes), want 20", tc.frac, got, len(out))
			}
			if got := strings.Count(out, "█"); got != 0 {
				t.Errorf("ProgressBar(%v, 20) filled %d cells, want 0", tc.frac, got)
			}
		})
	}
}

func TestProgressBarZeroTotalDivision(t *testing.T) {
	done, total := 0, 0
	out := stripANSI(ProgressBar(float64(done)/float64(total), 20))
	if got := utf8.RuneCountInString(out); got != 20 {
		t.Fatalf("ProgressBar(0/0, 20) rendered %d cells (%d bytes), want 20", got, len(out))
	}
	if got := strings.Count(out, "░"); got != 20 {
		t.Errorf("ProgressBar(0/0, 20) empty cells = %d, want 20", got)
	}
}

func TestMeterNonFiniteFraction(t *testing.T) {
	for _, frac := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		out := stripANSI(Meter(frac, 20))
		if got := utf8.RuneCountInString(out); got != 22 {
			t.Fatalf("Meter(%v, 20) rendered %d cells (%d bytes), want 22", frac, got, len(out))
		}
		if !strings.HasPrefix(out, "[") || !strings.HasSuffix(out, "]") {
			t.Errorf("Meter(%v, 20) lost its brackets: %q", frac, out)
		}
	}
}

func TestProgressBarWellFormed(t *testing.T) {
	for _, tc := range []struct {
		frac   float64
		width  int
		filled int
	}{
		{0, 20, 0},
		{0.25, 20, 5},
		{0.5, 20, 10},
		{0.999, 20, 19},
		{1, 20, 20},
		{2, 20, 20},
		{-1, 20, 0},
		{0.5, 0, 0},
		{0.5, -5, 0},
	} {
		want := strings.Repeat("█", tc.filled) + strings.Repeat("░", max(tc.width-tc.filled, 0))
		if got := stripANSI(ProgressBar(tc.frac, tc.width)); got != want {
			t.Errorf("ProgressBar(%v, %d) = %q, want %q", tc.frac, tc.width, got, want)
		}
	}
}

// styleSGR is the escape prefix a style emits, so a test can tell which colour a
// rendered row was painted with.
func styleSGR(tb testing.TB, sty lipgloss.Style) string {
	tb.Helper()
	rendered := sty.Render("x")
	idx := strings.Index(rendered, "x")
	if idx <= 0 {
		tb.Fatalf("style emitted no escape prefix: %q", rendered)
	}
	return rendered[:idx]
}

// TestSeriesPanelsSurviveThemeWithoutSeries covers a hand-built Theme literal:
// theme.New always fills Series, but theme.Use accepts any Theme, and indexing
// i%len(Series) on an empty slice is an integer divide by zero.
func TestSeriesPanelsSurviveThemeWithoutSeries(t *testing.T) {
	prev := *theme.Cur()
	t.Cleanup(func() { theme.Use(prev) })
	theme.Use(theme.Theme{})

	f := layout.NewFrame(40)
	data := []Datum{{"a", 3}, {"b", 1}}
	if out := Pie(f, "Mix", data, 10, fnum, "none"); !strings.Contains(stripANSI(out), "a") {
		t.Errorf("Pie dropped its data under a Series-less theme:\n%s", stripANSI(out))
	}
	if out := Spectrum(f, "EQ", []float64{0.5, 1}, 4, "none"); !strings.Contains(stripANSI(out), "█") {
		t.Errorf("Spectrum drew nothing under a Series-less theme:\n%s", stripANSI(out))
	}
}

// TestBarNonFiniteSignAndLabelMatchTheBar pins the three views of one datum to
// the same value: bar.go clamps the magnitude with finite() but used to read the
// raw value for both the sign (colour) and the printed label, so -Inf drew a
// zero-length bar in the negative colour and NaN printed "NaN" next to it.
func TestBarNonFiniteSignAndLabelMatchTheBar(t *testing.T) {
	cant := styleSGR(t, theme.Cur().Cant)
	data := []Datum{{"nan", math.NaN()}, {"neginf", math.Inf(-1)}, {"posinf", math.Inf(1)}, {"ok", 5}}
	out := Bar(layout.NewFrame(60), "Flow", data, 20, fnum, "")

	for _, line := range strings.Split(out, "\n") {
		plain := stripANSI(line)
		for _, label := range []string{"nan", "neginf", "posinf"} {
			if !strings.Contains(plain, label) {
				continue
			}
			if strings.Contains(line, cant) {
				t.Errorf("row %q is painted in the negative colour but draws a zero-length bar: %q",
					label, plain)
			}
			if !strings.Contains(plain, fnum(0)) {
				t.Errorf("row %q prints %q, want the clamped value %q", label, plain, fnum(0))
			}
			for _, raw := range []string{"NaN", "Inf"} {
				if strings.Contains(plain, raw) {
					t.Errorf("row %q leaked the raw non-finite value: %q", label, plain)
				}
			}
		}
	}
}

// TestSpectrumNaNIsSilenceNotFullScale guards the worst cosmetic default: int(NaN)
// is math.MinInt64, and eighths-cell*8 wraps back to a large positive number, so a
// NaN level used to render as a near-full-height bar.
func TestSpectrumNaNIsSilenceNotFullScale(t *testing.T) {
	out := stripANSI(Spectrum(layout.NewFrame(40), "EQ", []float64{math.NaN()}, 6, "silent"))
	for _, g := range append([]string{}, vBlocks[:]...) {
		if strings.Contains(out, g) {
			t.Fatalf("a NaN level drew %q — NaN must read as silence, not maximum signal:\n%s", g, out)
		}
	}
	if got := spectrumCell(math.NaN(), 0, 6); got != "" {
		t.Errorf("spectrumCell(NaN, 0, 6) = %q, want \"\"", got)
	}
	if got := peakCell(math.NaN(), 6); got != -1 {
		t.Errorf("peakCell(NaN, 6) = %d, want -1", got)
	}
}
