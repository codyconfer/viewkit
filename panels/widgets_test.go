package panels

import (
	"math"
	"strings"
	"testing"
	"unicode/utf8"
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
