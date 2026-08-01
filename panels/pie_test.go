package panels

import (
	"math"
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

func TestPieNonFiniteValues(t *testing.T) {
	t.Parallel()
	data := []Datum{{"posinf", math.Inf(1)}, {"nan", math.NaN()}, {"neginf", math.Inf(-1)}, {"ok", 4}}
	out := Pie(layout.DefaultFrame(), "Mix", data, 10, fnum, "nothing here")
	if len(out) > 1<<16 {
		t.Fatalf("Pie with non-finite data rendered %d bytes", len(out))
	}
	stripped := stripANSI(out)
	for _, l := range strings.Split(stripped, "\n") {
		if got := strings.Count(l, "█"); got > 10 {
			t.Errorf("Pie bar line is %d cells, want <= width 10", got)
		}
	}
	if !strings.Contains(stripped, "ok") {
		t.Errorf("Pie dropped the finite datum:\n%s", stripped)
	}
}

func TestPieAllNonFiniteValues(t *testing.T) {
	t.Parallel()
	data := []Datum{{"posinf", math.Inf(1)}, {"nan", math.NaN()}}
	out := Pie(layout.DefaultFrame(), "Mix", data, 10, fnum, "nothing here")
	if len(out) > 1<<16 {
		t.Fatalf("Pie with all non-finite data rendered %d bytes", len(out))
	}
	if !strings.Contains(out, "nothing here") {
		t.Errorf("Pie with all non-finite data missing placeholder:\n%s", out)
	}
}

func TestPieTotalOverflowsToInf(t *testing.T) {
	t.Parallel()
	data := []Datum{{"a", math.MaxFloat64}, {"b", math.MaxFloat64}}
	out := Pie(layout.DefaultFrame(), "Mix", data, 10, fnum, "nothing here")
	if len(out) > 1<<16 {
		t.Fatalf("Pie with overflowing total rendered %d bytes", len(out))
	}
	if !strings.Contains(out, "nothing here") {
		t.Errorf("Pie with overflowing total missing placeholder:\n%s", out)
	}
}

func TestPieEmpty(t *testing.T) {
	t.Parallel()
	out := Pie(layout.DefaultFrame(), "Mix", []Datum{{"a", 0}, {"b", -5}}, 20, fnum, "nothing here")
	if !strings.Contains(out, "nothing here") {
		t.Fatalf("empty pie missing placeholder:\n%s", out)
	}
}

func TestPieLegendAndProportions(t *testing.T) {
	t.Parallel()
	data := []Datum{{"cash", 75}, {"eggs", 25}}
	out := Pie(layout.DefaultFrame(), "Mix", data, 20, fnum, "")
	for _, want := range []string{"cash", "eggs", "75%", "25%", "■", "█"} {
		if !strings.Contains(out, want) {
			t.Errorf("pie output missing %q:\n%s", want, out)
		}
	}
}

func TestPieBarStaysWithinWidth(t *testing.T) {
	t.Parallel()
	data := []Datum{{"a", 1}, {"b", 1}, {"c", 1}}
	lines := strings.Split(stripANSI(Pie(layout.DefaultFrame(), "Mix", data, 10, fnum, "")), "\n")

	var barCells int
	for _, l := range lines {
		if c := strings.Count(l, "█"); c > barCells {
			barCells = c
		}
	}
	if barCells > 10 {
		t.Errorf("stacked bar is %d cells, want <= width 10", barCells)
	}
}
